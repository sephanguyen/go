package domain

import (
	"fmt"
	"time"
)

type AttendeeStates []*AttendeeState

func (a AttendeeStates) isValid() error {
	for _, e := range a {
		if err := e.isValid(); err != nil {
			return err
		}
	}
	return nil
}

func (a AttendeeStates) GetAttendeeIDs() []string {
	attendeeIDs := make([]string, 0, len(a))
	for _, attendee := range a {
		attendeeIDs = append(attendeeIDs, attendee.UserID)
	}

	return attendeeIDs
}

type AttendeeState struct {
	UserID           string                    // required field
	RaisingHandState *AttendeeRaisingHandState // required field
	AnnotationState  *AttendeeAnnotationState  // required field
	PollingAnswer    *AttendeePollingAnswerState
}

func (a *AttendeeState) isValid() error {
	if len(a.UserID) == 0 {
		return fmt.Errorf("user ID cannot be empty")
	}

	if a.RaisingHandState == nil {
		return fmt.Errorf("raising hand state cannot be empty")
	}

	if a.AnnotationState == nil {
		return fmt.Errorf("annotation state cannot be empty")
	}

	return nil
}

type AttendeeRaisingHandState struct {
	IsRaisingHand bool
	UpdatedAt     time.Time
}

type AttendeeAnnotationState struct {
	BeAllowed bool
	UpdatedAt time.Time
}

type AttendeePollingAnswerState struct {
	Answer    []string
	UpdatedAt time.Time
}
