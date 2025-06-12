package model

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

func TestRedis(t *testing.T) {
	conf := redis.RedisConf{Host: "127.0.0.1:6379", Type: "node"}
	rd := redis.MustNewRedis(conf)

	node := Node{Id: "123", OS: "linux", VmAPI: "libvirt", CPU: 4, Memory: 10000, LoginAt: time.Now().String(), RegisterAt: time.Now().String()}
	err := SetNode(context.Background(), rd, &node)
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

func TestTime(t *testing.T) {
	timeStr := "2025-06-11 16:16:07.6743289 +0800 CST m=+3.863716201"
	layout := "2006-01-02 15:04:05.999999999 -0700 MST"

	// 解析时忽略 "m=+..." 部分
	tim, err := time.Parse(layout, timeStr)
	if err != nil {
		fmt.Println("解析失败:", err)
		return
	}
	fmt.Println("解析结果:", tim) // 2025-06-11 16:16:07.6743289 +0800 CST
}

// func TestFiledExpire() {
// 	redisConf := redis.RedisConf{
// 		Host: "localhost:6379",
// 		Type: "node", // 单节点模式
// 	}

// 	// 获取 go-zero 的 Redis 操作对象
// 	r, err := redis.NewRedis(redisConf)
// 	if err != nil {
// 		panic(err)
// 	}

// 	defer goRedisClient.Close()
// }
