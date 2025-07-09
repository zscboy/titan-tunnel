package model

import (
	"testing"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

func TestBind(t *testing.T) {
	conf := redis.RedisConf{Host: "127.0.0.1:6379", Type: "node"}
	rd := redis.MustNewRedis(conf)
	BindNode(rd, "95e66ae2-5bca-11f0-9654-00163e0ced7c")
}

func TestUnbind(t *testing.T) {
	conf := redis.RedisConf{Host: "127.0.0.1:6379", Type: "node"}
	rd := redis.MustNewRedis(conf)
	UnbindNode(rd, "95e66ae2-5bca-11f0-9654-00163e0ced7c")
}
