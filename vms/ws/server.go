package ws

import (
	"titan-vm/vms/internal/svc"

	"github.com/zeromicro/go-zero/rest"

	"net/http"
)

type Server struct {
	tunMgr *TunnelManager
	server *rest.Server
	ctx    *svc.ServiceContext
	ws     *WsHandler
	cmd    *CmdHandler
}

func NewServer(server *rest.Server, ctx *svc.ServiceContext) *Server {
	tunMgr := newTunnelManager(ctx)
	wsHanler := newWsHandler(tunMgr)
	cmdHandler := newCmdHandler(tunMgr)

	ws := &Server{tunMgr: tunMgr, server: server, ws: wsHanler, cmd: cmdHandler, ctx: ctx}
	ws.registerHandlers()
	return ws
}

func (s *Server) registerHandlers() {
	s.server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/",
		Handler: s.ws.indexHandler,
	})

	var wsRoutes = []rest.Route{
		// /node?uuid={xxx}&os={windows/macos/linux}&vmapi={libvrit/multipass}
		{
			Method:  http.MethodGet,
			Path:    "/node",
			Handler: s.ws.nodeHandler,
		},
		// /vm?uuid={xxxx}&transport={raw/websocket}&vmapi={libvirt/multipass}&address=xxx
		{
			Method:  http.MethodGet,
			Path:    "/vm",
			Handler: s.ws.vmHandler,
		},
	}

	var cmdRoutes = []rest.Route{
		{
			Method:  http.MethodPost,
			Path:    "/cmd/downloadimage",
			Handler: s.cmd.downloadImageHandler,
		},
	}

	// rest.WithJwt(ws.ctx.Config.JwtAuth.AccessSecret),
	s.server.AddRoutes(wsRoutes)
	s.server.AddRoutes(cmdRoutes)

}
