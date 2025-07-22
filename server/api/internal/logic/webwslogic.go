package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"titan-tunnel/server/api/internal/svc"
)

type WebWSLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWebWSLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WebWSLogic {
	return &WebWSLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WebWSLogic) WebWS() error {
	// todo: add your logic here and delete this line

	return nil
}
