package domain

import (
	"fmt"
	"time"
)

type StudentSubscriptionAccessPaths []*StudentSubscriptionAccessPath

func (s StudentSubscriptionAccessPaths) IsValid() error {
	for i := range s {
		if err := s[i].IsValid(); err != nil {
			return err
		}
	}
	return nil
}

func (s StudentSubscriptionAccessPaths) GetSubscriptionIDs() []string {
	list := make([]string, 0, len(s))
	for _, subAccessPath := range s {
		list = append(list, subAccessPath.SubscriptionID)
	}
	return list
}

type StudentSubscriptionAccessPath struct {
	SubscriptionID string
	LocationID     string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      time.Time
}

func (s *StudentSubscriptionAccessPath) IsValid() error {
	if len(s.SubscriptionID) == 0 {
		return fmt.Errorf("SubscriptionID could not be empty")
	}

	if len(s.LocationID) == 0 {
		return fmt.Errorf("LocationID could not be empty")
	}

	return nil
}
