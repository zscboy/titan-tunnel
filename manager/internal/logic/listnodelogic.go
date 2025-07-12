package logic

import (
	"context"
	"fmt"

	"titan-tunnel/manager/internal/svc"
	"titan-tunnel/manager/internal/types"
	"titan-tunnel/server/rpc/serverapi"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListNodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListNodeLogic {
	return &ListNodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListNodeLogic) ListNode(req *types.ListNodeReq) (resp *types.ListNodeResp, err error) {
	api := l.svcCtx.ServerAPIs[req.PopID]
	if api == nil {
		return nil, fmt.Errorf("pop %s not found", req.PopID)
	}

	in := &serverapi.ListNodeReq{
		PopId: req.PopID,
		Type:  int32(req.Type),
		Start: int32(req.Start),
		End:   int32(req.End),
	}

	listNodeResp, err := api.ListNode(l.ctx, in)
	if err != nil {
		return nil, err
	}

	nodes := make([]*types.Node, 0, len(listNodeResp.Nodes))
	for _, node := range listNodeResp.Nodes {
		nodes = append(nodes, toNodeResp(node))
	}

	return &types.ListNodeResp{Nodes: nodes, Total: int(listNodeResp.Total)}, nil
}
