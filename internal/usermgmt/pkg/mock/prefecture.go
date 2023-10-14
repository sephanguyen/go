package mock

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type Prefecture struct {
	RandomPrefecture
}

type RandomPrefecture struct {
	entity.DefaultDomainPrefecture
	PrefectureID   field.String
	PrefectureCode field.String
}

func (s Prefecture) PrefectureID() field.String {
	return s.RandomPrefecture.PrefectureID
}

func (s Prefecture) PrefectureCode() field.String {
	return s.RandomPrefecture.PrefectureCode
}
