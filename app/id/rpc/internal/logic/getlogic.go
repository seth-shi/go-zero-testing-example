package logic

import (
	"context"

	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/id"
	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/internal/svc"
	"github.com/sony/sonyflake"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLogic {
	return &GetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetLogic) Get(in *id.IdRequest) (*id.IdResponse, error) {

	genId, err := l.svcCtx.Gen()
	if err != nil {
		return nil, err
	}

	return &id.IdResponse{
		Id:   genId,
		Node: sonyflake.MachineID(genId),
	}, nil
}
