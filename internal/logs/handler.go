package logs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	config "usergrowth/configs"

	elastic "github.com/elastic/go-elasticsearch/v8"
)

type MyEs struct {
	client *elastic.Client
	index  string
}

func NewEsClient(cfg *config.Config) *MyEs {
	esCfg := elastic.Config{
		Addresses: []string{fmt.Sprintf("http://%s:%s", cfg.ES.Host, strconv.Itoa(cfg.ES.Port))},
	}
	es, err := elastic.NewClient(esCfg)
	if err != nil {
		panic(err)
	}
	return &MyEs{client: es, index: "zap-logs"}
}

func (es *MyEs) Write(p []byte) (n int, err error) {
	res, err := es.client.Index(
		es.index,
		bytes.NewReader(p),
		es.client.Index.WithContext(context.Background()),
	)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("not close:", err)
		}
	}(res.Body)
	return len(p), nil
}
