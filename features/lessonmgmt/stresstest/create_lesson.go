package stresstest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/lessonmgmt"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
)

// Scenario_SchoolAdminCanCreateLessonWithAllRequiredFields simulate
//
//	Scenario: School admin can create a live lesson with all required fields
//	Given user signed in as school admin
//	And user get list locations by hasura
//	And user get list lessons management
//	And user get list teachers by hasura
//	And user get some student subscriptions
//	When user creates a new lesson with all required fields
//	Then returns "OK" status code
//	And the lesson was created
//	And user be redirected to list lessons management
func (s *Suite) Scenario_SchoolAdminCanCreateLessonWithAllRequiredFields(ctx context.Context) error {
	ctx = common.StepStateToContext(ctx, s.lessonSuite.CommonSuite.StepState)
	stepState := common.StepStateFromContext(ctx)
	err := s.ASignedInAsSchoolAdmin(ctx)
	if err != nil {
		return fmt.Errorf("ASignedInAsSchoolAdmin: %w", err)
	}

	_, err = s.GetLocationsByHasura(ctx, stepState.AuthToken, nil)
	if err != nil {
		return fmt.Errorf("GetLocationsByHasura: %w", err)
	}

	err = s.lessonSuite.RetrieveLowestLevelLocations(ctx)
	if err != nil {
		return fmt.Errorf("RetrieveLowestLevelLocations: %w", err)
	}
	if s.lessonSuite.CommonSuite.ResponseErr != nil {
		return fmt.Errorf("RetrieveLowestLevelLocations: %w", s.lessonSuite.CommonSuite.ResponseErr)
	}
	stepState.CenterIDs = stepState.LowestLevelLocationIDs

	_, err = s.lessonSuite.RetrieveListLessonManagement(ctx, "LESSON_TIME_FUTURE", "100", lessonmgmt.NIL_VALUE)
	if err != nil {
		return fmt.Errorf("RetrieveListLessonManagement: %w", err)
	}
	if s.lessonSuite.CommonSuite.ResponseErr != nil {
		return fmt.Errorf("RetrieveListLessonManagement: %w", s.lessonSuite.CommonSuite.ResponseErr)
	}

	ids, err := s.GetTeachersByHasura(ctx, stepState.AuthToken, stepState.CurrentSchoolID)
	if err != nil {
		return fmt.Errorf("GetTeachersByHasura: %w", err)
	}
	stepState.TeacherIDs = ids

	err = s.GetSomeStudentSubscriptions(ctx)
	if err != nil {
		return fmt.Errorf("GetSomeStudentSubscriptions: %w", err)
	}

	_, err = s.lessonSuite.CommonSuite.UserCreateALessonWithMissingFields(ctx, "materials")
	if err != nil {
		return fmt.Errorf("UserCreateALessonWithMissingFields: %w", err)
	}
	if s.lessonSuite.CommonSuite.ResponseErr != nil {
		return fmt.Errorf("UserCreateALessonWithMissingFields: %w", s.lessonSuite.CommonSuite.ResponseErr)
	}
	reqLesson := stepState.Request.(*bpb.CreateLessonRequest)

	_, err = s.lessonSuite.CommonSuite.ReturnsStatusCode(ctx, "OK")
	if err != nil {
		return fmt.Errorf("ReturnsStatusCode: %w", err)
	}

	// redirect to lesson list
	_, err = s.lessonSuite.RetrieveListLessonManagement(ctx, "LESSON_TIME_FUTURE", "10", lessonmgmt.NIL_VALUE)
	if err != nil {
		return fmt.Errorf("RetrieveListLessonManagement: %w", err)
	}
	if s.lessonSuite.CommonSuite.ResponseErr != nil {
		return fmt.Errorf("RetrieveListLessonManagement: %w", s.lessonSuite.CommonSuite.ResponseErr)
	}

	// go to detail page
	lesson, err := s.GetLessonDetailByHasura(ctx, stepState.AuthToken, stepState.CurrentLessonID)
	if err != nil {
		return fmt.Errorf("GetLessonDetailByHasura: %w", err)
	}
	if err = s.CheckCreatedLessonDetail(ctx, reqLesson, lesson); err != nil {
		return fmt.Errorf("CheckCreatedLessonDetail: %w", err)
	}

	return nil
}

