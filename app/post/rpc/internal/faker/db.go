package faker

import (
	"context"

	"github.com/samber/lo"
	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/id"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/model/entity"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type fakerModels struct {
	PostModel *entity.Post
}

func makeDatabase(db *gorm.DB) *fakerModels {

	// 创建表结构
	logx.Must(db.Migrator().CreateTable(&entity.Post{}))

	// 插入测试数据
	entity.SetIdGenerator(&databaseSeeder{})
	postModel := &entity.Post{
		Title:   lo.ToPtr("test"),
		Content: lo.ToPtr("content"),
	}
	logx.Must(db.Create(postModel).Error)
	entity.SetIdGenerator(nil)

	return &fakerModels{PostModel: postModel}
}

type databaseSeeder struct {
	id uint64
}

func (d *databaseSeeder) Get(ctx context.Context, in *id.IdRequest, opts ...grpc.CallOption) (*id.IdResponse, error) {
	d.id++
	return &id.IdResponse{Id: d.id}, nil
}
