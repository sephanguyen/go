package payloads

import "time"

type ListStudentSubScriptionsArgs struct {
	CourseIDs              []string
	SchoolID               string
	Grades                 []int32
	KeyWord                string
	Limit                  uint32
	ClassIDs               []string
	LocationIDs            []string
	StudentIDWithCourseIDs []string
	// used for filter
	StudentSubscriptionIDs []string
	GradesV2               []string
	// used for pagination
	StudentSubscriptionID string

	// used for student course duration and the lesson validation
	LessonDate time.Time
	// used for querying user_basic_info table instead of users table
}

type GetStudentCourseSubscriptions struct {
	LocationID            string
	StudentIDWithCourseID []string
}
