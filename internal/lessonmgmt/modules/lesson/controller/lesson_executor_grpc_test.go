package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	calendar_dto "github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/commands"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/producers"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/queries"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	calendar_mock_repositories "github.com/manabie-com/backend/mock/calendar/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	mock_clients "github.com/manabie-com/backend/mock/lessonmgmt/zoom/clients"
	mpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLessonExecutorService_GenerateLessonCSVTemplate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := new(mock_repositories.MockLessonRepo)
	lessonStr := `"partner_internal_id","start_date_time","end_date_time","teaching_method"` + "\n" +
		`"sample_center_id","` + time.Now().Format("2006-01-02 15:04:05") + `","` +
		time.Now().Add(2*time.Hour).Format("2006-01-02 15:04:05") + `","1"` + "\n"
	lessonStrV2 := `"partner_internal_id","start_date_time","end_date_time","teaching_method","teaching_medium","teacher_ids","student_course_ids"` + "\n" +
		`"sample_center_id","` + time.Now().Format("2006-01-02 15:04:05") + `","` +
		time.Now().Add(2*time.Hour).Format("2006-01-02 15:04:05") + `","1"` + `","1"` +
		`","teacherID1_teacherID2_teacherID3"` + `","studentID1/courseID1_studentID2/courseID2"` + "\n"

	s := &LessonExecutorService{
		lessonQueryHandler: queries.LessonQueryHandler{
			WrapperConnection: wrapperConnection,
			LessonRepo:        lessonRepo,
			UnleashClientIns:  mockUnleashClient,
		},
	}

	t.Run("successful", func(t *testing.T) {
		byteData := []byte(lessonStr)
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
		lessonRepo.On("GenerateLessonTemplate", ctx, db).Once().Return(byteData, nil)
		resp, err := s.GenerateLessonCSVTemplate(ctx, &lpb.GenerateLessonCSVTemplateRequest{})

		assert.Nil(t, err)
		assert.Equal(t, resp.Data, byteData)
	})

	t.Run("Export template V2", func(t *testing.T) {
		byteData := []byte(lessonStrV2)
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Once()
		lessonRepo.On("GenerateLessonTemplateV2", ctx, db).Once().Return(byteData, nil)
		resp, err := s.GenerateLessonCSVTemplate(ctx, &lpb.GenerateLessonCSVTemplateRequest{})

		assert.Nil(t, err)
		assert.Equal(t, resp.Data, byteData)
	})

	t.Run("return internal error when generate template failed", func(t *testing.T) {
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
		lessonRepo.On("GenerateLessonTemplate", ctx, db).Once().Return(nil, errors.New("sample error"))
		resp, err := s.GenerateLessonCSVTemplate(ctx, &lpb.GenerateLessonCSVTemplateRequest{})

		assert.Nil(t, resp.Data)
		assert.Equal(t, err, status.Error(codes.Internal, "sample error"))
	})
}

func TestLessonExecutorService_ExportClassrooms(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	classroomRepo := new(mock_repositories.MockClassroomRepo)
	s := &LessonExecutorService{
		classroomQueryHandler: queries.ClassroomQueryHandler{
			WrapperConnection: wrapperConnection,
			ClassroomRepo:     classroomRepo,
			Env:               "local",
			UnleashClientIns:  mockUnleashClient,
		},
	}

	classroomStrV1 := `"location_id","location_name","classroom_id","classroom_name","remarks","is_archived"` + "\n" +
		`"location-1","Location name 1","classroom-id-1","classroom name 1","","0"` + "\n" +
		`"location-2","Location name 2","classroom-id-2","classroom name 2","","0"` + "\n" +
		`"location-3","Location name 3","classroom-id-3","classroom name 3","","0"` + "\n"
	exportCols := []exporter.ExportColumnMap{
		{
			DBColumn: "location_id",
		},
		{
			DBColumn: "location_name",
		},
		{
			DBColumn: "classroom_id",
		},
		{
			DBColumn: "classroom_name",
		},
		{
			DBColumn: "remarks",
		},
	}

	listColsV1 := append(exportCols, exporter.ExportColumnMap{
		DBColumn: "is_archived",
	})

	listColsV2 := append(exportCols, []exporter.ExportColumnMap{
		{
			DBColumn: "room_area",
		},
		{
			DBColumn: "seat_capacity",
		}}...)

	t.Run("export classrom ver 1 successful", func(t *testing.T) {
		byteData := []byte(classroomStrV1)
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
		classroomRepo.On("ExportAllClassrooms", ctx, db, listColsV1).Once().Return(byteData, nil)
		resp, err := s.ExportClassrooms(ctx, &lpb.ExportClassroomsRequest{})

		assert.Nil(t, err)
		assert.Equal(t, resp.Data, byteData)
	})

	t.Run("export classrom ver 2 successful", func(t *testing.T) {
		byteData := []byte(`"location_id","location_name","classroom_id","classroom_name","remarks","room_area","seat_capacity"` + "\n" +
			`"location-1","Location name 1","classroom-id-1","classroom name 1","","floor 1","6"` + "\n" +
			`"location-2","Location name 2","classroom-id-2","classroom name 2","","floor 2","10"` + "\n" +
			`"location-3","Location name 3","classroom-id-3","classroom name 3","","floor 3","20"` + "\n")
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Once()
		classroomRepo.On("ExportAllClassrooms", ctx, db, listColsV2).Once().Return(byteData, nil)
		resp, err := s.ExportClassrooms(ctx, &lpb.ExportClassroomsRequest{})

		assert.Nil(t, err)
		assert.Equal(t, resp.Data, byteData)
	})

	t.Run("return internal error when export data failed", func(t *testing.T) {
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
		classroomRepo.On("ExportAllClassrooms", ctx, db, listColsV1).Once().Return(nil, errors.New("sample error"))
		resp, err := s.ExportClassrooms(ctx, &lpb.ExportClassroomsRequest{})

		assert.Nil(t, resp.Data)
		assert.Equal(t, err, status.Error(codes.Internal, "sample error"))
	})
}

