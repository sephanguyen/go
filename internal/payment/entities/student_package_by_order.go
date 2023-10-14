package entities

import "github.com/jackc/pgtype"

type StudentPackageByOrder struct {
	StudentPackageID                    pgtype.Text
	StudentID                           pgtype.Text
	PackageID                           pgtype.Text
	StartAt                             pgtype.Timestamptz
	EndAt                               pgtype.Timestamptz
	Properties                          pgtype.JSONB
	IsActive                            pgtype.Bool
	UpdatedAt                           pgtype.Timestamptz
	CreatedAt                           pgtype.Timestamptz
	DeletedAt                           pgtype.Timestamptz
	ResourcePath                        pgtype.Text
	LocationIDs                         pgtype.TextArray
	StudentSubscriptionStudentPackageID pgtype.Text
}

func (p *StudentPackageByOrder) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"student_package_id",
		"student_id",
		"package_id",
		"start_at",
		"end_at",
		"properties",
		"is_active",
		"updated_at",
		"created_at",
		"deleted_at",
		"location_ids",
		"resource_path",
		"student_subscription_student_package_id",
	}
	values = []interface{}{
		&p.StudentPackageID,
		&p.StudentID,
		&p.PackageID,
		&p.StartAt,
		&p.EndAt,
		&p.Properties,
		&p.IsActive,
		&p.UpdatedAt,
		&p.CreatedAt,
		&p.DeletedAt,
		&p.LocationIDs,
		&p.ResourcePath,
		&p.StudentSubscriptionStudentPackageID,
	}
	return
}

func (p *StudentPackageByOrder) TableName() string {
	return "student_package_by_order"
}

func (p *StudentPackageByOrder) GetProperties() (*PackageProperties, error) {
	pp := &PackageProperties{}
	err := p.Properties.AssignTo(pp)
	return pp, err
}
