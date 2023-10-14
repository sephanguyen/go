package services

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/discount/utils"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mockRepositories "github.com/manabie-com/backend/mock/discount/repositories"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDiscountTrackerService_TrackDiscount(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                         *mockDb.Ext
		studentDiscountTrackerRepo *mockRepositories.MockStudentDiscountTrackerRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentDiscountTrackerRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Fail case: Error when create student discount tracker",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentDiscountTrackerRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentDiscountTrackerRepo = new(mockRepositories.MockStudentDiscountTrackerRepo)

			testCase.Setup(testCase.Ctx)
			s := &DiscountTrackerService{
				DB:                         db,
				StudentDiscountTrackerRepo: studentDiscountTrackerRepo,
			}
			err := s.TrackDiscount(testCase.Ctx, db, &entities.StudentDiscountTracker{})

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentDiscountTrackerRepo)
		})
	}
}

func TestDiscountTrackerService_RetrieveSiblingDiscountTrackingHistoriesByStudentIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                         *mockDb.Ext
		studentDiscountTrackerRepo *mockRepositories.MockStudentDiscountTrackerRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentDiscountTrackerRepo.On("GetActiveTrackingByStudentIDs", ctx, mock.Anything, mock.Anything).Return([]entities.StudentDiscountTracker{
					{
						StudentID: pgtype.Text{String: "1", Status: pgtype.Present},
					},
					{
						StudentID: pgtype.Text{String: "2", Status: pgtype.Present},
					},
				}, nil)
			},
		},
		{
			Name:        "Fail case: Error when get active tracking of students",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentDiscountTrackerRepo.On("GetActiveTrackingByStudentIDs", ctx, mock.Anything, mock.Anything).Return([]entities.StudentDiscountTracker{}, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentDiscountTrackerRepo = new(mockRepositories.MockStudentDiscountTrackerRepo)

			testCase.Setup(testCase.Ctx)
			s := &DiscountTrackerService{
				DB:                         db,
				StudentDiscountTrackerRepo: studentDiscountTrackerRepo,
			}
			_, _, err := s.RetrieveSiblingDiscountTrackingHistoriesByStudentIDs(testCase.Ctx, db, []string{"1", "2"})

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentDiscountTrackerRepo)
		})
	}
}

func TestDiscountTrackerService_UpdateTrackingDurationByStudentProduct(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                         *mockDb.Ext
		studentDiscountTrackerRepo *mockRepositories.MockStudentDiscountTrackerRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentDiscountTrackerRepo.On("UpdateTrackingDurationByStudentProduct", ctx, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Fail case: Error when updating tracking duration of student product",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentDiscountTrackerRepo.On("UpdateTrackingDurationByStudentProduct", ctx, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentDiscountTrackerRepo = new(mockRepositories.MockStudentDiscountTrackerRepo)

			testCase.Setup(testCase.Ctx)
			s := &DiscountTrackerService{
				DB:                         db,
				StudentDiscountTrackerRepo: studentDiscountTrackerRepo,
			}
			err := s.UpdateTrackingDurationByStudentProduct(testCase.Ctx, db, entities.StudentProduct{})

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentDiscountTrackerRepo)
		})
	}
}
