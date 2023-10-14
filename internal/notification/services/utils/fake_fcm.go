package utils

type RetrievedPushNotificationMsg struct {
	Tokens []string
	Data   map[string]string
	Title  string
	Body   string
}
