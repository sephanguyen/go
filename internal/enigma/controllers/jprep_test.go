package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/enigma/dto"
	"github.com/manabie-com/backend/internal/enigma/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/enigma/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
	natsJS "github.com/nats-io/nats.go"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestJPREPController_topbClasses(t *testing.T) {
	t.Parallel()
	zapLog, _ := zap.NewDevelopment()
	t.Run("happy case", func(t *testing.T) {
		t.Parallel()
		classes := []dto.Class{
			{
				ActionKind: dto.ActionKindUpserted,
				CourseID:   3,
				ClassID:    3,
				ClassName:  "class-3",
				StartDate:  "2021/01/01",
				EndDate:    "2021/01/31",
			},
		}

		registerClasses, err := toPbClasses(classes, zapLog)
		assert.NoError(t, err)
		assert.Equal(t, "2021-01-31 14:59:59 +0000 UTC", registerClasses[0].EndDate.AsTime().String())
		assert.Equal(t, "2020-12-31 15:00:00 +0000 UTC", registerClasses[0].StartDate.AsTime().String())
	})
	t.Run("error no course id", func(t *testing.T) {
		t.Parallel()
		classes := []dto.Class{
			{
				ActionKind: dto.ActionKindUpserted,
				ClassID:    3,
				ClassName:  "class-3",
				StartDate:  "2021/01/01",
				EndDate:    "2021/01/31",
			},
		}

		_, err := toPbClasses(classes, zapLog)
		assert.Error(t, err)
		assert.Equal(t, "payload.m_regular_course[0].m_course_name_id is required", err.Error())
	})
}

func TestJPREPController_topbStudents(t *testing.T) {
	t.Parallel()
	zapLog, _ := zap.NewDevelopment()
	t.Run("happy case", func(t *testing.T) {
		t.Parallel()
		students := []dto.Student{
			{
				ActionKind: dto.ActionKindUpserted,
				StudentID:  "id",
				LastName:   "last name",
				GivenName:  "given name",
				Regularcourses: []struct {
					ClassID   int    "json:\"m_course_id\""
					Startdate string "json:\"startdate\""
					Enddate   string "json:\"enddate\""
				}{

					{ClassID: 1, Startdate: "2021/01/01", Enddate: "2021/01/31"}},
			},
		}

		registerStudents, err := toPbStudents(students, zapLog)
		assert.NoError(t, err)
		assert.Equal(t, "2021-01-31 14:59:59 +0000 UTC", registerStudents[0].Packages[0].EndDate.AsTime().String()) //EndDate.AsTime().String())
		assert.Equal(t, "2020-12-31 15:00:00 +0000 UTC", registerStudents[0].Packages[0].StartDate.AsTime().String())
	})
	t.Run("error missing last name", func(t *testing.T) {
		t.Parallel()
		students := []dto.Student{
			{
				ActionKind: dto.ActionKindUpserted,
				StudentID:  "id",
				GivenName:  "given name",
				Regularcourses: []struct {
					ClassID   int    "json:\"m_course_id\""
					Startdate string "json:\"startdate\""
					Enddate   string "json:\"enddate\""
				}{

					{ClassID: 1, Startdate: "2021/01/01", Enddate: "2021/01/31"}},
			},
		}

		_, err := toPbStudents(students, zapLog)
		assert.Error(t, err)
		assert.Equal(t, "payload.m_student[0].last_name is required", err.Error())
	})
}

