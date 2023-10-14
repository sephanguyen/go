package queries

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtuallesson/application/queries/payloads"
)

type VirtualLessonQuery struct {
	LessonmgmtDB        database.Ext
	WrapperDBConnection *support.WrapperDBConnection

	VirtualLessonRepo            infrastructure.VirtualLessonRepo
	LessonMemberRepo             infrastructure.LessonMemberRepo
	LessonTeacherRepo            infrastructure.LessonTeacherRepo
	StudentEnrollmentHistoryRepo infrastructure.StudentEnrollmentStatusHistoryRepo
	CourseClassRepo              infrastructure.CourseClassRepo
	OldClassRepo                 infrastructure.OldClassRepo
	StudentsRepo                 infrastructure.StudentsRepo
	ConfigRepo                   infrastructure.ConfigRepo
}

func (v *VirtualLessonQuery) GetLiveLessonsByLocations(ctx context.Context, payload *payloads.GetLiveLessonsByLocationsRequest) (*payloads.GetLiveLessonsByLocationsResponse, error) {
	conn, err := v.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	userID := interceptors.UserIDFromContext(ctx)
	isUserAStudent, err := v.StudentsRepo.IsUserIDAStudent(ctx, conn, userID)
	if err != nil {
		return nil, fmt.Errorf("error in StudentsRepo.IsUserIDAStudent, user %s: %w", userID, err)
	}

	params := &payloads.GetVirtualLessonsArgs{
		CourseIDs:                payload.CourseIDs,
		LocationIDs:              payload.LocationIDs,
		StartDate:                payload.StartDate,
		EndDate:                  payload.EndDate,
		LessonSchedulingStatuses: payload.LessonSchedulingStatuses,
		Limit:                    payload.Limit,
		Page:                     payload.Page,
	}

	if isUserAStudent {
		var whitelistCourses []string
		// get the whitelist course IDs from configs table
		// still uses old config table as currently being used by jprep only and maybe temporary
		if payload.GetWhitelistCourseIDs {
			resourcePath := golibs.ResourcePathFromCtx(ctx)

			configKey := "specificCourseIDsForLesson"
			configGroup := "lesson"
			configs, err := v.ConfigRepo.GetConfigWithResourcePath(ctx, conn, domain.CountryMaster, configGroup, []string{configKey}, resourcePath)
			if err != nil {
				return nil, fmt.Errorf("error in ConfigRepo.GetConfigWithResourcePath, config_key %s: %w", configKey, err)
			}
			if len(configs) > 0 {
				whitelistCourses = strings.Split(configs[0].Value, ",")
			}
		}
		// get the course IDs of student if the whitelist is empty
		if len(whitelistCourses) == 0 {
			whitelistCourses, err = v.getStudentValidCourses(ctx, userID)
			if err != nil {
				return nil, fmt.Errorf("error in getStudentValidCourses: %w", err)
			}
		}

		params.CourseIDs = whitelistCourses
		params.StudentIDs = append(params.StudentIDs, userID)
	} else {
		params.ReplaceCourseIDColumn = true
	}

	lessons, total, err := v.VirtualLessonRepo.GetVirtualLessons(ctx, conn, params)
	if err != nil {
		return nil, fmt.Errorf("error in VirtualLessonRepo.GetVirtualLessons: %w; parameters: %v", err, params)
	}

	return &payloads.GetLiveLessonsByLocationsResponse{
		Lessons: lessons,
		Total:   total,
	}, nil
}

func (v *VirtualLessonQuery) GetLearnersByLessonID(ctx context.Context, payload *payloads.GetLearnersByLessonIDArgs) (*payloads.GetLearnersByLessonIDResponse, error) {
	conn, err := v.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	students, err := v.LessonMemberRepo.GetLearnersByLessonIDWithPaging(ctx, conn, payload)
	if err != nil {
		return nil, fmt.Errorf("error in LessonMemberRepo.GetLearnersByLessonIDWithPaging, lesson %s: %w", payload.LessonID, err)
	}
	if len(students) == 0 {
		return &payloads.GetLearnersByLessonIDResponse{}, nil
	}

	lesson, err := v.VirtualLessonRepo.GetVirtualLessonOnlyByID(ctx, conn, payload.LessonID)
	if err != nil {
		return nil, fmt.Errorf("error in VirtualLessonRepo.GetVirtualLessonOnlyByID, lesson %s: %w", payload.LessonID, err)
	}

	studentIDs := make([]string, 0, len(students))
	for _, s := range students {
		studentIDs = append(studentIDs, s.UserID)
	}
	studentEnrollmentInfo, err := v.StudentEnrollmentHistoryRepo.GetStatusHistoryByStudentIDsAndLocationID(ctx, conn, studentIDs, lesson.CenterID)
	if err != nil {
		return nil, fmt.Errorf(`error in StudentEnrollmentHistoryRepo.GetStatusHistoryByStudentIDsAndLocationID: %w, parameters: lesson %s student IDs: %v`,
			err,
			payload.LessonID,
			studentIDs)
	}

	lastStudentItem := students[len(students)-1]
	return &payloads.GetLearnersByLessonIDResponse{
		StudentIDs:         studentIDs,
		StudentInfo:        studentEnrollmentInfo.GetStudentEnrollmentMap(),
		Limit:              payload.Limit,
		LastLessonCourseID: lastStudentItem.LessonID + lastStudentItem.CourseID,
		LastUserID:         lastStudentItem.UserID,
	}, nil
}

