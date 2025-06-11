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
	case pb.CommandType_DOWNLOAD_IMAGE:
		return c.downloadImage(msg.GetSessionId(), cmd.GetData())
	case pb.CommandType_DOWNLOAD_TASK_DELETE:
		return c.downloadTaskDelete(msg.GetSessionId(), cmd.GetData())
	case pb.CommandType_DOWNLOAD_TASK_LIST:
		return c.downloadTaskList(msg.GetSessionId())
	case pb.CommandType_AUTH:
		return c.auth(msg.GetSessionId(), cmd.GetData())
	case pb.CommandType_HOST_INFO:
		return c.getHostInfo(msg.GetSessionId())
	case pb.CommandType_LIST_NVME:
		return c.listNvmes(msg.GetSessionId())
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

func (c *Command) auth(sessionID string, reqData []byte) error {
	logx.Debugf("auth sessionID %s", sessionID)
	auth := cmds.NewAuthSSHAndMultipass(c.tunnel.vmapi)
	resp := auth.Auth(reqData)
	return c.cmdReplay(sessionID, resp)
}

func (c *Command) getHostInfo(sessionID string) error {
	logx.Debugf("getHostInfo sessionID %s", sessionID)
	hostInfo := cmds.NewHostInfo()
	resp := hostInfo.Get()
	err := c.cmdReplay(sessionID, resp)
	if err != nil {
		logx.Debugf("getHostInfo cmdReplay failed:%v, %v", err, resp.String())
	}
	return err
}

func (c *Command) listNvmes(sessionID string) error {
	nvmeInfo := cmds.NewNvmeInfo()
	resp := nvmeInfo.List()
	err := c.cmdReplay(sessionID, resp)
	if err != nil {
		logx.Debugf("getHostInfo cmdReplay failed:%v, %v", err, resp.String())
	}
	return err
}
