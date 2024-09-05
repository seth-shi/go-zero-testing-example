package main

import (
	"fmt"

	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/id"
	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/internal/config"
	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/internal/server"
	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

var getConfig = loadConfigByFile

func loadConfigByFile(filename string) config.Config {
	var c config.Config
	conf.MustLoad(filename, &c)
	return c
}

func main() {

	cfg := getConfig("etc/id.yaml")
	ctx := svc.NewServiceContext(cfg)

	s := zrpc.MustNewServer(
		ctx.Config.RpcServerConf, func(grpcServer *grpc.Server) {
			id.RegisterIdServer(grpcServer, server.NewIdServer(ctx))
		},
	)
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", ctx.Config.ListenOn)
	s.Start()
}
