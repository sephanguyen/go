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

type DomainSchoolRepo struct{}

type SchoolAttribute struct {
	ID                field.String
	Name              field.String
	NamePhonetic      field.String
	SchoolLevelID     field.String
	Address           field.String
	IsArchived        field.Boolean
	PartnerInternalID field.String
	OrganizationID    field.String
}

type School struct {
	SchoolAttribute

	CreatedAt field.Time
	UpdatedAt field.Time
	DeletedAt field.Time
}

func NewSchool(sh entity.DomainSchool) *School {
	now := field.NewTime(time.Now())
	return &School{
		SchoolAttribute: SchoolAttribute{
			ID:                sh.SchoolID(),
			SchoolLevelID:     sh.SchoolLevelID(),
			Name:              sh.Name(),
			NamePhonetic:      sh.NamePhonetic(),
			Address:           sh.Address(),
			IsArchived:        sh.IsArchived(),
			PartnerInternalID: sh.PartnerInternalID(),
			OrganizationID:    sh.OrganizationID(),
		},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: field.NewNullTime(),
	}
}

func (sh *School) SchoolID() field.String {
	return sh.SchoolAttribute.ID
}
func (sh *School) SchoolLevelID() field.String {
	return sh.SchoolAttribute.SchoolLevelID
}
func (sh *School) Name() field.String {
	return sh.SchoolAttribute.Name
}
func (sh *School) NamePhonetic() field.String {
	return sh.SchoolAttribute.NamePhonetic
}
func (sh *School) Address() field.String {
	return sh.SchoolAttribute.Address
}
func (sh *School) IsArchived() field.Boolean {
	return sh.SchoolAttribute.IsArchived
}
func (sh *School) OrganizationID() field.String {
	return sh.SchoolAttribute.OrganizationID
}
func (sh *School) PartnerInternalID() field.String {
	return sh.SchoolAttribute.PartnerInternalID
}

func (*School) TableName() string {
	return "school_info"
}

func (sh *School) FieldMap() ([]string, []interface{}) {
	return []string{
			"school_id",
			"school_name",
			"school_name_phonetic",
			"school_partner_id",
			"school_level_id",
			"address",
			"is_archived",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&sh.SchoolAttribute.ID,
			&sh.SchoolAttribute.Name,
			&sh.SchoolAttribute.NamePhonetic,
			&sh.SchoolAttribute.PartnerInternalID,
			&sh.SchoolAttribute.SchoolLevelID,
			&sh.SchoolAttribute.Address,
			&sh.SchoolAttribute.IsArchived,
			&sh.CreatedAt,
			&sh.UpdatedAt,
			&sh.DeletedAt,
			&sh.SchoolAttribute.OrganizationID,
		}
}

func (r *DomainSchoolRepo) GetByIDsAndGradeID(ctx context.Context, db database.QueryExecer, schoolIDs []string, gradeID string) (entity.DomainSchools, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainSchoolRepo.GetByGradeIDs")
	defer span.End()

	stmt := `
		SELECT si.%s FROM %s si
		INNER JOIN school_level sl ON si.school_level_id = sl.school_level_id
		INNER JOIN school_level_grade slg ON slg.school_level_id = sl.school_level_id
		WHERE slg.grade_id = $1 AND si.school_id = ANY($2) AND si.deleted_at IS NULL
	`
	school := NewSchool(entity.DefaultDomainSchool{})
	fieldNames, _ := school.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ", si."),
		school.TableName(),
	)

	rows, err := db.Query(ctx, stmt, database.Text(gradeID), database.TextArray(schoolIDs))
	if err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "db.Query")}
	}
	defer rows.Close()

	var result []entity.DomainSchool
	for rows.Next() {
		item := NewSchool(entity.DefaultDomainSchool{})

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "rows.Scan")}
		}

		result = append(result, item)
	}
	return result, nil
}

func (r *DomainSchoolRepo) GetByPartnerInternalIDs(ctx context.Context, db database.QueryExecer, partnerInternalIDs []string) (entity.DomainSchools, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainSchoolRepo.GetByPartnerInternalIDs")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE school_partner_id = ANY($1) AND deleted_at IS NULL`
	school := NewSchool(entity.DefaultDomainSchool{})

	fieldNames, _ := school.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		school.TableName(),
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

	var result []entity.DomainSchool
	for rows.Next() {
		item := NewSchool(entity.DefaultDomainSchool{})

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "rows.Scan")}
		}

		result = append(result, item)
	}
	return result, nil
}

func (r *DomainSchoolRepo) GetByIDs(ctx context.Context, db database.QueryExecer, ids []string) (entity.DomainSchools, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainSchoolRepo.GetByIDs")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE school_id = ANY($1) AND deleted_at IS NULL`
	school := NewSchool(entity.DefaultDomainSchool{})

	fieldNames, _ := school.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		school.TableName(),
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

	var result []entity.DomainSchool
	for rows.Next() {
		item := NewSchool(entity.DefaultDomainSchool{})

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "rows.Scan")}
		}

		result = append(result, item)
	}
	return result, nil
}
