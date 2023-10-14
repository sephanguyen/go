package repo

import (
	"github.com/manabie-com/backend/internal/golibs/database"
)

type CourseLocation struct {
	CourseID   string
	LocationID string
}

type CourseLocations []*CourseLocation

func (ss *CourseLocation) FieldMap() ([]string, []interface{}) {
	return []string{
			"course_id",
			"location_id",
		}, []interface{}{
			&ss.CourseID,
			&ss.LocationID,
		}
}

func (cc *CourseLocations) Add() database.Entity {
	e := &CourseLocation{}
	*cc = append(*cc, e)

	return e
}

func (ss *CourseLocation) TableName() string {
	return "lesson_student_subscriptions"
}
