package pkg

import (
	"os"
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

func TestCreateTempFile(t *testing.T) {
	tests := []string{
		".json",
		".yaml",
		".yml",
	}
	text := `{
	"a": "foo",
	"b": 1,
	"c": "${FOO}",
	"d": "abcd!@#$112"
}`
	t.Setenv("FOO", "2")

	for _, test := range tests {
		test := test
		t.Run(
			test, func(t *testing.T) {
				removeFile, tmpFile, err := CreateTempFile(test, text)
				require.NoError(t, err)
				defer removeFile()

				require.FileExists(t, tmpFile)
				content, err := os.ReadFile(tmpFile)
				require.NoError(t, err)
				require.Equal(t, text, string(content))
			},
		)
	}

}
