package mock

import "github.com/manabie-com/backend/internal/usermgmt/pkg/field"

type Grade struct {
	RandomGrade
}

type RandomGrade struct {
	GradeID           field.String
	Name              field.String
	IsArchived        field.Boolean
	PartnerInternalID field.String
	Sequence          field.Int32
	OrgranizationID   field.String
}

func (g Grade) GradeID() field.String {
	return g.RandomGrade.GradeID
}
func (g Grade) Name() field.String {
	return g.RandomGrade.Name
}
func (g Grade) IsArchived() field.Boolean {
	return g.RandomGrade.IsArchived
}
func (g Grade) PartnerInternalID() field.String {
	return g.RandomGrade.PartnerInternalID
}
func (g Grade) Sequence() field.Int32 {
	return g.RandomGrade.Sequence
}
func (g Grade) OrganizationID() field.String {
	return g.RandomGrade.OrgranizationID
}
