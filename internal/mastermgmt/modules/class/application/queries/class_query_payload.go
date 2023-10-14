package queries

import (
	"github.com/jackc/pgtype"
)

type GetByIds struct {
	IDs []string
}

type FindClassMemberFilter struct {
	ClassIDs []string
	Limit    uint32
	OffsetID string
	UserName string
}

type RetrieveByClassMembersFilter struct {
	SchoolID pgtype.Text

	ClassIDs      pgtype.TextArray
	StudentTagIDs pgtype.TextArray
	AllSchool     bool
	Unassigned    bool
	StudentIDs    pgtype.TextArray

	Limit  uint32
	Offset pgtype.Text
}
