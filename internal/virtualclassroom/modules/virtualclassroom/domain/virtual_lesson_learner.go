package domain

import (
	"fmt"
)

type LessonLearners []*LessonLearner

func (l LessonLearners) IsValid() error {
	for i := range l {
		if err := l[i].IsValid(); err != nil {
			return err
		}
	}

	learnerWithCourse := make(map[string]string)
	for _, learner := range l {
		if courseID, ok := learnerWithCourse[learner.LearnerID]; ok {
			return fmt.Errorf("duplicated learner %s and course %s", learner.LearnerID, courseID)
		}
		learnerWithCourse[learner.LearnerID] = learner.CourseID
	}

	return nil
}

func (l LessonLearners) GetCourseIDsOfLessonLearners() []string {
	ids := make(map[string]bool)
	for _, learner := range l {
		ids[learner.CourseID] = true
	}

	res := make([]string, 0, len(ids))
	for id := range ids {
		res = append(res, id)
	}

	return res
}

func (l LessonLearners) GetLearnerIDs() []string {
	ids := make([]string, 0, len(l))
	for _, u := range l {
		ids = append(ids, u.LearnerID)
	}
	return ids
}

type LessonLearner struct {
	LearnerID    string
	CourseID     string
	AttendStatus StudentAttendStatus
	LocationID   string
}

func (l LessonLearner) IsValid() error {
	if len(l.LearnerID) == 0 {
		return fmt.Errorf("Lesson.Learner.LearnerID cannot be empty")
	}

	if len(l.CourseID) == 0 {
		return fmt.Errorf("Lesson.Learner.CourseID of learner %s cannot be empty", l.LearnerID)
	}

	if len(l.AttendStatus) == 0 {
		return fmt.Errorf("Lesson.Learner.AttendStatus of learner %s cannot be empty", l.LearnerID)
	}

	if len(l.LocationID) == 0 {
		return fmt.Errorf("Lesson.Learner.LocationID of learner %s cannot be empty", l.LearnerID)
	}

	return nil
}
