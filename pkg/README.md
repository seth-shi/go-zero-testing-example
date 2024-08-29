## 开始

### Mock 方法介绍
* 参考此方案[https://taoshu.in/go/mock.html](https://taoshu.in/go/mock.html)
* 下面是简述上面链接的方案
```go
package main

import (
  "testing"
  "time"
)
////////////////////////////////////////
// 0x00 比如有一个这样的函数, 实际上我们是不可测试的, 因为 time.Now 不受代码控制
func Foo(t time.Time) {

  // 获取当前时间
  n := time.Now()
  if n.Sub(t) > 10*time.Minute {
    // ...
  }
}
////////////////////////////////////////
// 0x01 使用全局变量  (net/http.DefaultTransport 做法)
var (
  Now time.Time
)

func Foo(t time.Time) {

  // 获取当前时间
  if Now.Sub(t) > 10*time.Minute {
    // ...
  }
}
func TestTime(t *testing.T) {
  Now = time.Now().Add(time.Hour)
  Foo(time.Now())
}
////////////////////////////////////////
// 0x02 依赖注入接口 (io 下的基本都这种)
func Foo(n time.Time, t time.Time) {

  // 获取当前时间
  if n.Sub(t) > 10*time.Minute {
    // ...
  }
}
func TestTime(t *testing.T) {
  Foo(time.Now().Add(time.Hour), time.Now())
}

```
* 但不管哪种都需要你写代码的时候很痛苦(**不要纠结过高的代码覆盖率**)
* 第三种方案是猴子补丁, 不过我没用, 详情可查看: [https://github.com/bouk/monkey?tab=readme-ov-file](https://github.com/bouk/monkey?tab=readme-ov-file)

### 涉及测试的类型

* 单元测试
    * 业务的实现代码基本都是写单元测试, 比如在`go-zero`内部的`logic`
* 集成测试
    * 有服务依赖的, 比如数据库依赖, 其它服务依赖. 会去启动一个别的服务
    * 一般集成测试我会写在服务的根目录下

## 例子仓库地址
* [https://github.com/seth-shi/go-zero-testing-example](https://github.com/seth-shi/go-zero-testing-example)
* 服务的架构如下
    * **id** 服务是雪花**id**服务, 零依赖
    * **post** 服务依赖**雪花服务**, **数据库**,  **Redis**
```shell
├─app
│  ├─id
│  │  └─rpc
│  │      ├─etc
│  │      ├─id
│  │      └─internal
│  │          ├─config
│  │          ├─logic
│  │          ├─faker
│  │          ├─server
│  │          └─svc
│  └─post
│      └─rpc
│          ├─etc
│          ├─internal
│          │  ├─config
│          │  ├─logic
│          │  ├─faker
│          │  ├─model
│          │  │  ├─do
│          │  │  └─entity
│          │  ├─server
│          │  └─svc
│          └─post
└─pkg
└─go.mod
```

## 单元测试

### id 服务
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

### post 服务
* 这个服务中的**logic**所有依赖在单元测试的时候都需要**mock**出来
* 服务根目录下的依赖可以直接启动实际的数据库

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

* 业务实际代码
```go
////////////////////////////////////////
// svcCtx
package svc

import (
	"context"

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

	// 数据库表, 每个表一个字段
	Query   *do.Query
	PostDao do.IPostDo
}

func NewServiceContext(c config.Config) *ServiceContext {

	conn, err := gorm.Open(mysql.Open(c.DataSource))
	if err != nil {
		logx.Must(err)
	}
	
	idClient := id.NewIdClient(zrpc.MustNewClient(c.IdRpc).Conn())
	entity.SetIdGenerator(idClient)

	// 使用 redisv8, 而非 go-zero 自己的 redis
	rdb := redis.NewClient(
		&redis.Options{
			Addr:     c.RedisConf.Host,
			Password: c.RedisConf.Pass,
			DB:       0,
		},
	)

	// 使用 grom gen, 而非 go-zero 自己的 sqlx
	query := do.Use(conn)
	return &ServiceContext{
		Config:  c,
		Redis:   rdb,
		IdRpc:   idClient,
		Query:   query,
		PostDao: query.Post.WithContext(context.Background()),
	}
}


////////////////////////////////////////
// logic
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

	// 获取第一条记录
	p, err := l.
		svcCtx.
		PostDao.
		WithContext(l.ctx).
		Where(l.svcCtx.Query.Post.ID.Eq(in.GetId())).
		First()
	if err != nil {
		return nil, err
	}

	// 增加浏览量
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

### 开始写单元测试

```go
package logic

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/config"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/mock"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/model/do"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/svc"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/post"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

////////////////////////////////////////
// 注意, 此部分是单元测试, 不依赖任何外部依赖
// 逻辑的实现尽量通过接口的方式去实现
// 区别于服务根目录下的集成测试, 集成测试会启动服务包括依赖
func TestGetLogic_Get(t *testing.T) {

	var (
		mockIdClient         = &IdServer{}
		mockRedis, redisMock = redismock.NewClientMock()
		// 此 faker 实例可查看代码
		mockDao              = do.NewMockPostDao()
		svcCtx               = &svc.ServiceContext{
			Config:  config.Config{},
			Redis:   mockRedis,
			IdRpc:   mockIdClient,
			Query:   &do.Query{},
			PostDao: mockDao,
		}
		errNotFound      = errors.New("not found")
		errRedisNotFound = errors.New("redis not found")
	)
	// faker redis 返回值
	mockCall := mockDao.On("First", mock2.Anything).Return(1, nil)
	redisMock.ExpectIncr("post:1").SetVal(1)
	logic := NewGetLogic(context.Background(), svcCtx)

	// 正常的情况
	resp, err := logic.Get(&post.PostRequest{})
	require.NoError(t, err)
	require.Equal(t, uint64(1), resp.GetId())

	// redis 错误的情况
	redisMock.ExpectIncr("post:1").SetErr(errRedisNotFound)
	_, err3 := logic.Get(&post.PostRequest{})
	require.ErrorIs(t, err3, errRedisNotFound)

	// 数据库测试的情况
	mockCall.Unset()
	mockDao.On("First", mock2.Anything).Return(0, errNotFound)
	_, err2 := logic.Get(&post.PostRequest{})
	require.ErrorIs(t, err2, errNotFound)
}

type IdServer struct {
	mock.Mock
}

func (m *IdServer) Get(ctx context.Context, in *id.IdRequest, opts ...grpc.CallOption) (*id.IdResponse, error) {
	args := m.Called()
	idResp := args.Get(0).(uint64)

	return &id.IdResponse{
		Id:   idResp,
		Node: idResp,
	}, args.Error(1)
}


////////////////////////////////////////
// 这个需要放到 gorm 生成 do 包下
type MockPostDao struct {
	postDo
	mock.Mock
}

func NewMockPostDao() *MockPostDao {
	dao := &MockPostDao{}
	dao.withDO(new(gen.DO))
	return dao
}

func (d *MockPostDao) WithContext(ctx context.Context) IPostDo {
	return d
}

func (d *MockPostDao) Where(conds ...gen.Condition) IPostDo {
	return d
}

func (d *MockPostDao) First() (*entity.Post, error) {
	args := d.Called()
	return &entity.Post{
		ID:        uint64(args.Int(0)),
		CreatedAt: lo.ToPtr(time.Now()),
	}, args.Error(1)
}
```
* 至此, 上面的代码就可以 **100%** 覆盖率测试业务代码了


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
////////////////////////////////////////
package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/config"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/mock"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/model/do"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/svc"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/post"
	"github.com/seth-shi/go-zero-testing-example/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	mockModel   *mock.DatabaseModel
	rpcListenOn string
)

func TestMain(m *testing.M) {

	// 使用默认配置
	var (
		// 使用 miniredis
		_, addr, _ = pkg.FakerRedisServer()
		// 使用 go-mysql-server
		dsn        = pkg.FakerDatabaseServer()
		err        error
	)
	// 随机一个端口来启动服务
	rpcPort, err := pkg.GetAvailablePort()
	logx.Must(err)
	rpcListenOn = fmt.Sprintf(":%d", rpcPort)
	// 初始化数据库, 用来后续测试
	mockModel = mock.MakeDatabaseModel(dsn)
	svcCtxGet = func() (*svc.ServiceContext, error) {

		// 修改 main.go 的 svcCtxGet, 不要从文件中读取配置
		conn, err := gorm.Open(mysql.Open(dsn))
		if err != nil {
			logx.Must(err)
		}

		query := do.Use(conn)
		return &svc.ServiceContext{
			Config: config.Config{
				RpcServerConf: zrpc.RpcServerConf{
					ListenOn: rpcListenOn,
				},
			},
			Redis: redis.NewClient(
				&redis.Options{
					Addr: addr,
					DB:   0,
				},
			),
			// id 服务职能去 faker
			IdRpc:   &IdServer{},
			Query:   query,
			PostDao: query.Post.WithContext(context.Background()),
		}, nil
	}

	// 启动服务
	go main()

	// 运行测试
	code := m.Run()
	os.Exit(code)
}


// 测试 rpc 调用
func TestGet(t *testing.T) {

	conn, err := zrpc.NewClient(
		zrpc.RpcClientConf{
			Target:   rpcListenOn,
			NonBlock: false,
		},
	)

	require.NoError(t, err)
	client := post.NewPostClient(conn.Conn())
	resp, err := client.Get(context.Background(), &post.PostRequest{Id: mockModel.PostModel.ID})
	require.NoError(t, err)
	require.NotZero(t, resp.GetId())
	require.Equal(t, resp.GetId(), mockModel.PostModel.ID)
	require.Equal(t, resp.Title, lo.FromPtr(mockModel.PostModel.Title))
}

////////////////////////////////////////
// faker 包代码
// FakerDatabaseServer 测试环境可以使用容器化的 dsn/**
package pkg

import (
	"fmt"
	"log"

	"github.com/alicebob/miniredis/v2"
	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
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

func FakerRedisServer() (*miniredis.Miniredis, string, string) {
	m := miniredis.NewMiniRedis()
	if err := m.Start(); err != nil {
		log.Fatalf("could not start miniredis: %s", err)
	}

	return m, m.Addr(), redis.NodeType
}

////////////////////////////////////////
// 数据库初始化部分
package mock

import (
	"context"
	"math/rand"

	"github.com/samber/lo"
	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/id"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/model/entity"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DatabaseModel struct {
	PostModel *entity.Post
}

type fakerDatabaseKey struct{}

func (f *fakerDatabaseKey) Get(ctx context.Context, in *id.IdRequest, opts ...grpc.CallOption) (*id.IdResponse, error) {
	return &id.IdResponse{
		Id:   uint64(rand.Int63()),
		Node: 1,
	}, nil
}

func MakeDatabaseModel(dsn string) *DatabaseModel {

	db, err := gorm.Open(
		mysql.Open(dsn),
	)
	logx.Must(err)

	// createTables
	logx.Must(db.Migrator().CreateTable(&entity.Post{}))

	// test data
	entity.SetIdGenerator(&fakerDatabaseKey{})
	postModel := &entity.Post{
		Title:   lo.ToPtr("test"),
		Content: lo.ToPtr("content"),
	}
	logx.Must(db.Create(postModel).Error)
	entity.SetIdGenerator(nil)

	return &DatabaseModel{PostModel: postModel}
}

```

## End

* 很多时候不可能写这么多测试代码, 这里就给一个例子, 后续继续完善
* 完整代码 [https://github.com/seth-shi/go-zero-testing-example](https://github.com/seth-shi/go-zero-testing-example)