package logic

import (
	"context"
	"fmt"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/vms"

	"github.com/zeromicro/go-zero/core/logx"
)

type VMHosthostAddLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVMHosthostAddLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VMHosthostAddLogic {
	return &VMHosthostAddLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VMHosthostAddLogic) VMHosthostAdd(req *types.VMHostdevAddRequest) error {
	request := &vms.AddHostdevRequest{
		Id:               req.Id,
		VmName:           req.VmName,
		SourceAddrDomain: int32(req.SourceAddrDomain),
		SourceAddrBus:    int32(req.SourceAddrBus),
		SourceAddrSlot:   int32(req.SourceAddrSlot),
	}
	rsp, err := l.svcCtx.Vms.AddHostdevWithLibvirt(l.ctx, request)
	if err != nil {
		return err
	}

	if !rsp.Success {
		return fmt.Errorf(rsp.Message)
	}

	return nil
}
