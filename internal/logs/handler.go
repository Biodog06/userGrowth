package logs

import (
	"time"
)
import "github.com/olivere/elastic/v7"

type MyEs struct {
	*elastic.Client
}

func NewEsClient() *MyEs {
	client, err := elastic.NewClient(
		elastic.SetURL("http://127.0.0.1:9200"),
		elastic.SetHealthcheckInterval(10*time.Second),
		elastic.SetSniff(false),
		elastic.SetBasicAuth("", ""),
	)
	if err != nil {
		panic(err)
	}
	return &MyEs{client}
}
