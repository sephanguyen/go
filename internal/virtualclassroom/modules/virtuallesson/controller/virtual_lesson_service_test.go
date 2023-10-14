package controller_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure/repo"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtuallesson/application/queries"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtuallesson/application/queries/payloads"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtuallesson/controller"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/virtualclassroom/virtualclassroom/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestVirtualLessonReaderService_GetVirtualLessonByID(t *testing.T) {
	t.Parallel()

	lessonRepo := &mock_repositories.MockVirtualLessonRepo{}
	lessonGroupRepo := &mock_repositories.MockLessonGroupRepo{}
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	now := time.Now().UTC()
	nowString, err := now.MarshalText()
	require.NoError(t, err)
	validPresentMaterialJSON := database.JSONB(`
	{
		"current_material": {
			"media_id": "media-1",
			"updated_at": "` + string(nowString) + `",
			"video_state": {
				"current_time": "23m",
				"player_state": "PLAYER_STATE_PLAYING"
			}
		}
	}`)
	jsm := &mock_nats.JetStreamManagement{}

	tcs := []struct {
		name           string
		lessonID       string
		setup          func(ctx context.Context)
		hasError       bool
		expectedLesson *domain.VirtualLesson
	}{
		{
			name:     "retrieve successfully",
			lessonID: "lesson-1",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(3)
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonGroupRepo.
					On("GetByIDAndCourseID", ctx, db, "lesson-group-1", "course-1").
					Return(&repo.LessonGroupDTO{
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						MediaIDs:      database.TextArray([]string{"media-1", "media-2", "media-3"}),
					}, nil).Once()
			},
			expectedLesson: &domain.VirtualLesson{
				LessonID:      "lesson-1",
				LessonGroupID: "lesson-group-1",
				CourseID:      "course-1",
				RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
				LearnerIDs:    domain.LearnerIDs{LearnerIDs: []string{"learner-1", "learner-2", "learner-3"}},
				TeacherIDs:    domain.TeacherIDs{TeacherIDs: []string{"teacher-1", "teacher-2"}},
			},
			hasError: false,
		},
		{
			name:     "error lesson repo",
			lessonID: "lesson-1",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						fmt.Errorf("LessonRepo.FindByID: %s", err),
					).Once()
				lessonGroupRepo.
					On("GetByIDAndCourseID", ctx, db, "lesson-group-1", "course-1").
					Return(&repo.LessonGroupDTO{
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						MediaIDs:      database.TextArray([]string{"media-1", "media-2", "media-3"}),
					}, nil).Once()
			},
			expectedLesson: nil,
			hasError:       true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			tc.setup(ctx)

			jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {}).Return("", nil)

			srv := &controller.VirtualLessonReaderService{
				WrapperDBConnection: wrapperConnection,
				JSM:                 jsm,
				VirtualLessonRepo:   lessonRepo,
				LessonGroupRepo:     lessonGroupRepo,
			}
			lesson, err := srv.GetVirtualLessonByID(ctx, tc.lessonID, controller.IncludeLearnerIDs(), controller.IncludeTeacherIDs())
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expectedLesson, lesson)
			mock.AssertExpectationsForObjects(t, db, tx, mockUnleashClient)
		})
	}
}

