package ws

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/model"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/ssh"
)

const (
	userName = "root"
)

// WSMessage WebSocket 消息结构
type WSMessage struct {
	// 'error', 'stdin', 'stdout', 'resize'
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Cols uint   `json:"cols,omitempty"`
	Rows uint   `json:"rows,omitempty"`
}

type WSReq struct {
	Id   string `json:"id"`
	Type string `json:"type"`
	Addr string `json:"addr"`
}

type SSHWS struct {
	tunMgr *TunnelManager
}

func NewSSHWS(tunMgr *TunnelManager) *SSHWS {
	return &SSHWS{tunMgr: tunMgr}
}

// todo support multipass and auth client
func (ws *SSHWS) ServeWS(w http.ResponseWriter, r *http.Request, req *types.SSHWSReqeust) error {
	logx.Debugf("sshHandler")
	if len(req.NodeId) == 0 {
		return fmt.Errorf("request NodeId")
	}

	node, err := model.GetNode(ws.tunMgr.redis, req.NodeId)
	if err != nil {
		return err
	}

	if !node.Online {
		return fmt.Errorf("node %s offline", req.NodeId)
	}

	if node.OS != "linux" {
		return fmt.Errorf("ssh only support on linux")
	}

	privKeyBase64, err := os.ReadFile(ws.tunMgr.config.SSHPriKey)
	if err != nil {
		return err
	}

	signer, err := ssh.ParsePrivateKey(privKeyBase64)
	if err != nil {
		return fmt.Errorf("parse private key error: %s", err.Error())
	}

	sshPort := 22
	if node.SSHPort != 0 {
		sshPort = node.SSHPort
	}

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer wsConn.Close()

	sshConfig := &ssh.ClientConfig{
		User: userName,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer), // 使用私钥认证
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	ctx, cancle := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancle()

	url := fmt.Sprintf("ws://localhost:%d/ws/vm?id=%s&transport=raw&address=localhost:%d", ws.tunMgr.config.RestConf.Port, req.NodeId, sshPort)
	websocketConn, resp, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {
		var msg []byte
		if resp != nil {
			msg, _ = io.ReadAll(resp.Body)
		}
		logx.Errorf("websocket dial error %s, msg:%s", err.Error(), string(msg))
		wsConn.WriteJSON(WSMessage{Type: "error", Data: "websocket conn error " + err.Error()})
		return nil
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(websocketConn.NetConn(), fmt.Sprintf("%s:%d", "localhost", sshPort), sshConfig) // 替换为服务器IP和端口
	if err != nil {
		logx.Errorf("NewClientConn error %s", err.Error())
		wsConn.WriteJSON(WSMessage{Type: "error", Data: "NewClientConn error " + err.Error()})
		return nil
	}
	defer sshConn.Close()

	sshClient := ssh.NewClient(sshConn, chans, reqs)
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		log.Println("SSH session error:", err)
		wsConn.WriteJSON(WSMessage{Type: "error", Data: "SSH session failed"})
		return nil
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // open echo
		ssh.TTY_OP_ISPEED: 14400, // input rate
		ssh.TTY_OP_OSPEED: 14400, // output rate
	}
	if err := session.RequestPty("xterm-256color", 40, 80, modes); err != nil {
		wsConn.WriteJSON(WSMessage{Type: "error", Data: "Request PTY failed"})
		return nil
	}

	sshOut, err := session.StdoutPipe()
	if err != nil {
		wsConn.WriteJSON(WSMessage{Type: "error", Data: err.Error()})
		return nil
	}
	sshIn, err := session.StdinPipe()
	if err != nil {
		wsConn.WriteJSON(WSMessage{Type: "error", Data: err.Error()})
		return nil
	}

	if err := session.Shell(); err != nil {
		wsConn.WriteJSON(WSMessage{Type: "error", Data: err.Error()})
		return nil
	}

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := sshOut.Read(buf)
			if err != nil {
				if err != io.EOF {
					logx.Error("sshOut.Read error:", err)
				}
				break
			}
			msg := WSMessage{Type: "stdout", Data: string(buf[:n])}
			wsConn.WriteJSON(msg)
		}
		wsConn.Close()
	}()

	// 前端消息驱动 SSH 输入或 resize
	for {
		var msg WSMessage
		if err := wsConn.ReadJSON(&msg); err != nil {
			log.Println("ReadJSON error:", err)
			break
		}
		switch msg.Type {
		case "stdin":
			sshIn.Write([]byte(msg.Data))
		case "resize":
			// change windows size
			logx.Infof("ssh resize WindowChange:%d, %d", msg.Rows, msg.Cols)
			session.WindowChange(int(msg.Rows), int(msg.Cols))
		}
	}

	return nil
}
