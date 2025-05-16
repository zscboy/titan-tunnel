package logic

import (
	"context"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/vms"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListImageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListImageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListImageLogic {
	return &ListImageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListImageLogic) ListImage(req *types.ListImageRequest) (resp *types.ListImageResponse, err error) {
	// todo: add your logic here and delete this line
	request := &vms.ListImageRequest{
		Id: req.Id,
	}
	rsp, err := l.svcCtx.Vms.ListImage(l.ctx, request)
	if err != nil {
		return nil, err
	}

	return &types.ListImageResponse{Images: rsp.GetImages()}, nil
}
