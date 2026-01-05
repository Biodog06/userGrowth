package logs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	config "usergrowth/configs"

	elastic "github.com/elastic/go-elasticsearch/v8"
	"github.com/gogf/gf/v2/frame/g"
)

// EsLogsReq 查询日志请求参数
type EsLogsReq struct {
	g.Meta    `path:"/api/eslog" method:"get"`
	Keyword   string `p:"keyword"`     // 关键词
	Level     string `p:"level"`       // 日志级别
	Topic     string `p:"topic"`       // 日志主题
	StartTime string `p:"start_time"`  // 开始时间
	EndTime   string `p:"end_time"`    // 结束时间
	Page      int    `p:"page" d:"1"`  // 页码，默认1
	Size      int    `p:"size" d:"10"` // 每页条数，默认10
}

// EsLogsRes 日志查询响应
type EsLogsRes struct {
	Total int           `json:"total"`
	Data  []interface{} `json:"data"`
}

type EsController struct {
	client *elastic.Client
	index  string
}

func NewEsController(cfg *config.Config) *EsController {
	esCfg := elastic.Config{
		Addresses: []string{fmt.Sprintf("http://%s:%s", cfg.Elasticsearch.Host, strconv.Itoa(cfg.Elasticsearch.Port))},
	}
	client, err := elastic.NewClient(esCfg)
	if err != nil {
		panic(err)
	}
	return &EsController{
		client: client,
		index:  cfg.Elasticsearch.LogIndex,
	}
}

func (h *EsController) GetLogs(ctx context.Context, req *EsLogsReq) (res *EsLogsRes, err error) {
	r := g.RequestFromCtx(ctx)

	from := (req.Page - 1) * req.Size
	if from < 0 {
		from = 0
	}

	query := g.Map{
		"from": from,
		"size": req.Size,
		"sort": []g.Map{
			{"@timestamp": g.Map{"order": "desc"}},
		},
		"query": g.Map{
			"bool": g.Map{
				"must":   []g.Map{},
				"filter": []g.Map{},
			},
		},
	}

	boolQuery := query["query"].(g.Map)["bool"].(g.Map)
	mustList := boolQuery["must"].([]g.Map)
	filterList := boolQuery["filter"].([]g.Map)

	if req.Keyword != "" {
		mustList = append(mustList, g.Map{
			"match": g.Map{"message": req.Keyword},
		})
	}

	if req.Level != "" {
		filterList = append(filterList, g.Map{
			"term": g.Map{"level": req.Level},
		})
	}

	if req.Topic != "" {
		filterList = append(filterList, g.Map{
			"term": g.Map{"log_topic": req.Topic},
		})
	}

	if req.StartTime != "" || req.EndTime != "" {
		rangeQuery := g.Map{}
		if req.StartTime != "" {
			rangeQuery["gte"] = req.StartTime
		}
		if req.EndTime != "" {
			rangeQuery["lte"] = req.EndTime
		}
		filterList = append(filterList, g.Map{
			"range": g.Map{"@timestamp": rangeQuery},
		})
	}

	boolQuery["must"] = mustList
	boolQuery["filter"] = filterList

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		// 返回给前端的错误响应
		r.Response.WriteStatus(http.StatusInternalServerError)
		r.Response.WriteJson(g.Map{"code": 500, "message": "构建查询失败"})
		// 返回给中间件的错误（用于日志）
		return nil, fmt.Errorf("构建查询失败: %w", err)
	}

	// 执行 ES 查询
	esRes, err := h.client.Search(
		h.client.Search.WithContext(ctx),
		h.client.Search.WithIndex(h.index),
		h.client.Search.WithBody(&buf),
		h.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError)
		r.Response.WriteJson(g.Map{"code": 500, "message": "ES查询失败"})
		return nil, fmt.Errorf("ES查询失败: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(esRes.Body)

	// 检查 ES 响应状态
	if esRes.IsError() {
		var errRes map[string]interface{}
		_ = json.NewDecoder(esRes.Body).Decode(&errRes)
		r.Response.WriteStatus(esRes.StatusCode)
		r.Response.WriteJson(g.Map{
			"code":    esRes.StatusCode,
			"message": "ES返回错误",
			"data":    errRes,
		})
		return nil, fmt.Errorf("ES返回错误: %v", errRes)
	}

	// 解析 ES 响应体
	var rMap map[string]interface{}
	if err = json.NewDecoder(esRes.Body).Decode(&rMap); err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError)
		r.Response.WriteJson(g.Map{"code": 500, "message": "解析响应失败"})
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	hitsRoot, ok := rMap["hits"].(map[string]interface{})
	if !ok {
		r.Response.WriteStatus(http.StatusInternalServerError)
		r.Response.WriteJson(g.Map{"code": 500, "message": "ES响应格式异常"})
		return nil, fmt.Errorf("ES响应格式异常")
	}

	total := 0
	if totalMap, ok := hitsRoot["total"].(map[string]interface{}); ok {
		total = int(totalMap["value"].(float64))
	}

	results := make([]interface{}, 0)
	if rawHits, ok := hitsRoot["hits"].([]interface{}); ok {
		for _, hit := range rawHits {
			if source, ok := hit.(map[string]interface{})["_source"]; ok {
				results = append(results, source)
			}
		}
	}

	// 3. 返回结果 (遵循 login.go 模式，自己写 Response)
	r.Response.WriteJson(g.Map{
		"code":    200,
		"message": "success",
		"total":   total,
		"data":    results,
	})

	return nil, nil
}
