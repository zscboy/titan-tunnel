package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"titan-vm/vms/api/internal/svc"
)

type SshWSLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSshWSLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SshWSLogic {
	return &SshWSLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SshWSLogic) SshWS() error {
	// todo: add your logic here and delete this line

	return nil
}
