package logic

import (
	"context"
	"fmt"

	"titan-vm/vms/internal/svc"
	"titan-vm/vms/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	vmapiLibvirt   = "libvirt"
	vmapiMultipass = "multipass"
)

type CreateVMLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateVMLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateVMLogic {
	return &CreateVMLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// common
func (l *CreateVMLogic) CreateVM(in *pb.CreateVMRequest) (*pb.VMOperationResponse, error) {
	opts, err := getVirtOpts(l.svcCtx.Redis, in.Id)
	if err != nil {
		return nil, err
	}

	switch opts.VMAPI {
	case vmapiLibvirt:
		return l.createVMWithLibvirt(in)
	case vmapiMultipass:
		return l.createVMWithMultipass(in)
	default:
		return nil, fmt.Errorf("unsupport vmapi %s", opts.VMAPI)
	}

}

func (l *CreateVMLogic) createVMWithLibvirt(in *pb.CreateVMRequest) (*pb.VMOperationResponse, error) {
	defaultFormat := "qcow2"
	poolName := l.getDefaultPoolOrRandom()
	volLogin := NewCreateVolWithLibvirtLogic(l.ctx, l.svcCtx)
	createVolRequest := &pb.CreateVolWithLibvirtReqeust{
		Id:       in.Id,
		Name:     fmt.Sprintf("%s.qcow2", in.VmName),
		Pool:     poolName,
		Capacity: in.DiskSize,
		Format:   defaultFormat,
	}

	rsp, err := volLogin.CreateVolWithLibvirt(createVolRequest)
	if err != nil {
		return nil, err
	}

	logic := NewCreateVMWithLibvirtLogic(l.ctx, l.svcCtx)
	request := &pb.CreateVMWithLibvirtRequest{
		Id:     in.Id,
		VmName: in.VmName,
		Cpu:    in.Cpu,
		Memory: in.Memory,
		// storageVol key is same as path
		DiskPath: rsp.Key,
		IsoPath:  in.Image,
	}

	return logic.CreateVMWithLibvirt(request)
}

func (l *CreateVMLogic) createVMWithMultipass(in *pb.CreateVMRequest) (*pb.VMOperationResponse, error) {
	logic := NewCreateVMWithMultipassLogic(l.ctx, l.svcCtx)
	request := &pb.CreateVMWithMultipassRequest{
		Id:       in.Id,
		VmName:   in.VmName,
		Cpu:      in.Cpu,
		Memory:   fmt.Sprintf("%dM", in.Memory),
		DiskSize: fmt.Sprintf("%dG", in.DiskSize),
		Image:    in.Image,
	}

	return logic.CreateVMWithMultipass(request)
}

func (l *CreateVMLogic) getDefaultPoolOrRandom() string {
	// todo: list all pool and check default pool if exist
	defaultPool := "images"
	return defaultPool
}
