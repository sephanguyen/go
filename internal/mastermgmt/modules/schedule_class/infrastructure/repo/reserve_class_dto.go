package repo

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type ReserveClassDTO struct {
	ReserveClassID   pgtype.Text
	StudentID        pgtype.Text
	StudentPackageID pgtype.Text
	CourseID         pgtype.Text
	ClassID          pgtype.Text
	EffectiveDate    pgtype.Date
	UpdatedAt        pgtype.Timestamptz
	CreatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
}

type ReserveClassDTOs []*ReserveClassDTO

func (rcs *ReserveClassDTOs) Add() database.Entity {
	rc := &ReserveClassDTO{}
	*rcs = append(*rcs, rc)

	return rc
}

func (rc *ReserveClassDTO) FieldMap() ([]string, []interface{}) {
	return []string{
			"reserve_class_id",
			"student_id",
			"student_package_id",
			"course_id",
			"class_id",
			"effective_date",
			"updated_at",
			"created_at",
			"deleted_at",
		}, []interface{}{
			&rc.ReserveClassID,
			&rc.StudentID,
			&rc.StudentPackageID,
			&rc.CourseID,
			&rc.ClassID,
			&rc.EffectiveDate,
			&rc.UpdatedAt,
			&rc.CreatedAt,
			&rc.DeletedAt,
		}
}

func (rc *ReserveClassDTO) TableName() string {
	return "reserve_class"
}

func (rc *ReserveClassDTO) ToReserveClassDomain() *domain.ReserveClass {
	return &domain.ReserveClass{
		ReserveClassID:   rc.ReserveClassID.String,
		StudentID:        rc.StudentID.String,
		StudentPackageID: rc.StudentPackageID.String,
		CourseID:         rc.CourseID.String,
		ClassID:          rc.ClassID.String,
		EffectiveDate:    rc.EffectiveDate.Time,
		CreatedAt:        rc.CreatedAt.Time,
		UpdatedAt:        rc.UpdatedAt.Time,
		DeletedAt:        &rc.DeletedAt.Time,
	}
}

func NewReserveClassFromEntity(rc *domain.ReserveClass) (*ReserveClassDTO, error) {
	reserveClassDTO := &ReserveClassDTO{}
	database.AllNullEntity(reserveClassDTO)
	if err := multierr.Combine(
		reserveClassDTO.ReserveClassID.Set(rc.ReserveClassID),
		reserveClassDTO.StudentID.Set(rc.StudentID),
		reserveClassDTO.StudentPackageID.Set(rc.StudentPackageID),
		reserveClassDTO.CourseID.Set(rc.CourseID),
		reserveClassDTO.ClassID.Set(rc.ClassID),
		reserveClassDTO.EffectiveDate.Set(rc.EffectiveDate),
		reserveClassDTO.CreatedAt.Set(rc.CreatedAt),
		reserveClassDTO.UpdatedAt.Set(rc.UpdatedAt),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from reserve class entity to reserve class dto: %w", err)
	}
	if rc.DeletedAt != nil {
		if err := reserveClassDTO.DeletedAt.Set(rc.DeletedAt); err != nil {
			return nil, fmt.Errorf("could not set deleted_at: %w", err)
		}
	}
	return reserveClassDTO, nil
}
