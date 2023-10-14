package streams

import (
	"github.com/manabie-com/backend/internal/golibs/constants"

	nats_org "github.com/nats-io/nats.go"
)

func GetStreamConfigTimesheet() []*nats_org.StreamConfig {
	var arrStreamConfig = []*nats_org.StreamConfig{
		{
			Name:      constants.StreamTimesheetLesson,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{constants.SubjectTimesheetLesson},
		},
		{
			Name:      constants.StreamTimesheetActionLog,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{constants.SubjectTimesheetActionLog},
		},
		{
			Name:      constants.StreamTimesheetAutoCreateFlag,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{constants.SubjectTimesheetAutoCreateFlag},
		},
	}

	return arrStreamConfig
}
