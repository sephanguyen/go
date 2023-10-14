package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/pkg/errors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNotificationClassMemberRepo_Upsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	repo := &NotificationClassMemberRepo{}
	testCases := []struct {
		Name    string
		Request *entities.NotificationClassMember
		Err     error
		SetUp   func(ctx context.Context)
	}{
		{
			Name:    "happy case",
			Request: &entities.NotificationClassMember{},
			Err:     nil,
			SetUp: func(ctx context.Context) {
				e := &entities.NotificationClassMember{}
				fields := database.GetFieldNames(e)
				values := database.GetScanFields(e, fields)
				mockValues := make([]interface{}, 0, len(values)+2)
				mockValues = append(mockValues, mock.Anything)
				mockValues = append(mockValues, mock.AnythingOfType("string"))
				for range values {
					mockValues = append(mockValues, mock.Anything)
				}
				db.On("Exec", mockValues...).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
		},
		{
			Name:    "unique constraint",
			Request: &entities.NotificationClassMember{},
			Err:     errors.Wrap(&pgconn.PgError{Code: pgerrcode.UniqueViolation}, "r.DB.ExecEx"),
			SetUp: func(ctx context.Context) {
				e := &entities.NotificationClassMember{}
				fields := database.GetFieldNames(e)
				values := database.GetScanFields(e, fields)
				mockValues := make([]interface{}, 0, len(values)+2)
				mockValues = append(mockValues, mock.Anything)
				mockValues = append(mockValues, mock.AnythingOfType("string"))
				for range values {
					mockValues = append(mockValues, mock.Anything)
				}
				db.On("Exec", mockValues...).Once().Return(pgconn.CommandTag([]byte(`0`)), &pgconn.PgError{Code: pgerrcode.UniqueViolation})
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.SetUp(ctx)
		err := repo.Upsert(ctx, db, testCase.Request)
		if testCase.Err != nil {
			assert.Equal(t, testCase.Err.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.Err, err)
		}
	}
}

func TestNotificationClassMemberRepo_SoftDeleteByFilter(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	repo := &NotificationClassMemberRepo{}
	testCases := []struct {
		Name       string
		StudentID  string
		CourseID   string
		LocationID string
		Err        error
		Filter     *NotificationClassMemberFilter
		SetUp      func(ctx context.Context)
	}{
		{
			Name:       "happy case",
			StudentID:  "student-id",
			CourseID:   "course-id",
			LocationID: "location-id",
			Err:        nil,
			SetUp: func(ctx context.Context) {
				ctx, span := interceptors.StartSpan(ctx, "NotificationClassMemberRepo.SoftDeleteByFilter")
				defer span.End()
				db.On("Exec", ctx, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
		},
		{
			Name:      "err",
			StudentID: "student-id",
			CourseID:  "course-id",
			Err:       fmt.Errorf("err db.Exec: %w", puddle.ErrClosedPool),
			SetUp: func(ctx context.Context) {
				ctx, span := interceptors.StartSpan(ctx, "NotificationClassMemberRepo.SoftDeleteByFilter")
				defer span.End()
				db.On("Exec", ctx, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything).Once().Return(pgconn.CommandTag([]byte(`0`)), puddle.ErrClosedPool)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.SetUp(ctx)
		testCase.Filter = NewNotificationClassMemberFilter()
		_ = testCase.Filter.StudentIDs.Set([]string{testCase.StudentID})
		_ = testCase.Filter.CourseIDs.Set([]string{testCase.CourseID})
		_ = testCase.Filter.LocationIDs.Set([]string{testCase.LocationID})
		err := repo.SoftDeleteByFilter(ctx, db, testCase.Filter)
		if testCase.Err != nil {
			assert.Equal(t, testCase.Err.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.Err, err)
		}
	}
}

func TestNotificationClassMemberRepo_BulkUpsert(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}

	testCases := []struct {
		Name        string
		Req         interface{}
		ExpectedErr error
		SetUp       func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Req: []*entities.NotificationClassMember{
				{
					ClassID:    database.Text("class-id-1"),
					CourseID:   database.Text("course-id-1"),
					StudentID:  database.Text("student-id-1"),
					LocationID: database.Text("location-id-1"),
					StartAt:    database.Timestamptz(time.Now()),
					EndAt:      database.Timestamptz(time.Now()),
				},
				{
					ClassID:    database.Text("class-id-2"),
					CourseID:   database.Text("course-id-2"),
					StudentID:  database.Text("student-id-2"),
					LocationID: database.Text("location-id-2"),
					StartAt:    database.Timestamptz(time.Now()),
					EndAt:      database.Timestamptz(time.Now()),
				},
			},
			ExpectedErr: nil,
			SetUp: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}

	repo := &NotificationClassMemberRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.BulkUpsert(ctx, db, testCase.Req.([]*entities.NotificationClassMember))
			if testCase.ExpectedErr == nil {
				assert.Nil(t, err)
			}
		})
	}
}