func TestVirtualLessonReaderService_GetLiveLessonsByLocations(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	env := "local"
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := &mock_repositories.MockVirtualLessonRepo{}
	lessonMemberRepo := &mock_repositories.MockLessonMemberRepo{}
	courseClassRepo := &mock_repositories.MockCourseClassRepo{}
	oldClassRepo := &mock_repositories.MockOldClassRepo{}
	studentsRepo := &mock_repositories.MockStudentsRepo{}
	configRepo := &mock_repositories.MockConfigRepo{}
	jsm := &mock_nats.JetStreamManagement{}

	teacherID, studentID := "teacher_id1", "student_id1"
	lessonID, courseID := "lesson_id1", "course_id1"
	zoomLink := "sample_link"
	paging := &vpb.Pagination{
		Limit: 10,
		Page:  1,
	}

	now := time.Now().Add(1 * time.Hour)
	startDate := timestamppb.New(now)
	endDate := timestamppb.New(now.Add(7 * 24 * time.Hour))
	startDatePast := timestamppb.New(now.Add(-3 * time.Hour))
	endDatePast := timestamppb.New(now.Add(-2 * time.Hour))
	locationIDs := []string{"location_id1", "location_id2"}
	courseIDs := []string{courseID, "course_id2"}
	studentIDs := []string{studentID}
	status := []domain.LessonSchedulingStatus{domain.LessonSchedulingStatusPublished, domain.LessonSchedulingStatusCompleted}
	statusPB := []cpb.LessonSchedulingStatus{cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED, cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED}
	classID := int32(1234)
	classIDs := []int32{classID}

	configKey := "specificCourseIDsForLesson"
	configGroup := "lesson"
	configCountry := domain.CountryMaster
	resourcePath := "1"
	whitelistCourseIDs := []string{"course_id1", "course_id3"}
	configs := []*domain.Config{
		{
			Key:       configKey,
			Group:     configGroup,
			Country:   configCountry,
			Value:     "course_id1,course_id3",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	tcs := []struct {
		name         string
		userID       string
		req          *vpb.GetLiveLessonsByLocationsRequest
		expectedResp *vpb.GetLiveLessonsByLocationsResponse
		setup        func(ctx context.Context)
		hasError     bool
	}{
		{
			name:   "teacher get live lessons successfully",
			userID: teacherID,
			req: &vpb.GetLiveLessonsByLocationsRequest{
				CourseIds:        courseIDs,
				LocationIds:      locationIDs,
				SchedulingStatus: statusPB,
				From:             startDate,
				To:               endDate,
				Pagination:       paging,
			},
			expectedResp: &vpb.GetLiveLessonsByLocationsResponse{
				Lessons: []*vpb.Lesson{
					{
						LessonId:                 lessonID,
						CourseId:                 courseID,
						PresetStudyPlanWeeklyIds: "",
						Topic: &cpb.Topic{
							Attachments: []*cpb.Attachment{},
						},
						StartTime: startDate,
						EndTime:   endDate,
						ZoomLink:  zoomLink,
						Teacher: []*cpb.BasicProfile{
							{
								UserId: teacherID,
							},
						},
						Status:         cpb.LessonStatus_LESSON_STATUS_IN_PROGRESS,
						TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE,
					},
					{
						LessonId:                 lessonID,
						CourseId:                 courseID,
						PresetStudyPlanWeeklyIds: "",
						Topic: &cpb.Topic{
							Attachments: []*cpb.Attachment{},
						},
						StartTime: startDatePast,
						EndTime:   endDatePast,
						ZoomLink:  zoomLink,
						Teacher: []*cpb.BasicProfile{
							{
								UserId: teacherID,
							},
						},
						Status:         cpb.LessonStatus_LESSON_STATUS_COMPLETED,
						TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_CLASS_DO,
					},
				},
				Total: int32(1),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, teacherID).Once().
					Return(false, nil)

				params := &payloads.GetVirtualLessonsArgs{
					CourseIDs:                courseIDs,
					LocationIDs:              locationIDs,
					StartDate:                startDate.AsTime(),
					EndDate:                  endDate.AsTime(),
					ReplaceCourseIDColumn:    true,
					Limit:                    paging.Limit,
					Page:                     paging.Page,
					LessonSchedulingStatuses: status,
				}

				lessonRepo.On("GetVirtualLessons", ctx, mock.Anything, params).Once().
					Return([]*domain.VirtualLesson{
						{
							LessonID:       lessonID,
							CourseID:       courseID,
							TeacherID:      teacherID,
							StartTime:      startDate.AsTime(),
							EndTime:        endDate.AsTime(),
							EndAt:          nil,
							ZoomLink:       zoomLink,
							TeachingMedium: domain.LessonTeachingMediumOnline,
						},
						{
							LessonID:       lessonID,
							CourseID:       courseID,
							TeacherID:      teacherID,
							StartTime:      startDatePast.AsTime(),
							EndTime:        endDatePast.AsTime(),
							EndAt:          nil,
							ZoomLink:       zoomLink,
							TeachingMedium: domain.LessonTeachingMediumClassDo,
						},
					}, int32(1), nil)
			},
		},
		{
			name:   "teacher get live lessons successfully with no teachers in lessons",
			userID: teacherID,
			req: &vpb.GetLiveLessonsByLocationsRequest{
				CourseIds:        courseIDs,
				LocationIds:      locationIDs,
				SchedulingStatus: statusPB,
				From:             startDate,
				To:               endDate,
				Pagination:       paging,
			},
			expectedResp: &vpb.GetLiveLessonsByLocationsResponse{
				Lessons: []*vpb.Lesson{
					{
						LessonId:                 lessonID,
						CourseId:                 courseID,
						PresetStudyPlanWeeklyIds: "",
						Topic: &cpb.Topic{
							Attachments: []*cpb.Attachment{},
						},
						StartTime:      startDate,
						EndTime:        endDate,
						ZoomLink:       zoomLink,
						Status:         cpb.LessonStatus_LESSON_STATUS_IN_PROGRESS,
						TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ZOOM,
					},
				},
				Total: int32(1),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, teacherID).Once().
					Return(false, nil)

				params := &payloads.GetVirtualLessonsArgs{
					CourseIDs:                courseIDs,
					LocationIDs:              locationIDs,
					StartDate:                startDate.AsTime(),
					EndDate:                  endDate.AsTime(),
					ReplaceCourseIDColumn:    true,
					Limit:                    paging.Limit,
					Page:                     paging.Page,
					LessonSchedulingStatuses: status,
				}

				lessonRepo.On("GetVirtualLessons", ctx, mock.Anything, params).Once().
					Return([]*domain.VirtualLesson{
						{
							LessonID:       lessonID,
							CourseID:       courseID,
							StartTime:      startDate.AsTime(),
							EndTime:        endDate.AsTime(),
							EndAt:          nil,
							ZoomLink:       zoomLink,
							TeachingMedium: domain.LessonTeachingMediumZoom,
						},
					}, int32(1), nil)
			},
		},
		{
			name:   "student get live lessons successfully",
			userID: studentID,
			req: &vpb.GetLiveLessonsByLocationsRequest{
				LocationIds: locationIDs,
				From:        startDate,
				To:          endDate,
				Pagination:  paging,
			},
			expectedResp: &vpb.GetLiveLessonsByLocationsResponse{
				Lessons: []*vpb.Lesson{
					{
						LessonId:                 lessonID,
						CourseId:                 courseID,
						PresetStudyPlanWeeklyIds: "",
						Topic: &cpb.Topic{
							Attachments: []*cpb.Attachment{},
						},
						StartTime: startDate,
						EndTime:   endDate,
						ZoomLink:  zoomLink,
						Teacher: []*cpb.BasicProfile{
							{
								UserId: teacherID,
							},
						},
						Status:         cpb.LessonStatus_LESSON_STATUS_IN_PROGRESS,
						TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_HYBRID,
					},
				},
				Total: int32(1),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Twice()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, studentID).Once().
					Return(true, nil)

				oldClassRepo.On("FindJoined", ctx, mock.Anything, studentID).Once().
					Return(domain.OldClasses{
						{
							ID: classID,
						},
					}, nil)

				courseClassRepo.On("FindActiveCourseClassByID", ctx, mock.Anything, classIDs).Once().
					Return(nil, nil)

				lessonMemberRepo.On("GetCourseAccessible", ctx, mock.Anything, studentID).Once().
					Return(courseIDs, nil)

				params := &payloads.GetVirtualLessonsArgs{
					StudentIDs:            studentIDs,
					CourseIDs:             courseIDs,
					LocationIDs:           locationIDs,
					StartDate:             startDate.AsTime(),
					EndDate:               endDate.AsTime(),
					ReplaceCourseIDColumn: false,
					Limit:                 paging.Limit,
					Page:                  paging.Page,
				}

				lessonRepo.On("GetVirtualLessons", ctx, mock.Anything, params).Once().
					Return([]*domain.VirtualLesson{
						{
							LessonID:       lessonID,
							CourseID:       courseID,
							TeacherID:      teacherID,
							StartTime:      startDate.AsTime(),
							EndTime:        endDate.AsTime(),
							EndAt:          nil,
							ZoomLink:       zoomLink,
							TeachingMedium: domain.LessonTeachingMediumHybrid,
						},
					}, int32(1), nil)
			},
		},
		{
			name:   "student get live lessons successfully with whitelist",
			userID: studentID,
			req: &vpb.GetLiveLessonsByLocationsRequest{
				LocationIds:      locationIDs,
				SchedulingStatus: statusPB,
				From:             startDate,
				To:               endDate,
				Pagination:       paging,
			},
			expectedResp: &vpb.GetLiveLessonsByLocationsResponse{
				Lessons: []*vpb.Lesson{
					{
						LessonId:                 lessonID,
						CourseId:                 courseID,
						PresetStudyPlanWeeklyIds: "",
						Topic: &cpb.Topic{
							Attachments: []*cpb.Attachment{},
						},
						StartTime: startDate,
						EndTime:   endDate,
						ZoomLink:  zoomLink,
						Teacher: []*cpb.BasicProfile{
							{
								UserId: teacherID,
							},
						},
						Status:         cpb.LessonStatus_LESSON_STATUS_IN_PROGRESS,
						TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
					},
				},
				Total: int32(1),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(true, nil).Once()

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, studentID).Once().
					Return(true, nil)

				configRepo.On("GetConfigWithResourcePath", ctx, mock.Anything, configCountry, configGroup, []string{configKey}, resourcePath).
					Return(configs, nil).Once()

				params := &payloads.GetVirtualLessonsArgs{
					StudentIDs:               studentIDs,
					CourseIDs:                whitelistCourseIDs,
					LocationIDs:              locationIDs,
					StartDate:                startDate.AsTime(),
					EndDate:                  endDate.AsTime(),
					ReplaceCourseIDColumn:    false,
					Limit:                    paging.Limit,
					Page:                     paging.Page,
					LessonSchedulingStatuses: status,
				}

				lessonRepo.On("GetVirtualLessons", ctx, mock.Anything, params).Once().
					Return([]*domain.VirtualLesson{
						{
							LessonID:       lessonID,
							CourseID:       courseID,
							TeacherID:      teacherID,
							StartTime:      startDate.AsTime(),
							EndTime:        endDate.AsTime(),
							EndAt:          nil,
							ZoomLink:       zoomLink,
							TeachingMedium: domain.LessonTeachingMediumOffline,
						},
					}, int32(1), nil)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithUserID(ctx, tc.userID)
			ctx = golibs.ResourcePathToCtx(ctx, resourcePath)
			tc.setup(ctx)

			query := queries.VirtualLessonQuery{
				LessonmgmtDB:        db,
				WrapperDBConnection: wrapperConnection,
				VirtualLessonRepo:   lessonRepo,
				LessonMemberRepo:    lessonMemberRepo,
				CourseClassRepo:     courseClassRepo,
				OldClassRepo:        oldClassRepo,
				StudentsRepo:        studentsRepo,
				ConfigRepo:          configRepo,
			}
			service := &controller.VirtualLessonReaderService{
				WrapperDBConnection: wrapperConnection,
				JSM:                 jsm,
				Env:                 env,
				VirtualLessonQuery:  query,
				VirtualLessonRepo:   lessonRepo,
				UnleashClient:       mockUnleashClient,
			}

			response, err := service.GetLiveLessonsByLocations(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expectedResp, response)
			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo, lessonMemberRepo, courseClassRepo, oldClassRepo, studentsRepo, mockUnleashClient)
		})
	}
}

