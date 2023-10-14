package streams

import (
	"github.com/manabie-com/backend/internal/golibs/constants"

	nats_org "github.com/nats-io/nats.go"
)

func GetStreamConfigEureka() []*nats_org.StreamConfig {
	var arrStreamConfig = []*nats_org.StreamConfig{
		{
			Name:      constants.StreamLearningObjectives,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"LearningObjective.*"},
		},
		{
			Name:      constants.StreamStudyPlanItems,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"StudyPlanItems.*"},
		},
		{
			Name:      constants.StreamStudentPackageEventNats,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"StudentPackage.*"},
		},
		{
			Name:      constants.StreamStudentPackageEventNatsV2,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"StudentPackageV2.*"},
		},
		{
			Name:      constants.StreamAssignments,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"Assignments.*"},
		},
		{
			Name:      constants.StreamSyncStudentPackage,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"SyncStudentPackage.*"},
		},
		{
			Name:      constants.StreamClass,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"Class.*"},
		},
	}

	return arrStreamConfig
}
