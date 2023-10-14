package commands

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	classDomain "github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_class_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/class/infrastructure/repo"
	mock_schedule_class_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/schedule_class/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func setupMock() (*ReserveClassCommandHandler, *mock_database.Ext, *mock_database.Tx, *mock_schedule_class_repo.MockReserveClassRepo, *mock_class_repo.MockClassMemberRepo, *mock_nats.JetStreamManagement) {
	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockReserveClassRepo := new(mock_schedule_class_repo.MockReserveClassRepo)
	mockJsm := new(mock_nats.JetStreamManagement)
	mockClassMemberRepo := new(mock_class_repo.MockClassMemberRepo)

	rcc := &ReserveClassCommandHandler{
		DB:               mockDB,
		ReserveClassRepo: mockReserveClassRepo,
		JSM:              mockJsm,
		ClassMemberRepo:  mockClassMemberRepo,
	}

	return rcc, mockDB, tx, mockReserveClassRepo, mockClassMemberRepo, mockJsm
}

func TestReserveClassCommandHandler_UpsertReserveClass(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	now := time.Now()
	rcc, mockDB, tx, mockReserveClassRepo, _, _ := setupMock()
	reserveClass := &domain.ReserveClass{
		ReserveClassID:   "reserve_class_id",
		StudentID:        "student_id",
		StudentPackageID: "student_pacakge_id",
		CourseID:         "course_id",
		ClassID:          "class_id",
		EffectiveDate:    now,
	}

	testCases := []struct {
		name        string
		req         *domain.ReserveClass
		expectedErr error
		setup       func(ctx context.Context)
	}{
		{
			name:        "success",
			req:         reserveClass,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				mockReserveClassRepo.On("DeleteOldReserveClass", ctx, tx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(pgtype.Text{String: "old_class_id"}, pgtype.Date{Time: now.Add(24 * time.Hour)}, nil)
				mockReserveClassRepo.On("InsertOne", ctx, tx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "insert fail",
			req:         reserveClass,
			expectedErr: fmt.Errorf("UpsertReserveClass: insert fail"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				tx.On("Rollback", ctx).Return(nil).Once()
				mockReserveClassRepo.On("DeleteOldReserveClass", ctx, tx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(pgtype.Text{String: "old_class_id"}, pgtype.Date{Time: now.Add(24 * time.Hour)}, nil)
				mockReserveClassRepo.On("InsertOne", ctx, tx, mock.Anything).Once().Return(fmt.Errorf("insert fail"))
			},
		},
		{
			name:        "delete fail",
			req:         reserveClass,
			expectedErr: fmt.Errorf("UpsertReserveClass: delete fail"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				tx.On("Rollback", ctx).Return(nil).Once()
				mockReserveClassRepo.On("DeleteOldReserveClass", ctx, tx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(pgtype.Text{}, pgtype.Date{}, fmt.Errorf("delete fail"))
				mockReserveClassRepo.On("InsertOne", ctx, tx, mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			_, _, err := rcc.UpsertReserveClass(ctx, tc.req)
			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReserveClassCommandHandler_CheckWillReserveClass(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	now := time.Now()
	rcc, mockDB, _, _, mockClassMemberRepo, _ := setupMock()

	testCases := []struct {
		name                   string
		req                    *mpb.ScheduleStudentClassRequest
		expectedResp           bool
		expectedCurrentClassID string
		expectedErr            error
		setup                  func(ctx context.Context)
	}{
		{
			name: "return err when effective date on the past",
			req: &mpb.ScheduleStudentClassRequest{
				StudentId:        "student_id_01",
				StudentPackageId: "student_package_id_01",
				CourseId:         "course_id_01",
				ClassId:          "class_id_01",
				StartTime:        timestamppb.Now(),
				EndTime:          timestamppb.New(now.Add(100 * 24 * time.Hour)),
				EffectiveDate:    timestamppb.New(now.Add(-10 * 24 * time.Hour)),
			},
			expectedResp:           false,
			expectedCurrentClassID: "",
			expectedErr:            fmt.Errorf("invalid effective date"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "return err when effective date is greater than course end date",
			req: &mpb.ScheduleStudentClassRequest{
				StudentId:        "student_id_01",
				StudentPackageId: "student_package_id_01",
				CourseId:         "course_id_01",
				ClassId:          "class_id_01",
				StartTime:        timestamppb.Now(),
				EndTime:          timestamppb.New(now.Add(19 * 24 * time.Hour)),
				EffectiveDate:    timestamppb.New(now.Add(20 * 24 * time.Hour)),
			},
			expectedResp:           false,
			expectedCurrentClassID: "",
			expectedErr:            fmt.Errorf("invalid effective date"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "register class when course has not started",
			req: &mpb.ScheduleStudentClassRequest{
				StudentId:        "student_id_01",
				StudentPackageId: "student_package_id_01",
				CourseId:         "course_id_01",
				ClassId:          "class_id_01",
				StartTime:        timestamppb.New(now.Add(10 * 24 * time.Hour)),
				EndTime:          timestamppb.New(now.Add(100 * 24 * time.Hour)),
				EffectiveDate:    timestamppb.New(now.Add(20 * 24 * time.Hour)),
			},
			expectedResp:           false,
			expectedCurrentClassID: "",
			expectedErr:            nil,
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "register class when effective date is today",
			req: &mpb.ScheduleStudentClassRequest{
				StudentId:        "student_id_01",
				StudentPackageId: "student_package_id_01",
				CourseId:         "course_id_01",
				ClassId:          "class_id_01",
				StartTime:        timestamppb.New(now.Add(10 * 24 * time.Hour)),
				EndTime:          timestamppb.New(now.Add(100 * 24 * time.Hour)),
				EffectiveDate:    timestamppb.Now(),
			},
			expectedResp:           false,
			expectedCurrentClassID: "",
			expectedErr:            nil,
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "reserve class when student course has active class",
			req: &mpb.ScheduleStudentClassRequest{
				StudentId:        "student_id_01",
				StudentPackageId: "student_package_id_01",
				CourseId:         "course_id_01",
				ClassId:          "class_id_01",
				StartTime:        timestamppb.New(now.Add(-10 * 24 * time.Hour)),
				EndTime:          timestamppb.New(now.Add(100 * 24 * time.Hour)),
				EffectiveDate:    timestamppb.New(now.Add(20 * 24 * time.Hour)),
			},
			expectedResp:           true,
			expectedCurrentClassID: "current_class_id",
			expectedErr:            nil,
			setup: func(ctx context.Context) {
				mockClassMemberRepo.On("GetByUserAndCourse", mock.Anything, mockDB, "student_id_01", "course_id_01").Once().Return(map[string]*classDomain.ClassMember{
					"student_id_01": {
						ClassMemberID: idutil.ULIDNow(),
						ClassID:       "current_class_id",
						UserID:        "student_id_01",
						StartDate:     now.Add(-10 * 24 * time.Hour),
						EndDate:       now.Add(10 * 24 * time.Hour),
					},
				}, nil)
			},
		},
		{
			name: "register class when student course has outdated class",
			req: &mpb.ScheduleStudentClassRequest{
				StudentId:        "student_id_01",
				StudentPackageId: "student_package_id_01",
				CourseId:         "course_id_01",
				ClassId:          "class_id_01",
				StartTime:        timestamppb.New(now.Add(-10 * 24 * time.Hour)),
				EndTime:          timestamppb.New(now.Add(100 * 24 * time.Hour)),
				EffectiveDate:    timestamppb.New(now.Add(20 * 24 * time.Hour)),
			},
			expectedResp:           false,
			expectedCurrentClassID: "",
			expectedErr:            nil,
			setup: func(ctx context.Context) {
				mockClassMemberRepo.On("GetByUserAndCourse", mock.Anything, mockDB, "student_id_01", "course_id_01").Once().Return(map[string]*classDomain.ClassMember{
					"student_id_01": {
						ClassMemberID: idutil.ULIDNow(),
						ClassID:       "outdated_class_id",
						UserID:        "student_id_01",
						StartDate:     now.Add(-10 * 24 * time.Hour),
						EndDate:       now.Add(-5 * 24 * time.Hour),
					},
				}, nil)
			},
		},
		{
			name: "return err when query GetByUserAndCourse fail",
			req: &mpb.ScheduleStudentClassRequest{
				StudentId:        "student_id_01",
				StudentPackageId: "student_package_id_01",
				CourseId:         "course_id_01",
				ClassId:          "class_id_01",
				StartTime:        timestamppb.New(now.Add(-10 * 24 * time.Hour)),
				EndTime:          timestamppb.New(now.Add(100 * 24 * time.Hour)),
				EffectiveDate:    timestamppb.New(now.Add(20 * 24 * time.Hour)),
			},
			expectedResp:           false,
			expectedCurrentClassID: "",
			expectedErr:            fmt.Errorf("query class members fail: query err"),
			setup: func(ctx context.Context) {
				mockClassMemberRepo.On("GetByUserAndCourse", mock.Anything, mockDB, "student_id_01", "course_id_01").Once().Return(nil, fmt.Errorf("query err"))
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			isReserveClass, currentClassID, err := rcc.CheckWillReserveClass(ctx, tc.req)
			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedResp, isReserveClass)
			assert.Equal(t, tc.expectedCurrentClassID, currentClassID)
		})
	}
}

func TestReserveClassCommandHandler_ReserveStudentClass(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	now := time.Now()
	rcc, mockDB, tx, mockReserveClassRepo, _, mockJsm := setupMock()

	testCases := []struct {
		name        string
		req         *mpb.ScheduleStudentClassRequest
		expectedErr error
		setup       func(ctx context.Context)
	}{
		{
			name: "return err when build reserve class domain fail",
			req: &mpb.ScheduleStudentClassRequest{
				StudentPackageId: "student_package_id_01",
				CourseId:         "course_id_01",
				ClassId:          "class_id_01",
				StartTime:        timestamppb.New(now.Add(-10 * 24 * time.Hour)),
				EndTime:          timestamppb.New(now.Add(100 * 24 * time.Hour)),
				EffectiveDate:    timestamppb.New(now.Add(20 * 24 * time.Hour)),
			},
			expectedErr: fmt.Errorf("build reserve class err: invalid reserve class: ReserveClass.StudentID cannot be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "return err when build reserve class domain fail",
			req: &mpb.ScheduleStudentClassRequest{
				StudentId:        "student_id_01",
				StudentPackageId: "student_package_id_01",
				CourseId:         "course_id_01",
				ClassId:          "class_id_01",
				StartTime:        timestamppb.New(now.Add(-10 * 24 * time.Hour)),
				EndTime:          timestamppb.New(now.Add(100 * 24 * time.Hour)),
				EffectiveDate:    timestamppb.New(now.Add(20 * 24 * time.Hour)),
			},
			expectedErr: fmt.Errorf("UpsertReserveClass: insert fail"),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				tx.On("Rollback", ctx).Return(nil).Once()
				mockReserveClassRepo.On("DeleteOldReserveClass", ctx, tx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(pgtype.Text{String: "old_class_id"}, pgtype.Date{Time: now.Add(24 * time.Hour)}, nil)
				mockReserveClassRepo.On("InsertOne", ctx, tx, mock.Anything).Once().Return(fmt.Errorf("insert fail"))
			},
		},
		{
			name: "success",
			req: &mpb.ScheduleStudentClassRequest{
				StudentId:        "student_id_01",
				StudentPackageId: "student_package_id_01",
				CourseId:         "course_id_01",
				ClassId:          "class_id_01",
				StartTime:        timestamppb.New(now.Add(-10 * 24 * time.Hour)),
				EndTime:          timestamppb.New(now.Add(100 * 24 * time.Hour)),
				EffectiveDate:    timestamppb.New(now.Add(20 * 24 * time.Hour)),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				mockReserveClassRepo.On("DeleteOldReserveClass", ctx, tx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(pgtype.Text{String: "old_class_id"}, pgtype.Date{Time: now.Add(24 * time.Hour)}, nil)
				mockReserveClassRepo.On("InsertOne", ctx, tx, mock.Anything).Once().Return(nil)
				mockJsm.On("PublishAsyncContext", ctx, mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			err := rcc.ReserveStudentClass(ctx, tc.req, "current_class_id")
			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
