package export

import (
	"titan-ipoverlay/ippop/api/internal/config"
	"titan-ipoverlay/ippop/api/internal/handler"
	"titan-ipoverlay/ippop/api/internal/svc"
	"titan-ipoverlay/ippop/api/socks5"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

type APIServerConfig config.Config

func AddAPIService(group *service.ServiceGroup, c APIServerConfig) {
	server := rest.MustNewServer(c.RestConf)
	group.Add(server)
	// defer server.Stop()

	ctx := svc.NewServiceContext(config.Config(c))
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
	// defer socks5.Stop()

	// socks5.Startup()

	// group.Add(server)
	group.Add(socks5)
	// logx.Infof("API server start at %s:%d...", c.Host, c.Port)
	// server.Start()
}
