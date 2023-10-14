package payloads

import (
	"fmt"

	"github.com/manabie-com/backend/internal/calendar/domain/dto"
)

type GetStaffRequest struct {
	LocationID                string
	IsUsingUserBasicInfoTable bool
}

type GetStaffResponse struct {
	User []*dto.User
}

func (r *GetStaffRequest) Validate() error {
	if len(r.LocationID) == 0 {
		return fmt.Errorf("location id cannot be empty")
	}

	return nil
}

type GetStaffByLocationIDsAndNameOrEmailRequest struct {
	LocationIDs        []string
	Keyword            string
	FilteredTeacherIDs []string
	Limit              int
}

type GetStaffByLocationIDsAndNameOrEmailResponse struct {
	User []*dto.User
}

func (r *GetStaffByLocationIDsAndNameOrEmailRequest) Validate() error {
	if len(r.LocationIDs) == 0 {
		return fmt.Errorf("location id list cannot be empty")
	}

	return nil
}
