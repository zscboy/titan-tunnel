package main

import (
	"flag"
	"fmt"
	api "titan-vm/vms/api/export"
	"titan-vm/vms/internal/config"
	"titan-vm/vms/internal/server"
	"titan-vm/vms/internal/svc"
	"titan-vm/vms/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/vms.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	restServer := rest.MustNewServer(c.RestConf)
	defer restServer.Stop()

	ctx := svc.NewServiceContext(c)
	// ws.NewServer(restServer, ctx)
	api.RegisterHandlers(restServer, c)

	group := service.NewServiceGroup()

	// 启动 RPC 服务
	rpcServer := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterVmsServer(grpcServer, server.NewVmsServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer rpcServer.Stop()

	group.Add(rpcServer)
	group.Add(restServer)

	// api.RegisterHandlers(restServer, c)

	fmt.Printf("http server listen on %s:%d\n", c.RestConf.Host, c.RestConf.Port)
	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	// 启动所有服务
	group.Start()
}
