package payloads

import "time"

type GetClassroomListArg struct {
	LocationIDs []string
	LessonID    string // will ignore this lesson when checking occupied classroom
	Limit       int32
	Offset      int32
	KeyWord     string
	Timezone    string
	StartTime   time.Time
	EndTime     time.Time
}
