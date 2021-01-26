package main_test

import (
	"context"
	"log"
	"testing"
	"time"
)

func TestCompact(t *testing.T){
	cli := Client()
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := cli.Get(ctx, "foo")
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	compRev := resp.Header.Revision // specify compact revision of your choice

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	err = cli.Compact(ctx, compRev)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
}
