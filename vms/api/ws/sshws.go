package ws

import (
	"io"
	"log"
	"net/http"
	"titan-vm/vms/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/ssh"
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

func (ws *SSHWS) ServeWS(w http.ResponseWriter, r *http.Request, req *types.SSHWSReqeust) {
	logx.Debugf("sshHandler")
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer wsConn.Close()

	sshConfig := &ssh.ClientConfig{
		User:            "root",
		Auth:            []ssh.AuthMethod{ssh.Password("123")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	sshClient, err := ssh.Dial("tcp", "192.168.0.132:22", sshConfig)
	if err != nil {
		log.Println("SSH dial error:", err)
		wsConn.WriteJSON(WSMessage{Type: "error", Data: "SSH connection failed"})
		return
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		log.Println("SSH session error:", err)
		wsConn.WriteJSON(WSMessage{Type: "error", Data: "SSH session failed"})
		return
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // open echo
		ssh.TTY_OP_ISPEED: 14400, // input rate
		ssh.TTY_OP_OSPEED: 14400, // output rate
	}
	if err := session.RequestPty("xterm-256color", 40, 80, modes); err != nil {
		log.Println("RequestPty error:", err)
		wsConn.WriteJSON(WSMessage{Type: "error", Data: "Request PTY failed"})
		return
	}

	sshOut, err := session.StdoutPipe()
	if err != nil {
		log.Println("StdoutPipe error:", err)
		return
	}
	sshIn, err := session.StdinPipe()
	if err != nil {
		log.Println("StdinPipe error:", err)
		return
	}

	if err := session.Shell(); err != nil {
		log.Println("session.Shell error:", err)
		return
	}

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := sshOut.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Println("sshOut.Read error:", err)
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
}