func TestLessonExecutorService_ImportLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	now := time.Now()
	db := new(mock_database.Ext)
	tx := &mock_database.Tx{}
	jsm := &mock_nats.JetStreamManagement{}
	masterDataRepo := new(mock_repositories.MockMasterDataRepo)
	lessonRepo := new(mock_repositories.MockLessonRepo)
	dateInfoRepo := new(calendar_mock_repositories.MockDateInfoRepo)
	schedulerRepo := new(calendar_mock_repositories.MockSchedulerRepo)
	dateInfos := []*calendar_dto.DateInfo{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	expectedlocationVN, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	mockSchedulerClient := &mock_clients.MockSchedulerClient{}
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")

	s := &LessonExecutorService{
		lessonCommandHandler: commands.LessonCommandHandler{
			WrapperConnection: wrapperConnection,
			SchedulerRepo:     schedulerRepo,
			LessonRepo:        lessonRepo,
			UnleashClientIns:  mockUnleashClient,
			Env:               "local",
			MasterDataPort:    masterDataRepo,
			DateInfoRepo:      dateInfoRepo,
			SchedulerClient:   mockSchedulerClient,
		},
		LessonProducer: producers.LessonProducer{
			JSM: jsm,
		},
		UnleashClientIns: mockUnleashClient,
		Env:              "local",
	}

	testCases := []TestCase{
		{
			name: "parsing valid file (without validation error lines in response)",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &lpb.ImportLessonRequest{
				Payload: []byte(fmt.Sprintf(`partner_internal_id,start_date_time,end_date_time,teaching_method
				pid_1,2023-01-02 05:40:00,2023-01-02 06:45:00,1
				pid_2,2023-01-05 08:30:00,2023-01-05 14:45:00,2`)),
				TimeZone: "Asia/Ho_Chi_Minh",
			},
			expectedResp: &lpb.ImportLessonResponse{
				Errors: []*lpb.ImportLessonResponse_ImportLessonError{},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Twice()
				tx.On("Commit", ctx).Return(nil).Twice()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(3)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Twice()
				masterDataRepo.On("GetLowestLocationsByPartnerInternalIDs", ctx, mock.Anything, []string{"pid_1", "pid_2"}).Once().Return(map[string]*domain.Location{
					"pid_1": {LocationID: "center-1", Name: "Center 1"},
					"pid_2": {LocationID: "center-2", Name: "Center 2"},
				}, nil)

				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, tx, mock.Anything, mock.Anything, mock.Anything,
				).Return([]*calendar_dto.DateInfo{}, nil).Twice()

				mockSchedulerClient.On("CreateScheduler", ctx, mock.Anything).Twice().Return(&mpb.CreateSchedulerResponse{
					SchedulerId: "SchedulerID",
				}, nil)

				masterDataRepo.On("GetLocationByID", ctx, tx, "center-1").
					Return(&domain.Location{
						LocationID: "center-1",
						Name:       "Center 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()

				masterDataRepo.On("GetLocationByID", ctx, tx, "center-2").
					Return(&domain.Location{
						LocationID: "center-2",
						Name:       "Center 2",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()

				expectedLessons := []*domain.Lesson{
					{
						LessonID:         "test-id-1",
						LocationID:       "center-1",
						StartTime:        time.Date(2023, 01, 02, 05, 40, 0, 0, expectedlocationVN),
						EndTime:          time.Date(2023, 01, 02, 06, 45, 0, 0, expectedlocationVN),
						SchedulingStatus: domain.LessonSchedulingStatusDraft,
						TeachingMedium:   domain.LessonTeachingMediumOffline,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						DateInfos:        dateInfos,
						MasterDataPort:   masterDataRepo,
						DateInfoRepo:     dateInfoRepo,
					},
					{
						LessonID:         "test-id-2",
						LocationID:       "center-2",
						StartTime:        time.Date(2023, 01, 05, 8, 30, 0, 0, expectedlocationVN),
						EndTime:          time.Date(2023, 01, 05, 14, 45, 0, 0, expectedlocationVN),
						SchedulingStatus: domain.LessonSchedulingStatusDraft,
						TeachingMedium:   domain.LessonTeachingMediumOffline,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						DateInfos:        dateInfos,
						MasterDataPort:   masterDataRepo,
						DateInfoRepo:     dateInfoRepo,
					},
				}
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Twice()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				lessonRepo.On("InsertLesson", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						actualLesson := args[2].(*domain.Lesson)
						actualLesson.LessonID = expectedLessons[0].LessonID
						actualLesson.MasterDataPort = masterDataRepo
						actualLesson.UserModulePort = nil
						actualLesson.DateInfoRepo = dateInfoRepo
						actualLesson.ClassroomRepo = nil

						assert.NotEmpty(t, actualLesson.LessonID)
						expectedLessons[0].SchedulerID = actualLesson.SchedulerID
						expectedLessons[0].CreatedAt = actualLesson.CreatedAt
						expectedLessons[0].UpdatedAt = actualLesson.UpdatedAt
						assert.EqualValues(t, expectedLessons[0], actualLesson)
					}).Return(expectedLessons[0], nil).Once()

				lessonRepo.On("InsertLesson", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						actualLesson := args[2].(*domain.Lesson)
						actualLesson.LessonID = expectedLessons[1].LessonID
						actualLesson.MasterDataPort = masterDataRepo
						actualLesson.UserModulePort = nil
						actualLesson.DateInfoRepo = dateInfoRepo
						actualLesson.ClassroomRepo = nil

						assert.NotEmpty(t, actualLesson.LessonID)
						expectedLessons[1].SchedulerID = actualLesson.SchedulerID
						expectedLessons[1].CreatedAt = actualLesson.CreatedAt
						expectedLessons[1].UpdatedAt = actualLesson.UpdatedAt
						assert.EqualValues(t, expectedLessons[1], actualLesson)
					}).Return(expectedLessons[1], nil).Once()

				jsm.On("PublishAsyncContext", ctx, constants.SubjectLessonCreated, mock.Anything).Twice().Return("", nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.ImportLesson(testCase.ctx, testCase.req.(*lpb.ImportLessonRequest))
			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.NotNil(t, resp)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.expectedResp.(*lpb.ImportLessonResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, err.RowNumber, expectedResp.Errors[i].RowNumber)
					assert.Contains(t, err.Error, expectedResp.Errors[i].Error)
				}
			}
			mock.AssertExpectationsForObjects(t, db, tx, masterDataRepo, dateInfoRepo, schedulerRepo, lessonRepo, mockUnleashClient)
		})
	}
}

