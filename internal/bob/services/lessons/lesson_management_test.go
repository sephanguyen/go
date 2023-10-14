package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestLessonManagementService_RetrieveLessonsV2(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now().UTC()
	mockLessonManagement := &MockLessonManagement{}

	s := &LessonManagementService{
		RetrieveLessonsV2: mockLessonManagement.RetrieveLessonsV2,
	}
	var schoolIDs pgtype.Int4Array
	_ = schoolIDs.Set([]int{5})

	courses := pgtype.TextArray{}
	_ = courses.Set([]string{"course-1"})

	teachers := pgtype.TextArray{}
	_ = teachers.Set([]string{"teacher-1"})

	students := pgtype.TextArray{}
	_ = students.Set([]string{"student-1"})

	centers := pgtype.TextArray{}
	_ = centers.Set([]string{"center-1"})

	locationIds := []string{"center-1", "center-2"}
	locations := pgtype.TextArray{}
	_ = locations.Set(locationIds)

	classIds := []string{"class-1", "class-2"}

	someErrors := status.Error(codes.Internal, fmt.Errorf("Some errors").Error())
	testCases := []TestCase{
		{
			name: "School Admin get list lessons future successfully with filter",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &bpb.RetrieveLessonsRequestV2{
				Paging:      &cpb.Paging{Limit: 2},
				LessonTime:  bpb.LessonTime(bpb.LessonTime_value["LESSON_TIME_FUTURE"]),
				CurrentTime: timestamppb.New(now),
				Keyword:     "Lesson Name",
				Filter: &bpb.RetrieveLessonsFilterV2{
					DateOfWeeks: []cpb.DateOfWeek{
						cpb.DateOfWeek_DATE_OF_WEEK_SUNDAY,
						cpb.DateOfWeek_DATE_OF_WEEK_MONDAY,
						cpb.DateOfWeek_DATE_OF_WEEK_TUESDAY,
						cpb.DateOfWeek_DATE_OF_WEEK_WEDNESDAY,
						cpb.DateOfWeek_DATE_OF_WEEK_THURSDAY,
						cpb.DateOfWeek_DATE_OF_WEEK_FRIDAY,
						cpb.DateOfWeek_DATE_OF_WEEK_SATURDAY,
					},
					TimeZone:         "UTC",
					FromTime:         durationpb.New(2 * time.Hour),
					ToTime:           durationpb.New(10 * time.Hour),
					FromDate:         timestamppb.New(now.Add(-30 * time.Hour)),
					ToDate:           timestamppb.New(now.Add(30 * time.Hour)),
					TeacherIds:       []string{"teacher-1"},
					CenterIds:        []string{"center-1"},
					CourseIds:        []string{"course-1"},
					StudentIds:       []string{"student-1"},
					Grades:           []int32{5, 6},
					SchedulingStatus: []cpb.LessonSchedulingStatus{cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED, cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED},
					ClassIds:         classIds,
				},
				LocationIds: locationIds,
			},
			expectedErr: nil,
			expectedResp: &bpb.RetrieveLessonsResponseV2{
				Items: []*bpb.RetrieveLessonsResponseV2_Lesson{
					{
						Id:               "lesson-1",
						Name:             "Lesson Name",
						CenterId:         "center-1",
						StartTime:        timestamppb.New(now),
						EndTime:          timestamppb.New(now),
						TeacherIds:       []string{"teacher-1", "teacher-2"},
						TeachingMethod:   cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
						TeachingMedium:   cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE,
						CourseId:         "course-1",
						ClassId:          "class-1",
						SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED,
					},
					{
						Id:               "lesson-2",
						Name:             "Lesson Name",
						CenterId:         "center-1",
						StartTime:        timestamppb.New(now),
						EndTime:          timestamppb.New(now),
						TeacherIds:       []string{"teacher-1"},
						TeachingMethod:   cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
						TeachingMedium:   cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
						CourseId:         "course-1",
						ClassId:          "class-1",
						SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED,
					},
				},
				NextPage: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "lesson-2",
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "",
					},
				},
				TotalLesson: 99,
				TotalItems:  99,
			},
			setup: func(ctx context.Context) {
				mockLessonManagement.On("RetrieveLessonsV2", mock.Anything, &lpb.RetrieveLessonsRequest{
					Paging:      &cpb.Paging{Limit: 2},
					LessonTime:  lpb.LessonTime(bpb.LessonTime_value["LESSON_TIME_FUTURE"]),
					CurrentTime: timestamppb.New(now),
					Keyword:     "Lesson Name",
					Filter: &lpb.RetrieveLessonsFilter{
						DateOfWeeks: []cpb.DateOfWeek{
							cpb.DateOfWeek_DATE_OF_WEEK_SUNDAY,
							cpb.DateOfWeek_DATE_OF_WEEK_MONDAY,
							cpb.DateOfWeek_DATE_OF_WEEK_TUESDAY,
							cpb.DateOfWeek_DATE_OF_WEEK_WEDNESDAY,
							cpb.DateOfWeek_DATE_OF_WEEK_THURSDAY,
							cpb.DateOfWeek_DATE_OF_WEEK_FRIDAY,
							cpb.DateOfWeek_DATE_OF_WEEK_SATURDAY,
						},
						TimeZone:         "UTC",
						FromTime:         durationpb.New(2 * time.Hour),
						ToTime:           durationpb.New(10 * time.Hour),
						FromDate:         timestamppb.New(now.Add(-30 * time.Hour)),
						ToDate:           timestamppb.New(now.Add(30 * time.Hour)),
						TeacherIds:       []string{"teacher-1"},
						LocationIds:      []string{"center-1"},
						CourseIds:        []string{"course-1"},
						StudentIds:       []string{"student-1"},
						Grades:           []int32{5, 6},
						SchedulingStatus: []cpb.LessonSchedulingStatus{cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED, cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED},
						ClassIds:         classIds,
					},
					LocationIds: locationIds,
				}).Once().Return(&lpb.RetrieveLessonsResponse{
					Items: []*lpb.RetrieveLessonsResponse_Lesson{
						{
							Id:               "lesson-1",
							Name:             "Lesson Name",
							CenterId:         "center-1",
							StartTime:        timestamppb.New(now),
							EndTime:          timestamppb.New(now),
							TeacherIds:       []string{"teacher-1", "teacher-2"},
							TeachingMethod:   cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
							TeachingMedium:   cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE,
							CourseId:         "course-1",
							ClassId:          "class-1",
							SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED,
						},
						{
							Id:               "lesson-2",
							Name:             "Lesson Name",
							CenterId:         "center-1",
							StartTime:        timestamppb.New(now),
							EndTime:          timestamppb.New(now),
							TeacherIds:       []string{"teacher-1"},
							TeachingMethod:   cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
							TeachingMedium:   cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
							CourseId:         "course-1",
							ClassId:          "class-1",
							SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED,
						},
					},
					NextPage: &cpb.Paging{
						Limit: 2,
						Offset: &cpb.Paging_OffsetString{
							OffsetString: "lesson-2",
						},
					},
					PreviousPage: &cpb.Paging{
						Limit: 2,
						Offset: &cpb.Paging_OffsetString{
							OffsetString: "",
						},
					},
					TotalLesson: 99,
					TotalItems:  99,
				}, nil)
			},
		},
		{
			name: "School Admin get list lessons future successfully without filter",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &bpb.RetrieveLessonsRequestV2{
				Paging:      &cpb.Paging{Limit: 2},
				LessonTime:  bpb.LessonTime(bpb.LessonTime_value["LESSON_TIME_FUTURE"]),
				CurrentTime: timestamppb.New(now),
			},
			expectedErr: nil,
			expectedResp: &bpb.RetrieveLessonsResponseV2{
				Items: []*bpb.RetrieveLessonsResponseV2_Lesson{
					{
						Id:             "lesson-1",
						Name:           "Name",
						CenterId:       "center-1",
						StartTime:      timestamppb.New(now),
						EndTime:        timestamppb.New(now),
						TeacherIds:     []string{"teacher-1"},
						TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
						TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE,
						CourseId:       "course-1",
						ClassId:        "class-1",
					},
					{
						Id:             "lesson-2",
						Name:           "Name",
						CenterId:       "center-1",
						StartTime:      timestamppb.New(now),
						EndTime:        timestamppb.New(now),
						TeacherIds:     []string{"teacher-1"},
						TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
						TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
						CourseId:       "",
						ClassId:        "",
					},
				},
				NextPage: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "lesson-2",
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "",
					},
				},
				TotalLesson: 99,
				TotalItems:  99,
			},
			setup: func(ctx context.Context) {
				mockLessonManagement.On("RetrieveLessonsV2", mock.Anything, &lpb.RetrieveLessonsRequest{
					Paging:      &cpb.Paging{Limit: 2},
					LessonTime:  lpb.LessonTime(lpb.LessonTime_value["LESSON_TIME_FUTURE"]),
					CurrentTime: timestamppb.New(now),
				}).Once().Return(&lpb.RetrieveLessonsResponse{
					Items: []*lpb.RetrieveLessonsResponse_Lesson{
						{
							Id:             "lesson-1",
							Name:           "Name",
							CenterId:       "center-1",
							StartTime:      timestamppb.New(now),
							EndTime:        timestamppb.New(now),
							TeacherIds:     []string{"teacher-1"},
							TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
							TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE,
							CourseId:       "course-1",
							ClassId:        "class-1",
						},
						{
							Id:             "lesson-2",
							Name:           "Name",
							CenterId:       "center-1",
							StartTime:      timestamppb.New(now),
							EndTime:        timestamppb.New(now),
							TeacherIds:     []string{"teacher-1"},
							TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
							TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
							CourseId:       "",
							ClassId:        "",
						},
					},
					NextPage: &cpb.Paging{
						Limit: 2,
						Offset: &cpb.Paging_OffsetString{
							OffsetString: "lesson-2",
						},
					},
					PreviousPage: &cpb.Paging{
						Limit: 2,
						Offset: &cpb.Paging_OffsetString{
							OffsetString: "",
						},
					},
					TotalLesson: 99,
					TotalItems:  99,
				}, nil)
			},
		},
		{
			name: "Teacher get list lessons future successfully without filter",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &bpb.RetrieveLessonsRequestV2{
				Paging:      &cpb.Paging{Limit: 2},
				LessonTime:  bpb.LessonTime(bpb.LessonTime_value["LESSON_TIME_FUTURE"]),
				CurrentTime: timestamppb.New(now),
			},
			expectedErr: nil,
			expectedResp: &bpb.RetrieveLessonsResponseV2{
				Items: []*bpb.RetrieveLessonsResponseV2_Lesson{
					{
						Id:             "lesson-1",
						Name:           "Name",
						CenterId:       "center-1",
						StartTime:      timestamppb.New(now),
						EndTime:        timestamppb.New(now),
						TeacherIds:     []string{"teacher-1"},
						TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
						TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE,
						CourseId:       "course-1",
						ClassId:        "class-1",
					},
					{
						Id:             "lesson-2",
						Name:           "Name",
						CenterId:       "center-1",
						StartTime:      timestamppb.New(now),
						EndTime:        timestamppb.New(now),
						TeacherIds:     []string{"teacher-1"},
						TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
						TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
						CourseId:       "",
						ClassId:        "",
					},
				},
				NextPage: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "lesson-2",
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "",
					},
				},
				TotalLesson: 99,
				TotalItems:  99,
			},
			setup: func(ctx context.Context) {
				mockLessonManagement.On("RetrieveLessonsV2", mock.Anything, &lpb.RetrieveLessonsRequest{
					Paging:      &cpb.Paging{Limit: 2},
					LessonTime:  lpb.LessonTime(lpb.LessonTime_value["LESSON_TIME_FUTURE"]),
					CurrentTime: timestamppb.New(now),
				}).Once().Return(&lpb.RetrieveLessonsResponse{
					Items: []*lpb.RetrieveLessonsResponse_Lesson{
						{
							Id:             "lesson-1",
							Name:           "Name",
							CenterId:       "center-1",
							StartTime:      timestamppb.New(now),
							EndTime:        timestamppb.New(now),
							TeacherIds:     []string{"teacher-1"},
							TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
							TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE,
							CourseId:       "course-1",
							ClassId:        "class-1",
						},
						{
							Id:             "lesson-2",
							Name:           "Name",
							CenterId:       "center-1",
							StartTime:      timestamppb.New(now),
							EndTime:        timestamppb.New(now),
							TeacherIds:     []string{"teacher-1"},
							TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
							TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
							CourseId:       "",
							ClassId:        "",
						},
					},
					NextPage: &cpb.Paging{
						Limit: 2,
						Offset: &cpb.Paging_OffsetString{
							OffsetString: "lesson-2",
						},
					},
					PreviousPage: &cpb.Paging{
						Limit: 2,
						Offset: &cpb.Paging_OffsetString{
							OffsetString: "",
						},
					},
					TotalLesson: 99,
					TotalItems:  99,
				}, nil)
			},
		},
		{
			name: "School Admin return list lessons past successfully without filter",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &bpb.RetrieveLessonsRequestV2{
				Paging:      &cpb.Paging{Limit: 2},
				LessonTime:  bpb.LessonTime(bpb.LessonTime_value["LESSON_TIME_PAST"]),
				CurrentTime: timestamppb.New(now),
			},
			expectedErr: nil,
			expectedResp: &bpb.RetrieveLessonsResponseV2{
				Items: []*bpb.RetrieveLessonsResponseV2_Lesson{
					{
						Id:             "lesson-1",
						Name:           "Name",
						CenterId:       "center-1",
						StartTime:      timestamppb.New(now),
						EndTime:        timestamppb.New(now),
						TeacherIds:     []string{"teacher-1"},
						TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
						TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE,
						CourseId:       "course-1",
						ClassId:        "class-1",
					},
					{
						Id:             "lesson-2",
						Name:           "Name",
						CenterId:       "center-1",
						StartTime:      timestamppb.New(now),
						EndTime:        timestamppb.New(now),
						TeacherIds:     []string{"teacher-1"},
						TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
						TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
						CourseId:       "",
						ClassId:        "",
					},
				},
				NextPage: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "lesson-2",
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "",
					},
				},
				TotalLesson: 99,
				TotalItems:  99,
			},
			setup: func(ctx context.Context) {
				mockLessonManagement.On("RetrieveLessonsV2", mock.Anything, &lpb.RetrieveLessonsRequest{
					Paging:      &cpb.Paging{Limit: 2},
					LessonTime:  lpb.LessonTime(lpb.LessonTime_value["LESSON_TIME_PAST"]),
					CurrentTime: timestamppb.New(now),
				}).Once().Return(&lpb.RetrieveLessonsResponse{
					Items: []*lpb.RetrieveLessonsResponse_Lesson{
						{
							Id:             "lesson-1",
							Name:           "Name",
							CenterId:       "center-1",
							StartTime:      timestamppb.New(now),
							EndTime:        timestamppb.New(now),
							TeacherIds:     []string{"teacher-1"},
							TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
							TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE,
							CourseId:       "course-1",
							ClassId:        "class-1",
						},
						{
							Id:             "lesson-2",
							Name:           "Name",
							CenterId:       "center-1",
							StartTime:      timestamppb.New(now),
							EndTime:        timestamppb.New(now),
							TeacherIds:     []string{"teacher-1"},
							TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
							TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
							CourseId:       "",
							ClassId:        "",
						},
					},
					NextPage: &cpb.Paging{
						Limit: 2,
						Offset: &cpb.Paging_OffsetString{
							OffsetString: "lesson-2",
						},
					},
					PreviousPage: &cpb.Paging{
						Limit: 2,
						Offset: &cpb.Paging_OffsetString{
							OffsetString: "",
						},
					},
					TotalLesson: 99,
					TotalItems:  99,
				}, nil)
			},
		},
		{
			name: "Return list empty successfully",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &bpb.RetrieveLessonsRequestV2{
				Paging:      &cpb.Paging{Limit: 2},
				LessonTime:  bpb.LessonTime(bpb.LessonTime_value["LESSON_TIME_FUTURE"]),
				CurrentTime: timestamppb.New(now),
			},
			expectedErr: nil,
			expectedResp: &bpb.RetrieveLessonsResponseV2{
				Items: []*bpb.RetrieveLessonsResponseV2_Lesson{},
				NextPage: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "",
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "",
					},
				},
				TotalLesson: 0,
				TotalItems:  0,
			},
			setup: func(ctx context.Context) {
				mockLessonManagement.On("RetrieveLessonsV2", mock.Anything, &lpb.RetrieveLessonsRequest{
					Paging:      &cpb.Paging{Limit: 2},
					LessonTime:  lpb.LessonTime(lpb.LessonTime_value["LESSON_TIME_FUTURE"]),
					CurrentTime: timestamppb.New(now),
				}).Once().Return(&lpb.RetrieveLessonsResponse{
					Items: []*lpb.RetrieveLessonsResponse_Lesson{},
					NextPage: &cpb.Paging{
						Limit: 2,
						Offset: &cpb.Paging_OffsetString{
							OffsetString: "",
						},
					},
					PreviousPage: &cpb.Paging{
						Limit: 2,
						Offset: &cpb.Paging_OffsetString{
							OffsetString: "",
						},
					},
					TotalLesson: 0,
					TotalItems:  0,
				}, nil)
			},
		},
		{
			name: "School Admin get list lessons fail",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &bpb.RetrieveLessonsRequestV2{
				Paging:      &cpb.Paging{Limit: 2},
				LessonTime:  bpb.LessonTime(bpb.LessonTime_value["LESSON_TIME_FUTURE"]),
				CurrentTime: timestamppb.New(now),
				Keyword:     "Lesson Name",
				LocationIds: locationIds,
			},
			expectedErr:  someErrors,
			expectedResp: nil,
			setup: func(ctx context.Context) {
				mockLessonManagement.On("RetrieveLessonsV2", mock.Anything, &lpb.RetrieveLessonsRequest{
					Paging:      &cpb.Paging{Limit: 2},
					LessonTime:  lpb.LessonTime(bpb.LessonTime_value["LESSON_TIME_FUTURE"]),
					CurrentTime: timestamppb.New(now),
					Keyword:     "Lesson Name",
					LocationIds: locationIds,
				}).Once().Return(nil, someErrors)
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*bpb.RetrieveLessonsRequestV2)
			resp, err := s.RetrieveLessons(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestLessonManagementService_DeleteLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonRepo := &mock_repositories.MockLessonRepo{}
	lessonReportRepo := &mock_repositories.MockLessonReportRepo{}
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	jsm := &mock_nats.JetStreamManagement{}
	jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {}).Return("", nil)
	mockDeleteLessonManagement := &MockLessonManagement{}

	lessonID := "lesson-id-1"

	tcs := []struct {
		name     string
		req      *bpb.DeleteLessonRequest
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "delete successfully",
			req: &bpb.DeleteLessonRequest{
				LessonId:     lessonID,
				SavingOption: &bpb.DeleteLessonRequest_SavingOption{Method: bpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME},
			},
			setup: func(ctx context.Context) {
				mockDeleteLessonManagement.On("DeleteLessonV2", ctx,
					&lpb.DeleteLessonRequest{
						LessonId: lessonID,
						SavingOption: &lpb.DeleteLessonRequest_SavingOption{
							Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
						},
					}).Return(&lpb.DeleteLessonResponse{}, nil).Once()

			},
			hasError: false,
		},
		{
			name: "delete successfully without saving_option",
			req: &bpb.DeleteLessonRequest{
				LessonId: lessonID,
			},
			setup: func(ctx context.Context) {
				mockDeleteLessonManagement.On("DeleteLessonV2", ctx,
					&lpb.DeleteLessonRequest{
						LessonId: lessonID,
					}).Return(&lpb.DeleteLessonResponse{}, nil).Once()

			},
			hasError: false,
		},
		{
			name: "delete failed",
			req:  &bpb.DeleteLessonRequest{LessonId: lessonID},
			setup: func(ctx context.Context) {
				mockDeleteLessonManagement.On("DeleteLessonV2", ctx, &lpb.DeleteLessonRequest{LessonId: lessonID}).Return(nil, fmt.Errorf("some err")).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			srv := &LessonManagementService{
				DeleteLessonV2: mockDeleteLessonManagement.DeleteLessonV2,
			}
			_, err := srv.DeleteLesson(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo, lessonReportRepo)
		})
	}
}

