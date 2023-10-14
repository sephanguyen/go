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

type DomainStudentPackageRepo struct{}

var _ entity.DomainStudentPackage = (*StudentPackage)(nil)

type StudentPackage struct {
	StudentPackageIDAttr field.String
	StudentIDAttr        field.String
	PackageIDAttr        field.String
	StartDateAttr        field.Time
	EndDateAttr          field.Time
	IsActiveAttr         field.Boolean
	LocationIDsAttr      []string
	UpdatedAt            field.Time
	CreatedAt            field.Time
	DeletedAt            field.Time
	OrganizationIDAttr   field.String
}

func NewStudentPackage(s entity.DomainStudentPackage) *StudentPackage {
	now := field.NewTime(time.Now())
	return &StudentPackage{
		StudentPackageIDAttr: s.StudentPackageID(),
		StudentIDAttr:        s.StudentID(),
		PackageIDAttr:        s.PackageID(),
		StartDateAttr:        s.StartDate(),
		EndDateAttr:          s.EndDate(),
		IsActiveAttr:         s.IsActive(),
		LocationIDsAttr:      make([]string, 0),
		CreatedAt:            now,
		UpdatedAt:            now,
		DeletedAt:            field.NewNullTime(),
		OrganizationIDAttr:   s.OrganizationID(),
	}
}

func (s *StudentPackage) StudentPackageID() field.String {
	return s.StudentPackageIDAttr
}

func (s *StudentPackage) StartDate() field.Time {
	return s.StartDateAttr
}

func (s *StudentPackage) EndDate() field.Time {
	return s.EndDateAttr
}

func (s *StudentPackage) IsActive() field.Boolean {
	return s.IsActiveAttr
}

func (s *StudentPackage) LocationIDs() []field.String {
	locationIDs := make([]field.String, 0, len(s.LocationIDsAttr))
	for _, locationID := range s.LocationIDsAttr {
		locationIDs = append(locationIDs, field.NewString(locationID))
	}
	return locationIDs
}

func (s *StudentPackage) PackageID() field.String {
	return s.PackageIDAttr
}

func (s *StudentPackage) StudentID() field.String {
	return s.StudentIDAttr
}

func (s *StudentPackage) OrganizationID() field.String {
	return s.OrganizationIDAttr
}

func (s *StudentPackage) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_package_id",
			"student_id",
			"package_id",
			"start_at",
			"end_at",
			"is_active",
			"location_ids",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&s.StudentPackageIDAttr,
			&s.StudentIDAttr,
			&s.PackageIDAttr,
			&s.StartDateAttr,
			&s.EndDateAttr,
			&s.IsActiveAttr,
			&s.LocationIDsAttr,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.DeletedAt,
			&s.OrganizationIDAttr,
		}
}

func (*StudentPackage) TableName() string {
	return "student_packages"
}

func (d *DomainStudentPackageRepo) GetByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) (entity.DomainStudentPackages, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainStudentPackageRepo.GetByStudentIDs")
	defer span.End()

	studentPackage := &StudentPackage{}
	query := fmt.Sprintf(`
		SELECT %s 
		FROM student_packages
		WHERE
			student_id = ANY($1) AND
			deleted_at IS NULL
	`, strings.Join(database.GetFieldNames(studentPackage), ", "))

	rows, err := db.Query(ctx, query, database.TextArray(studentIDs))
	if err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "db.Query")}
	}
	defer rows.Close()

	studentPackages := make(entity.DomainStudentPackages, 0)
	for rows.Next() {
		studentPackage := NewStudentPackage(entity.DefaultDomainStudentPackage{})
		err := rows.Scan(database.GetScanFields(studentPackage, database.GetFieldNames(studentPackage))...)
		if err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "row.Scan")}
		}
		studentPackages = append(studentPackages, studentPackage)
	}

	if err := rows.Err(); err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "row.Err")}
	}
	return studentPackages, nil
}

func (d *DomainStudentPackageRepo) GetByStudentCourseAndLocationIDs(ctx context.Context, db database.QueryExecer, studentID string, courseID string, locationIDs []string) (entity.DomainStudentPackages, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainStudentPackageRepo.GetByStudentCourseAndLocationIDs")
	defer span.End()

	studentPackage := &StudentPackage{}
	query := fmt.Sprintf(`
		SELECT %s 
		FROM student_packages
		WHERE properties->'can_do_quiz' @> '["%s"]'
		AND student_id = $1
		AND ((ARRAY_LENGTH($2::TEXT[], 1) IS NULL) OR (location_ids && $2::TEXT[]));
	`, strings.Join(database.GetFieldNames(studentPackage), ", "), courseID)

	rows, err := db.Query(ctx, query, database.Text(studentID), database.TextArray(locationIDs))
	if err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "db.Query")}
	}
	defer rows.Close()

	studentPackages := make(entity.DomainStudentPackages, 0)
	for rows.Next() {
		studentPackage := NewStudentPackage(entity.DefaultDomainStudentPackage{})
		err := rows.Scan(database.GetScanFields(studentPackage, database.GetFieldNames(studentPackage))...)
		if err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "row.Scan")}
		}
		studentPackages = append(studentPackages, studentPackage)
	}

	if err := rows.Err(); err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "row.Err")}
	}
	return studentPackages, nil
}

func (d *DomainStudentPackageRepo) GetByStudentIDAndCourseID(ctx context.Context, db database.QueryExecer, studentID string, courseID string) (entity.DomainStudentPackages, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainStudentPackageRepo.GetByStudentIDAndCourseID")
	defer span.End()

	studentPackage := &StudentPackage{}
	query := fmt.Sprintf(`
		SELECT %s 
		FROM student_packages
		WHERE properties->'can_do_quiz' @> '["%s"]'
		AND student_id = $1
	`, strings.Join(database.GetFieldNames(studentPackage), ", "), courseID)

	rows, err := db.Query(ctx, query, database.Text(studentID))
	if err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "db.Query")}
	}
	defer rows.Close()

	studentPackages := make(entity.DomainStudentPackages, 0)
	for rows.Next() {
		studentPackage := NewStudentPackage(entity.DefaultDomainStudentPackage{})
		err := rows.Scan(database.GetScanFields(studentPackage, database.GetFieldNames(studentPackage))...)
		if err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "row.Scan")}
		}
		studentPackages = append(studentPackages, studentPackage)
	}

	if err := rows.Err(); err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "row.Err")}
	}
	return studentPackages, nil
}
