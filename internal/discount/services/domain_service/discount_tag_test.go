package services

import (
	"context"
	"fmt"
	"testing"
	"time"

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

func TestDiscountTagService_RetrieveEligibleDiscountTagsOfStudentInLocation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                  *mockDb.Ext
		userDiscountTagRepo *mockRepositories.MockUserDiscountTagRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				userDiscountTagRepo.On("GetDiscountTagsByUserIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.UserDiscountTag{}, nil)
			},
		},
		{
			Name:        "Fail case: Error when get discount tag by student ID and location ID",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				userDiscountTagRepo.On("GetDiscountTagsByUserIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.UserDiscountTag{}, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			userDiscountTagRepo = new(mockRepositories.MockUserDiscountTagRepo)
			testCase.Setup(testCase.Ctx)
			s := &DiscountTagService{
				DB:                  db,
				UserDiscountTagRepo: userDiscountTagRepo,
			}
			_, err := s.RetrieveEligibleDiscountTagsOfStudentInLocation(testCase.Ctx, db, mock.Anything, mock.Anything)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, userDiscountTagRepo)
		})
	}
}

func TestDiscountTagService_RetrieveDiscountEligibilityOfStudentProduct(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                  *mockDb.Ext
		userDiscountTagRepo *mockRepositories.MockUserDiscountTagRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				userDiscountTagRepo.On("GetDiscountEligibilityOfStudentProduct", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.UserDiscountTag{}, nil)
			},
		},
		{
			Name:        "Fail case: Error when get discount eligibility of student product",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				userDiscountTagRepo.On("GetDiscountEligibilityOfStudentProduct", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.UserDiscountTag{}, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			userDiscountTagRepo = new(mockRepositories.MockUserDiscountTagRepo)
			testCase.Setup(testCase.Ctx)
			s := &DiscountTagService{
				DB:                  db,
				UserDiscountTagRepo: userDiscountTagRepo,
			}
			_, err := s.RetrieveDiscountEligibilityOfStudentProduct(testCase.Ctx, db, mock.Anything, mock.Anything, mock.Anything)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, userDiscountTagRepo)
		})
	}
}

func TestDiscountTagService_RetrieveDiscountTagsWithActivityOnDate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                  *mockDb.Ext
		userDiscountTagRepo *mockRepositories.MockUserDiscountTagRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				userDiscountTagRepo.On("GetDiscountTagsWithActivityOnDate", ctx, mock.Anything, mock.Anything).Return([]*entities.UserDiscountTag{}, nil)
			},
		},
		{
			Name:        "Fail case: Error when get discount eligibility of student product",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				userDiscountTagRepo.On("GetDiscountTagsWithActivityOnDate", ctx, mock.Anything, mock.Anything).Return([]*entities.UserDiscountTag{}, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			userDiscountTagRepo = new(mockRepositories.MockUserDiscountTagRepo)
			testCase.Setup(testCase.Ctx)
			s := &DiscountTagService{
				DB:                  db,
				UserDiscountTagRepo: userDiscountTagRepo,
			}
			_, err := s.RetrieveDiscountTagsWithActivityOnDate(testCase.Ctx, db, time.Now())

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, userDiscountTagRepo)
		})
	}
}

func TestDiscountTagService_RetrieveUserIDsWithActivityOnDate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                  *mockDb.Ext
		userDiscountTagRepo *mockRepositories.MockUserDiscountTagRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				userDiscountTagRepo.On("GetUserIDsWithActivityOnDate", ctx, mock.Anything, mock.Anything).Return([]pgtype.Text{}, nil)
			},
		},
		{
			Name:        "Fail case: Error when get discount eligibility of student product",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				userDiscountTagRepo.On("GetUserIDsWithActivityOnDate", ctx, mock.Anything, mock.Anything).Return([]pgtype.Text{}, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			userDiscountTagRepo = new(mockRepositories.MockUserDiscountTagRepo)
			testCase.Setup(testCase.Ctx)
			s := &DiscountTagService{
				DB:                  db,
				UserDiscountTagRepo: userDiscountTagRepo,
			}
			_, err := s.RetrieveUserIDsWithActivityOnDate(testCase.Ctx, db, time.Now())

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, userDiscountTagRepo)
		})
	}
}

func TestDiscountTagService_UpdateDiscountTagOfStudentIDWithTimeSegment(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                  *mockDb.Ext
		userDiscountTagRepo *mockRepositories.MockUserDiscountTagRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				userDiscountTagRepo.On("SoftDeleteByTypesAndUserID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				userDiscountTagRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Happy case: No old data to soft delete",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				userDiscountTagRepo.On("SoftDeleteByTypesAndUserID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("0 RowsAffected"))
				userDiscountTagRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Fail case: error on soft delete",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				userDiscountTagRepo.On("SoftDeleteByTypesAndUserID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error on create user discount tag",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				userDiscountTagRepo.On("SoftDeleteByTypesAndUserID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				userDiscountTagRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			userDiscountTagRepo = new(mockRepositories.MockUserDiscountTagRepo)
			testCase.Setup(testCase.Ctx)
			s := &DiscountTagService{
				DB:                  db,
				UserDiscountTagRepo: userDiscountTagRepo,
			}
			err := s.UpdateDiscountTagOfStudentIDWithTimeSegment(testCase.Ctx, db, mock.Anything, mock.Anything, []string{mock.Anything}, []entities.TimestampSegment{
				{
					StartDate: time.Now(),
					EndDate:   time.Now().AddDate(0, 1, 0),
				},
			})

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, userDiscountTagRepo)
		})
	}
}
