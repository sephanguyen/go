package domain

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
)

type ReserveClass struct {
	ReserveClassID   string
	StudentID        string
	StudentPackageID string
	CourseID         string
	ClassID          string
	EffectiveDate    time.Time
	UpdatedAt        time.Time
	CreatedAt        time.Time
	DeletedAt        *time.Time
	ResourcePath     string
	Repo             ReserveClassRepo
}

type ReserveClasses []*ReserveClass

type ReserveClassBuilder struct {
	reserveClass *ReserveClass
}

func NewReserveClassBuilder() *ReserveClassBuilder {
	return &ReserveClassBuilder{
		reserveClass: &ReserveClass{},
	}
}

func (rc *ReserveClass) IsValid() error {
	if len(rc.StudentID) == 0 {
		return fmt.Errorf("ReserveClass.StudentID cannot be empty")
	}

	if len(rc.StudentPackageID) == 0 {
		return fmt.Errorf("ReserveClass.StudentPackageID cannot be empty")
	}

	if len(rc.CourseID) == 0 {
		return fmt.Errorf("ReserveClass.CourseID cannot be empty")
	}

	if len(rc.ClassID) == 0 {
		return fmt.Errorf("ReserveClass.ClassID cannot be empty")
	}
	if rc.EffectiveDate.IsZero() {
		return fmt.Errorf("ReserveClass.EffectiveDate cannot be empty")
	}
	if rc.CreatedAt.IsZero() {
		return fmt.Errorf("ReserveClass.CreatedAt cannot be empty")
	}
	if rc.UpdatedAt.IsZero() {
		return fmt.Errorf("ReserveClass.UpdatedAt cannot be empty")
	}

	if rc.UpdatedAt.Before(rc.CreatedAt) {
		return fmt.Errorf("ReserveClass.UpdatedAt cannot before ReserveClass.CreatedAt")
	}

	return nil
}

func (rc *ReserveClassBuilder) Build() (*ReserveClass, error) {
	if err := rc.reserveClass.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid reserve class: %w", err)
	}
	if rc.reserveClass.ReserveClassID == "" {
		rc.reserveClass.ReserveClassID = idutil.ULIDNow()
	}
	return rc.reserveClass, nil
}

func (rc *ReserveClassBuilder) WithReserveClassRepo(repo ReserveClassRepo) *ReserveClassBuilder {
	rc.reserveClass.Repo = repo
	return rc
}

func (rc *ReserveClassBuilder) WithReserveClassID(reserveClassID string) *ReserveClassBuilder {
	rc.reserveClass.ReserveClassID = reserveClassID
	if reserveClassID == "" {
		rc.reserveClass.ReserveClassID = idutil.ULIDNow()
	}
	return rc
}

func (rc *ReserveClassBuilder) WithStudentID(studentID string) *ReserveClassBuilder {
	rc.reserveClass.StudentID = studentID
	return rc
}

func (rc *ReserveClassBuilder) WithStudentPackageID(studentPackageID string) *ReserveClassBuilder {
	rc.reserveClass.StudentPackageID = studentPackageID
	return rc
}

func (rc *ReserveClassBuilder) WithCourseID(courseID string) *ReserveClassBuilder {
	rc.reserveClass.CourseID = courseID
	return rc
}

func (rc *ReserveClassBuilder) WithClassID(classID string) *ReserveClassBuilder {
	rc.reserveClass.ClassID = classID
	return rc
}

func (rc *ReserveClassBuilder) WithEffectiveDate(effectiveDate time.Time) *ReserveClassBuilder {
	rc.reserveClass.EffectiveDate = effectiveDate
	return rc
}

func (rc *ReserveClassBuilder) WithModificationTime(createdAt, updatedAt time.Time) *ReserveClassBuilder {
	rc.reserveClass.CreatedAt = createdAt
	rc.reserveClass.UpdatedAt = updatedAt
	return rc
}

func (rc *ReserveClassBuilder) WithDeletedTime(deletedAt *time.Time) *ReserveClassBuilder {
	rc.reserveClass.DeletedAt = deletedAt
	return rc
}

func (rc *ReserveClassBuilder) GetReserveClass() *ReserveClass {
	return rc.reserveClass
}
