package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/queries"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"

	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_lesson_es "github.com/manabie-com/backend/mock/lessonmgmt/lesson/elasticsearch"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	mock_user_repo "github.com/manabie-com/backend/mock/lessonmgmt/user/repositories"

	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const LessonTest1, UserTest1 = "lesson-1", "user-1"

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestLessonReaderService_RetrieveLessonByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := new(mock_repositories.MockLessonRepo)
	now := time.Now()
	s := &LessonReaderService{
		wrapperConnection: wrapperConnection,
		retrieveLessonCommand: application.RetrieveLessonCommand{
			WrapperConnection: wrapperConnection,
			LessonRepo:        lessonRepo,
		},
	}
	testCases := []TestCase{
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &lpb.RetrieveLessonByIDRequest{LessonId: LessonTest1},
			expectedErr: nil,
			expectedResp: &lpb.RetrieveLessonByIDResponse{
				Lesson: &lpb.Lesson{
					LessonId:         LessonTest1,
					LocationId:       "location-id-1",
					StartTime:        timestamppb.New(now),
					EndTime:          timestamppb.New(now),
					CreatedAt:        timestamppb.New(now),
					UpdatedAt:        timestamppb.New(now),
					SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT,
					TeachingMethod:   cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
					TeachingMedium:   cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE,
					LearnerMembers: []*lpb.LessonMember{
						{
							StudentId:        UserTest1,
							CourseId:         "course-id-1",
							LocationId:       "center-id-1",
							AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						},
					},
					TeacherIds: []string{"teacher-1"},
					MediaIds:   []string{"media-1"},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonByID", ctx, db, LessonTest1).
					Return(&domain.Lesson{
						LessonID:         LessonTest1,
						LocationID:       "location-id-1",
						CreatedAt:        now,
						UpdatedAt:        now,
						StartTime:        now,
						EndTime:          now,
						SchedulingStatus: domain.LessonSchedulingStatusDraft,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						Learners: domain.LessonLearners{
							{
								LearnerID:    UserTest1,
								CourseID:     "course-id-1",
								AttendStatus: domain.StudentAttendStatusAttend,
								LocationID:   "center-id-1",
							},
						},
						Teachers: domain.LessonTeachers{
							{
								TeacherID: "teacher-1",
							},
						},
						Material: &domain.LessonMaterial{MediaIDs: []string{"media-1"}},
					}, nil).Once()
			},
		},
		{
			name:        "error case",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &lpb.RetrieveLessonByIDRequest{LessonId: LessonTest1},
			expectedErr: fmt.Errorf("rpc error: code = Internal desc = cannot get lesson by id lesson-1"),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonByID", ctx, db, LessonTest1).
					Return(nil, fmt.Errorf("errSubString")).Once()
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*lpb.RetrieveLessonByIDRequest)
			resp, err := s.RetrieveLessonByID(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}
}

func TestLessonManagementService_RetrieveLessons(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	now := time.Now().UTC()
	searchRepo := &mock_lesson_es.MockSearchRepo{}

	s := &LessonReaderService{
		wrapperConnection: wrapperConnection,
		retrieveLessonCommand: application.RetrieveLessonCommand{
			WrapperConnection: wrapperConnection,
			SearchRepo:        searchRepo,
		},
	}

	courseId := "course-1"
	courses := []string{courseId}
	teachers := []string{"teacher-1"}
	students := []string{"student-1"}
	centers := []string{"center-1"}
	locations := []string{"center-1", "center-2"}
	class := "class-1"

	testCases := []TestCase{
		{
			name: "School Admin get list lessons future successfully with filter",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsRequest{
				Paging:      &cpb.Paging{Limit: 2},
				LessonTime:  lpb.LessonTime(lpb.LessonTime_value["LESSON_TIME_FUTURE"]),
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
					TimeZone:    "UTC",
					FromTime:    durationpb.New(2 * time.Hour),
					ToTime:      durationpb.New(10 * time.Hour),
					FromDate:    timestamppb.New(now.Add(-30 * time.Hour)),
					ToDate:      timestamppb.New(now.Add(30 * time.Hour)),
					TeacherIds:  []string{"teacher-1"},
					LocationIds: []string{"center-1"},
					CourseIds:   []string{"course-1"},
					StudentIds:  []string{"student-1"},
					Grades:      []int32{5, 6},
				},
				LocationIds: locations,
			},
			expectedErr: nil,
			expectedResp: &lpb.RetrieveLessonsResponse{
				Items: []*lpb.RetrieveLessonsResponse_Lesson{
					{
						Id:             "lesson-1",
						Name:           "Lesson Name",
						CenterId:       "center-1",
						StartTime:      timestamppb.New(now),
						EndTime:        timestamppb.New(now),
						TeacherIds:     []string{"teacher-1", "teacher-2"},
						TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
						TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE,
						ClassId:        class,
						CourseId:       courseId,
					},
					{
						Id:             "lesson-2",
						Name:           "Lesson Name",
						CenterId:       "center-1",
						StartTime:      timestamppb.New(now),
						EndTime:        timestamppb.New(now),
						TeacherIds:     []string{"teacher-1"},
						TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
						TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
						ClassId:        "",
						CourseId:       "",
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
				searchRepo.On("Search", mock.Anything, &domain.ListLessonArgs{
					Limit:       2,
					LessonID:    "",
					CurrentTime: now,
					Compare:     ">=",
					LessonTime:  "future",
					CourseIDs:   courses,
					TeacherIDs:  teachers,
					StudentIDs:  students,
					FromDate:    now.Add(-30 * time.Hour),
					ToDate:      now.Add(30 * time.Hour),
					FromTime:    "02:00:00",
					ToTime:      "10:00:00",
					KeyWord:     "Lesson Name",
					LocationIDs: centers,
					Dow:         []domain.DateOfWeek{0, 1, 2, 3, 4, 5, 6},
					Grades:      []int32{5, 6},
					TimeZone:    "UTC",
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:       "lesson-1",
						Name:           "Lesson Name",
						LocationID:     "center-1",
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodGroup,
						TeachingMedium: domain.LessonTeachingMediumOnline,
						Teachers: domain.LessonTeachers{
							&domain.LessonTeacher{TeacherID: "teacher-1"},
							&domain.LessonTeacher{TeacherID: "teacher-2"},
						},
						ClassID:  class,
						CourseID: courseId,
					},
					{
						LessonID:       "lesson-2",
						Name:           "Lesson Name",
						LocationID:     "center-1",
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodIndividual,
						TeachingMedium: domain.LessonTeachingMediumOffline,
						Teachers: domain.LessonTeachers{
							&domain.LessonTeacher{TeacherID: "teacher-1"},
						},
						ClassID:  "",
						CourseID: "",
					},
				}, uint32(99), "", nil)
			},
		},
		{
			name: "School Admin get list lessons future successfully without filter",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsRequest{
				Paging:      &cpb.Paging{Limit: 2},
				LessonTime:  lpb.LessonTime(lpb.LessonTime_value["LESSON_TIME_FUTURE"]),
				CurrentTime: timestamppb.New(now),
			},
			expectedErr: nil,
			expectedResp: &lpb.RetrieveLessonsResponse{
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
						ClassId:        class,
						CourseId:       courseId,
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
						ClassId:        "",
						CourseId:       "",
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
				searchRepo.On("Search", mock.Anything, &domain.ListLessonArgs{
					Limit:       2,
					CurrentTime: now,
					Compare:     ">=",
					LessonTime:  "future",
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:   "lesson-1",
						Name:       "Name",
						LocationID: "center-1",
						Teachers: domain.LessonTeachers{
							&domain.LessonTeacher{TeacherID: "teacher-1"},
						},
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodGroup,
						TeachingMedium: domain.LessonTeachingMediumOnline,
						ClassID:        class,
						CourseID:       courseId,
					},
					{
						LessonID:   "lesson-2",
						Name:       "Name",
						LocationID: "center-1",
						Teachers: domain.LessonTeachers{
							&domain.LessonTeacher{TeacherID: "teacher-1"},
						},
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodIndividual,
						TeachingMedium: domain.LessonTeachingMediumOffline,
						ClassID:        "",
						CourseID:       "",
					},
				}, uint32(99), "", nil)
			},
		},
		{
			name: "Teacher get list lessons future successfully without filter",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsRequest{
				Paging:      &cpb.Paging{Limit: 2},
				LessonTime:  lpb.LessonTime(lpb.LessonTime_value["LESSON_TIME_FUTURE"]),
				CurrentTime: timestamppb.New(now),
			},
			expectedErr: nil,
			expectedResp: &lpb.RetrieveLessonsResponse{
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
						ClassId:        class,
						CourseId:       courseId,
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
						ClassId:        "",
						CourseId:       "",
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
				searchRepo.On("Search", mock.Anything, &domain.ListLessonArgs{
					Limit:       2,
					CurrentTime: now,
					Compare:     ">=",
					LessonTime:  "future",
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:   "lesson-1",
						Name:       "Name",
						LocationID: "center-1",
						Teachers: domain.LessonTeachers{
							&domain.LessonTeacher{TeacherID: "teacher-1"},
						},
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodGroup,
						TeachingMedium: domain.LessonTeachingMediumOnline,
						ClassID:        class,
						CourseID:       courseId,
					},
					{
						LessonID:   "lesson-2",
						Name:       "Name",
						LocationID: "center-1",
						Teachers: domain.LessonTeachers{
							&domain.LessonTeacher{TeacherID: "teacher-1"},
						},
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodIndividual,
						TeachingMedium: domain.LessonTeachingMediumOffline,
						ClassID:        "",
						CourseID:       "",
					},
				}, uint32(99), "", nil)
			},
		},
		{
			name: "School Admin return list lessons past successfully without filter",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsRequest{
				Paging:      &cpb.Paging{Limit: 2},
				LessonTime:  lpb.LessonTime(lpb.LessonTime_value["LESSON_TIME_PAST"]),
				CurrentTime: timestamppb.New(now),
			},
			expectedErr: nil,
			expectedResp: &lpb.RetrieveLessonsResponse{
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
						ClassId:        class,
						CourseId:       courseId,
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
						ClassId:        "",
						CourseId:       "",
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
				searchRepo.On("Search", mock.Anything, &domain.ListLessonArgs{
					Limit:       2,
					CurrentTime: now,
					Compare:     "<",
					LessonTime:  "past",
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:   "lesson-1",
						Name:       "Name",
						LocationID: "center-1",
						Teachers: domain.LessonTeachers{
							&domain.LessonTeacher{TeacherID: "teacher-1"},
						},
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodGroup,
						TeachingMedium: domain.LessonTeachingMediumOnline,
						ClassID:        class,
						CourseID:       courseId,
					},
					{
						LessonID:   "lesson-2",
						Name:       "Name",
						LocationID: "center-1",
						Teachers: domain.LessonTeachers{
							&domain.LessonTeacher{TeacherID: "teacher-1"},
						},
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodIndividual,
						TeachingMedium: domain.LessonTeachingMediumOffline,
						ClassID:        "",
						CourseID:       "",
					},
				}, uint32(99), "", nil)
			},
		},
		{
			name: "Return list empty successfully",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsRequest{
				Paging:      &cpb.Paging{Limit: 2},
				LessonTime:  lpb.LessonTime(lpb.LessonTime_value["LESSON_TIME_FUTURE"]),
				CurrentTime: timestamppb.New(now),
			},
			expectedErr: nil,
			expectedResp: &lpb.RetrieveLessonsResponse{
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
			},
			setup: func(ctx context.Context) {

				searchRepo.On("Search", mock.Anything, &domain.ListLessonArgs{
					Limit:       2,
					CurrentTime: now,
					Compare:     ">=",
					LessonTime:  "future",
				}).Once().Return([]*domain.Lesson{}, uint32(0), "", nil)
			},
		},
		{
			name: "Return fail missing page",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsRequest{
				LessonTime:  lpb.LessonTime(lpb.LessonTime_value["LESSON_TIME_FUTURE"]),
				CurrentTime: timestamppb.New(now),
			},
			expectedErr:  status.Error(codes.Internal, "missing paging info"),
			expectedResp: &lpb.RetrieveLessonsResponse{},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "Return fail missing current time",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsRequest{
				Paging:     &cpb.Paging{Limit: 2},
				LessonTime: lpb.LessonTime(lpb.LessonTime_value["LESSON_TIME_FUTURE"]),
			},
			expectedErr:  status.Error(codes.Internal, "missing current time"),
			expectedResp: &lpb.RetrieveLessonsResponse{},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "Return fail missing timezone",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsRequest{
				Paging:      &cpb.Paging{Limit: 2},
				LessonTime:  lpb.LessonTime(lpb.LessonTime_value["LESSON_TIME_FUTURE"]),
				CurrentTime: timestamppb.New(now),
				Filter: &lpb.RetrieveLessonsFilter{
					DateOfWeeks: []cpb.DateOfWeek{
						cpb.DateOfWeek_DATE_OF_WEEK_MONDAY,
					},
				},
			},
			expectedErr:  status.Error(codes.Internal, "missing timezone"),
			expectedResp: &lpb.RetrieveLessonsResponse{},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "Return fail missing timezone filter FromTime",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsRequest{
				Paging:      &cpb.Paging{Limit: 2},
				LessonTime:  lpb.LessonTime(lpb.LessonTime_value["LESSON_TIME_FUTURE"]),
				CurrentTime: timestamppb.New(now),
				Filter: &lpb.RetrieveLessonsFilter{
					FromTime: durationpb.New(2 * time.Hour),
				},
			},
			expectedErr:  status.Error(codes.Internal, "missing timezone"),
			expectedResp: &lpb.RetrieveLessonsResponse{},
			setup: func(ctx context.Context) {
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*lpb.RetrieveLessonsRequest)
			resp, err := s.RetrieveLessons(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}
}