func TestLessonManagementService_CreateLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		mockLessonManagement := &MockLessonManagement{}
		srv := &LessonManagementService{
			CreateLessonV2: mockLessonManagement.CreateLessonV2,
		}
		now := time.Now()
		req := &bpb.CreateLessonRequest{
			StartTime:      timestamppb.New(now),
			EndTime:        timestamppb.New(now),
			TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
			TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
			TeacherIds:     []string{"teacher-id-1", "teacher-id-2"},
			CenterId:       "center-id-1",
			StudentInfoList: []*bpb.CreateLessonRequest_StudentInfo{
				{
					StudentId:        "user-id-1",
					CourseId:         "course-id-1",
					AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
					AttendanceNotice: bpb.StudentAttendanceNotice_NOTICE_EMPTY,
					AttendanceReason: bpb.StudentAttendanceReason_REASON_EMPTY,
				},
				{
					StudentId:        "user-id-2",
					CourseId:         "course-id-2",
					AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_EMPTY,
					AttendanceNotice: bpb.StudentAttendanceNotice_ON_THE_DAY,
					AttendanceReason: bpb.StudentAttendanceReason_SCHOOL_EVENT,
				},
				{
					StudentId:        "user-id-3",
					CourseId:         "course-id-3",
					AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
					AttendanceNotice: bpb.StudentAttendanceNotice_IN_ADVANCE,
					AttendanceReason: bpb.StudentAttendanceReason_PHYSICAL_CONDITION,
					AttendanceNote:   "sample-attendance-note",
				},
			},
			Materials: []*bpb.Material{
				{
					Resource: &bpb.Material_MediaId{
						MediaId: "media-id-1",
					},
				},
				{
					Resource: &bpb.Material_MediaId{
						MediaId: "media-id-2",
					},
				},
				{
					Resource: &bpb.Material_BrightcoveVideo_{
						BrightcoveVideo: &bpb.Material_BrightcoveVideo{
							Name: "brightcove-video",
							Url:  "https://bri.com/video",
						},
					},
				},
			},
			SavingOption: &bpb.CreateLessonRequest_SavingOption{
				Method: bpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
			},
			ClassId:      "class-id",
			CourseId:     "course-id",
			ClassroomIds: []string{"classroom-id-1, classroom-id-2"},
		}
		lessonId := "lesson-id"
		mockLessonManagement.On("CreateLessonV2", mock.Anything, mock.MatchedBy(func(lReq *lpb.CreateLessonRequest) bool {
			assert.Equal(t, req.StartTime, lReq.StartTime)
			assert.Equal(t, req.EndTime, lReq.EndTime)
			assert.Equal(t, req.TeachingMedium, lReq.TeachingMedium)
			assert.Equal(t, req.TeachingMethod, lReq.TeachingMethod)
			assert.Equal(t, req.TeacherIds, lReq.TeacherIds)
			assert.Equal(t, req.CenterId, lReq.LocationId)
			for _, v := range req.StudentInfoList {
				assert.Equal(t, true, checkInfoList(v, lReq.StudentInfoList))
			}
			for _, v := range req.Materials {
				assert.Equal(t, true, checkMaterials(v, lReq.Materials))
			}
			assert.Equal(t, req.SavingOption.Method.String(), lReq.SavingOption.Method.Enum().String())
			assert.Equal(t, req.ClassId, lReq.ClassId)
			assert.Equal(t, req.CourseId, lReq.CourseId)
			assert.Equal(t, req.ClassroomIds, lReq.ClassroomIds)
			return true
		})).Once().Return(&lpb.CreateLessonResponse{Id: lessonId}, nil)

		lRes, err := srv.CreateLesson(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, lRes.Id, lessonId)
	})

	t.Run("fail by call new service has error", func(t *testing.T) {
		mockLessonManagement := &MockLessonManagement{}
		srv := &LessonManagementService{
			CreateLessonV2: mockLessonManagement.CreateLessonV2,
		}
		req := &bpb.CreateLessonRequest{}
		expectError := status.Error(codes.Internal, `unexpected saving option method `)

		mockLessonManagement.On("CreateLessonV2", mock.Anything, mock.Anything).Once().Return(nil, expectError)

		_, err := srv.CreateLesson(ctx, req)
		assert.Equal(t, expectError, err)
	})
}

