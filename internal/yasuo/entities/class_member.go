package entities

import (
	"github.com/manabie-com/backend/internal/bob/entities"
)

const (
	ClassMemberStatusActive   = "CLASS_MEMBER_STATUS_ACTIVE"
	ClassMemberStatusInactive = "CLASS_MEMBER_STATUS_INACTIVE"
)

type ClassMember struct {
	entities.ClassMember
}
