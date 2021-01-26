package main_test

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"testing"
)

func TestWatchPrefix(t *testing.T){
	cli := Client()
	defer cli.Close()

	rch := cli.Watch(context.Background(), "foo", clientv3.WithPrefix())

	// 现在 PUT "foo1" : "bar" 看看结果
	for wresp := range rch {
		for _, ev := range wresp.Events {
			fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
		}
	}
}