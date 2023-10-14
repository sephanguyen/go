package virtualclassroom

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"github.com/bxcodec/faker/v3/support/slice"
	"github.com/jackc/pgtype"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) userGetsLiveLessonsWithoutFilter(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.GetLiveLessonsByLocationsRequest{
		LocationIds: stepState.CenterIDs,
		SchedulingStatus: []cpb.LessonSchedulingStatus{
			cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
			cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED,
		},
		Pagination: &vpb.Pagination{
			Limit: 15,
			Page:  1,
		},
	}

	stepState.Request = req

	return s.userGetsLiveLessons(ctx, req)
}

func (s *suite) userGetsLiveLessonsWithFilters(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	now := time.Now()
	oneDayDuration := time.Second * 60 * 60 * 24
	startDate := now.Add(-oneDayDuration * 2)
	endDate := now.Add(oneDayDuration * 2)
	req := &vpb.GetLiveLessonsByLocationsRequest{
		LocationIds: stepState.CenterIDs,
		CourseIds:   stepState.CourseIDs,
		SchedulingStatus: []cpb.LessonSchedulingStatus{
			cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
			cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED,
		},
		From: timestamppb.New(startDate),
		To:   timestamppb.New(endDate),
		Pagination: &vpb.Pagination{
			Limit: 15,
			Page:  1,
		},
	}

	stepState.Request = req

	return s.userGetsLiveLessons(ctx, req)
}

func (s *suite) userGetsLiveLessonsWithPagingOnly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.GetLiveLessonsByLocationsRequest{
		Pagination: &vpb.Pagination{
			Limit: 15,
			Page:  1,
		},
	}

	stepState.Request = req

	return s.userGetsLiveLessons(ctx, req)
}

