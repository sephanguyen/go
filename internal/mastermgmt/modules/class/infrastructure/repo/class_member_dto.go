package repo

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type ClassMember struct {
	ClassMemberID pgtype.Text
	ClassID       pgtype.Text
	UserID        pgtype.Text
	UpdatedAt     pgtype.Timestamptz
	CreatedAt     pgtype.Timestamptz
	DeletedAt     pgtype.Timestamptz
	StartDate     pgtype.Timestamptz
	EndDate       pgtype.Timestamptz
}

func (c *ClassMember) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"class_member_id", "class_id", "user_id", "created_at", "updated_at", "deleted_at", "start_date", "end_date"}
	values = []interface{}{&c.ClassMemberID, &c.ClassID, &c.UserID, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt, &c.StartDate, &c.EndDate}
	return
}

func (c *ClassMember) TableName() string {
	return "class_member"
}

func (c *ClassMember) ToClassMemberEntity() *domain.ClassMember {
	classMember := &domain.ClassMember{
		ClassMemberID: c.ClassMemberID.String,
		ClassID:       c.ClassID.String,
		UserID:        c.UserID.String,
		CreatedAt:     c.CreatedAt.Time,
		UpdatedAt:     c.UpdatedAt.Time,
		StartDate:     c.StartDate.Time,
		EndDate:       c.EndDate.Time,
	}
	if c.DeletedAt.Status == pgtype.Present {
		classMember.DeletedAt = &c.DeletedAt.Time
	}
	return classMember
}

func NewClassMemberFromEntity(c *domain.ClassMember) (*ClassMember, error) {
	classMemberDTO := &ClassMember{}
	database.AllNullEntity(classMemberDTO)
	if err := multierr.Combine(
		classMemberDTO.ClassMemberID.Set(c.ClassMemberID),
		classMemberDTO.ClassID.Set(c.ClassID),
		classMemberDTO.UserID.Set(c.UserID),
		classMemberDTO.CreatedAt.Set(c.CreatedAt),
		classMemberDTO.UpdatedAt.Set(c.UpdatedAt),
		classMemberDTO.StartDate.Set(c.StartDate),
		classMemberDTO.EndDate.Set(c.EndDate),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from class member entity to class dto: %w", err)
	}
	if c.DeletedAt != nil {
		if err := classMemberDTO.DeletedAt.Set(c.DeletedAt); err != nil {
			return nil, fmt.Errorf("could not set deleted_at: %w", err)
		}
	}
	return classMemberDTO, nil
}
