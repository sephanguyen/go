package domain

import (
	"fmt"
	"time"
)

type (
	ClassroomStatus string
)

const (
	Available ClassroomStatus = "AVAILABLE"
	InUsed    ClassroomStatus = "IN_USED"
)

type Classroom struct {
	ClassroomID     string
	Name            string
	LocationID      string
	Remarks         string
	RoomArea        string
	SeatCapacity    int32
	IsArchived      bool
	ClassroomStatus ClassroomStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

func NewClassroom(classroomID string) *Classroom {
	clr := &Classroom{
		ClassroomID:     classroomID,
		ClassroomStatus: Available,
	}
	return clr
}

func (c Classroom) IsValid() error {
	if len(c.ClassroomID) == 0 {
		return fmt.Errorf("ClassroomID cannot be empty")
	}
	return nil
}

func (c *Classroom) WithModificationTime(createdAt, updatedAt time.Time) *Classroom {
	c.CreatedAt = createdAt
	c.UpdatedAt = updatedAt
	return c
}

func (c *Classroom) WithName(name string) *Classroom {
	c.Name = name
	return c
}

func (c *Classroom) WithLocationID(locationID string) *Classroom {
	c.LocationID = locationID
	return c
}

func (c *Classroom) WithRemark(remarks string) *Classroom {
	c.Remarks = remarks
	return c
}

func (c *Classroom) WithRoomArea(area string) *Classroom {
	c.RoomArea = area
	return c
}

func (c *Classroom) WithSeatCapacity(capacity int) *Classroom {
	c.SeatCapacity = int32(capacity)
	return c
}

func (c *Classroom) WithIsArchived(isArchived bool) *Classroom {
	c.IsArchived = isArchived
	return c
}
