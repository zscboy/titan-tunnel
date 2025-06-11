package logic

import (
	"context"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/api/ws"

	"github.com/zeromicro/go-zero/core/logx"
)

type DownloadImageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDownloadImageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DownloadImageLogic {
	return &DownloadImageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DownloadImageLogic) DownloadImage(req *types.DownloadImageRequest) (resp *types.DownloadImageResponse, err error) {
	task := ws.NewDownloadtask(l.svcCtx.TunMgr)
	return task.DownloadImage(l.ctx, req)
}
