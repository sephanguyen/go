package domain

type EventLog map[string]interface{}

func (e EventLog) CollectionName() string {
	return "event_logs"
}
