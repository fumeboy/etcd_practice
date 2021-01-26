package main_test

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	cli := Client()
	defer cli.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := cli.Get(ctx, "foo")
	cancel()
	if err != nil {
		fmt.Printf("get from etcd failed")
		panic(err)
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%s:%s\n", ev.Key, ev.Value)
	}
}
