package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	// Redis     redis.RedisConf
	PopID     string
	APIServer string
}
