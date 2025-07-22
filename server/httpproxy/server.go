package httpproxy

import (
	"context"
	"log"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type Handle struct {
	routes map[string]http.HandlerFunc
	tunMgr *TunnelManager
}

func newHandler(tunManager *TunnelManager) *Handle {
	h := &Handle{routes: make(map[string]http.HandlerFunc), tunMgr: tunManager}
	h.routes["/ws/web"] = h.handleWS
	return h
}

func (h *Handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	handle := h.routes[path]
	if handle != nil {
		handle(w, r)
	} else {
		h.handleHttpProxy(w, r)
	}
}

func (h *Handle) handleWS(w http.ResponseWriter, r *http.Request) {
	var req WebWSReq
	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	browserws := newBrowserWS(h.tunMgr)
	err := browserws.ServeWS(w, r, &req)
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	} else {
		httpx.Ok(w)
	}
}

func (h *Handle) handleHttpProxy(w http.ResponseWriter, r *http.Request) {
	httpProxy := newHttProxy(h.tunMgr)
	httpProxy.HandleProxy(w, r)
}

type Server struct {
	addr    string
	server  *http.Server
	Handler http.Handler
	// router  *http.ServeMux
}

func NewServer(addr string, redisConf redis.RedisConf) *Server {
	redis := redis.MustNewRedis(redisConf)
	tunManager := NewTunnelManager(redis)
	return &Server{addr: addr, Handler: newHandler(tunManager)}
}

func (s *Server) Start() {
	if s.server != nil {
		logx.Error("server already start")
		return
	}

	s.server = &http.Server{
		Addr:    s.addr,
		Handler: s.Handler,
	}

	go func() {
		logx.Infof("Starting HTTP server on %s", s.addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

}

func (s *Server) Stop() {
	if s.server == nil {
		logx.Error("server not start")
		return
	}

	log.Println("Shutting down server...")
	if err := s.server.Shutdown(context.TODO()); err != nil {
		logx.Error("server shutdown failed: %v", err)
	}

}
