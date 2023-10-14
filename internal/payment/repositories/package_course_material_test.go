package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_db "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func PackageCourseMaterialRepoWithSqlMock() (*PackageCourseMaterialRepo, *testutil.MockDB, *mock_database.Tx) {
	packageCourseMaterialRepo := &PackageCourseMaterialRepo{}
	return packageCourseMaterialRepo, testutil.NewMockDB(), &mock_database.Tx{}
}
func TestPackageCourseMaterial_Upsert(t *testing.T) {
	t.Parallel()
	mockDb := &mock_db.Ext{}
	packageCourseMaterialRepo := &PackageCourseMaterialRepo{}
	testCases := []utils.TestCase{
		{
			Name: "happy case",
			Req: []*entities.PackageCourseMaterial{
				{
					PackageID:  pgtype.Text{String: "1", Status: pgtype.Present},
					CourseID:   pgtype.Text{String: constant.CourseName, Status: pgtype.Present},
					MaterialID: pgtype.Text{String: "3", Status: pgtype.Present},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				batchResults := &mock_db.BatchResults{}
				mockDb.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(constant.SuccessCommandTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			Name: "happy case: upsert multiple parents",
			Req: []*entities.PackageCourseMaterial{
				{
					PackageID:  pgtype.Text{String: "1", Status: pgtype.Present},
					CourseID:   pgtype.Text{String: constant.CourseName, Status: pgtype.Present},
					MaterialID: pgtype.Text{String: "3", Status: pgtype.Present},
				},
				{
					PackageID:  pgtype.Text{String: "1", Status: pgtype.Present},
					CourseID:   pgtype.Text{String: "Course-4", Status: pgtype.Present},
					MaterialID: pgtype.Text{String: "4", Status: pgtype.Present},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				batchResults := &mock_db.BatchResults{}
				mockDb.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(constant.SuccessCommandTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			Name: "error send batch",
			Req: []*entities.PackageCourseMaterial{
				{
					PackageID:  pgtype.Text{String: "1", Status: pgtype.Present},
					CourseID:   pgtype.Text{String: constant.CourseName, Status: pgtype.Present},
					MaterialID: pgtype.Text{String: "3", Status: pgtype.Present},
				},
			},
			ExpectedErr: errors.Wrap(puddle.ErrClosedPool, "batchResults.Exec"),
			Setup: func(ctx context.Context) {
				batchResults := &mock_db.BatchResults{}
				mockDb.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.Setup(ctx)
		err := packageCourseMaterialRepo.Upsert(ctx, mockDb, pgtype.Text{String: "1", Status: pgtype.Present}, testCase.Req.([]*entities.PackageCourseMaterial))
		if testCase.ExpectedErr != nil {
			assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.ExpectedErr, err)
		}
	}
}

func TestPackageCourseMaterial_GetToTalAssociatedByCourseIDAndPackageID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var total int32
	t.Run(constant.HappyCase, func(t *testing.T) {
		repo, mockDB, _ := PackageCourseMaterialRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Scan", &total).Once().Return(nil)
		_, err := repo.GetToTalAssociatedByCourseIDAndPackageID(ctx, mockDB.DB, "1", []string{})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run(constant.FailCaseErrorRow, func(t *testing.T) {
		repo, mockDB, _ := PackageCourseFeeRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Scan", &total).Once().Return(constant.ErrDefault)
		_, err := repo.GetToTalAssociatedByCourseIDAndPackageID(ctx, mockDB.DB, "1", []string{})
		assert.NotNil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
