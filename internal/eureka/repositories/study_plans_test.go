package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func studyPlanRepoWithMock() (*StudyPlanRepo, *testutil.MockDB) {
	r := &StudyPlanRepo{}
	return r, testutil.NewMockDB()
}

type InsertTestCase struct {
	name         string
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context, db *mock_database.QueryExecer, row *mock_database.Row)
	queryCheck   func(db *mock_database.QueryExecer) error
}

func TestInsert(t *testing.T) {
	t.Parallel()
	studyPlanRepo := &StudyPlanRepo{}
	studyPlan := &entities.StudyPlan{
		ID: pgtype.Text{String: "study-plan", Status: pgtype.Present},
	}
	queryCheck := func(db *mock_database.QueryExecer) error {
		query := db.Calls[0].Arguments[1].(string)
		if !strings.Contains(query, "RETURNING study_plan_id") {
			return fmt.Errorf("Insert does not return study_plan_id")
		}
		return nil
	}
	testCases := []InsertTestCase{
		{
			name:        "error scan",
			req:         studyPlan,
			expectedErr: fmt.Errorf("error insert: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context, db *mock_database.QueryExecer, row *mock_database.Row) {
				fieldNames := database.GetFieldNames(studyPlan)
				scanFields := database.GetScanFields(studyPlan, fieldNames)
				var args []interface{}
				args = append(args, mock.Anything, mock.Anything)
				args = append(args, scanFields...)
				db.On("QueryRow", args...).Once().Return(row)
				row.On("Scan", mock.Anything).Once().Return(pgx.ErrNoRows)
			},
			queryCheck: queryCheck,
		},
		{
			name:        "happy case",
			req:         studyPlan,
			expectedErr: nil,
			setup: func(ctx context.Context, db *mock_database.QueryExecer, row *mock_database.Row) {
				fieldNames := database.GetFieldNames(studyPlan)
				scanFields := database.GetScanFields(studyPlan, fieldNames)
				var args []interface{}
				args = append(args, mock.Anything, mock.Anything)
				args = append(args, scanFields...)
				db.On("QueryRow", args...).Once().Return(row)
				row.On("Scan", mock.Anything).Once().Return(nil)
			},
			queryCheck: queryCheck,
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		db := &mock_database.QueryExecer{}
		row := &mock_database.Row{}
		testCase.setup(ctx, db, row)
		_, err := studyPlanRepo.Insert(ctx, db, testCase.req.(*entities.StudyPlan))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
		err = testCase.queryCheck(db)
		if err != nil {
			assert.Nil(t, err)
		}
	}

	return
}

func TestRetrieveByCourseID(t *testing.T) {
	t.Parallel()
	courseID := database.Text("course-id")
	limit := uint32(10)
	studyPlanName := database.Text("study-plan-name")
	studyPlanID := database.Text("study-plan-id")
	r, mockDB := studyPlanRepoWithMock()

	t.Run("err select", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		args := &RetrieveStudyPlanByCourseArgs{
			CourseID:      courseID,
			Limit:         limit,
			StudyPlanName: studyPlanName,
			StudyPlanID:   studyPlanID,
		}

		mockDB.MockQueryArgs(
			t,
			pgx.ErrNoRows,
			mock.Anything,
			mock.AnythingOfType("string"),
			&args.CourseID,
			&args.StudyPlanName,
			&args.StudyPlanID,
			&args.Limit,
		)

		contents, err := r.RetrieveByCourseID(ctx, mockDB.DB, args)

		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, contents)
	})

	t.Run("success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		args := &RetrieveStudyPlanByCourseArgs{
			CourseID:      courseID,
			Limit:         limit,
			StudyPlanName: studyPlanName,
			StudyPlanID:   studyPlanID,
		}

		mockDB.MockQueryArgs(
			t,
			nil,
			mock.Anything,
			mock.AnythingOfType("string"),
			&args.CourseID,
			&args.StudyPlanName,
			&args.StudyPlanID,
			&args.Limit,
		)
		e := &entities.StudyPlan{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{values})
		contents, err := r.RetrieveByCourseID(ctx, mockDB.DB, args)
		assert.Nil(t, err)
		assert.Equal(t, []*entities.StudyPlan{e}, contents)
		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})
}