func TestVirtualLessonReaderService_GetLearnersByLessonID(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := &mock_repositories.MockVirtualLessonRepo{}
	lessonMemberRepo := &mock_repositories.MockLessonMemberRepo{}
	studentESHRepo := &mock_repositories.MockStudentEnrollmentStatusHistoryRepo{}

	userID, lessonID, locationID := "main_user1", "lesson_id1", "location_id1"
	now := time.Now()
	startDate := now
	endDate := now.Add(7 * 24 * time.Hour)

	paging := &cpb.Paging{
		Limit: 15,
		Offset: &cpb.Paging_OffsetMultipleCombined{
			OffsetMultipleCombined: &cpb.Paging_MultipleCombined{
				Combined: []*cpb.Paging_Combined{
					{
						OffsetString: "last_lesson_course_id",
					},
					{
						OffsetString: "last_user_id",
					},
				},
			},
		},
	}

	tcs := []struct {
		name         string
		userID       string
		req          *vpb.GetLearnersByLessonIDRequest
		expectedResp *vpb.GetLearnersByLessonIDResponse
		setup        func(ctx context.Context)
		hasError     bool
	}{
		{
			name:   "user get learners successfully",
			userID: userID,
			req: &vpb.GetLearnersByLessonIDRequest{
				LessonId: lessonID,
				Paging:   paging,
			},
			expectedResp: &vpb.GetLearnersByLessonIDResponse{
				Learners: []*vpb.LearnerInfo{
					{
						LearnerId: "learner_id1",
						EnrollmentStatusInfo: []*vpb.LearnerInfo_EnrollmentStatusInfo{
							{
								LocationId: locationID,
								StartDate:  timestamppb.New(startDate),
								EndDate:    timestamppb.New(endDate),
							},
						},
					},
					{
						LearnerId: "learner_id2",
						EnrollmentStatusInfo: []*vpb.LearnerInfo_EnrollmentStatusInfo{
							{
								LocationId: locationID,
								StartDate:  timestamppb.New(startDate),
								EndDate:    timestamppb.New(endDate),
							},
						},
					},
					{
						LearnerId: "learner_id3",
						EnrollmentStatusInfo: []*vpb.LearnerInfo_EnrollmentStatusInfo{
							{
								LocationId: locationID,
								StartDate:  timestamppb.New(startDate),
								EndDate:    timestamppb.New(endDate),
							},
						},
					},
				},
				NextPage: &cpb.Paging{
					Limit: paging.Limit,
					Offset: &cpb.Paging_OffsetMultipleCombined{
						OffsetMultipleCombined: &cpb.Paging_MultipleCombined{
							Combined: []*cpb.Paging_Combined{
								{
									OffsetString: lessonID + "course_id3",
								},
								{
									OffsetString: "learner_id3",
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
				studentIDs := []string{
					"learner_id1",
					"learner_id2",
					"learner_id3",
				}

				params := &payloads.GetLearnersByLessonIDArgs{
					LessonID:       lessonID,
					Limit:          int32(paging.Limit),
					LessonCourseID: "last_lesson_course_id",
					UserID:         "last_user_id",
				}

				lessonMemberRepo.On("GetLearnersByLessonIDWithPaging", ctx, mock.Anything, params).Once().
					Return([]domain.LessonMember{
						{
							LessonID: lessonID,
							UserID:   "learner_id1",
							CourseID: "course_id1",
						},
						{
							LessonID: lessonID,
							UserID:   "learner_id2",
							CourseID: "course_id2",
						},
						{
							LessonID: lessonID,
							UserID:   "learner_id3",
							CourseID: "course_id3",
						},
					}, nil)

				lessonRepo.On("GetVirtualLessonOnlyByID", ctx, mock.Anything, params.LessonID).Once().
					Return(&domain.VirtualLesson{
						LessonID: lessonID,
						CenterID: locationID,
					}, nil)
				studentESHRepo.On("GetStatusHistoryByStudentIDsAndLocationID", ctx, mock.Anything, studentIDs, locationID).Once().
					Return(domain.StudentEnrollmentStatusHistories{
						{
							StudentID:  "learner_id1",
							LocationID: locationID,
							StartDate:  startDate,
							EndDate:    endDate,
						},
						{
							StudentID:  "learner_id2",
							LocationID: locationID,
							StartDate:  startDate,
							EndDate:    endDate,
						},
						{
							StudentID:  "learner_id3",
							LocationID: locationID,
							StartDate:  startDate,
							EndDate:    endDate,
						},
					}, nil)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithUserID(ctx, tc.userID)
			tc.setup(ctx)

			query := queries.VirtualLessonQuery{
				LessonmgmtDB:                 db,
				WrapperDBConnection:          wrapperConnection,
				VirtualLessonRepo:            lessonRepo,
				LessonMemberRepo:             lessonMemberRepo,
				StudentEnrollmentHistoryRepo: studentESHRepo,
			}
			service := &controller.VirtualLessonReaderService{
				WrapperDBConnection: wrapperConnection,
				VirtualLessonQuery:  query,
				UnleashClient:       mockUnleashClient,
				Env:                 "local",
			}

			response, err := service.GetLearnersByLessonID(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expectedResp, response)
			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo, lessonMemberRepo, studentESHRepo, mockUnleashClient)
		})
	}
}

func TestVirtualLessonReaderService_GetLessons(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := &mock_repositories.MockVirtualLessonRepo{}
	lessonTeacherRepo := &mock_repositories.MockLessonTeacherRepo{}

	userID := "sample-id1"
	schoolID := "123456"
	now := time.Now()

	tcs := []struct {
		name         string
		userID       string
		req          *vpb.GetLessonsRequest
		expectedResp *vpb.GetLessonsResponse
		setup        func(ctx context.Context)
		hasError     bool
	}{
		{
			name:   "user get lessons with empty paging",
			userID: userID,
			req: &vpb.GetLessonsRequest{
				CurrentTime:       timestamppb.New(now),
				LessonTimeCompare: vpb.LessonTimeCompare_LESSON_TIME_COMPARE_FUTURE,
				TimeLookup:        vpb.TimeLookup_TIME_LOOKUP_END_TIME,
				SortAsc:           true,
			},
			expectedResp: nil,
			setup:        func(ctx context.Context) {},
			hasError:     true,
		},
		{
			name:   "user get lessons with empty current time",
			userID: userID,
			req: &vpb.GetLessonsRequest{
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "lesson_id1",
					},
				},
				LessonTimeCompare: vpb.LessonTimeCompare_LESSON_TIME_COMPARE_FUTURE,
				TimeLookup:        vpb.TimeLookup_TIME_LOOKUP_END_TIME,
				SortAsc:           true,
			},
			expectedResp: nil,
			setup:        func(ctx context.Context) {},
			hasError:     true,
		},
		{
			name:   "user get lessons successfully",
			userID: userID,
			req: &vpb.GetLessonsRequest{
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "lesson_id1",
					},
				},
				CurrentTime:       timestamppb.New(now),
				LessonTimeCompare: vpb.LessonTimeCompare_LESSON_TIME_COMPARE_FUTURE,
				TimeLookup:        vpb.TimeLookup_TIME_LOOKUP_END_TIME,
				SortAsc:           true,
			},
			expectedResp: &vpb.GetLessonsResponse{
				Items: []*vpb.GetLessonsResponse_Lesson{
					{
						Id:               "lesson-id1",
						Name:             "lesson 1",
						CenterId:         "location-id1",
						StartTime:        timestamppb.New(now),
						EndTime:          timestamppb.New(now),
						TeacherIds:       []string{"teacher-id1", "teacher-id2"},
						TeachingMethod:   cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
						TeachingMedium:   cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE,
						CourseId:         "course-id1",
						ClassId:          "",
						SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
						LessonCapacity:   32,
						EndAt:            timestamppb.New(now),
						ZoomLink:         "zoom-link-1",
					},
					{
						Id:               "lesson-id2",
						Name:             "lesson 2",
						CenterId:         "location-id2",
						StartTime:        timestamppb.New(now),
						EndTime:          timestamppb.New(now),
						TeacherIds:       []string{"teacher-id3", "teacher-id4"},
						TeachingMethod:   cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
						TeachingMedium:   cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ZOOM,
						CourseId:         "course-id2",
						ClassId:          "class-id2",
						SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
						LessonCapacity:   31,
						EndAt:            timestamppb.New(now),
						ZoomLink:         "zoom-link-2",
					},
					{
						Id:               "lesson-id3",
						Name:             "lesson 3",
						CenterId:         "location-id3",
						StartTime:        timestamppb.New(now),
						EndTime:          timestamppb.New(now),
						TeacherIds:       []string{"teacher-id1", "teacher-id3"},
						TeachingMethod:   cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
						TeachingMedium:   cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_HYBRID,
						CourseId:         "",
						ClassId:          "",
						SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED,
						LessonCapacity:   30,
						EndAt:            nil,
						ZoomLink:         "",
					},
					{
						Id:               "lesson-id4",
						Name:             "lesson 4",
						CenterId:         "location-id4",
						StartTime:        timestamppb.New(now),
						EndTime:          timestamppb.New(now),
						TeacherIds:       []string{"teacher-id1", "teacher-id4"},
						TeachingMethod:   cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
						TeachingMedium:   cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE,
						CourseId:         "",
						ClassId:          "",
						SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT,
						LessonCapacity:   29,
						EndAt:            timestamppb.New(now),
						ZoomLink:         "",
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "",
					},
				},
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "lesson-id4",
					},
				},
				TotalLesson: 4,
				TotalItems:  4,
			},
			setup: func(ctx context.Context) {
				lessonRepo.On("GetLessons", ctx, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						payload := args.Get(2).(payloads.GetLessonsArgs)
						assert.Equal(t, payload.CurrentTime.UTC(), now.UTC())
						assert.Equal(t, payload.LessonTimeCompare, payloads.LessonTimeCompareFuture)
						assert.Equal(t, payload.TimeLookup, payloads.TimeLookupEndTime)
						assert.Equal(t, payload.SchoolID, schoolID)
						assert.Equal(t, payload.OffsetLessonID, "lesson_id1")
						assert.Equal(t, payload.Limit, uint32(10))
						assert.True(t, payload.SortAscending)
						assert.Empty(t, payload.LocationIDs)
						assert.Empty(t, payload.TeacherIDs)
						assert.Empty(t, payload.StudentIDs)
						assert.Empty(t, payload.CourseIDs)
						assert.Empty(t, payload.LessonSchedulingStatuses)
						assert.Equal(t, payload.LiveLessonStatus, payloads.LiveLessonStatusNone)
						assert.Empty(t, payload.FromDate)
						assert.Empty(t, payload.ToDate)
					}).
					Return([]domain.VirtualLesson{
						{
							LessonID:         "lesson-id1",
							Name:             "lesson 1",
							CenterID:         "location-id1",
							StartTime:        now,
							EndTime:          now,
							TeachingMethod:   domain.LessonTeachingMethodIndividual,
							TeachingMedium:   domain.LessonTeachingMediumOnline,
							CourseID:         "course-id1",
							ClassID:          "",
							SchedulingStatus: domain.LessonSchedulingStatusPublished,
							LessonCapacity:   32,
							EndAt:            &now,
							ZoomLink:         "zoom-link-1",
						},
						{
							LessonID:         "lesson-id2",
							Name:             "lesson 2",
							CenterID:         "location-id2",
							StartTime:        now,
							EndTime:          now,
							TeachingMethod:   domain.LessonTeachingMethodGroup,
							TeachingMedium:   domain.LessonTeachingMediumZoom,
							CourseID:         "course-id2",
							ClassID:          "class-id2",
							SchedulingStatus: domain.LessonSchedulingStatusPublished,
							LessonCapacity:   31,
							EndAt:            &now,
							ZoomLink:         "zoom-link-2",
						},
						{
							LessonID:         "lesson-id3",
							Name:             "lesson 3",
							CenterID:         "location-id3",
							StartTime:        now,
							EndTime:          now,
							TeachingMethod:   domain.LessonTeachingMethodGroup,
							TeachingMedium:   domain.LessonTeachingMediumHybrid,
							CourseID:         "",
							ClassID:          "",
							SchedulingStatus: domain.LessonSchedulingStatusCompleted,
							LessonCapacity:   30,
							EndAt:            nil,
							ZoomLink:         "",
						},
						{
							LessonID:         "lesson-id4",
							Name:             "lesson 4",
							CenterID:         "location-id4",
							StartTime:        now,
							EndTime:          now,
							TeachingMethod:   domain.LessonTeachingMethodIndividual,
							TeachingMedium:   domain.LessonTeachingMediumOnline,
							CourseID:         "",
							ClassID:          "",
							SchedulingStatus: domain.LessonSchedulingStatusDraft,
							LessonCapacity:   29,
							EndAt:            &now,
							ZoomLink:         "",
						},
					}, uint32(4), "", uint32(4), nil).Once()

				lessonTeacherRepo.On("GetTeacherIDsOnlyByLessonIDs", ctx, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						lessonIDs := args.Get(2).([]string)
						assert.Equal(t, lessonIDs, []string{"lesson-id1", "lesson-id2", "lesson-id3", "lesson-id4"})
					}).
					Return(map[string][]string{
						"lesson-id1": {"teacher-id1", "teacher-id2"},
						"lesson-id2": {"teacher-id3", "teacher-id4"},
						"lesson-id3": {"teacher-id1", "teacher-id3"},
						"lesson-id4": {"teacher-id1", "teacher-id4"},
					}, nil).Once()
			},
		},
		{
			name:   "user get lessons successfully with filter",
			userID: userID,
			req: &vpb.GetLessonsRequest{
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "lesson_id1",
					},
				},
				CurrentTime:       timestamppb.New(now),
				LessonTimeCompare: vpb.LessonTimeCompare_LESSON_TIME_COMPARE_PAST_AND_EQUAL,
				TimeLookup:        vpb.TimeLookup_TIME_LOOKUP_END_TIME_INCLUDE_WITHOUT_END_AT,
				SortAsc:           false,
				Filter: &vpb.GetLessonsFilter{
					TeacherIds: []string{"teacher-id1", "teacher-id2", "teacher-id3", "teacher-id4"},
					StudentIds: []string{"student-id1", "student-id2"},
					CourseIds:  []string{"course-id1", "course-id2", "course-id3", "course-id4"},
					SchedulingStatus: []cpb.LessonSchedulingStatus{
						cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
						cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED,
						cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT,
						cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED,
					},
					FromDate:         timestamppb.New(now),
					ToDate:           timestamppb.New(now),
					LiveLessonStatus: vpb.LiveLessonStatus_LIVE_LESSON_STATUS_ENDED,
				},
			},
			expectedResp: &vpb.GetLessonsResponse{
				Items: []*vpb.GetLessonsResponse_Lesson{
					{
						Id:               "lesson-id1",
						Name:             "lesson 1",
						CenterId:         "location-id1",
						StartTime:        timestamppb.New(now),
						EndTime:          timestamppb.New(now),
						TeacherIds:       []string{"teacher-id1", "teacher-id2"},
						TeachingMethod:   cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
						TeachingMedium:   cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE,
						CourseId:         "course-id1",
						ClassId:          "",
						SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
						LessonCapacity:   32,
						EndAt:            timestamppb.New(now),
						ZoomLink:         "zoom-link-1",
					},
					{
						Id:               "lesson-id2",
						Name:             "lesson 2",
						CenterId:         "location-id2",
						StartTime:        timestamppb.New(now),
						EndTime:          timestamppb.New(now),
						TeacherIds:       []string{"teacher-id3", "teacher-id4"},
						TeachingMethod:   cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
						TeachingMedium:   cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ZOOM,
						CourseId:         "course-id2",
						ClassId:          "class-id2",
						SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
						LessonCapacity:   31,
						EndAt:            timestamppb.New(now),
						ZoomLink:         "zoom-link-2",
					},
					{
						Id:               "lesson-id3",
						Name:             "lesson 3",
						CenterId:         "location-id3",
						StartTime:        timestamppb.New(now),
						EndTime:          timestamppb.New(now),
						TeacherIds:       []string{"teacher-id1", "teacher-id3"},
						TeachingMethod:   cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
						TeachingMedium:   cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_HYBRID,
						CourseId:         "",
						ClassId:          "",
						SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED,
						LessonCapacity:   30,
						EndAt:            nil,
						ZoomLink:         "",
					},
					{
						Id:               "lesson-id4",
						Name:             "lesson 4",
						CenterId:         "location-id4",
						StartTime:        timestamppb.New(now),
						EndTime:          timestamppb.New(now),
						TeacherIds:       []string{"teacher-id1", "teacher-id4"},
						TeachingMethod:   cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
						TeachingMedium:   cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE,
						CourseId:         "",
						ClassId:          "",
						SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT,
						LessonCapacity:   29,
						EndAt:            timestamppb.New(now),
						ZoomLink:         "",
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "",
					},
				},
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "lesson-id4",
					},
				},
				TotalLesson: 4,
				TotalItems:  4,
			},
			setup: func(ctx context.Context) {
				lessonRepo.On("GetLessons", ctx, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						payload := args.Get(2).(payloads.GetLessonsArgs)
						assert.Equal(t, payload.CurrentTime.UTC(), now.UTC())
						assert.Equal(t, payload.LessonTimeCompare, payloads.LessonTimeComparePastAndEqual)
						assert.Equal(t, payload.TimeLookup, payloads.TimeLookupEndTimeIncludeWithoutEndAt)
						assert.Equal(t, payload.SchoolID, schoolID)
						assert.Equal(t, payload.OffsetLessonID, "lesson_id1")
						assert.Equal(t, payload.Limit, uint32(10))
						assert.Equal(t, payload.TeacherIDs, []string{"teacher-id1", "teacher-id2", "teacher-id3", "teacher-id4"})
						assert.Equal(t, payload.StudentIDs, []string{"student-id1", "student-id2"})
						assert.Equal(t, payload.CourseIDs, []string{"course-id1", "course-id2", "course-id3", "course-id4"})
						assert.Equal(t, payload.LessonSchedulingStatuses, []domain.LessonSchedulingStatus{
							domain.LessonSchedulingStatusPublished,
							domain.LessonSchedulingStatusCompleted,
							domain.LessonSchedulingStatusDraft,
							domain.LessonSchedulingStatusCanceled,
						})
						assert.Equal(t, payload.LiveLessonStatus, payloads.LiveLessonStatusEnded)
						assert.Equal(t, payload.FromDate.UTC(), now.UTC())
						assert.Equal(t, payload.ToDate.UTC(), now.UTC())
						assert.False(t, payload.SortAscending)
					}).
					Return([]domain.VirtualLesson{
						{
							LessonID:         "lesson-id1",
							Name:             "lesson 1",
							CenterID:         "location-id1",
							StartTime:        now,
							EndTime:          now,
							TeachingMethod:   domain.LessonTeachingMethodIndividual,
							TeachingMedium:   domain.LessonTeachingMediumOnline,
							CourseID:         "course-id1",
							ClassID:          "",
							SchedulingStatus: domain.LessonSchedulingStatusPublished,
							LessonCapacity:   32,
							EndAt:            &now,
							ZoomLink:         "zoom-link-1",
						},
						{
							LessonID:         "lesson-id2",
							Name:             "lesson 2",
							CenterID:         "location-id2",
							StartTime:        now,
							EndTime:          now,
							TeachingMethod:   domain.LessonTeachingMethodGroup,
							TeachingMedium:   domain.LessonTeachingMediumZoom,
							CourseID:         "course-id2",
							ClassID:          "class-id2",
							SchedulingStatus: domain.LessonSchedulingStatusPublished,
							LessonCapacity:   31,
							EndAt:            &now,
							ZoomLink:         "zoom-link-2",
						},
						{
							LessonID:         "lesson-id3",
							Name:             "lesson 3",
							CenterID:         "location-id3",
							StartTime:        now,
							EndTime:          now,
							TeachingMethod:   domain.LessonTeachingMethodGroup,
							TeachingMedium:   domain.LessonTeachingMediumHybrid,
							CourseID:         "",
							ClassID:          "",
							SchedulingStatus: domain.LessonSchedulingStatusCompleted,
							LessonCapacity:   30,
							EndAt:            nil,
							ZoomLink:         "",
						},
						{
							LessonID:         "lesson-id4",
							Name:             "lesson 4",
							CenterID:         "location-id4",
							StartTime:        now,
							EndTime:          now,
							TeachingMethod:   domain.LessonTeachingMethodIndividual,
							TeachingMedium:   domain.LessonTeachingMediumOnline,
							CourseID:         "",
							ClassID:          "",
							SchedulingStatus: domain.LessonSchedulingStatusDraft,
							LessonCapacity:   29,
							EndAt:            &now,
							ZoomLink:         "",
						},
					}, uint32(4), "", uint32(4), nil).Once()

				lessonTeacherRepo.On("GetTeacherIDsOnlyByLessonIDs", ctx, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						lessonIDs := args.Get(2).([]string)
						assert.Equal(t, lessonIDs, []string{"lesson-id1", "lesson-id2", "lesson-id3", "lesson-id4"})
					}).
					Return(map[string][]string{
						"lesson-id1": {"teacher-id1", "teacher-id2"},
						"lesson-id2": {"teacher-id3", "teacher-id4"},
						"lesson-id3": {"teacher-id1", "teacher-id3"},
						"lesson-id4": {"teacher-id1", "teacher-id4"},
					}, nil).Once()
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithUserID(ctx, tc.userID)
			ctxWithResourcePath := golibs.ResourcePathToCtx(ctx, schoolID)
			tc.setup(ctxWithResourcePath)

			query := queries.VirtualLessonQuery{
				LessonmgmtDB:        db,
				WrapperDBConnection: wrapperConnection,
				VirtualLessonRepo:   lessonRepo,
				LessonTeacherRepo:   lessonTeacherRepo,
			}
			service := &controller.VirtualLessonReaderService{
				WrapperDBConnection: wrapperConnection,
				VirtualLessonQuery:  query,
			}

			response, err := service.GetLessons(ctxWithResourcePath, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expectedResp, response)
			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo, lessonTeacherRepo, mockUnleashClient)
		})
	}
}

