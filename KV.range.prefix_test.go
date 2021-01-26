package main_test

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"log"
	"testing"
	"time"
)

func TestRangePrefix(t *testing.T){
	cli := Client()
	defer cli.Close()

	for i := range make([]int, 3) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, err := cli.Put(ctx, fmt.Sprintf("key_%d", i), "value")
		cancel()
		if err != nil {
			log.Fatal(err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := cli.Get(ctx, "key", clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%s : %s\n", ev.Key, ev.Value)
	}
}
