package svc

import (
	"titan-tunnel/server/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type ServiceContext struct {
	Config config.Config
	Redis  *redis.Redis
	// TunMgr *ws.TunnelManager
	// JwtMiddleware rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	redis := redis.MustNewRedis(c.Redis.RedisConf)
	return &ServiceContext{
		Config: c,
		Redis:  redis,
		// TunMgr: ws.NewTunnelManager(c, redis),
		// JwtMiddleware: middleware.NewJwtMiddleware(c.JwtAuth.AccessSecret).Handle,
	}
}
