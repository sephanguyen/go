package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCourseStudyPlanUpsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	courseStudyPlanRepo := &CourseStudyPlanRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.CourseStudyPlan{
				{
					CourseID:    database.Text("course-id"),
					StudyPlanID: database.Text("study-plan-id"),
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "error send batch",
			req: []*entities.CourseStudyPlan{
				{
					CourseID:    database.Text("course-id-1"),
					StudyPlanID: database.Text("study-plan-id-1"),
				},
				{
					CourseID:    database.Text("course-id-2"),
					StudyPlanID: database.Text("study-plan-id-2"),
				},
				{
					CourseID:    database.Text("course-id-3"),
					StudyPlanID: database.Text("study-plan-id-3"),
				},
				{
					CourseID:    database.Text("course-id-4"),
					StudyPlanID: database.Text("study-plan-id-4"),
				},
			},
			expectedErr: fmt.Errorf("batchResults.Exec: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, pgx.ErrTxClosed)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := courseStudyPlanRepo.BulkUpsert(ctx, db, testCase.req.([]*entities.CourseStudyPlan))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}

	return
}

func TestSoftDeleteCourseStudyPlanBy(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	courseStudyPlanRepo := &CourseStudyPlanRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: &entities.CourseStudyPlan{
				CourseID:    database.Text("course-id"),
				StudyPlanID: database.Text("study-plan-id"),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := courseStudyPlanRepo.DeleteCourseStudyPlanBy(ctx, db, testCase.req.(*entities.CourseStudyPlan))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}

	return
}

func TestCourseStudyPlan_FindByCourseIDs(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	courseStudyPlanRepo := &CourseStudyPlanRepo{}
	courseIDs := database.TextArray([]string{"course-id-1", "course-id-2"})
	rows := mockDB.Rows
	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, &courseIDs).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "error no rows",
			expectedErr: fmt.Errorf("[Other Error]:%v", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, &courseIDs).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(pgx.ErrNoRows)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := courseStudyPlanRepo.FindByCourseIDs(ctx, mockDB.DB, courseIDs)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestCourseStudyPlan_ListCourseStudyPlans(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	courseStudyPlanRepo := &CourseStudyPlanRepo{}
	courseIDs := database.TextArray([]string{"course-id-1", "course-id-2"})
	bookIDs := database.TextArray([]string{"book-id-1", "book-id-2"})
	args := ListCourseStudyPlansArgs{
		CourseIDs: courseIDs,
		BookIDs:   bookIDs,
	}
	rows := mockDB.Rows
	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, &courseIDs, &bookIDs).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "error no rows",
			expectedErr: fmt.Errorf("[Other Error]:%v", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, &courseIDs, &bookIDs).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(pgx.ErrNoRows)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := courseStudyPlanRepo.ListCourseStudyPlans(ctx, mockDB.DB, &args)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
func TestCourseStudyPlan_ListCourseStatisticItems(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	courseStudyPlanRepo := &CourseStudyPlanRepo{}

	args := ListCourseStatisticItemsArgs{
		CourseID:    database.Text("CourseID"),
		StudyPlanID: database.Text("StudyPlanID"),
		ClassID:     pgtype.Text{Status: pgtype.Null},
	}
	rows := mockDB.Rows
	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, args.CourseID, args.StudyPlanID, args.ClassID).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "scan error",
			expectedErr: fmt.Errorf("rows.Scan: %v", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, args.CourseID, args.StudyPlanID, args.ClassID).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
		{
			name:        "error no rows",
			expectedErr: fmt.Errorf("rows.Err: %v", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, args.CourseID, args.StudyPlanID, args.ClassID).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(pgx.ErrNoRows)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := courseStudyPlanRepo.ListCourseStatisticItems(ctx, mockDB.DB, &args)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

// // write unit test for ListCourseStatisticV4 function
func TestCourseStudyPlan_ListCourseStatisticV4(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	courseStudyPlanRepo := &CourseStudyPlanRepo{}

	args := ListCourseStatisticItemsArgsV3{
		CourseID:    database.Text("CourseID"),
		StudyPlanID: database.Text("StudyPlanID"),
		ClassID:     pgtype.TextArray{Status: pgtype.Null},
		TagIDs:      pgtype.TextArray{Status: pgtype.Null},
		StudentIDs:  pgtype.TextArray{Status: pgtype.Null},
	}

	rows := mockDB.Rows
	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, &args.CourseID, &args.StudyPlanID, &args.TagIDs, &args.StudentIDs, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "scan error",
			expectedErr: fmt.Errorf("ItemStatistic rows.Scan: %v", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, &args.CourseID, &args.StudyPlanID, &args.TagIDs, &args.StudentIDs, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
		{
			name:        "error no rows",
			expectedErr: fmt.Errorf("%v", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, &args.CourseID, &args.StudyPlanID, &args.TagIDs, &args.StudentIDs, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(pgx.ErrNoRows)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, _, err := courseStudyPlanRepo.ListCourseStatisticV4(ctx, mockDB.DB, &args)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
