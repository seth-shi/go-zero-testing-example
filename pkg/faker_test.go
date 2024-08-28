package pkg

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	errTransactionInsert = errors.New("insert error")
)

func TestFakerRedisServer(t *testing.T) {

	var (
		key = "key"
		val = "val"
	)

	m, _ := FakerRedisServer()
	t.Cleanup(m.Close)

	assert.Nil(t, m.Set(key, val))
	savedVal, err := m.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, val, savedVal)
}

func TestFakerDatabaseServer(t *testing.T) {

	dsn := FakerDatabaseServer()
	db, err := gorm.Open(
		mysql.Open(dsn),
	)
	logx.Must(err)
	assert.NotNil(t, db)

	type user struct {
		gorm.Model
		Name string `gorm:"size:255;index:idx_name,unique"`
	}
	assert.False(t, db.Migrator().HasTable(&user{}))

	err = db.Migrator().CreateTable(&user{})
	assert.NoError(t, err)
	assert.True(t, db.Migrator().HasTable(&user{}))

	var (
		count     int64
		wantCount int64 = 1
		model           = &user{}
	)
	assert.NoError(t, db.Model(model).Create(&user{Name: "test"}).Error)
	assert.NoError(t, db.Model(model).Count(&count).Error)
	assert.Equal(t, count, wantCount)
}
