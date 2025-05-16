package logic

import (
	"context"

	"titan-vm/vms/internal/svc"
	"titan-vm/vms/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type MultipassExecLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMultipassExecLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MultipassExecLogic {
	return &MultipassExecLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *MultipassExecLogic) MultipassExec(in *pb.MultipassExecRequest) (*pb.MultipassExecResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.MultipassExecResponse{}, nil
}
