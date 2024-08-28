package mock

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

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
