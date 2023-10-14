package controller

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	calendar_constants "github.com/manabie-com/backend/internal/calendar/domain/constants"
	calendar_dto "github.com/manabie-com/backend/internal/calendar/domain/dto"
	calendar_entities "github.com/manabie-com/backend/internal/calendar/domain/entities"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/commands"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/producers"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	user_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	zoom_service "github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/service"
	calendar_mock_repositories "github.com/manabie-com/backend/mock/calendar/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_media_module_adapter "github.com/manabie-com/backend/mock/lessonmgmt/lesson/media_module_adapter"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	mock_user_module_adapter "github.com/manabie-com/backend/mock/lessonmgmt/lesson/usermodadapter"
	mock_lesson_report_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson_report/repositories"
	mock_user_repo "github.com/manabie-com/backend/mock/lessonmgmt/user/repositories"
	mock_clients "github.com/manabie-com/backend/mock/lessonmgmt/zoom/clients"
	mock_zoom_repo "github.com/manabie-com/backend/mock/lessonmgmt/zoom/repositories"
	mock_service "github.com/manabie-com/backend/mock/lessonmgmt/zoom/service"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestLessonManagementGRPCService_CreateLessonIndividual(t *testing.T) {
	t.Parallel()
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	jsm := &mock_nats.JetStreamManagement{}
	masterDataRepo := new(mock_repositories.MockMasterDataRepo)
	mediaModulePort := new(mock_media_module_adapter.MockMediaModuleAdapter)
	userModuleAdapter := new(mock_user_module_adapter.MockUserModuleAdapter)
	lessonRepo := new(mock_repositories.MockLessonRepo)
	courseRepo := new(mock_repositories.MockCourseRepo)
	schedulerRepo := new(calendar_mock_repositories.MockSchedulerRepo)
	dateInfoRepo := new(calendar_mock_repositories.MockDateInfoRepo)
	dateInfos := []*calendar_dto.DateInfo{}
	studentSubscriptionRepo := new(mock_user_repo.MockStudentSubscriptionRepo)
	userRepo := new(mock_user_repo.MockUserRepo)
	classroomRepo := new(mock_repositories.MockClassroomRepo)
	reallocationRepo := new(mock_repositories.MockReallocationRepo)
	zoomAccountRepo := new(mock_zoom_repo.MockZoomAccountRepo)
	mockExternalConfigService := &mock_service.MockExternalConfigService{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	mockSchedulerClient := &mock_clients.MockSchedulerClient{}
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")

	mockHTTPClient := &mock_clients.MockHTTPClient{}
	zcf := &configs.ZoomConfig{}

	zoomService := zoom_service.InitZoomService(zcf, mockExternalConfigService, mockHTTPClient)
	tcs := []struct {
		name     string
		req      *lpb.CreateLessonRequest
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "full fields",
			req: &lpb.CreateLessonRequest{
				StartTime:      timestamppb.New(now),
				EndTime:        timestamppb.New(now),
				TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
				TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
				TeacherIds:     []string{"teacher-id-1", "teacher-id-2"},
				ClassroomIds:   []string{"classroom-id-1", "classroom-id-2"},
				LocationId:     "center-id-1",
				StudentInfoList: []*lpb.CreateLessonRequest_StudentInfo{
					{
						StudentId:        "user-id-1",
						CourseId:         "course-id-1",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						AttendanceNotice: lpb.StudentAttendanceNotice_NOTICE_EMPTY,
						AttendanceReason: lpb.StudentAttendanceReason_REASON_EMPTY,
						LocationId:       "center-id-1",
					},
					{
						StudentId:        "user-id-2",
						CourseId:         "course-id-2",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_EMPTY,
						AttendanceNotice: lpb.StudentAttendanceNotice_NOTICE_EMPTY,
						AttendanceReason: lpb.StudentAttendanceReason_REASON_EMPTY,
						LocationId:       "center-id-1",
					},
					{
						StudentId:        "user-id-3",
						CourseId:         "course-id-3",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
						AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE,
						AttendanceReason: lpb.StudentAttendanceReason_PHYSICAL_CONDITION,
						AttendanceNote:   "sample-attendance-note",
						LocationId:       "center-id-1",
					},
				},
				Materials: []*lpb.Material{
					{
						Resource: &lpb.Material_MediaId{
							MediaId: "media-id-1",
						},
					},
					{
						Resource: &lpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
				},
				SavingOption: &lpb.CreateLessonRequest_SavingOption{
					Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
				},
				ClassId:  "",
				CourseId: "",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, tx, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, tx, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()
				mediaModulePort.
					On(
						"RetrieveMediasByIDs",
						ctx,
						[]string{"media-id-1", "media-id-2"},
					).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).
					Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						now.UTC(),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2", "user-id-3", "course-id-3"},
					).
					Return(nil).
					Once()
				classroomRepo.
					On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				// student
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-1",
					).
					Return(&user_domain.User{ID: "user-id-1", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-2",
					).
					Return(&user_domain.User{ID: "user-id-2", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-3",
					).
					Return(&user_domain.User{ID: "user-id-3", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				// teacher
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-1",
					).
					Return(&user_domain.User{ID: "teacher-id-1", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-2",
					).
					Return(&user_domain.User{ID: "teacher-id-2", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				expectedLesson := &domain.Lesson{
					LessonID:         "test-id-1",
					LocationID:       "center-id-1",
					StartTime:        now.UTC(),
					EndTime:          now.UTC(),
					SchedulingStatus: domain.LessonSchedulingStatusPublished,
					TeachingMedium:   domain.LessonTeachingMediumOffline,
					TeachingMethod:   domain.LessonTeachingMethodIndividual,
					Learners: domain.LessonLearners{
						{
							LearnerID:        "user-id-1",
							CourseID:         "course-id-1",
							AttendStatus:     domain.StudentAttendStatusAttend,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.NoticeEmpty,
							AttendanceReason: domain.ReasonEmpty,
						},
						{
							LearnerID:        "user-id-2",
							CourseID:         "course-id-2",
							AttendStatus:     domain.StudentAttendStatusEmpty,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.NoticeEmpty,
							AttendanceReason: domain.ReasonEmpty,
						},
						{
							LearnerID:        "user-id-3",
							CourseID:         "course-id-3",
							AttendStatus:     domain.StudentAttendStatusAbsent,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.InAdvance,
							AttendanceReason: domain.PhysicalCondition,
							AttendanceNote:   "sample-attendance-note",
						},
					},
					Teachers: domain.LessonTeachers{
						{
							TeacherID: "teacher-id-1",
						},
						{
							TeacherID: "teacher-id-2",
						},
					},
					Material: &domain.LessonMaterial{
						MediaIDs: []string{"media-id-1", "media-id-2"},
					},
					ClassID:   "",
					CourseID:  "",
					DateInfos: dateInfos,
					Classrooms: domain.LessonClassrooms{
						{
							ClassroomID: "classroom-id-1",
						},
						{
							ClassroomID: "classroom-id-2",
						},
					},
					PreparationTime: -1,
					BreakTime:       -1,
				}

				mockSchedulerClient.On("CreateScheduler", ctx, mock.Anything).Once().Return(&mpb.CreateSchedulerResponse{
					SchedulerId: "SchedulerID",
				}, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				lessonRepo.
					On("InsertLesson", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						actualLesson := args[2].(*domain.Lesson)
						actualLesson.LessonID = expectedLesson.LessonID
						actualLesson.MasterDataPort = nil
						actualLesson.UserModulePort = nil
						actualLesson.MediaModulePort = nil
						actualLesson.DateInfoRepo = nil
						actualLesson.DateInfos = dateInfos
						actualLesson.ClassroomRepo = nil

						assert.NotEmpty(t, actualLesson.LessonID)
						expectedLesson.SchedulerID = actualLesson.SchedulerID
						expectedLesson.CreatedAt = actualLesson.CreatedAt
						expectedLesson.UpdatedAt = actualLesson.CreatedAt
						assert.EqualValues(t, expectedLesson, actualLesson)
					}).Return(expectedLesson, nil).Once()
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectLessonCreated, mock.MatchedBy(func(data []byte) bool {
					r := &bpb.EvtLesson{}
					if err := proto.Unmarshal(data, r); err != nil {
						return false
					}
					if len(r.GetCreateLessons().Lessons) > 0 {
						lesson := r.GetCreateLessons().Lessons[0]
						return lesson.LessonId == expectedLesson.LessonID && lesson.Name == expectedLesson.Name && reflect.DeepEqual(lesson.LearnerIds, expectedLesson.GetLearnersIDs())
					}
					return false
				})).Return("", nil).Once()

			},
		},
		{
			name: "missing material",
			req: &lpb.CreateLessonRequest{
				StartTime:      timestamppb.New(now),
				EndTime:        timestamppb.New(now),
				TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
				TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
				TeacherIds:     []string{"teacher-id-1", "teacher-id-2"},
				ClassroomIds:   []string{"classroom-id-1", "classroom-id-2"},
				LocationId:     "center-id-1",
				StudentInfoList: []*lpb.CreateLessonRequest_StudentInfo{
					{
						StudentId:        "user-id-1",
						CourseId:         "course-id-1",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						AttendanceNotice: lpb.StudentAttendanceNotice_NOTICE_EMPTY,
						AttendanceReason: lpb.StudentAttendanceReason_REASON_EMPTY,
						LocationId:       "center-id-1",
					},
					{
						StudentId:        "user-id-2",
						CourseId:         "course-id-2",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_EMPTY,
						AttendanceNotice: lpb.StudentAttendanceNotice_NOTICE_EMPTY,
						AttendanceReason: lpb.StudentAttendanceReason_REASON_EMPTY,
						LocationId:       "center-id-1",
					},
					{
						StudentId:        "user-id-3",
						CourseId:         "course-id-3",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
						AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE,
						AttendanceReason: lpb.StudentAttendanceReason_PHYSICAL_CONDITION,
						AttendanceNote:   "sample-attendance-note",
						LocationId:       "center-id-1",
					},
				},
				SavingOption: &lpb.CreateLessonRequest_SavingOption{
					Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
				},
				ClassId:  "",
				CourseId: "",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, tx, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, tx, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).
					Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						now.UTC(),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2", "user-id-3", "course-id-3"},
					).
					Return(nil).
					Once()
				classroomRepo.
					On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				// student
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-1",
					).
					Return(&user_domain.User{ID: "user-id-1", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-2",
					).
					Return(&user_domain.User{ID: "user-id-2", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-3",
					).
					Return(&user_domain.User{ID: "user-id-3", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				// teacher
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-1",
					).
					Return(&user_domain.User{ID: "teacher-id-1", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-2",
					).
					Return(&user_domain.User{ID: "teacher-id-2", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				expectedLesson := &domain.Lesson{
					LessonID:         "test-id-1",
					LocationID:       "center-id-1",
					StartTime:        now.UTC(),
					EndTime:          now.UTC(),
					SchedulingStatus: domain.LessonSchedulingStatusPublished,
					TeachingMedium:   domain.LessonTeachingMediumOffline,
					TeachingMethod:   domain.LessonTeachingMethodIndividual,
					Learners: domain.LessonLearners{
						{
							LearnerID:        "user-id-1",
							CourseID:         "course-id-1",
							AttendStatus:     domain.StudentAttendStatusAttend,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.NoticeEmpty,
							AttendanceReason: domain.ReasonEmpty,
						},
						{
							LearnerID:        "user-id-2",
							CourseID:         "course-id-2",
							AttendStatus:     domain.StudentAttendStatusEmpty,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.NoticeEmpty,
							AttendanceReason: domain.ReasonEmpty,
						},
						{
							LearnerID:        "user-id-3",
							CourseID:         "course-id-3",
							AttendStatus:     domain.StudentAttendStatusAbsent,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.InAdvance,
							AttendanceReason: domain.PhysicalCondition,
							AttendanceNote:   "sample-attendance-note",
						},
					},
					Teachers: domain.LessonTeachers{
						{
							TeacherID: "teacher-id-1",
						},
						{
							TeacherID: "teacher-id-2",
						},
					},
					ClassID:   "",
					CourseID:  "",
					DateInfos: dateInfos,
					Classrooms: domain.LessonClassrooms{
						{
							ClassroomID: "classroom-id-1",
						},
						{
							ClassroomID: "classroom-id-2",
						},
					},
					PreparationTime: -1,
					BreakTime:       -1,
				}

				mockSchedulerClient.On("CreateScheduler", ctx, mock.Anything).Once().Return(&mpb.CreateSchedulerResponse{
					SchedulerId: "SchedulerID",
				}, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				lessonRepo.
					On("InsertLesson", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						actualLesson := args[2].(*domain.Lesson)
						actualLesson.LessonID = expectedLesson.LessonID
						actualLesson.MasterDataPort = nil
						actualLesson.UserModulePort = nil
						actualLesson.MediaModulePort = nil
						actualLesson.DateInfoRepo = nil
						actualLesson.ClassroomRepo = nil

						assert.NotEmpty(t, actualLesson.LessonID)
						expectedLesson.SchedulerID = actualLesson.SchedulerID
						expectedLesson.CreatedAt = actualLesson.CreatedAt
						expectedLesson.UpdatedAt = actualLesson.CreatedAt
						assert.EqualValues(t, expectedLesson, actualLesson)
					}).Return(expectedLesson, nil).Once()
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectLessonCreated, mock.MatchedBy(func(data []byte) bool {
					r := &bpb.EvtLesson{}
					if err := proto.Unmarshal(data, r); err != nil {
						return false
					}
					if len(r.GetCreateLessons().Lessons) > 0 {
						lesson := r.GetCreateLessons().Lessons[0]
						return lesson.LessonId == expectedLesson.LessonID && lesson.Name == expectedLesson.Name && reflect.DeepEqual(lesson.LearnerIds, expectedLesson.GetLearnersIDs())
					}
					return false
				})).Return("", nil).Once()

			},
		},
		{
			name: "insert lesson failed",
			req: &lpb.CreateLessonRequest{
				StartTime:      timestamppb.New(now),
				EndTime:        timestamppb.New(now),
				TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
				TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
				TeacherIds:     []string{"teacher-id-1", "teacher-id-2"},
				ClassroomIds:   []string{"classroom-id-1", "classroom-id-2"},
				LocationId:     "center-id-1",
				StudentInfoList: []*lpb.CreateLessonRequest_StudentInfo{
					{
						StudentId:        "user-id-1",
						CourseId:         "course-id-1",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						AttendanceNotice: lpb.StudentAttendanceNotice_NOTICE_EMPTY,
						AttendanceReason: lpb.StudentAttendanceReason_REASON_EMPTY,
						LocationId:       "center-id-1",
					},
					{
						StudentId:        "user-id-2",
						CourseId:         "course-id-2",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_EMPTY,
						AttendanceNotice: lpb.StudentAttendanceNotice_NOTICE_EMPTY,
						AttendanceReason: lpb.StudentAttendanceReason_REASON_EMPTY,
						LocationId:       "center-id-1",
					},
					{
						StudentId:        "user-id-3",
						CourseId:         "course-id-3",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
						AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE,
						AttendanceReason: lpb.StudentAttendanceReason_PHYSICAL_CONDITION,
						AttendanceNote:   "sample-attendance-note",
						LocationId:       "center-id-1",
					},
				},
				Materials: []*lpb.Material{
					{
						Resource: &lpb.Material_MediaId{
							MediaId: "media-id-1",
						},
					},
					{
						Resource: &lpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
				},
				SavingOption: &lpb.CreateLessonRequest_SavingOption{
					Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
				},
				ClassId:  "",
				CourseId: "",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, tx, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, tx, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()
				mediaModulePort.
					On(
						"RetrieveMediasByIDs",
						ctx,
						[]string{"media-id-1", "media-id-2"},
					).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).
					Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						now.UTC(),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2", "user-id-3", "course-id-3"},
					).
					Return(nil).
					Once()
				classroomRepo.
					On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				// student
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-1",
					).
					Return(&user_domain.User{ID: "user-id-1", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-2",
					).
					Return(&user_domain.User{ID: "user-id-2", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-3",
					).
					Return(&user_domain.User{ID: "user-id-3", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				// teacher
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-1",
					).
					Return(&user_domain.User{ID: "teacher-id-1", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-2",
					).
					Return(&user_domain.User{ID: "teacher-id-2", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				expectedLesson := &domain.Lesson{
					LessonID:         "test-id-1",
					LocationID:       "center-id-1",
					StartTime:        now.UTC(),
					EndTime:          now.UTC(),
					SchedulingStatus: domain.LessonSchedulingStatusPublished,
					TeachingMedium:   domain.LessonTeachingMediumOffline,
					TeachingMethod:   domain.LessonTeachingMethodIndividual,
					Learners: domain.LessonLearners{
						{
							LearnerID:        "user-id-1",
							CourseID:         "course-id-1",
							AttendStatus:     domain.StudentAttendStatusAttend,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.NoticeEmpty,
							AttendanceReason: domain.ReasonEmpty,
						},
						{
							LearnerID:        "user-id-2",
							CourseID:         "course-id-2",
							AttendStatus:     domain.StudentAttendStatusEmpty,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.NoticeEmpty,
							AttendanceReason: domain.ReasonEmpty,
						},
						{
							LearnerID:        "user-id-3",
							CourseID:         "course-id-3",
							AttendStatus:     domain.StudentAttendStatusAbsent,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.InAdvance,
							AttendanceReason: domain.PhysicalCondition,
							AttendanceNote:   "sample-attendance-note",
						},
					},
					Teachers: domain.LessonTeachers{
						{
							TeacherID: "teacher-id-1",
						},
						{
							TeacherID: "teacher-id-2",
						},
					},
					Material: &domain.LessonMaterial{
						MediaIDs: []string{"media-id-1", "media-id-2"},
					},
					ClassID:   "",
					CourseID:  "",
					DateInfos: dateInfos,
					Classrooms: domain.LessonClassrooms{
						{
							ClassroomID: "classroom-id-1",
						},
						{
							ClassroomID: "classroom-id-2",
						},
					},
					PreparationTime: -1,
					BreakTime:       -1,
				}
				mockSchedulerClient.On("CreateScheduler", ctx, mock.Anything).Once().Return(&mpb.CreateSchedulerResponse{
					SchedulerId: "SchedulerID",
				}, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				lessonRepo.
					On("InsertLesson", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						actualLesson := args[2].(*domain.Lesson)
						actualLesson.MasterDataPort = nil
						actualLesson.UserModulePort = nil
						actualLesson.MediaModulePort = nil
						actualLesson.DateInfoRepo = nil
						actualLesson.ClassroomRepo = nil
						actualLesson.LessonID = expectedLesson.LessonID

						assert.NotEmpty(t, actualLesson.LessonID)
						expectedLesson.SchedulerID = actualLesson.SchedulerID
						expectedLesson.CreatedAt = actualLesson.CreatedAt
						expectedLesson.UpdatedAt = actualLesson.CreatedAt
						assert.EqualValues(t, expectedLesson, actualLesson)
					}).Return(nil, errors.New("could not insert lesson")).Once()
			},
			hasError: true,
		},
		{
			name: "duplicated learner",
			req: &lpb.CreateLessonRequest{
				StartTime:      timestamppb.New(now),
				EndTime:        timestamppb.New(now),
				TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
				TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
				TeacherIds:     []string{"teacher-id-1", "teacher-id-2"},
				ClassroomIds:   []string{"classroom-id-1", "classroom-id-2"},
				LocationId:     "center-id-1",
				StudentInfoList: []*lpb.CreateLessonRequest_StudentInfo{
					{
						StudentId:        "user-id-1",
						CourseId:         "course-id-1",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						AttendanceNotice: lpb.StudentAttendanceNotice_NOTICE_EMPTY,
						AttendanceReason: lpb.StudentAttendanceReason_REASON_EMPTY,
					},
					{
						StudentId:        "user-id-1",
						CourseId:         "course-id-2",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
						AttendanceNotice: lpb.StudentAttendanceNotice_ON_THE_DAY,
						AttendanceReason: lpb.StudentAttendanceReason_SCHOOL_EVENT,
					},
					{
						StudentId:        "user-id-2",
						CourseId:         "course-id-2",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_EMPTY,
						AttendanceNotice: lpb.StudentAttendanceNotice_NOTICE_EMPTY,
						AttendanceReason: lpb.StudentAttendanceReason_REASON_EMPTY,
					},
					{
						StudentId:        "user-id-3",
						CourseId:         "course-id-3",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
						AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE,
						AttendanceReason: lpb.StudentAttendanceReason_PHYSICAL_CONDITION,
						AttendanceNote:   "sample-attendance-note",
					},
				},
				Materials: []*lpb.Material{
					{
						Resource: &lpb.Material_MediaId{
							MediaId: "media-id-1",
						},
					},
					{
						Resource: &lpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
				},
				SavingOption: &lpb.CreateLessonRequest_SavingOption{
					Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
				},
				ClassId:  "",
				CourseId: "",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()

				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, tx, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()
				mockSchedulerClient.On("CreateScheduler", ctx, mock.Anything).Once().Return(&mpb.CreateSchedulerResponse{
					SchedulerId: "SchedulerID",
				}, nil)
				// student
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-1",
					).
					Return(&user_domain.User{ID: "user-id-1", Group: "USER_GROUP_STUDENT"}, nil)
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-2",
					).
					Return(&user_domain.User{ID: "user-id-2", Group: "USER_GROUP_STUDENT"}, nil)
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-3",
					).
					Return(&user_domain.User{ID: "user-id-3", Group: "USER_GROUP_STUDENT"}, nil)
				// teacher
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-1",
					).
					Return(&user_domain.User{ID: "teacher-id-1", Group: "USER_GROUP_TEACHER"}, nil)
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-2",
					).
					Return(&user_domain.User{ID: "teacher-id-2", Group: "USER_GROUP_TEACHER"}, nil)
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-3",
					).
					Return(&user_domain.User{ID: "teacher-id-3", Group: "USER_GROUP_TEACHER"}, nil)
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			srv := NewLessonModifierService(
				wrapperConnection,
				jsm,
				lessonRepo,
				masterDataRepo,
				userModuleAdapter,
				mediaModulePort,
				dateInfoRepo,
				classroomRepo,
				nil,
				"local",
				mockUnleashClient,
				schedulerRepo,
				studentSubscriptionRepo,
				reallocationRepo,
				nil,
				zoomService,
				zoomAccountRepo,
				nil,
				nil,
				mockSchedulerClient,
				nil,
			)
			res, err := srv.CreateLesson(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, res.Id)
			}
			mock.AssertExpectationsForObjects(
				t,
				db,
				tx,
				jsm,
				masterDataRepo,
				userModuleAdapter,
				mediaModulePort,
				courseRepo,
				mockUnleashClient,
			)
		})
	}
}

func Test_LessonModifierService_UpdateLessonSchedulingStatus(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := new(mock_repositories.MockLessonRepo)
	jsm := &mock_nats.JetStreamManagement{}
	now := time.Now()
	s := &LessonModifierService{
		wrapperConnection: wrapperConnection,
		RetrieveLessonCommand: application.RetrieveLessonCommand{
			WrapperConnection: wrapperConnection,
			LessonRepo:        lessonRepo,
		},
		LessonProducer: producers.LessonProducer{
			JSM: jsm,
		},
		LessonCommandHandler: commands.LessonCommandHandler{
			WrapperConnection: wrapperConnection,
			LessonRepo:        lessonRepo,
		},
	}
	testCases := []TestCase{
		{
			name:         "happy case",
			ctx:          interceptors.ContextWithUserID(ctx, "id"),
			req:          &lpb.UpdateLessonSchedulingStatusRequest{LessonId: LessonTest1, SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED},
			expectedErr:  nil,
			expectedResp: &lpb.UpdateLessonSchedulingStatusResponse{},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonWithSchedulerInfoByLessonID", ctx, db, LessonTest1).
					Return(&domain.Lesson{
						LessonID:         LessonTest1,
						LocationID:       "location-id-1",
						CreatedAt:        now,
						UpdatedAt:        now,
						SchedulingStatus: domain.LessonSchedulingStatusDraft,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
					}, nil).Once()
				lessonRepo.On("UpdateLessonSchedulingStatus", ctx, db, mock.Anything).
					Return(&domain.Lesson{
						LessonID:         LessonTest1,
						LocationID:       "location-id-1",
						CreatedAt:        now,
						UpdatedAt:        now,
						SchedulingStatus: domain.LessonSchedulingStatusPublished,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
					}, nil)
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectLessonUpdated, mock.MatchedBy(func(data []byte) bool {
					evt := &bpb.EvtLesson{}
					if err := proto.Unmarshal(data, evt); err != nil {
						return false
					}
					l := evt.GetUpdateLesson()
					if domain.LessonSchedulingStatus(l.SchedulingStatusBefore.String()) != domain.LessonSchedulingStatusDraft {
						return false
					}
					if domain.LessonSchedulingStatus(l.SchedulingStatusAfter.String()) != domain.LessonSchedulingStatusPublished {
						return false
					}
					return true
				})).Return("", nil).Once()
			},
		},
		{
			name: "error case",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.UpdateLessonSchedulingStatusRequest{
				LessonId:         LessonTest1,
				SchedulingStatus: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED},
			expectedErr: fmt.Errorf("rpc error: code = Internal desc = l.LessonRepo.GetLessonByID: errSubString"),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonWithSchedulerInfoByLessonID", ctx, db, LessonTest1).
					Return(nil, fmt.Errorf("errSubString")).Once()
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*lpb.UpdateLessonSchedulingStatusRequest)
			resp, err := s.UpdateLessonSchedulingStatus(testCase.ctx, req)
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

func TestLessonManagementGRPCService_CreateLessonGroup(t *testing.T) {
	t.Parallel()
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	jsm := &mock_nats.JetStreamManagement{}
	masterDataRepo := new(mock_repositories.MockMasterDataRepo)
	mediaModulePort := new(mock_media_module_adapter.MockMediaModuleAdapter)
	userModuleAdapter := new(mock_user_module_adapter.MockUserModuleAdapter)
	lessonRepo := new(mock_repositories.MockLessonRepo)
	courseRepo := new(mock_repositories.MockCourseRepo)
	schedulerRepo := new(calendar_mock_repositories.MockSchedulerRepo)
	dateInfoRepo := new(calendar_mock_repositories.MockDateInfoRepo)
	dateInfos := []*calendar_dto.DateInfo{}
	studentSubscriptionRepo := new(mock_user_repo.MockStudentSubscriptionRepo)
	classroomRepo := new(mock_repositories.MockClassroomRepo)
	userRepo := new(mock_user_repo.MockUserRepo)
	reallocationRepo := new(mock_repositories.MockReallocationRepo)
	zoomAccountRepo := new(mock_zoom_repo.MockZoomAccountRepo)
	mockExternalConfigService := &mock_service.MockExternalConfigService{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	mockHTTPClient := &mock_clients.MockHTTPClient{}
	zcf := &configs.ZoomConfig{}
	mockSchedulerClient := &mock_clients.MockSchedulerClient{}
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")

	zoomService := zoom_service.InitZoomService(zcf, mockExternalConfigService, mockHTTPClient)

	tcs := []struct {
		name     string
		req      *lpb.CreateLessonRequest
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "full fields",
			req: &lpb.CreateLessonRequest{
				StartTime:      timestamppb.New(now),
				EndTime:        timestamppb.New(now),
				TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
				TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
				TeacherIds:     []string{"teacher-id-1", "teacher-id-2"},
				ClassroomIds:   []string{"classroom-id-1", "classroom-id-2"},
				LocationId:     "center-id-1",
				StudentInfoList: []*lpb.CreateLessonRequest_StudentInfo{
					{
						StudentId:        "user-id-1",
						CourseId:         "course-id-1",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						AttendanceNotice: lpb.StudentAttendanceNotice_NOTICE_EMPTY,
						AttendanceReason: lpb.StudentAttendanceReason_REASON_EMPTY,
						LocationId:       "center-id-1",
					},
					{
						StudentId:        "user-id-2",
						CourseId:         "course-id-2",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_EMPTY,
						AttendanceNotice: lpb.StudentAttendanceNotice_NOTICE_EMPTY,
						AttendanceReason: lpb.StudentAttendanceReason_REASON_EMPTY,
						LocationId:       "center-id-1",
					},
					{
						StudentId:        "user-id-3",
						CourseId:         "course-id-3",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
						AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE,
						AttendanceReason: lpb.StudentAttendanceReason_PHYSICAL_CONDITION,
						AttendanceNote:   "sample-attendance-note",
						LocationId:       "center-id-1",
					},
				},
				Materials: []*lpb.Material{
					{
						Resource: &lpb.Material_MediaId{
							MediaId: "media-id-1",
						},
					},
					{
						Resource: &lpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
				},
				SavingOption: &lpb.CreateLessonRequest_SavingOption{
					Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
				},
				ClassId:  "mock-class-id-1",
				CourseId: "mock-course-id-1",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, tx, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, tx, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()
				masterDataRepo.
					On("GetCourseByID", ctx, tx, "mock-course-id-1").
					Return(&domain.Course{
						CourseID:  "mock-course-id-1",
						Name:      "mock course name 1",
						UpdatedAt: now,
						CreatedAt: now,
					}, nil).Once()
				masterDataRepo.
					On("GetClassByID", ctx, tx, "mock-class-id-1").
					Return(&domain.Class{
						ClassID:   "mock-class-id-1",
						Name:      "mock course name 1",
						UpdatedAt: now,
						CreatedAt: now,
					}, nil).Once()
				classroomRepo.
					On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				mediaModulePort.
					On(
						"RetrieveMediasByIDs",
						ctx,
						[]string{"media-id-1", "media-id-2"},
					).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				// student
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-1",
					).
					Return(&user_domain.User{ID: "user-id-1", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-2",
					).
					Return(&user_domain.User{ID: "user-id-2", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-3",
					).
					Return(&user_domain.User{ID: "user-id-3", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				// teacher
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-1",
					).
					Return(&user_domain.User{ID: "teacher-id-1", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-2",
					).
					Return(&user_domain.User{ID: "teacher-id-2", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).
					Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						now.UTC(),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2", "user-id-3", "course-id-3"},
					).
					Return(nil).
					Once()

				expectedLesson := &domain.Lesson{
					LessonID:         "test-id-1",
					LocationID:       "center-id-1",
					StartTime:        now.UTC(),
					EndTime:          now.UTC(),
					SchedulingStatus: domain.LessonSchedulingStatusPublished,
					TeachingMedium:   domain.LessonTeachingMediumOffline,
					TeachingMethod:   domain.LessonTeachingMethodGroup,
					Learners: domain.LessonLearners{
						{
							LearnerID:        "user-id-1",
							CourseID:         "course-id-1",
							AttendStatus:     domain.StudentAttendStatusAttend,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.NoticeEmpty,
							AttendanceReason: domain.ReasonEmpty,
						},
						{
							LearnerID:        "user-id-2",
							CourseID:         "course-id-2",
							AttendStatus:     domain.StudentAttendStatusEmpty,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.NoticeEmpty,
							AttendanceReason: domain.ReasonEmpty,
						},
						{
							LearnerID:        "user-id-3",
							CourseID:         "course-id-3",
							AttendStatus:     domain.StudentAttendStatusAbsent,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.InAdvance,
							AttendanceReason: domain.PhysicalCondition,
							AttendanceNote:   "sample-attendance-note",
						},
					},
					Teachers: domain.LessonTeachers{
						{
							TeacherID: "teacher-id-1",
						},
						{
							TeacherID: "teacher-id-2",
						},
					},
					Material: &domain.LessonMaterial{
						MediaIDs: []string{"media-id-1", "media-id-2"},
					},
					ClassID:   "mock-class-id-1",
					CourseID:  "mock-course-id-1",
					DateInfos: dateInfos,
					Classrooms: domain.LessonClassrooms{
						{
							ClassroomID: "classroom-id-1",
						},
						{
							ClassroomID: "classroom-id-2",
						},
					},
					PreparationTime: -1,
					BreakTime:       -1,
				}

				mockSchedulerClient.On("CreateScheduler", ctx, mock.Anything).Once().Return(&mpb.CreateSchedulerResponse{
					SchedulerId: "SchedulerID",
				}, nil)

				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				lessonRepo.
					On("InsertLesson", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						actualLesson := args[2].(*domain.Lesson)
						actualLesson.LessonID = expectedLesson.LessonID
						actualLesson.MasterDataPort = nil
						actualLesson.UserModulePort = nil
						actualLesson.MediaModulePort = nil
						actualLesson.DateInfoRepo = nil
						actualLesson.ClassroomRepo = nil

						assert.NotEmpty(t, actualLesson.LessonID)
						expectedLesson.SchedulerID = actualLesson.SchedulerID
						expectedLesson.CreatedAt = actualLesson.CreatedAt
						expectedLesson.UpdatedAt = actualLesson.CreatedAt
						assert.EqualValues(t, expectedLesson, actualLesson)
					}).Return(expectedLesson, nil).Once()
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectLessonCreated, mock.MatchedBy(func(data []byte) bool {
					r := &bpb.EvtLesson{}
					if err := proto.Unmarshal(data, r); err != nil {
						return false
					}
					if len(r.GetCreateLessons().Lessons) > 0 {
						lesson := r.GetCreateLessons().Lessons[0]
						return lesson.LessonId == expectedLesson.LessonID && lesson.Name == expectedLesson.Name && reflect.DeepEqual(lesson.LearnerIds, expectedLesson.GetLearnersIDs())
					}
					return false
				})).Return("", nil).Once()
			},
		},
		{
			name: "missing material",
			req: &lpb.CreateLessonRequest{
				StartTime:      timestamppb.New(now),
				EndTime:        timestamppb.New(now),
				TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
				TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
				TeacherIds:     []string{"teacher-id-1", "teacher-id-2"},
				ClassroomIds:   []string{"classroom-id-1", "classroom-id-2"},
				LocationId:     "center-id-1",
				StudentInfoList: []*lpb.CreateLessonRequest_StudentInfo{
					{
						StudentId:        "user-id-1",
						CourseId:         "course-id-1",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						AttendanceNotice: lpb.StudentAttendanceNotice_NOTICE_EMPTY,
						AttendanceReason: lpb.StudentAttendanceReason_REASON_EMPTY,
						LocationId:       "center-id-1",
					},
					{
						StudentId:        "user-id-2",
						CourseId:         "course-id-2",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_EMPTY,
						AttendanceNotice: lpb.StudentAttendanceNotice_NOTICE_EMPTY,
						AttendanceReason: lpb.StudentAttendanceReason_REASON_EMPTY,
						LocationId:       "center-id-1",
					},
					{
						StudentId:        "user-id-3",
						CourseId:         "course-id-3",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
						AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE,
						AttendanceReason: lpb.StudentAttendanceReason_PHYSICAL_CONDITION,
						AttendanceNote:   "sample-attendance-note",
						LocationId:       "center-id-1",
					},
				},
				SavingOption: &lpb.CreateLessonRequest_SavingOption{
					Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
				},
				ClassId:  "mock-class-id-1",
				CourseId: "mock-course-id-1",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, tx, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, tx, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()
				masterDataRepo.
					On("GetCourseByID", ctx, tx, "mock-course-id-1").
					Return(&domain.Course{
						CourseID:  "mock-course-id-1",
						Name:      "mock course name 1",
						UpdatedAt: now,
						CreatedAt: now,
					}, nil).Once()
				masterDataRepo.
					On("GetClassByID", ctx, tx, "mock-class-id-1").
					Return(&domain.Class{
						ClassID:   "mock-class-id-1",
						Name:      "mock course name 1",
						UpdatedAt: now,
						CreatedAt: now,
					}, nil).Once()
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).
					Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						now.UTC(),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2", "user-id-3", "course-id-3"},
					).
					Return(nil).
					Once()
				classroomRepo.
					On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				// student
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-1",
					).
					Return(&user_domain.User{ID: "user-id-1", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-2",
					).
					Return(&user_domain.User{ID: "user-id-2", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-3",
					).
					Return(&user_domain.User{ID: "user-id-3", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				// teacher
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-1",
					).
					Return(&user_domain.User{ID: "teacher-id-1", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-2",
					).
					Return(&user_domain.User{ID: "teacher-id-2", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-3",
					).
					Return(&user_domain.User{ID: "teacher-id-2", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				expectedLesson := &domain.Lesson{
					LessonID:         "test-id-1",
					LocationID:       "center-id-1",
					StartTime:        now.UTC(),
					EndTime:          now.UTC(),
					SchedulingStatus: domain.LessonSchedulingStatusPublished,
					TeachingMedium:   domain.LessonTeachingMediumOffline,
					TeachingMethod:   domain.LessonTeachingMethodGroup,
					Learners: domain.LessonLearners{
						{
							LearnerID:        "user-id-1",
							CourseID:         "course-id-1",
							AttendStatus:     domain.StudentAttendStatusAttend,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.NoticeEmpty,
							AttendanceReason: domain.ReasonEmpty,
						},
						{
							LearnerID:        "user-id-2",
							CourseID:         "course-id-2",
							AttendStatus:     domain.StudentAttendStatusEmpty,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.NoticeEmpty,
							AttendanceReason: domain.ReasonEmpty,
						},
						{
							LearnerID:        "user-id-3",
							CourseID:         "course-id-3",
							AttendStatus:     domain.StudentAttendStatusAbsent,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.InAdvance,
							AttendanceReason: domain.PhysicalCondition,
							AttendanceNote:   "sample-attendance-note",
						},
					},
					Teachers: domain.LessonTeachers{
						{
							TeacherID: "teacher-id-1",
						},
						{
							TeacherID: "teacher-id-2",
						},
					},
					ClassID:   "mock-class-id-1",
					CourseID:  "mock-course-id-1",
					DateInfos: dateInfos,
					Classrooms: domain.LessonClassrooms{
						{
							ClassroomID: "classroom-id-1",
						},
						{
							ClassroomID: "classroom-id-2",
						},
					},
					PreparationTime: -1,
					BreakTime:       -1,
				}

				mockSchedulerClient.On("CreateScheduler", ctx, mock.Anything).Once().Return(&mpb.CreateSchedulerResponse{
					SchedulerId: "SchedulerID",
				}, nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				lessonRepo.
					On("InsertLesson", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						actualLesson := args[2].(*domain.Lesson)
						actualLesson.LessonID = expectedLesson.LessonID
						actualLesson.MasterDataPort = nil
						actualLesson.UserModulePort = nil
						actualLesson.MediaModulePort = nil
						actualLesson.DateInfoRepo = nil
						actualLesson.ClassroomRepo = nil

						assert.NotEmpty(t, actualLesson.LessonID)
						expectedLesson.SchedulerID = actualLesson.SchedulerID
						expectedLesson.CreatedAt = actualLesson.CreatedAt
						expectedLesson.UpdatedAt = actualLesson.UpdatedAt
						assert.EqualValues(t, expectedLesson, actualLesson)
					}).Return(expectedLesson, nil).Once()
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectLessonCreated, mock.MatchedBy(func(data []byte) bool {
					r := &bpb.EvtLesson{}
					if err := proto.Unmarshal(data, r); err != nil {
						return false
					}
					if len(r.GetCreateLessons().Lessons) > 0 {
						lesson := r.GetCreateLessons().Lessons[0]
						return lesson.LessonId == expectedLesson.LessonID && lesson.Name == expectedLesson.Name && reflect.DeepEqual(lesson.LearnerIds, expectedLesson.GetLearnersIDs())
					}
					return false
				})).Return("", nil).Once()

			},
		},
		{
			name: "insert lesson failed",
			req: &lpb.CreateLessonRequest{
				StartTime:      timestamppb.New(now),
				EndTime:        timestamppb.New(now),
				TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
				TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
				TeacherIds:     []string{"teacher-id-1", "teacher-id-2"},
				ClassroomIds:   []string{"classroom-id-1", "classroom-id-2"},
				LocationId:     "center-id-1",
				StudentInfoList: []*lpb.CreateLessonRequest_StudentInfo{
					{
						StudentId:        "user-id-1",
						CourseId:         "course-id-1",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						AttendanceNotice: lpb.StudentAttendanceNotice_NOTICE_EMPTY,
						AttendanceReason: lpb.StudentAttendanceReason_REASON_EMPTY,
						LocationId:       "center-id-1",
					},
					{
						StudentId:        "user-id-2",
						CourseId:         "course-id-2",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_EMPTY,
						AttendanceNotice: lpb.StudentAttendanceNotice_NOTICE_EMPTY,
						AttendanceReason: lpb.StudentAttendanceReason_REASON_EMPTY,
						LocationId:       "center-id-1",
					},
					{
						StudentId:        "user-id-3",
						CourseId:         "course-id-3",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
						AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE,
						AttendanceReason: lpb.StudentAttendanceReason_PHYSICAL_CONDITION,
						AttendanceNote:   "sample-attendance-note",
						LocationId:       "center-id-1",
					},
				},
				Materials: []*lpb.Material{
					{
						Resource: &lpb.Material_MediaId{
							MediaId: "media-id-1",
						},
					},
					{
						Resource: &lpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
				},
				SavingOption: &lpb.CreateLessonRequest_SavingOption{
					Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
				},
				ClassId:  "mock-class-id-1",
				CourseId: "mock-course-id-1",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, tx, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, tx, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()
				masterDataRepo.
					On("GetCourseByID", ctx, tx, "mock-course-id-1").
					Return(&domain.Course{
						CourseID:  "mock-course-id-1",
						Name:      "mock course name 1",
						UpdatedAt: now,
						CreatedAt: now,
					}, nil).Once()
				masterDataRepo.
					On("GetClassByID", ctx, tx, "mock-class-id-1").
					Return(&domain.Class{
						ClassID:   "mock-class-id-1",
						Name:      "mock course name 1",
						UpdatedAt: now,
						CreatedAt: now,
					}, nil).Once()
				mediaModulePort.
					On(
						"RetrieveMediasByIDs",
						ctx,
						[]string{"media-id-1", "media-id-2"},
					).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).
					Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						now.UTC(),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2", "user-id-3", "course-id-3"},
					).
					Return(nil).
					Once()
				classroomRepo.
					On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				// student
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-1",
					).
					Return(&user_domain.User{ID: "user-id-1", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-2",
					).
					Return(&user_domain.User{ID: "user-id-2", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-3",
					).
					Return(&user_domain.User{ID: "user-id-3", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				// teacher
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-1",
					).
					Return(&user_domain.User{ID: "teacher-id-1", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-2",
					).
					Return(&user_domain.User{ID: "teacher-id-2", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-3",
					).
					Return(&user_domain.User{ID: "teacher-id-2", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				expectedLesson := &domain.Lesson{
					LessonID:         "test-id-1",
					LocationID:       "center-id-1",
					StartTime:        now.UTC(),
					EndTime:          now.UTC(),
					SchedulingStatus: domain.LessonSchedulingStatusPublished,
					TeachingMedium:   domain.LessonTeachingMediumOffline,
					TeachingMethod:   domain.LessonTeachingMethodGroup,
					Learners: domain.LessonLearners{
						{
							LearnerID:        "user-id-1",
							CourseID:         "course-id-1",
							AttendStatus:     domain.StudentAttendStatusAttend,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.NoticeEmpty,
							AttendanceReason: domain.ReasonEmpty,
						},
						{
							LearnerID:        "user-id-2",
							CourseID:         "course-id-2",
							AttendStatus:     domain.StudentAttendStatusEmpty,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.NoticeEmpty,
							AttendanceReason: domain.ReasonEmpty,
						},
						{
							LearnerID:        "user-id-3",
							CourseID:         "course-id-3",
							AttendStatus:     domain.StudentAttendStatusAbsent,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.InAdvance,
							AttendanceReason: domain.PhysicalCondition,
							AttendanceNote:   "sample-attendance-note",
						},
					},
					Teachers: domain.LessonTeachers{
						{
							TeacherID: "teacher-id-1",
						},
						{
							TeacherID: "teacher-id-2",
						},
					},
					Material: &domain.LessonMaterial{
						MediaIDs: []string{"media-id-1", "media-id-2"},
					},
					ClassID:   "mock-class-id-1",
					CourseID:  "mock-course-id-1",
					DateInfos: dateInfos,
					Classrooms: domain.LessonClassrooms{
						{
							ClassroomID: "classroom-id-1",
						},
						{
							ClassroomID: "classroom-id-2",
						},
					},
					PreparationTime: -1,
					BreakTime:       -1,
				}

				mockSchedulerClient.On("CreateScheduler", ctx, mock.Anything).Twice().Return(&mpb.CreateSchedulerResponse{
					SchedulerId: "SchedulerID",
				}, nil)

				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				lessonRepo.
					On("InsertLesson", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						actualLesson := args[2].(*domain.Lesson)
						actualLesson.LessonID = expectedLesson.LessonID
						actualLesson.MasterDataPort = nil
						actualLesson.UserModulePort = nil
						actualLesson.MediaModulePort = nil
						actualLesson.DateInfoRepo = nil
						actualLesson.ClassroomRepo = nil

						assert.NotEmpty(t, actualLesson.LessonID)
						expectedLesson.SchedulerID = actualLesson.SchedulerID
						expectedLesson.CreatedAt = actualLesson.CreatedAt
						expectedLesson.UpdatedAt = actualLesson.UpdatedAt
						assert.EqualValues(t, expectedLesson, actualLesson)
					}).Return(nil, errors.New("could not insert lesson")).Once()

			},
			hasError: true,
		},
		{
			name: "duplicated learner",
			req: &lpb.CreateLessonRequest{
				StartTime:      timestamppb.New(now),
				EndTime:        timestamppb.New(now),
				TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
				TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
				TeacherIds:     []string{"teacher-id-1", "teacher-id-2"},
				ClassroomIds:   []string{"classroom-id-1", "classroom-id-2"},
				LocationId:     "center-id-1",
				StudentInfoList: []*lpb.CreateLessonRequest_StudentInfo{
					{
						StudentId:        "user-id-1",
						CourseId:         "course-id-1",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						AttendanceNotice: lpb.StudentAttendanceNotice_NOTICE_EMPTY,
						AttendanceReason: lpb.StudentAttendanceReason_REASON_EMPTY,
					},
					{
						StudentId:        "user-id-1",
						CourseId:         "course-id-2",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
						AttendanceNotice: lpb.StudentAttendanceNotice_ON_THE_DAY,
						AttendanceReason: lpb.StudentAttendanceReason_SCHOOL_EVENT,
					},
					{
						StudentId:        "user-id-2",
						CourseId:         "course-id-2",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_EMPTY,
						AttendanceNotice: lpb.StudentAttendanceNotice_NOTICE_EMPTY,
						AttendanceReason: lpb.StudentAttendanceReason_REASON_EMPTY,
					},
					{
						StudentId:        "user-id-3",
						CourseId:         "course-id-3",
						AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
						AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE,
						AttendanceReason: lpb.StudentAttendanceReason_PHYSICAL_CONDITION,
						AttendanceNote:   "sample-attendance-note",
					},
				},
				Materials: []*lpb.Material{
					{
						Resource: &lpb.Material_MediaId{
							MediaId: "media-id-1",
						},
					},
					{
						Resource: &lpb.Material_MediaId{
							MediaId: "media-id-2",
						},
					},
				},
				SavingOption: &lpb.CreateLessonRequest_SavingOption{
					Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
				},
				ClassId:  "mock-class-id-1",
				CourseId: "mock-course-id-1",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, tx, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()

				// student
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-1",
					).
					Return(&user_domain.User{ID: "user-id-1", Group: "USER_GROUP_STUDENT"}, nil)
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-2",
					).
					Return(&user_domain.User{ID: "user-id-2", Group: "USER_GROUP_STUDENT"}, nil)
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-3",
					).
					Return(&user_domain.User{ID: "user-id-3", Group: "USER_GROUP_STUDENT"}, nil)
				// teacher
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-1",
					).
					Return(&user_domain.User{ID: "teacher-id-1", Group: "USER_GROUP_TEACHER"}, nil)
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-2",
					).
					Return(&user_domain.User{ID: "teacher-id-2", Group: "USER_GROUP_TEACHER"}, nil)
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-3",
					).
					Return(&user_domain.User{ID: "teacher-id-2", Group: "USER_GROUP_TEACHER"}, nil)
				mockSchedulerClient.On("CreateScheduler", ctx, mock.Anything).Once().Return(&mpb.CreateSchedulerResponse{
					SchedulerId: "SchedulerID",
				}, nil)
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			srv := NewLessonModifierService(
				wrapperConnection,
				jsm,
				lessonRepo,
				masterDataRepo,
				userModuleAdapter,
				mediaModulePort,
				dateInfoRepo,
				classroomRepo,
				nil,
				"local",
				mockUnleashClient,
				schedulerRepo,
				studentSubscriptionRepo,
				reallocationRepo,
				nil,
				zoomService,
				zoomAccountRepo,
				nil,
				nil,
				mockSchedulerClient,
				nil,
			)
			res, err := srv.CreateLesson(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, res.Id)
			}
			mock.AssertExpectationsForObjects(
				t,
				db,
				tx,
				masterDataRepo,
				userModuleAdapter,
				mediaModulePort,
				lessonRepo,
				courseRepo,
				mockUnleashClient,
			)
		})
	}
}

func TestLessonManagementService_DeleteLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonRepo := &mock_repositories.MockLessonRepo{}
	lessonReportRepo := &mock_lesson_report_repositories.MockLessonReportRepo{}
	masterDataRepo := &mock_repositories.MockMasterDataRepo{}
	userModuleAdapter := &mock_user_module_adapter.MockUserModuleAdapter{}
	mediaModulePort := &mock_media_module_adapter.MockMediaModuleAdapter{}
	calendarRepo := new(calendar_mock_repositories.MockSchedulerRepo)
	dateInfoRepo := new(calendar_mock_repositories.MockDateInfoRepo)
	classroomRepo := new(mock_repositories.MockClassroomRepo)
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	jsm := &mock_nats.JetStreamManagement{}
	studentSubscriptionRepo := new(mock_user_repo.MockStudentSubscriptionRepo)
	mockSchedulerClient := &mock_clients.MockSchedulerClient{}

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	zoomAccountRepo := new(mock_zoom_repo.MockZoomAccountRepo)
	mockExternalConfigService := &mock_service.MockExternalConfigService{}
	mockHTTPClient := &mock_clients.MockHTTPClient{}
	zcf := &configs.ZoomConfig{}
	lessonMemberRepo := new(mock_repositories.MockLessonMemberRepo)

	zoomService := zoom_service.InitZoomService(zcf, mockExternalConfigService, mockHTTPClient)
	lessonIDs := []string{"lesson-id-1", "lesson-id-2", "lesson-id-3"}

	tcs := []struct {
		name     string
		context  context.Context
		req      *lpb.DeleteLessonRequest
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name:    "delete successfully without saving_option",
			context: ctx,
			req:     &lpb.DeleteLessonRequest{LessonId: "lesson-id-1"},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonWithSchedulerInfoByLessonID", ctx, db, "lesson-id-1").Return(&domain.Lesson{
					LessonID: "lesson-id-1",
					ZoomID:   "",
					SchedulerInfo: &domain.SchedulerInfo{
						Freq: "once",
					},
				}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				lessonReportRepo.
					On("DeleteReportsBelongToLesson", ctx, tx, []string{"lesson-id-1"}).
					Return(nil).Once()
				lessonRepo.
					On("Delete", ctx, tx, []string{"lesson-id-1"}).
					Return(nil).Once()
				jsm.On("PublishAsyncContext", ctx, constants.SubjectLessonDeleted, mock.Anything).Once().Return("", nil)
			},
			hasError: false,
		},
		{
			name:    "delete successfully with saving_option type one time",
			context: ctx,
			req: &lpb.DeleteLessonRequest{
				LessonId:     "lesson-id-1",
				SavingOption: &lpb.DeleteLessonRequest_SavingOption{Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonWithSchedulerInfoByLessonID", ctx, db, "lesson-id-1").Return(&domain.Lesson{
					LessonID: "lesson-id-1",
					ZoomID:   "",
					SchedulerInfo: &domain.SchedulerInfo{
						Freq: "once",
					},
				}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				lessonReportRepo.
					On("DeleteReportsBelongToLesson", ctx, tx, []string{"lesson-id-1"}).
					Return(nil).Once()
				lessonRepo.
					On("Delete", ctx, tx, []string{"lesson-id-1"}).
					Return(nil).Once()
				jsm.On("PublishAsyncContext", ctx, constants.SubjectLessonDeleted, mock.Anything).Once().Return("", nil)
			},
			hasError: false,
		},
		{
			name:    "delete successfully with saving_option type recurring",
			context: ctx,
			req: &lpb.DeleteLessonRequest{
				LessonId:     "lesson-id-1",
				SavingOption: &lpb.DeleteLessonRequest_SavingOption{Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetFutureRecurringLessonIDs", ctx, db, "lesson-id-1").Once().Return(lessonIDs, nil)
				db.On("Begin", ctx).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				lessonReportRepo.
					On("DeleteReportsBelongToLesson", ctx, tx, lessonIDs).
					Return(nil).Once()
				lessonRepo.
					On("Delete", ctx, tx, lessonIDs).
					Return(nil).Once()
				jsm.On("PublishAsyncContext", ctx, constants.SubjectLessonDeleted, mock.Anything).Once().Return("", nil)
			},
			hasError: false,
		},
		{
			name:    "delete fail with saving_option type recurring",
			context: ctx,
			req: &lpb.DeleteLessonRequest{
				LessonId:     "lesson-id-1",
				SavingOption: &lpb.DeleteLessonRequest_SavingOption{Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetFutureRecurringLessonIDs", ctx, db, "lesson-id-1").Once().Return([]string{}, fmt.Errorf("some errors"))
			},
			hasError: true,
		},
		{
			name:    "delete failed",
			context: ctx,
			req:     &lpb.DeleteLessonRequest{LessonId: "lesson-id-1"},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonWithSchedulerInfoByLessonID", ctx, db, "lesson-id-1").Return(&domain.Lesson{
					LessonID: "lesson-id-1",
					ZoomID:   "",
					SchedulerInfo: &domain.SchedulerInfo{
						Freq: "once",
					},
				}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				lessonReportRepo.
					On("DeleteReportsBelongToLesson", ctx, tx, []string{"lesson-id-1"}).
					Return(nil).Once()
				lessonRepo.
					On("Delete", ctx, tx, []string{"lesson-id-1"}).
					Return(errors.New("error")).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(tc.context)
			srv := NewLessonModifierService(
				wrapperConnection,
				jsm,
				lessonRepo,
				masterDataRepo,
				userModuleAdapter,
				mediaModulePort,
				dateInfoRepo,
				classroomRepo,
				lessonReportRepo,
				"",
				mockUnleashClient,
				calendarRepo,
				studentSubscriptionRepo,
				nil,
				lessonMemberRepo,
				zoomService,
				zoomAccountRepo,
				nil,
				nil,
				mockSchedulerClient,
				nil,
			)
			_, err := srv.DeleteLesson(tc.context, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo, lessonReportRepo, mockUnleashClient)
		})
	}
}

func TestLessonModifierService_UpdateLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonRepo := &mock_repositories.MockLessonRepo{}
	lessonReportRepo := &mock_lesson_report_repositories.MockLessonReportRepo{}
	masterDataRepo := &mock_repositories.MockMasterDataRepo{}
	userModuleAdapter := &mock_user_module_adapter.MockUserModuleAdapter{}
	mediaModulePort := &mock_media_module_adapter.MockMediaModuleAdapter{}
	schedulerRepo := new(calendar_mock_repositories.MockSchedulerRepo)
	dateInfoRepo := new(calendar_mock_repositories.MockDateInfoRepo)
	dateInfos := []*calendar_dto.DateInfo{}
	classroomRepo := new(mock_repositories.MockClassroomRepo)
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	jsm := &mock_nats.JetStreamManagement{}
	studentSubscriptionRepo := new(mock_user_repo.MockStudentSubscriptionRepo)
	userRepo := new(mock_user_repo.MockUserRepo)
	reallocationRepo := new(mock_repositories.MockReallocationRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	now := time.Now()
	zoomAccountRepo := new(mock_zoom_repo.MockZoomAccountRepo)
	mockExternalConfigService := &mock_service.MockExternalConfigService{}
	mockHTTPClient := &mock_clients.MockHTTPClient{}
	zcf := &configs.ZoomConfig{}
	mockSchedulerClient := &mock_clients.MockSchedulerClient{}
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")

	zoomService := zoom_service.InitZoomService(zcf, mockExternalConfigService, mockHTTPClient)
	req := &lpb.UpdateLessonRequest{
		LessonId:       "lesson-id",
		StartTime:      timestamppb.New(now),
		EndTime:        timestamppb.New(now),
		TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
		TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
		TeacherIds:     []string{"teacher-id-1", "teacher-id-2"},
		ClassroomIds:   []string{"classroom-id-1", "classroom-id-2"},
		LocationId:     "center-id-1",
		StudentInfoList: []*lpb.UpdateLessonRequest_StudentInfo{
			{
				StudentId:        "user-id-1",
				CourseId:         "course-id-1",
				AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
				LocationId:       "center-id-1",
				AttendanceNotice: lpb.StudentAttendanceNotice_NOTICE_EMPTY,
				AttendanceReason: lpb.StudentAttendanceReason_REASON_EMPTY,
			},
			{
				StudentId:        "user-id-2",
				CourseId:         "course-id-2",
				AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_EMPTY,
				LocationId:       "center-id-1",
				AttendanceNotice: lpb.StudentAttendanceNotice_NOTICE_EMPTY,
				AttendanceReason: lpb.StudentAttendanceReason_REASON_EMPTY,
			},
			{
				StudentId:        "user-id-3",
				CourseId:         "course-id-3",
				AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
				LocationId:       "center-id-1",
				AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE,
				AttendanceReason: lpb.StudentAttendanceReason_PHYSICAL_CONDITION,
				AttendanceNote:   "sample-attendance-note",
			},
		},
		Materials: []*lpb.Material{
			{
				Resource: &lpb.Material_MediaId{
					MediaId: "media-id-1",
				},
			},
			{
				Resource: &lpb.Material_MediaId{
					MediaId: "media-id-2",
				},
			},
		},
		SavingOption: &lpb.UpdateLessonRequest_SavingOption{
			Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
		},
		ClassId:  "class-id",
		CourseId: "course-id",
	}

	tcs := []struct {
		name     string
		context  context.Context
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name:    "update successfully",
			context: ctx,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Twice()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, tx, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()

				lessonRepo.On("GetLessonByID", mock.Anything, mock.Anything, req.LessonId).Once().Return(&domain.Lesson{
					LessonID:         req.LessonId,
					SchedulingStatus: domain.LessonSchedulingStatusCompleted,
					StartTime:        now,
					EndTime:          now,
				}, nil)

				//for check invalid in domain.lesson

				lessonRepo.On("GetLessonByID", ctx, tx, req.LessonId).Once().Return(&domain.Lesson{
					LessonID:         req.LessonId,
					SchedulingStatus: domain.LessonSchedulingStatusCompleted,
					StartTime:        now,
					EndTime:          now,
					TeachingMethod:   domain.LessonTeachingMethodIndividual,
				}, nil)

				masterDataRepo.
					On("GetLocationByID", ctx, tx, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()
				mediaModulePort.
					On(
						"RetrieveMediasByIDs",
						ctx,
						[]string{"media-id-1", "media-id-2"},
					).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).
					Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						now.UTC(),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2", "user-id-3", "course-id-3"},
					).
					Return(nil).
					Once()
				classroomRepo.
					On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				// student
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-1",
					).
					Return(&user_domain.User{ID: "user-id-1", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-2",
					).
					Return(&user_domain.User{ID: "user-id-2", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-3",
					).
					Return(&user_domain.User{ID: "user-id-3", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				// teacher
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-1",
					).
					Return(&user_domain.User{ID: "teacher-id-1", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-2",
					).
					Return(&user_domain.User{ID: "teacher-id-2", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-3",
					).
					Return(&user_domain.User{ID: "teacher-id-2", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				currentLesson := &domain.Lesson{
					LessonID:    "lesson-id",
					SchedulerID: "scheduler-id",
					LocationID:  "center-id-1",
				}
				schedulerRepo.
					On("GetByID", ctx, currentLesson.SchedulerID).Return(&calendar_entities.Scheduler{
					SchedulerID: currentLesson.SchedulerID,
					Frequency:   calendar_constants.FrequencyOnce,
				}, nil).Once()
				lessonRepo.
					On("UpdateLesson", ctx, tx, mock.MatchedBy(func(l *domain.Lesson) bool {
						assert.Equal(t, req.StartTime.AsTime(), l.StartTime)
						assert.Equal(t, req.EndTime.AsTime(), l.EndTime)
						assert.Equal(t, domain.LessonTeachingMedium(req.TeachingMedium.String()), l.TeachingMedium)
						assert.Equal(t, domain.LessonTeachingMethod(req.TeachingMethod.String()), l.TeachingMethod)
						assert.Equal(t, req.TeacherIds, l.GetTeacherIDs())
						assert.Equal(t, req.LocationId, l.LocationID)
						for _, v := range req.StudentInfoList {
							assert.Equal(t, true, checkInfoListOnUpdateLesson(v, l.Learners))
						}
						for _, v := range req.Materials {
							assert.Equal(t, true, checkMaterials(v, l.Material.MediaIDs))
						}
						assert.Equal(t, req.ClassId, l.ClassID)
						assert.Equal(t, req.CourseId, l.CourseID)
						return true
					})).Return(&domain.Lesson{
					LocationID:       "center-id-1",
					StartTime:        now.UTC(),
					EndTime:          now.UTC(),
					SchedulingStatus: domain.LessonSchedulingStatusPublished,
					TeachingMedium:   domain.LessonTeachingMediumOffline,
					TeachingMethod:   domain.LessonTeachingMethodIndividual,
					Learners: domain.LessonLearners{
						{
							LearnerID:        "user-id-1",
							CourseID:         "course-id-1",
							AttendStatus:     domain.StudentAttendStatusAttend,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.NoticeEmpty,
							AttendanceReason: domain.ReasonEmpty,
						},
						{
							LearnerID:        "user-id-2",
							CourseID:         "course-id-2",
							AttendStatus:     domain.StudentAttendStatusEmpty,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.NoticeEmpty,
							AttendanceReason: domain.ReasonEmpty,
						},
						{
							LearnerID:        "user-id-3",
							CourseID:         "course-id-3",
							AttendStatus:     domain.StudentAttendStatusAbsent,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.InAdvance,
							AttendanceReason: domain.PhysicalCondition,
							AttendanceNote:   "sample-attendance-note",
						},
					},
					Teachers: domain.LessonTeachers{
						{
							TeacherID: "teacher-id-1",
						},
						{
							TeacherID: "teacher-id-2",
						},
					},
					Material: &domain.LessonMaterial{
						MediaIDs: []string{"media-id-1", "media-id-2"},
					},
					ClassID:  "",
					CourseID: "",
					Classrooms: domain.LessonClassrooms{
						{
							ClassroomID: "classroom-id-1",
						},
						{
							ClassroomID: "classroom-id-2",
						},
					},
					PreparationTime: -1,
					BreakTime:       -1,
				}, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				reallocationRepo.
					On("GetByNewLessonID", ctx, tx, []string{}, mock.AnythingOfType("string")).
					Return([]*domain.Reallocation{}, nil).Once()
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectLessonUpdated, mock.Anything).Once().Return("", nil)

			},
			hasError: false,
		},
		{
			name:    "update failed",
			context: ctx,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Twice()
				lessonRepo.On("GetLessonByID", mock.Anything, mock.Anything, req.LessonId).Once().Return(&domain.Lesson{
					LessonID:         req.LessonId,
					SchedulingStatus: domain.LessonSchedulingStatusCompleted,
					StartTime:        now,
					EndTime:          now,
				}, nil)

				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", ctx).Once().Return(nil)
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, tx, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()
				//for check invalid in domain.lesson

				lessonRepo.On("GetLessonByID", mock.Anything, mock.Anything, req.LessonId).Once().Return(&domain.Lesson{
					LessonID:         req.LessonId,
					SchedulingStatus: domain.LessonSchedulingStatusCompleted,
					StartTime:        now,
					EndTime:          now,
					TeachingMethod:   domain.LessonTeachingMethodIndividual,
				}, nil)

				masterDataRepo.
					On("GetLocationByID", ctx, tx, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()
				mediaModulePort.
					On(
						"RetrieveMediasByIDs",
						ctx,
						[]string{"media-id-1", "media-id-2"},
					).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).
					Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						now.UTC(),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2", "user-id-3", "course-id-3"},
					).
					Return(nil).
					Once()
				classroomRepo.
					On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				// student
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-1",
					).
					Return(&user_domain.User{ID: "user-id-1", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-2",
					).
					Return(&user_domain.User{ID: "user-id-2", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-3",
					).
					Return(&user_domain.User{ID: "user-id-3", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				// teacher
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-1",
					).
					Return(&user_domain.User{ID: "teacher-id-1", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-2",
					).
					Return(&user_domain.User{ID: "teacher-id-2", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-3",
					).
					Return(&user_domain.User{ID: "teacher-id-2", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				currentLesson := &domain.Lesson{
					LessonID:    "lesson-id",
					SchedulerID: "scheduler-id",
					LocationID:  "center-id-1",
				}
				schedulerRepo.
					On("GetByID", ctx, currentLesson.SchedulerID).Return(&calendar_entities.Scheduler{
					SchedulerID: currentLesson.SchedulerID,
					Frequency:   calendar_constants.FrequencyOnce,
				}, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Once()
				lessonRepo.
					On("UpdateLesson", ctx, tx, mock.MatchedBy(func(l *domain.Lesson) bool {
						assert.Equal(t, req.StartTime.AsTime(), l.StartTime)
						assert.Equal(t, req.EndTime.AsTime(), l.EndTime)
						assert.Equal(t, domain.LessonTeachingMedium(req.TeachingMedium.String()), l.TeachingMedium)
						assert.Equal(t, domain.LessonTeachingMethod(req.TeachingMethod.String()), l.TeachingMethod)
						assert.Equal(t, req.TeacherIds, l.GetTeacherIDs())
						assert.Equal(t, req.LocationId, l.LocationID)
						for _, v := range req.StudentInfoList {
							assert.Equal(t, true, checkInfoListOnUpdateLesson(v, l.Learners))
						}
						for _, v := range req.Materials {
							assert.Equal(t, true, checkMaterials(v, l.Material.MediaIDs))
						}
						assert.Equal(t, req.ClassId, l.ClassID)
						assert.Equal(t, req.CourseId, l.CourseID)
						return true
					})).Return(nil, pgx.ErrNoRows).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(tc.context)
			srv := NewLessonModifierService(
				wrapperConnection,
				jsm,
				lessonRepo,
				masterDataRepo,
				userModuleAdapter,
				mediaModulePort,
				dateInfoRepo,
				classroomRepo,
				lessonReportRepo,
				"",
				mockUnleashClient,
				schedulerRepo,
				studentSubscriptionRepo,
				reallocationRepo,
				nil,
				zoomService,
				zoomAccountRepo,
				nil,
				nil,
				mockSchedulerClient,
				nil,
			)
			_, err := srv.UpdateLesson(tc.context, req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo, masterDataRepo, mediaModulePort, userModuleAdapter, dateInfoRepo, mockUnleashClient)
		})
	}
}

func checkInfoListOnUpdateLesson(v *lpb.UpdateLessonRequest_StudentInfo, lc domain.LessonLearners) bool {
	for _, b := range lc {
		if v.CourseId == b.CourseID && v.LocationId == b.LocationID && domain.StudentAttendStatus(v.AttendanceStatus.String()) == b.AttendStatus && v.StudentId == b.LearnerID {
			return true
		}
	}
	return false
}

func checkMaterials(v *lpb.Material, lc []string) bool {
	for _, b := range lc {
		switch lr := v.Resource.(type) {
		case *lpb.Material_BrightcoveVideo_:
			return false
		case *lpb.Material_MediaId:
			if b == lr.MediaId {
				return true
			}
		}
	}
	return false
}

func TestLessonModifierService_UpdateToRecurrence(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonRepo := &mock_repositories.MockLessonRepo{}
	lessonReportRepo := &mock_lesson_report_repositories.MockLessonReportRepo{}
	masterDataRepo := &mock_repositories.MockMasterDataRepo{}
	userModuleAdapter := &mock_user_module_adapter.MockUserModuleAdapter{}
	mediaModulePort := &mock_media_module_adapter.MockMediaModuleAdapter{}
	schedulerRepo := new(calendar_mock_repositories.MockSchedulerRepo)
	dateInfoRepo := new(calendar_mock_repositories.MockDateInfoRepo)
	dateInfos := []*calendar_dto.DateInfo{}
	classroomRepo := new(mock_repositories.MockClassroomRepo)
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	jsm := &mock_nats.JetStreamManagement{}
	studentSubscriptionRepo := new(mock_user_repo.MockStudentSubscriptionRepo)
	// userRepo := new(mock_user_repo.MockUserRepo)
	reallocationRepo := new(mock_repositories.MockReallocationRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	now := time.Now()
	zoomAccountRepo := new(mock_zoom_repo.MockZoomAccountRepo)
	mockExternalConfigService := &mock_service.MockExternalConfigService{}
	mockHTTPClient := &mock_clients.MockHTTPClient{}
	zcf := &configs.ZoomConfig{}
	mockSchedulerClient := &mock_clients.MockSchedulerClient{}
	startTime := time.Date(2023, 4, 13, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 4, 13, 10, 0, 0, 0, time.UTC)
	zoomService := zoom_service.InitZoomService(zcf, mockExternalConfigService, mockHTTPClient)
	req := &lpb.UpdateToRecurrenceRequest{
		LessonId:       "01GM10XAQ7GGXTT6KFCAEXZFH5",
		StartTime:      timestamppb.New(startTime),
		EndTime:        timestamppb.New(endTime),
		TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
		TeachingMethod: cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
		TeacherIds:     []string{"teacher-id-1", "teacher-id-2"},
		ClassroomIds:   []string{"classroom-id-1", "classroom-id-2"},
		LocationId:     "center-id-1",
		StudentInfoList: []*lpb.UpdateToRecurrenceRequest_StudentInfo{
			{
				StudentId:        "student-id-1",
				CourseId:         "course-id-1",
				AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
				LocationId:       "center-id-1",
				AttendanceNotice: lpb.StudentAttendanceNotice_NOTICE_EMPTY,
				AttendanceReason: lpb.StudentAttendanceReason_REASON_EMPTY,
			},
			{
				StudentId:        "student-id-2",
				CourseId:         "course-id-2",
				AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_EMPTY,
				LocationId:       "center-id-1",
				AttendanceNotice: lpb.StudentAttendanceNotice_NOTICE_EMPTY,
				AttendanceReason: lpb.StudentAttendanceReason_REASON_EMPTY,
			},
			{
				StudentId:        "student-id-3",
				CourseId:         "course-id-3",
				AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
				LocationId:       "center-id-1",
				AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE,
				AttendanceReason: lpb.StudentAttendanceReason_PHYSICAL_CONDITION,
				AttendanceNote:   "sample-attendance-note",
			},
		},
		Materials: []*lpb.Material{
			{
				Resource: &lpb.Material_MediaId{
					MediaId: "media-id-1",
				},
			},
			{
				Resource: &lpb.Material_MediaId{
					MediaId: "media-id-2",
				},
			},
		},
		SavingOption: &lpb.UpdateToRecurrenceRequest_SavingOption{
			Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE,
			Recurrence: &lpb.Recurrence{
				EndDate: timestamppb.New(time.Date(2023, 4, 28, 9, 0, 0, 0, time.UTC)),
			},
		},
		ClassId:  "class-id",
		CourseId: "course-id",
	}

	tcs := []struct {
		name     string
		context  context.Context
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name:    "update successfully",
			context: ctx,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Twice()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.On("GetLessonByID", mock.Anything, mock.Anything, req.LessonId).Once().Return(&domain.Lesson{
					LessonID:         req.LessonId,
					SchedulingStatus: domain.LessonSchedulingStatusPublished,
					StartTime:        now,
					EndTime:          now,
					Teachers: domain.LessonTeachers{
						{
							TeacherID: "teacher-1",
						},
						{
							TeacherID: "teacher-2",
						},
					},
					LocationID: "location-1",
				}, nil)

				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, db, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()

				mockSchedulerClient.On("CreateScheduler", ctx, mock.Anything).Once().Return(&mpb.CreateSchedulerResponse{
					SchedulerId: "scheduler-1",
				}, nil)
				studentSubscriptionRepo.On("GetStudentCourseSubscriptions", ctx,
					mock.Anything,
					mock.Anything,
					[]string{
						"student-id-1",
						"course-id-1",
						"student-id-2",
						"course-id-2",
						"student-id-3",
						"course-id-3",
					}).
					Return(user_domain.StudentSubscriptions{
						{
							SubscriptionID: "subscription-id-1",
							StudentID:      "student-id-1",
							CourseID:       "course-id-1",
							LocationIDs:    []string{"location-id-1", "location-id-3"},
							StartAt:        startTime.Add(-24 * time.Hour),
							EndAt:          startTime.AddDate(0, 1, 0),
						},
						{
							SubscriptionID: "subscription-id-2",
							StudentID:      "student-id-2",
							CourseID:       "course-id-2",
							LocationIDs:    []string{"location-id-1", "location-id-3"},
							StartAt:        startTime.Add(-24 * time.Hour),
							EndAt:          startTime.AddDate(0, 1, 0),
						},
						{
							SubscriptionID: "subscription-id-3",
							StudentID:      "student-id-3",
							CourseID:       "course-id-3",
							LocationIDs:    []string{"location-id-1", "location-id-2", "location-id-5"},
							StartAt:        startTime.Add(-24 * time.Hour),
							EndAt:          startTime.AddDate(0, 1, 0),
						},
					}, nil).Once()

				lessonRepo.On("GetLessonByID", ctx, tx, req.LessonId).Once().Return(&domain.Lesson{
					LessonID:         req.LessonId,
					SchedulingStatus: domain.LessonSchedulingStatusPublished,
					StartTime:        now,
					EndTime:          now,
					TeachingMethod:   domain.LessonTeachingMethodIndividual,
				}, nil)

				masterDataRepo.
					On("GetLocationByID", ctx, tx, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()
				mediaModulePort.
					On(
						"RetrieveMediasByIDs",
						ctx,
						[]string{"media-id-1", "media-id-2"},
					).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).
					Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						startTime.UTC(),
						[]string{"student-id-1", "course-id-1", "student-id-2", "course-id-2", "student-id-3", "course-id-3"},
					).
					Return(nil).
					Once()
				classroomRepo.
					On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Twice()
				lessonRepo.On("UpsertLessons", ctx, tx, mock.MatchedBy(func(recurringLesson *domain.RecurringLesson) bool {
					lessons := recurringLesson.Lessons
					if len(lessons) != 3 {
						return false
					}
					return true
				})).Return([]string{}, nil).Once()
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectLessonUpdated, mock.Anything).Once().Return("", nil)
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectLessonCreated, mock.Anything).Once().Return("", nil)
			},
			hasError: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(tc.context)
			srv := NewLessonModifierService(
				wrapperConnection,
				jsm,
				lessonRepo,
				masterDataRepo,
				userModuleAdapter,
				mediaModulePort,
				dateInfoRepo,
				classroomRepo,
				lessonReportRepo,
				"",
				mockUnleashClient,
				schedulerRepo,
				studentSubscriptionRepo,
				reallocationRepo,
				nil,
				zoomService,
				zoomAccountRepo,
				nil,
				nil,
				mockSchedulerClient,
				nil,
			)
			_, err := srv.UpdateToRecurrence(tc.context, req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo, masterDataRepo, mediaModulePort, userModuleAdapter, dateInfoRepo, mockUnleashClient)
		})
	}
}
