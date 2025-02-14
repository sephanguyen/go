package constant

type LessonUpdateType int

const (
	LessonUpdateTypeNone LessonUpdateType = iota
	LessonUpdateTypeDraftToPublished
	LessonUpdateTypePublishedToDraft
	LessonUpdateTypeChangeLessonDate
	LessonUpdateTypeChangeLocationID
	LessonUpdateTypeChangeTeacherID
	LessonUpdateTypeChangeLessonStartTime
)
