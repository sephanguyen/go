package commands

type VirtualClassroomCommand struct {
	CommanderID string
	LessonID    string
}

func (c *VirtualClassroomCommand) GetCommander() string {
	return c.CommanderID
}

func (c *VirtualClassroomCommand) GetLessonID() string {
	return c.LessonID
}

func (c *VirtualClassroomCommand) InitBasicData(commanderID, lessonID string) {
	c.CommanderID = commanderID
	c.LessonID = lessonID
}
