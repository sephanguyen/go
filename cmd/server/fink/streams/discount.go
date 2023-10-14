package streams

import (
	"github.com/manabie-com/backend/internal/golibs/constants"

	nats_org "github.com/nats-io/nats.go"
)

func GetStreamConfigDiscount() []*nats_org.StreamConfig {
	var arrStreamConfig = []*nats_org.StreamConfig{
		{
			Name:      constants.StreamUpdateStudentProduct,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{constants.SubjectUpdateStudentProductCreated},
		},
	}
	return arrStreamConfig
}
