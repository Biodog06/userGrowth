package logs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync"
	config "usergrowth/configs"

	elastic "github.com/elastic/go-elasticsearch/v8"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// SearchRequest 查询参数
// GoFrame 使用 `p` 标签来获取参数 (Query/Form/Json 通用)
// 使用 `d` 标签设置默认值
type SearchRequest struct {
	Keyword   string `p:"keyword"`     // 关键词
	Level     string `p:"level"`       // 日志级别
	StartTime string `p:"start_time"`  // 开始时间
	EndTime   string `p:"end_time"`    // 结束时间
	Page      int    `p:"page" d:"1"`  // 页码，默认1
	Size      int    `p:"size" d:"10"` // 每页条数，默认10
}

type MyAsyncEs struct {
	client *elastic.Client
	index  string
	// 既然已经改用 Filebeat 写日志了，下面的这些写相关的字段其实可以删掉了
	// 这里保留是为了不破坏你的结构体定义
	logChan   chan []byte
	quitChan  chan struct{}
	workers   int
	wg        sync.WaitGroup
	batchSize int
}

func NewEsClient(cfg *config.Config) *MyAsyncEs {
	esCfg := elastic.Config{
		Addresses: []string{fmt.Sprintf("http://%s:%s", cfg.ES.Host, strconv.Itoa(cfg.ES.Port))},
	}
	client, err := elastic.NewClient(esCfg)
	if err != nil {
		panic(err)
	}
	return &MyAsyncEs{
		client:    client,
		index:     cfg.ES.LogIndex,
		logChan:   make(chan []byte, cfg.ES.MaxQueueSize),
		quitChan:  make(chan struct{}),
		workers:   cfg.ES.Workers,
		batchSize: cfg.ES.MaxBatchSize,
	}
}

// GetLogs 搜索日志 Handler
// 返回值从 gin.HandlerFunc 改为 ghttp.HandlerFunc
func GetLogs(h *MyAsyncEs) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		var req SearchRequest
		// 1. 绑定并校验参数 (GoFrame 会自动处理默认值)
		if err := r.Parse(&req); err != nil {
			r.Response.WriteJson(g.Map{
				"code": 400,
				"msg":  "参数错误: " + err.Error(),
			})
			return
		}

		// 计算分页 offset
		from := (req.Page - 1) * req.Size
		if from < 0 {
			from = 0
		}

		// 2. 构建 ES DSL 查询语句
		// 使用 g.Map 简化写法
		query := g.Map{
			"from": from,
			"size": req.Size,
			"sort": []g.Map{
				// 注意：如果你用 Filebeat，标准时间字段通常是 "@timestamp"
				// 如果你代码里存的是 "ts"，请保持 "ts"
				{"@timestamp": g.Map{"order": "desc"}},
			},
			"query": g.Map{
				"bool": g.Map{
					"must":   []g.Map{},
					"filter": []g.Map{},
				},
			},
		}

		// 动态添加过滤条件
		// 这里需要做类型断言来修改 map 内部的值
		boolQuery := query["query"].(g.Map)["bool"].(g.Map)
		mustList := boolQuery["must"].([]g.Map)
		filterList := boolQuery["filter"].([]g.Map)

		// A. 关键词模糊匹配
		if req.Keyword != "" {
			// 假设你配置 Filebeat 时解析出来的字段在 "message" 或者 "content"
			// 甚至可以使用 "multi_match" 同时搜多个字段
			mustList = append(mustList, g.Map{
				"match": g.Map{"message": req.Keyword},
			})
		}

		// B. 级别过滤
		if req.Level != "" {
			filterList = append(filterList, g.Map{
				"term": g.Map{"level": req.Level},
			})
		}

		// C. 时间范围过滤
		if req.StartTime != "" || req.EndTime != "" {
			rangeQuery := g.Map{}
			if req.StartTime != "" {
				rangeQuery["gte"] = req.StartTime
			}
			if req.EndTime != "" {
				rangeQuery["lte"] = req.EndTime
			}
			// 注意：这里的时间字段名要和 sort 里的一致
			filterList = append(filterList, g.Map{
				"range": g.Map{"@timestamp": rangeQuery},
			})
		}

		// 把修改后的切片赋值回去
		boolQuery["must"] = mustList
		boolQuery["filter"] = filterList

		// 3. 执行查询
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(query); err != nil {
			r.Response.WriteJson(g.Map{"code": 500, "msg": "构建查询失败"})
			return
		}

		// 调用 ES 客户端
		res, err := h.client.Search(
			h.client.Search.WithContext(r.Context()), // 使用 r.Context()
			h.client.Search.WithIndex(h.index),       // 确保 index 名称正确 (如 usergrowth-*)
			h.client.Search.WithBody(&buf),
			h.client.Search.WithTrackTotalHits(true),
		)
		if err != nil {
			g.Log().Error(r.Context(), "ES query failed", err)
			r.Response.WriteJson(g.Map{"code": 500, "msg": "ES查询失败"})
			return
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(res.Body)

		// 4. 处理 ES 响应错误
		if res.IsError() {
			var errRes map[string]interface{}
			_ = json.NewDecoder(res.Body).Decode(&errRes)
			g.Log().Error(r.Context(), "ES response error", errRes)

			r.Response.WriteJson(g.Map{
				"code": res.StatusCode,
				"msg":  "ES返回错误",
				"data": errRes,
			})
			return
		}

		// 5. 解析结果
		var rMap map[string]interface{}
		if err = json.NewDecoder(res.Body).Decode(&rMap); err != nil {
			r.Response.WriteJson(g.Map{"code": 500, "msg": "解析响应失败"})
			return
		}

		// 安全提取 hits
		hitsRoot, ok := rMap["hits"].(map[string]interface{})
		if !ok {
			r.Response.WriteJson(g.Map{"code": 500, "msg": "ES响应格式异常"})
			return
		}

		// 提取总数
		total := 0
		if totalMap, ok := hitsRoot["total"].(map[string]interface{}); ok {
			total = int(totalMap["value"].(float64))
		}

		// 提取数据列表
		results := make([]interface{}, 0)
		if rawHits, ok := hitsRoot["hits"].([]interface{}); ok {
			for _, hit := range rawHits {
				if source, ok := hit.(map[string]interface{})["_source"]; ok {
					results = append(results, source)
				}
			}
		}

		// 6. 返回成功响应
		r.Response.WriteJson(g.Map{
			"code":  200,
			"msg":   "success",
			"total": total,
			"data":  results,
		})
	}
}
