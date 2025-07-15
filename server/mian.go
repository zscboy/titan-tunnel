package main

import (
	"flag"
	"fmt"
	api "titan-tunnel/server/api/export"
	"titan-tunnel/server/config"
	rpc "titan-tunnel/server/rpc/export"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/server.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	group := service.NewServiceGroup()
	api.AddAPIService(group, c.APIServerConfig)
	rpc.AddRPCService(group, c.RPCServerConfig)

	fmt.Printf("Starting api server at %s:%d\n", c.APIServerConfig.Host, c.APIServerConfig.Port)
	fmt.Printf("Starting rpc server at %s...\n", c.RPCServerConfig.ListenOn)
	group.Start()

}