func checkInfoList(v *bpb.CreateLessonRequest_StudentInfo, lc []*lpb.CreateLessonRequest_StudentInfo) bool {
	for _, b := range lc {
		if v.CourseId == b.CourseId && v.LocationId == b.LocationId && v.AttendanceStatus == bpb.StudentAttendStatus(b.AttendanceStatus) && v.StudentId == b.StudentId {
			return true
		}
	}
	return false
}

func checkMaterials(v *bpb.Material, lc []*lpb.Material) bool {
	for _, b := range lc {
		switch br := v.Resource.(type) {
		case *bpb.Material_BrightcoveVideo_:
			switch lr := b.Resource.(type) {
			case *lpb.Material_BrightcoveVideo_:
				if br.BrightcoveVideo.Name == lr.BrightcoveVideo.Name && br.BrightcoveVideo.Url == lr.BrightcoveVideo.Url {
					return true
				}
			}
		case *bpb.Material_MediaId:
			switch lr := b.Resource.(type) {
			case *lpb.Material_MediaId:
				if br.MediaId == lr.MediaId {
					return true
				}
			}
		}
	}
	return false
}

func TestLessonManagementService_UpdateLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		mockLessonManagement := &MockLessonManagement{}
		srv := &LessonManagementService{
			UpdateLessonV2: mockLessonManagement.UpdateLessonV2,
		}
		now := time.Now()
		req := &bpb.UpdateLessonRequest{
			StartTime:      timestamppb.New(now),
			EndTime:        timestamppb.New(now),
			TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
			TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
			TeacherIds:     []string{"teacher-id-1", "teacher-id-2"},
			CenterId:       "center-id-1",
			StudentInfoList: []*bpb.UpdateLessonRequest_StudentInfo{
				{
					StudentId:        "user-id-1",
					CourseId:         "course-id-1",
					AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
					AttendanceNotice: bpb.StudentAttendanceNotice_NOTICE_EMPTY,
					AttendanceReason: bpb.StudentAttendanceReason_REASON_EMPTY,
				},
				{
					StudentId:        "user-id-2",
					CourseId:         "course-id-2",
					AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_EMPTY,
					AttendanceNotice: bpb.StudentAttendanceNotice_ON_THE_DAY,
					AttendanceReason: bpb.StudentAttendanceReason_SCHOOL_EVENT,
				},
				{
					StudentId:        "user-id-3",
					CourseId:         "course-id-3",
					AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
					AttendanceNotice: bpb.StudentAttendanceNotice_IN_ADVANCE,
					AttendanceReason: bpb.StudentAttendanceReason_PHYSICAL_CONDITION,
					AttendanceNote:   "sample-attendance-note",
				},
			},
			Materials: []*bpb.Material{
				{
					Resource: &bpb.Material_MediaId{
						MediaId: "media-id-1",
					},
				},
				{
					Resource: &bpb.Material_MediaId{
						MediaId: "media-id-2",
					},
				},
				{
					Resource: &bpb.Material_BrightcoveVideo_{
						BrightcoveVideo: &bpb.Material_BrightcoveVideo{
							Name: "brightcove-video",
							Url:  "https://bri.com/video",
						},
					},
				},
			},
			SavingOption: &bpb.UpdateLessonRequest_SavingOption{
				Method: bpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
			},
			ClassId:      "class-id",
			CourseId:     "course-id",
			ClassroomIds: []string{"classroom-id-1, classroom-id-2"},
		}
		mockLessonManagement.On("UpdateLessonV2", mock.Anything, mock.MatchedBy(func(lReq *lpb.UpdateLessonRequest) bool {
			assert.Equal(t, req.StartTime, lReq.StartTime)
			assert.Equal(t, req.EndTime, lReq.EndTime)
			assert.Equal(t, req.TeachingMedium, lReq.TeachingMedium)
			assert.Equal(t, req.TeachingMethod, lReq.TeachingMethod)
			assert.Equal(t, req.TeacherIds, lReq.TeacherIds)
			assert.Equal(t, req.CenterId, lReq.LocationId)
			for _, v := range req.StudentInfoList {
				assert.Equal(t, true, checkInfoListOnUpdateLesson(v, lReq.StudentInfoList))
			}
			for _, v := range req.Materials {
				assert.Equal(t, true, checkMaterials(v, lReq.Materials))
			}
			assert.Equal(t, req.SavingOption.Method.String(), lReq.SavingOption.Method.Enum().String())
			assert.Equal(t, req.ClassId, lReq.ClassId)
			assert.Equal(t, req.CourseId, lReq.CourseId)
			assert.Equal(t, req.ClassroomIds, lReq.ClassroomIds)
			return true
		})).Once().Return(&lpb.UpdateLessonResponse{}, nil)

		_, err := srv.UpdateLesson(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("fail with lessonId", func(t *testing.T) {
		mockLessonManagement := &MockLessonManagement{}
		srv := &LessonManagementService{
			UpdateLessonV2: mockLessonManagement.UpdateLessonV2,
		}
		now := time.Now()
		req := &bpb.UpdateLessonRequest{
			StartTime:      timestamppb.New(now),
			EndTime:        timestamppb.New(now),
			TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
			TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
			TeacherIds:     []string{"teacher-id-1", "teacher-id-2"},
			CenterId:       "center-id-1",
			StudentInfoList: []*bpb.UpdateLessonRequest_StudentInfo{
				{
					StudentId:        "user-id-1",
					CourseId:         "course-id-1",
					AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
				},
				{
					StudentId:        "user-id-2",
					CourseId:         "course-id-2",
					AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_EMPTY,
				},
				{
					StudentId:        "user-id-3",
					CourseId:         "course-id-3",
					AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_INFORMED_ABSENT,
				},
			},
			Materials: []*bpb.Material{
				{
					Resource: &bpb.Material_MediaId{
						MediaId: "media-id-1",
					},
				},
				{
					Resource: &bpb.Material_MediaId{
						MediaId: "media-id-2",
					},
				},
			},
			SavingOption: &bpb.UpdateLessonRequest_SavingOption{
				Method: bpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
			},
			ClassId:      "class-id",
			CourseId:     "course-id",
			ClassroomIds: []string{"classroom-id-1, classroom-id-2"},
		}

		mockLessonManagement.On("UpdateLessonV2", mock.Anything, mock.Anything).Once().Return(&lpb.UpdateLessonResponse{}, fmt.Errorf("some error"))

		_, err := srv.UpdateLesson(ctx, req)
		assert.EqualError(t, err, "some error")
	})
}

