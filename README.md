## 开始
![Coverage](https://img.shields.io/badge/Coverage-53.6%25-yellow)

#### 涉及测试的类型
* 单元测试
    * 业务的实现代码基本都是写单元测试, 比如在`go-zero`内部的`logic`
    * 所有的依赖都使用`mock`, 比如数据库就使用[sql-mock](https://github.com/DATA-DOG/go-sqlmock), [redis-mock](https://github.com/go-redis/redismock), 其它依赖使用接口[testify-mock](https://github.com/stretchr/testify?tab=readme-ov-file#mock-package)
    * 更多**Mock**方案可参考[[https://github.com/bouk/monkey?tab=readme-ov-file](https://github.com/bouk/monkey?tab=readme-ov-file)
      ]([https://github.com/bouk/monkey?tab=readme-ov-file](https://github.com/bouk/monkey?tab=readme-ov-file)
      )
* 集成测试
    * 有服务依赖的, 比如数据库依赖, 其它服务依赖. 会去启动一个别的服务
    * 数据库依赖使用`go-mysql-server`, `redis`使用`mini-redis`(也可以启动一个真正的数据库来测试)

#### 例子仓库地址
* [https://github.com/seth-shi/go-zero-testing-example](https://github.com/seth-shi/go-zero-testing-example)
* 服务的架构如下
    * **id** 服务是雪花**id**服务, 零依赖
    * **post** 服务依赖**雪花服务**, **数据库**,  **Redis**
```shell
├─app
│  ├─id
│  │  └─rpc
│  │      ├─etc (配置文件)
│  │      ├─id (grpc 代码生成)
│  │      └─internal
│  │          ├─config (配置定义)
│  │          ├─logic (业务逻辑)
│  │          ├─mock (单元测试数据)
│  │          ├─server (go-zero 服务端生成)
│  │          └─svc (服务依赖定义)
│  └─post
│      └─rpc
│          ├─etc (配置文件)
│          ├─internal
│          │  ├─config (配置定义)
│          │  ├─faker (集成测试数据)
│          │  ├─logic (业务逻辑)
│          │  ├─mock (单元测试数据)
│          │  ├─model (gorm 生成)
│          │  │  ├─do (数据库查询操作)
│          │  │  └─entity (gorm gen 生成模型定义)
│          │  ├─server (go-zero 服务端生成)
│          │  └─svc (服务依赖定义)
│          └─post (grpc 代码生成)
└─pkg (公共代码)
└─go.mod
```

## 单元测试

#### id 服务

```protobuf
syntax = "proto3";

package id;
option go_package="./id";

message IdRequest {
}

message IdResponse {
  uint64 id = 1;
  uint64 node = 2;
}

service Id {
  rpc Get(IdRequest) returns(IdResponse);
}
```

* 这个服务使用索尼的雪花算法生成**id** ([https://github.com/sony/sonyflake/issues](https://github.com/sony/sonyflake/issues)), 代码非常简单, 我们这里直接跳过说明

#### post 服务

```protobuf
syntax = "proto3";

package post;
option go_package="./post";

message PostRequest {
  uint64  id = 1;
}

message PostResponse {
  uint64 id = 1;
  string title = 2;
  string content = 3;
  uint64 createdAt = 4;
  uint64 viewCount = 5;
}

service Post {
  rpc Get(PostRequest) returns(PostResponse);
}


```

#### 代码

* 服务依赖

```go
package svc

import (
	"github.com/redis/go-redis/v9"
	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/id"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/config"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/model/do"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/model/entity"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	Redis  *redis.Client
	IdRpc  id.IdClient

	Query *do.Query
}

func NewServiceContext(c config.Config) *ServiceContext {

	conn, err := gorm.Open(mysql.Open(c.DataSource))
	logx.Must(err)

	idClient := id.NewIdClient(zrpc.MustNewClient(c.IdRpc).Conn())
	entity.SetIdGenerator(idClient)

	rdb := redis.NewClient(
		&redis.Options{
			Addr:     c.RedisConf.Host,
			Password: c.RedisConf.Pass,
			DB:       0,
		},
	)

	return &ServiceContext{
		Config: c,
		Redis:  rdb,
		IdRpc:  idClient,
		Query:  do.Use(conn),
	}
}
```

* **!!!业务逻辑**

```go
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

	// 数据库查询数据
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

	// redis + 1 浏览量
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
```

#### 开始写单元测试

* 业务逻辑单测

```go
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

```

* mock 包代码组装数据

```go
package mock

import (
	"context"
	"sync"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/id"
	"github.com/stretchr/testify/mock"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type value struct {
	IdServer     *idMock
	Database     *gorm.DB
	DatabaseMock sqlmock.Sqlmock
	RedisMock    redismock.ClientMock
	Redis        *redis.Client

	cacheStore sync.Map
}

var GetValue = sync.OnceValue(
	func() value {

		db, dbMock := makeDatabase()
		redis, redisMock := redismock.NewClientMock()

		return value{
			IdServer:     &idMock{},
			Database:     db,
			DatabaseMock: dbMock,
			Redis:        redis,
			RedisMock:    redisMock,
			cacheStore:   sync.Map{},
		}
	},
)

type idMock struct {
	mock.Mock
}

func (m *idMock) Get(ctx context.Context, in *id.IdRequest, opts ...grpc.CallOption) (*id.IdResponse, error) {
	args := m.Called()
	idResp := uint64(args.Int(0))

	return &id.IdResponse{
		Id:   idResp,
		Node: idResp,
	}, args.Error(1)
}

func makeDatabase() (*gorm.DB, sqlmock.Sqlmock) {

	db, dbMock, err := sqlmock.New()
	logx.Must(err)
	dbMock.
		ExpectQuery("SELECT VERSION()").
		WillReturnRows(sqlmock.NewRows([]string{"VERSION()"}).AddRow("5.7"))

	gormDB, err := gorm.Open(
		mysql.New(
			mysql.Config{
				Conn: db,
			},
		), &gorm.Config{},
	)
	logx.Must(err)

	return gormDB, dbMock
}

```

* 至此, 我们就完成此业务代码的 **100%** 测试覆盖

## 集成测试

* 需要改造一下**main**方法

```go
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

```

* 集成测试方法

```go
package main

import (
	"context"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/config"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/faker"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/model/do"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/svc"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/post"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/zrpc"
)

func TestMain(m *testing.M) {

	// 使用默认配置
	svcCtxGet = func() (*svc.ServiceContext, error) {

		fakerVal := faker.GetValue()
		return &svc.ServiceContext{
			Config: config.Config{
				RpcServerConf: zrpc.RpcServerConf{
					ListenOn: fakerVal.RpcListen,
				},
			},
			Redis: redis.NewClient(
				&redis.Options{
					Addr: fakerVal.RedisAddr,
					DB:   0,
				},
			),
			IdRpc: fakerVal.IdServer,
			Query: do.Use(fakerVal.Gorm),
		}, nil
	}

	go main()

	// 运行测试
	code := m.Run()
	os.Exit(code)
}

// 集成测试
func TestGet(t *testing.T) {

	var (
		fakerVal  = faker.GetValue()
		postModel = fakerVal.Models.PostModel
	)
	conn, err := zrpc.NewClient(
		zrpc.RpcClientConf{
			Target:   fakerVal.RpcListen,
			NonBlock: false,
		},
	)
	require.NoError(t, err)
	client := post.NewPostClient(conn.Conn())
	resp, err := client.Get(context.Background(), &post.PostRequest{Id: postModel.ID})
	require.NoError(t, err)
	require.NotZero(t, resp.GetId())
	require.Equal(t, resp.GetId(), postModel.ID)
	require.Equal(t, resp.Title, lo.FromPtr(postModel.Title))
}

```

* 集成测试依赖服务

```go
package faker

import (
	"context"
	"fmt"
	"math/rand"
	"sync"

	"github.com/alicebob/miniredis/v2"
	"github.com/samber/lo"
	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/id"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/model/entity"
	"github.com/seth-shi/go-zero-testing-example/pkg"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type value struct {
	IdServer    *idGenerator
	Redis       *miniredis.Miniredis
	Models      *fakerModels
	Gorm        *gorm.DB
	RedisAddr   string
	RpcListen   string
	DatabaseDsn string
}

var GetValue = sync.OnceValue(
	func() value {

		redis, redisAddr := pkg.FakerRedisServer()
		dsn := pkg.FakerDatabaseServer()

		rpcPort, err := pkg.GetAvailablePort()
		logx.Must(err)

		conn, err := gorm.Open(mysql.Open(dsn))
		logx.Must(err)

		idGen := &idGenerator{
			startId: uint64(rand.Int() + 1),
			locker:  &sync.RWMutex{},
		}
		return value{
			IdServer:    idGen,
			Redis:       redis,
			RedisAddr:   redisAddr,
			DatabaseDsn: dsn,
			Models:      makeDatabase(dsn, idGen),
			RpcListen:   fmt.Sprintf(":%d", rpcPort),
			Gorm:        conn,
		}
	},
)

type idGenerator struct {
	startId uint64
	locker  sync.Locker
}

func (m *idGenerator) Get(ctx context.Context, in *id.IdRequest, opts ...grpc.CallOption) (*id.IdResponse, error) {

	m.locker.Lock()
	defer m.locker.Unlock()

	m.startId++

	return &id.IdResponse{
		Id:   m.startId,
		Node: 1,
	}, nil
}


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
```

#### 集成测试依赖数据库

```go
package pkg

import (
	"fmt"
	"log"

	"github.com/alicebob/miniredis/v2"
	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/zeromicro/go-zero/core/logx"
)

// FakerDatabaseServer 测试环境可以使用容器化的 dsn/**
func FakerDatabaseServer() string {

	var (
		username = "root"
		password = ""
		host     = "localhost"
		dbname   = "test_db"
		port     int
		err      error
	)

	db := memory.NewDatabase(dbname)
	db.BaseDatabase.EnablePrimaryKeyIndexes()
	provider := memory.NewDBProvider(db)
	engine := sqle.NewDefault(provider)
	mysqlDb := engine.Analyzer.Catalog.MySQLDb
	mysqlDb.SetEnabled(true)
	mysqlDb.AddRootAccount()

	port, err = GetAvailablePort()
	logx.Must(err)

	config := server.Config{
		Protocol: "tcp",
		Address:  fmt.Sprintf("%s:%d", host, port),
	}
	s, err := server.NewServer(
		config,
		engine,
		memory.NewSessionBuilder(provider),
		nil,
	)
	logx.Must(err)
	go func() {
		logx.Must(s.Start())
	}()

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&loc=Local&parseTime=true",
		username,
		password,
		host,
		port,
		dbname,
	)

	return dsn
}

func FakerRedisServer() (*miniredis.Miniredis, string) {
	m := miniredis.NewMiniRedis()
	if err := m.Start(); err != nil {
		log.Fatalf("could not start miniredis: %s", err)
	}

	return m, m.Addr()
}
```

* 至此, 就完成了集成测试的部分

## End

* 很多时候不可能写这么多测试代码, 这里就给一个例子, 后续继续完善
* 完整代码 [https://github.com/seth-shi/go-zero-testing-example](https://github.com/seth-shi/go-zero-testing-example)