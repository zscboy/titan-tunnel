package logic

import (
	"context"
	"fmt"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/pb"
	"titan-vm/vms/vms"

	"github.com/zeromicro/go-zero/core/logx"
)

type VMInterfaceAddLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVMInterfaceAddLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VMInterfaceAddLogic {
	return &VMInterfaceAddLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VMInterfaceAddLogic) VMInterfaceAdd(req *types.VMInterfaceAddRequest) error {
	request := &vms.AddNetworkInterfaceRequest{
		Id:              req.Id,
		VmName:          req.VmName,
		Type:            pb.InterfaceType(req.InterfaceType),
		SourceDirectDev: req.SourceDirectDev,
		Model:           pb.InterfaceSourceDirectModel(req.Model),
	}
	rsp, err := l.svcCtx.Vms.AddNetworkInterfaceWithLibvirt(l.ctx, request)
	if err != nil {
		return err
	}

	if !rsp.Success {
		return fmt.Errorf(rsp.Message)
	}

	return nil
}
