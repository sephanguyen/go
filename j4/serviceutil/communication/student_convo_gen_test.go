package serviceutil

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/manabie-com/backend/j4/infras"
	"github.com/manabie-com/backend/j4/serviceutil"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mocks "github.com/manabie-com/backend/mock/j4/serviceutil"
	"github.com/manabie-com/backend/mock/testutil"
	spb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupMockForStudentPoolHasNoConvoInDB(db *mock_database.Ext, shamir *mocks.ShamirClient,
	userSvc *mocks.GrpcClient) {

	studentID := "student_1"
	convID := "conv_1"
	email := "email1"
	createUserRows := func() pgx.Rows {
		dbutil := testutil.NewMockDB()
		scannedValues := [][]interface{}{
			{&studentID, &email},
		}

		dbutil.MockScanArray(nil, []string{"user_id", "email"}, scannedValues)
		return dbutil.Rows
	}
	createEmptyConvoRows := func() pgx.Rows {
		rows := &mock_database.Rows{}
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(true)
		rows.On("Err").Once().Return(nil)
		return rows
	}
	createConvoRows := func() pgx.Rows {
		dbutil := testutil.NewMockDB()
		scannedValues := [][]interface{}{
			{&studentID, &convID},
		}

		dbutil.MockScanArray(nil, []string{"student_id", "conversation_id"}, scannedValues)
		return dbutil.Rows
	}

	// 1st scan see nothing
	db.On("Query", mock.Anything, mock.MatchedBy(func(query string) bool {
		return strings.Contains(query, "conversation_students")
	}), mock.Anything, mock.Anything).Once().Return(func(_ context.Context, _ string, _ ...interface{}) pgx.Rows {
		return createEmptyConvoRows()
	}, nil)
	// 2nd scan see convos created
	db.On("Query", mock.Anything, mock.MatchedBy(func(query string) bool {
		return strings.Contains(query, "conversation_students")
	}), mock.Anything, mock.Anything).Return(func(_ context.Context, _ string, _ ...interface{}) pgx.Rows {
		return createConvoRows()
	}, nil)

	// scan user in db return mock rows
	db.On("Query", mock.Anything, mock.MatchedBy(func(query string) bool {
		return strings.Contains(query, "from users")
	}), mock.Anything).Return(func(_ context.Context, _ string, _ ...interface{}) pgx.Rows {
		return createUserRows()
	}, nil)

	shamir.On("GenerateFakeToken", mock.Anything, mock.Anything).Return(&spb.GenerateFakeTokenResponse{}, nil)
	userSvc.On("ImportStudent", mock.Anything, mock.Anything).Return(&upb.ImportStudentResponse{Errors: nil}, nil)
}

func setupMockForStudentPoolAlwaysReturnRecord(
	db *mock_database.Ext,
	shamir *mocks.ShamirClient,
	userSvc *mocks.GrpcClient) {

	studentID := "student_1"
	convID := "conv_1"
	email := "email1"
	createUserRows := func() pgx.Rows {
		dbutil := testutil.NewMockDB()
		scannedValues := [][]interface{}{
			{&studentID, &email},
		}

		dbutil.MockScanArray(nil, []string{"user_id", "email"}, scannedValues)
		return dbutil.Rows
	}
	createConvoRows := func() pgx.Rows {
		dbutil := testutil.NewMockDB()
		scannedValues := [][]interface{}{
			{&studentID, &convID},
		}

		dbutil.MockScanArray(nil, []string{"student_id", "conversation_id"}, scannedValues)
		return dbutil.Rows
	}

	// scan convo in db return mock rows
	db.On("Query", mock.Anything, mock.MatchedBy(func(query string) bool {
		return strings.Contains(query, "conversation_students")
	}), mock.Anything, mock.Anything).Return(func(_ context.Context, _ string, _ ...interface{}) pgx.Rows {
		return createConvoRows()
	}, nil)

	// scan user in db return mock rows
	db.On("Query", mock.Anything, mock.MatchedBy(func(query string) bool {
		return strings.Contains(query, "from users")
	}), mock.Anything).Return(func(_ context.Context, _ string, _ ...interface{}) pgx.Rows {
		return createUserRows()
	}, nil)

	shamir.On("GenerateFakeToken", mock.Anything, mock.Anything).Return(&spb.GenerateFakeTokenResponse{}, nil)
	userSvc.On("ImportStudent", mock.Anything, mock.Anything).Return(&upb.ImportStudentResponse{Errors: nil}, nil)
}

func Test_StudentConvoPool(t *testing.T) {
	t.Parallel()
	type testSuites struct {
		setupFunc func(
			db *mock_database.Ext,
			shamir *mocks.ShamirClient,
			userSvc *mocks.GrpcClient)
		name string
	}
	tcases := []testSuites{
		{
			name:      "found available in db",
			setupFunc: setupMockForStudentPoolAlwaysReturnRecord,
		},
		{
			name:      "not found in db, create new student",
			setupFunc: setupMockForStudentPoolHasNoConvoInDB,
		},
	}
	for _, item := range tcases {
		t.Run(item.name, func(t *testing.T) {
			shamir := &mocks.ShamirClient{}
			userSvc := &mocks.GrpcClient{}
			mockDB := &mock_database.Ext{}
			item.setupFunc(mockDB, shamir, userSvc)

			pool := &StudentConvoPool{
				genPerBatch:    1,
				j4Cfg:          &infras.ManabieJ4Config{},
				bobDB:          mockDB,
				tomDB:          mockDB,
				userSvc:        userSvc,
				tokenGenerator: &serviceutil.TokenGenerator{ShamirCl: shamir},
				studentPoolMu:  &sync.Mutex{},
			}
			go pool.checkStudentPool(context.Background())
			for try := 0; try < 2; try++ {
				stu := pool.GetOne(context.Background())
				assert.Equal(t, stu.ConvID, "conv_1")
				assert.Equal(t, stu.UserID, "student_1")
			}
		})
	}

}
