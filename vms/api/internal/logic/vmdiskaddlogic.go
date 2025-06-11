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

type VMDiskAddLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVMDiskAddLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VMDiskAddLogic {
	return &VMDiskAddLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VMDiskAddLogic) VMDiskAdd(req *types.VMDiskAddRequest) error {
	request := &vms.AddDiskRequest{
		Id:         req.Id,
		VmName:     req.VmName,
		DiskType:   pb.VMDiskType(req.DiskType),
		SourcePath: req.SourcePath,
		TargetDev:  req.TargetDev,
		TargetBus:  req.TargetBus,
	}
	rsp, err := l.svcCtx.Vms.AddDiskWithLibvirt(l.ctx, request)
	if err != nil {
		return err
	}

	if !rsp.Success {
		return fmt.Errorf(rsp.Message)
	}

	return nil
}