func TestLessonManagementService_RetrieveLessonsV2(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	now := time.Now().UTC()
	lessonRepo := new(mock_repositories.MockLessonRepo)
	lessonTeacherRepo := new(mock_repositories.MockLessonTeacherRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	s := &LessonReaderService{
		wrapperConnection: wrapperConnection,
		lessonQueryHandler: queries.LessonQueryHandler{
			WrapperConnection: wrapperConnection,
			LessonRepo:        lessonRepo,
			LessonTeacherRepo: lessonTeacherRepo,
		},
		env:              "local",
		unleashClientIns: mockUnleashClient,
	}

	courseId := "course-1"
	courses := []string{courseId}
	teachers := []string{"teacher-1"}
	students := []string{"student-1"}
	centers := []string{"center-1"}
	locations := []string{"center-1", "center-2"}
	classes := []string{"class-1", "class-2"}
	zoomLink := "wzoomcom-link"

	testCases := []TestCase{
		{
			name: "School Admin get list lessons future successfully with filter",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsRequest{
				Paging:      &cpb.Paging{Limit: 2},
				LessonTime:  lpb.LessonTime(lpb.LessonTime_value["LESSON_TIME_FUTURE"]),
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
					ClassIds:         classes,
				},
				LocationIds: locations,
			},
			expectedErr: nil,
			expectedResp: &lpb.RetrieveLessonsResponse{
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
						ClassId:          classes[0],
						CourseId:         courseId,
						SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED,
						EndAt:            nil,
					},
					{
						Id:               "lesson-2",
						Name:             "Lesson Name",
						CenterId:         "center-1",
						StartTime:        timestamppb.New(now),
						EndTime:          timestamppb.New(now),
						TeacherIds:       []string{"teacher-1"},
						TeachingMethod:   cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
						TeachingMedium:   cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
						ClassId:          "",
						CourseId:         "",
						SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED,
						EndAt:            timestamppb.New(now),
						ZoomLink:         zoomLink,
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
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				lessonRepo.On("Retrieve", mock.Anything, db, &payloads.GetLessonListArg{
					Limit:                    2,
					LessonID:                 "",
					CurrentTime:              now,
					Compare:                  ">=",
					LessonTime:               "future",
					CourseIDs:                courses,
					TeacherIDs:               teachers,
					StudentIDs:               students,
					FromDate:                 now.Add(-30 * time.Hour),
					ToDate:                   now.Add(30 * time.Hour),
					FromTime:                 "02:00:00",
					ToTime:                   "10:00:00",
					KeyWord:                  "Lesson Name",
					LocationIDs:              centers,
					Dow:                      []domain.DateOfWeek{0, 1, 2, 3, 4, 5, 6},
					Grades:                   []int32{5, 6},
					TimeZone:                 "UTC",
					LessonSchedulingStatuses: []domain.LessonSchedulingStatus{domain.LessonSchedulingStatusCanceled, domain.LessonSchedulingStatusCompleted},
					ClassIDs:                 classes,
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:         "lesson-1",
						Name:             "Lesson Name",
						LocationID:       "center-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          classes[0],
						CourseID:         courseId,
						SchedulingStatus: domain.LessonSchedulingStatusCanceled,
						EndAt:            nil,
					},
					{
						LessonID:         "lesson-2",
						Name:             "Lesson Name",
						LocationID:       "center-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOffline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						EndAt:            &now,
						ZoomLink:         zoomLink,
					},
				}, uint32(99), "pre_id", uint32(2), nil)

				lessonTeacherRepo.On("GetTeachersByLessonIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(map[string]domain.LessonTeachers{
					"lesson-1": {
						&domain.LessonTeacher{TeacherID: "teacher-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2"},
					},
					"lesson-2": {
						&domain.LessonTeacher{TeacherID: "teacher-1"},
					},
				}, nil)
			},
		},
		{
			name: "School Admin get list lessons future successfully without filter",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsRequest{
				Paging:      &cpb.Paging{Limit: 2},
				LessonTime:  lpb.LessonTime(lpb.LessonTime_value["LESSON_TIME_FUTURE"]),
				CurrentTime: timestamppb.New(now),
			},
			expectedErr: nil,
			expectedResp: &lpb.RetrieveLessonsResponse{
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
						ClassId:        classes[0],
						CourseId:       courseId,
						EndAt:          nil,
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
						ClassId:        "",
						CourseId:       "",
						EndAt:          timestamppb.New(now),
						ZoomLink:       zoomLink,
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
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				lessonRepo.On("Retrieve", mock.Anything, db, &payloads.GetLessonListArg{
					Limit:       2,
					CurrentTime: now,
					Compare:     ">=",
					LessonTime:  "future",
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:       "lesson-1",
						Name:           "Name",
						LocationID:     "center-1",
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodGroup,
						TeachingMedium: domain.LessonTeachingMediumOnline,
						ClassID:        classes[0],
						CourseID:       courseId,
					},
					{
						LessonID:       "lesson-2",
						Name:           "Name",
						LocationID:     "center-1",
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodIndividual,
						TeachingMedium: domain.LessonTeachingMediumOffline,
						ClassID:        "",
						CourseID:       "",
						EndAt:          &now,
						ZoomLink:       zoomLink,
					},
				}, uint32(99), "pre_id", uint32(2), nil)

				lessonTeacherRepo.On("GetTeachersByLessonIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(map[string]domain.LessonTeachers{
					"lesson-1": {
						&domain.LessonTeacher{TeacherID: "teacher-1"},
					},
					"lesson-2": {
						&domain.LessonTeacher{TeacherID: "teacher-1"},
					},
				}, nil)
			},
		},
		{
			name: "Teacher get list lessons future successfully without filter",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsRequest{
				Paging:      &cpb.Paging{Limit: 2},
				LessonTime:  lpb.LessonTime(lpb.LessonTime_value["LESSON_TIME_FUTURE"]),
				CurrentTime: timestamppb.New(now),
			},
			expectedErr: nil,
			expectedResp: &lpb.RetrieveLessonsResponse{
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
						ClassId:        classes[0],
						CourseId:       courseId,
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
						ClassId:        "",
						CourseId:       "",
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
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				lessonRepo.On("Retrieve", mock.Anything, db, &payloads.GetLessonListArg{
					Limit:       2,
					CurrentTime: now,
					Compare:     ">=",
					LessonTime:  "future",
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:       "lesson-1",
						Name:           "Name",
						LocationID:     "center-1",
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodGroup,
						TeachingMedium: domain.LessonTeachingMediumOnline,
						ClassID:        classes[0],
						CourseID:       courseId,
					},
					{
						LessonID:       "lesson-2",
						Name:           "Name",
						LocationID:     "center-1",
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodIndividual,
						TeachingMedium: domain.LessonTeachingMediumOffline,
						ClassID:        "",
						CourseID:       "",
					},
				}, uint32(99), "pre_id", uint32(2), nil)

				lessonTeacherRepo.On("GetTeachersByLessonIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(map[string]domain.LessonTeachers{
					"lesson-1": {
						&domain.LessonTeacher{TeacherID: "teacher-1"},
					},
					"lesson-2": {
						&domain.LessonTeacher{TeacherID: "teacher-1"},
					},
				}, nil)
			},
		},
		{
			name: "School Admin return list lessons past successfully without filter",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsRequest{
				Paging:      &cpb.Paging{Limit: 2},
				LessonTime:  lpb.LessonTime(lpb.LessonTime_value["LESSON_TIME_PAST"]),
				CurrentTime: timestamppb.New(now),
			},
			expectedErr: nil,
			expectedResp: &lpb.RetrieveLessonsResponse{
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
						ClassId:        classes[0],
						CourseId:       courseId,
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
						ClassId:        "",
						CourseId:       "",
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
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				lessonRepo.On("Retrieve", mock.Anything, db, &payloads.GetLessonListArg{
					Limit:       2,
					CurrentTime: now,
					Compare:     "<",
					LessonTime:  "past",
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:       "lesson-1",
						Name:           "Name",
						LocationID:     "center-1",
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodGroup,
						TeachingMedium: domain.LessonTeachingMediumOnline,
						ClassID:        classes[0],
						CourseID:       courseId,
					},
					{
						LessonID:       "lesson-2",
						Name:           "Name",
						LocationID:     "center-1",
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodIndividual,
						TeachingMedium: domain.LessonTeachingMediumOffline,
						ClassID:        "",
						CourseID:       "",
					},
				}, uint32(99), "pre_id", uint32(2), nil)

				lessonTeacherRepo.On("GetTeachersByLessonIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(map[string]domain.LessonTeachers{
					"lesson-1": {
						&domain.LessonTeacher{TeacherID: "teacher-1"},
					},
					"lesson-2": {
						&domain.LessonTeacher{TeacherID: "teacher-1"},
					},
				}, nil)
			},
		},
		{
			name: "Return list empty successfully",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsRequest{
				Paging:      &cpb.Paging{Limit: 2},
				LessonTime:  lpb.LessonTime(lpb.LessonTime_value["LESSON_TIME_FUTURE"]),
				CurrentTime: timestamppb.New(now),
			},
			expectedErr: nil,
			expectedResp: &lpb.RetrieveLessonsResponse{
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
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				lessonRepo.On("Retrieve", mock.Anything, db, &payloads.GetLessonListArg{
					Limit:       2,
					CurrentTime: now,
					Compare:     ">=",
					LessonTime:  "future",
				}).Once().Return([]*domain.Lesson{}, uint32(0), "", uint32(0), nil)

				lessonTeacherRepo.On("GetTeachersByLessonIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(map[string]domain.LessonTeachers{}, nil)
			},
		},
		{
			name: "Return fail missing page",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsRequest{
				LessonTime:  lpb.LessonTime(lpb.LessonTime_value["LESSON_TIME_FUTURE"]),
				CurrentTime: timestamppb.New(now),
			},
			expectedErr:  status.Error(codes.Internal, "missing paging info"),
			expectedResp: &lpb.RetrieveLessonsResponse{},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "Return fail missing current time",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsRequest{
				Paging:     &cpb.Paging{Limit: 2},
				LessonTime: lpb.LessonTime(lpb.LessonTime_value["LESSON_TIME_FUTURE"]),
			},
			expectedErr:  status.Error(codes.Internal, "missing current time"),
			expectedResp: &lpb.RetrieveLessonsResponse{},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "Return fail missing timezone",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsRequest{
				Paging:      &cpb.Paging{Limit: 2},
				LessonTime:  lpb.LessonTime(lpb.LessonTime_value["LESSON_TIME_FUTURE"]),
				CurrentTime: timestamppb.New(now),
				Filter: &lpb.RetrieveLessonsFilter{
					DateOfWeeks: []cpb.DateOfWeek{
						cpb.DateOfWeek_DATE_OF_WEEK_MONDAY,
					},
				},
			},
			expectedErr:  status.Error(codes.Internal, "missing timezone"),
			expectedResp: &lpb.RetrieveLessonsResponse{},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "Return fail missing timezone filter FromTime",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsRequest{
				Paging:      &cpb.Paging{Limit: 2},
				LessonTime:  lpb.LessonTime(lpb.LessonTime_value["LESSON_TIME_FUTURE"]),
				CurrentTime: timestamppb.New(now),
				Filter: &lpb.RetrieveLessonsFilter{
					FromTime: durationpb.New(2 * time.Hour),
				},
			},
			expectedErr:  status.Error(codes.Internal, "missing timezone"),
			expectedResp: &lpb.RetrieveLessonsResponse{},
			setup: func(ctx context.Context) {
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*lpb.RetrieveLessonsRequest)
			resp, err := s.RetrieveLessonsV2(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}

			mock.AssertExpectationsForObjects(t, lessonRepo, lessonTeacherRepo, mockUnleashClient)
		})
	}
}

