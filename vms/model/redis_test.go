package model

import (
	"context"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

func TestRedis(t *testing.T) {
	conf := redis.RedisConf{Host: "127.0.0.1:6379", Type: "node"}
	rd := redis.MustNewRedis(conf)

	node := Node{Id: "123", OS: "linux", VmAPI: "libvirt", CPU: 4, Memory: 10000, LoginAt: time.Now().String(), RegisterAt: time.Now().String()}
	err := RegisterNode(context.Background(), rd, &node)
	if err != nil {
		t.Fatalf("register node %s", err.Error())
	}
}

func TestGetAccount(t *testing.T) {
	conf := redis.RedisConf{Host: "127.0.0.1:6379", Type: "node"}
	rd := redis.MustNewRedis(conf)

	ac, err := GetAccount(rd, "abc")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("account:%v", ac)
}
