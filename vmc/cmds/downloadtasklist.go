package cmds

import (
	"titan-vm/pb"
	"titan-vm/vmc/downloader"

	"github.com/zeromicro/go-zero/core/logx"
)

type DownloadTaskList struct {
	dm *downloader.Manager
}

func NewDownloadTaskList(dm *downloader.Manager) *DownloadTaskList {
	return &DownloadTaskList{dm: dm}
}

func (d *DownloadTaskList) ListDownloadTask() *pb.CmdDownloadTaskListResponse {
	logx.Debug("ListDownloadTask")
	return d.listDownloadTask()
}

func (d *DownloadTaskList) listDownloadTask() *pb.CmdDownloadTaskListResponse {
	tasks := d.dm.TasksToProto()
	return &pb.CmdDownloadTaskListResponse{Tasks: tasks}

}
