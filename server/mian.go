package main

import (
	"flag"
	"fmt"
	api "titan-tunnel/server/api/export"
	"titan-tunnel/server/config"
	"titan-tunnel/server/httpproxy"
	rpc "titan-tunnel/server/rpc/export"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

var configFile = flag.String("f", "etc/server.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// Override Log and Redis
	c.APIServer.Log = c.Log
	c.RPCServer.Redis = redis.RedisKeyConf{RedisConf: c.Redis}
	c.RPCServer.Log = c.Log
	c.RPCServer.APIServer = fmt.Sprintf("localhost:%d", c.APIServer.Port)

	group := service.NewServiceGroup()
	api.AddAPIService(group, c.APIServer)
	rpc.AddRPCService(group, c.RPCServer)

	httpProxyServer := httpproxy.NewServer(c.HTTPProxy, c.Redis)
	group.Add(httpProxyServer)

	logx.Infof("Starting api server at %s:%d", c.APIServer.Host, c.APIServer.Port)
	logx.Infof("Starting rpc server at %s...", c.RPCServer.ListenOn)
	group.Start()

}