func TestRetrieveStudyPlanIdentity(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	rows := &mock_database.Rows{}
	repo := &StudyPlanRepo{}

	studyPlanItemIDs := []string{"study-plan-item-id-1", "study-plan-item-id-2"}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         studyPlanItemIDs,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "query error",
			req:         studyPlanItemIDs,
			expectedErr: fmt.Errorf("StudyPlanRepo.RetrieveStudyPlanIdentity.Query: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
				rows.On("Close").Once().Return(nil)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "scan error",
			req:         studyPlanItemIDs,
			expectedErr: fmt.Errorf("StudyPlanRepo.RetrieveStudyPlanIdentity.Scan: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.([]string)
		_, err := repo.RetrieveStudyPlanIdentity(ctx, db, database.TextArray(req))
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestRecursiveSoftDeleteStudyPlanByStudyPlanIDInCourse(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	rows := &mock_database.Rows{}

	studyPlanRepo := &StudyPlanRepo{}

	var args pgtype.Text
	args.Set("true")

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         args,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := studyPlanRepo.RecursiveSoftDeleteStudyPlanByStudyPlanIDInCourse(ctx, db, testCase.req.(pgtype.Text))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}

	return
}

func Test_RetrieveMasterByCourseIDs(t *testing.T) {
	t.Parallel()
	type TestCase struct {
		name         string
		intput1      pgtype.Text
		intput2      pgtype.TextArray
		expectedResp interface{}
		expectedErr  error
		setup        func(ctx context.Context)
	}
	r, mockDB := studyPlanRepoWithMock()
	typeMock := database.Text("mock-type")
	courseIDs := database.TextArray([]string{"course-id-1", "course-id-2"})
	testCases := []TestCase{
		{
			name:        "happy case",
			intput1:     typeMock,
			intput2:     courseIDs,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				e := &entities.StudyPlan{}
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &courseIDs, &typeMock)
				fields, values := e.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
		{
			name:        "err query",
			intput1:     typeMock,
			intput2:     courseIDs,
			expectedErr: fmt.Errorf("err db.Query: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				e := &entities.StudentStudyPlan{}
				mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, &courseIDs, &typeMock)
				fields, values := e.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.setup(ctx)
			_, err := r.RetrieveMasterByCourseIDs(ctx, mockDB.DB, tc.intput1, tc.intput2)
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, tc.expectedErr, err)
			}
			if tc.expectedErr == nil {
				e := &entities.StudyPlan{}
				fields, _ := e.FieldMap()
				mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
				mockDB.RawStmt.AssertSelectedFields(t, fields...)
			}
		})
	}
}

func Test_RetrieveCombineStudent(t *testing.T) {
	t.Parallel()
	type TestCase struct {
		name         string
		input1       pgtype.TextArray
		expectedResp interface{}
		expectedErr  error
		setup        func(ctx context.Context)
	}
	mockDB := testutil.NewMockDB()
	rows := mockDB.Rows
	r := &StudyPlanRepo{}
	e := &entities.StudyPlan{}
	fields, _ := e.FieldMap()
	scanFields := database.GetScanFields(e, fields)
	var (
		studentID pgtype.Text
	)
	scanFields = append(scanFields, &studentID)

	bookIDs := database.TextArray([]string{"book-1", "book-2", "book-3"})
	testCases := []TestCase{
		{
			name:        "happy case",
			input1:      bookIDs,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &bookIDs)
				mockDB.DB.On("Query").Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", scanFields...).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "err query",
			input1:      bookIDs,
			expectedErr: fmt.Errorf("row.Err: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &bookIDs)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(pgx.ErrNoRows)
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.setup(ctx)
			_, err := r.RetrieveCombineStudent(ctx, mockDB.DB, tc.input1)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}
