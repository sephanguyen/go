package domain

import "fmt"

type Zoom struct {
	ZoomID       string
	ZoomLink     string
	AccountID    string
	OccurrenceID string
}

func (z *Zoom) Empty() {
	z.ZoomID = ""
	z.ZoomLink = ""
	z.AccountID = ""
	z.OccurrenceID = ""
}

func (z *Zoom) IsEmpty() bool {
	return z == nil || z.ZoomID == ""
}

func (z *Zoom) Validate() error {
	if len(z.ZoomLink) == 0 {
		return fmt.Errorf("Lesson.Zoom.ZoomLink cannot be empty")
	}
	if len(z.AccountID) == 0 {
		return fmt.Errorf("Lesson.Zoom.AccountID cannot be empty")
	}
	return nil
}
