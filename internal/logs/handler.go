package logs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// SearchRequest 查询参数
type SearchRequest struct {
	Keyword   string `p:"keyword"`
	Level     string `p:"level"`
	StartTime string `p:"start_time"`
	EndTime   string `p:"end_time"`
	Page      int    `p:"page"`
	Size      int    `p:"size"`
}

// 标准返回格式
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Total   int         `json:"total"`
}

func GetLogs(h *MyAsyncEs) func(r *ghttp.Request) {
	return func(r *ghttp.Request) {
		var req SearchRequest
		// 1. 绑定并校验参数
		if err := r.Parse(&req); err != nil {
			r.Response.WriteJson(ghttp.DefaultHandlerResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})
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
				"range": map[string]interface{}{"ts": rangeQuery},
			})
		}

		// 3. 执行查询
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(query); err != nil {
			r.Response.WriteJson(ghttp.DefaultHandlerResponse{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			})
			return
		}

		res, err := h.client.Search(
			h.client.Search.WithContext(context.Background()),
			h.client.Search.WithIndex(h.index),
			h.client.Search.WithBody(&buf),
			h.client.Search.WithTrackTotalHits(true),
		)
		if err != nil {
			r.Response.WriteJson(ghttp.DefaultHandlerResponse{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			})
			return
		}
		defer func(Body io.ReadCloser) {
			err = Body.Close()
			if err != nil {
				fmt.Println(err)
				return
			}
		}(res.Body)

		// 1. 检查 HTTP 状态码是否为 200
		if res.IsError() {
			var errRes map[string]interface{}
			err = json.NewDecoder(res.Body).Decode(&errRes)
			if err != nil {
				r.Response.WriteJson(ghttp.DefaultHandlerResponse{
					Code:    http.StatusInternalServerError,
					Message: err.Error(),
				})
				return
			}
			fmt.Printf("ES 查询报错: %+v\n", errRes) // 在控制台打印具体的 DSL 错误
			r.Response.WriteStatus(res.StatusCode)
			r.Response.WriteJson(g.Map{
				"error":  "ES查询失败",
				"detail": errRes,
			})
			return
		}

		var raw map[string]interface{}
		if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
			r.Response.WriteJson(ghttp.DefaultHandlerResponse{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			})
			return
		}

		// 2. 安全检查 hits 字段是否存在
		hitsRoot, ok := raw["hits"].(map[string]interface{})
		if !ok {
			// 如果走到这里，说明响应里根本没有 hits
			fmt.Printf("ES 响应异常，完整内容: %+v\n", r)
			r.Response.WriteJson(ghttp.DefaultHandlerResponse{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			})
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
		r.Response.WriteStatus(http.StatusOK)
		r.Response.WriteJson(g.Map{
			"total": total,
			"data":  results,
		})
	}
}
