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

type VMDiskDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVMDiskDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VMDiskDeleteLogic {
	return &VMDiskDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VMDiskDeleteLogic) VMDiskDelete(req *types.VMDiskDeleteRequest) error {
	request := &vms.DeleteDiskRequest{
		Id:               req.Id,
		VmName:           req.VmName,
		DiskType:         pb.VMDiskType(req.DiskType),
		SourcePciAddrBus: int32(req.SourcePciAddrBus),
		TargetDev:        req.TargetDev,
	}
	rsp, err := l.svcCtx.Vms.DeleteDiskWithLibvirt(l.ctx, request)
	if err != nil {
		return err
	}

	if !rsp.Success {
		return fmt.Errorf(rsp.Message)
	}

	return nil
}
