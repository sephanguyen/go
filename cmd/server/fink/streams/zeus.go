package streams

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"

	nats_org "github.com/nats-io/nats.go"
)

func GetStreamConfigZeus() []*nats_org.StreamConfig {
	var arrStreamConfig = []*nats_org.StreamConfig{
		{
			Name:      constants.StreamActivityLog,
			Retention: nats_org.LimitsPolicy,
			Subjects:  []string{"ActivityLog.*"},
			Replicas:  3,
			MaxAge:    time.Hour * 24 * 30, // 1 month
			MaxBytes:  1024 * 1024 * 1024,  // 1GB
		},
	}
	return arrStreamConfig
}