func TestLessonExecutorService_ImportClassroom(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	tx := &mock_database.Tx{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	masterDataRepo := new(mock_repositories.MockMasterDataRepo)
	classroomRepo := new(mock_repositories.MockClassroomRepo)

	s := &LessonExecutorService{
		classroomCommandHandler: commands.ClassroomCommandHandler{
			WrapperConnection: wrapperConnection,
			ClassroomRepo:     classroomRepo,
			MasterDataPort:    masterDataRepo,
		},
	}
	testCases := []TestCase{
		{
			name: "happy case - import classroom success",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &lpb.ImportClassroomRequest{
				Payload: []byte(fmt.Sprintf(`location_id,location_name,classroom_id,classroom_name,room_area,seat_capacity,remarks
					location-1,location-name-1,,classroom name 1,floor 1,24,
					location-2,location-name-2,classroom-id-1,teacher room 1,floor 1,30,teacher seat`)),
			},
			expectedResp: &lpb.ImportClassroomResponse{
				Errors: []*lpb.ImportError{},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Twice()
				tx.On("Commit", ctx).Return(nil).Twice()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				locationName := map[string]string{
					"location-1": "location-name-1",
					"location-2": "location-name-2",
				}
				masterDataRepo.On("CheckLocationByIDs", ctx, mock.Anything, []string{"location-1", "location-2"}, locationName).Once().Return(nil)
				classroomRepo.On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1"}).Return(nil).Once()
				classroomRepo.On("UpsertClassrooms", ctx, mock.Anything, mock.Anything).Return(nil).Once()
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.ImportClassroom(testCase.ctx, testCase.req.(*lpb.ImportClassroomRequest))
			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.NotNil(t, resp)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.expectedResp.(*lpb.ImportClassroomResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, err.RowNumber, expectedResp.Errors[i].RowNumber)
					assert.Contains(t, err.Error, expectedResp.Errors[i].Error)
				}
			}
			mock.AssertExpectationsForObjects(t, masterDataRepo, classroomRepo)
		})
	}
}

