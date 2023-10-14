package commands

type StateModifyCommand interface {
	GetCommander() string
	GetLessonID() string
	InitBasicData(commanderID, lessonID string)
}
