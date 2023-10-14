package entities

import (
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type StudentPackage struct {
	ID          pgtype.Text `sql:"student_package_id,pk"`
	StudentID   pgtype.Text `sql:"student_id"`
	PackageID   pgtype.Text `sql:"package_id"`
	StartAt     pgtype.Timestamptz
	EndAt       pgtype.Timestamptz
	Properties  pgtype.JSONB
	IsActive    pgtype.Bool
	LocationIDs pgtype.TextArray
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
}

func (rcv *StudentPackage) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"student_package_id", "student_id", "package_id", "start_at", "end_at", "properties", "is_active", "location_ids", "created_at", "updated_at"}
	values = []interface{}{&rcv.ID, &rcv.StudentID, &rcv.PackageID, &rcv.StartAt, &rcv.EndAt, &rcv.Properties, &rcv.IsActive, &rcv.LocationIDs, &rcv.CreatedAt, &rcv.UpdatedAt}
	return
}

func (*StudentPackage) TableName() string {
	return "student_packages"
}

type StudentPackageProps PackageProperties

func (rcv *StudentPackage) GetProperties() (*StudentPackageProps, error) {
	p := &StudentPackageProps{}
	err := rcv.Properties.AssignTo(p)
	return p, err
}

type StudentPackages []*StudentPackage

func (s *StudentPackages) Add() database.Entity {
	e := &StudentPackage{}
	*s = append(*s, e)

	return e
}

func (rcv *StudentPackage) GetCourseIDs() ([]string, error) {
	prop, err := rcv.GetProperties()
	if err != nil {
		return nil, err
	}
	courseIDs := make([]string, 0)
	courseIDs = append(courseIDs, prop.CanDoQuiz...)
	courseIDs = append(courseIDs, prop.CanViewStudyGuide...)
	courseIDs = append(courseIDs, prop.CanWatchVideo...)

	courseIDs = golibs.Uniq(courseIDs)
	return courseIDs, nil
}
