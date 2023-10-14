package field

import "time"

type TimeWithoutTz struct {
	status Status
	value  time.Time
}

func NewUndefinedTimeWithoutTz() TimeWithoutTz {
	return TimeWithoutTz{
		status: StatusUndefined,
	}
}

func NewNullTimeWithoutTz() TimeWithoutTz {
	return TimeWithoutTz{
		status: StatusNull,
	}
}

func NewTimeWithoutTz(value time.Time) TimeWithoutTz {
	return TimeWithoutTz{
		status: StatusPresent,
		value:  value,
	}
}

func (field TimeWithoutTz) Status() Status {
	return field.status
}

func (field TimeWithoutTz) Ptr() *TimeWithoutTz {
	ptr := &field
	return ptr
}
