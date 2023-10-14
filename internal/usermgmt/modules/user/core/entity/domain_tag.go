package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

const (
	UserTagTypeStudent         = "USER_TAG_TYPE_STUDENT"
	UserTagTypeStudentDiscount = "USER_TAG_TYPE_STUDENT_DISCOUNT"
	UserTagTypeParent          = "USER_TAG_TYPE_PARENT"
	UserTagTypeParentDiscount  = "USER_TAG_TYPE_PARENT_DISCOUNT"
	UserTagTypeStaff           = "USER_TAG_TYPE_STAFF"
)

var (
	StudentTags = []string{UserTagTypeStudent, UserTagTypeStudentDiscount}
	ParentTags  = []string{UserTagTypeParent, UserTagTypeParentDiscount}
	StaffTags   = []string{UserTagTypeStaff}
)

type Tag interface {
	valueobj.HasTagID
	TagName() field.String
	TagType() field.String
	IsArchived() field.Boolean
}

type DomainTag interface {
	Tag
	valueobj.HasOrganizationID
	valueobj.HasPartnerInternalID
}

type TagWillBeDelegated struct {
	Tag
	valueobj.HasOrganizationID
	valueobj.HasPartnerInternalID
}

type EmptyDomainTag struct{}

func (e EmptyDomainTag) TagID() field.String {
	return field.NewUndefinedString()
}

func (e EmptyDomainTag) TagName() field.String {
	return field.NewNullString()
}

func (e EmptyDomainTag) TagType() field.String {
	return field.NewNullString()
}

func (e EmptyDomainTag) IsArchived() field.Boolean {
	return field.NewNullBoolean()
}

func (e EmptyDomainTag) PartnerInternalID() field.String {
	return field.NewNullString()
}

func (e EmptyDomainTag) OrganizationID() field.String {
	return field.NewNullString()
}

func IsStudentTag(t DomainTag) bool {
	return t.TagType().String() == UserTagTypeStudent ||
		t.TagType().String() == UserTagTypeStudentDiscount
}

func IsParentTag(t DomainTag) bool {
	return t.TagType().String() == UserTagTypeParent ||
		t.TagType().String() == UserTagTypeParentDiscount
}

func IsStaffTag(t DomainTag) bool {
	return t.TagType().String() == UserTagTypeStaff
}

type DomainTags []DomainTag

func (domainTags DomainTags) ToTaggedUser(user valueobj.HasUserID) DomainTaggedUsers {
	taggedUsers := make(DomainTaggedUsers, 0, len(domainTags))
	for _, tag := range domainTags {
		taggedUsers = append(taggedUsers, TaggedUserWillBeDelegated{
			HasTagID:  tag,
			HasUserID: user,
		})
	}
	return taggedUsers
}

func (domainTags DomainTags) TagIDs() []string {
	ids := make([]string, 0, len(domainTags))
	for _, tag := range domainTags {
		ids = append(ids, tag.TagID().String())
	}
	return ids
}

func (domainTags DomainTags) PartnerInternalIDs() []string {
	partnerInternalIDs := []string{}
	for _, tag := range domainTags {
		partnerInternalIDs = append(partnerInternalIDs, tag.PartnerInternalID().String())
	}
	return partnerInternalIDs
}

func (domainTags DomainTags) ContainIDs(tagIDs ...string) bool {
	existedDomainTags := map[string]struct{}{}
	for _, domainTag := range domainTags {
		existedDomainTags[domainTag.TagID().String()] = struct{}{}
	}
	for _, tagID := range tagIDs {
		if _, ok := existedDomainTags[tagID]; !ok {
			return false
		}
	}
	return true
}

func (domainTags DomainTags) ContainPartnerInternalIDs(partnerInternalIDs ...string) bool {
	existedDomainTags := map[string]struct{}{}
	for _, domainTag := range domainTags {
		existedDomainTags[domainTag.PartnerInternalID().String()] = struct{}{}
	}
	for _, partnerInternalID := range partnerInternalIDs {
		if _, ok := existedDomainTags[partnerInternalID]; !ok {
			return false
		}
	}
	return true
}
