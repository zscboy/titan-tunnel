package cmds

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	"github.com/urfave/cli/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

const (
	defaultKeyFile = "id_rsa"
)

var SshCmd = &cli.Command{
	Name:  "ssh",
	Usage: "vmadm ssh {node-id}",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "key",
			Usage: "--key",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "user",
			Usage: "--user",
			Value: "root",
		},
		&cli.IntFlag{
			Name:  "remote-ssh-port",
			Usage: "--remote-ssh-port",
			Value: 22,
		},
	},
	Action: func(cctx *cli.Context) error {
		server := cctx.String("server")
		keyPath := cctx.String("key")
		userName := cctx.String("user")
		sshPort := cctx.Int("remote-ssh-port")
		id := cctx.Args().Get(0)
		if len(id) == 0 {
			return fmt.Errorf("need node id, example: vmadm ssh {node-id}")
		}

		if sshPort == 0 {
			return fmt.Errorf("need node ssh port, example: vmadm ssh --remote-ssh-port 22")
		}

		if len(userName) == 0 {
			return fmt.Errorf("need user name, example: vmadm ssh --user root")
		}

		if len(server) == 0 {
			return fmt.Errorf("need server url, example: vmadm ssh --server ws://localhost:7777/vm")
		}

		logx.Debugf("server:%s, id:%s, user name:%s, ssh port:%d", server, id, userName, sshPort)

		signer, err := newSSHSigner(keyPath)
		if err != nil {
			log.Fatal("私钥解析失败:", err)
		}
		config := &ssh.ClientConfig{
			User: userName,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		ctx, cancle := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancle()

		remoteSshAddr := fmt.Sprintf("%s:%d", "localhost", sshPort)
		url := fmt.Sprintf("ws://%s/api/ws/vm?id=%s&transport=raw&address=%s", server, id, remoteSshAddr)
		websocketConn, resp, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
		if err != nil {
			var msg []byte
			if resp != nil {
				msg, _ = io.ReadAll(resp.Body)
			}
			log.Fatalf("Dial failed: %s, msg:%s", err, string(msg))
		}

		log.Printf("websocket connect to %s\n", url)
		// 连接远程服务器
		sshConn, chans, reqs, err := ssh.NewClientConn(websocketConn.NetConn(), remoteSshAddr, config) // 替换为服务器IP和端口
		if err != nil {
			log.Fatalf("SSH连接失败: %v", err)
		}
		defer sshConn.Close()

		log.Println("new ssh conn")
		client := ssh.NewClient(sshConn, chans, reqs)
		defer client.Close()

		session, err := client.NewSession()
		if err != nil {
			log.Fatalf("Failed to create session: %s", err)
		}
		defer session.Close()

		// 设置终端模式
		session.Stdout = os.Stdout
		session.Stderr = os.Stderr
		session.Stdin = os.Stdin

		fd := int(os.Stdin.Fd())
		oldState, err := term.MakeRaw(fd)
		if err != nil {
			log.Fatalf("Failed to enter raw mode: %v", err)
		}
		defer term.Restore(fd, oldState)

		// 设置终端类型
		modes := ssh.TerminalModes{
			ssh.ECHO:          1,
			ssh.TTY_OP_ISPEED: 14400,
			ssh.TTY_OP_OSPEED: 14400,
		}
		if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
			log.Fatalf("request for pseudo terminal failed: %s", err)
		}

		// 启动远程 shell
		if err := session.Shell(); err != nil {
			log.Fatalf("Failed to start shell: %s", err)
		}

		// 等待结束
		err = session.Wait()
		if err != nil && err != io.EOF {
			log.Printf("SSH session ended with error: %s", err)
		}

		return nil
	},
}

func newSSHSigner(keyPath string) (ssh.Signer, error) {
	if len(keyPath) == 0 {
		currentUser, err := user.Current()
		if err != nil {
			return nil, err
		}

		sshDir := filepath.Join(currentUser.HomeDir, ".ssh")
		keyPath = filepath.Join(sshDir, defaultKeyFile)
	}

	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", keyPath, err)
	}
	return signer, nil
}
