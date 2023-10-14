package http

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

var mapStudentParentRelationship = map[int32]constant.FamilyRelationship{
	1: constant.FamilyRelationshipFather,
	2: constant.FamilyRelationshipMother,
	3: constant.FamilyRelationshipGrandfather,
	4: constant.FamilyRelationshipGrandmother,
	5: constant.FamilyRelationshipUncle,
	6: constant.FamilyRelationshipAunt,
	7: constant.FamilyRelationshipOther,
}

type UpsertParentRequest struct {
	Parents []ParentProfile `json:"parents"`
}

type ParentProfile struct {
	entity.NullDomainParent
	UserIDAttr           field.String `json:"-"`
	FullNameAttr         field.String `json:"-"`
	FullNamePhoneticAttr field.String `json:"-"`
	LoginEmailAttr       field.String `json:"-"`

	ExternalUserIDAttr       field.String            `json:"external_user_id"`
	UserNameAttr             field.String            `json:"username"`
	FirstNameAttr            field.String            `json:"first_name"`
	LastNameAttr             field.String            `json:"last_name"`
	FirstNamePhoneticAttr    field.String            `json:"first_name_phonetic"`
	LastNamePhoneticAttr     field.String            `json:"last_name_phonetic"`
	EmailAttr                field.String            `json:"email"`
	PrimaryPhoneNumberAttr   field.String            `json:"primary_phone_number"`
	SecondaryPhoneNumberAttr field.String            `json:"secondary_phone_number"`
	RemarksAttr              field.String            `json:"remarks"`
	ParentTagsAttr           []field.String          `json:"parent_tags"`
	GenderAttr               field.Int32             `json:"gender"`
	ChildrenAttr             []ParentChildrenPayload `json:"children"`
}

type ParentChildrenPayload struct {
	entity.NullDomainStudentParentRelationship

	StudentIDAttr field.String `json:"_"`

	StudentEmailAttr field.String `json:"student_email"`
	RelationshipAttr field.Int32  `json:"relationship"`
}

func (child ParentChildrenPayload) StudentID() field.String {
	return child.StudentIDAttr
}

func (child ParentChildrenPayload) Relationship() field.String {
	return field.NewString(string(mapStudentParentRelationship[child.RelationshipAttr.Int32()]))
}

func (parent ParentProfile) UserID() field.String {
	return parent.UserIDAttr
}

func (parent ParentProfile) ExternalUserID() field.String {
	return parent.ExternalUserIDAttr
}

func (parent ParentProfile) UserName() field.String {
	return parent.UserNameAttr
}

func (parent ParentProfile) FirstName() field.String {
	return parent.FirstNameAttr
}

func (parent ParentProfile) LastName() field.String {
	return parent.LastNameAttr
}

func (parent ParentProfile) FullName() field.String {
	return parent.FullNameAttr
}

func (parent ParentProfile) Email() field.String {
	return parent.EmailAttr
}

func (parent ParentProfile) FirstNamePhonetic() field.String {
	return parent.FirstNamePhoneticAttr
}

func (parent ParentProfile) LastNamePhonetic() field.String {
	return parent.LastNamePhoneticAttr
}

func (parent ParentProfile) FullNamePhonetic() field.String {
	return parent.FullNamePhoneticAttr
}

func (parent ParentProfile) Remarks() field.String {
	return parent.RemarksAttr
}

func (parent ParentProfile) Gender() field.String {
	gender := parent.GenderAttr
	if field.IsPresent(gender) {
		return field.NewString(constant.UserGenderMap[int(gender.Int32())])
	}

	return field.NewNullString()
}

func (parent ParentProfile) LoginEmail() field.String {
	return parent.LoginEmailAttr
}
func (parent ParentProfile) UserRole() field.String {
	return field.NewString(string(constant.UserRoleParent))
}
