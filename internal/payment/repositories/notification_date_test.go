package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	"github.com/manabie-com/backend/mock/testutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func NotificationDateRepoWithSqlMock() (*NotificationDateRepo, *testutil.MockDB) {
	notificationDateRepo := &NotificationDateRepo{}
	return notificationDateRepo, testutil.NewMockDB()
}

func TestNotificationDateRepo_Upsert(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	notificationDateRepo, mockDB := NotificationDateRepoWithSqlMock()
	db := mockDB.DB

	mockEntities := entities.StudentPackages{}
	_, fieldMap := mockEntities.FieldMap()
	tag := pgconn.CommandTag{1}
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	testCases := []utils.TestCase{
		{
			Name:        "Fail case: Error when delete query",
			Ctx:         ctx,
			Req:         &entities.NotificationDate{},
			ExpectedErr: fmt.Errorf("err delete NotificationDateRepo: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				db.On("Exec", args...).Return(tag, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when exec query",
			Ctx:         ctx,
			Req:         &entities.NotificationDate{},
			ExpectedErr: fmt.Errorf("error when upsert notification date: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				db.On("Exec", args...).Once().Return(tag, nil)
				db.On("Exec", args...).Once().Return(tag, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when have no affected row",
			Ctx:         ctx,
			Req:         &entities.NotificationDate{},
			ExpectedErr: fmt.Errorf("upsert notification date have no row affected"),
			Setup: func(ctx context.Context) {
				db.On("Exec", args...).Once().Return(tag, nil)
				db.On("Exec", args...).Once().Return(constant.FailCommandTag, nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req:  &entities.NotificationDate{},
			Setup: func(ctx context.Context) {
				db.On("Exec", args...).Once().Return(tag, nil)
				db.On("Exec", args...).Once().Return(constant.SuccessCommandTag, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			notificationDateRepo, mockDB = NotificationDateRepoWithSqlMock()
			db = mockDB.DB

			testCase.Setup(ctx)

			err := notificationDateRepo.Upsert(ctx, mockDB.DB, testCase.Req.(*entities.NotificationDate))
			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, testCase.ExpectedErr, err)
		})
	}
}

func TestNotificationDateRepo_GetByOrderType(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		notificationDateRepoWithSqlMock *NotificationDateRepo
		mockDB                          *testutil.MockDB
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when query row",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				pb.OrderType_ORDER_TYPE_LOA.String(),
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				_, fieldValues := (&entities.NotificationDate{}).FieldMap()
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(mockDB.Row)
				mockDB.Row.On("Scan", fieldValues...).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				pb.OrderType_ORDER_TYPE_LOA.String(),
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				_, fieldValues := (&entities.NotificationDate{}).FieldMap()
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(mockDB.Row)
				mockDB.Row.On("Scan", fieldValues...).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			notificationDateRepoWithSqlMock, mockDB = NotificationDateRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)

			orderType := testCase.Req.([]interface{})[0].(string)

			_, err := notificationDateRepoWithSqlMock.GetByOrderType(testCase.Ctx, mockDB.DB, orderType)
			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}
