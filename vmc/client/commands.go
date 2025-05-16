package client

import (
	"fmt"
	"titan-vm/pb"
	"titan-vm/vmc/cmds"
	"titan-vm/vmc/downloader"

	"github.com/golang/protobuf/proto"
	"github.com/zeromicro/go-zero/core/logx"
)

type Command struct {
	tunnel          *Tunnel
	downloadManager *downloader.Manager
}

func (c *Command) cmdReplay(sessionID string, cmdReplay proto.Message) error {
	bytes, err := proto.Marshal(cmdReplay)
	if err != nil {
		return err
	}

	msg := &pb.Message{Type: pb.MessageType_CONTROL, SessionId: sessionID, Payload: bytes}
	bytes, err = proto.Marshal(msg)
	if err != nil {
		return err
	}

	return c.tunnel.write(bytes)
}

func (c *Command) exec(sessionID string, cmdPyaload []byte) error {
	return nil
}

func (c *Command) downloadImage(sessionID string, reqData []byte) error {
	logx.Debugf("downloadImage sessionID %s", sessionID)

	downloadImage := cmds.NewDownloadImage(c.downloadManager)
	resp := downloadImage.DownloadImage(reqData)
	return c.cmdReplay(sessionID, resp)
}

func (c *Command) downloadTaskControl(sessionID string, reqData []byte) error {
	logx.Debugf("downloadImage sessionID %s", sessionID)

	taskControl := &pb.DownloadTaskControlRequest{}
	err := proto.Unmarshal(reqData, taskControl)
	if err != nil {
		return c.cmdReplay(sessionID, &pb.CmdDownloadTaskControlResponse{Success: false, Message: err.Error()})
	}

	switch taskControl.Action {
	case pb.DownloadTaskAction_START:
	case pb.DownloadTaskAction_STOP:
	case pb.DownloadTaskAction_DELETE:
	default:
		return c.cmdReplay(sessionID, &pb.CmdDownloadTaskControlResponse{Success: false, Message: fmt.Sprintf("unsupport action %s", taskControl.Action)})
	}
	return nil
}
