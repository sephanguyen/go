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

type DomainSchoolCourseRepo struct{}

type SchoolCourseAttribute struct {
	ID                field.String
	Name              field.String
	NamePhonetic      field.String
	SchoolID          field.String
	IsArchived        field.Boolean
	PartnerInternalID field.String
	OrganizationID    field.String
}

type SchoolCourse struct {
	SchoolCourseAttribute

	CreatedAt field.Time
	UpdatedAt field.Time
	DeletedAt field.Time
}

func NewSchoolCourse(sh entity.DomainSchoolCourse) *SchoolCourse {
	now := field.NewTime(time.Now())
	return &SchoolCourse{
		SchoolCourseAttribute: SchoolCourseAttribute{
			ID:                sh.SchoolCourseID(),
			SchoolID:          sh.SchoolID(),
			Name:              sh.Name(),
			NamePhonetic:      sh.NamePhonetic(),
			IsArchived:        sh.IsArchived(),
			PartnerInternalID: sh.PartnerInternalID(),
			OrganizationID:    sh.OrganizationID(),
		},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: field.NewNullTime(),
	}
}

func (sh *SchoolCourse) SchoolCourseID() field.String {
	return sh.SchoolCourseAttribute.ID
}
func (sh *SchoolCourse) SchoolID() field.String {
	return sh.SchoolCourseAttribute.SchoolID
}
func (sh *SchoolCourse) Name() field.String {
	return sh.SchoolCourseAttribute.Name
}
func (sh *SchoolCourse) NamePhonetic() field.String {
	return sh.SchoolCourseAttribute.NamePhonetic
}
func (sh *SchoolCourse) IsArchived() field.Boolean {
	return sh.SchoolCourseAttribute.IsArchived
}
func (sh *SchoolCourse) OrganizationID() field.String {
	return sh.SchoolCourseAttribute.OrganizationID
}
func (sh *SchoolCourse) PartnerInternalID() field.String {
	return sh.SchoolCourseAttribute.PartnerInternalID
}

func (*SchoolCourse) TableName() string {
	return "school_course"
}

func (sh *SchoolCourse) FieldMap() ([]string, []interface{}) {
	return []string{
			"school_course_id",
			"school_course_name",
			"school_course_name_phonetic",
			"school_course_partner_id",
			"school_id",
			"is_archived",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&sh.SchoolCourseAttribute.ID,
			&sh.SchoolCourseAttribute.Name,
			&sh.SchoolCourseAttribute.NamePhonetic,
			&sh.SchoolCourseAttribute.PartnerInternalID,
			&sh.SchoolCourseAttribute.SchoolID,
			&sh.SchoolCourseAttribute.IsArchived,
			&sh.CreatedAt,
			&sh.UpdatedAt,
			&sh.DeletedAt,
			&sh.SchoolCourseAttribute.OrganizationID,
		}
}

func (r *DomainSchoolCourseRepo) GetByPartnerInternalIDsAndSchoolIDs(ctx context.Context, db database.QueryExecer, partnerInternalIDs []string, schoolIDs []string) (entity.DomainSchoolCourses, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainSchoolCourseRepo.GetByPartnerInternalIDsAndSchoolIDs")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE school_course_partner_id = ANY($1) AND school_id = ANY($2) AND deleted_at is NULL`
	schoolCourse := NewSchoolCourse(entity.DefaultDomainSchoolCourse{})

	fieldNames, _ := schoolCourse.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		schoolCourse.TableName(),
	)

	rows, err := db.Query(
		ctx,
		stmt,
		database.TextArray(partnerInternalIDs),
		database.TextArray(schoolIDs),
	)
	if err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "db.Query")}
	}

	defer rows.Close()

	var result []entity.DomainSchoolCourse
	for rows.Next() {
		item := NewSchoolCourse(entity.DefaultDomainSchoolCourse{})

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "rows.Scan")}
		}

		result = append(result, item)
	}
	return result, nil
}

func (r *DomainSchoolCourseRepo) GetByIDs(ctx context.Context, db database.QueryExecer, ids []string) (entity.DomainSchoolCourses, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainSchoolCourseRepo.GetByIDs")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE school_course_id = ANY($1) AND deleted_at is NULL`
	schoolCourse := NewSchoolCourse(entity.DefaultDomainSchoolCourse{})

	fieldNames, _ := schoolCourse.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		schoolCourse.TableName(),
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

	var result []entity.DomainSchoolCourse
	for rows.Next() {
		item := NewSchoolCourse(entity.DefaultDomainSchoolCourse{})

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "rows.Scan")}
		}

		result = append(result, item)
	}
	return result, nil
}

func (r *DomainSchoolCourseRepo) GetByPartnerInternalIDs(ctx context.Context, db database.QueryExecer, partnerInternalIDs []string) (entity.DomainSchoolCourses, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainSchoolCourseRepo.GetByPartnerInternalIDs")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE school_course_partner_id = ANY($1) AND deleted_at IS NULL`
	schoolCourse := NewSchoolCourse(entity.DefaultDomainSchoolCourse{})

	fieldNames, _ := schoolCourse.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		schoolCourse.TableName(),
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

	var result entity.DomainSchoolCourses
	for rows.Next() {
		item := NewSchoolCourse(entity.DefaultDomainSchoolCourse{})

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "rows.Scan")}
		}

		result = append(result, item)
	}
	return result, nil
}
