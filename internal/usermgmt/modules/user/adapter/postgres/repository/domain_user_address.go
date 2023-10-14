package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type DomainUserAddressRepo struct{}

type UserAddressAttribute struct {
	UserAddressID  field.String
	UserID         field.String
	AddressType    field.String
	PostalCode     field.String
	PrefectureID   field.String
	City           field.String
	FirstStreet    field.String
	SecondStreet   field.String
	OrganizationID field.String
}

type UserAddress struct {
	UserAddressAttribute

	// These attributes belong to postgres database context
	UpdatedAt field.Time
	CreatedAt field.Time
	DeletedAt field.Time
}

func NewUserAddress(userAddress entity.DomainUserAddress) *UserAddress {
	now := field.NewTime(time.Now())
	address := &UserAddress{
		UserAddressAttribute: UserAddressAttribute{
			UserAddressID:  userAddress.UserAddressID(),
			UserID:         userAddress.UserID(),
			AddressType:    userAddress.AddressType(),
			PostalCode:     userAddress.PostalCode(),
			City:           userAddress.City(),
			PrefectureID:   userAddress.PrefectureID(),
			FirstStreet:    userAddress.FirstStreet(),
			SecondStreet:   userAddress.SecondStreet(),
			OrganizationID: userAddress.OrganizationID(),
		},
		UpdatedAt: now,
		CreatedAt: now,
		DeletedAt: field.NewNullTime(),
	}
	field.SetUndefinedFieldsToNull(address)

	return address
}

func (ud *UserAddress) UserAddressID() field.String {
	return ud.UserAddressAttribute.UserAddressID
}

func (ud *UserAddress) UserID() field.String {
	return ud.UserAddressAttribute.UserID
}

func (ud *UserAddress) AddressType() field.String {
	return ud.UserAddressAttribute.AddressType
}

func (ud *UserAddress) PostalCode() field.String {
	return ud.UserAddressAttribute.PostalCode
}

func (ud *UserAddress) City() field.String {
	return ud.UserAddressAttribute.City
}

func (ud *UserAddress) PrefectureID() field.String {
	return ud.UserAddressAttribute.PrefectureID
}

func (ud *UserAddress) FirstStreet() field.String {
	return ud.UserAddressAttribute.FirstStreet
}

func (ud *UserAddress) SecondStreet() field.String {
	return ud.UserAddressAttribute.SecondStreet
}

func (ud *UserAddress) OrganizationID() field.String {
	return ud.UserAddressAttribute.OrganizationID
}

// FieldMap returns field in UserAddresses table
func (ud *UserAddress) FieldMap() ([]string, []interface{}) {
	return []string{
			"user_address_id",
			"user_id",
			"address_type",
			"postal_code",
			"prefecture_id",
			"city",
			"first_street",
			"second_street",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&ud.UserAddressAttribute.UserAddressID,
			&ud.UserAddressAttribute.UserID,
			&ud.UserAddressAttribute.AddressType,
			&ud.UserAddressAttribute.PostalCode,
			&ud.UserAddressAttribute.PrefectureID,
			&ud.UserAddressAttribute.City,
			&ud.UserAddressAttribute.FirstStreet,
			&ud.UserAddressAttribute.SecondStreet,
			&ud.CreatedAt,
			&ud.UpdatedAt,
			&ud.DeletedAt,
			&ud.UserAddressAttribute.OrganizationID,
		}
}

func (*UserAddress) TableName() string {
	return "user_address"
}

func (r *DomainUserAddressRepo) SoftDeleteByUserIDs(ctx context.Context, db database.QueryExecer, userIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainUserAddressRepo.SoftDeleteByUserIDs")
	defer span.End()

	sql := `UPDATE user_address SET deleted_at = now() WHERE user_id = ANY($1) AND deleted_at IS NULL`
	_, err := db.Exec(ctx, sql, database.TextArray(userIDs))
	if err != nil {
		return InternalError{RawError: errors.Wrap(err, "db.Exec")}
	}

	return nil
}

func (r *DomainUserAddressRepo) UpsertMultiple(ctx context.Context, db database.QueryExecer, userAddresses ...entity.DomainUserAddress) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainUserAddressRepo.UpsertMultiple")
	defer span.End()

	batch := &pgx.Batch{}
	r.queueUpsert(batch, userAddresses...)

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := batchResults.Exec()
		if err != nil {
			return InternalError{RawError: errors.Wrap(err, "batchResults.Exec")}
		}
	}

	return nil
}

func (r *DomainUserAddressRepo) queueUpsert(batch *pgx.Batch, userAddresses ...entity.DomainUserAddress) {
	queue := func(b *pgx.Batch, userAddress *UserAddress) {
		fieldNames := database.GetFieldNames(userAddress)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		stmt := fmt.Sprintf(`
			INSERT INTO %s (%s) VALUES (%s)
			ON CONFLICT ON CONSTRAINT user_address__pk 
			DO UPDATE SET user_id = EXCLUDED.user_id, address_type = EXCLUDED.address_type, postal_code = EXCLUDED.postal_code, 
            prefecture_id = EXCLUDED.prefecture_id, city = EXCLUDED.city, 
            first_street = EXCLUDED.first_street, second_street = EXCLUDED.second_street, 
            updated_at = now(), deleted_at = NULL`,
			userAddress.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
		)

		b.Queue(stmt, database.GetScanFields(userAddress, fieldNames)...)
	}

	for _, userAddress := range userAddresses {
		repoDomainUserAddress := NewUserAddress(userAddress)

		queue(batch, repoDomainUserAddress)
	}
}

func (r *DomainUserAddressRepo) GetByUserID(ctx context.Context, db database.QueryExecer, userID pgtype.Text) ([]entity.DomainUserAddress, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserAddressRepo.GetByUserID")
	defer span.End()

	userAddressRepo := NewUserAddress(entity.DefaultDomainUserAddress{})
	fields, _ := userAddressRepo.FieldMap()
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1 AND deleted_at IS NULL", strings.Join(fields, ","), userAddressRepo.TableName())

	rows, err := db.Query(ctx, stmt, &userID)
	if err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "db.Query")}
	}
	defer rows.Close()

	var userAddresses []entity.DomainUserAddress
	for rows.Next() {
		userAddress := NewUserAddress(entity.DefaultDomainUserAddress{})
		_, fieldValues := userAddress.FieldMap()
		if err := rows.Scan(fieldValues...); err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "row.Scan")}
		}

		userAddresses = append(userAddresses, userAddress)
	}

	return userAddresses, nil
}
