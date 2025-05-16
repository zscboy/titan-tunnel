package logic

import (
	"context"

	"titan-vm/vms/internal/svc"
	"titan-vm/vms/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetVolLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetVolLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVolLogic {
	return &GetVolLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetVolLogic) GetVol(in *pb.GetVolRequest) (*pb.GetVolResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetVolResponse{}, nil
}
