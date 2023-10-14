package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type ParentAdditionalInfo struct {
	Relationship string
}

type Parent struct {
	User

	ID        pgtype.Text
	SchoolID  pgtype.Int4
	UpdatedAt pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz

	ParentAdditionalInfo *ParentAdditionalInfo
}

// FieldMap return a map of field name and pointer to field
func (e *Parent) FieldMap() ([]string, []interface{}) {
	return []string{
			"parent_id", "school_id", "updated_at", "created_at", "deleted_at",
		}, []interface{}{
			&e.ID, &e.SchoolID, &e.UpdatedAt, &e.CreatedAt, &e.DeletedAt,
		}
}

// TableName returns "students"
func (e *Parent) TableName() string {
	return "parents"
}

type Parents []*Parent

func (parents Parents) Len() int {
	return len(parents)
}

func (parents Parents) Ids() []string {
	ids := make([]string, 0, len(parents))
	for _, parent := range parents {
		ids = append(ids, parent.ID.String)
	}
	return ids
}

func (parents Parents) Emails() []string {
	emails := make([]string, 0, len(parents))
	for _, parent := range parents {
		emails = append(emails, parent.User.Email.String)
	}
	return emails
}

func (parents Parents) PhoneNumbers() []string {
	phoneNumbers := make([]string, 0, len(parents))
	for _, parent := range parents {
		if parent.PhoneNumber.Status == pgtype.Present {
			phoneNumbers = append(phoneNumbers, parent.PhoneNumber.String)
		}
	}
	return phoneNumbers
}

func (parents Parents) Users() Users {
	users := make(Users, 0, len(parents))
	for _, parent := range parents {
		users = append(users, &parent.User)
	}
	return users
}

func (parents Parents) UserGroups() ([]*UserGroup, error) {
	userGroups := make([]*UserGroup, 0, len(parents))
	for _, parent := range parents {
		userGroup := &UserGroup{}
		database.AllNullEntity(userGroup)

		err := multierr.Combine(
			userGroup.UserID.Set(parent.ID.String),
			userGroup.GroupID.Set(userGroup),
			userGroup.IsOrigin.Set(true),
			userGroup.Status.Set(UserGroupStatusActive),
		)
		if err != nil {
			return nil, err
		}
		userGroups = append(userGroups, userGroup)
	}
	return userGroups, nil
}

func (parents Parents) FindByID(id string) *Parent {
	for _, parent := range parents {
		if parent.ID.String == id {
			return parent
		}
	}
	return nil
}
