package svc

import (
	"titan-tunnel/server/internal/config"
	"titan-tunnel/server/internal/middleware"
	"titan-tunnel/server/ws"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config        config.Config
	Redis         *redis.Redis
	TunMgr        *ws.TunnelManager
	JwtMiddleware rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	redis := redis.MustNewRedis(c.Redis)
	return &ServiceContext{
		Config:        c,
		Redis:         redis,
		TunMgr:        ws.NewTunnelManager(c, redis),
		JwtMiddleware: middleware.NewJwtMiddleware(c.JwtAuth.AccessSecret).Handle,
	}
}
