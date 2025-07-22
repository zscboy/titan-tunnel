package config

import (
	api "titan-tunnel/server/api/export"
	rpc "titan-tunnel/server/rpc/export"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type Config struct {
	APIServer api.APIServerConfig
	RPCServer rpc.RPCServerConfig
	Redis     redis.RedisConf
	Log       logx.LogConf
	// HTTPProxy string
}
