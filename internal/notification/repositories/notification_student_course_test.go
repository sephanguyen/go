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
	"github.com/manabie-com/backend/mock/testutil"
	"github.com/pkg/errors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNotificationStudentCourseRepo_Upsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	studentSubscriptionRepo := &NotificationStudentCourseRepo{}
	testCases := []struct {
		Name    string
		Request *entities.NotificationStudentCourse
		Err     error
		SetUp   func(ctx context.Context)
	}{
		{
			Name:    "happy case",
			Request: &entities.NotificationStudentCourse{},
			Err:     nil,
			SetUp: func(ctx context.Context) {
				e := &entities.NotificationStudentCourse{}
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
			Request: &entities.NotificationStudentCourse{},
			Err:     errors.Wrap(&pgconn.PgError{Code: pgerrcode.UniqueViolation}, "r.DB.ExecEx"),
			SetUp: func(ctx context.Context) {
				e := &entities.NotificationStudentCourse{}
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
		err := studentSubscriptionRepo.Upsert(ctx, db, testCase.Request)
		if testCase.Err != nil {
			assert.Equal(t, testCase.Err.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.Err, err)
		}
	}
}

func TestNotificationStudentCourseRepo_SoftDelete(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	repo := &NotificationStudentCourseRepo{}
	testCases := []struct {
		Name             string
		SoftDeleteFilter *SoftDeleteNotificationStudentCourseFilter
		Err              error
		Setup            func(ctx context.Context)
	}{
		{
			Name: "happy case",
			SoftDeleteFilter: &SoftDeleteNotificationStudentCourseFilter{
				StudentCourseIDs: database.TextArray([]string{"student-course-id"}),
				StudentIDs:       database.TextArray([]string{"student-id"}),
				CourseIDs:        database.TextArray([]string{"course-id"}),
				LocationIDs:      database.TextArray([]string{"location-id"}),
			},
			Err: nil,
			Setup: func(ctx context.Context) {
				ctx, span := interceptors.StartSpan(ctx, "TestNotificationStudentCourseRepo_SoftDelete")
				defer span.End()

				mockDB.DB.On("Exec", ctx, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
		},
		{
			Name: "case err",
			SoftDeleteFilter: &SoftDeleteNotificationStudentCourseFilter{
				StudentCourseIDs: database.TextArray([]string{"student-course-id"}),
				CourseIDs:        database.TextArray([]string{"course-id"}),
				StudentIDs:       database.TextArray([]string{"student-id"}),
				LocationIDs:      database.TextArray([]string{"location-id"}),
			},
			Err: fmt.Errorf("err db.Exec: %w", puddle.ErrClosedPool),
			Setup: func(ctx context.Context) {
				ctx, span := interceptors.StartSpan(ctx, "TestNotificationStudentCourseRepo_SoftDelete")
				defer span.End()

				mockDB.DB.On("Exec", ctx, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(pgconn.CommandTag([]byte(`0`)), puddle.ErrClosedPool)
			},
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		tc.Setup(ctx)
		err := repo.SoftDelete(ctx, db, tc.SoftDeleteFilter)
		if tc.Err != nil {
			assert.Equal(t, tc.Err.Error(), err.Error())
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestNotificationStudentCourseRepo_BulkUpsert(t *testing.T) {
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
			Req: []*entities.NotificationStudentCourse{
				{
					CourseID:   database.Text("course-id-1"),
					StudentID:  database.Text("student-id-1"),
					LocationID: database.Text("location-id-1"),
					StartAt:    database.Timestamptz(time.Now()),
					EndAt:      database.Timestamptz(time.Now()),
				},
				{
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

	repo := &NotificationStudentCourseRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.BulkUpsert(ctx, db, testCase.Req.([]*entities.NotificationStudentCourse))
			if testCase.ExpectedErr == nil {
				assert.Nil(t, err)
			}
		})
	}
}
