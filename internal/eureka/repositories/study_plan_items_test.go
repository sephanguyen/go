package repositories

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStudyPlanItemRepo_BulkUpsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	studyPlanItemRepo := &StudyPlanItemRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.StudyPlanItem{
				{
					ID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
			},
		},
		{
			name: "error exec error",
			req: []*entities.StudyPlanItem{
				{
					ID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
			},
			expectedErr: fmt.Errorf("eureka_db.BulkUpsertStudyPlanItem error: error exec error"),
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, fmt.Errorf("error exec error"))

			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := studyPlanItemRepo.BulkInsert(ctx, db, testCase.req.([]*entities.StudyPlanItem))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, nil, err)
		}
	}

	return
}

func TestStudyPlanItemRepo_BulkCopy(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	studyPlanItemRepo := &StudyPlanItemRepo{}

	testCases := []TestCase{
		{
			name: "error \"original study plan ids and new study plan ids not match\"",
			req: []pgtype.TextArray{
				database.TextArray([]string{"1"}),
				database.TextArray([]string{"1", "2"}),
			},
			expectedErr: fmt.Errorf("original study plan ids and new study plan ids not match"),
			setup: func(ctx context.Context) {
				// no-op
			},
		},
		{
			name: "success",
			req: []pgtype.TextArray{
				database.TextArray([]string{"1", "2"}),
				database.TextArray([]string{"1", "2"}),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				batchResults.On("Close").Once().Return(nil)

				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)

				batchResults.On("Exec").Once().Return(nil, nil)
				batchResults.On("Exec").Once().Return(nil, nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)

		req := testCase.req.([]pgtype.TextArray)
		err := studyPlanItemRepo.BulkCopy(ctx, db, req[0], req[1])

		if err != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, nil)
		}
	}
}

func TestStudyPlanItemRepo_FindByStudyPlanID(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	studyPlanItemRepo := &StudyPlanItemRepo{}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         database.Text("study-plan-item-id"),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name:        "error no rows",
			req:         database.Text("study-plan-item-id"),
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		studyPlanItemID := testCase.req.(pgtype.Text)
		err := studyPlanItemRepo.MarkItemCompleted(ctx, db, studyPlanItemID)
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestStudyPlanItemRepo_FindAndSortByStudyPlanID(t *testing.T) {
	t.Parallel()
	rows := &mock_database.Rows{}
	db := &mock_database.Ext{}
	studyPlanItemRepo := &StudyPlanItemRepo{}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         database.TextArray([]string{"study-plan-item-id"}),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				fields := database.GetFieldNames(&entities.StudyPlanItem{})
				fieldDescriptions := make([]pgproto3.FieldDescription, 0, len(fields))
				for _, f := range fields {
					fieldDescriptions = append(fieldDescriptions, pgproto3.FieldDescription{Name: []byte(f)})
				}
				rows.On("FieldDescriptions").Return(fieldDescriptions)
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", database.GetScanFields(&entities.StudyPlanItem{}, fields)...).Once().Return(nil)

				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "error no rows",
			req:         database.TextArray([]string{"study-plan-item-id"}),
			expectedErr: fmt.Errorf("err db.Query: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)

			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		studyPlanItemID := testCase.req.(pgtype.TextArray)
		_, err := studyPlanItemRepo.FindAndSortByIDs(ctx, db, studyPlanItemID)
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestStudyPlanItemRepo_UnMarkItemCompleted(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	studyPlanItemRepo := &StudyPlanItemRepo{}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         database.Text("study-plan-item-id"),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name:        "error no rows",
			req:         database.Text("study-plan-item-id"),
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		studyPlanItemID := testCase.req.(pgtype.Text)
		err := studyPlanItemRepo.UnMarkItemCompleted(ctx, db, studyPlanItemID)
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestCountStudentPerStudyPlanItem(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	rows := &mock_database.Rows{}
	studyPlanItemRepo := &StudyPlanItemRepo{}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         []interface{}{database.Text("study_plan_item_id"), database.Bool(false)},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("pgtype.Text"), mock.AnythingOfType("pgtype.Bool")).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.AnythingOfType("*pgtype.Text"), mock.AnythingOfType("*pgtype.Int8")).Once().Return(nil)

				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "error no rows",
			req:         []interface{}{database.Text("study_plan_item_id"), database.Bool(false)},
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("pgtype.Text"), mock.AnythingOfType("pgtype.Bool")).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.([]interface{})
		_, _, err := studyPlanItemRepo.CountStudentInStudyPlanItem(ctx, db, req[0].(pgtype.Text), req[1].(pgtype.Bool))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}

	return
}

