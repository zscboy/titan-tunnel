package cmds

import (
	"fmt"
	"titan-vm/pb"
	"titan-vm/vmc/downloader"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type DownloadTaskGet struct {
	dm *downloader.Manager
}

func NewDownloadTaskGet(dm *downloader.Manager) *DownloadTaskGet {
	return &DownloadTaskGet{dm: dm}
}

func (d *DownloadTaskGet) GetDownloadTask(req []byte) *pb.CmdDownloadTaskGetResponse {
	logx.Debug("ListDownloadTask")
	return d.getDownloadTask(req)
}

func (d *DownloadTaskGet) getDownloadTask(req []byte) *pb.CmdDownloadTaskGetResponse {
	downloadTaskGet := &pb.CmdDownloadTaskGetRequest{}
	err := proto.Unmarshal(req, downloadTaskGet)
	if err != nil {
		return &pb.CmdDownloadTaskGetResponse{ErrMsg: err.Error()}
	}

	task := d.dm.GetTask(downloadTaskGet.GetTaskId())
	if task == nil {
		return &pb.CmdDownloadTaskGetResponse{ErrMsg: fmt.Sprintf("not found task %s", downloadTaskGet.GetTaskId())}
	}

	return &pb.CmdDownloadTaskGetResponse{Success: true, Task: d.dm.TaskToProto(task)}

}
