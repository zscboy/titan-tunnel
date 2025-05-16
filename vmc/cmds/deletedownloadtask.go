package cmds

import (
	"titan-vm/pb"
	"titan-vm/vmc/downloader"
)

type DeleteDownloadTask struct {
	downloadManager *downloader.Manager
}

func NewDeleteDownloadTask() *DeleteDownloadTask {
	return &DeleteDownloadTask{}
}

func (d *DownloadImage) DeleteDownloadTask(req []byte) *pb.CmdDownloadTaskControlResponse {
	// 	logx.Debug("DownloadImage")
	// 	err := d.deleteDownloadTask(req)
	// 	if err != nil {
	// 		return &pb.CmdDownloadImageResponse{Success: false, Message: err.Error()}
	// 	}
	// 	return &pb.CmdDownloadImageResponse{Success: true}
	return nil
}

func (d *DownloadImage) deleteDownloadTask(req []byte) error {
	// downloadImageRequest := &pb.CmdDownloadImageRequest{}
	// err := proto.Unmarshal(req, downloadImageRequest)
	// if err != nil {
	// 	return err
	// }

	// opts := downloader.TaskOptions{}
	// task := downloader.NewTask(&opts)
	// if err := d.dm.AddTask(task); err != nil {
	// 	return err
	// }

	// return task.Start()
	return nil
}
