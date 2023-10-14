package jprep

import (
	"fmt"
	"time"
)

func (s *suite) ExecuteWithRetry(process func() error, waitTime time.Duration, retryTime int) error {
	var count int
	var err error
	for count <= retryTime {
		err = process()
		if err == nil {
			return err
		}
		time.Sleep(waitTime)
		count++
	}
	return err
}

func toJprepAcedemicYearID(v int) string {
	return toJprepID("ACADEMIC_YEAR", v)
}

func toJprepCourseID(v int) string {
	return toJprepID("COURSE", v)
}

func toJprepLessonID(v int) string {
	return toJprepID("LESSON", v)
}

func toJprepID(typeID string, v int) string {
	return fmt.Sprintf("JPREP_%s_%09d", typeID, v)
}