func TestLessonExecutorService_ExportCourseTeachingTime(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	courseRepo := new(mock_repositories.MockCourseRepo)
	s := &LessonExecutorService{
		CourseTeachingTimeQueryHandler: queries.CourseTeachingTimeQueryHandler{
			WrapperConnection: wrapperConnection,
			CourseRepo:        courseRepo,
		},
	}

	courseStr := `"course_id","course_name","preparation_time","break_time"` + "\n" +
		`"course-id-1","Course name 1","120","10"` + "\n" +
		`"course-id-2","Course name 2","150","15"` + "\n" +
		`"course-id-3","Course name 3","180","20"` + "\n"
	exportCols := []exporter.ExportColumnMap{
		{
			DBColumn: "course_id",
		},
		{
			DBColumn:  "name",
			CSVColumn: "course_name",
		},
		{
			DBColumn: "preparation_time",
		},
		{
			DBColumn: "break_time",
		},
	}

	t.Run("successful - export course with teaching time info", func(t *testing.T) {
		byteData := []byte(courseStr)
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		courseRepo.On("ExportAllCoursesWithTeachingTimeValue", ctx, db, exportCols).Once().Return(byteData, nil)
		resp, err := s.ExportCourseTeachingTime(ctx, &lpb.ExportCourseTeachingTimeRequest{})

		assert.Nil(t, err)
		assert.Equal(t, resp.Data, byteData)
	})

	t.Run("return internal error when export data failed", func(t *testing.T) {
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		courseRepo.On("ExportAllCoursesWithTeachingTimeValue", ctx, db, exportCols).Once().Return(nil, errors.New("sample error"))
		resp, err := s.ExportCourseTeachingTime(ctx, &lpb.ExportCourseTeachingTimeRequest{})

		assert.Nil(t, resp.Data)
		assert.Equal(t, err, status.Error(codes.Internal, "sample error"))
	})
}