func TestRetrieveChildStudyPlanItem(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	rows := &mock_database.Rows{}
	studyPlanItemRepo := &StudyPlanItemRepo{}

	studyPlanItemID := database.Text("study-plan-item-id")
	userIDs := database.TextArray([]string{"user-id-1"})
	item := &entities.StudyPlanItem{}
	fields, values := item.FieldMap()
	var userID pgtype.Text

	allvalues := make([]interface{}, 0, len(fields)+1)
	allvalues = append(allvalues, &userID)
	allvalues = append(allvalues, values...)

	testCases := []TestCase{
		{
			name:         "happy case",
			req:          []interface{}{studyPlanItemID, userIDs},
			expectedErr:  nil,
			expectedResp: map[string]string{"user-id": "study-plan-item-id"},
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.AnythingOfType("string"), studyPlanItemID, userIDs).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", allvalues...).Once().Return(nil)
				rows.On("Next").Once().Return(false)
			},
		},
		{
			name:        "error no rows",
			req:         []interface{}{studyPlanItemID, userIDs},
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.AnythingOfType("string"), studyPlanItemID, userIDs).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.([]interface{})

		_, err := studyPlanItemRepo.RetrieveChildStudyPlanItem(ctx, db, req[0].(pgtype.Text), req[1].(pgtype.TextArray))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}

	return
}

func TestCountStudentStudyPlanItemsInClass(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	studyPlanItemRepo := &StudyPlanItemRepo{}
	row := &mock_database.Row{}

	filter := &CountStudentStudyPlanItemsInClassFilter{
		ClassID:         database.Text("class-id"),
		StudyPlanItemID: database.Text("study-plan-item-id"),
		IsCompleted:     database.Bool(true),
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         []interface{}{database.Text("class-id"), database.Text("study-plan-item-id"), database.Bool(true)},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &filter.ClassID, &filter.StudyPlanItemID, &filter.IsCompleted).Once().Return(row, nil)
				row.On("Scan", mock.AnythingOfType("*pgtype.Int8")).Once().Return(nil)
			},
		},
		{
			name:        "error no rows",
			req:         []interface{}{database.Text("class-id"), database.Text("study-plan-item-id"), database.Bool(true)},
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				db.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &filter.ClassID, &filter.StudyPlanItemID, &filter.IsCompleted).Once().Return(row, pgx.ErrNoRows)
				row.On("Scan", mock.AnythingOfType("*pgtype.Int8")).Once().Return(pgx.ErrNoRows)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.([]interface{})
		_, err := studyPlanItemRepo.CountStudentStudyPlanItemsInClass(ctx, db, &CountStudentStudyPlanItemsInClassFilter{
			ClassID:         req[0].(pgtype.Text),
			StudyPlanItemID: req[1].(pgtype.Text),
			IsCompleted:     req[2].(pgtype.Bool),
		})
		assert.Equal(t, testCase.expectedErr, err)

	}
}

func TestRetrieveStudyPlanContentStructuresByBooks(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	rows := &mock_database.Rows{}
	studyPlanItemRepo := &StudyPlanItemRepo{}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         []interface{}{"book1", "book2"},
			expectedErr: nil,
			expectedResp: map[string][]entities.ContentStructure{
				"sp1": {
					{
						BookID:   "book1",
						CourseID: "",
					},
				},
				"sp2": {
					{

						BookID:   "book2",
						CourseID: "",
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.AnythingOfType("string")).Once().Return(rows, nil)

				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					reflect.ValueOf(args[0]).Elem().SetString("sp1")
					reflect.ValueOf(args[1]).Elem().SetString("book1")
					reflect.ValueOf(args[2]).Elem().SetString("")
				}).Return(nil)

				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					reflect.ValueOf(args[0]).Elem().SetString("sp2")
					reflect.ValueOf(args[1]).Elem().SetString("book2")
					reflect.ValueOf(args[2]).Elem().SetString("")
				}).Return(nil)

				rows.On("Next").Once().Return(false)

				rows.On("Close").Once().Return(nil)
				rows.On("Err").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.([]interface{})

		resp, err := studyPlanItemRepo.RetrieveStudyPlanContentStructuresByBooks(
			ctx,
			db,
			database.TextArray([]string{req[0].(string), req[1].(string)}),
		)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedResp, resp)
		}
	}
}

