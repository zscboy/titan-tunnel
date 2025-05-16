package logic

import (
	"context"
	"fmt"

	"titan-vm/vms/internal/svc"
	"titan-vm/vms/pb"
	multipassPb "titan-vm/vms/virt/multipass/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateVMWithMultipassLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateVMWithMultipassLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateVMWithMultipassLogic {
	return &CreateVMWithMultipassLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Multipass 相关操作
func (l *CreateVMWithMultipassLogic) CreateVMWithMultipass(in *pb.CreateVMWithMultipassRequest) (*pb.VMOperationResponse, error) {
	opts, err := getVirtOpts(l.svcCtx.Redis, in.Id)
	if err != nil {
		return nil, err
	}

	vmAPI := l.svcCtx.Virt.GetVMAPI(opts)
	if vmAPI == nil {
		return nil, fmt.Errorf("can not find vm api:%s", opts.VMAPI)
	}

	// progressChan := make(chan *multipassPb.LaunchProgress)
	go func() {
		progressChan := make(chan *multipassPb.LaunchProgress)
		go func() {
			for {
				progress, ok := <-progressChan
				if !ok {
					break
				}

				l.Logger.Infof("progress %s:%s", progress.GetType().String(), progress.GetPercentComplete())
			}

		}()

		err = vmAPI.CreateVMWithMultipass(context.Background(), in, progressChan)
		if err != nil {
			l.Logger.Errorf("CreateVMWithMultipassLogic.CreateVMWithMultipass %s", err.Error())
			return
		}

		l.Logger.Infof("instance %s create complete", in.VmName)

	}()

	return &pb.VMOperationResponse{Success: true}, nil
}
