package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrderItemCourseRepo_MultiCreate(t *testing.T) {
	t.Parallel()
	db := &mockDb.Ext{}
	orderItemCourseRepo := &OrderItemCourseRepo{}
	testCases := []utils.TestCase{
		{
			Name: "happy case",
			Req: []entities.OrderItemCourse{
				{
					PackageID: pgtype.Text{String: "1", Status: pgtype.Present},
					OrderID:   pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					CourseID:  pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
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
			Req: []entities.OrderItemCourse{
				{
					PackageID: pgtype.Text{String: "1", Status: pgtype.Present},
					OrderID:   pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					CourseID:  pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				},
				{
					PackageID: pgtype.Text{String: "2", Status: pgtype.Present},
					OrderID:   pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					CourseID:  pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
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
			Req: []entities.OrderItemCourse{
				{
					PackageID: pgtype.Text{String: "1", Status: pgtype.Present},
					OrderID:   pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					CourseID:  pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
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
		{
			Name: "error : order item course not inserted",
			Req: []entities.OrderItemCourse{
				{
					PackageID: pgtype.Text{String: "1", Status: pgtype.Present},
					OrderID:   pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					CourseID:  pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				},
			},
			ExpectedErr: fmt.Errorf("order item course not inserted"),
			Setup: func(ctx context.Context) {
				batchResults := &mockDb.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(constant.FailCommandTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.Setup(ctx)
		err := orderItemCourseRepo.MultiCreate(ctx, db, testCase.Req.([]entities.OrderItemCourse))
		if testCase.ExpectedErr != nil {
			assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.ExpectedErr, err)
		}
	}
}

func TestOrderItemCourseRepo_GetMapOrderItemCourseByOrderIDAndPackageID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockOrderItemCourseRepo := OrderItemCourseRepo{}
	db := testutil.NewMockDB()
	t.Run("Success", func(t *testing.T) {
		db.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.OrderItemCourse{
			CourseID: pgtype.Text{Status: pgtype.Present, String: "1"},
		}
		fields, values := e.FieldMap()

		db.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		orderItemCourses, err := mockOrderItemCourseRepo.GetMapOrderItemCourseByOrderIDAndPackageID(ctx, db.DB, "1", "1")
		assert.Nil(t, err)
		assert.NotNil(t, orderItemCourses)

	})
	t.Run("err case scan row", func(t *testing.T) {
		db.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.OrderItemCourse{}
		fields, values := e.FieldMap()

		db.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{
			values,
		})
		_, err := mockOrderItemCourseRepo.GetMapOrderItemCourseByOrderIDAndPackageID(ctx, db.DB, "1", "1")
		assert.True(t, errors.Is(err, pgx.ErrNoRows))

	})
	t.Run("err case query", func(t *testing.T) {
		db.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything,
			mock.Anything,
			mock.Anything)
		_, err := mockOrderItemCourseRepo.GetMapOrderItemCourseByOrderIDAndPackageID(ctx, db.DB, "1", "1")
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
	})
}