func TestBulkSync(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	studyPlanItemRepo := &StudyPlanItemRepo{}

	testCases := []TestCase{
		{
			name: "all study plan items are new",
			req: []*entities.StudyPlanItem{
				{
					ID: pgtype.Text{String: "new1", Status: pgtype.Present},
				},
				{
					ID: pgtype.Text{String: "new2", Status: pgtype.Present},
				},
				{
					ID: pgtype.Text{String: "new3", Status: pgtype.Present},
				},
			},
			expectedErr: nil,
			expectedResp: []*entities.StudyPlanItem{
				{
					ID: pgtype.Text{String: "new1", Status: pgtype.Present},
				},
				{
					ID: pgtype.Text{String: "new2", Status: pgtype.Present},
				},
				{
					ID: pgtype.Text{String: "new3", Status: pgtype.Present},
				},
			},
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				batchResults.On("Close").Once().Return(nil)

				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)

				row1 := &mock_database.Row{}
				batchResults.On("QueryRow").Once().Return(row1)
				row1.On("Scan", mock.Anything).Once().Run(func(args mock.Arguments) {
					id := pgtype.Text{String: "new1", Status: pgtype.Present}
					reflect.ValueOf(args[0]).Elem().Set(reflect.ValueOf(id))
				}).Return(nil)

				row2 := &mock_database.Row{}
				batchResults.On("QueryRow").Once().Return(row2)
				row2.On("Scan", mock.Anything).Once().Run(func(args mock.Arguments) {
					id := pgtype.Text{String: "new2", Status: pgtype.Present}
					reflect.ValueOf(args[0]).Elem().Set(reflect.ValueOf(id))
				}).Return(nil)

				row3 := &mock_database.Row{}
				batchResults.On("QueryRow").Once().Return(row3)
				row3.On("Scan", mock.Anything).Once().Run(func(args mock.Arguments) {
					id := pgtype.Text{String: "new3", Status: pgtype.Present}
					reflect.ValueOf(args[0]).Elem().Set(reflect.ValueOf(id))
				}).Return(nil)
			},
		},
		{
			name: "some study plan items exist in DB",
			req: []*entities.StudyPlanItem{
				{
					ID: pgtype.Text{String: "exist1", Status: pgtype.Present},
				},
				{
					ID: pgtype.Text{String: "new1", Status: pgtype.Present},
				},
				{
					ID: pgtype.Text{String: "new2", Status: pgtype.Present},
				},
			},
			expectedErr: nil,
			expectedResp: []*entities.StudyPlanItem{
				{
					ID: pgtype.Text{String: "new1", Status: pgtype.Present},
				},
				{
					ID: pgtype.Text{String: "new2", Status: pgtype.Present},
				},
			},
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				batchResults.On("Close").Once().Return(nil)

				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)

				row1 := &mock_database.Row{}
				batchResults.On("QueryRow").Once().Return(row1)
				row1.On("Scan", mock.Anything).Once().Run(func(args mock.Arguments) {
					id := pgtype.Text{String: "exist2", Status: pgtype.Present}
					reflect.ValueOf(args[0]).Elem().Set(reflect.ValueOf(id))
				}).Return(nil)

				row2 := &mock_database.Row{}
				batchResults.On("QueryRow").Once().Return(row2)
				row2.On("Scan", mock.Anything).Once().Run(func(args mock.Arguments) {
					id := pgtype.Text{String: "new1", Status: pgtype.Present}
					reflect.ValueOf(args[0]).Elem().Set(reflect.ValueOf(id))
				}).Return(nil)

				row3 := &mock_database.Row{}
				batchResults.On("QueryRow").Once().Return(row3)
				row3.On("Scan", mock.Anything).Once().Run(func(args mock.Arguments) {
					id := pgtype.Text{String: "new2", Status: pgtype.Present}
					reflect.ValueOf(args[0]).Elem().Set(reflect.ValueOf(id))
				}).Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)

		resp, err := studyPlanItemRepo.BulkSync(ctx, db, testCase.req.([]*entities.StudyPlanItem))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedResp, resp)
		}
	}
}

