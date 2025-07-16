package logic

import (
	"context"
	"fmt"
	"net"

	"titan-tunnel/server/api/internal/svc"
	"titan-tunnel/server/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetServerInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetServerInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetServerInfoLogic {
	return &GetServerInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetServerInfoLogic) GetServerInfo() (resp *types.ServerInfo, err error) {
	// todo: add your logic here and delete this line
	_, port, err := net.SplitHostPort(l.svcCtx.Config.Socks5.Addr)
	if err != nil {
		return nil, err
	}
	socks5Addr := fmt.Sprintf("%s:%s", l.svcCtx.Config.Socks5.ServerIP, port)

	domain := l.svcCtx.Config.Socks5.ServerIP
	if len(l.svcCtx.Config.Domain) > 0 {
		domain = l.svcCtx.Config.Domain
	}
	wsServerURl := fmt.Sprintf("ws://%s:%d/ws/node", domain, l.svcCtx.Config.Port)

	return &types.ServerInfo{Socks5Addr: socks5Addr, WSServerURL: wsServerURl}, nil
}
