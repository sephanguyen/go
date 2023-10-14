package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type GradeOrganizationRepo struct{}

type GradeOrganization struct {
	gradeOrganizationID field.String
	gradeID             field.String
	gradeValue          field.Int32
	organizationID      field.String
	deletedAt           field.Time
}

func NewGradeOrganization() *GradeOrganization {
	return &GradeOrganization{
		gradeOrganizationID: field.NewNullString(),
		gradeID:             field.NewNullString(),
		gradeValue:          field.NewNullInt32(),
		organizationID:      field.NewNullString(),
	}
}

func (g *GradeOrganization) GradeOrganizationID() field.String {
	return g.gradeOrganizationID
}
func (g *GradeOrganization) GradeID() field.String {
	return g.gradeID
}
func (g *GradeOrganization) GradeValue() field.Int32 {
	return g.gradeValue
}
func (g *GradeOrganization) OrganizationID() field.String {
	return g.organizationID
}

func (g *GradeOrganization) FieldMap() ([]string, []interface{}) {
	return []string{
			"grade_organization_id",
			"grade_id",
			"grade_value",
			"resource_path",
			"deleted_at",
		}, []interface{}{
			&g.gradeOrganizationID,
			&g.gradeID,
			&g.gradeValue,
			&g.organizationID,
			&g.deletedAt,
		}
}

func (g *GradeOrganization) TableName() string {
	return "grade_organization"
}

func (r *GradeOrganizationRepo) GetByGradeIDs(ctx context.Context, db database.QueryExecer, gradeIDs []string) ([]*GradeOrganization, error) {
	ctx, span := interceptors.StartSpan(ctx, "GradeOrganizationRepo.GetByGradeIDs")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE grade_id = ANY($1) and deleted_at is NULL`
	d := NewGradeOrganization()

	fieldNames, _ := d.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		d.TableName(),
	)

	rows, err := db.Query(
		ctx,
		stmt,
		database.TextArray(gradeIDs),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []*GradeOrganization
	for rows.Next() {
		item := NewGradeOrganization()

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		result = append(result, item)
	}
	return result, nil
}

func (r *GradeOrganizationRepo) GetByGradeValues(ctx context.Context, db database.QueryExecer, gradeValues []int32) ([]*GradeOrganization, error) {
	ctx, span := interceptors.StartSpan(ctx, "GradeOrganizationRepo.GetByGradeValue")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE grade_value = ANY($1) and deleted_at is NULL`
	d := NewGradeOrganization()

	fieldNames, _ := d.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		d.TableName(),
	)

	rows, err := db.Query(
		ctx,
		stmt,
		database.Int4Array(gradeValues),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []*GradeOrganization
	for rows.Next() {
		item := NewGradeOrganization()

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		result = append(result, item)
	}
	return result, nil
}
