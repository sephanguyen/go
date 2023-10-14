package entities

import (
	"errors"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/calendar/domain/constants"
)

type DateType struct {
	DateTypeID  constants.DateTypeID
	DisplayName string
	IsArchived  bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   time.Time
}

func GetDateTypeID(id string) (constants.DateTypeID, error) {
	switch strings.ToLower(id) {
	case "regular":
		return constants.RegularDay, nil
	case "seasonal":
		return constants.SeasonalDay, nil
	case "spare":
		return constants.SpareDay, nil
	case "closed":
		return constants.ClosedDay, nil
	}

	return "", errors.New("unsupported date type id")
}
