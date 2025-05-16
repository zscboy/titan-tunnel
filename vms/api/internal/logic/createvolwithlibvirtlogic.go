package logic

import (
	"context"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/vms"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateVolWithLibvirtLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateVolWithLibvirtLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateVolWithLibvirtLogic {
	return &CreateVolWithLibvirtLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateVolWithLibvirtLogic) CreateVolWithLibvirt(req *types.CreateVolWithLibvirtReqeust) (resp *types.CreateVolWithLibvirtResponse, err error) {
	// todo: add your logic here and delete this line
	request := &vms.CreateVolWithLibvirtReqeust{
		Id:       req.Id,
		Pool:     req.Pool,
		Name:     req.Name,
		Capacity: req.Capacity,
		Format:   req.Format,
	}
	rsp, err := l.svcCtx.Vms.CreateVolWithLibvirt(l.ctx, request)
	if err != nil {
		return nil, err
	}

	return &types.CreateVolWithLibvirtResponse{Pool: rsp.GetPool(), Name: rsp.GetName(), Key: rsp.GetKey()}, nil
}
