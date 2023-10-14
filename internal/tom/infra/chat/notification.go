package chat

type notificationType int

const (
	_ notificationType = iota
	conversation
)

func (n notificationType) String() string {
	return [...]string{"_", "conversation"}[n]
}

const clickAction = "FLUTTER_NOTIFICATION_CLICK"
