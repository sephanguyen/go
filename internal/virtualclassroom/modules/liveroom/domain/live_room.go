package domain

import (
	"errors"
	"time"
)

var (
	ErrNoChannelCreated = errors.New("no channel created")
	ErrChannelNotFound  = errors.New("channel not found")
	ErrNoChannelUpdated = errors.New("no channel updated")

	ErrNoLiveRoomActivityLogCreated = errors.New("did not create new live room activity log")
)

type LiveRoom struct {
	ChannelID        string
	ChannelName      string
	WhiteboardRoomID string
	EndedAt          *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

type LiveRoomBuilder struct {
	liveRoom *LiveRoom
}

func NewLiveRoom() *LiveRoomBuilder {
	return &LiveRoomBuilder{
		liveRoom: &LiveRoom{},
	}
}

func (l *LiveRoomBuilder) WithChannelID(id string) *LiveRoomBuilder {
	l.liveRoom.ChannelID = id
	return l
}

func (l *LiveRoomBuilder) WithChannelName(name string) *LiveRoomBuilder {
	l.liveRoom.ChannelName = name
	return l
}

func (l *LiveRoomBuilder) WithWhiteboardRoomID(roomID string) *LiveRoomBuilder {
	l.liveRoom.WhiteboardRoomID = roomID
	return l
}

func (l *LiveRoomBuilder) WithEndedAt(endedAt *time.Time) *LiveRoomBuilder {
	l.liveRoom.EndedAt = endedAt
	return l
}

func (l *LiveRoomBuilder) WithModifiedTime(createdAt, updatedAt time.Time) *LiveRoomBuilder {
	l.liveRoom.CreatedAt = createdAt
	l.liveRoom.UpdatedAt = updatedAt
	return l
}

func (l *LiveRoomBuilder) WithDeletedAt(deletedAt *time.Time) *LiveRoomBuilder {
	l.liveRoom.DeletedAt = deletedAt
	return l
}

func (l *LiveRoomBuilder) BuildDraft() *LiveRoom {
	return l.liveRoom
}
