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

type DomainGradeRepo struct{}

type Grade struct {
	gradeID           field.String
	name              field.String
	isArchived        field.Boolean
	partnerInternalID field.String
	sequence          field.Int32
	organizationID    field.String
	updatedAt         field.Time
	createdAt         field.Time
	deletedAt         field.Time
}

func NewGrade(grade entity.DomainGrade) *Grade {
	now := field.NewTime(time.Now())
	return &Grade{
		gradeID:           grade.GradeID(),
		name:              grade.Name(),
		isArchived:        grade.IsArchived(),
		partnerInternalID: grade.PartnerInternalID(),
		sequence:          grade.Sequence(),
		organizationID:    grade.OrganizationID(),
		updatedAt:         now,
		createdAt:         now,
		deletedAt:         field.NewNullTime(),
	}
}

func (grade *Grade) GradeID() field.String {
	return grade.gradeID
}
func (grade *Grade) Name() field.String {
	return grade.name
}
func (grade *Grade) IsArchived() field.Boolean {
	return grade.isArchived
}
func (grade *Grade) PartnerInternalID() field.String {
	return grade.partnerInternalID
}
func (grade *Grade) Sequence() field.Int32 {
	return grade.sequence
}
func (grade *Grade) OrganizationID() field.String {
	return grade.organizationID
}

func (grade *Grade) FieldMap() ([]string, []interface{}) {
	return []string{
			"grade_id",
			"name",
			"is_archived",
			"partner_internal_id",
			"sequence",
			"resource_path",
			"updated_at",
			"created_at",
			"deleted_at",
		}, []interface{}{
			&grade.gradeID,
			&grade.name,
			&grade.isArchived,
			&grade.partnerInternalID,
			&grade.sequence,
			&grade.organizationID,
			&grade.updatedAt,
			&grade.createdAt,
			&grade.deletedAt,
		}
}

func (grade *Grade) TableName() string {
	return "grade"
}

func (r *DomainGradeRepo) GetByPartnerInternalIDs(ctx context.Context, db database.QueryExecer, partnerInternalIDs []string) ([]entity.DomainGrade, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainGradeRepo.GetByPartnerInternalIDs")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE partner_internal_id = ANY($1) and deleted_at is NULL`
	grade := NewGrade(entity.NullDomainGrade{})

	fieldNames, _ := grade.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		grade.TableName(),
	)

	rows, err := db.Query(
		ctx,
		stmt,
		database.TextArray(partnerInternalIDs),
	)
	if err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "db.Query")}
	}

	defer rows.Close()

	var result []entity.DomainGrade
	for rows.Next() {
		item := NewGrade(entity.NullDomainGrade{})

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "rows.Scan")}
		}

		result = append(result, item)
	}
	return result, nil
}

func (r *DomainGradeRepo) GetByIDs(ctx context.Context, db database.QueryExecer, ids []string) ([]entity.DomainGrade, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainGradeRepo.GetByIDs")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE grade_id = ANY($1) and deleted_at is NULL`
	grade := NewGrade(entity.NullDomainGrade{})

	fieldNames, _ := grade.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		grade.TableName(),
	)

	rows, err := db.Query(
		ctx,
		stmt,
		database.TextArray(ids),
	)
	if err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "db.Query")}
	}

	defer rows.Close()

	var result []entity.DomainGrade
	for rows.Next() {
		item := NewGrade(entity.NullDomainGrade{})

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "rows.Scan")}
		}

		result = append(result, item)
	}
	return result, nil
}
func (r *DomainGradeRepo) GetAll(ctx context.Context, db database.QueryExecer) ([]entity.DomainGrade, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainGradeRepo.GetAll")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE deleted_at is NULL`
	grade := NewGrade(entity.NullDomainGrade{})

	fieldNames, _ := grade.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		grade.TableName(),
	)

	rows, err := db.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "db.Query")}
	}

	defer rows.Close()

	var result []entity.DomainGrade
	for rows.Next() {
		item := NewGrade(entity.NullDomainGrade{})

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "rows.Scan")}
		}

		result = append(result, item)
	}
	return result, nil
}
