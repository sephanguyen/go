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

type DomainTaggedUserRepo struct{}

type TaggedUserAttribute struct {
	TagID          field.String
	UserID         field.String
	OrganizationID field.String
}

type TaggedUser struct {
	TaggedUserAttribute

	CreatedAt field.Time
	UpdatedAt field.Time
	DeletedAt field.Time
}

func (ut *TaggedUser) TagID() field.String {
	return ut.TaggedUserAttribute.TagID
}

func (ut *TaggedUser) UserID() field.String {
	return ut.TaggedUserAttribute.UserID
}

func (ut *TaggedUser) OrganizationID() field.String {
	return ut.TaggedUserAttribute.OrganizationID
}

func NewTaggedUser(taggedUser entity.DomainTaggedUser) *TaggedUser {
	now := time.Now()
	return &TaggedUser{
		TaggedUserAttribute: TaggedUserAttribute{
			UserID:         taggedUser.UserID(),
			TagID:          taggedUser.TagID(),
			OrganizationID: taggedUser.OrganizationID(),
		},

		CreatedAt: field.NewTime(now),
		UpdatedAt: field.NewTime(now),
		DeletedAt: field.NewNullTime(),
	}
}

func (ut *TaggedUser) FieldMap() ([]string, []interface{}) {
	return []string{
			"user_id",
			"tag_id",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&ut.TaggedUserAttribute.UserID,
			&ut.TaggedUserAttribute.TagID,
			&ut.CreatedAt,
			&ut.UpdatedAt,
			&ut.DeletedAt,
			&ut.TaggedUserAttribute.OrganizationID,
		}
}

func (*TaggedUser) TableName() string {
	return "tagged_user"
}

func (ut *DomainTaggedUserRepo) GetByUserIDs(ctx context.Context, db database.QueryExecer, userIDs []string) ([]entity.DomainTaggedUser, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainTaggedUserRepo.GetByUserIDs")
	defer span.End()

	var taggedUsers []entity.DomainTaggedUser
	taggedUser := NewTaggedUser(&entity.EmptyDomainTaggedUser{})
	fieldNames := database.GetFieldNames(taggedUser)
	query := fmt.Sprintf(
		`
		SELECT %s
		FROM %s
		WHERE
			user_id = ANY($1) AND
			deleted_at IS NULL
		`,
		strings.Join(fieldNames, ","),
		taggedUser.TableName(),
	)
	rows, err := db.Query(ctx, query, database.TextArray(userIDs))
	if err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "db.Query")}
	}

	defer rows.Close()
	if err := rows.Err(); err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "rows.Err")}
	}

	for rows.Next() {
		taggedUser := NewTaggedUser(&entity.EmptyDomainTaggedUser{})
		_, fields := taggedUser.FieldMap()
		if err := rows.Scan(fields...); err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "rows.Scan")}
		}
		taggedUsers = append(taggedUsers, taggedUser)
	}

	return taggedUsers, nil
}

func (ut *DomainTaggedUserRepo) UpsertBatch(ctx context.Context, db database.QueryExecer, taggedUsers ...entity.DomainTaggedUser) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainTaggedUserRepo.UpsertBatch")
	defer span.End()

	batch := &pgx.Batch{}
	ut.queueUpsert(batch, taggedUsers...)

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for range taggedUsers {
		if _, err := batchResults.Exec(); err != nil {
			return InternalError{RawError: errors.Wrap(err, "batchResults.Exec")}
		}
	}

	return nil
}

func (ut *DomainTaggedUserRepo) queueUpsert(batch *pgx.Batch, taggedUsers ...entity.DomainTaggedUser) {
	queue := func(batch *pgx.Batch, taggedUser *TaggedUser) {
		fieldNames := database.GetFieldNames(taggedUser)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		stmt := fmt.Sprintf(
			`INSERT INTO %s (%s) VALUES (%s)
			ON CONFLICT ON CONSTRAINT pk__tagged_user
			DO UPDATE SET updated_at = now(), deleted_at = NULL`,
			taggedUser.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
		)
		batch.Queue(stmt, database.GetScanFields(taggedUser, fieldNames)...)
	}
	for _, taggedUser := range taggedUsers {
		queue(batch, NewTaggedUser(taggedUser))
	}
}

func (ut *DomainTaggedUserRepo) SoftDelete(ctx context.Context, db database.QueryExecer, taggedUsers ...entity.DomainTaggedUser) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainTaggedUserRepo.SoftDelete")
	defer span.End()

	tagIDs := []string{}
	userIDs := []string{}
	for _, ugm := range taggedUsers {
		tagIDs = append(tagIDs, ugm.TagID().String())
		userIDs = append(userIDs, ugm.UserID().String())
	}

	query := fmt.Sprintf(`
	  UPDATE
	      %s
	  SET deleted_at = now(),
	      updated_at = now()
	  WHERE
	      tag_id = ANY($1) AND
	      user_id = ANY($2) AND
	      deleted_at IS NULL`,
		new(TaggedUser).TableName(),
	)

	if _, err := db.Exec(ctx, query, &tagIDs, &userIDs); err != nil {
		return InternalError{RawError: errors.Wrap(err, "db.Exec")}
	}
	return nil
}

func (ut *DomainTaggedUserRepo) SoftDeleteByUserIDs(ctx context.Context, db database.QueryExecer, userIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainTaggedUserRepo.SoftDeleteByUserIDs")
	defer span.End()

	taggedUser := TaggedUser{}

	sql := fmt.Sprintf(`UPDATE %s SET deleted_at = now() WHERE user_id = ANY($1) AND deleted_at IS NULL`, taggedUser.TableName())
	_, err := db.Exec(ctx, sql, database.TextArray(userIDs))
	if err != nil {
		return InternalError{RawError: errors.Wrap(err, "db.Exec")}
	}

	return nil
}
