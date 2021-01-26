package main_test

import (
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

func Client() *clientv3.Client{
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"192.168.123.2:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println("connect to etcd failed")
		panic(err)
	}
	fmt.Println("connect to etcd success")
	return cli
}