func checkInfoListOnUpdateLesson(v *bpb.UpdateLessonRequest_StudentInfo, lc []*lpb.UpdateLessonRequest_StudentInfo) bool {
	for _, b := range lc {
		if v.CourseId == b.CourseId && v.LocationId == b.LocationId && v.AttendanceStatus == bpb.StudentAttendStatus(b.AttendanceStatus) && v.StudentId == b.StudentId {
			return true
		}
	}
	return false
}

type MockLessonManagement struct {
	mock.Mock
}

func (r *MockLessonManagement) CreateLessonV2(arg1 context.Context, arg2 *lpb.CreateLessonRequest) (*lpb.CreateLessonResponse, error) {
	args := r.Called(arg1, arg2)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*lpb.CreateLessonResponse), args.Error(1)
}

func (r *MockLessonManagement) DeleteLessonV2(arg1 context.Context, arg2 *lpb.DeleteLessonRequest) (*lpb.DeleteLessonResponse, error) {
	args := r.Called(arg1, arg2)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*lpb.DeleteLessonResponse), args.Error(1)
}

func (r *MockLessonManagement) UpdateLessonV2(arg1 context.Context, arg2 *lpb.UpdateLessonRequest) (*lpb.UpdateLessonResponse, error) {
	args := r.Called(arg1, arg2)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*lpb.UpdateLessonResponse), args.Error(1)
}

func (r *MockLessonManagement) RetrieveLessonsV2(arg1 context.Context, arg2 *lpb.RetrieveLessonsRequest) (*lpb.RetrieveLessonsResponse, error) {
	args := r.Called(arg1, arg2)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*lpb.RetrieveLessonsResponse), args.Error(1)
}
