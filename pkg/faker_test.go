package pkg

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestFakerRedisServer(t *testing.T) {

	var (
		key = "key"
		val = "val"
	)

	m, _ := FakerRedisServer()
	t.Cleanup(m.Close)

	require.Nil(t, m.Set(key, val))
	savedVal, err := m.Get(key)
	require.NoError(t, err)
	require.Equal(t, val, savedVal)
}

func TestFakerDatabaseServer(t *testing.T) {

	dsn := FakerDatabaseServer()
	db, err := gorm.Open(
		mysql.Open(dsn),
	)
	logx.Must(err)
	require.NotNil(t, db)

	type user struct {
		gorm.Model
		Name string `gorm:"size:255;index:idx_name,unique"`
	}
	require.False(t, db.Migrator().HasTable(&user{}))

	err = db.Migrator().CreateTable(&user{})
	require.NoError(t, err)
	require.True(t, db.Migrator().HasTable(&user{}))

	var (
		count     int64
		wantCount int64 = 1
		model           = &user{}
	)
	require.NoError(t, db.Model(model).Create(&user{Name: "test"}).Error)
	require.NoError(t, db.Model(model).Count(&count).Error)
	require.Equal(t, count, wantCount)
}
