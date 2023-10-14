package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func PackageCourseRepoWithSqlMock() (*PackageCourseRepo, *testutil.MockDB, *mock_database.Tx) {
	packageCourseRepo := &PackageCourseRepo{}
	return packageCourseRepo, testutil.NewMockDB(), &mock_database.Tx{}
}

func TestPackageCourseRepo_GetByPackageIDForUpdate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var productId string = "1"
	mockDB := testutil.NewMockDB()
	packageCourseRepo := &PackageCourseRepo{}
	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, productId)
		entities := &entities.PackageCourse{}
		fields, values := entities.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		packageCourses, err := packageCourseRepo.GetByPackageIDForUpdate(ctx, mockDB.DB, productId)
		assert.Nil(t, err)
		assert.NotNil(t, packageCourses)
	})
	t.Run("err case scan row", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, productId)
		entities := &entities.PackageCourse{}
		fields, values := entities.FieldMap()
		mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{
			values,
		})
		packageCourses, err := packageCourseRepo.GetByPackageIDForUpdate(ctx, mockDB.DB, productId)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, packageCourses)
	})

	t.Run("err case query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, productId)
		packageCourses, err := packageCourseRepo.GetByPackageIDForUpdate(ctx, mockDB.DB, productId)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, packageCourses)
	})
}

func TestPackageCourseRepo_Upsert(t *testing.T) {
	t.Parallel()
	db := &mockDb.Ext{}
	packageCourseRepo := &PackageCourseRepo{}
	testCases := []utils.TestCase{
		{
			Name: "happy case",
			Req: []*entities.PackageCourse{
				{
					PackageID:     pgtype.Text{String: "1", Status: pgtype.Present},
					CourseID:      pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					CourseWeight:  pgtype.Int4{Int: 2, Status: pgtype.Present},
					MandatoryFlag: pgtype.Bool{Bool: true, Status: pgtype.Present},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				batchResults := &mockDb.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(constant.SuccessCommandTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			Name: "happy case: upsert multiple parents",
			Req: []*entities.PackageCourse{
				{
					PackageID:     pgtype.Text{String: "1", Status: pgtype.Present},
					CourseID:      pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					CourseWeight:  pgtype.Int4{Int: 2, Status: pgtype.Present},
					MandatoryFlag: pgtype.Bool{Bool: true, Status: pgtype.Present},
				},
				{
					PackageID:     pgtype.Text{String: "1", Status: pgtype.Present},
					CourseID:      pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					CourseWeight:  pgtype.Int4{Int: 3, Status: pgtype.Present},
					MandatoryFlag: pgtype.Bool{Bool: true, Status: pgtype.Present},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				batchResults := &mockDb.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(constant.SuccessCommandTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			Name: "error send batch",
			Req: []*entities.PackageCourse{
				{
					PackageID:     pgtype.Text{String: "1", Status: pgtype.Present},
					CourseID:      pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					CourseWeight:  pgtype.Int4{Int: 2, Status: pgtype.Present},
					MandatoryFlag: pgtype.Bool{Bool: true, Status: pgtype.Present},
				},
			},
			ExpectedErr: errors.Wrap(puddle.ErrClosedPool, "batchResults.Exec"),
			Setup: func(ctx context.Context) {
				batchResults := &mockDb.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.Setup(ctx)
		err := packageCourseRepo.Upsert(ctx, db, pgtype.Text{String: "1", Status: pgtype.Present}, testCase.Req.([]*entities.PackageCourse))
		if testCase.ExpectedErr != nil {
			assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.ExpectedErr, err)
		}
	}
}

func TestPackageCourse_GetByPackageIDAndCourseID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB, _ := PackageCourseRepoWithSqlMock()
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		entities := &entities.PackageCourse{}
		fields, values := entities.FieldMap()
		mockDB.MockRowScanFields(nil, fields, values)
		material, err := repo.GetByPackageIDAndCourseID(ctx, mockDB.DB, mock.Anything, mock.Anything)
		assert.Nil(t, err)
		assert.NotNil(t, material)

	})
	t.Run(constant.FailCaseErrorRow, func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		entities := &entities.PackageCourse{}
		fields, values := entities.FieldMap()
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		material, err := repo.GetByPackageIDAndCourseID(ctx, mockDB.DB, mock.Anything, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, material)
	})
}
