package validation

import (
	"fmt"

	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/dto"
)

var (
	ErrMissingReferenceID = "error System Notification must have Reference ID value"
	ErrMissingValidFrom   = "error System Notification must have Valid From value"
	ErrMissingRecipients  = "error System Notification must have recipients"
	ErrMissingContents    = "error System Notification must have contents"
	ErrMissingURL         = "error System Notification must have URL"
)

// TODO: add validation for content and url
func ValidateSystemNotificationRequiredFields(event *dto.SystemNotification) error {
	if event.ReferenceID == "" {
		return fmt.Errorf(ErrMissingReferenceID)
	}
	// in case it's delete event, ignore other checks
	if !event.IsDeleted {
		if event.ValidFrom.IsZero() {
			return fmt.Errorf(ErrMissingValidFrom)
		}
		if len(event.Recipients) == 0 {
			return fmt.Errorf(ErrMissingRecipients)
		}
		if len(event.Content) == 0 {
			return fmt.Errorf(ErrMissingContents)
		}
		if event.URL == "" {
			return fmt.Errorf(ErrMissingURL)
		}
	}
	return nil
}
