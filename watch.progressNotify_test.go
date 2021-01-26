package main_test

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"log"
	"testing"
	"time"
)

func TestWatchProgressNotify(t *testing.T){
	cli := Client()
	rch := cli.Watch(context.Background(), "foo", clientv3.WithProgressNotify())
	closedch := make(chan bool)
	go func() {
		// This assumes that cluster is configured with frequent WatchProgressNotifyInterval
		// e.g. WatchProgressNotifyInterval: 200 * time.Millisecond.
		time.Sleep(time.Second)
		err := cli.Close()
		if err != nil {
			log.Fatal(err)
		}
		close(closedch)
	}()
	wresp := <-rch
	for _, _ = range wresp.Events {
		// 设置 progress_notify 后 ，如果最近没有事件，etcd 服务器将定期的发送不带任何事件的 WatchResponse 给新的观察者
		// 当客户端希望从最近已知的修订版本开始恢复断开的观察者时有用
		// etcd 服务器将基于当前负载决定它发送通知的频率。
		panic("")
	}
	fmt.Println("wresp.IsProgressNotify:", wresp.IsProgressNotify())
	<-closedch
}
