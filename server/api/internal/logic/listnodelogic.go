package logic

import (
	"context"
	"fmt"

	"titan-tunnel/server/api/internal/svc"
	"titan-tunnel/server/api/internal/types"
	"titan-tunnel/server/api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	reqNodeType = iota
	reqNodeTypeAll
	reqNodeTypeUnbind
	reqNodeTypeBind
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
	if req.Type != reqNodeTypeAll && req.Type != reqNodeTypeUnbind && req.Type != reqNodeTypeBind {
		return nil, fmt.Errorf("request type [%d] not match", req.Type)
	}

	switch req.Type {
	case reqNodeTypeAll:
		return l.listNodeWithAll(req)
	case reqNodeTypeUnbind:
		return l.listNodeWithUnbinde(req)
	case reqNodeTypeBind:
		return l.listNodeWithBind(req)
	}
	return nil, fmt.Errorf("unsupport request type %d", req.Type)
}

func (l *ListNodeLogic) listNodeWithAll(req *types.ListNodeReq) (resp *types.ListNodeResp, err error) {
	nodes, err := model.ListNode(l.ctx, l.svcCtx.Redis, req.Start, req.End)
	if err != nil {
		return nil, err
	}

	ns := make([]types.Node, 0, len(nodes))
	for _, node := range nodes {
		n := types.Node{Id: node.Id, IP: node.IP, BindUser: node.BindUser, Online: node.Online}
		ns = append(ns, n)
	}

	total, err := model.GetNodeLen(l.svcCtx.Redis)
	if err != nil {
		return nil, err
	}

	return &types.ListNodeResp{Nodes: ns, Total: total}, nil
}

func (l *ListNodeLogic) listNodeWithUnbinde(req *types.ListNodeReq) (resp *types.ListNodeResp, err error) {
	nodes, err := model.ListUnbindNode(l.ctx, l.svcCtx.Redis, req.Start, req.End)
	if err != nil {
		return nil, err
	}

	ns := make([]types.Node, 0, len(nodes))
	for _, node := range nodes {
		n := types.Node{Id: node.Id, IP: node.IP, BindUser: node.BindUser, Online: node.Online}
		ns = append(ns, n)
	}

	total, err := model.GetUnbindNodeLen(l.svcCtx.Redis)
	if err != nil {
		return nil, err
	}

	return &types.ListNodeResp{Nodes: ns, Total: total}, nil
}

func (l *ListNodeLogic) listNodeWithBind(req *types.ListNodeReq) (resp *types.ListNodeResp, err error) {
	nodes, err := model.ListBindNode(l.ctx, l.svcCtx.Redis, req.Start, req.End)
	if err != nil {
		return nil, err
	}

	ns := make([]types.Node, 0, len(nodes))
	for _, node := range nodes {
		n := types.Node{Id: node.Id, IP: node.IP, BindUser: node.BindUser, Online: node.Online}
		ns = append(ns, n)
	}

	total, err := model.GetbindNodeLen(l.svcCtx.Redis)
	if err != nil {
		return nil, err
	}

	return &types.ListNodeResp{Nodes: ns, Total: total}, nil
}
