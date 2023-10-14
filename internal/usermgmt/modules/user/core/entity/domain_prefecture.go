package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type DomainPrefecture interface {
	PrefectureCode() field.String
	Country() field.String
	Name() field.String

	valueobj.HasPrefectureID
}

type DefaultDomainPrefecture struct{}

func (p DefaultDomainPrefecture) PrefectureCode() field.String {
	return field.NewNullString()
}

func (p DefaultDomainPrefecture) Country() field.String {
	return field.NewNullString()
}

func (p DefaultDomainPrefecture) Name() field.String {
	return field.NewNullString()
}

func (p DefaultDomainPrefecture) PrefectureID() field.String {
	return field.NewUndefinedString()
}

type DomainPrefectures []DomainPrefecture
