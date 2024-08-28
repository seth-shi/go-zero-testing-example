package entity

import (
	"errors"

	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/id"
	"gorm.io/gorm"
)

var (
	idClient              id.IdClient
	errNotInitIdGenerator = errors.New("not init id generator")
)

func idGenerator(tx *gorm.DB) (uint64, error) {

	if idClient == nil {
		return 0, errNotInitIdGenerator
	}

	resp, err := idClient.Get(tx.Statement.Context, &id.IdRequest{})
	if err != nil {
		return 0, err
	}

	return resp.GetId(), nil
}

func SetIdGenerator(f id.IdClient) {
	idClient = f
}

// BeforeCreate 创建使用 id
func (data *Post) BeforeCreate(tx *gorm.DB) error {

	if data.ID > 0 {
		return nil
	}

	snowId, err := idGenerator(tx)
	if err != nil {
		return err
	}

	data.ID = snowId

	return nil
}
