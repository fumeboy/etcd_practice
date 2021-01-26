package main_test

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"testing"
)

func TestWatchRange(t *testing.T)  {
	cli := Client()
	// watches within ['foo1', 'foo4'), in lexicographical order
	rch := cli.Watch(context.Background(), "foo1", clientv3.WithRange("foo4"))

	go func() {
		cli.Put(context.Background(), "foo1", "bar1")
		cli.Put(context.Background(), "foo5", "bar5")
		cli.Put(context.Background(), "foo2", "bar2")
		cli.Put(context.Background(), "foo3", "bar3")
	}()

	i := 0
	for wresp := range rch {
		for _, ev := range wresp.Events {
			fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			i++
			if i == 3 {
				// After 3 messages we are done.
				cli.Delete(context.Background(), "foo", clientv3.WithPrefix())
				cli.Close()
				return
			}
		}
	}
	// Output:
	// PUT "foo1" : "bar1"
	// PUT "foo2" : "bar2"
	// PUT "foo3" : "bar3"
}
