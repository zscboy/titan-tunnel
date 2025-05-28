package ws

import (
	"context"
	"fmt"
	"time"
	"titan-vm/pb"
	"titan-vm/vms/api/internal/types"

	"google.golang.org/protobuf/proto"
)

type CmdHandler struct {
	tunMgr *TunnelManager
}

func NewCmdHandler(tunMgr *TunnelManager) *CmdHandler {
	return &CmdHandler{tunMgr: tunMgr}

}

func (cmd *CmdHandler) DownloadImage(ctx context.Context, req *types.DownloadImageRequest) (*types.DownloadImageResponse, error) {
	tun := cmd.tunMgr.getTunnel(req.Id)
	if tun == nil {
		return nil, fmt.Errorf("not found %s", req.Id)
	}

	request := &pb.CmdDownloadImageRequest{Url: req.URL, Md5: req.MD5, Path: req.Path}
	bytes, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	resp := &pb.CmdDownloadImageResponse{}
	payload := &pb.Command{Type: pb.CommandType_DOWNLOAD_IMAGE, Data: bytes}
	err = tun.sendCommand(ctx, payload, resp)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.ErrMsg)
	}

	return &types.DownloadImageResponse{TaskId: resp.TaskId}, nil
}

func (cmd *CmdHandler) DownloadTaskDelete(ctx context.Context, req *types.DownloadTaskDeleteRequest) error {
	tun := cmd.tunMgr.getTunnel(req.Id)
	if tun == nil {
		return fmt.Errorf("not found task %s", req.Id)
	}

	request := &pb.CmdDownloadTaskDeleteRequest{TaskId: req.TaskId}
	bytes, err := proto.Marshal(request)
	if err != nil {
		return err
	}

	resp := &pb.CmdDownloadTaskDeleteResponse{}
	payload := &pb.Command{Type: pb.CommandType_DOWNLOAD_TASK_DELETE, Data: bytes}
	err = tun.sendCommand(ctx, payload, resp)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}

	return nil
}

func (cmd *CmdHandler) DownloadTaskList(tx context.Context, req *types.DownloadTaskListRequest) (*types.DownloadTaskListResponse, error) {
	tun := cmd.tunMgr.getTunnel(req.Id)
	if tun == nil {
		return nil, fmt.Errorf("not found task %s", req.Id)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	downloadTaskResp := &pb.CmdDownloadTaskListResponse{}
	payload := &pb.Command{Type: pb.CommandType_DOWNLOAD_TASK_LIST}
	err := tun.sendCommand(ctx, payload, downloadTaskResp)
	if err != nil {
		return nil, err
	}

	tasks := make([]*types.DownloadTask, 0, len(downloadTaskResp.Tasks))
	for _, downloadTask := range downloadTaskResp.GetTasks() {
		task := &types.DownloadTask{
			TaskId:       downloadTask.Id,
			URL:          downloadTask.Url,
			MD5:          downloadTask.Md5,
			Path:         downloadTask.Path,
			TotalSize:    downloadTask.TotalSize,
			DownloadSize: downloadTask.DownloadSize,
			Running:      downloadTask.Running,
			ErrMsg:       downloadTask.ErrMsg,
			Success:      downloadTask.Success,
		}
		tasks = append(tasks, task)
	}
	return &types.DownloadTaskListResponse{Tasks: tasks}, nil
}

func (cmd *CmdHandler) DownloadTaskGet(ctx context.Context, req *types.DownloadTaskGetRequest) (*types.DownloadTask, error) {
	tun := cmd.tunMgr.getTunnel(req.Id)
	if tun == nil {
		return nil, fmt.Errorf("not found task %s", req.Id)
	}

	request := &pb.CmdDownloadTaskGetRequest{TaskId: req.TaskId}
	bytes, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	downloadTask := &pb.DownloadTask{}
	payload := &pb.Command{Type: pb.CommandType_DOWNLOAD_TASK_GET, Data: bytes}
	err = tun.sendCommand(ctx, payload, downloadTask)
	if err != nil {
		return nil, err
	}

	if !downloadTask.Success {
		return nil, fmt.Errorf(downloadTask.ErrMsg)
	}

	task := &types.DownloadTask{
		TaskId:       downloadTask.Id,
		URL:          downloadTask.Url,
		MD5:          downloadTask.Md5,
		Path:         downloadTask.Path,
		TotalSize:    downloadTask.TotalSize,
		DownloadSize: downloadTask.DownloadSize,
		Running:      downloadTask.Running,
	}
	return task, nil
}
