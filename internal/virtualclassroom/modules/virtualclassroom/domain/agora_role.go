package domain

import (
	"github.com/manabie-com/backend/internal/golibs/agoratokenbuilder"
)

type AgoraRole string

const (
	RoleAttendee   AgoraRole = "ATTENDEE"
	RolePublisher  AgoraRole = "PUBLISHER"
	RoleSubscriber AgoraRole = "SUBSCRIBER"
	RoleAdmin      AgoraRole = "ADMIN"
)

var (
	AgoraRoleMap = map[AgoraRole]agoratokenbuilder.Role{
		RoleAttendee:   agoratokenbuilder.RoleAttendee,
		RolePublisher:  agoratokenbuilder.RolePublisher,
		RoleSubscriber: agoratokenbuilder.RoleSubscriber,
		RoleAdmin:      agoratokenbuilder.RoleAdmin,
	}
)
