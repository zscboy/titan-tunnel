package api

import (
	"titan-vm/vms/api/internal/handler"
	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/internal/config"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, c config.Config) {
	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

}
