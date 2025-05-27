package logic

import (
	"context"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/vms"

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

func (l *ListNodeLogic) ListNode(req *types.ListNodeReqeust) (resp *types.ListNodeResponse, err error) {
	request := &vms.ListNodeRequest{
		Start: int32(req.Start),
		End:   int32(req.End),
	}
	rsp, err := l.svcCtx.Vms.ListNode(l.ctx, request)
	if err != nil {
		return nil, err
	}

	nodes := make([]*types.Node, 0, len(rsp.GetNodes()))
	for _, pbNode := range rsp.GetNodes() {
		node := &types.Node{
			Id:             pbNode.GetId(),
			OS:             pbNode.GetOs(),
			VmType:         pbNode.GetVmType(),
			TotalCpu:       int(pbNode.GetTotalCpu()),
			TotalMemory:    int(pbNode.GetTotalMemory()),
			SystemDiskSize: int(pbNode.GetSystemDiskSize()),
			IP:             pbNode.GetIp(),
			Online:         pbNode.Online,
		}
		nodes = append(nodes, node)
	}

	return &types.ListNodeResponse{Nodes: nodes}, nil
}
