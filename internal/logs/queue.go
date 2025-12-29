package logs

import (
	"fmt"
	"strconv"
	"sync"
	config "usergrowth/configs"

	elastic "github.com/elastic/go-elasticsearch/v8"
)

type MyAsyncEs struct {
	client    *elastic.Client
	index     string
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
	es := &MyAsyncEs{
		client:    client,
		index:     cfg.ES.LogIndex,
		logChan:   make(chan []byte, cfg.ES.MaxQueueSize),
		quitChan:  make(chan struct{}),
		workers:   cfg.ES.Workers,
		batchSize: cfg.ES.MaxBatchSize,
	}
	es.StartWorkerPools(es.workers)
	return es
}

func (es *MyAsyncEs) Write(p []byte) (n int, err error) {
	// 【关键】必须进行深拷贝！
	// Zap 传入的 p 指向的是它内部的缓冲区，该缓冲区会被立刻复用。
	data := make([]byte, len(p))
	copy(data, p)
	select {
	case es.logChan <- data:
		return len(data), nil
	default:
		fmt.Println("channel full")
		return 0, fmt.Errorf("channel full")
	}
}

func (es *MyAsyncEs) Sync() error {
	close(es.logChan)
	es.wg.Wait()
	return nil
}
