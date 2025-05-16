package logic

import (
	"context"

	"titan-vm/vms/internal/svc"
	"titan-vm/vms/model"
	"titan-vm/vms/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetNodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNodeLogic {
	return &GetNodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetNodeLogic) GetNode(in *pb.GetNodeRequest) (*pb.GetNodeResponse, error) {
	host, err := model.GetNode(l.svcCtx.Redis, in.GetId())
	if err != nil {
		return nil, err
	}

	node := &pb.Node{
		Id:          host.Id,
		Os:          host.OS,
		VmType:      host.VmAPI,
		TotalCpu:    int32(host.CPU),
		TotalMemory: int32(host.Memory),
	}

	return &pb.GetNodeResponse{Node: node}, nil
}
