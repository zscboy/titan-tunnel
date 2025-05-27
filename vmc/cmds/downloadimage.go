package cmds

import (
	"titan-vm/pb"
	"titan-vm/vmc/downloader"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type DownloadImage struct {
	dm *downloader.Manager
}

func NewDownloadImage(dm *downloader.Manager) *DownloadImage {
	return &DownloadImage{dm: dm}
}

func (d *DownloadImage) DownloadImage(req []byte) *pb.CmdDownloadImageResponse {
	logx.Debug("DownloadImage")
	return d.downloadImage(req)
}

func (d *DownloadImage) downloadImage(req []byte) *pb.CmdDownloadImageResponse {
	downloadImageRequest := &pb.CmdDownloadImageRequest{}
	err := proto.Unmarshal(req, downloadImageRequest)
	if err != nil {
		return &pb.CmdDownloadImageResponse{ErrMsg: err.Error()}
	}

	if len(downloadImageRequest.Url) == 0 {
		return &pb.CmdDownloadImageResponse{ErrMsg: "url can not emtpy"}
	}

	if len(downloadImageRequest.Md5) == 0 {
		return &pb.CmdDownloadImageResponse{ErrMsg: "md5 can not emtpy"}
	}

	if len(downloadImageRequest.Path) == 0 {
		return &pb.CmdDownloadImageResponse{ErrMsg: "path can not emtpy"}
	}

	opts := downloader.TaskOptions{
		Id:   uuid.NewString(),
		URL:  downloadImageRequest.Url,
		MD5:  downloadImageRequest.Md5,
		Path: downloadImageRequest.Path,
	}
	task := downloader.NewTask(&opts)
	if err := d.dm.AddTask(task); err != nil {
		return &pb.CmdDownloadImageResponse{ErrMsg: err.Error()}
	}

	if err = task.Start(); err != nil {
		return &pb.CmdDownloadImageResponse{ErrMsg: err.Error()}
	}

	return &pb.CmdDownloadImageResponse{Success: true, TaskId: task.GetId()}
}
