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

	"github.com/pkg/errors"
)

type DomainTagRepo struct{}

type TagAttribute struct {
	TagID          field.String
	TagName        field.String
	TagType        field.String
	TagPartnerID   field.String
	IsArchived     field.Boolean
	OrganizationID field.String
}

type Tag struct {
	TagAttribute
	CreatedAt field.Time
	UpdatedAt field.Time
	DeletedAt field.Time
}

func NewTag(tag entity.DomainTag) *Tag {
	now := field.NewTime(time.Now())
	return &Tag{
		TagAttribute: TagAttribute{
			TagID:          tag.TagID(),
			TagName:        tag.TagName(),
			TagType:        tag.TagType(),
			TagPartnerID:   tag.PartnerInternalID(),
			IsArchived:     tag.IsArchived(),
			OrganizationID: tag.OrganizationID(),
		},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: field.NewNullTime(),
	}
}

func (t *Tag) TagID() field.String {
	return t.TagAttribute.TagID
}

func (t *Tag) TagName() field.String {
	return t.TagAttribute.TagName
}

func (t *Tag) TagType() field.String {
	return t.TagAttribute.TagType
}

func (t *Tag) IsArchived() field.Boolean {
	return t.TagAttribute.IsArchived
}

func (t *Tag) PartnerInternalID() field.String {
	return t.TagAttribute.TagPartnerID
}

func (t *Tag) OrganizationID() field.String {
	return t.TagAttribute.OrganizationID
}

func (t *Tag) FieldMap() ([]string, []interface{}) {
	return []string{
			"user_tag_id",
			"user_tag_name",
			"user_tag_type",
			"is_archived",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
			"user_tag_partner_id",
		}, []interface{}{
			&t.TagAttribute.TagID,
			&t.TagAttribute.TagName,
			&t.TagAttribute.TagType,
			&t.TagAttribute.IsArchived,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.DeletedAt,
			&t.TagAttribute.OrganizationID,
			&t.TagAttribute.TagPartnerID,
		}
}

func (*Tag) TableName() string {
	return "user_tag"
}

func (r *DomainTagRepo) GetByIDs(ctx context.Context, db database.QueryExecer, tagIDs []string) (entity.DomainTags, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainTagRepo.GetByIDs")
	defer span.End()

	if len(tagIDs) == 0 {
		return entity.DomainTags{}, nil
	}

	var tags []entity.DomainTag
	tag := NewTag(entity.EmptyDomainTag{})
	fieldNames, _ := tag.FieldMap()
	query := fmt.Sprintf(
		`
		SELECT %s
		FROM %s
		WHERE
			user_tag_id = ANY($1) AND
			deleted_at IS NULL
		`,
		strings.Join(fieldNames, ","),
		tag.TableName(),
	)
	rows, err := db.Query(ctx, query, database.TextArray(tagIDs))
	if err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "db.Query")}
	}

	defer rows.Close()
	if err := rows.Err(); err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "rows.Err")}
	}

	for rows.Next() {
		tag := NewTag(entity.EmptyDomainTag{})
		_, fields := tag.FieldMap()
		if err := rows.Scan(fields...); err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "rows.Scan")}
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (r *DomainTagRepo) GetByPartnerInternalIDs(ctx context.Context, db database.QueryExecer, userTagPartnerIDs []string) (entity.DomainTags, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainTagRepo.GetByPartnerInternalIDs")
	defer span.End()

	var tags []entity.DomainTag
	tag := NewTag(entity.EmptyDomainTag{})
	fieldNames, _ := tag.FieldMap()
	query := fmt.Sprintf(
		`
		SELECT %s
		FROM %s
		WHERE
			user_tag_partner_id = ANY($1) AND
			deleted_at IS NULL
		`,
		strings.Join(fieldNames, ","),
		tag.TableName(),
	)
	rows, err := db.Query(ctx, query, database.TextArray(userTagPartnerIDs))
	if err != nil {
		return nil, InternalError{
			RawError: fmt.Errorf("db.Query: %v", err),
		}
	}

	defer rows.Close()
	if err := rows.Err(); err != nil {
		return nil, InternalError{
			RawError: fmt.Errorf("rows.Err: %v", err),
		}
	}

	for rows.Next() {
		tag := NewTag(entity.EmptyDomainTag{})
		_, fields := tag.FieldMap()
		if err := rows.Scan(fields...); err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "rows.Scan")}
		}
		tags = append(tags, tag)
	}

	return tags, nil
}
