package entities

import (
	"github.com/jackc/pgtype"
)

type StudentPackageOrder struct {
	ID                        pgtype.Text `sql:"student_package_order_id,pk"`
	UserID                    pgtype.Text
	OrderID                   pgtype.Text
	CourseID                  pgtype.Text
	StartAt                   pgtype.Timestamptz
	EndAt                     pgtype.Timestamptz
	StudentPackageObject      pgtype.JSONB
	StudentPackageID          pgtype.Text
	IsCurrentStudentPackage   pgtype.Bool
	CreatedAt                 pgtype.Timestamptz
	UpdatedAt                 pgtype.Timestamptz
	DeletedAt                 pgtype.Timestamptz
	FromStudentPackageOrderID pgtype.Text
	IsExecutedByCronJob       pgtype.Bool
	ExecutedError             pgtype.Text
}

func (p *StudentPackageOrder) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"student_package_order_id",
		"user_id",
		"order_id",
		"course_id",
		"start_at",
		"end_at",
		"student_package_object",
		"student_package_id",
		"is_current_student_package",
		"created_at",
		"updated_at",
		"deleted_at",
		"from_student_package_order_id",
		"is_executed_by_cronjob",
		"executed_error",
	}
	values = []interface{}{
		&p.ID,
		&p.UserID,
		&p.OrderID,
		&p.CourseID,
		&p.StartAt,
		&p.EndAt,
		&p.StudentPackageObject,
		&p.StudentPackageID,
		&p.IsCurrentStudentPackage,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.DeletedAt,
		&p.FromStudentPackageOrderID,
		&p.IsExecutedByCronJob,
		&p.ExecutedError,
	}
	return
}

func (p *StudentPackageOrder) TableName() string {
	return "student_package_order"
}

type StudentPackagePosition int

const (
	PastStudentPackage StudentPackagePosition = iota
	CurrentStudentPackage
	FutureStudentPackage
)

func (p *StudentPackageOrder) GetStudentPackageObject() (StudentPackages, error) {
	pp := StudentPackages{}
	err := p.StudentPackageObject.AssignTo(&pp)
	return pp, err
}
