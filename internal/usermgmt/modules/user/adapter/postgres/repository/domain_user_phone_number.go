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

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type DomainUserPhoneNumberRepo struct{}

type UserPhoneNumberAttribute struct {
	ID             field.String
	PhoneNumber    field.String
	Type           field.String
	UserID         field.String
	OrganizationID field.String
}

type UserPhoneNumber struct {
	UserPhoneNumberAttribute

	CreatedAt field.Time
	UpdatedAt field.Time
	DeletedAt field.Time
}

func NewDomainUserPhoneNumber(upn entity.DomainUserPhoneNumber) *UserPhoneNumber {
	now := field.NewTime(time.Now())
	return &UserPhoneNumber{
		UserPhoneNumberAttribute: UserPhoneNumberAttribute{
			ID:             upn.UserPhoneNumberID(),
			UserID:         upn.UserID(),
			PhoneNumber:    upn.PhoneNumber(),
			Type:           upn.Type(),
			OrganizationID: upn.OrganizationID(),
		},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: field.NewNullTime(),
	}
}

func (upn *UserPhoneNumber) UserPhoneNumberID() field.String {
	return upn.UserPhoneNumberAttribute.ID
}
func (upn *UserPhoneNumber) PhoneNumber() field.String {
	return upn.UserPhoneNumberAttribute.PhoneNumber
}
func (upn *UserPhoneNumber) Type() field.String {
	return upn.UserPhoneNumberAttribute.Type
}
func (upn *UserPhoneNumber) UserID() field.String {
	return upn.UserPhoneNumberAttribute.UserID
}
func (upn *UserPhoneNumber) OrganizationID() field.String {
	return upn.UserPhoneNumberAttribute.OrganizationID
}

func (*UserPhoneNumber) TableName() string {
	return "user_phone_number"
}

func (upn *UserPhoneNumber) FieldMap() ([]string, []interface{}) {
	return []string{
			"user_phone_number_id",
			"user_id",
			"phone_number",
			"type",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&upn.UserPhoneNumberAttribute.ID,
			&upn.UserPhoneNumberAttribute.UserID,
			&upn.UserPhoneNumberAttribute.PhoneNumber,
			&upn.UserPhoneNumberAttribute.Type,
			&upn.CreatedAt,
			&upn.UpdatedAt,
			&upn.DeletedAt,
			&upn.UserPhoneNumberAttribute.OrganizationID,
		}
}

func (repo *DomainUserPhoneNumberRepo) GetByUserIDs(ctx context.Context, db database.QueryExecer, userIDs []string) (entity.DomainUserPhoneNumbers, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainUserPhoneNumberRepo.GetByUserIDs")
	defer span.End()

	var userPhoneNumbers []entity.DomainUserPhoneNumber
	userPhoneNumber := NewDomainUserPhoneNumber(&entity.DefaultDomainUserPhoneNumber{})
	fieldNames := database.GetFieldNames(userPhoneNumber)
	query := fmt.Sprintf(
		`
		SELECT %s
		FROM %s
		WHERE
			user_id = ANY($1) AND
			deleted_at IS NULL
		`,
		strings.Join(fieldNames, ","),
		userPhoneNumber.TableName(),
	)
	rows, err := db.Query(ctx, query, database.TextArray(userIDs))
	if err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "db.Query")}
	}

	defer rows.Close()
	if err := rows.Err(); err != nil {
		return nil, InternalError{
			RawError: fmt.Errorf("rows.Err: %w", err),
		}
	}

	for rows.Next() {
		userPhoneNumber := NewDomainUserPhoneNumber(&entity.DefaultDomainUserPhoneNumber{})
		_, fields := userPhoneNumber.FieldMap()
		if err := rows.Scan(fields...); err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "rows.Scan")}
		}
		userPhoneNumbers = append(userPhoneNumbers, userPhoneNumber)
	}

	return userPhoneNumbers, nil
}

func (repo *DomainUserPhoneNumberRepo) SoftDeleteByUserIDs(ctx context.Context, db database.QueryExecer, userIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainUserPhoneNumberRepo.SoftDeleteByUserIDs")
	defer span.End()

	upn := UserPhoneNumber{}

	sql := fmt.Sprintf(`UPDATE %s SET deleted_at = now() WHERE user_id = ANY($1) AND deleted_at IS NULL`, upn.TableName())
	_, err := db.Exec(ctx, sql, database.TextArray(userIDs))
	if err != nil {
		return InternalError{RawError: errors.Wrap(err, "db.Exec")}
	}

	return nil
}

func (repo *DomainUserPhoneNumberRepo) UpsertMultiple(ctx context.Context, db database.QueryExecer, userPhoneNumbers ...entity.DomainUserPhoneNumber) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainUserPhoneNumberRepo.UpsertMultiple")
	defer span.End()

	upsertPhoneNumbers := make([]entity.DomainUserPhoneNumber, 0, len(userPhoneNumbers))
	for _, userPhoneNumber := range userPhoneNumbers {
		// if not Present
		if !field.IsPresent(userPhoneNumber.PhoneNumber()) {
			continue
		}
		upsertPhoneNumbers = append(upsertPhoneNumbers, userPhoneNumber)
	}

	queueFn := func(b *pgx.Batch, userPhoneNumber *UserPhoneNumber) {
		fields, values := userPhoneNumber.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)
			ON CONFLICT ON CONSTRAINT user_phone_number__pk 
			DO UPDATE SET phone_number = excluded.phone_number,
			              type = excluded.type,
			              updated_at = NOW(),
			              deleted_at = NULL
			`,
			userPhoneNumber.TableName(),
			strings.Join(fields, ","),
			placeHolders,
		)

		b.Queue(stmt, values...)
	}

	batch := &pgx.Batch{}

	for _, upsertPhoneNumber := range upsertPhoneNumbers {
		repoUserPhoneNumber := NewDomainUserPhoneNumber(upsertPhoneNumber)
		queueFn(batch, repoUserPhoneNumber)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < len(upsertPhoneNumbers); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return InternalError{RawError: errors.Wrap(err, "batchResults.Exec")}
		}

		if cmdTag.RowsAffected() != 1 {
			return InternalError{RawError: errors.Errorf("user_phone_number was not upserted")}
		}
	}

	return nil
}
