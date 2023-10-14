package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/crypt"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/pkg/errors"
)

type DomainAPIKeypairRepo struct {
	EncryptedKey  string
	InitialVector string
}

type APIKeyPair struct {
	publicKey      field.String
	privateKey     field.String
	userID         field.String
	organizationID field.String
	updatedAt      field.Time
	createdAt      field.Time
	deletedAt      field.Time
}

func newAPIKeyPair(apiKeypair entity.DomainAPIKeypair) *APIKeyPair {
	now := field.NewTime(time.Now())
	return &APIKeyPair{
		publicKey:      apiKeypair.PublicKey(),
		privateKey:     apiKeypair.PrivateKey(),
		userID:         apiKeypair.UserID(),
		organizationID: apiKeypair.OrganizationID(),
		updatedAt:      now,
		createdAt:      now,
		deletedAt:      field.NewNullTime(),
	}
}

func (apiKeypair *APIKeyPair) PublicKey() field.String {
	return apiKeypair.publicKey
}
func (apiKeypair *APIKeyPair) PrivateKey() field.String {
	return apiKeypair.privateKey
}
func (apiKeypair *APIKeyPair) UserID() field.String {
	return apiKeypair.userID
}
func (apiKeypair *APIKeyPair) OrganizationID() field.String {
	return apiKeypair.organizationID
}

func (apiKeypair *APIKeyPair) FieldMap() ([]string, []interface{}) {
	return []string{
			"public_key",
			"private_key",
			"user_id",
			"resource_path",
			"updated_at",
			"created_at",
			"deleted_at",
		}, []interface{}{
			&apiKeypair.publicKey,
			&apiKeypair.privateKey,
			&apiKeypair.userID,
			&apiKeypair.organizationID,
			&apiKeypair.updatedAt,
			&apiKeypair.createdAt,
			&apiKeypair.deletedAt,
		}
}

func (apiKeypair *APIKeyPair) TableName() string {
	return "api_keypair"
}

func (r *DomainAPIKeypairRepo) Create(ctx context.Context, db database.QueryExecer, apiKeyToCreate aggregate.DomainAPIKeypair) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainAPIKeypairRepo.Create")
	defer span.End()
	databaseAPIKeypairToCreate := newAPIKeyPair(apiKeyToCreate)
	encryptedPrivateKey, err := crypt.AESEncrypt(databaseAPIKeypairToCreate.privateKey.String(), []byte(r.EncryptedKey), []byte(r.InitialVector))
	if err != nil {
		return fmt.Errorf("crypt.AESEncrypt err: %v", err)
	}
	databaseAPIKeypairToCreate.privateKey = field.NewString(crypt.EncodeBase64(encryptedPrivateKey))
	cmdTag, err := database.Insert(ctx, databaseAPIKeypairToCreate, db.Exec)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() != 1 {
		return ErrNoRowAffected
	}

	return nil
}

func (r *DomainAPIKeypairRepo) GetByPublicKey(ctx context.Context, db database.QueryExecer, publicKey string) (entity.DomainAPIKeypair, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainAPIKeypairRepo.GetByPublicKey")
	defer span.End()

	apiKeypair := newAPIKeyPair(&APIKeyPair{})

	stmt := `SELECT %s FROM %s WHERE public_key = $1 and deleted_at is NULL`
	fieldNames, values := apiKeypair.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		apiKeypair.TableName(),
	)

	err := db.QueryRow(
		ctx,
		stmt,
		database.Text(publicKey),
	).Scan(values...)
	if err != nil {
		return nil, err
	}

	privateKey, err := crypt.AESDecryptBase64(apiKeypair.PrivateKey().String(), []byte(r.EncryptedKey), []byte(r.InitialVector))
	if err != nil {
		return nil, errors.Wrap(err, "crypt.AESDecryptBase64")
	}

	apiKeypair.privateKey = field.NewString(string(privateKey))

	return apiKeypair, nil
}