func TestLessonReaderService_ListStudentsByLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonMemberRepo := new(mock_repositories.MockLessonMemberRepo)
	s := &LessonReaderService{
		wrapperConnection: wrapperConnection,
		retrieveLessonCommand: application.RetrieveLessonCommand{
			WrapperConnection: wrapperConnection,
			LessonMemberRepo:  lessonMemberRepo,
		},
	}
	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.ListStudentsByLessonRequest{LessonId: LessonTest1, Paging: &cpb.Paging{
				Limit: 10,
			}},
			expectedErr: nil,
			expectedResp: &lpb.ListStudentsByLessonResponse{
				Students: []*cpb.BasicProfile{
					{
						UserId:     UserTest1,
						Name:       "given last",
						Avatar:     "avatar",
						Group:      cpb.UserGroup(cpb.UserGroup_value["group"]),
						FacebookId: "facebook-id",
					},
				},
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetMultipleCombined{
						OffsetMultipleCombined: &cpb.Paging_MultipleCombined{
							Combined: []*cpb.Paging_Combined{
								{
									OffsetString: "given last",
								},
								{
									OffsetString: UserTest1,
								},
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonMemberRepo.On("ListStudentsByLessonArgs", ctx, db, &domain.ListStudentsByLessonArgs{
					LessonID: LessonTest1,
					Limit:    10,
					UserName: "",
					UserID:   "",
				}).
					Return([]*domain.User{{
						ID:         UserTest1,
						GivenName:  "given",
						LastName:   "last",
						Avatar:     "avatar",
						Group:      "group",
						FacebookID: "facebook-id",
					}}, nil).Once()
			},
		},
		{
			name:        "error case",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &lpb.ListStudentsByLessonRequest{LessonId: LessonTest1},
			expectedErr: fmt.Errorf("LessonMemberRepo.ListStudentsByLessonArgs err: errSubString"),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonMemberRepo.On("ListStudentsByLessonArgs", ctx, db, &domain.ListStudentsByLessonArgs{
					LessonID: LessonTest1,
					Limit:    10,
					UserName: "",
					UserID:   "",
				}).
					Return(nil, fmt.Errorf("errSubString")).Once()
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*lpb.ListStudentsByLessonRequest)
			resp, err := s.RetrieveStudentsByLesson(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}
}

func TestLessonReaderService_RetrieveLessonMedias(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	now := time.Now()
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := new(mock_repositories.MockLessonRepo)
	lessonGroupRepo := new(mock_repositories.MockLessonGroupRepo)
	s := &LessonReaderService{
		wrapperConnection: wrapperConnection,
		retrieveLessonCommand: application.RetrieveLessonCommand{
			WrapperConnection: wrapperConnection,
			LessonRepo:        lessonRepo,
			LessonGroupRepo:   lessonGroupRepo,
		},
	}
	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.ListLessonMediasRequest{LessonId: LessonTest1, Paging: &cpb.Paging{
				Limit: 10,
			}},
			expectedErr: nil,
			expectedResp: &lpb.ListLessonMediasResponse{
				Items: []*lpb.Media{
					{
						MediaId:   "id-1",
						Name:      "name",
						Resource:  "resource",
						CreatedAt: timestamppb.New(now),
						UpdatedAt: timestamppb.New(now),
						Comments:  []*lpb.Comment{},
						Images:    []*lpb.ConvertedImage{},
						Type:      lpb.MediaType(lpb.MediaType_value[string(media_domain.MediaTypeImage)]),
					},
				},
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "id-1",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonGroupRepo.On("ListMediaByLessonArgs", ctx, db, &domain.ListMediaByLessonArgs{
					LessonID: LessonTest1,
					Limit:    10,
					Offset:   "",
				}).
					Return(media_domain.Medias{{
						ID:        "id-1",
						Name:      "name",
						Resource:  "resource",
						Type:      media_domain.MediaTypeImage,
						CreatedAt: now,
						UpdatedAt: now,
					}}, nil).Once()
			},
		},
		{
			name: "happy case with media type audio",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.ListLessonMediasRequest{LessonId: LessonTest1, Paging: &cpb.Paging{
				Limit: 10,
			}},
			expectedErr: nil,
			expectedResp: &lpb.ListLessonMediasResponse{
				Items: []*lpb.Media{
					{
						MediaId:   "id-2",
						Name:      "name",
						Resource:  "resource",
						CreatedAt: timestamppb.New(now),
						UpdatedAt: timestamppb.New(now),
						Comments:  []*lpb.Comment{},
						Images:    []*lpb.ConvertedImage{},
						Type:      lpb.MediaType(lpb.MediaType_value[string(media_domain.MediaTypeAudio)]),
					},
				},
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "id-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonGroupRepo.On("ListMediaByLessonArgs", ctx, db, &domain.ListMediaByLessonArgs{
					LessonID: LessonTest1,
					Limit:    10,
					Offset:   "",
				}).
					Return(media_domain.Medias{{
						ID:        "id-2",
						Name:      "name",
						Resource:  "resource",
						Type:      media_domain.MediaTypeAudio,
						CreatedAt: now,
						UpdatedAt: now,
					}}, nil).Once()
			},
		},
		{
			name: "error case",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.ListLessonMediasRequest{LessonId: LessonTest1, Paging: &cpb.Paging{
				Limit: 10,
			}},
			expectedErr: fmt.Errorf("LessonGroupRepo.ListMediaByLessonArgs err: errSubString"),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonGroupRepo.On("ListMediaByLessonArgs", ctx, db, &domain.ListMediaByLessonArgs{
					LessonID: LessonTest1,
					Limit:    10,
					Offset:   "",
				}).Return(media_domain.Medias{}, fmt.Errorf("errSubString")).Once()
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*lpb.ListLessonMediasRequest)
			resp, err := s.RetrieveLessonMedias(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}
}

