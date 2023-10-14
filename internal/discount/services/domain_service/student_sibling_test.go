package services

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/utils"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mockRepositories "github.com/manabie-com/backend/mock/discount/repositories"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStudentSibling_RetrieveStudentSiblingIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                *mockDb.Ext
		studentParentRepo *mockRepositories.MockStudentParentRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        constant.HappyCase + " single sibling",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentParentRepo.On("GetSiblingIDsByStudentID", ctx, db, mock.Anything).Return([]string{"sibling-1"}, nil)
			},
		},
		{
			Name:        constant.HappyCase + " multi sibling",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentParentRepo.On("GetSiblingIDsByStudentID", ctx, db, mock.Anything).Return([]string{"sibling-1", "sibling-2"}, nil)
			},
		},
		{
			Name:        "Fail case: Error when getting student sibling",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentParentRepo.On("GetSiblingIDsByStudentID", ctx, db, mock.Anything).Return([]string{}, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentParentRepo = new(mockRepositories.MockStudentParentRepo)

			testCase.Setup(testCase.Ctx)
			s := &StudentSiblingService{
				DB:                db,
				StudentParentRepo: studentParentRepo,
			}
			_, err := s.RetrieveStudentSiblingIDs(testCase.Ctx, db, mock.Anything)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentParentRepo)
		})
	}
}
