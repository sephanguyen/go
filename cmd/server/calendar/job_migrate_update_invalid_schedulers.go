package calendar

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/calendar/domain/constants"
	cld_dto "github.com/manabie-com/backend/internal/calendar/domain/dto"
	cld_repo "github.com/manabie-com/backend/internal/calendar/infrastructure/repositories"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/configurations"
	lesson_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
)

func init() {
	bootstrap.RegisterJob("update_invalid_schedulers", jobMigrateUpdateInvalidSchedulers).
		Desc("migrate update invalid scheduler_id of lessons").
		StringVar(&resourcePath, "resourcePath", "", "orgId of partner").
		StringVar(&userID, "userID", "", "userID of school admin")
}

func jobMigrateUpdateInvalidSchedulers(ctx context.Context, _ configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	zLogger := zapLogger.Sugar()
	lessonDB := rsc.DBWith("lessonmgmt")
	calendarDB := rsc.DBWith("calendar")

	lessonRepo := lesson_repo.LessonRepo{}
	calendarRepo := cld_repo.SchedulerRepo{}

	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
			ResourcePath: resourcePath,
			UserID:       userID,
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, claim)

	lessons, err := lessonRepo.GetLessonsWithInvalidSchedulerID(ctx, lessonDB)
	if err != nil {
		return fmt.Errorf("get lessons with invalid scheduler_id failed: %w", err)
	}

	createSchedulersParams := make([]*cld_dto.CreateSchedulerParamWithIdentity, 0)

	mapSchedulerLesson := make(map[string][]*lesson_repo.Lesson, 0)

	for _, lesson := range lessons {
		schedulerID := lesson.SchedulerID.String
		if _, ok := mapSchedulerLesson[schedulerID]; !ok {
			mapSchedulerLesson[schedulerID] = make([]*lesson_repo.Lesson, 0)
		}
		mapSchedulerLesson[schedulerID] = append(mapSchedulerLesson[schedulerID], lesson)
	}

	for _, m := range mapSchedulerLesson {
		freq := string(constants.FrequencyOnce)
		if len(m) > 1 {
			freq = string(constants.FrequencyWeekly)
		}

		createSchedulersParams = append(createSchedulersParams, &cld_dto.CreateSchedulerParamWithIdentity{
			ID: m[0].LessonID.String, // not important
			CreateSchedulerParam: cld_dto.CreateSchedulerParams{
				SchedulerID: m[0].SchedulerID.String,
				StartDate:   m[0].StartTime.Time,
				EndDate:     m[len(m)-1].EndTime.Time.Add(24 * time.Hour),
				Frequency:   freq,
			},
		})
	}

	_, err = calendarRepo.CreateMany(ctx, calendarDB, createSchedulersParams)

	if err != nil {
		return fmt.Errorf("create schedulers failed: %w", err)
	}
	zLogger.Infof("========== Done to update invalid schedulers on partner %s ==========", resourcePath)
	return nil
}
