package logs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	config "usergrowth/configs"

	elastic "github.com/elastic/go-elasticsearch/v8"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtrace"
	"github.com/gogf/gf/v2/text/gstr"
)

type EsLogsReq struct {
	g.Meta    `path:"/api/eslog" method:"get"`
	Keyword   string `p:"keyword"`
	Level     string `p:"level"`
	Topic     string `p:"topic"`
	StartTime string `p:"start_time"`
	EndTime   string `p:"end_time"`
	Page      int    `p:"page" d:"1"`
	Size      int    `p:"size" d:"10"`
}

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
	ctx, span := gtrace.NewSpan(ctx, "eslog")
	defer span.End()

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
		level := gstr.ToUpper(req.Level)
		switch level {
		case "ERROR":
			level = "ERRO"
		case "DEBUG":
			level = "DEBU"
		case "WARNING":
			level = "WARN"
		}

		filterList = append(filterList, g.Map{
			"term": g.Map{"level": level},
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
		r.Response.WriteJson(g.Map{"code": 500, "message": "构建查询失败"})
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
		r.Response.WriteJson(g.Map{"code": 500, "message": "ES查询失败"})
		return nil, fmt.Errorf("ES查询失败: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(esRes.Body)

	if esRes.IsError() {
		var errRes map[string]interface{}
		_ = json.NewDecoder(esRes.Body).Decode(&errRes)
		r.Response.WriteJson(g.Map{
			"code":    esRes.StatusCode,
			"message": "ES返回错误",
			"data":    errRes,
		})
		return nil, fmt.Errorf("ES返回错误: %v", errRes)
	}

	// 解析 ES 响应
	var rMap map[string]interface{}
	if err = json.NewDecoder(esRes.Body).Decode(&rMap); err != nil {
		r.Response.WriteJson(g.Map{"code": 500, "message": "解析响应失败"})
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	hitsRoot, ok := rMap["hits"].(map[string]interface{})
	if !ok {
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

	r.Response.WriteJson(g.Map{
		"code":    200,
		"message": "success",
		"total":   total,
		"data":    results,
	})

	return nil, nil
}
