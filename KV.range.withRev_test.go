package main_test

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"log"
	"testing"
	"time"
)

func TestRangeWithRev(t *testing.T){
	cli := Client()
	defer cli.Close()

	presp, err := cli.Put(context.TODO(), "foo", "bar1")
	if err != nil {
		log.Fatal(err)
	}
	_, err = cli.Put(context.TODO(), "foo", "bar2")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := cli.Get(ctx, "foo", clientv3.WithRev(presp.Header.Revision)) // 指定版本
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%s : %s\n", ev.Key, ev.Value)
	}
}
