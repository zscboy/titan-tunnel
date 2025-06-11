package logic

import (
	"context"
	"fmt"

	"titan-vm/vms/internal/svc"
	"titan-vm/vms/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetVMInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetVMInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVMInfoLogic {
	return &GetVMInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetVMInfoLogic) GetVMInfo(in *pb.GetVMInfoRequest) (*pb.GetVMInfoResponse, error) {
	opts, err := getVirtOpts(l.svcCtx.Redis, in.Id)
	if err != nil {
		return nil, err
	}

	vmAPI := l.svcCtx.Virt.GetVMAPI(opts)
	if vmAPI == nil {
		return nil, fmt.Errorf("can not find vm api:%s", opts.VMAPI)
	}

	return vmAPI.GetVMInfo(l.ctx, in)
}
