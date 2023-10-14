package queries

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/calendar/application/queries/payloads"
	"github.com/manabie-com/backend/internal/calendar/infrastructure"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	lesson_payloads "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/queries/payloads"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
)

type LessonQueryHandler struct {
	LessonRepo          infrastructure.LessonPort
	LessonTeacherRepo   infrastructure.LessonTeacherPort
	LessonMemberRepo    infrastructure.LessonMemberPort
	LessonClassroomRepo infrastructure.LessonClassroomPort
	LessonGroupRepo     infrastructure.LessonGroupPort
	SchedulerRepo       infrastructure.SchedulerPort
	UserRepo            infrastructure.UserPort

	Env           string
	UnleashClient unleashclient.ClientInstance
}

func (l *LessonQueryHandler) GetLessonDetail(ctx context.Context, db database.QueryExecer, req *payloads.GetLessonDetailRequest) (*payloads.GetLessonDetailResponse, error) {
	isUnleashToggled, err := l.UnleashClient.IsFeatureEnabledOnOrganization("Lesson_LessonManagement_BackOffice_SwitchNewDBConnection", l.Env, golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to unleash: %w", err)
	}
	useUserBasicInfoTable := isUnleashToggled

	lesson, err := l.LessonRepo.GetLessonWithNamesByID(ctx, db, req.LessonID)
	if err != nil {
		return nil, fmt.Errorf("LessonRepo.GetLessonWithNamesByID: %w", err)
	}

	lessonIDs := []string{req.LessonID}
	lessonTeachersMap, err := l.LessonTeacherRepo.GetTeachersWithNamesByLessonIDs(ctx, db, lessonIDs, useUserBasicInfoTable)
	if err != nil {
		return nil, fmt.Errorf("LessonTeacherRepo.GetTeachersWithNamesByLessonIDs: %w", err)
	}
	lesson.AddTeachers(lessonTeachersMap[req.LessonID])

	lessonLearnersMap, err := l.LessonMemberRepo.GetLessonLearnersWithCourseAndNamesByLessonIDs(ctx, db, lessonIDs, useUserBasicInfoTable)
	if err != nil {
		return nil, fmt.Errorf("LessonMemberRepo.GetLessonLearnersWithCourseAndNamesByLessonIDs: %w", err)
	}
	if lessonLearners, isLessonLearnersPresent := lessonLearnersMap[req.LessonID]; isLessonLearnersPresent {
		learnerIDs := lessonLearners.GetLearnerIDs()

		learnerGradesMap, err := l.UserRepo.GetStudentCurrentGradeByUserIDs(ctx, db, learnerIDs, useUserBasicInfoTable)
		if err != nil {
			return nil, fmt.Errorf("UserRepo.GetStudentCurrentGradeByUserIDs: %w", err)
		}

		for _, learner := range lessonLearners {
			if grade, isLearnerGradePresent := learnerGradesMap[learner.LearnerID]; isLearnerGradePresent {
				learner.AddGrade(grade)
			}
		}

		lesson.AddLearners(lessonLearnersMap[lesson.LessonID])
	}

	lessonClassroomsMap, err := l.LessonClassroomRepo.GetLessonClassroomsWithNamesByLessonIDs(ctx, db, lessonIDs)
	if err != nil {
		return nil, fmt.Errorf("LessonClassroomRepo.GetLessonClassroomsWithNamesByLessonIDs: %w", err)
	}
	lesson.AddClassrooms(lessonClassroomsMap[req.LessonID])

	argsMedia := &lesson_domain.ListMediaByLessonArgs{
		LessonID: req.LessonID,
		Limit:    50,
	}
	lessonMedias, err := l.LessonGroupRepo.ListMediaByLessonArgs(ctx, db, argsMedia)
	if err != nil {
		return nil, fmt.Errorf("LessonGroupRepo.ListMediaByLessonArgs: %w", err)
	}
	lesson.AddMaterials(lessonMedias.GetMediaIDs())

	scheduler, err := l.SchedulerRepo.GetByID(ctx, db, lesson.SchedulerID)
	if err != nil {
		return nil, fmt.Errorf("SchedulerRepo.GetByID: %w", err)
	}

	return &payloads.GetLessonDetailResponse{
		Lesson:    lesson,
		Scheduler: scheduler,
	}, nil
}

