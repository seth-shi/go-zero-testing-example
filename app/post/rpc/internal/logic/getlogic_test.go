package logic

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/config"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/mock"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/model/do"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/svc"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/post"
	"github.com/stretchr/testify/require"
)

// 注意, 此部分是单元测试, 不依赖任何外部依赖
// 逻辑的实现尽量通过接口的方式去实现
// 区别于服务根目录下的集成测试, 集成测试会启动服务包括依赖
func TestGetLogic_Get(t *testing.T) {

	var (
		mockVal = mock.GetValue()
		svcCtx  = &svc.ServiceContext{
			Config: config.Config{},
			Redis:  mockVal.Redis,
			IdRpc:  mockVal.IdServer,
			Query:  do.Use(mockVal.Database),
		}
		errNotFound      = errors.New("not found")
		errRedisNotFound = errors.New("redis not found")
		selectSql        = "SELECT (.+) FROM `posts` WHERE `posts`.`id` = (.+)"
		columns          = []string{"id", "title", "content", "created_at", "updated_at"}
		row              = []driver.Value{1, "title", "content", time.Now(), time.Now()}
	)

	logic := NewGetLogic(context.Background(), svcCtx)

	// mock 数据库返回
	mockVal.
		DatabaseMock.
		ExpectQuery(selectSql).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(row...))
	mockVal.RedisMock.ExpectIncr("post:1").SetVal(1)
	resp, err := logic.Get(&post.PostRequest{Id: 1})
	require.NoError(t, err)
	require.Equal(t, uint64(1), resp.GetId())
	require.Equal(t, "title", resp.GetTitle())

	// redis 返回错误的场景
	mockVal.
		DatabaseMock.
		ExpectQuery(selectSql).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(row...))
	mockVal.RedisMock.ExpectIncr("post:1").SetErr(errRedisNotFound)
	_, err3 := logic.Get(&post.PostRequest{Id: 1})
	require.ErrorIs(t, err3, errRedisNotFound)

	// 数据库返回错误的场景
	mockVal.
		DatabaseMock.
		ExpectQuery(selectSql).
		WithArgs(1, 1).
		WillReturnError(errNotFound)
	_, err2 := logic.Get(&post.PostRequest{Id: 1})
	require.ErrorIs(t, err2, errNotFound)
}
