package entity

import (
	"context"
	"errors"
	"math/rand"
	"testing"

	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/id"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/mock"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestPost_BeforeCreate(t *testing.T) {

	var (
		mockVal = mock.GetValue()
		data    Post
		tx      = gorm.DB{
			Statement: &gorm.Statement{
				Context: context.Background(),
			},
		}
	)

	data.ID = 1
	assert.NoError(t, data.BeforeCreate(&tx))
	data.ID = 0
	assert.ErrorIs(t, data.BeforeCreate(&tx), errNotInitIdGenerator)

	mockCall := mockVal.
		IdServer.
		On("Get", mock2.Anything).
		Return(1, nil)
	defer mockCall.Unset()

	idClient = mockVal.IdServer
	assert.NoError(t, data.BeforeCreate(&tx))
}

func TestSetIdGeneratorSuccess(t *testing.T) {

	var (
		mockVal = mock.GetValue()
		tx      = gorm.DB{
			Statement: &gorm.Statement{
				Context: context.Background(),
			},
		}
		createId = rand.Int()
	)
	mockCall := mockVal.
		IdServer.
		On("Get", mock2.Anything).
		Return(createId, nil)
	defer mockCall.Unset()

	SetIdGenerator(mockVal.IdServer)
	res, err := idGenerator(&tx)
	assert.NoError(t, err)
	assert.Equal(t, uint64(createId), res)
}

func TestSetIdGenerator(t *testing.T) {

	var (
		mockVal  = mock.GetValue()
		errGetId = errors.New("getid")
		tx       = gorm.DB{
			Statement: &gorm.Statement{
				Context: context.Background(),
			},
		}
	)
	mockCall := mockVal.
		IdServer.
		On("Get", mock2.Anything).
		Return(0, errGetId)
	defer mockCall.Unset()

	tests := []struct {
		name string
		args id.IdClient
		want error
	}{
		{
			args: nil,
			name: "nil",
			want: errNotInitIdGenerator,
		},
		{
			args: mockVal.IdServer,
			name: "mock",
			want: errGetId,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				SetIdGenerator(tt.args)
				_, err := idGenerator(&tx)
				assert.Equal(t, err, tt.want)
			},
		)
	}
}
