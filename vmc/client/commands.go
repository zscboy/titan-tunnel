package client

import (
	"titan-vm/pb"
	"titan-vm/vmc/cmds"
	"titan-vm/vmc/downloader"

	log "github.com/sirupsen/logrus"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type Command struct {
	tunnel          *Tunnel
	downloadManager *downloader.Manager
}

func (c *Command) onCommandMessage(msg *pb.Message) error {
	cmd := &pb.Command{}
	err := proto.Unmarshal(msg.GetPayload(), cmd)
	if err != nil {
		return err
	}

	log.Debugf("Tunnel.onControlMessage, cmd type:%s", cmd.GetType().String())

	switch cmd.GetType() {
	case pb.CommandType_DownloadImage:
		return c.downloadImage(msg.GetSessionId(), cmd.GetData())
	case pb.CommandType_DownloadTaskDelete:
		return c.downloadTaskDelete(msg.GetSessionId(), cmd.GetData())
	case pb.CommandType_DownloadTaskList:
		return c.downloadTaskList(msg.GetSessionId())
	}
	return nil
}

func (c *Command) cmdReplay(sessionID string, cmdReplay proto.Message) error {
	bytes, err := proto.Marshal(cmdReplay)
	if err != nil {
		return err
	}

	msg := &pb.Message{Type: pb.MessageType_COMMAND, SessionId: sessionID, Payload: bytes}
	bytes, err = proto.Marshal(msg)
	if err != nil {
		return err
	}

	return c.tunnel.write(bytes)
}

// func (c *Command) exec(sessionID string, cmdPyaload []byte) error {
// 	return nil
// }

func (c *Command) downloadImage(sessionID string, reqData []byte) error {
	logx.Debugf("downloadImage sessionID %s", sessionID)

	downloadImage := cmds.NewDownloadImage(c.downloadManager)
	resp := downloadImage.DownloadImage(reqData)
	return c.cmdReplay(sessionID, resp)
}

func (c *Command) downloadTaskDelete(sessionID string, reqData []byte) error {
	logx.Debugf("downloadTaskDelete sessionID %s", sessionID)

	downloadTaskDelete := cmds.NewDownloadTaskDelete(c.downloadManager)
	resp := downloadTaskDelete.DeleteDownloadTask(reqData)
	return c.cmdReplay(sessionID, resp)
}

func (c *Command) downloadTaskList(sessionID string) error {
	logx.Debugf("listDownloadTask sessionID %s", sessionID)

	downloadTaskList := cmds.NewDownloadTaskList(c.downloadManager)
	resp := downloadTaskList.ListDownloadTask()
	return c.cmdReplay(sessionID, resp)
}
