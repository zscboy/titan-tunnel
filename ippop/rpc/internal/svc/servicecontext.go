package svc

import (
	"titan-ipoverlay/ippop/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type ServiceContext struct {
	Config config.Config
	Redis  *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	redis := redis.MustNewRedis(c.Redis.RedisConf)
	return &ServiceContext{
		Config: c,
		Redis:  redis,
	}
}
