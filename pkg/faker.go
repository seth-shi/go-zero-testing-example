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
