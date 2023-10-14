package lessonmgmt

import (
	"context"
	"fmt"
	"sync"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/lessonmgmt/configurations"
)

var (
	schoolID   string
	schoolName string
)

func init() {
	bootstrap.RegisterJob("sync_lesson_data_to_elasticseach", syncLessonDataToElasticSearch).
		Desc("job sync lesson data to elasticsearch").
		StringVar(&schoolID, "schoolID", "", "migrate for specific school").
		StringVar(&schoolName, "schoolName", "", "migrate for specific school")
}

func syncLessonDataToElasticSearch(ctx context.Context, cfg configurations.Config, _ *bootstrap.Resources) error {
	var wg sync.WaitGroup
	msgChan := make(chan Message, 5)
	syncJob := NewSyncJob(ctx, &wg, msgChan, &cfg)
	if err := syncJob.ResetLessonIndex(&cfg); err != nil {
		return err
	}

	numberWorker := 5
	for i := 0; i < numberWorker; i++ {
		go syncJob.Subscribe()
	}
	if err := syncJob.Start(ctx); err != nil {
		return fmt.Errorf("syncJob.Start: %s", err)
	}
	wg.Wait()
	return nil
}
