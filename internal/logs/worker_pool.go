package logs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"
)

func (es *MyAsyncEs) StartWorkerPools(workers int) {
	for i := 0; i < es.workers; i++ {
		es.wg.Add(workers)
		go es.BatchProcess()
	}
}

func (es *MyAsyncEs) BatchProcess() {
	defer es.wg.Done()
	batch := make([][]byte, 0, es.batchSize)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case zapLog, ok := <-es.logChan:
			if !ok {
				if len(batch) > 0 {
					es.doFlush(batch)
				}
				fmt.Println("channel closed")
				return
			}
			batch = append(batch, zapLog)
			fmt.Printf("batch size: %d\n", len(batch))
			if len(batch) >= es.batchSize {
				es.doFlush(batch)
				batch = batch[:0]
				ticker.Reset(5 * time.Second)
			}
		case <-ticker.C:
			if len(batch) > 0 {
				es.doFlush(batch)
				batch = batch[:0]
			}
		}
	}
}

func (es *MyAsyncEs) doFlush(batch [][]byte) {
	// 不如直接用 BulkIndex
	var buf bytes.Buffer
	for _, b := range batch {
		buf.WriteString(`{"index":{}}`)
		buf.WriteByte('\n')
		buf.Write(b)
		if !bytes.HasSuffix(b, []byte("\n")) {
			buf.WriteByte('\n')
		}
	}
	res, err := es.client.Bulk(
		bytes.NewReader(buf.Bytes()),
		es.client.Bulk.WithContext(context.Background()),
		es.client.Bulk.WithIndex(es.index),
	)
	if err != nil {
		fmt.Println("bulk fail")
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("close fail")
		}
	}(res.Body)
}
