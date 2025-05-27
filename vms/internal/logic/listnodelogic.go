package logic

import (
	"context"

	"titan-vm/vms/internal/svc"
	"titan-vm/vms/model"
	"titan-vm/vms/pb"

	"github.com/zeromicro/go-zero/core/logx"
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

// node
func (l *ListNodeLogic) ListNode(in *pb.ListNodeRequest) (*pb.ListNodeResponse, error) {
	modelNodes, err := model.ListNode(l.ctx, l.svcCtx.Redis, int(in.Start), int(in.End))
	if err != nil {
		return nil, err
	}

	nodes := make([]*pb.Node, 0, len(modelNodes))
	for _, modelNode := range modelNodes {
		node := &pb.Node{
			Id:          modelNode.Id,
			Os:          modelNode.OS,
			VmType:      modelNode.VmAPI,
			TotalCpu:    int32(modelNode.CPU),
			TotalMemory: int32(modelNode.Memory),
			Ip:          modelNode.IP,
			Online:      modelNode.Online,
		}
		nodes = append(nodes, node)
	}

	len, err := model.GetNodeLen(l.svcCtx.Redis)
	if err != nil {
		return nil, err
	}

	return &pb.ListNodeResponse{Nodes: nodes, Total: int32(len)}, nil
}
