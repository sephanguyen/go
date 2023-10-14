package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/pkg/errors"
)

type DomainCourseRepo struct{}

type Course struct {
	CourseIDAttr        field.String
	CoursePartnerIDAttr field.String
	OrganizationIDAttr  field.String
}

func NewCourse(course entity.Course) *Course {
	return &Course{
		CourseIDAttr:        course.CourseID(),
		CoursePartnerIDAttr: course.CoursePartnerID(),
	}
}

func (course *Course) CourseID() field.String {
	return course.CourseIDAttr
}

func (course *Course) CoursePartnerID() field.String {
	return course.CoursePartnerIDAttr
}

func (course *Course) OrganizationID() field.String {
	return course.OrganizationIDAttr
}

func (course *Course) TableName() string {
	return "courses"
}

func (course *Course) FieldMap() ([]string, []interface{}) {
	return []string{
			"course_id",
			"course_partner_id",
			"resource_path",
		}, []interface{}{
			&course.CourseIDAttr,
			&course.CoursePartnerIDAttr,
			&course.OrganizationIDAttr,
		}
}

func (r *DomainCourseRepo) GetByCoursePartnerIDs(ctx context.Context, db database.QueryExecer, coursePartnerIDs []string) (entity.DomainCourses, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainCourseRepo.GetByCoursePartnerIDs")
	defer span.End()
	stmt := `SELECT %s FROM %s WHERE course_partner_id = ANY($1) and deleted_at is NULL`
	course := NewCourse(entity.DefaultDomainCourse{})

	fieldNames, _ := course.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		course.TableName(),
	)
	rows, err := db.Query(
		ctx,
		stmt,
		database.TextArray(coursePartnerIDs),
	)
	if err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "db.Query")}
	}
	defer rows.Close()
	var result []entity.DomainCourse
	for rows.Next() {
		item := NewCourse(entity.DefaultDomainCourse{})

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "rows.Scan")}
		}

		result = append(result, item)
	}
	return result, nil
}
