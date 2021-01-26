package main_test

import (
	"context"
	"fmt"
	"testing"
)

// watch 监听数据变动

func TestWatch(t *testing.T){
	cli := Client()
	defer cli.Close()
	// watch key:foo change
	w := cli.Watch(context.Background(), "foo") // <-chan WatchResponse
	// 现在你可以在别处 put 这个 key 的值，看看这个程序的反应
	for resp := range w {
		for _, ev := range resp.Events {
			fmt.Printf("Type: %s Key:%s Value:%s\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
		}
	}
}
