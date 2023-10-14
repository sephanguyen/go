package queries

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	user_infras "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LessonQueryHandler struct {
	WrapperConnection *support.WrapperDBConnection

	// ports
	LessonRepo          infrastructure.LessonRepo
	LessonMemberRepo    infrastructure.LessonMemberRepo
	LessonTeacherRepo   infrastructure.LessonTeacherRepo
	LessonClassroomRepo infrastructure.LessonClassroomRepo
	UserRepo            user_infras.UserRepo
	UnleashClientIns    unleashclient.ClientInstance
	Env                 string
}

type RetrieveLessonsResponse struct {
	Lessons  []*domain.Lesson
	Total    uint32
	OffsetID string
	Error    error
}

func (l *LessonQueryHandler) RetrieveLesson(ctx context.Context, payload *payloads.GetLessonListArg) *RetrieveLessonsResponse {
	var (
		preTotal uint32
		lessons  []*domain.Lesson
		total    uint32
		offsetID string
		err      error
	)

	connectionDB, err := l.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return &RetrieveLessonsResponse{Error: status.Error(codes.Internal, err.Error())}
	}
	lessons, total, offsetID, preTotal, err = l.LessonRepo.Retrieve(ctx, connectionDB, payload)

	if err != nil {
		return &RetrieveLessonsResponse{Error: status.Error(codes.Internal, err.Error())}
	}

	if preTotal <= payload.Limit {
		offsetID = ""
	}

	lessonIDs := getListLessonIDs(lessons)
	teachers, err := l.LessonTeacherRepo.GetTeachersByLessonIDs(ctx, connectionDB, lessonIDs)
	if err != nil {
		return &RetrieveLessonsResponse{
			Error: status.Error(codes.Internal, fmt.Errorf("LessonRepo.GetTeacherIDsByLessonIDs: %w", err).Error()),
		}
	}
	for _, lesson := range lessons {
		lesson.AddTeachers(teachers[lesson.LessonID])
		if lesson.TeachingMethod == domain.LessonTeachingMethodIndividual {
			lesson.CourseID, lesson.ClassID = "", ""
		}
	}
	return &RetrieveLessonsResponse{
		Lessons:  lessons,
		Total:    total,
		OffsetID: offsetID,
		Error:    err,
	}
}

func getListLessonIDs(lessons []*domain.Lesson) []string {
	lessonIDs := make([]string, 0, len(lessons))
	for _, lesson := range lessons {
		lessonIDs = append(lessonIDs, lesson.LessonID)
	}
	return lessonIDs
}

func (l *LessonQueryHandler) RetrieveLessonsOnCalendar(ctx context.Context, payload *payloads.GetLessonListOnCalendarArgs) *RetrieveLessonsResponse {
	// get lesson list
	connectionDB, err := l.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return &RetrieveLessonsResponse{
			Error: status.Error(codes.Internal, err.Error()),
		}
	}
	lessons, err := l.LessonRepo.GetLessonsOnCalendar(ctx, connectionDB, payload)
	if err != nil {
		return &RetrieveLessonsResponse{
			Error: status.Error(codes.Internal, fmt.Errorf("LessonRepo.GetLessonsOnCalendar: %w", err).Error()),
		}
	}

	var (
		teachersMap         map[string]domain.LessonTeachers
		lessonLearnersMap   map[string]domain.LessonLearners
		lessonClassroomsMap map[string]domain.LessonClassrooms
	)

	// get teachers, learners, and classrooms only if there are lessons
	if len(lessons) > 0 {
		lessonIDs := getListLessonIDs(lessons)

		teachersMap, err = l.LessonTeacherRepo.GetTeachersWithNamesByLessonIDs(ctx, connectionDB, lessonIDs, true)
		if err != nil {
			return &RetrieveLessonsResponse{
				Error: status.Error(codes.Internal, fmt.Errorf("LessonTeacherRepo.GetTeachersWithNamesByLessonIDs: %w", err).Error()),
			}
		}

		lessonLearnersMap, err = l.LessonMemberRepo.GetLessonLearnersWithCourseAndNamesByLessonIDs(ctx, connectionDB, lessonIDs, true)
		if err != nil {
			return &RetrieveLessonsResponse{
				Error: status.Error(codes.Internal, fmt.Errorf("LessonMemberRepo.GetLessonLearnersWithCourseAndNamesByLessonIDs: %w", err).Error()),
			}
		}

		lessonClassroomsMap, err = l.LessonClassroomRepo.GetLessonClassroomsWithNamesByLessonIDs(ctx, connectionDB, lessonIDs)
		if err != nil {
			return &RetrieveLessonsResponse{
				Error: status.Error(codes.Internal, fmt.Errorf("LessonClassroomRepo.GetLessonClassroomsWithNamesByLessonIDs: %w", err).Error()),
			}
		}
	}

	// add teachers and lesson learners to lessons
	for _, lesson := range lessons {
		lesson.AddTeachers(teachersMap[lesson.LessonID])
		lesson.AddClassrooms(lessonClassroomsMap[lesson.LessonID])

		lessonLearners, isLessonLearnersPresent := lessonLearnersMap[lesson.LessonID]

		// get lesson learners grades for individual lesson
		if isLessonLearnersPresent {
			if lesson.TeachingMethod == domain.LessonTeachingMethodIndividual {
				learnerIDs := lessonLearners.GetLearnerIDs()

				learnerGradesMap, err := l.UserRepo.GetStudentCurrentGradeByUserIDs(ctx, connectionDB, learnerIDs)
				if err != nil {
					return &RetrieveLessonsResponse{
						Error: status.Error(codes.Internal, fmt.Errorf("UserRepo.GetStudentCurrentGradeByUserIDs: %w", err).Error()),
					}
				}

				// add the grade to the lesson learner
				for _, learner := range lessonLearners {
					if grade, isLearnerGradePresent := learnerGradesMap[learner.LearnerID]; isLearnerGradePresent {
						learner.AddGrade(grade)
					}
				}
			}
			lesson.AddLearners(lessonLearnersMap[lesson.LessonID])
		}
	}

	return &RetrieveLessonsResponse{
		Lessons: lessons,
		Error:   err,
	}
}

func (l *LessonQueryHandler) GenerateLessonCSVTemplate(ctx context.Context) (data []byte, err error) {
	isUnleashToggled, err := l.UnleashClientIns.IsFeatureEnabled("Lesson_LessonManagement_BackOffice_ImportLessonByCSVV2", l.Env)
	if err != nil {
		return nil, fmt.Errorf("l.connectToUnleash: %w", err)
	}
	conn, err := l.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	if isUnleashToggled {
		return l.LessonRepo.GenerateLessonTemplateV2(ctx, conn)
	}
	return l.LessonRepo.GenerateLessonTemplate(ctx, conn)
}
