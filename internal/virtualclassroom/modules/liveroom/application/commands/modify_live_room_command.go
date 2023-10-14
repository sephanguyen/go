package commands

type ModifyLiveRoomCommand struct {
	CommanderID string
	ChannelID   string
}

func (m *ModifyLiveRoomCommand) GetCommander() string {
	return m.CommanderID
}

func (m *ModifyLiveRoomCommand) GetChannelID() string {
	return m.ChannelID
}

func (m *ModifyLiveRoomCommand) InitBasicData(commanderID, channelID string) {
	m.CommanderID = commanderID
	m.ChannelID = channelID
}