func (s *Suite) CreateALiveLessonWithTeachersAndStudents(ctx context.Context, teacherIDs, studentIDs []string) (string, error) {
	// create lesson which contain all students teachers, and have some materials
	stepState := common.StepStateFromContext(ctx)
	err := s.ASignedInAsSchoolAdmin(ctx)
	if err != nil {
		return "", fmt.Errorf("ASignedInAsSchoolAdmin: %w", err)
	}

	stepState.TeacherIDs = teacherIDs
	for _, studentID := range studentIDs {
		stepState.StudentIDWithCourseID = append(
			stepState.StudentIDWithCourseID,
			studentID,
			stepState.CurrentCourseID,
		)
	}

	_, err = s.lessonSuite.CommonSuite.UpsertValidMediaList(ctx)
	if err != nil {
		return "", fmt.Errorf("UpsertValidMediaList: %w", err)
	}

	_, err = s.lessonSuite.CommonSuite.UserCreateALiveLessonWithMissingFields(ctx, "")
	if err != nil {
		return "", fmt.Errorf("UserCreateALessonWithMissingFields: %w", err)
	}
	if s.lessonSuite.CommonSuite.ResponseErr != nil {
		return "", fmt.Errorf("UserCreateALessonWithMissingFields: %w", s.lessonSuite.CommonSuite.ResponseErr)
	}

	_, err = s.lessonSuite.CommonSuite.ReturnsStatusCode(ctx, "OK")
	if err != nil {
		return "", fmt.Errorf("ReturnsStatusCode: %w", err)
	}

	resp := s.lessonSuite.CommonSuite.Response.(*bpb.CreateLessonResponse)
	return resp.Id, nil
}

func (s *Suite) GetSomeStudentSubscriptions(ctx context.Context) error {
	stepState := common.StepStateFromContext(ctx)
	_, err := s.lessonSuite.UserRetrieveStudentSubscription(ctx, 5, 0, "", "", "")
	if err != nil {
		return fmt.Errorf("lessonSuite.UserRetrieveStudentSubscription: %w", err)
	}
	if s.lessonSuite.CommonSuite.ResponseErr != nil {
		return fmt.Errorf("lessonSuite.UserRetrieveStudentSubscription: %w", s.lessonSuite.CommonSuite.ResponseErr)
	}
	subs := s.lessonSuite.CommonSuite.Response.(*bpb.RetrieveStudentSubscriptionResponse)

	for _, item := range subs.Items {
		stepState.StudentIDWithCourseID = append(
			stepState.StudentIDWithCourseID,
			item.StudentId,
			item.CourseId,
		)
	}
	common.StepStateToContext(ctx, stepState)

	return nil
}

