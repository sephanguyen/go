package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	IStudentEventLogRepositoryReader interface {
		RetrieveStudentEventLogsByStudyPlanIdentities(context.Context, database.QueryExecer, []*repositories.StudyPlanItemIdentity) ([]*entities.StudentEventLog, error)
	}
)

// fetch the logs chunk by chunk to avoid Seq Scan
// because of too large amount of records will be over the effective_cache_size
// Postgres will choose Seq Scan rather than Index Scan
func retrieveStudentEventLogsConcurrentlyByStudyPlanItemIdentities(
	ctx context.Context,
	db database.QueryExecer,
	studyPlanItemIdentities []*repositories.StudyPlanItemIdentity,
	repo IStudentEventLogRepositoryReader,
) ([]*entities.StudentEventLog, error) {
	studentEventLogs := make([]*entities.StudentEventLog, 0)
	chunkSize := 50
	numberOfRoutines := int(math.Ceil(float64(len(studyPlanItemIdentities)) / float64(chunkSize)))
	var wg sync.WaitGroup
	wg.Add(numberOfRoutines)
	cLogs := make(chan []*entities.StudentEventLog, numberOfRoutines)
	cErrs := make(chan error, numberOfRoutines)
	defer func() {
		close(cLogs)
		close(cErrs)
	}()
	for i := 0; i < len(studyPlanItemIdentities); i += chunkSize {
		go func(i int) {
			defer wg.Done()
			t := int(math.Min(float64(i+chunkSize), float64(len(studyPlanItemIdentities))))
			logs, err := repo.RetrieveStudentEventLogsByStudyPlanIdentities(ctx, db, studyPlanItemIdentities[i:t])
			if err != nil {
				cErrs <- status.Error(codes.Internal, fmt.Errorf("retrieveStudentEventLogsConcurrentlyByStudyPlanItemIdentities: %v", err).Error())
				return
			}
			cLogs <- logs
		}(i)
	}
	wg.Wait()
	for i := 0; i < numberOfRoutines; i++ {
		select {
		case logs := <-cLogs:
			studentEventLogs = append(studentEventLogs, logs...)

		case err := <-cErrs:
			return nil, err
		}
	}
	sort.Slice(studentEventLogs, func(i, j int) bool {
		return studentEventLogs[i].CreatedAt.Time.Before(studentEventLogs[j].CreatedAt.Time)
	})
	return studentEventLogs, nil
}