func (v *VirtualLessonQuery) getStudentValidCourses(ctx context.Context, studentID string) ([]string, error) {
	conn, err := v.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	var validCourseIDs []string

	classes, err := v.OldClassRepo.FindJoined(ctx, conn, studentID)
	if err != nil {
		return nil, fmt.Errorf("error in OldClassRepo.FindJoined, user %v: %w", studentID, err)
	}
	classIDs := classes.GetIDs()

	courseClassesMap, err := v.CourseClassRepo.FindActiveCourseClassByID(ctx, conn, classIDs)
	if err != nil {
		return nil, fmt.Errorf("error in CourseClassRepo.FindActiveCourseClassByID: %w; class IDs: %v", err, classIDs)
	}
	for _, courseClass := range courseClassesMap {
		validCourseIDs = append(validCourseIDs, courseClass...)
	}

	courseIDs, err := v.LessonMemberRepo.GetCourseAccessible(ctx, conn, studentID)
	if err != nil {
		return nil, fmt.Errorf("error in LessonMemberRepo.GetCourseAccessible, user %v: %w", studentID, err)
	}
	validCourseIDs = append(validCourseIDs, courseIDs...)

	return validCourseIDs, err
}

func (v *VirtualLessonQuery) GetLearnersByLessonIDs(ctx context.Context, lessonIDs []string) (map[string]domain.LessonLearners, error) {
	lessonLearners, err := v.LessonMemberRepo.GetLessonLearnersByLessonIDs(ctx, v.LessonmgmtDB, lessonIDs)
	if err != nil {
		return nil, fmt.Errorf("error in LessonMemberRepo.GetLessonLearnersByLessonIDs, lessons %s: %w", lessonIDs, err)
	}

	return lessonLearners, nil
}

func (v *VirtualLessonQuery) GetLessons(ctx context.Context, payload payloads.GetLessonsArgs) (lessons []domain.VirtualLesson, total uint32, offsetID string, err error) {
	var preTotal uint32

	lessons, total, offsetID, preTotal, err = v.VirtualLessonRepo.GetLessons(ctx, v.LessonmgmtDB, payload)
	if err != nil {
		err = fmt.Errorf("error in VirtualLessonRepo.GetLessons: %w, payload: %v", err, payload)
		return
	}
	if preTotal <= payload.Limit {
		offsetID = ""
	}

	lessonIDs := domain.GetLessonIDs(lessons)
	if len(lessonIDs) > 0 {
		teacherIDs, repoErr := v.LessonTeacherRepo.GetTeacherIDsOnlyByLessonIDs(ctx, v.LessonmgmtDB, lessonIDs)
		if repoErr != nil {
			err = fmt.Errorf("error in LessonTeacherRepo.GetTeacherIDsOnlyByLessonIDs: %w, lessonIDs: %s", repoErr, lessonIDs)
			return
		}
		for i := range lessons {
			lessons[i].AddTeacherIDs(teacherIDs[lessons[i].LessonID])
		}
	}

	return
}

func (v *VirtualLessonQuery) GetClassDoInfoByLessonID(ctx context.Context, lessonID string) (string, error) {
	conn, err := v.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return "", err
	}

	lesson, err := v.VirtualLessonRepo.GetVirtualLessonOnlyByID(ctx, conn, lessonID)
	if err != nil {
		return "", fmt.Errorf("error in VirtualLessonRepo.GetVirtualLessonOnlyByID, lesson %s: %w", lessonID, err)
	}

	return lesson.ClassDoLink, nil
}