func (s *Suite) GetLessonDetailByHasura(ctx context.Context, jwt, lessonID string) (*domain.Lesson, error) {
	// TODO: get query from deployments/helm/manabie-all-in-one/charts/bob/files/hasura/metadata/query_collections.yaml later
	body := []byte(fmt.Sprintf("{\"query\":\"query LessonByLessonIdForLessonManagement($lesson_id: String!) {\\n  lessons(where: {lesson_id: {_eq: $lesson_id}}) {\\n    lesson_id\\n    center_id\\n    lesson_group_id\\n    teaching_medium\\n    teaching_method\\n    lesson_type\\n    scheduling_status\\n    start_time\\n    end_time\\n    lessons_teachers {\\n      teacher {\\n        users {\\n          user_id\\n          name\\n          email\\n        }\\n      }\\n    }\\n    lesson_members {\\n      attendance_remark\\n      attendance_status\\n      course {\\n        course_id\\n        name\\n        subject\\n      }\\n      user {\\n        user_id\\n        name\\n        email\\n        student {\\n          current_grade\\n        }\\n      }\\n    }\\n  }\\n}\\n\",\"variables\":{\"lesson_id\":\"%s\"}}", lessonID))
	resp, err := s.st.QueryHasura(ctx, body, jwt)
	if err != nil {
		return nil, fmt.Errorf("QueryHasura: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected status 200 but got %d", resp.StatusCode)
	}

	res := struct {
		Data struct {
			Lessons []struct {
				LessonId         string    `json:"lesson_id"`
				CenterId         string    `json:"center_id"`
				LessonGroupId    string    `json:"lesson_group_id"`
				TeachingMedium   string    `json:"teaching_medium"`
				TeachingMethod   string    `json:"teaching_method"`
				LessonType       string    `json:"lesson_type"`
				SchedulingStatus string    `json:"scheduling_status"`
				StartTime        time.Time `json:"start_time"`
				EndTime          time.Time `json:"end_time"`
				LessonsTeachers  []struct {
					Teacher struct {
						Users struct {
							UserId string  `json:"user_id"`
							Name   string  `json:"name"`
							Email  *string `json:"email"`
						} `json:"users"`
					} `json:"teacher"`
				} `json:"lessons_teachers"`
				LessonMembers []struct {
					AttendanceRemark interface{} `json:"attendance_remark"`
					AttendanceStatus string      `json:"attendance_status"`
					AttendanceReason string      `json:"attendance_reason"`
					AttendanceNotice string      `json:"attendance_notice"`
					Course           struct {
						CourseId string `json:"course_id"`
						Name     string `json:"name"`
						Subject  string `json:"subject"`
					} `json:"course"`
					User struct {
						UserId  string `json:"user_id"`
						Name    string `json:"name"`
						Email   string `json:"email"`
						Student struct {
							CurrentGrade int `json:"current_grade"`
						} `json:"student"`
					} `json:"user"`
				} `json:"lesson_members"`
			} `json:"lessons"`
		} `json:"data"`
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	if len(res.Data.Lessons) == 0 {
		return nil, fmt.Errorf("could not get details of lesson %s", lessonID)
	}
	detail := res.Data.Lessons[0]
	builder := domain.NewLesson().
		WithID(detail.LessonId).
		WithLocationID(detail.CenterId).
		WithTimeRange(detail.StartTime, detail.EndTime).
		WithTeachingMedium(domain.LessonTeachingMedium(detail.TeachingMedium)).
		WithTeachingMethod(domain.LessonTeachingMethod(detail.TeachingMethod)).
		WithSchedulingStatus(domain.LessonSchedulingStatus(detail.SchedulingStatus))

	teacherIDS := make([]string, 0, len(detail.LessonsTeachers))
	for _, teacher := range detail.LessonsTeachers {
		teacherIDS = append(teacherIDS, teacher.Teacher.Users.UserId)
	}
	builder = builder.WithTeacherIDs(teacherIDS)

	for _, member := range detail.LessonMembers {
		learner := domain.NewLessonLearner(
			member.User.UserId,
			member.Course.CourseId,
			detail.CenterId,
			member.AttendanceStatus,
			member.AttendanceNotice,
			member.AttendanceReason,
			"",
		)
		builder = builder.AddLearner(learner)
	}

	return builder.BuildDraft(), nil
}

func (s *Suite) CheckCreatedLessonDetail(ctx context.Context, req *bpb.CreateLessonRequest, actual *domain.Lesson) error {
	if _, err := s.lessonSuite.ValidateLessonForCreatedRequestMGMT(ctx, actual, req); err != nil {
		return fmt.Errorf("actual created lesson detail not match: %v", err)
	}
	return nil
}