func TestLessonReaderService_ToMediaPbV1(t *testing.T) {
	t.Parallel()
	media := generateMediaEnt()

	mediaPb := toMediaLpb(&media)

	assert.Equal(t, media.ID, mediaPb.MediaId)
	assert.Equal(t, media.Name, mediaPb.Name)
	assert.Equal(t, media.Resource, mediaPb.Resource)
	for i := range mediaPb.Comments {
		assert.Equal(t, media.Comments[i].Comment, mediaPb.Comments[i].Comment)
	}
	for i := range mediaPb.Images {
		assert.Equal(t, media.ConvertedImages[i].Width, mediaPb.Images[i].Width)
		assert.Equal(t, media.ConvertedImages[i].Height, mediaPb.Images[i].Height)
		assert.Equal(t, media.ConvertedImages[i].ImageURL, mediaPb.Images[i].ImageUrl)
	}

	assert.True(t, media.CreatedAt.Equal(mediaPb.CreatedAt.AsTime()))
	assert.True(t, media.UpdatedAt.Equal(mediaPb.UpdatedAt.AsTime()))
}

func generateMediaEnt() media_domain.Media {
	return media_domain.Media{
		ID:        idutil.ULIDNow(),
		Name:      "media-name",
		Resource:  "media Resource",
		Type:      media_domain.MediaType(media_domain.MediaTypeImage),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Comments: []media_domain.Comment{
			{
				Comment:  "comment-1",
				Duration: 1,
			},
		},
		ConvertedImages: []media_domain.ConvertedImage{
			{
				Width:    1,
				Height:   1,
				ImageURL: "url",
			},
		},
	}
}

