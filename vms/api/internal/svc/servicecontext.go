package svc

import (
	"titan-vm/vms/internal/config"
	"titan-vm/vms/vms"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	Vms    vms.Vms
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		Vms:    vms.NewVms(zrpc.MustNewClient(c.RpcServer)),
	}
}
