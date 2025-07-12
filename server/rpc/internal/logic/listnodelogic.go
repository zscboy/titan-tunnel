package logic

import (
	"context"
	"fmt"

	"titan-tunnel/server/api/model"
	"titan-tunnel/server/rpc/internal/svc"
	"titan-tunnel/server/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	reqNodeType = iota
	reqNodeTypeAll
	reqNodeTypeUnbind
	reqNodeTypeBind
)

type ListNodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListNodeLogic {
	return &ListNodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListNodeLogic) ListNode(in *pb.ListNodeReq) (*pb.ListNodeResp, error) {
	if in.Type != reqNodeTypeAll && in.Type != reqNodeTypeUnbind && in.Type != reqNodeTypeBind {
		return nil, fmt.Errorf("request type [%d] not match", in.Type)
	}

	switch in.Type {
	case reqNodeTypeAll:
		return l.listNodeWithAll(in)
	case reqNodeTypeUnbind:
		return l.listNodeWithUnbinde(in)
	case reqNodeTypeBind:
		return l.listNodeWithBind(in)
	}
	return nil, fmt.Errorf("unsupport request type %d", in.Type)

}

func (l *ListNodeLogic) listNodeWithAll(req *pb.ListNodeReq) (resp *pb.ListNodeResp, err error) {
	nodes, err := model.ListNode(l.ctx, l.svcCtx.Redis, int(req.Start), int(req.End))
	if err != nil {
		return nil, err
	}

	ns := make([]*pb.Node, 0, len(nodes))
	for _, node := range nodes {
		n := &pb.Node{Id: node.Id, Ip: node.IP, BindUser: node.BindUser, Online: node.Online}
		ns = append(ns, n)
	}

	total, err := model.GetNodeLen(l.svcCtx.Redis)
	if err != nil {
		return nil, err
	}

	return &pb.ListNodeResp{Nodes: ns, Total: int32(total)}, nil
}

func (l *ListNodeLogic) listNodeWithUnbinde(req *pb.ListNodeReq) (resp *pb.ListNodeResp, err error) {
	nodes, err := model.ListUnbindNode(l.ctx, l.svcCtx.Redis, int(req.Start), int(req.End))
	if err != nil {
		return nil, err
	}

	ns := make([]*pb.Node, 0, len(nodes))
	for _, node := range nodes {
		n := &pb.Node{Id: node.Id, Ip: node.IP, BindUser: node.BindUser, Online: node.Online}
		ns = append(ns, n)
	}

	total, err := model.GetUnbindNodeLen(l.svcCtx.Redis)
	if err != nil {
		return nil, err
	}

	return &pb.ListNodeResp{Nodes: ns, Total: int32(total)}, nil
}

func (l *ListNodeLogic) listNodeWithBind(req *pb.ListNodeReq) (resp *pb.ListNodeResp, err error) {
	nodes, err := model.ListBindNode(l.ctx, l.svcCtx.Redis, int(req.Start), int(req.End))
	if err != nil {
		return nil, err
	}

	ns := make([]*pb.Node, 0, len(nodes))
	for _, node := range nodes {
		n := &pb.Node{Id: node.Id, Ip: node.IP, BindUser: node.BindUser, Online: node.Online}
		ns = append(ns, n)
	}

	total, err := model.GetbindNodeLen(l.svcCtx.Redis)
	if err != nil {
		return nil, err
	}

	return &pb.ListNodeResp{Nodes: ns, Total: int32(total)}, nil
}
