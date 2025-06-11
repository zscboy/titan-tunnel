package virt

import (
	"context"
	"titan-vm/vms/pb"
	"titan-vm/vms/virt/golibvirt"
	"titan-vm/vms/virt/multipass"
	multipassPb "titan-vm/vms/virt/multipass/pb"
)

const (
	vmapiMultipass = "multipass"
	vmapiLibvirt   = "libvirt"
)

type VirtInterface interface {
	// Libvirt 相关操作
	CreateVMWithLibvirt(ctx context.Context, in *pb.CreateVMWithLibvirtRequest) error
	CreateVolWithLibvirt(ctx context.Context, in *pb.CreateVolWithLibvirtReqeust) (*pb.CreateVolWithLibvirtResponse, error)
	GetVol(ctx context.Context, in *pb.GetVolRequest) (*pb.GetVolResponse, error)

	// ListHostNetworkInterfaceWithLibvirt(ctx context.Context, in *pb.ListHostNetworkInterfaceRequest) (*pb.ListHostNetworkInterfaceResponse, error)
	// ListVMNetwrokInterfaceWithLibvirt(ctx context.Context, in *pb.ListVMNetwrokInterfaceReqeust) (*pb.ListVMNetworkInterfaceResponse, error)
	AddNetworkInterfaceWithLibvirt(ctx context.Context, in *pb.AddNetworkInterfaceRequest) error
	DeleteNetworkInterfaceWithLibvirt(ctx context.Context, in *pb.DeleteNetworkInterfaceRequest) error

	// ListHostDiskWithLibvirt(ctx context.Context, in *pb.ListHostDiskRequest) (*pb.ListDiskResponse, error)
	// ListVMDiskWithLibvirt(ctx context.Context, in *pb.ListVMDiskRequest) (*pb.ListVMDiskResponse, error)
	AddDiskWithLibvirt(ctx context.Context, in *pb.AddDiskRequest) error
	DeleteDiskWithLibvirt(ctx context.Context, in *pb.DeleteDiskRequest) error
	AddHostdevWithLibvirt(ctx context.Context, in *pb.AddHostdevRequest) error
	DeleteHostdevWithLibvirt(ctx context.Context, in *pb.DeleteHostdevRequest) error
	GetVncPortWithLibvirt(ctx context.Context, in *pb.VMVncPortRequest) (*pb.VMVncPortResponse, error)

	// libvirt 与multipass 通用
	StartVM(ctx context.Context, in *pb.StartVMRequest) error
	StopVM(ctx context.Context, in *pb.StopVMRequest) error
	DeleteVM(ctx context.Context, in *pb.DeleteVMRequest) error
	UpdateVM(ctx context.Context, in *pb.UpdateVMRequest) error
	ListVMInstance(ctx context.Context, in *pb.ListVMInstanceReqeust) (*pb.ListVMInstanceResponse, error)
	ListImage(ctx context.Context, in *pb.ListImageRequest) (*pb.ListImageResponse, error)
	DeleteImage(_ context.Context, request *pb.DeleteImageRequest) error
	GetVMInfo(ctx context.Context, in *pb.GetVMInfoRequest) (*pb.GetVMInfoResponse, error)

	// Multipass 相关操作
	// if return err, will not close progressChan
	CreateVMWithMultipass(ctx context.Context, in *pb.CreateVMWithMultipassRequest, progressChan chan<- *multipassPb.LaunchProgress) error
}

type Virt struct {
	goLibvirt *golibvirt.GoLibvirt
	multipass *multipass.Multipass
}

type VirtOptions struct {
	OS     string
	VMAPI  string
	Online bool
}

func NewVirt(serverURL string, certProvider multipass.CertProvider) *Virt {
	goLibvirt := golibvirt.NewGoLibvirt(serverURL)
	multipass := multipass.NewMultipass(serverURL, certProvider)
	return &Virt{goLibvirt: goLibvirt, multipass: multipass}
}

func (v *Virt) GetVMAPI(opts *VirtOptions) VirtInterface {
	switch opts.VMAPI {
	case vmapiLibvirt:
		return v.goLibvirt
	case vmapiMultipass:
		return v.multipass
	default:
		return nil
	}
}