func (s *suite) userGetsLiveLessons(ctx context.Context, req *vpb.GetLiveLessonsByLocationsRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualLessonReaderServiceClient(s.VirtualClassroomConn).
		GetLiveLessonsByLocations(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userReceivesLiveLessonsThatMatchesWithTheFilters(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request := stepState.Request.(*vpb.GetLiveLessonsByLocationsRequest)
	response := stepState.Response.(*vpb.GetLiveLessonsByLocationsResponse)

	lessonsCount := len(response.GetLessons())
	if lessonsCount == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting live lessons but got 0")
	}

	// get all needed data to check lessons response
	lessonIDs := make([]string, 0, lessonsCount)
	for _, lesson := range response.GetLessons() {
		lessonIDs = append(lessonIDs, lesson.LessonId)
	}

	lessonCenterMap, err := s.getLessonCenterMap(ctx, lessonIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error in getting lesson centers: %w", err)
	}

	lessonSchedulingStatusMap, err := s.getLessonSchedulingStatusMap(ctx, lessonIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error in getting lesson scheduling statuses: %w", err)
	}

	reqCourseIDs := request.GetCourseIds()
	lessonCourseMap := make(map[string][]string)
	if len(reqCourseIDs) > 0 {
		lessonCourseMap, err = s.getLessonCourseMap(ctx, lessonIDs)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error in getting lesson courses: %w", err)
		}
	}

	lessonMemberMap := make(map[string][]string)
	if user == studentType {
		lessonMemberMap, err = s.getLessonMembersMap(ctx, lessonIDs)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error in getting lesson members: %w", err)
		}
	}

	// actual checking of lesson response
	for _, lesson := range response.GetLessons() {
		locationIDs := request.GetLocationIds()
		if !slice.Contains(locationIDs, lessonCenterMap[lesson.LessonId]) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s location %s does not match in filter %v", lesson.LessonId, lessonCenterMap[lesson.LessonId], locationIDs)
		}

		reqStartDate := request.GetFrom()
		reqEndDate := request.GetTo()
		if reqStartDate != nil && reqEndDate != nil {
			if lesson.StartTime.AsTime().After(reqEndDate.AsTime()) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s start time %v is found after the end time filter %v", lesson.LessonId, lesson.StartTime.AsTime(), reqEndDate.AsTime())
			}
			if lesson.EndTime.AsTime().Before(reqStartDate.AsTime()) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s end time %v is found before the start time filter %v", lesson.LessonId, lesson.EndTime.AsTime(), reqStartDate.AsTime())
			}
		}

		if lessonSchedulingStatusMap[lesson.LessonId] != cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED.String() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecting lesson %s status to be published but got %s", lesson.LessonId, lessonSchedulingStatusMap[lesson.LessonId])
		}

		if len(reqCourseIDs) > 0 && user != studentType {
			lessonCourseIDs := lessonCourseMap[lesson.LessonId]

			foundCourses := sliceutils.Filter(lessonCourseIDs,
				func(courseID string) bool {
					return slices.Contains(reqCourseIDs, courseID)
				},
			)

			if len(foundCourses) == 0 {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s does not contain any course IDs filter %v", lesson.LessonId, reqCourseIDs)
			}
		}

		if user == studentType {
			if !slices.Contains(lessonMemberMap[lesson.LessonId], stepState.CurrentUserID) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s does not have student %s", lesson.LessonId, stepState.CurrentUserID)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userReceivesLiveLessons(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := stepState.Response.(*vpb.GetLiveLessonsByLocationsResponse)

	lessonsCount := len(response.GetLessons())
	if lessonsCount == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting live lessons but got 0")
	}

	// get all needed data to check lessons response
	lessonIDs := make([]string, 0, lessonsCount)
	for _, lesson := range response.GetLessons() {
		lessonIDs = append(lessonIDs, lesson.LessonId)
	}

	lessonSchedulingStatusMap, err := s.getLessonSchedulingStatusMap(ctx, lessonIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error in getting lesson scheduling statuses: %w", err)
	}

	lessonMemberMap := make(map[string][]string)
	if user == "student" {
		lessonMemberMap, err = s.getLessonMembersMap(ctx, lessonIDs)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error in getting lesson members: %w", err)
		}
	}

	// actual checking of lesson response - only status
	for _, lesson := range response.GetLessons() {
		if lessonSchedulingStatusMap[lesson.LessonId] != cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED.String() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecting lesson %s status to be published but got %s", lesson.LessonId, lessonSchedulingStatusMap[lesson.LessonId])
		}

		if user == "student" {
			if !slices.Contains(lessonMemberMap[lesson.LessonId], stepState.CurrentUserID) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s does not have student %s", lesson.LessonId, stepState.CurrentUserID)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getLessonCenterMap(ctx context.Context, lessonIDs []string) (map[string]string, error) {
	query := `SELECT lesson_id, center_id 
				FROM lessons 
				WHERE lesson_id = ANY($1) 
				AND deleted_at IS NULL`
	rows, err := s.LessonmgmtDB.Query(ctx, query, lessonIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var lessonID, centerID pgtype.Text
	lessonCenterMap := make(map[string]string, len(lessonIDs))
	for rows.Next() {
		if err := rows.Scan(&lessonID, &centerID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		lessonCenterMap[lessonID.String] = centerID.String
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return lessonCenterMap, nil
}

func (s *suite) getLessonSchedulingStatusMap(ctx context.Context, lessonIDs []string) (map[string]string, error) {
	query := `SELECT lesson_id, scheduling_status 
				FROM lessons 
				WHERE lesson_id = ANY($1) 
				AND deleted_at IS NULL`
	rows, err := s.LessonmgmtDB.Query(ctx, query, lessonIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var lessonID, schedulingStatus pgtype.Text
	lessonStatusMap := make(map[string]string, len(lessonIDs))
	for rows.Next() {
		if err := rows.Scan(&lessonID, &schedulingStatus); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		lessonStatusMap[lessonID.String] = schedulingStatus.String
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return lessonStatusMap, nil
}

func (s *suite) getLessonCourseMap(ctx context.Context, lessonIDs []string) (map[string][]string, error) {
	query := `SELECT lesson_id, course_id 
				FROM lessons_courses 
				WHERE lesson_id = ANY($1) 
				AND deleted_at IS NULL`
	rows, err := s.LessonmgmtDB.Query(ctx, query, lessonIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var lessonID, courseID pgtype.Text
	lessonCourseMap := make(map[string][]string, len(lessonIDs))
	for rows.Next() {
		if err := rows.Scan(&lessonID, &courseID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		lessonCourseMap[lessonID.String] = append(lessonCourseMap[lessonID.String], courseID.String)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return lessonCourseMap, nil
}

func (s *suite) getLessonMembersMap(ctx context.Context, lessonIDs []string) (map[string][]string, error) {
	query := `SELECT lesson_id, user_id 
				FROM lesson_members 
				WHERE lesson_id = ANY($1) 
				AND deleted_at IS NULL`
	rows, err := s.LessonmgmtDB.Query(ctx, query, lessonIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var lessonID, userID pgtype.Text
	lessonLearnersMap := make(map[string][]string, len(lessonIDs))
	for rows.Next() {
		if err := rows.Scan(&lessonID, &userID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		lessonLearnersMap[lessonID.String] = append(lessonLearnersMap[lessonID.String], userID.String)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return lessonLearnersMap, nil
}