func TestLessonExecutorService_ImportCourseTeachingTime(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	tx := &mock_database.Tx{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	courseRepo := new(mock_repositories.MockCourseRepo)
	lessonRepo := new(mock_repositories.MockLessonRepo)
	masterDataRepo := new(mock_repositories.MockMasterDataRepo)
	timezone := "Asia/Saigon"

	s := &LessonExecutorService{
		CourseTeachingTimeCommandHandler: commands.CourseTeachingTimeCommandHandler{
			WrapperConnection: wrapperConnection,
			CourseRepo:        courseRepo,
		},
		lessonCommandHandler: commands.LessonCommandHandler{
			WrapperConnection: wrapperConnection,
			LessonRepo:        lessonRepo,
			MasterDataPort:    masterDataRepo,
		},
	}
	testCases := []TestCase{
		{
			name: "happy case - import course success",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &lpb.ImportCourseTeachingTimeRequest{
				Payload: []byte(fmt.Sprintf(`course_id,course_name,preparation_time,break_time,action
					course-id-1,course-name-1,120,15,upsert
					course-id-2,course-name-2,150,10,upsert`)),
				Timezone: timezone,
			},
			expectedResp: &lpb.ImportCourseTeachingTimeResponse{
				Errors: []*lpb.ImportError{},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Twice()
				tx.On("Commit", ctx).Return(nil).Twice()
				courses := domain.Courses{
					{
						CourseID:        "course-id-1",
						PreparationTime: 120,
						BreakTime:       15,
					},
					{
						CourseID:        "course-id-2",
						PreparationTime: 150,
						BreakTime:       10,
					},
				}
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(3)
				courseRepo.On("CheckCourseIDs", ctx, mock.Anything, []string{"course-id-1", "course-id-2"}).Return(nil).Once()
				courseRepo.On("RegisterCourseTeachingTime", ctx, mock.Anything, courses).Return(nil).Once()
				lessonRepo.On("GetFutureLessonsByCourseIDs", ctx, mock.Anything, []string{"course-id-1", "course-id-2"}, timezone).Return([]*domain.Lesson{
					{
						LessonID: "lesson-id-1",
						CourseID: "course-id-1",
					},
					{
						LessonID: "lesson-id-2",
						CourseID: "course-id-2",
					},
				}, nil).Once()
				masterDataRepo.
					On("GetCourseTeachingTimeByIDs", ctx, tx, []string{"course-id-1", "course-id-2"}).
					Return(map[string]*domain.Course{
						"course-id-1": {
							CourseID:        "course-id-1",
							PreparationTime: 120,
							BreakTime:       15,
						},
						"course-id-2": {
							CourseID:        "course-id-2",
							PreparationTime: 150,
							BreakTime:       10,
						},
					}, nil).Once()
				lessonRepo.On("UpdateLessonsTeachingTime", ctx, mock.Anything, mock.Anything).Return(nil).Once()
			},
		},
		{
			name: "error case - invalid action",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &lpb.ImportCourseTeachingTimeRequest{
				Payload: []byte(fmt.Sprintf(`course_id,course_name,preparation_time,break_time,action
					course-id-1,course-name-1,120,15,update
					course-id-2,course-name-2,150,10,upsert`)),
				Timezone: timezone,
			},
			expectedResp: &lpb.ImportCourseTeachingTimeResponse{
				Errors: []*lpb.ImportError{
					{
						RowNumber: 2,
						Error:     "invalid action",
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Twice()
				tx.On("Commit", ctx).Return(nil).Twice()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Twice()
				courseRepo.On("CheckCourseIDs", ctx, mock.Anything, []string{"course-id-1", "course-id-2"}).Return(nil).Once()
			},
		},
		{
			name: "error case - invalid time values",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &lpb.ImportCourseTeachingTimeRequest{
				Payload: []byte(fmt.Sprintf(`course_id,course_name,preparation_time,break_time,action
					course-id-1,course-name-1,12.0,15,upsert
					course-id-2,course-name-2,150,10,upsert`)),
				Timezone: timezone,
			},
			expectedResp: &lpb.ImportCourseTeachingTimeResponse{
				Errors: []*lpb.ImportError{
					{
						RowNumber: 2,
						Error:     "preparation time must be numeric and should be greater than or equal to 0",
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Twice()
				tx.On("Commit", ctx).Return(nil).Twice()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Twice()
				courseRepo.On("CheckCourseIDs", ctx, mock.Anything, []string{"course-id-1", "course-id-2"}).Return(nil).Once()
			},
		},
		{
			name: "error case - course id not exist",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &lpb.ImportCourseTeachingTimeRequest{
				Payload: []byte(fmt.Sprintf(`course_id,course_name,preparation_time,break_time,action
					course-id-1,course-name-1,120,15,upsert
					course-id-2,course-name-2,150,10,upsert`)),
				Timezone: timezone,
			},
			expectedResp: &lpb.ImportCourseTeachingTimeResponse{
				Errors: []*lpb.ImportError{
					{
						RowNumber: 0,
						Error:     "course ids not exists",
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Twice()
				tx.On("Commit", ctx).Return(nil).Twice()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Twice()
				courseRepo.On("CheckCourseIDs", ctx, mock.Anything, []string{"course-id-1", "course-id-2"}).Return(
					errors.New("course ids not exists"),
				).Once()
			},
		},
		{
			name: "error case - invalid file header",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &lpb.ImportCourseTeachingTimeRequest{
				Payload: []byte(fmt.Sprintf(`course_id,name,preparation_time,break_time,action
					course-id-1,course-name-1,120,15,upsert
					course-id-2,course-name-2,150,10,upsert`)),
				Timezone: timezone,
			},
			expectedResp: &lpb.ImportCourseTeachingTimeResponse{
				Errors: []*lpb.ImportError{
					{
						RowNumber: 1,
						Error:     "invalid format: the column have index 1 (toLowerCase) should be 'course_name'",
					},
				},
			},
			setup: func(ctx context.Context) {},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.ImportCourseTeachingTime(testCase.ctx, testCase.req.(*lpb.ImportCourseTeachingTimeRequest))
			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.NotNil(t, resp)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.expectedResp.(*lpb.ImportCourseTeachingTimeResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, err.RowNumber, expectedResp.Errors[i].RowNumber)
					assert.Contains(t, err.Error, expectedResp.Errors[i].Error)
				}
			}
			mock.AssertExpectationsForObjects(t, courseRepo)
		})
	}
}
