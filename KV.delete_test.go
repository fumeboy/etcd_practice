package main_test

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"log"
	"testing"
	"time"
)

func TestDelete(t *testing.T){
	cli := Client()
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, _ = cli.Put(ctx, "foo1", "bar")
	_, _ = cli.Put(ctx, "foo2", "bar")
	_, _ = cli.Put(ctx, "foo3", "bar")

	gresp, err := cli.Get(ctx, "foo", clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
	}

	dresp, err := cli.Delete(ctx, "foo", clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
	}

	cancel()
	// 比较删除量和原来持有量
	fmt.Println("Deleted all keys:", int64(len(gresp.Kvs)) == dresp.Deleted)
}
