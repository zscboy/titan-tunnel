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

func (d *DownloadImage) DownloadImage(req []byte) *pb.CmdDownloadTaskControlResponse {
	logx.Debug("DownloadImage")
	err := d.downloadImage(req)
	if err != nil {
		return &pb.CmdDownloadTaskControlResponse{Success: false, Message: err.Error()}
	}
	return &pb.CmdDownloadTaskControlResponse{Success: true}
}

func (d *DownloadImage) downloadImage(req []byte) error {
	downloadImageRequest := &pb.CmdDownloadImageRequest{}
	err := proto.Unmarshal(req, downloadImageRequest)
	if err != nil {
		return err
	}

	opts := downloader.TaskOptions{
		Id:   uuid.NewString(),
		URL:  downloadImageRequest.Url,
		MD5:  downloadImageRequest.Md5,
		Path: downloadImageRequest.Path,
	}
	task := downloader.NewTask(&opts)
	if err := d.dm.AddTask(task); err != nil {
		return err
	}

	return task.Start()
}
