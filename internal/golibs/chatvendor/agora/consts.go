package agora

import "time"

const (
	TokenUserExpire = 3600
	TokenAppExpire  = 3600

	AuthHeaderKey        = "Authorization"
	ContentTypeHeaderKey = "Content-Type"

	MaxUsersOnChatGroup = 1000
	RequestTimeout      = 20 * time.Second
)

type Method string

const (
	MethodGet  Method = "GET"
	MethodPost Method = "POST"
	MethodPut  Method = "PUT"
	MethodDel  Method = "DELETE"
)

type Endpoint string

const (
	Users         Endpoint = "/users"
	ChatGroups    Endpoint = "/chatgroups"
	RecallMessage Endpoint = "/messages/msg_recall"
)
