package logic

import (
	"context"
	"fmt"

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
	for _, pop := range l.svcCtx.Config.Pops {
		totalNode, err := l.getTotalNode(pop.ID)
		if err != nil {
			return nil, err
		}

		p := &types.Pop{ID: pop.ID, Area: pop.Area, TotalNode: totalNode, Socks5Addr: pop.Socks5Addr}
		pops = append(pops, p)
	}
	return &types.GetPopsResp{Pops: pops}, nil
}

func (l *GetPopsLogic) getTotalNode(popID string) (int, error) {
	api := l.svcCtx.ServerAPIs[popID]
	if api == nil {
		return 0, fmt.Errorf("pop %s not found", popID)
	}

	listNodeResp, err := api.ListNode(l.ctx, &serverapi.ListNodeReq{PopId: popID, Type: 1, Start: 0, End: 1})
	if err != nil {
		return 0, err
	}

	return int(listNodeResp.Total), nil
}
