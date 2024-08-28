package main

import (
	"flag"
	"fmt"

	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/config"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/server"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/svc"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/post"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var svcCtxGet = getCtxByConfigFile

func getCtxByConfigFile() (*svc.ServiceContext, error) {
	flag.Parse()
	var c config.Config
	if err := conf.Load("etc/post.yaml", &c); err != nil {
		return nil, err
	}

	return svc.NewServiceContext(c), nil
}

func main() {

	ctx, err := svcCtxGet()
	logx.Must(err)
	s := zrpc.MustNewServer(
		ctx.Config.RpcServerConf, func(grpcServer *grpc.Server) {
			post.RegisterPostServer(grpcServer, server.NewPostServer(ctx))

			if ctx.Config.Mode == service.DevMode || ctx.Config.Mode == service.TestMode {
				reflection.Register(grpcServer)
			}
		},
	)
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", ctx.Config.ListenOn)
	s.Start()
}