func TestVirtualLessonReaderService_GetClassDoURL(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := &mock_repositories.MockVirtualLessonRepo{}

	lessonID := "lesson_id1"
	classDoLink := "sample-link.com"

	tcs := []struct {
		name         string
		req          *vpb.GetClassDoURLRequest
		expectedResp *vpb.GetClassDoURLResponse
		setup        func(ctx context.Context)
		hasError     bool
	}{
		{
			name: "user get classdo link successfully",
			req: &vpb.GetClassDoURLRequest{
				LessonId: lessonID,
			},
			expectedResp: &vpb.GetClassDoURLResponse{
				ClassdoLink: classDoLink,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()

				lessonRepo.On("GetVirtualLessonOnlyByID", ctx, mock.Anything, lessonID).Once().
					Return(&domain.VirtualLesson{
						LessonID:    lessonID,
						ClassDoLink: classDoLink,
					}, nil)
			},
		},
		{
			name: "user get classdo link with error",
			req: &vpb.GetClassDoURLRequest{
				LessonId: lessonID,
			},
			expectedResp: nil,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()

				lessonRepo.On("GetVirtualLessonOnlyByID", ctx, mock.Anything, lessonID).Once().
					Return(nil, fmt.Errorf("error"))
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			tc.setup(ctx)

			query := queries.VirtualLessonQuery{
				LessonmgmtDB:        db,
				WrapperDBConnection: wrapperConnection,
				VirtualLessonRepo:   lessonRepo,
			}
			service := &controller.VirtualLessonReaderService{
				WrapperDBConnection: wrapperConnection,
				VirtualLessonQuery:  query,
				UnleashClient:       mockUnleashClient,
				Env:                 "local",
			}

			response, err := service.GetClassDoURL(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expectedResp, response)
			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo, mockUnleashClient)
		})
	}
}