func TestJPREPController_logSyncData(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	partnerSyncDataLogRepo := new(mock_repositories.MockPartnerSyncDataLogRepo)
	partnerSyncDataLogSplitRepo := new(mock_repositories.MockPartnerSyncDataLogSplitRepo)
	jsm := new(mock_nats.JetStreamManagement)
	j := &JPREPController{
		Logger:                      zap.NewNop(),
		DB:                          db,
		JSM:                         jsm,
		PartnerSyncDataLogRepo:      partnerSyncDataLogRepo,
		PartnerSyncDataLogSplitRepo: partnerSyncDataLogSplitRepo,
	}
	userID := ksuid.New().String()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	payload, _ := json.Marshal([]string{"student-1", "student-2"})
	testCases := []struct {
		name         string
		userID       string
		payload      []byte
		signature    string
		expectedResp interface{}
		expectedErr  error
		setup        func(ctx context.Context)
	}{
		{
			name:        "success write log",
			userID:      userID,
			payload:     payload,
			signature:   "signature-1",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				partnerSyncDataLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "error write log",
			userID:      userID,
			payload:     payload,
			signature:   "signature-1",
			expectedErr: fmt.Errorf("err PartnerSyncDataLogRepo.Create: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				partnerSyncDataLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			_, err := j.logSyncData(ctx, testCase.payload, testCase.signature)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJPREPController_logSyncDataSplit(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	partnerSyncDataLogRepo := new(mock_repositories.MockPartnerSyncDataLogRepo)
	partnerSyncDataLogSplitRepo := new(mock_repositories.MockPartnerSyncDataLogSplitRepo)
	jsm := new(mock_nats.JetStreamManagement)
	j := &JPREPController{
		Logger:                      zap.NewNop(),
		DB:                          db,
		JSM:                         jsm,
		PartnerSyncDataLogRepo:      partnerSyncDataLogRepo,
		PartnerSyncDataLogSplitRepo: partnerSyncDataLogSplitRepo,
	}
	userID := ksuid.New().String()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	payload, _ := json.Marshal([]string{"student-1", "student-2"})
	testCases := []struct {
		name         string
		userID       string
		logStruct    *LogStructure
		signature    string
		expectedResp interface{}
		expectedErr  error
		setup        func(ctx context.Context)
	}{
		{
			name:   "success write log",
			userID: userID,
			logStruct: &LogStructure{
				PartnerSyncDataLogID: "log-id",
				Payload:              payload,
				Kind:                 string(entities.KindLesson),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				partnerSyncDataLogSplitRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:   "error write log",
			userID: userID,
			logStruct: &LogStructure{
				PartnerSyncDataLogID: "log-id",
				Payload:              payload,
				Kind:                 string(entities.KindLesson),
			},
			expectedErr: fmt.Errorf("err PartnerSyncDataLogSplitRepo.Create: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				partnerSyncDataLogSplitRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			log, err := j.logSyncDataSplit(ctx, testCase.logStruct)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, log.Kind.String, testCase.logStruct.Kind)
				assert.Equal(t, log.PartnerSyncDataLogID.String, testCase.logStruct.PartnerSyncDataLogID)
			}
		})
	}
}

func TestJPREPController_getPartnerLogReport(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	partnerSyncDataLogRepo := new(mock_repositories.MockPartnerSyncDataLogRepo)
	partnerSyncDataLogSplitRepo := new(mock_repositories.MockPartnerSyncDataLogSplitRepo)
	jsm := new(mock_nats.JetStreamManagement)
	j := &JPREPController{
		Logger:                      zap.NewNop(),
		DB:                          db,
		JSM:                         jsm,
		PartnerSyncDataLogRepo:      partnerSyncDataLogRepo,
		PartnerSyncDataLogSplitRepo: partnerSyncDataLogSplitRepo,
	}
	userID := ksuid.New().String()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	testCases := []struct {
		name         string
		userID       string
		request      *dto.PartnerLogRequestByDate
		expectedResp interface{}
		expectedErr  error
		setup        func(ctx context.Context)
	}{
		{
			name:   "err get report when from_date null",
			userID: userID,
			request: &dto.PartnerLogRequestByDate{
				Timestamp: int(time.Now().Unix()),
				Payload: struct {
					FromDate string `json:"from_date"`
					ToDate   string `json:"to_date"`
				}{
					FromDate: "",
					ToDate:   "2022-01-03",
				},
			},
			expectedErr: fmt.Errorf("from_date and to_date is required"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:   "err get report when from_date > to_date",
			userID: userID,
			request: &dto.PartnerLogRequestByDate{
				Timestamp: int(time.Now().Unix()),
				Payload: struct {
					FromDate string `json:"from_date"`
					ToDate   string `json:"to_date"`
				}{
					FromDate: "2022-01-21",
					ToDate:   "2022-01-20",
				},
			},
			expectedErr: fmt.Errorf("Start date must come before End date"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:   "err get report when period more than 30 days",
			userID: userID,
			request: &dto.PartnerLogRequestByDate{
				Timestamp: int(time.Now().Unix()),
				Payload: struct {
					FromDate string `json:"from_date"`
					ToDate   string `json:"to_date"`
				}{
					FromDate: "2022-01-21",
					ToDate:   "2022-03-20",
				},
			},
			expectedErr: fmt.Errorf("Please choose a period less than or equal to 30 days"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:   "success get report ",
			userID: userID,
			request: &dto.PartnerLogRequestByDate{
				Timestamp: int(time.Now().Unix()),
				Payload: struct {
					FromDate string `json:"from_date"`
					ToDate   string `json:"to_date"`
				}{
					FromDate: "2022-01-01",
					ToDate:   "2022-01-03",
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				partnerSyncDataLogSplitRepo.On("GetLogsReportByDate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.PartnerSyncDataLogReport{}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)

			_, err := j.getPartnerLogReport(ctx, testCase.request)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJPREPController_recoverData(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	partnerSyncDataLogRepo := new(mock_repositories.MockPartnerSyncDataLogRepo)
	partnerSyncDataLogSplitRepo := new(mock_repositories.MockPartnerSyncDataLogSplitRepo)
	jsm := new(mock_nats.JetStreamManagement)
	j := &JPREPController{
		Logger:                      zap.NewNop(),
		DB:                          db,
		JSM:                         jsm,
		PartnerSyncDataLogRepo:      partnerSyncDataLogRepo,
		PartnerSyncDataLogSplitRepo: partnerSyncDataLogSplitRepo,
	}
	userID := ksuid.New().String()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	testCases := []struct {
		name         string
		userID       string
		request      *dto.PartnerLogRequestByDate
		expectedResp interface{}
		expectedErr  error
		setup        func(ctx context.Context)
	}{
		{
			name:   "err recover when from_date null",
			userID: userID,
			request: &dto.PartnerLogRequestByDate{
				Timestamp: int(time.Now().Unix()),
				Payload: struct {
					FromDate string `json:"from_date"`
					ToDate   string `json:"to_date"`
				}{
					FromDate: "",
					ToDate:   "2022-01-03",
				},
			},
			expectedErr: fmt.Errorf("from_date and to_date is required"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:   "err recover when from_date > to_date",
			userID: userID,
			request: &dto.PartnerLogRequestByDate{
				Timestamp: int(time.Now().Unix()),
				Payload: struct {
					FromDate string `json:"from_date"`
					ToDate   string `json:"to_date"`
				}{
					FromDate: "2022-01-21",
					ToDate:   "2022-01-20",
				},
			},
			expectedErr: fmt.Errorf("Start date must come before End date"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:   "err recover when period more than 30 days",
			userID: userID,
			request: &dto.PartnerLogRequestByDate{
				Timestamp: int(time.Now().Unix()),
				Payload: struct {
					FromDate string `json:"from_date"`
					ToDate   string `json:"to_date"`
				}{
					FromDate: "2022-01-21",
					ToDate:   "2022-03-20",
				},
			},
			expectedErr: fmt.Errorf("Please choose a period less than or equal to 30 days"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:   "success recover students",
			userID: userID,
			request: &dto.PartnerLogRequestByDate{
				Timestamp: int(time.Now().Unix()),
				Payload: struct {
					FromDate string `json:"from_date"`
					ToDate   string `json:"to_date"`
				}{
					FromDate: "2022-01-01",
					ToDate:   "2022-01-03",
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				partnerSyncDataLogSplitRepo.On("GetLogsByDateRange", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.PartnerSyncDataLogSplit{
					{
						PartnerSyncDataLogSplitID: database.Text("1"),
						Status:                    database.Text(string(entities.StatusProcessing)),
						RetryTimes:                pgtype.Int4{Int: 2},
						Kind:                      database.Text(string(entities.KindStudent)),
						Payload: database.JSONB(`
						[{"last_name": "Last name 01G1D6ZRTP1VSGMZRA8PVSHX7M", "given_name": "Given name 01G1D6ZRTP1VSGMZRA8R037V83", "student_id": "1", "action_kind": 1}]`),
						CreatedAt: pgtype.Timestamptz{Time: time.Date(2021, 12, 12, 0, 0, 0, 0, time.UTC)},
					},
				}, nil)
				partnerSyncDataLogSplitRepo.On("UpdateLogsStatusAndRetryTime", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				partnerSyncDataLogRepo.On("UpdateTime", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUserRegistrationNatsJS, mock.Anything).Once().Return(&natsJS.PubAck{}, nil)
			},
		},
		{
			name:   "success recover staff",
			userID: userID,
			request: &dto.PartnerLogRequestByDate{
				Timestamp: int(time.Now().Unix()),
				Payload: struct {
					FromDate string `json:"from_date"`
					ToDate   string `json:"to_date"`
				}{
					FromDate: "2022-01-01",
					ToDate:   "2022-01-03",
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				partnerSyncDataLogSplitRepo.On("GetLogsByDateRange", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.PartnerSyncDataLogSplit{
					{
						PartnerSyncDataLogSplitID: database.Text("1"),
						Status:                    database.Text(string(entities.StatusProcessing)),
						RetryTimes:                pgtype.Int4{Int: 2},
						Kind:                      database.Text(string(entities.KindStaff)),
						Payload: database.JSONB(`
						[{"last_name": "Last name 01G1D6ZRTP1VSGMZRA8PVSHX7M", "given_name": "Given name 01G1D6ZRTP1VSGMZRA8R037V83", "student_id": "1", "action_kind": 1}]`),
						CreatedAt: pgtype.Timestamptz{Time: time.Date(2021, 12, 12, 0, 0, 0, 0, time.UTC)},
					},
				}, nil)
				partnerSyncDataLogSplitRepo.On("UpdateLogsStatusAndRetryTime", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				partnerSyncDataLogRepo.On("UpdateTime", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUserRegistrationNatsJS, mock.Anything).Once().Return(&natsJS.PubAck{}, nil)
			},
		},
		{
			name:   "success recover course",
			userID: userID,
			request: &dto.PartnerLogRequestByDate{
				Timestamp: int(time.Now().Unix()),
				Payload: struct {
					FromDate string `json:"from_date"`
					ToDate   string `json:"to_date"`
				}{
					FromDate: "2022-01-01",
					ToDate:   "2022-01-03",
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				partnerSyncDataLogSplitRepo.On("GetLogsByDateRange", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.PartnerSyncDataLogSplit{
					{
						PartnerSyncDataLogSplitID: database.Text("1"),
						Status:                    database.Text(string(entities.StatusProcessing)),
						RetryTimes:                pgtype.Int4{Int: 2},
						Kind:                      database.Text(string(entities.KindCourse)),
						Payload: database.JSONB(`
						[{"last_name": "Last name 01G1D6ZRTP1VSGMZRA8PVSHX7M", "given_name": "Given name 01G1D6ZRTP1VSGMZRA8R037V83", "student_id": "1", "action_kind": 1}]`),
						CreatedAt: pgtype.Timestamptz{Time: time.Date(2021, 12, 12, 0, 0, 0, 0, time.UTC)},
					},
				}, nil)
				partnerSyncDataLogSplitRepo.On("UpdateLogsStatusAndRetryTime", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				partnerSyncDataLogRepo.On("UpdateTime", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectSyncMasterRegistration, mock.Anything).Once().Return(&natsJS.PubAck{}, nil)
			},
		},
		{
			name:   "success recover class",
			userID: userID,
			request: &dto.PartnerLogRequestByDate{
				Timestamp: int(time.Now().Unix()),
				Payload: struct {
					FromDate string `json:"from_date"`
					ToDate   string `json:"to_date"`
				}{
					FromDate: "2022-01-01",
					ToDate:   "2022-01-03",
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				partnerSyncDataLogSplitRepo.On("GetLogsByDateRange", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.PartnerSyncDataLogSplit{
					{
						PartnerSyncDataLogSplitID: database.Text("1"),
						Status:                    database.Text(string(entities.StatusProcessing)),
						RetryTimes:                pgtype.Int4{Int: 2},
						Kind:                      database.Text(string(entities.KindClass)),
						Payload: database.JSONB(`
						[{"last_name": "Last name 01G1D6ZRTP1VSGMZRA8PVSHX7M", "given_name": "Given name 01G1D6ZRTP1VSGMZRA8R037V83", "student_id": "1", "action_kind": 1}]`),
						CreatedAt: pgtype.Timestamptz{Time: time.Date(2021, 12, 12, 0, 0, 0, 0, time.UTC)},
					},
				}, nil)
				partnerSyncDataLogSplitRepo.On("UpdateLogsStatusAndRetryTime", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				partnerSyncDataLogRepo.On("UpdateTime", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectSyncMasterRegistration, mock.Anything).Once().Return(&natsJS.PubAck{}, nil)
			},
		},
		{
			name:   "success recover lesson",
			userID: userID,
			request: &dto.PartnerLogRequestByDate{
				Timestamp: int(time.Now().Unix()),
				Payload: struct {
					FromDate string `json:"from_date"`
					ToDate   string `json:"to_date"`
				}{
					FromDate: "2022-01-01",
					ToDate:   "2022-01-03",
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				partnerSyncDataLogSplitRepo.On("GetLogsByDateRange", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.PartnerSyncDataLogSplit{
					{
						PartnerSyncDataLogSplitID: database.Text("1"),
						Status:                    database.Text(string(entities.StatusProcessing)),
						RetryTimes:                pgtype.Int4{Int: 2},
						Kind:                      database.Text(string(entities.KindLesson)),
						Payload: database.JSONB(`
						[{"last_name": "Last name 01G1D6ZRTP1VSGMZRA8PVSHX7M", "given_name": "Given name 01G1D6ZRTP1VSGMZRA8R037V83", "student_id": "1", "action_kind": 1}]`),
						CreatedAt: pgtype.Timestamptz{Time: time.Date(2021, 12, 12, 0, 0, 0, 0, time.UTC)},
					},
				}, nil)
				partnerSyncDataLogSplitRepo.On("UpdateLogsStatusAndRetryTime", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				partnerSyncDataLogRepo.On("UpdateTime", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectSyncMasterRegistration, mock.Anything).Once().Return(&natsJS.PubAck{}, nil)
			},
		},
		{
			name:   "success recover AcademicYear",
			userID: userID,
			request: &dto.PartnerLogRequestByDate{
				Timestamp: int(time.Now().Unix()),
				Payload: struct {
					FromDate string `json:"from_date"`
					ToDate   string `json:"to_date"`
				}{
					FromDate: "2022-01-01",
					ToDate:   "2022-01-03",
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				partnerSyncDataLogSplitRepo.On("GetLogsByDateRange", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.PartnerSyncDataLogSplit{
					{
						PartnerSyncDataLogSplitID: database.Text("1"),
						Status:                    database.Text(string(entities.StatusProcessing)),
						RetryTimes:                pgtype.Int4{Int: 2},
						Kind:                      database.Text(string(entities.KindAcademicYear)),
						Payload: database.JSONB(`
						[{"last_name": "Last name 01G1D6ZRTP1VSGMZRA8PVSHX7M", "given_name": "Given name 01G1D6ZRTP1VSGMZRA8R037V83", "student_id": "1", "action_kind": 1}]`),
						CreatedAt: pgtype.Timestamptz{Time: time.Date(2021, 12, 12, 0, 0, 0, 0, time.UTC)},
					},
				}, nil)
				partnerSyncDataLogSplitRepo.On("UpdateLogsStatusAndRetryTime", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				partnerSyncDataLogRepo.On("UpdateTime", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectSyncMasterRegistration, mock.Anything).Once().Return(&natsJS.PubAck{}, nil)
			},
		},
		{
			name:   "success recover student lessons",
			userID: userID,
			request: &dto.PartnerLogRequestByDate{
				Timestamp: int(time.Now().Unix()),
				Payload: struct {
					FromDate string `json:"from_date"`
					ToDate   string `json:"to_date"`
				}{
					FromDate: "2022-01-01",
					ToDate:   "2022-01-03",
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				partnerSyncDataLogSplitRepo.On("GetLogsByDateRange", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.PartnerSyncDataLogSplit{
					{
						PartnerSyncDataLogSplitID: database.Text("1"),
						Status:                    database.Text(string(entities.StatusProcessing)),
						RetryTimes:                pgtype.Int4{Int: 2},
						Kind:                      database.Text(string(entities.KindStudentLessons)),
						Payload: database.JSONB(`
						[{"last_name": "Last name 01G1D6ZRTP1VSGMZRA8PVSHX7M", "given_name": "Given name 01G1D6ZRTP1VSGMZRA8R037V83", "student_id": "1", "action_kind": 1}]`),
						CreatedAt: pgtype.Timestamptz{Time: time.Date(2021, 12, 12, 0, 0, 0, 0, time.UTC)},
					},
				}, nil)
				partnerSyncDataLogSplitRepo.On("UpdateLogsStatusAndRetryTime", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				partnerSyncDataLogRepo.On("UpdateTime", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectJPREPSyncUserCourseNatsJS, mock.Anything).Once().Return(&natsJS.PubAck{}, nil)
			},
		},
		{
			name:   "success recover student lessons and AcademicYear",
			userID: userID,
			request: &dto.PartnerLogRequestByDate{
				Timestamp: int(time.Now().Unix()),
				Payload: struct {
					FromDate string `json:"from_date"`
					ToDate   string `json:"to_date"`
				}{
					FromDate: "2022-01-01",
					ToDate:   "2022-01-03",
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				partnerSyncDataLogSplitRepo.On("GetLogsByDateRange", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.PartnerSyncDataLogSplit{
					{
						PartnerSyncDataLogSplitID: database.Text("1"),
						Status:                    database.Text(string(entities.StatusProcessing)),
						RetryTimes:                pgtype.Int4{Int: 2},
						Kind:                      database.Text(string(entities.KindStudentLessons)),
						Payload: database.JSONB(`
						[{"last_name": "Last name 01G1D6ZRTP1VSGMZRA8PVSHX7M", "given_name": "Given name 01G1D6ZRTP1VSGMZRA8R037V83", "student_id": "1", "action_kind": 1}]`),
						CreatedAt: pgtype.Timestamptz{Time: time.Date(2021, 12, 12, 0, 0, 0, 0, time.UTC)},
					},
					{
						PartnerSyncDataLogSplitID: database.Text("1"),
						Status:                    database.Text(string(entities.StatusProcessing)),
						RetryTimes:                pgtype.Int4{Int: 2},
						Kind:                      database.Text(string(entities.KindAcademicYear)),
						Payload: database.JSONB(`
						[{"last_name": "Last name 01G1D6ZRTP1VSGMZRA8PVSHX7M", "given_name": "Given name 01G1D6ZRTP1VSGMZRA8R037V83", "student_id": "1", "action_kind": 1}]`),
						CreatedAt: pgtype.Timestamptz{Time: time.Date(2021, 12, 12, 0, 0, 0, 0, time.UTC)},
					},
				}, nil)
				partnerSyncDataLogSplitRepo.On("UpdateLogsStatusAndRetryTime", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				partnerSyncDataLogRepo.On("UpdateTime", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectJPREPSyncUserCourseNatsJS, mock.Anything).Once().Return(&natsJS.PubAck{}, nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectSyncMasterRegistration, mock.Anything).Once().Return(&natsJS.PubAck{}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)

			err := j.recoverData(ctx, testCase.request)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJPREPController_toPbLessons(t *testing.T) {
	t.Parallel()
	zapLog, _ := zap.NewDevelopment()
	now := time.Now()
	t.Run("happy case", func(t *testing.T) {
		t.Parallel()
		lessons := []dto.Lesson{
			{
				ActionKind:    dto.ActionKindUpserted,
				LessonID:      1,
				LessonType:    "online",
				CourseID:      1,
				StartDatetime: int(now.Unix()),
				EndDatetime:   int(now.Unix()),
				ClassName:     "class name",
				Week:          "1",
			},
		}

		registerLessons, err := toPbLessons(lessons, zapLog)
		assert.NoError(t, err)
		assert.Equal(t, cpb.LessonType_LESSON_TYPE_ONLINE, registerLessons[0].LessonType)
		assert.Equal(t, fmt.Sprintf("JPREP_LESSON_%09d", 1), registerLessons[0].LessonId)
		assert.Equal(t, fmt.Sprintf("JPREP_COURSE_%09d", 1), registerLessons[0].CourseId)
	})
	t.Run("error no course id", func(t *testing.T) {
		t.Parallel()
		lessons := []dto.Lesson{
			{
				ActionKind:    dto.ActionKindUpserted,
				LessonID:      1,
				LessonType:    "online",
				StartDatetime: int(now.Unix()),
				EndDatetime:   int(now.Unix()),
				ClassName:     "class name",
				Week:          "1",
			},
		}

		_, err := toPbLessons(lessons, zapLog)
		assert.Error(t, err)
		assert.Equal(t, "payload.m_lesson[0].m_course_name_id is required", err.Error())
	})
}

func TestJPREPController_toPbStudentLessons(t *testing.T) {
	t.Parallel()
	zapLog, _ := zap.NewDevelopment()
	t.Run("happy case", func(t *testing.T) {
		t.Parallel()
		studentLessons := []dto.StudentLesson{
			{
				StudentID:  "1",
				LessonIDs:  []int{1},
				ActionKind: dto.ActionKindUpserted,
			},
		}

		registers, err := toPbStudentLessons(studentLessons, zapLog)
		assert.NoError(t, err)
		assert.Equal(t, "1", registers[0].StudentId)
	})
	t.Run("error no student id", func(t *testing.T) {
		t.Parallel()
		studentLessons := []dto.StudentLesson{
			{
				LessonIDs:  []int{1},
				ActionKind: dto.ActionKindUpserted,
			},
		}

		_, err := toPbStudentLessons(studentLessons, zapLog)
		assert.Error(t, err)
		assert.Equal(t, "payload.student_id is required", err.Error())
	})
}

func TestJPREPController_toPbCourses(t *testing.T) {
	t.Parallel()
	zapLog, _ := zap.NewDevelopment()
	t.Run("happy case", func(t *testing.T) {
		t.Parallel()
		courses := []dto.Course{
			{
				ActionKind:         dto.ActionKindUpserted,
				CourseID:           1,
				CourseName:         "course-juku",
				CourseStudentDivID: dto.CourseIDJuku,
			},
			{
				ActionKind:         dto.ActionKindUpserted,
				CourseID:           2,
				CourseName:         "course-kids",
				CourseStudentDivID: dto.CourseIDKid,
			},
			{
				ActionKind:         dto.ActionKindUpserted,
				CourseID:           3,
				CourseName:         "course-aplus",
				CourseStudentDivID: dto.CourseIDAPlus,
			},
		}

		registers, err := toPbCourses(courses, zapLog)
		assert.NoError(t, err)
		assert.Len(t, registers, 3)
		assert.Equal(t, fmt.Sprintf("JPREP_COURSE_%09d", 1), registers[0].CourseId)
		assert.Equal(t, "course-juku", registers[0].CourseName)
		assert.Equal(t, cpb.CourseStatus_COURSE_STATUS_INACTIVE, registers[0].Status)

		assert.Equal(t, fmt.Sprintf("JPREP_COURSE_%09d", 2), registers[1].CourseId)
		assert.Equal(t, "course-kids", registers[1].CourseName)
		assert.Equal(t, cpb.CourseStatus_COURSE_STATUS_ACTIVE, registers[1].Status)

		assert.Equal(t, fmt.Sprintf("JPREP_COURSE_%09d", 3), registers[2].CourseId)
		assert.Equal(t, "course-aplus", registers[2].CourseName)
		assert.Equal(t, cpb.CourseStatus_COURSE_STATUS_ACTIVE, registers[2].Status)
	})

	t.Run("error no student id", func(t *testing.T) {
		t.Parallel()
		courses := []dto.Course{
			{
				ActionKind: dto.ActionKindUpserted,
				CourseName: "course-name-with-actionKind-upsert",
			},
		}

		_, err := toPbCourses(courses, zapLog)
		assert.Error(t, err)
		assert.Equal(t, "payload.m_course_name[0].m_course_name_id is required", err.Error())
	})
}
