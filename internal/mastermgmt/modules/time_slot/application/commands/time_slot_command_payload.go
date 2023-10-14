package commands

import "github.com/manabie-com/backend/internal/mastermgmt/modules/time_slot/domain"

type ImportTimeSlotPayload struct {
	TimeSlots []*domain.TimeSlot
}
