package svc

import (
	"fmt"
	"titan-vm/vms/internal/config"
	"titan-vm/vms/virt"
	"titan-vm/vms/virt/multipass"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type ServiceContext struct {
	Config config.Config
	Virt   *virt.Virt
	Redis  *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	serverURL := fmt.Sprintf("ws://localhost:%d/ws/vm", c.RestConf.Port)
	return &ServiceContext{
		Config: c,
		Virt:   virt.NewVirt(serverURL, multipass.CertProvider{CertFile: c.MultipassCert, KeyFile: c.MultipassKey}),
		Redis:  redis.MustNewRedis(c.Redis.RedisConf),
	}
}
