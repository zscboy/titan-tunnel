package logic

import (
	"context"

	"titan-tunnel/manager/internal/svc"
	"titan-tunnel/manager/internal/types"
	"titan-tunnel/server/rpc/serverapi"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPopsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPopsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPopsLogic {
	return &GetPopsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPopsLogic) GetPops() (resp *types.GetPopsResp, err error) {
	pops := make([]*types.Pop, 0, len(l.svcCtx.Config.Pops))
	for id, server := range l.svcCtx.Servers {
		listNodeResp, err := server.API.ListNode(l.ctx, &serverapi.ListNodeReq{PopId: id, Type: 1, Start: 0, End: 1})
		if err != nil {
			return nil, err
		}

		p := &types.Pop{ID: id, Area: server.Area, TotalNode: int(listNodeResp.Total), Socks5Addr: server.Socks5Addr}
		pops = append(pops, p)
	}
	return &types.GetPopsResp{Pops: pops}, nil
}
