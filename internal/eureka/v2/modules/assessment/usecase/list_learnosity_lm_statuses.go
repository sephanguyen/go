package usecase

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/helper"
	"github.com/manabie-com/backend/internal/golibs/learnosity"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
)

// ListLearnositySessionStatuses
// Get all completed sessions from Learnosity
// Then all learning materials in the responses are completed, otherwise uncompleted
func (a *AssessmentUsecaseImpl) ListLearnositySessionStatuses(ctx context.Context, courseID, userID string, learningMaterialIDs []string) (map[string]bool, error) {
	asmTuples := sliceutils.Map(learningMaterialIDs, func(l string) domain.Assessment {
		return domain.Assessment{CourseID: courseID, LearningMaterialID: l}
	})
	assessments, err := a.AssessmentRepo.GetManyByLMAndCourseIDs(ctx, a.DB, asmTuples)
	if err != nil {
		return nil, errors.New("AssessmentUsecase.ListLearnositySessionStatuses", err)
	}

	now := time.Now()
	security := helper.NewLearnositySecurity(ctx, a.LearnosityConfig, "localhost", now)

	pagedSize := 300 // upto 1000
	chunks := sliceutils.Chunk(assessments, pagedSize)
	userIDs := []string{userID}
	completed := []string{string(learnosity.SessionStatusCompleted)}
	asmSessionMap := make(map[string]domain.Session, len(assessments))
	for _, v := range assessments {
		asmSessionMap[v.ID] = domain.Session{
			AssessmentID:       v.ID,
			CourseID:           v.CourseID,
			LearningMaterialID: v.LearningMaterialID,
			UserID:             userID,
			Status:             domain.SessionStatusIncomplete,
		}
	}

	for _, chunk := range chunks {
		asmIDs := sliceutils.Map(chunk, func(asm domain.Assessment) string {
			return asm.ID
		})
		dataRequest := learnosity.Request{
			"activity_id": asmIDs,
			"user_id":     userIDs,
			"status":      completed,
		}
		completedSessions, err := a.LearnositySessionRepo.GetSessionStatuses(ctx, security, dataRequest)
		if err != nil {
			return nil, errors.New("AssessmentUsecase.ListLearnositySessionStatuses", err)
		}

		for _, s := range completedSessions {
			v, ok := asmSessionMap[s.AssessmentID]
			if ok {
				v.Status = domain.SessionStatusCompleted
				v.ID = s.ID
				asmSessionMap[s.AssessmentID] = v
			}
		}
	}

	status := make(map[string]bool, len(asmSessionMap))
	for _, v := range asmSessionMap {
		s, ok := status[v.LearningMaterialID]
		if ok && s {
			continue
		}
		status[v.LearningMaterialID] = v.Status == domain.SessionStatusCompleted
	}

	return status, nil
}
