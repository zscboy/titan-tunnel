package logic

import (
	"context"

	"titan-tunnel/server/internal/svc"
	"titan-tunnel/server/internal/types"
	"titan-tunnel/server/model"

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
	// todo: add your logic here and delete this line
	pops := make([]types.Pop, 0, len(l.svcCtx.Config.Pops))
	for _, pop := range l.svcCtx.Config.Pops {
		totalNode := l.getTotalNode(pop.ID)
		p := types.Pop{ID: pop.ID, Area: pop.Area, TotalNode: totalNode}
		pops = append(pops, p)
	}
	return &types.GetPopsResp{Pops: pops}, nil
}

func (l *GetPopsLogic) getTotalNode(podID string) int {
	if podID == l.svcCtx.Config.PopID {
		total, err := model.GetNodeLen(l.svcCtx.Redis)
		if err != nil {
			logx.Errorf("GetNodeLen %v", err)
			return 0
		}

		return total
	}

	return 0
}
