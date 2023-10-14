package entities

import "github.com/jackc/pgtype"

type UpcomingStudentPackage struct {
	UpcomingStudentPackageID            pgtype.Text
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
	ExecutedError                       pgtype.Text
	IsExecutedByCronjob                 pgtype.Bool
}

func (p *UpcomingStudentPackage) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"upcoming_student_package_id",
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
		"executed_error",
		"is_executed_by_cronjob",
	}
	values = []interface{}{
		&p.UpcomingStudentPackageID,
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
		&p.ExecutedError,
		&p.IsExecutedByCronjob,
	}
	return
}

func (p *UpcomingStudentPackage) TableName() string {
	return "upcoming_student_package"
}

func (p *UpcomingStudentPackage) GetProperties() (*PackageProperties, error) {
	pp := &PackageProperties{}
	err := p.Properties.AssignTo(pp)
	return pp, err
}
