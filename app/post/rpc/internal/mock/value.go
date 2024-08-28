package mock

import (
	"sync"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
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

func (m *value) TableScheme(dest any) (*schema.Schema, error) {
	return schema.Parse(dest, &m.cacheStore, m.Database.NamingStrategy)
}
