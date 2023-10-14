package utils

import (
	"errors"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ValidateDueDateAndExpiryDate(dueDate *timestamppb.Timestamp, expiryDate *timestamppb.Timestamp) error {
	if dueDate == nil {
		return errors.New("invalid DueDate value")
	}

	if expiryDate == nil {
		return errors.New("invalid ExpiryDate value")
	}

	if dueDate.AsTime() != time.Now().UTC() && dueDate.AsTime().Before(time.Now().UTC()) {
		return errors.New("invalid date: DueDate must be today or after")
	}

	if expiryDate.AsTime() != time.Now().UTC() && expiryDate.AsTime().Before(time.Now().UTC()) {
		return errors.New("invalid date: ExpiryDate must be today or after")
	}

	if dueDate.AsTime().After(expiryDate.AsTime()) {
		return errors.New("invalid date: DueDate must be before ExpiryDate")
	}

	return nil
}
