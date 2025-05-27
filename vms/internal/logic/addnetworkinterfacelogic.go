package logic

import (
	"context"

	"titan-vm/vms/internal/svc"
	"titan-vm/vms/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddNetworkInterfaceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddNetworkInterfaceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddNetworkInterfaceLogic {
	return &AddNetworkInterfaceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AddNetworkInterfaceLogic) AddNetworkInterface(in *pb.AddNetworkInterfaceRequest) (*pb.VMOperationResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.VMOperationResponse{}, nil
}
