package entities

import (
	pb "github.com/manabie-com/backend/pkg/genproto/yasuo"
)

type TargetConditions struct {
	Subscription []string `json:"subscription,omitempty"`
	School       []string `json:"school,omitempty"`
	Grade        []string `json:"grade,omitempty"`
	Hub          []string `json:"hub,omitempty"`
	Province     []string `json:"province"`
	District     []string `json:"district"`
	// todo: consider set UserGroup to USER_GROUP_STUDENT as default value
	UserGroup []string `json:"user_group,omitempty"`
	// todo: consider set Country as COUNTRY_VN as default value
	Country []string `json:"country,omitempty"`
	// todo: should we remove those in this version
	Platform  []string `json:"platform,omitempty"`
	CreatedAt []string `json:"created_at,omitempty"`
	IsTester  bool     `json:"is_tester"`
}

func (o TargetConditions) IsValid() bool {
	for _, ug := range o.UserGroup {
		if pb.UserGroup_value[ug] == 0 && ug != pb.USER_GROUP_STUDENT.String() {
			return false
		}
	}

	return true
}
