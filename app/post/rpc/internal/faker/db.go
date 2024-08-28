package faker

import (
	"github.com/samber/lo"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/model/entity"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type fakerModels struct {
	PostModel *entity.Post
}

func makeDatabase(dsn string, gen *idGenerator) *fakerModels {

	db, err := gorm.Open(
		mysql.Open(dsn),
	)
	logx.Must(err)

	// 创建表结构
	logx.Must(db.Migrator().CreateTable(&entity.Post{}))

	// 插入测试数据
	entity.SetIdGenerator(gen)
	postModel := &entity.Post{
		Title:   lo.ToPtr("test"),
		Content: lo.ToPtr("content"),
	}
	logx.Must(db.Create(postModel).Error)
	entity.SetIdGenerator(nil)

	return &fakerModels{PostModel: postModel}
}
