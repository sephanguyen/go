package constants

import v1 "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"

type (
	Frequency string
)

const (
	FrequencyOnce   Frequency = "once"
	FrequencyWeekly Frequency = "weekly"
)

var MapFrequencyToProtoBuf = map[Frequency]v1.Frequency{
	FrequencyOnce:   v1.Frequency_ONCE,
	FrequencyWeekly: v1.Frequency_WEEKLY,
}
