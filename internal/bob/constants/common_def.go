package constants

import (
	"math"
)

const (
	SubjectClassEventNats  = "class_event"
	SubjectLessonEventNats = "lesson_event"

	SubjectStudentLearning = "student_learning"

	SubjectAllocateStudentQuestionNats               = "allocate_student_questions"
	SubjectAllocateStudentQuestionAfter10SecondsNats = "allocate_student_questions_after_10s"
	SubjectAllocateStudentQuestionAfter30SecondsNats = "allocate_student_questions_after_30s"
	SubjectAllocateStudentQuestionAfter60SecondsNats = "allocate_student_questions_after_60s"

	StreamStudentEventLog          = "studenteventlogs"
	SubjectStudentEventLogsCreated = "StudentEventLogs.Created"
	DeliverStudentEventLog         = "deliver.student-event-logs"
	QueueStudentEventLogsCreated   = "queue-student-event-logs-created"
	DurableStudentEventLogsCreated = "durable-student-event-logs-created"

	ManabieSchool   = math.MinInt32
	ManabieCity     = math.MinInt32
	ManabieDistrict = math.MinInt32

	Staging         = "stag"
	Local           = "local"
	Production      = "prod"
	UnAssignClassID = "UNASSIGN_CLASS_ID"
)