func TestLessonReaderService_RetrieveLessonsOnCalendar(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	now := time.Now().UTC()
	lessonRepo := new(mock_repositories.MockLessonRepo)
	lessonMemberRepo := new(mock_repositories.MockLessonMemberRepo)
	lessonTeacherRepo := new(mock_repositories.MockLessonTeacherRepo)
	lessonClassroomRepo := new(mock_repositories.MockLessonClassroomRepo)
	userRepo := new(mock_user_repo.MockUserRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonIDs := []string{"lesson-1", "lesson-2", "lesson-3"}

	lessonReaderService := &LessonReaderService{
		wrapperConnection: wrapperConnection,
		lessonQueryHandler: queries.LessonQueryHandler{
			WrapperConnection:   wrapperConnection,
			LessonRepo:          lessonRepo,
			LessonMemberRepo:    lessonMemberRepo,
			LessonTeacherRepo:   lessonTeacherRepo,
			LessonClassroomRepo: lessonClassroomRepo,
			UserRepo:            userRepo,
		},
		env:              "local",
		unleashClientIns: mockUnleashClient,
	}

	testCases := []TestCase{
		{
			name: "user successfully retrieves lessons on calendar",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsOnCalendarRequest{
				CalendarView: lpb.CalendarView_WEEKLY,
				LocationId:   "location-id-1",
				FromDate:     timestamppb.New(now),
				ToDate:       timestamppb.New(now.Add(7 * 24 * time.Hour)),
				Timezone:     "sample-timezone",
				LocationIds:  []string{"location-id-1", "location-id-2"},
			},
			expectedErr: nil,
			expectedResp: &lpb.RetrieveLessonsOnCalendarResponse{
				Items: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson{
					{
						TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
						LessonId:       "lesson-1",
						LessonName:     "Lesson Name 1",
						StartTime:      timestamppb.New(now),
						EndTime:        timestamppb.New(now),
						LessonTeachers: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonTeacher{
							{
								TeacherId:   "teacher-1",
								TeacherName: "teacher-name-1",
							},
							{
								TeacherId:   "teacher-2",
								TeacherName: "teacher-name-2",
							},
						},
						LessonMembers: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonMember{
							{
								StudentId:   "student-id-1",
								CourseId:    "course-1",
								Grade:       "Grade 5",
								StudentName: "student-name 1",
								CourseName:  "course-name-1",
							},
							{
								StudentId:   "student-id-2",
								CourseId:    "course-2",
								Grade:       "Grade 5",
								StudentName: "student-name 2",
								CourseName:  "course-name-2",
							},
							{
								StudentId:   "student-id-3",
								CourseId:    "course-1",
								Grade:       "Grade 6",
								StudentName: "student-name 3",
								CourseName:  "course-name-1",
							},
						},
						SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED,
						LessonClassrooms: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonClassroom{
							{
								ClassroomId:   "classroom-id-1",
								ClassroomName: "classroom-name-1",
							},
							{
								ClassroomId:   "classroom-id-2",
								ClassroomName: "classroom-name-2",
							},
						},
						SchedulerId: "scheduler-id-1",
					},
					{
						TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
						LessonId:       "lesson-2",
						LessonName:     "Lesson Name 2",
						StartTime:      timestamppb.New(now),
						EndTime:        timestamppb.New(now),
						LessonTeachers: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonTeacher{
							{
								TeacherId:   "teacher-1",
								TeacherName: "teacher-name-1",
							},
						},
						LessonMembers: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonMember{
							{
								StudentId:   "student-id-4",
								CourseId:    "course-2",
								Grade:       "Grade 4",
								StudentName: "student-name 4",
								CourseName:  "course-name-2",
							},
							{
								StudentId:   "student-id-5",
								CourseId:    "course-2",
								Grade:       "Grade 5",
								StudentName: "student-name 5",
								CourseName:  "course-name-2",
							},
						},
						SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED,
						LessonClassrooms: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonClassroom{
							{
								ClassroomId:   "classroom-id-1",
								ClassroomName: "classroom-name-1",
							},
							{
								ClassroomId:   "classroom-id-2",
								ClassroomName: "classroom-name-2",
							},
						},
						SchedulerId: "scheduler-id-2",
					},
					{
						TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
						LessonId:       "lesson-3",
						LessonName:     "Lesson Name 3",
						StartTime:      timestamppb.New(now),
						EndTime:        timestamppb.New(now),
						LessonTeachers: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonTeacher{
							{
								TeacherId:   "teacher-1",
								TeacherName: "teacher-name-1",
							},
							{
								TeacherId:   "teacher-2",
								TeacherName: "teacher-name-2",
							},
						},
						ClassId:          "class-1",
						ClassName:        "class-name-1",
						CourseId:         "course-2",
						CourseName:       "course-name-2",
						LessonMembers:    []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonMember{},
						SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED,
						LessonClassrooms: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonClassroom{
							{
								ClassroomId:   "classroom-id-1",
								ClassroomName: "classroom-name-1",
							},
							{
								ClassroomId:   "classroom-id-2",
								ClassroomName: "classroom-name-2",
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonsOnCalendar", mock.Anything, db, &payloads.GetLessonListOnCalendarArgs{
					View:       payloads.Weekly,
					FromDate:   now,
					ToDate:     now.Add(7 * 24 * time.Hour),
					LocationID: "location-id-1",
					Timezone:   "sample-timezone",
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:         "lesson-1",
						Name:             "Lesson Name 1",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						CourseName:       "",
						ClassName:        "",
						SchedulerID:      "scheduler-id-1",
					},
					{
						LessonID:         "lesson-2",
						Name:             "Lesson Name 2",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCanceled,
						CourseName:       "",
						ClassName:        "",
						SchedulerID:      "scheduler-id-2",
					},
					{
						LessonID:         "lesson-3",
						Name:             "Lesson Name 3",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "class-1",
						CourseID:         "course-2",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						CourseName:       "course-name-2",
						ClassName:        "class-name-1",
					},
				}, nil)

				lessonTeacherRepo.On("GetTeachersWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonTeachers{
					"lesson-1": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
					"lesson-2": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
					},
					"lesson-3": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
				}, nil)

				lessonMemberRepo.On("GetLessonLearnersWithCourseAndNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonLearners{
					"lesson-1": {
						&domain.LessonLearner{
							LearnerID:   "student-id-1",
							CourseID:    "course-1",
							LearnerName: "student-name 1",
							CourseName:  "course-name-1",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-2",
							CourseID:    "course-2",
							LearnerName: "student-name 2",
							CourseName:  "course-name-2",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-3",
							CourseID:    "course-1",
							LearnerName: "student-name 3",
							CourseName:  "course-name-1",
						},
					},
					"lesson-2": {
						&domain.LessonLearner{
							LearnerID:   "student-id-4",
							CourseID:    "course-2",
							LearnerName: "student-name 4",
							CourseName:  "course-name-2",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-5",
							CourseID:    "course-2",
							LearnerName: "student-name 5",
							CourseName:  "course-name-2",
						},
					},
				}, nil)

				lessonClassroomRepo.On("GetLessonClassroomsWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs).Once().Return(map[string]domain.LessonClassrooms{
					"lesson-1": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
					"lesson-2": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
					"lesson-3": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
				}, nil)

				userRepo.On("GetStudentCurrentGradeByUserIDs", mock.Anything, mock.Anything, []string{"student-id-1", "student-id-2", "student-id-3"}).Once().Return(map[string]string{
					"student-id-1": "Grade 5",
					"student-id-2": "Grade 5",
					"student-id-3": "Grade 6",
				}, nil)

				userRepo.On("GetStudentCurrentGradeByUserIDs", mock.Anything, mock.Anything, []string{"student-id-4", "student-id-5"}).Once().Return(map[string]string{
					"student-id-4": "Grade 4",
					"student-id-5": "Grade 5",
				}, nil)
			},
		},
		{
			name: "user successfully retrieves lessons on calendar with filter",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsOnCalendarRequest{
				CalendarView: lpb.CalendarView_WEEKLY,
				LocationId:   "location-id-1",
				FromDate:     timestamppb.New(now),
				ToDate:       timestamppb.New(now.Add(7 * 24 * time.Hour)),
				Timezone:     "sample-timezone",
				LocationIds:  []string{"location-id-1", "location-id-2"},
				Filter: &lpb.RetrieveLessonsOnCalendarRequest_Filter{
					StudentIds: []string{"student-id-1", "student-id-2", "student-id-3", "student-id-4", "student-id-5"},
					CourseIds:  []string{"course-1", "course-2"},
					TeacherIds: []string{"teacher-1", "teacher-2"},
					ClassIds:   []string{"class-1"},
				},
			},
			expectedErr: nil,
			expectedResp: &lpb.RetrieveLessonsOnCalendarResponse{
				Items: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson{
					{
						TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
						LessonId:       "lesson-1",
						LessonName:     "Lesson Name 1",
						StartTime:      timestamppb.New(now),
						EndTime:        timestamppb.New(now),
						LessonTeachers: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonTeacher{
							{
								TeacherId:   "teacher-1",
								TeacherName: "teacher-name-1",
							},
							{
								TeacherId:   "teacher-2",
								TeacherName: "teacher-name-2",
							},
						},
						LessonMembers: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonMember{
							{
								StudentId:   "student-id-1",
								CourseId:    "course-1",
								Grade:       "Grade 5",
								StudentName: "student-name 1",
								CourseName:  "course-name-1",
							},
							{
								StudentId:   "student-id-2",
								CourseId:    "course-2",
								Grade:       "Grade 5",
								StudentName: "student-name 2",
								CourseName:  "course-name-2",
							},
							{
								StudentId:   "student-id-3",
								CourseId:    "course-1",
								Grade:       "Grade 6",
								StudentName: "student-name 3",
								CourseName:  "course-name-1",
							},
						},
						SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED,
						LessonClassrooms: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonClassroom{
							{
								ClassroomId:   "classroom-id-1",
								ClassroomName: "classroom-name-1",
							},
							{
								ClassroomId:   "classroom-id-2",
								ClassroomName: "classroom-name-2",
							},
						},
						SchedulerId: "scheduler-id-1",
					},
					{
						TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
						LessonId:       "lesson-2",
						LessonName:     "Lesson Name 2",
						StartTime:      timestamppb.New(now),
						EndTime:        timestamppb.New(now),
						LessonTeachers: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonTeacher{
							{
								TeacherId:   "teacher-1",
								TeacherName: "teacher-name-1",
							},
						},
						LessonMembers: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonMember{
							{
								StudentId:   "student-id-4",
								CourseId:    "course-2",
								Grade:       "Grade 4",
								StudentName: "student-name 4",
								CourseName:  "course-name-2",
							},
							{
								StudentId:   "student-id-5",
								CourseId:    "course-2",
								Grade:       "Grade 5",
								StudentName: "student-name 5",
								CourseName:  "course-name-2",
							},
						},
						SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED,
						LessonClassrooms: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonClassroom{
							{
								ClassroomId:   "classroom-id-1",
								ClassroomName: "classroom-name-1",
							},
							{
								ClassroomId:   "classroom-id-2",
								ClassroomName: "classroom-name-2",
							},
						},
						SchedulerId: "scheduler-id-2",
					},
					{
						TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
						LessonId:       "lesson-3",
						LessonName:     "Lesson Name 3",
						StartTime:      timestamppb.New(now),
						EndTime:        timestamppb.New(now),
						LessonTeachers: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonTeacher{
							{
								TeacherId:   "teacher-1",
								TeacherName: "teacher-name-1",
							},
							{
								TeacherId:   "teacher-2",
								TeacherName: "teacher-name-2",
							},
						},
						ClassId:          "class-1",
						ClassName:        "class-name-1",
						CourseId:         "course-2",
						CourseName:       "course-name-2",
						LessonMembers:    []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonMember{},
						SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED,
						LessonClassrooms: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonClassroom{
							{
								ClassroomId:   "classroom-id-1",
								ClassroomName: "classroom-name-1",
							},
							{
								ClassroomId:   "classroom-id-2",
								ClassroomName: "classroom-name-2",
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonsOnCalendar", mock.Anything, db, &payloads.GetLessonListOnCalendarArgs{
					View:       payloads.Weekly,
					FromDate:   now,
					ToDate:     now.Add(7 * 24 * time.Hour),
					LocationID: "location-id-1",
					Timezone:   "sample-timezone",
					StudentIDs: []string{"student-id-1", "student-id-2", "student-id-3", "student-id-4", "student-id-5"},
					CourseIDs:  []string{"course-1", "course-2"},
					TeacherIDs: []string{"teacher-1", "teacher-2"},
					ClassIDs:   []string{"class-1"},
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:         "lesson-1",
						Name:             "Lesson Name 1",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-1",
					},
					{
						LessonID:         "lesson-2",
						Name:             "Lesson Name 2",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCanceled,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-2",
					},
					{
						LessonID:         "lesson-3",
						Name:             "Lesson Name 3",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "class-1",
						CourseID:         "course-2",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						CourseName:       "course-name-2",
						ClassName:        "class-name-1",
					},
				}, nil)

				lessonTeacherRepo.On("GetTeachersWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonTeachers{
					"lesson-1": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
					"lesson-2": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
					},
					"lesson-3": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
				}, nil)

				lessonMemberRepo.On("GetLessonLearnersWithCourseAndNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonLearners{
					"lesson-1": {
						&domain.LessonLearner{
							LearnerID:   "student-id-1",
							CourseID:    "course-1",
							LearnerName: "student-name 1",
							CourseName:  "course-name-1",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-2",
							CourseID:    "course-2",
							LearnerName: "student-name 2",
							CourseName:  "course-name-2",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-3",
							CourseID:    "course-1",
							LearnerName: "student-name 3",
							CourseName:  "course-name-1",
						},
					},
					"lesson-2": {
						&domain.LessonLearner{
							LearnerID:   "student-id-4",
							CourseID:    "course-2",
							LearnerName: "student-name 4",
							CourseName:  "course-name-2",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-5",
							CourseID:    "course-2",
							LearnerName: "student-name 5",
							CourseName:  "course-name-2",
						},
					},
				}, nil)

				lessonClassroomRepo.On("GetLessonClassroomsWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs).Once().Return(map[string]domain.LessonClassrooms{
					"lesson-1": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
					"lesson-2": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
					"lesson-3": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
				}, nil)

				userRepo.On("GetStudentCurrentGradeByUserIDs", mock.Anything, mock.Anything, []string{"student-id-1", "student-id-2", "student-id-3"}).Once().Return(map[string]string{
					"student-id-1": "Grade 5",
					"student-id-2": "Grade 5",
					"student-id-3": "Grade 6",
				}, nil)

				userRepo.On("GetStudentCurrentGradeByUserIDs", mock.Anything, mock.Anything, []string{"student-id-4", "student-id-5"}).Once().Return(map[string]string{
					"student-id-4": "Grade 4",
					"student-id-5": "Grade 5",
				}, nil)
			},
		},
		{
			name: "user successfully retrieves lessons on calendar but missing location ids",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsOnCalendarRequest{
				CalendarView: lpb.CalendarView_WEEKLY,
				LocationId:   "location-id-1",
				FromDate:     timestamppb.New(now),
				ToDate:       timestamppb.New(now.Add(7 * 24 * time.Hour)),
				Timezone:     "sample-timezone",
			},
			expectedErr: nil,
			expectedResp: &lpb.RetrieveLessonsOnCalendarResponse{
				Items: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson{
					{
						TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
						LessonId:       "lesson-1",
						LessonName:     "Lesson Name 1",
						StartTime:      timestamppb.New(now),
						EndTime:        timestamppb.New(now),
						LessonTeachers: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonTeacher{
							{
								TeacherId:   "teacher-1",
								TeacherName: "teacher-name-1",
							},
							{
								TeacherId:   "teacher-2",
								TeacherName: "teacher-name-2",
							},
						},
						LessonMembers: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonMember{
							{
								StudentId:   "student-id-1",
								CourseId:    "course-1",
								Grade:       "Grade 5",
								StudentName: "student-name 1",
								CourseName:  "course-name-1",
							},
							{
								StudentId:   "student-id-2",
								CourseId:    "course-2",
								Grade:       "Grade 5",
								StudentName: "student-name 2",
								CourseName:  "course-name-2",
							},
							{
								StudentId:   "student-id-3",
								CourseId:    "course-1",
								Grade:       "Grade 6",
								StudentName: "student-name 3",
								CourseName:  "course-name-1",
							},
						},
						SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED,
						LessonClassrooms: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonClassroom{
							{
								ClassroomId:   "classroom-id-1",
								ClassroomName: "classroom-name-1",
							},
							{
								ClassroomId:   "classroom-id-2",
								ClassroomName: "classroom-name-2",
							},
						},
						SchedulerId: "scheduler-id-1",
					},
					{
						TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
						LessonId:       "lesson-2",
						LessonName:     "Lesson Name 2",
						StartTime:      timestamppb.New(now),
						EndTime:        timestamppb.New(now),
						LessonTeachers: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonTeacher{
							{
								TeacherId:   "teacher-1",
								TeacherName: "teacher-name-1",
							},
						},
						LessonMembers: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonMember{
							{
								StudentId:   "student-id-4",
								CourseId:    "course-2",
								Grade:       "Grade 4",
								StudentName: "student-name 4",
								CourseName:  "course-name-2",
							},
							{
								StudentId:   "student-id-5",
								CourseId:    "course-2",
								Grade:       "Grade 5",
								StudentName: "student-name 5",
								CourseName:  "course-name-2",
							},
						},
						SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED,
						LessonClassrooms: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonClassroom{
							{
								ClassroomId:   "classroom-id-1",
								ClassroomName: "classroom-name-1",
							},
							{
								ClassroomId:   "classroom-id-2",
								ClassroomName: "classroom-name-2",
							},
						},
						SchedulerId: "scheduler-id-2",
					},
					{
						TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
						LessonId:       "lesson-3",
						LessonName:     "Lesson Name 3",
						StartTime:      timestamppb.New(now),
						EndTime:        timestamppb.New(now),
						LessonTeachers: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonTeacher{
							{
								TeacherId:   "teacher-1",
								TeacherName: "teacher-name-1",
							},
							{
								TeacherId:   "teacher-2",
								TeacherName: "teacher-name-2",
							},
						},
						ClassId:          "class-1",
						ClassName:        "class-name-1",
						CourseId:         "course-2",
						CourseName:       "course-name-2",
						LessonMembers:    []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonMember{},
						SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED,
						LessonClassrooms: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonClassroom{
							{
								ClassroomId:   "classroom-id-1",
								ClassroomName: "classroom-name-1",
							},
							{
								ClassroomId:   "classroom-id-2",
								ClassroomName: "classroom-name-2",
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonsOnCalendar", mock.Anything, db, &payloads.GetLessonListOnCalendarArgs{
					View:       payloads.Weekly,
					FromDate:   now,
					ToDate:     now.Add(7 * 24 * time.Hour),
					LocationID: "location-id-1",
					Timezone:   "sample-timezone",
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:         "lesson-1",
						Name:             "Lesson Name 1",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-1",
					},
					{
						LessonID:         "lesson-2",
						Name:             "Lesson Name 2",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCanceled,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-2",
					},
					{
						LessonID:         "lesson-3",
						Name:             "Lesson Name 3",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "class-1",
						CourseID:         "course-2",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						CourseName:       "course-name-2",
						ClassName:        "class-name-1",
					},
				}, nil)

				lessonTeacherRepo.On("GetTeachersWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonTeachers{
					"lesson-1": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
					"lesson-2": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
					},
					"lesson-3": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
				}, nil)

				lessonMemberRepo.On("GetLessonLearnersWithCourseAndNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonLearners{
					"lesson-1": {
						&domain.LessonLearner{
							LearnerID:   "student-id-1",
							CourseID:    "course-1",
							LearnerName: "student-name 1",
							CourseName:  "course-name-1",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-2",
							CourseID:    "course-2",
							LearnerName: "student-name 2",
							CourseName:  "course-name-2",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-3",
							CourseID:    "course-1",
							LearnerName: "student-name 3",
							CourseName:  "course-name-1",
						},
					},
					"lesson-2": {
						&domain.LessonLearner{
							LearnerID:   "student-id-4",
							CourseID:    "course-2",
							LearnerName: "student-name 4",
							CourseName:  "course-name-2",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-5",
							CourseID:    "course-2",
							LearnerName: "student-name 5",
							CourseName:  "course-name-2",
						},
					},
				}, nil)

				lessonClassroomRepo.On("GetLessonClassroomsWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs).Once().Return(map[string]domain.LessonClassrooms{
					"lesson-1": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
					"lesson-2": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
					"lesson-3": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
				}, nil)

				userRepo.On("GetStudentCurrentGradeByUserIDs", mock.Anything, mock.Anything, []string{"student-id-1", "student-id-2", "student-id-3"}).Once().Return(map[string]string{
					"student-id-1": "Grade 5",
					"student-id-2": "Grade 5",
					"student-id-3": "Grade 6",
				}, nil)

				userRepo.On("GetStudentCurrentGradeByUserIDs", mock.Anything, mock.Anything, []string{"student-id-4", "student-id-5"}).Once().Return(map[string]string{
					"student-id-4": "Grade 4",
					"student-id-5": "Grade 5",
				}, nil)
			},
		},
		{
			name: "returns empty list",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsOnCalendarRequest{
				CalendarView: lpb.CalendarView_DAILY,
				LocationId:   "location-id-1",
				FromDate:     timestamppb.New(now),
				ToDate:       timestamppb.New(now.Add(7 * 24 * time.Hour)),
				Timezone:     "sample-timezone",
				LocationIds:  []string{"location-id-1", "location-id-2"},
			},
			expectedErr: nil,
			expectedResp: &lpb.RetrieveLessonsOnCalendarResponse{
				Items: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonsOnCalendar", mock.Anything, db, mock.Anything).Once().Return([]*domain.Lesson{}, nil)
			},
		},
		{
			name: "selected location id is not part of the location ids list",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsOnCalendarRequest{
				CalendarView: lpb.CalendarView_DAILY,
				LocationId:   "location-id-1",
				FromDate:     timestamppb.New(now),
				ToDate:       timestamppb.New(now.Add(7 * 24 * time.Hour)),
				Timezone:     "sample-timezone",
				LocationIds:  []string{"location-id-2", "location-id-3"},
			},
			expectedErr: nil,
			expectedResp: &lpb.RetrieveLessonsOnCalendarResponse{
				Items: []*lpb.RetrieveLessonsOnCalendarResponse_Lesson{},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "missing location id on request",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsOnCalendarRequest{
				CalendarView: lpb.CalendarView_DAILY,
				FromDate:     timestamppb.New(now),
				ToDate:       timestamppb.New(now.Add(7 * 24 * time.Hour)),
				Timezone:     "sample-timezone",
				LocationIds:  []string{"location-id-1", "location-id-2"},
			},
			expectedErr: fmt.Errorf("rpc error: code = Internal desc = request missing location ID"),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "missing timezone on request",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsOnCalendarRequest{
				CalendarView: lpb.CalendarView_DAILY,
				LocationId:   "location-id-1",
				FromDate:     timestamppb.New(now),
				ToDate:       timestamppb.New(now.Add(7 * 24 * time.Hour)),
				LocationIds:  []string{"location-id-1", "location-id-2"},
			},
			expectedErr: fmt.Errorf("rpc error: code = Internal desc = request missing timezone"),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "to date is before from date on request",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsOnCalendarRequest{
				CalendarView: lpb.CalendarView_DAILY,
				LocationId:   "location-id-1",
				FromDate:     timestamppb.New(now.Add(7 * 24 * time.Hour)),
				ToDate:       timestamppb.New(now),
				Timezone:     "sample-timezone",
				LocationIds:  []string{"location-id-1", "location-id-2"},
			},
			expectedErr: fmt.Errorf("rpc error: code = Internal desc = to date cannot be before from date"),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "failed to fetch lessons on calendar from lesson repo",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsOnCalendarRequest{
				CalendarView: lpb.CalendarView_WEEKLY,
				LocationId:   "location-id-1",
				FromDate:     timestamppb.New(now),
				ToDate:       timestamppb.New(now.Add(7 * 24 * time.Hour)),
				Timezone:     "sample-timezone",
				LocationIds:  []string{"location-id-1", "location-id-2"},
			},
			expectedErr: fmt.Errorf("rpc error: code = Internal desc = LessonRepo.GetLessonsOnCalendar: some-lesson-repo-error"),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonsOnCalendar", mock.Anything, db, &payloads.GetLessonListOnCalendarArgs{
					View:       payloads.Weekly,
					FromDate:   now,
					ToDate:     now.Add(7 * 24 * time.Hour),
					LocationID: "location-id-1",
					Timezone:   "sample-timezone",
				}).Once().Return(nil, fmt.Errorf("some-lesson-repo-error"))
			},
		},
		{
			name: "failed to fetch teachers by lesson IDs from lesson teacher repo",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsOnCalendarRequest{
				CalendarView: lpb.CalendarView_WEEKLY,
				LocationId:   "location-id-1",
				FromDate:     timestamppb.New(now),
				ToDate:       timestamppb.New(now.Add(7 * 24 * time.Hour)),
				Timezone:     "sample-timezone",
				LocationIds:  []string{"location-id-1", "location-id-2"},
			},
			expectedErr: fmt.Errorf("rpc error: code = Internal desc = LessonTeacherRepo.GetTeachersWithNamesByLessonIDs: some-lesson-teacher-repo-error"),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonsOnCalendar", mock.Anything, db, &payloads.GetLessonListOnCalendarArgs{
					View:       payloads.Weekly,
					FromDate:   now,
					ToDate:     now.Add(7 * 24 * time.Hour),
					LocationID: "location-id-1",
					Timezone:   "sample-timezone",
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:         "lesson-1",
						Name:             "Lesson Name 1",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-1",
					},
					{
						LessonID:         "lesson-2",
						Name:             "Lesson Name 2",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCanceled,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-2",
					},
					{
						LessonID:         "lesson-3",
						Name:             "Lesson Name 3",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "class-1",
						CourseID:         "course-2",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						CourseName:       "course-name-2",
						ClassName:        "class-name-1",
					},
				}, nil)

				lessonTeacherRepo.On("GetTeachersWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().
					Return(map[string]domain.LessonTeachers{}, fmt.Errorf("some-lesson-teacher-repo-error"))
			},
		},
		{
			name: "failed to fetch lesson learners by lesson IDs from lesson member repo",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsOnCalendarRequest{
				CalendarView: lpb.CalendarView_WEEKLY,
				LocationId:   "location-id-1",
				FromDate:     timestamppb.New(now),
				ToDate:       timestamppb.New(now.Add(7 * 24 * time.Hour)),
				Timezone:     "sample-timezone",
				LocationIds:  []string{"location-id-1", "location-id-2"},
			},
			expectedErr: fmt.Errorf("rpc error: code = Internal desc = LessonMemberRepo.GetLessonLearnersWithCourseAndNamesByLessonIDs: some-lesson-member-repo-error"),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonsOnCalendar", mock.Anything, db, &payloads.GetLessonListOnCalendarArgs{
					View:       payloads.Weekly,
					FromDate:   now,
					ToDate:     now.Add(7 * 24 * time.Hour),
					LocationID: "location-id-1",
					Timezone:   "sample-timezone",
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:         "lesson-1",
						Name:             "Lesson Name 1",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-1",
					},
					{
						LessonID:         "lesson-2",
						Name:             "Lesson Name 2",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCanceled,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-2",
					},
					{
						LessonID:         "lesson-3",
						Name:             "Lesson Name 3",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "class-1",
						CourseID:         "course-2",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						CourseName:       "course-name-2",
						ClassName:        "class-name-1",
					},
				}, nil)

				lessonTeacherRepo.On("GetTeachersWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonTeachers{
					"lesson-1": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
					"lesson-2": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
					},
					"lesson-3": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
				}, nil)

				lessonMemberRepo.On("GetLessonLearnersWithCourseAndNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().
					Return(map[string]domain.LessonLearners{}, fmt.Errorf("some-lesson-member-repo-error"))
			},
		},
		{
			name: "failed to fetch lesson classrooms by lesson IDs from lesson classroom repo",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsOnCalendarRequest{
				CalendarView: lpb.CalendarView_WEEKLY,
				LocationId:   "location-id-1",
				FromDate:     timestamppb.New(now),
				ToDate:       timestamppb.New(now.Add(7 * 24 * time.Hour)),
				Timezone:     "sample-timezone",
				LocationIds:  []string{"location-id-1", "location-id-2"},
			},
			expectedErr: fmt.Errorf("rpc error: code = Internal desc = LessonClassroomRepo.GetLessonClassroomsWithNamesByLessonIDs: some-lesson-classroom-repo-error"),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonsOnCalendar", mock.Anything, db, &payloads.GetLessonListOnCalendarArgs{
					View:       payloads.Weekly,
					FromDate:   now,
					ToDate:     now.Add(7 * 24 * time.Hour),
					LocationID: "location-id-1",
					Timezone:   "sample-timezone",
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:         "lesson-1",
						Name:             "Lesson Name 1",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-1",
					},
					{
						LessonID:         "lesson-2",
						Name:             "Lesson Name 2",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCanceled,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-2",
					},
					{
						LessonID:         "lesson-3",
						Name:             "Lesson Name 3",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "class-1",
						CourseID:         "course-2",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						CourseName:       "course-name-2",
						ClassName:        "class-name-1",
					},
				}, nil)

				lessonTeacherRepo.On("GetTeachersWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonTeachers{
					"lesson-1": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
					"lesson-2": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
					},
					"lesson-3": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
				}, nil)

				lessonMemberRepo.On("GetLessonLearnersWithCourseAndNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonLearners{
					"lesson-1": {
						&domain.LessonLearner{
							LearnerID:   "student-id-1",
							CourseID:    "course-1",
							LearnerName: "student-name 1",
							CourseName:  "course-name-1",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-2",
							CourseID:    "course-2",
							LearnerName: "student-name 2",
							CourseName:  "course-name-2",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-3",
							CourseID:    "course-1",
							LearnerName: "student-name 3",
							CourseName:  "course-name-1",
						},
					},
					"lesson-2": {
						&domain.LessonLearner{
							LearnerID:   "student-id-4",
							CourseID:    "course-2",
							LearnerName: "student-name 4",
							CourseName:  "course-name-2",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-5",
							CourseID:    "course-2",
							LearnerName: "student-name 5",
							CourseName:  "course-name-2",
						},
					},
				}, nil)

				lessonClassroomRepo.On("GetLessonClassroomsWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs).Once().
					Return(map[string]domain.LessonClassrooms{}, fmt.Errorf("some-lesson-classroom-repo-error"))
			},
		},
		{
			name: "failed to fetch student grades of individual lessons from user repo",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsOnCalendarRequest{
				CalendarView: lpb.CalendarView_WEEKLY,
				LocationId:   "location-id-1",
				FromDate:     timestamppb.New(now),
				ToDate:       timestamppb.New(now.Add(7 * 24 * time.Hour)),
				Timezone:     "sample-timezone",
				LocationIds:  []string{"location-id-1", "location-id-2"},
			},
			expectedErr: fmt.Errorf("rpc error: code = Internal desc = UserRepo.GetStudentCurrentGradeByUserIDs: some-user-repo-error"),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonsOnCalendar", mock.Anything, db, &payloads.GetLessonListOnCalendarArgs{
					View:       payloads.Weekly,
					FromDate:   now,
					ToDate:     now.Add(7 * 24 * time.Hour),
					LocationID: "location-id-1",
					Timezone:   "sample-timezone",
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:         "lesson-1",
						Name:             "Lesson Name 1",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-1",
					},
					{
						LessonID:         "lesson-2",
						Name:             "Lesson Name 2",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCanceled,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-2",
					},
					{
						LessonID:         "lesson-3",
						Name:             "Lesson Name 3",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "class-1",
						CourseID:         "course-2",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						CourseName:       "course-name-2",
						ClassName:        "class-name-1",
					},
				}, nil)

				lessonTeacherRepo.On("GetTeachersWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonTeachers{
					"lesson-1": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
					"lesson-2": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
					},
					"lesson-3": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
				}, nil)

				lessonMemberRepo.On("GetLessonLearnersWithCourseAndNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonLearners{
					"lesson-1": {
						&domain.LessonLearner{
							LearnerID:   "student-id-1",
							CourseID:    "course-1",
							LearnerName: "student-name 1",
							CourseName:  "course-name-1",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-2",
							CourseID:    "course-2",
							LearnerName: "student-name 2",
							CourseName:  "course-name-2",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-3",
							CourseID:    "course-1",
							LearnerName: "student-name 3",
							CourseName:  "course-name-1",
						},
					},
					"lesson-2": {
						&domain.LessonLearner{
							LearnerID:   "student-id-4",
							CourseID:    "course-2",
							LearnerName: "student-name 4",
							CourseName:  "course-name-2",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-5",
							CourseID:    "course-2",
							LearnerName: "student-name 5",
							CourseName:  "course-name-2",
						},
					},
				}, nil)

				lessonClassroomRepo.On("GetLessonClassroomsWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs).Once().Return(map[string]domain.LessonClassrooms{
					"lesson-1": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
					"lesson-2": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
					"lesson-3": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
				}, nil)

				userRepo.On("GetStudentCurrentGradeByUserIDs", mock.Anything, mock.Anything, []string{"student-id-1", "student-id-2", "student-id-3"}).Once().
					Return(map[string]string{}, fmt.Errorf("some-user-repo-error"))
			},
		},
		{
			name: "failed to fetch student grades of individual lessons from user repo in second loop iteration",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveLessonsOnCalendarRequest{
				CalendarView: lpb.CalendarView_WEEKLY,
				LocationId:   "location-id-1",
				FromDate:     timestamppb.New(now),
				ToDate:       timestamppb.New(now.Add(7 * 24 * time.Hour)),
				Timezone:     "sample-timezone",
				LocationIds:  []string{"location-id-1", "location-id-2"},
			},
			expectedErr: fmt.Errorf("rpc error: code = Internal desc = UserRepo.GetStudentCurrentGradeByUserIDs: some-user-repo-error"),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonsOnCalendar", mock.Anything, db, &payloads.GetLessonListOnCalendarArgs{
					View:       payloads.Weekly,
					FromDate:   now,
					ToDate:     now.Add(7 * 24 * time.Hour),
					LocationID: "location-id-1",
					Timezone:   "sample-timezone",
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:         "lesson-1",
						Name:             "Lesson Name 1",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-1",
					},
					{
						LessonID:         "lesson-2",
						Name:             "Lesson Name 2",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCanceled,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-2",
					},
					{
						LessonID:         "lesson-3",
						Name:             "Lesson Name 3",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "class-1",
						CourseID:         "course-2",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						CourseName:       "course-name-2",
						ClassName:        "class-name-1",
					},
				}, nil)

				lessonTeacherRepo.On("GetTeachersWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonTeachers{
					"lesson-1": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
					"lesson-2": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
					},
					"lesson-3": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
				}, nil)

				lessonMemberRepo.On("GetLessonLearnersWithCourseAndNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonLearners{
					"lesson-1": {
						&domain.LessonLearner{
							LearnerID:   "student-id-1",
							CourseID:    "course-1",
							LearnerName: "student-name 1",
							CourseName:  "course-name-1",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-2",
							CourseID:    "course-2",
							LearnerName: "student-name 2",
							CourseName:  "course-name-2",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-3",
							CourseID:    "course-1",
							LearnerName: "student-name 3",
							CourseName:  "course-name-1",
						},
					},
					"lesson-2": {
						&domain.LessonLearner{
							LearnerID:   "student-id-4",
							CourseID:    "course-2",
							LearnerName: "student-name 4",
							CourseName:  "course-name-2",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-5",
							CourseID:    "course-2",
							LearnerName: "student-name 5",
							CourseName:  "course-name-2",
						},
					},
				}, nil)

				lessonClassroomRepo.On("GetLessonClassroomsWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs).Once().Return(map[string]domain.LessonClassrooms{
					"lesson-1": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
					"lesson-2": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
					"lesson-3": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
				}, nil)

				userRepo.On("GetStudentCurrentGradeByUserIDs", mock.Anything, mock.Anything, []string{"student-id-1", "student-id-2", "student-id-3"}).Once().Return(map[string]string{
					"student-id-1": "Grade 5",
					"student-id-2": "Grade 5",
					"student-id-3": "Grade 6",
				}, nil)

				userRepo.On("GetStudentCurrentGradeByUserIDs", mock.Anything, mock.Anything, []string{"student-id-4", "student-id-5"}).Once().
					Return(map[string]string{}, fmt.Errorf("some-user-repo-error"))
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*lpb.RetrieveLessonsOnCalendarRequest)
			resp, err := lessonReaderService.RetrieveLessonsOnCalendar(testCase.ctx, req)

			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}

			mock.AssertExpectationsForObjects(t, lessonRepo, lessonTeacherRepo, lessonMemberRepo, userRepo, mockUnleashClient)
		})
	}
}
