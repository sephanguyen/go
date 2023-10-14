package mock

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type Tag struct {
	entity.EmptyDomainTag

	TagIDAttr             field.String
	TagTypeAttr           field.String
	PartnerInternalIDAttr field.String
}

func NewTag(tagID, tagType string) *Tag {
	return &Tag{
		TagIDAttr:   field.NewString(tagID),
		TagTypeAttr: field.NewString(tagType),
	}
}

func (t Tag) TagID() field.String {
	return t.TagIDAttr
}

func (t Tag) TagType() field.String {
	return t.TagTypeAttr
}

func (t Tag) PartnerInternalID() field.String {
	return t.PartnerInternalIDAttr
}

type TaggedUser struct {
	entity.EmptyDomainTaggedUser

	TagIDAttr  field.String
	UserIDAttr field.String
}

func NewTaggedUser(tagID, userID string) *TaggedUser {
	return &TaggedUser{
		TagIDAttr:  field.NewString(tagID),
		UserIDAttr: field.NewString(userID),
	}
}

func (t *TaggedUser) TagID() field.String {
	return t.TagIDAttr
}

func (t *TaggedUser) UserID() field.String {
	return t.UserIDAttr
}
