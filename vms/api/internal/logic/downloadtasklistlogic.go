package logic

import (
	"context"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/api/ws"

	"github.com/zeromicro/go-zero/core/logx"
)

type DownloadTaskListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDownloadTaskListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DownloadTaskListLogic {
	return &DownloadTaskListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DownloadTaskListLogic) DownloadTaskList(req *types.DownloadTaskListRequest) (resp *types.DownloadTaskListResponse, err error) {
	cmd := ws.NewCmdHandler(l.svcCtx.TunMgr)
	return cmd.DownloadTaskList(l.ctx, req)
}
