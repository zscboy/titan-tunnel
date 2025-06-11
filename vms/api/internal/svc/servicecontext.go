package svc

import (
	"titan-vm/vms/api/internal/middleware"
	"titan-vm/vms/api/ws"
	"titan-vm/vms/internal/config"
	"titan-vm/vms/vms"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config        config.Config
	Vms           vms.Vms
	Redis         *redis.Redis
	TunMgr        *ws.TunnelManager
	JwtMiddleware rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	redis := redis.MustNewRedis(c.Redis.RedisConf)
	return &ServiceContext{
		Config:        c,
		Vms:           vms.NewVms(zrpc.MustNewClient(c.RpcClient)),
		Redis:         redis,
		TunMgr:        ws.NewTunnelManager(c, redis),
		JwtMiddleware: middleware.NewJwtMiddleware(c.JwtAuth.AccessSecret).Handle,
	}
}
