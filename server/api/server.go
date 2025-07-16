package main

import (
	"flag"

	"titan-tunnel/server/api/internal/config"
	"titan-tunnel/server/api/internal/handler"
	"titan-tunnel/server/api/internal/svc"
	"titan-tunnel/server/api/socks5"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/server-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	opts := &socks5.Socks5ServerOptions{
		Address:      c.Socks5.Addr,
		UDPServerIP:  c.Socks5.ServerIP,
		UDPPortStart: c.Socks5.UDPPortStart,
		UDPPortEnd:   c.Socks5.UDPPortEnd,
		EnableAuth:   c.Socks5.EnableAuth,
		Handler:      ctx.TunMgr,
	}
	socks5, err := socks5.New(opts)
	if err != nil {
		panic(err)
	}
	defer socks5.Stop()

	socks5.Start()

	logx.Infof("API server start at %s:%d...", c.Host, c.Port)
	server.Start()
}
