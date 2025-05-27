package svc

import (
	"titan-vm/vms/api/ws"
	"titan-vm/vms/internal/config"
	"titan-vm/vms/vms"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	Vms    vms.Vms
	Redis  *redis.Redis
	TunMgr *ws.TunnelManager
}

func NewServiceContext(c config.Config) *ServiceContext {
	redis := redis.MustNewRedis(c.Redis.RedisConf)
	return &ServiceContext{
		Config: c,
		Vms:    vms.NewVms(zrpc.MustNewClient(c.RpcClient)),
		Redis:  redis,
		TunMgr: ws.NewTunnelManager(redis),
	}
}
