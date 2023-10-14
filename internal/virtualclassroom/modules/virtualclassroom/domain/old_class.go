package domain

import "time"

type ClassStatus string
type ClassMemberStatus string

const (
	ClassStatusActive   ClassStatus = "CLASS_STATUS_ACTIVE"
	ClassStatusInactive ClassStatus = "CLASS_STATUS_INACTIVE"

	ClassMemberStatusActive   ClassMemberStatus = "CLASS_MEMBER_STATUS_ACTIVE"
	ClassMemberStatusInactive ClassMemberStatus = "CLASS_MEMBER_STATUS_INACTIVE"
)

type OldClass struct {
	ID        int32
	Name      string
	Status    string
	UpdatedAt time.Time
	CreatedAt time.Time
}

type OldClasses []*OldClass

func (o OldClasses) GetIDs() []int32 {
	ids := make([]int32, 0, len(o))
	for _, class := range o {
		ids = append(ids, class.ID)
	}
	return ids
}
