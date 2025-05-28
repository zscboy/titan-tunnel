package api

import (
	"titan-vm/vms/api/internal/handler"
	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/internal/config"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, c config.Config) {
	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

}

type CreateVMRequest types.CreateVMRequest
type CreateVolWithLibvirtReqeust types.CreateVolWithLibvirtReqeust
type CreateVolWithLibvirtResponse types.CreateVolWithLibvirtResponse
type DeleteVMRequest types.DeleteVMRequest
type ListImageRequest types.ListImageRequest
type ListImageResponse types.ListImageResponse
type ListVMInstanceReqeust types.ListVMInstanceReqeust
type ListVMInstanceResponse types.ListVMInstanceResponse
type StartVMRequest types.StartVMRequest
type StopVMRequest types.StopVMRequest
type UpdateVMRequest types.UpdateVMRequest
type VMInfo types.VMInfo
type VMOperationResponse types.VMOperationResponse
type MultipassExecRequest types.MultipassExecRequest
type MultipassExecResponse types.MultipassExecRequest
type ListNodeReqeust types.ListNodeReqeust
type ListNodeResponse types.ListNodeResponse
type Node types.Node

type DownloadImageRequest types.DownloadImageRequest
type DownloadImageResponse types.DownloadImageResponse
type DownloadTaskDeleteRequest types.DownloadTaskDeleteRequest
type DownloadTaskDeleteResponse types.DownloadTaskDeleteResponse
type DownloadTaskListRequest types.DownloadTaskListRequest
type DownloadTask types.DownloadTask
type DownloadTaskListResponse types.DownloadTaskListResponse
type DownloadTaskGetRequest types.DownloadTaskGetRequest

type NodeWSRequest types.NodeWSRequest
type VMWSRequest types.VMWSRequest
type SSHWSReqeust types.SSHWSReqeust
type SSHWSMessage types.SSHWSMessage
