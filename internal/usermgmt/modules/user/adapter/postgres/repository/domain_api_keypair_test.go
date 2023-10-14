package repository

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/crypt"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func DomainAPIKeypairRepoWithSqlMock() (*DomainAPIKeypairRepo, *testutil.MockDB) {
	r := &DomainAPIKeypairRepo{
		InitialVector: "initial---vector",
		EncryptedKey:  "random-encrypted-key-24b",
	}
	return r, testutil.NewMockDB()
}

func TestDomainAPIKeypairRepo_Create(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	apiKeypair := APIKeyPair{
		privateKey: field.NewString("random-data"),
	}
	_, apiKeypairValues := apiKeypair.FieldMap()

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := DomainAPIKeypairRepoWithSqlMock()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(apiKeypairValues))...)
		cmdTag := pgconn.CommandTag(`1`)
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, aggregate.DomainAPIKeypair{DomainAPIKeypair: &apiKeypair})
		assert.Nil(t, err)
	})
	t.Run("create fail", func(t *testing.T) {
		repo, mockDB := DomainAPIKeypairRepoWithSqlMock()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(apiKeypairValues))...)
		mockDB.DB.On("Exec", args...).Return(nil, puddle.ErrClosedPool)

		err := repo.Create(ctx, mockDB.DB, aggregate.DomainAPIKeypair{DomainAPIKeypair: &apiKeypair})
		assert.Equal(t, puddle.ErrClosedPool, err)
	})
}

func TestDomainAPIKeypairRepo_GetByPublicKey(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	publicKey := "valid-public-key"
	privateKey := "valid-private-key"

	apiKeypair := newAPIKeyPair(&APIKeyPair{})
	_, values := apiKeypair.FieldMap()
	args := append([]interface{}{}, genSliceMock(len(values))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := DomainAPIKeypairRepoWithSqlMock()
		encryptedPrivateKey, _ := crypt.AESEncrypt(privateKey, []byte(repo.EncryptedKey), []byte(repo.InitialVector))
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, database.Text(publicKey))
		mockDB.DB.On("QueryRow").Once().Return(mockDB.Row)
		mockDB.Row.On("Scan", args...).Once().Run(func(args mock.Arguments) {
			reflect.ValueOf(args[1]).Elem().Set(reflect.ValueOf(field.NewString(crypt.EncodeBase64(encryptedPrivateKey))))
		}).Return(nil)

		apiKeypair, err := repo.GetByPublicKey(ctx, mockDB.DB, publicKey)
		assert.Nil(t, err)
		assert.NotNil(t, apiKeypair)
	})
	t.Run("db Query returns error", func(t *testing.T) {
		repo, mockDB := DomainAPIKeypairRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, database.Text(publicKey))
		mockDB.DB.On("QueryRow").Once().Return(mockDB.Row)
		mockDB.Row.On("Scan", args...).Once().Return(pgx.ErrTxClosed)
		apiKeypair, err := repo.GetByPublicKey(ctx, mockDB.DB, publicKey)
		assert.NotNil(t, err)
		assert.Nil(t, apiKeypair)
	})

}
