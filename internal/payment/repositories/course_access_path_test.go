package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/utils"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func CourseAccessPathRepoWithSqlMock() (*CourseAccessPathRepo, *testutil.MockDB) {
	r := &CourseAccessPathRepo{}
	return r, testutil.NewMockDB()
}

func TestCourseAccessPathRepo_GetCourseAccessPathByUCourseIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		courseAccessPathRepoWithSqlMock *CourseAccessPathRepo
		mockDB                          *testutil.MockDB
	)

	testcases := []utils.TestCase{
		{
			Name:        constant.FailCaseErrorQuery,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, constant.ErrDefault,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
			},
		},
		{
			Name:        constant.FailCaseErrorQuery,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan",
					mock.Anything,
					mock.Anything,
				).Once().Return(constant.ErrDefault)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan",
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan",
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
				mockDB.Rows.On("Next").Once().Return(false)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			courseAccessPathRepoWithSqlMock, mockDB = CourseAccessPathRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)

			_, err := courseAccessPathRepoWithSqlMock.GetCourseAccessPathByUCourseIDs(testCase.Ctx, mockDB.DB, []string{"1", "2"})

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}
