package commands

import "github.com/manabie-com/backend/internal/mastermgmt/modules/working_hours/domain"

type ImportWorkingHoursPayload struct {
	WorkingHours []*domain.WorkingHours
}
