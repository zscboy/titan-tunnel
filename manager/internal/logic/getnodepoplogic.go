package logic

import (
	"context"
	"fmt"

	"titan-tunnel/manager/internal/svc"
	"titan-tunnel/manager/internal/types"
	"titan-tunnel/server/rpc/serverapi"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetNodePopLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetNodePopLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNodePopLogic {
	return &GetNodePopLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNodePopLogic) GetNodePop(req *types.GetNodePopReq) (resp *types.GetNodePopResp, err error) {
	pop := l.getNodePop(req)
	if pop == nil {
		return nil, fmt.Errorf("not found pop for node %s", req.NodeId)
	}

	getTokenResp, err := pop.API.GetNodeAccessToken(l.ctx, &serverapi.GetNodeAccessTokenReq{NodeId: req.NodeId})
	if err != nil {
		return nil, err
	}

	return &types.GetNodePopResp{ServerURL: pop.WSServerURL, AccessToken: getTokenResp.Token}, nil
}

func (l *GetNodePopLogic) getNodePop(_ *types.GetNodePopReq) *svc.Server {
	for _, pop := range l.svcCtx.Servers {
		return pop
	}
	return nil
}
