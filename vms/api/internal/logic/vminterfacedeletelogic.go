package logic

import (
	"context"
	"fmt"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/vms"

	"github.com/zeromicro/go-zero/core/logx"
)

type VMInterfaceDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVMInterfaceDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VMInterfaceDeleteLogic {
	return &VMInterfaceDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VMInterfaceDeleteLogic) VMInterfaceDelete(req *types.VMInterfaceDeleteRequest) error {
	request := &vms.DeleteNetworkInterfaceRequest{
		Id:     req.Id,
		VmName: req.VmName,
		Mac:    req.Mac,
	}
	rsp, err := l.svcCtx.Vms.DeleteNetworkInterfaceWithLibvirt(l.ctx, request)
	if err != nil {
		return err
	}

	if !rsp.Success {
		return fmt.Errorf(rsp.Message)
	}

	return nil
}
