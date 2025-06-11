package logic

import (
	"context"

	"titan-vm/vms/internal/svc"
	"titan-vm/vms/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetVncPortWithLibvirtLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetVncPortWithLibvirtLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVncPortWithLibvirtLogic {
	return &GetVncPortWithLibvirtLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetVncPortWithLibvirtLogic) GetVncPortWithLibvirt(in *pb.VMVncPortRequest) (*pb.VMVncPortResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.VMVncPortResponse{}, nil
}