func TestBulkSoftDeleteStudyPlanItemBy(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	studyPlanItemRepo := &StudyPlanItemRepo{}

	var args pgtype.TextArray
	args.Set([]string{"true"})

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         args,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := studyPlanItemRepo.DeleteStudyPlanItemsByStudyPlans(ctx, db, testCase.req.(pgtype.TextArray))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}

	return
}

func TestStudyPlanItemRepo_UpdateCompletedAtByID(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	repo := &StudyPlanItemRepo{}
	ctx := context.Background()
	ID := database.Text("id")
	time := database.Timestamptz(time.Now())

	t.Run("err update", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mock.Anything, mock.Anything)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), pgx.ErrTxClosed, args...)

		err := repo.UpdateCompletedAtByID(ctx, mockDB.DB, ID, time)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mock.Anything, mock.Anything)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
		err := repo.UpdateCompletedAtByID(ctx, mockDB.DB, ID, time)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertUpdatedTable(t, "study_plan_items")
		mockDB.RawStmt.AssertUpdatedFields(t, "completed_at", "updated_at")
	})
}

func TestSoftDeleteByStudyPlanItemIDs(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	repo := &StudyPlanItemRepo{}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         database.TextArray([]string{"assignment_id"}),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name:        "error exec",
			req:         database.TextArray([]string{"assignment_id"}),
			expectedErr: fmt.Errorf("db.Exec: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := repo.SoftDeleteByStudyPlanItemIDs(ctx, db, testCase.req.(pgtype.TextArray))
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestUpdateSchoolDateByStudyPlanItemIdentity(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	repo := &StudyPlanItemRepo{}

	type request struct {
		studentIDs  pgtype.TextArray
		lmID        pgtype.Text
		studyPlanID pgtype.Text
		schoolDate  pgtype.Timestamptz
	}

	testCases := []TestCase{
		{
			name: "happy case",
			req: &request{
				studentIDs:  database.TextArray([]string{"student_id_1"}),
				lmID:        database.Text("lm_id"),
				studyPlanID: database.Text("study_plan_id"),
				schoolDate:  database.Timestamptz(time.Now()),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name: "error exec",
			req: &request{
				studentIDs:  database.TextArray([]string{"student_id_1"}),
				lmID:        database.Text("lm_id"),
				studyPlanID: database.Text("study_plan_id"),
				schoolDate:  database.Timestamptz(time.Now()),
			},
			expectedErr: fmt.Errorf("db.Exec: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.(*request)
		err := repo.UpdateSchoolDateByStudyPlanItemIdentity(ctx, db, req.lmID, req.studyPlanID, req.studentIDs, req.schoolDate)
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestUpdateSchoolDateStudyPlanItems(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	repo := &StudyPlanItemRepo{}

	type Req struct {
		ids        pgtype.TextArray
		studentID  pgtype.Text
		schoolDate pgtype.Timestamptz
	}

	testCases := []TestCase{
		{
			name: "happy case",
			req: &Req{
				ids:        database.TextArray([]string{"id-1"}),
				studentID:  database.Text("student-1"),
				schoolDate: database.Timestamptz(time.Now()),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name: "error exec",
			req: &Req{
				ids:        database.TextArray([]string{"id-1"}),
				studentID:  database.Text("student-1"),
				schoolDate: database.Timestamptz(time.Now()),
			},
			expectedErr: fmt.Errorf("db.Exec: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.(*Req)
		err := repo.UpdateSchoolDate(ctx, db, req.ids, req.studentID, req.schoolDate)
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestUpdateStatusStudyPlanItems(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	repo := &StudyPlanItemRepo{}

	type Req struct {
		ids       pgtype.TextArray
		studentID pgtype.Text
		status    pgtype.Text
	}

	testCases := []TestCase{
		{
			name: "happy case",
			req: &Req{
				ids:       database.TextArray([]string{"id-1"}),
				studentID: database.Text("student-1"),
				status:    database.Text("status"),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name: "error exec",
			req: &Req{
				ids:       database.TextArray([]string{"id-1"}),
				studentID: database.Text("student-1"),
				status:    database.Text("status"),
			},
			expectedErr: fmt.Errorf("db.Exec: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.(*Req)
		err := repo.UpdateStatus(ctx, db, req.ids, req.studentID, req.status)
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestFetchByStudyProgressRequest(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	rows := &mock_database.Rows{}
	repo := &StudyPlanItemRepo{}

	type Req struct {
		courseID  pgtype.Text
		bookID    pgtype.Text
		studentID pgtype.Text
	}

	testCases := []TestCase{
		{
			name: "happy case",
			req: &Req{
				courseID:  database.Text("course_id"),
				bookID:    database.Text("book_id"),
				studentID: database.Text("student_id"),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name: "query error",
			req: &Req{
				courseID:  database.Text("course_id"),
				bookID:    database.Text("book_id"),
				studentID: database.Text("student_id"),
			},
			expectedErr: fmt.Errorf("StudyPlanItemRepo.FetchByStudyProgressRequest.Query: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
				rows.On("Close").Once().Return(nil)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name: "scan error",
			req: &Req{
				courseID:  database.Text("course_id"),
				bookID:    database.Text("book_id"),
				studentID: database.Text("student_id"),
			},
			expectedErr: fmt.Errorf("StudyPlanItemRepo.FetchByStudyProgressRequest.Scan: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.(*Req)
		_, err := repo.FetchByStudyProgressRequest(ctx, db, req.courseID, req.bookID, req.studentID)
		assert.Equal(t, testCase.expectedErr, err)
	}
}
func studyPlanItemRepoWithMock() (*StudyPlanItemRepo, *testutil.MockDB) {
	r := &StudyPlanItemRepo{}
	return r, testutil.NewMockDB()
}

func Test_RetrieveByBookContent(t *testing.T) {
	t.Parallel()
	type TestCase struct {
		name         string
		input1       pgtype.TextArray
		input2       pgtype.TextArray
		input3       pgtype.TextArray
		expectedResp interface{}
		expectedErr  error
		setup        func(ctx context.Context)
	}
	r, mockDB := studyPlanItemRepoWithMock()
	bookIDs := database.TextArray([]string{"book-id-1", "book-id-2", "book-id-3"})
	loIDs := database.TextArray([]string{"lo-1", "lo-2", "lo-3"})
	assignmentIDs := database.TextArray([]string{"assignment-1", "assignment-2", "assignment-3"})
	testCases := []TestCase{
		{
			name:        "happy case",
			input1:      bookIDs,
			input2:      loIDs,
			input3:      assignmentIDs,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				e := &entities.StudyPlanItem{}
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &bookIDs, &loIDs, &assignmentIDs)
				fields, values := e.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
		{
			name:        "err query",
			input1:      bookIDs,
			input2:      loIDs,
			input3:      assignmentIDs,
			expectedErr: fmt.Errorf("err db.Query: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				e := &entities.StudyPlanItem{}
				mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, &bookIDs, &loIDs, &assignmentIDs)
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
			_, err := r.RetrieveByBookContent(ctx, mockDB.DB, tc.input1, tc.input2, tc.input3)
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr, err)
			} else {
				e := &entities.StudyPlanItem{}
				fields, _ := e.FieldMap()
				mockDB.RawStmt.AssertSelectedFields(t, fields...)
			}
		})
	}
}
func Test_BulkUpdateStartEndDate(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	studyPlanItemRepo := &StudyPlanItemRepo{}
	type Req struct {
		studyPlanItemIDs   pgtype.TextArray
		updateType         sspb.UpdateStudyPlanItemsStartEndDateFields
		startDate, endDate pgtype.Timestamptz
	}

	testCases := []TestCase{
		{
			name: "happy case update start date and end date",
			req: &Req{
				studyPlanItemIDs: database.TextArray([]string{"study-plan-item-id-1", "study-plan-item-id-2"}),
				updateType:       sspb.UpdateStudyPlanItemsStartEndDateFields_ALL,
				startDate:        database.Timestamptz(time.Now()),
				endDate:          database.Timestamptz(time.Now()),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name: "happy case only update start date",
			req: &Req{
				studyPlanItemIDs: database.TextArray([]string{"study-plan-item-id-1", "study-plan-item-id-2"}),
				updateType:       sspb.UpdateStudyPlanItemsStartEndDateFields_START_DATE,
				startDate:        database.Timestamptz(time.Now()),
				endDate:          database.Timestamptz(time.Now()),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name: "happy case only update end date",
			req: &Req{
				studyPlanItemIDs: database.TextArray([]string{"study-plan-item-id-1", "study-plan-item-id-2"}),
				updateType:       sspb.UpdateStudyPlanItemsStartEndDateFields_END_DATE,
				startDate:        database.Timestamptz(time.Now()),
				endDate:          database.Timestamptz(time.Now()),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name: "db exec return error",
			req: &Req{
				studyPlanItemIDs: database.TextArray([]string{"study-plan-item-id-1", "study-plan-item-id-2"}),
				updateType:       sspb.UpdateStudyPlanItemsStartEndDateFields_ALL,
				startDate:        database.Timestamptz(time.Now()),
				endDate:          database.Timestamptz(time.Now()),
			},
			expectedErr: fmt.Errorf("db.Exec: error"),
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.(*Req)
		_, err := studyPlanItemRepo.BulkUpdateStartEndDate(ctx, db, pgtype.TextArray{}, req.updateType, req.startDate, req.endDate)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentPlanItems_UpdateCompletedAtToNullBySPIIdentity(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &StudyPlanItemRepo{}

	shuffleQuizSet := entities.ShuffledQuizSet{
		StudyPlanItemID:    database.Text("study-plan-item-id"),
		StudyPlanID:        database.Text("study-plan-id"),
		StudentID:          database.Text("student-id"),
		LearningMaterialID: database.Text("lm-id"),
	}
	studyPlanItem := entities.StudyPlanItem{
		ID: shuffleQuizSet.StudyPlanItemID,
	}

	stmt := fmt.Sprintf(`UPDATE %s
	SET completed_at = NULL, updated_at = NOW()
	WHERE study_plan_item_id = (
		SELECT DISTINCT study_plan_item_id	
		FROM shuffled_quiz_sets sqs 
		WHERE sqs.learning_material_id = $1
			AND sqs.student_id = $2
			AND sqs.study_plan_id = $3
			AND deleted_at IS NULL
	) AND deleted_at IS NULL`, studyPlanItem.TableName())

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", mock.Anything, stmt, shuffleQuizSet.LearningMaterialID, shuffleQuizSet.StudentID, shuffleQuizSet.StudyPlanID).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
			req: StudyPlanItemIdentity{
				StudentID:          shuffleQuizSet.StudentID,
				StudyPlanID:        shuffleQuizSet.StudyPlanID,
				LearningMaterialID: shuffleQuizSet.LearningMaterialID,
			},
			expectedErr:  nil,
			expectedResp: int64(1),
		},
		{
			name: "no row",
			setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", mock.Anything, stmt, database.Text("random-lm-id"), shuffleQuizSet.StudentID, shuffleQuizSet.StudyPlanID).Once().Return(pgconn.CommandTag([]byte(`0`)), pgx.ErrNoRows)
			},
			req: StudyPlanItemIdentity{
				StudentID:          shuffleQuizSet.StudentID,
				StudyPlanID:        shuffleQuizSet.StudyPlanID,
				LearningMaterialID: database.Text("random-lm-id"),
			},
			expectedErr:  fmt.Errorf("db.Exec: %w", pgx.ErrNoRows),
			expectedResp: int64(0),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.UpdateCompletedAtToNullByStudyPlanItemIdentity(ctx, mockDB.DB, testCase.req.(StudyPlanItemIdentity))
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}

func TestUpdateStudyPlanItemsStatus(t *testing.T) {
	//TODO: do later
}

func TestStudyPlanItemRepo_FindLearningMaterialByStudyPlanID(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}
	rows := &mock_database.Rows{}
	r := &StudyPlanItemRepo{}
	e := entities.StudyPlanItem{}

	stmt := fmt.Sprintf(`
	SELECT
	COALESCE(NULLIF(content_structure ->> 'lo_id', ''), content_structure->>'assignment_id', '') AS learning_material_id, study_plan_item_id 
		FROM %s spi
	WHERE study_plan_id  = $1
	AND deleted_at IS NULL`, e.TableName())

	spID := database.Text("_SP_ID")

	testCases := []TestCase{
		{
			name: "Happy case",
			req:  spID,
			setup: func(ctx context.Context) {
				mockDB.On("Query", mock.Anything, stmt, spID).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.setup(ctx)

			_, err := r.FindLearningMaterialByStudyPlanID(ctx, mockDB, tc.req.(pgtype.Text))

			assert.NoError(t, err)
		})
	}

}
