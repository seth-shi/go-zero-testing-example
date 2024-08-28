package logic

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/svc"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/post"

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

func (l *GetLogic) Get(in *post.PostRequest) (*post.PostResponse, error) {

	p, err := l.
		svcCtx.
		Query.
		Post.
		WithContext(l.ctx).
		Where(l.svcCtx.Query.Post.ID.Eq(in.GetId())).
		First()
	if err != nil {
		return nil, err
	}

	redisKey := fmt.Sprintf("post:%d", p.ID)
	val, err := l.svcCtx.Redis.Incr(l.ctx, redisKey).Result()
	if err != nil {
		return nil, err
	}

	resp := &post.PostResponse{
		Id:        p.ID,
		Title:     lo.FromPtr(p.Title),
		Content:   lo.FromPtr(p.Content),
		CreatedAt: uint64(p.CreatedAt.Unix()),
		ViewCount: uint64(val),
	}
	return resp, nil
}
