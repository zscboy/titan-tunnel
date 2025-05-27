package logic

import (
	"context"

	"titan-vm/vms/internal/svc"
	"titan-vm/vms/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListHostDiskWithLibvirtLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListHostDiskWithLibvirtLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListHostDiskWithLibvirtLogic {
	return &ListHostDiskWithLibvirtLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListHostDiskWithLibvirtLogic) ListHostDiskWithLibvirt(in *pb.ListHostDiskRequest) (*pb.ListDiskResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.ListDiskResponse{}, nil
}
