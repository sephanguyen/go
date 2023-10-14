package commands

type ModifyStateCommand interface {
	GetCommander() string
	GetChannelID() string
	InitBasicData(commanderID, channelID string)
}
