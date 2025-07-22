package httpproxy

import (
	"net"

	"github.com/zeromicro/go-zero/core/logx"
)

type Req struct {
	ID string
	// TODO: change to tcpConn
	conn             net.Conn
	tun              *Tunnel
	rspDataLength    int64
	rspContentLength int64
}

func (r *Req) write(data []byte) error {
	_, err := r.conn.Write(data)
	return err
}

func (r *Req) proxy() error {
	conn := r.conn
	// defer conn.Close()

	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			logx.Infof("req.proxy read: %v", err)
			break
		}

		if err := r.tun.onHTTPData(r.ID, buf[:n]); err != nil {
			return err
		}
	}

	return nil
}
