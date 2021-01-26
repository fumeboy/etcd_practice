package main_test

import (
	"context"
	"fmt"
	"testing"
	"time"
)
// put 设置键值对数据
func TestPut(t *testing.T) {
	cli := Client()
	defer cli.Close()
	// put
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err := cli.Put(ctx, "foo", "bar2")
	cancel()
	if err != nil {
		fmt.Printf("put to etcd failed")
		panic(err)
	}
}