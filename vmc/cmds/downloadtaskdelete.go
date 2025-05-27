package cmds

import (
	"fmt"
	"titan-vm/pb"
	"titan-vm/vmc/downloader"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type DownloadTaskDelete struct {
	dm *downloader.Manager
}

func NewDownloadTaskDelete(dm *downloader.Manager) *DownloadTaskDelete {
	return &DownloadTaskDelete{dm: dm}
}

func (d *DownloadTaskDelete) DeleteDownloadTask(req []byte) *pb.CmdDownloadTaskDeleteResponse {
	logx.Debug("DeleteDownloadTask")
	err := d.deleteDownloadTask(req)
	if err != nil {
		return &pb.CmdDownloadTaskDeleteResponse{Success: false, Message: err.Error()}
	}
	return &pb.CmdDownloadTaskDeleteResponse{Success: true}
}

func (d *DownloadTaskDelete) deleteDownloadTask(req []byte) error {
	downloadTaskDelete := &pb.CmdDownloadTaskDeleteRequest{}
	err := proto.Unmarshal(req, downloadTaskDelete)
	if err != nil {
		return err
	}

	task := d.dm.GetTask(downloadTaskDelete.GetTaskId())
	if task == nil {
		return fmt.Errorf("task %s not found", downloadTaskDelete.GetTaskId())
	}

	if task.IsRunning() {
		task.Stop()
	}

	return d.dm.DeleteTask(task)
}
