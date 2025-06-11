package logic

import (
	"context"
	"fmt"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/vms"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteImageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteImageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteImageLogic {
	return &DeleteImageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteImageLogic) DeleteImage(req *types.DeleteImageRequest) error {
	request := &vms.DeleteImageRequest{
		Id:   req.Id,
		Path: req.Path,
	}
	rsp, err := l.svcCtx.Vms.DeleteImage(l.ctx, request)
	if err != nil {
		return err
	}

	if !rsp.Success {
		return fmt.Errorf(rsp.Message)
	}

	return nil
}
