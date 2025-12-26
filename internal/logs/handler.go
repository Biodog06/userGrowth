package logs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SearchRequest 查询参数
type SearchRequest struct {
	Keyword   string `form:"keyword"`    // 关键词
	Level     string `form:"level"`      // 日志级别
	StartTime string `form:"start_time"` // 开始时间 (ISO8601)
	EndTime   string `form:"end_time"`   // 结束时间
	Page      int    `form:"page"`       // 页码
	Size      int    `form:"size"`       // 每页条数
}

// 标准返回格式
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Total   int         `json:"total"`
}

func GetLogs(h *MyAsyncEs) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SearchRequest
		// 1. 绑定并校验参数
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 设置分页默认值
		if req.Size <= 0 {
			req.Size = 10
		}
		from := (req.Page - 1) * req.Size
		if from < 0 {
			from = 0
		}

		// 2. 构建复杂的 ES DSL 查询语句
		// 我们使用 bool 查询：must 用于搜索关键词，filter 用于过滤级别和时间（filter不计分，性能更好）
		query := map[string]interface{}{
			"from": from,
			"size": req.Size,
			"sort": []map[string]interface{}{
				{"ts": map[string]interface{}{"order": "desc"}}, // 按时间倒序
			},
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"must":   []map[string]interface{}{},
					"filter": []map[string]interface{}{},
				},
			},
		}

		// 动态添加过滤条件
		boolQuery := query["query"].(map[string]interface{})["bool"].(map[string]interface{})

		// 如果有关键词，对 message 字段进行模糊匹配
		if req.Keyword != "" {
			must := boolQuery["must"].([]map[string]interface{})
			boolQuery["must"] = append(must, map[string]interface{}{
				"match": map[string]interface{}{"message": req.Keyword},
			})
		}

		// 如果有级别过滤
		if req.Level != "" {
			filter := boolQuery["filter"].([]map[string]interface{})
			boolQuery["filter"] = append(filter, map[string]interface{}{
				"term": map[string]interface{}{"level": req.Level},
			})
		}

		// 如果有时间范围过滤
		if req.StartTime != "" || req.EndTime != "" {
			rangeQuery := map[string]interface{}{}
			if req.StartTime != "" {
				rangeQuery["gte"] = req.StartTime
			}
			if req.EndTime != "" {
				rangeQuery["lte"] = req.EndTime
			}

			filter := boolQuery["filter"].([]map[string]interface{})
			boolQuery["filter"] = append(filter, map[string]interface{}{
				"range": map[string]interface{}{"timestamp": rangeQuery},
			})
		}

		// 3. 执行查询
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(query); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "JSON encode error"})
		}

		res, err := h.client.Search(
			h.client.Search.WithContext(context.Background()),
			h.client.Search.WithIndex(h.index),
			h.client.Search.WithBody(&buf),
			h.client.Search.WithTrackTotalHits(true),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ES query error"})
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Println(err)
				return
			}
		}(res.Body)

		// 1. 检查 HTTP 状态码是否为 200
		if res.IsError() {
			var errRes map[string]interface{}
			json.NewDecoder(res.Body).Decode(&errRes)
			fmt.Printf("ES 查询报错: %+v\n", errRes) // 在控制台打印具体的 DSL 错误
			c.JSON(res.StatusCode, gin.H{
				"error":  "ES查询失败",
				"detail": errRes,
			})
			return
		}

		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			c.JSON(500, gin.H{"error": "解析响应失败"})
			return
		}

		// 2. 安全检查 hits 字段是否存在
		hitsRoot, ok := r["hits"].(map[string]interface{})
		if !ok {
			// 如果走到这里，说明响应里根本没有 hits
			fmt.Printf("ES 响应异常，完整内容: %+v\n", r)
			c.JSON(500, gin.H{"error": "ES响应格式不符合预期"})
			return
		}

		// --- 关键防御代码结束 ---

		// 提取数据
		total := 0
		if totalMap, ok := hitsRoot["total"].(map[string]interface{}); ok {
			total = int(totalMap["value"].(float64))
		}

		rawHits := hitsRoot["hits"].([]interface{})
		results := make([]interface{}, 0)
		for _, hit := range rawHits {
			if source, ok := hit.(map[string]interface{})["_source"]; ok {
				results = append(results, source)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"code":  200,
			"total": total,
			"data":  results,
		})
	}
}
