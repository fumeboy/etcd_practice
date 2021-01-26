package main_test

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"log"
	"testing"
)

func TestDo(t *testing.T){
	cli := Client()
	defer cli.Close()

	ops := []clientv3.Op{
		clientv3.OpPut("put-key", "123"),
		clientv3.OpGet("put-key"),
		clientv3.OpPut("put-key", "456")}

	for _, op := range ops {
		if _, err := cli.Do(context.TODO(), op); err != nil {
			log.Fatal(err)
		}
	}
}