func (l *LessonQueryHandler) GetLessonIDsForBulkStatusUpdate(ctx context.Context, db database.QueryExecer, req *payloads.GetLessonIDsForBulkStatusUpdateRequest) ([]*payloads.GetLessonIDsForBulkStatusUpdateResponse, error) {
	res := make([]*payloads.GetLessonIDsForBulkStatusUpdateResponse, 0, 2)

	queryParams := lesson_payloads.GetLessonsByLocationStatusAndDateTimeRangeArgs{
		LocationID: req.LocationID,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Timezone:   req.Timezone,
	}

	switch req.Action {
	case lesson_domain.LessonBulkActionCancel:
		{
			queryParams.LessonStatus = lesson_domain.LessonSchedulingStatusCompleted
			completedLessons, err := l.LessonRepo.GetLessonsByLocationStatusAndDateTimeRange(ctx, db, &queryParams)
			if err != nil {
				return nil, fmt.Errorf("LessonRepo.GetLessonsByLocationStatusAndDateTimeRange (completed lessons): %w", err)
			}
			completedLessonsCount := uint32(len(completedLessons))

			queryParams.LessonStatus = lesson_domain.LessonSchedulingStatusPublished
			publishedLessons, err := l.LessonRepo.GetLessonsByLocationStatusAndDateTimeRange(ctx, db, &queryParams)
			if err != nil {
				return nil, fmt.Errorf("LessonRepo.GetLessonsByLocationStatusAndDateTimeRange (published lessons): %w", err)
			}
			publishedLessonsCount := uint32(len(publishedLessons))

			res = append(res,
				&payloads.GetLessonIDsForBulkStatusUpdateResponse{
					LessonStatus:           lesson_domain.LessonSchedulingStatusCompleted,
					ModifiableLessonsCount: completedLessonsCount,
					LessonsCount:           completedLessonsCount,
					LessonIDs:              getLessonIDs(completedLessons, func(li *lesson_domain.Lesson) bool { return true }),
				},
				&payloads.GetLessonIDsForBulkStatusUpdateResponse{
					LessonStatus:           lesson_domain.LessonSchedulingStatusPublished,
					ModifiableLessonsCount: publishedLessonsCount,
					LessonsCount:           publishedLessonsCount,
					LessonIDs:              getLessonIDs(publishedLessons, func(li *lesson_domain.Lesson) bool { return true }),
				},
			)
		}
	case lesson_domain.LessonBulkActionPublish:
		{
			queryParams.LessonStatus = lesson_domain.LessonSchedulingStatusDraft
			draftLessons, err := l.LessonRepo.GetLessonsByLocationStatusAndDateTimeRange(ctx, db, &queryParams)
			if err != nil {
				return nil, fmt.Errorf("LessonRepo.GetLessonsByLocationStatusAndDateTimeRange (draft lessons): %w", err)
			}
			lessonIDs := getLessonIDs(draftLessons, func(li *lesson_domain.Lesson) bool { return true })

			if len(lessonIDs) > 0 {
				teachersMap, err := l.LessonTeacherRepo.GetTeachersByLessonIDs(ctx, db, lessonIDs)
				if err != nil {
					return nil, fmt.Errorf("LessonTeacherRepo.GetTeachersByLessonIDs: %w", err)
				}

				for _, lesson := range draftLessons {
					lesson.AddTeachers(teachersMap[lesson.LessonID])
				}
			}
			modifiableLessonIDs := getLessonIDs(draftLessons, func(li *lesson_domain.Lesson) bool { return li.CheckLessonValidForPublish() })

			res = append(res,
				&payloads.GetLessonIDsForBulkStatusUpdateResponse{
					LessonStatus:           lesson_domain.LessonSchedulingStatusDraft,
					ModifiableLessonsCount: uint32(len(modifiableLessonIDs)),
					LessonsCount:           uint32(len(draftLessons)),
					LessonIDs:              modifiableLessonIDs,
				},
			)
		}
	default:
		return nil, fmt.Errorf("unsupported action for getting lesson ids")
	}

	return res, nil
}

func getLessonIDs(lessons []*lesson_domain.Lesson, evaluate func(lessonItem *lesson_domain.Lesson) bool) []string {
	ids := make([]string, 0, len(lessons))
	for _, lessonItem := range lessons {
		if evaluate(lessonItem) {
			ids = append(ids, lessonItem.LessonID)
		}
	}
	return ids
}
