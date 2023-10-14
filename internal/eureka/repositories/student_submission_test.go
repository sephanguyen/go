package repositories

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/multierr"
)

func studentSubmissionRepoWithMockSQL() (*StudentSubmissionRepo, *testutil.MockDB) {
	r := &StudentSubmissionRepo{}
	return r, testutil.NewMockDB()
}

func TestStudentSubmissionRepo_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := studentSubmissionRepoWithMockSQL()
	t.Run("simple success insert", func(t *testing.T) {
		// mocking 'now'
		now := time.Now()
		timeutil.Now = func() time.Time { return now }

		e := &entities.StudentSubmission{}
		expectExecArgs := append([]interface{}{ctx, mock.AnythingOfType("string")}, database.GetScanFields(e, database.GetFieldNames(e))...)
		mockDB.MockExecArgs(t, pgconn.CommandTag{}, nil, expectExecArgs...)

		assert.NoError(t, r.Create(ctx, mockDB.DB, e), "expecting no error returned")

		mockDB.RawStmt.AssertInsertedFields(t, database.GetFieldNames(e)...)
		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())

		assert.Equal(t, pgtype.Present, e.CreatedAt.Status, "expecting entities.CreatedAt with value")
		assert.True(t, now.Equal(e.CreatedAt.Time), "expecting entities.CreatedAt have 'now' value")
		assert.Equal(t, e.CreatedAt, e.UpdatedAt, "expecting updatedAt and createdAt is the same")
	})
}

func TestStudentSubmissionRepo_Get(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := studentSubmissionRepoWithMockSQL()
	t.Run("simple success select", func(t *testing.T) {
		id := database.Text("zaq123")
		e := &entities.StudentSubmission{}

		mockDB.MockQueryArgs(t, nil, ctx, mock.AnythingOfType("string"), &id)
		mockDB.MockScanFields(nil, database.GetFieldNames(e), database.GetScanFields(e, database.GetFieldNames(e))) // called by scan one

		r.Get(ctx, mockDB.DB, id)
		mockDB.RawStmt.AssertSelectedFields(t, database.GetFieldNames(e)...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"student_submission_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"deleted_at":            {HasNullTest: true},
		})
	})
}

func TestStudentSubmissionRepo_List(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	e := &entities.StudentSubmission{}
	fields := strings.Join(database.GetFieldNames(e), ",student_latest_submissions.")

	t.Run("list without course", func(t *testing.T) {
		filter := &StudentSubmissionFilter{}
		filter.CourseID.Set(nil)
		filter.Limit = 10

		m := mockDB{
			QueryFn: func(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
				expectedQuery := fmt.Sprintf(listStmtTpl1, fields, filter.Limit)
				if query != expectedQuery {
					t.Errorf("unexpected query: got: %v, want: %v", query, expectedQuery)
				}
				return nil, pgx.ErrNoRows
			},
		}

		r := &StudentSubmissionRepo{}
		r.List(ctx, m, filter)
	})

	t.Run("list with course", func(t *testing.T) {
		filter := &StudentSubmissionFilter{}
		filter.CourseID.Set("cid")
		filter.Limit = 10

		m := mockDB{
			QueryFn: func(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
				expectedQuery := fmt.Sprintf(listWithCourseStmtTpl1, fields, filter.Limit)
				if query != expectedQuery {
					t.Errorf("unexpected query: got: %v, want: %v", query, expectedQuery)
				}
				return nil, pgx.ErrNoRows
			},
		}

		r := &StudentSubmissionRepo{}
		r.List(ctx, m, filter)
	})
}

func TestStudentSubmissionRepo_UpdateGradeStatus(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	pgtextNull := pgtype.Text{
		Status: pgtype.Null,
	}
	t.Run("err update", func(t *testing.T) {
		m := mockDB{
			ExecFn: func(ctx context.Context, sqlstmt string, args ...interface{}) (pgconn.CommandTag, error) {
				return nil, pgx.ErrTxClosed
			},
		}
		r := &StudentSubmissionRepo{}
		err := r.UpdateGradeStatus(ctx, m, pgtextNull, pgtextNull, pgtextNull, pgtextNull)
		assert.Error(t, err)
		assert.Equal(t, pgx.ErrTxClosed, err)
	})
	t.Run("happy case", func(t *testing.T) {
		m := mockDB{
			ExecFn: func(ctx context.Context, sqlstmt string, args ...interface{}) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
		}
		r := &StudentSubmissionRepo{}
		err := r.UpdateGradeStatus(ctx, m, pgtextNull, pgtextNull, pgtextNull, pgtextNull)
		assert.NoError(t, err)
		assert.Equal(t, nil, err)
	})
}

func TestStudentSubmissionRepo_BulkUpdateStatus(t *testing.T) {
	t.Parallel()
	type BulkUpdateStatusInput struct {
		EditorID string
		Status   string
		Grades   []*entities.StudentSubmissionGrade
	}
	db := &mock_database.QueryExecer{}
	StudentSubmissionRepo := &StudentSubmissionRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: &BulkUpdateStatusInput{
				EditorID: "editor-id",
				Status:   "status",
				Grades: []*entities.StudentSubmissionGrade{
					{
						ID:       database.Text("grade-id"),
						GraderID: database.Text("grader-id"),
					},
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
			req: &BulkUpdateStatusInput{
				EditorID: "editor-id",
				Status:   "status",
				Grades: []*entities.StudentSubmissionGrade{
					{
						ID:       database.Text("grade-id 1"),
						GraderID: database.Text("grader-id"),
					},
					{
						ID:       database.Text("grade-id 2"),
						GraderID: database.Text("grader-id"),
					},
				},
			},
			expectedErr: fmt.Errorf("batchResults.Exec: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, pgx.ErrTxClosed)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		input := testCase.req.(*BulkUpdateStatusInput)
		err := StudentSubmissionRepo.BulkUpdateStatus(ctx, db, database.Text(input.EditorID), database.Text(input.Status), input.Grades)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentSubmissionRepo_DeleteByStudyPlanItemIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	pgTextArrayNull := pgtype.TextArray{
		Status: pgtype.Null,
	}
	pgTextNull := pgtype.Text{
		Status: pgtype.Null,
	}

	t.Run("err delete", func(t *testing.T) {
		m := mockDB{
			ExecFn: func(ctx context.Context, sqlstmt string, args ...interface{}) (pgconn.CommandTag, error) {
				return nil, pgx.ErrTxClosed
			},
		}
		r := &StudentSubmissionRepo{}
		err := r.DeleteByStudyPlanItemIDs(ctx, m, pgTextArrayNull, pgTextNull)
		assert.Error(t, err)
		assert.Equal(t, pgx.ErrTxClosed, err)
	})

	t.Run("happy case", func(t *testing.T) {
		m := mockDB{
			ExecFn: func(ctx context.Context, sqlstmt string, args ...interface{}) (pgconn.CommandTag, error) {
				return pgconn.CommandTag("1"), nil
			},
		}
		r := &StudentSubmissionRepo{}
		err := r.DeleteByStudyPlanItemIDs(ctx, m, pgTextArrayNull, pgTextNull)
		assert.NoError(t, err)
		assert.Equal(t, nil, err)
	})
}

func TestStudentSubmissionRepo_RetrieveByStudyPlanIdentities(t *testing.T) {
	t.Parallel()
	StudentSubmissionRepo, mockDB := studentSubmissionRepoWithMockSQL()
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*StudyPlanItemIdentity{
				{
					StudyPlanID:        database.Text("sp-id-1"),
					LearningMaterialID: database.Text("lm-id-1"),
					StudentID:          database.Text("student-id-1"),
				},
			},
			setup: func(ctx context.Context) {
				ss := &StudentSubmissionInfo{
					StudentSubmission: entities.StudentSubmission{ID: database.Text("ss-id-1")},
				}
				fields, values := ss.FieldMap()
				fields = append(fields, "course_id", "start_date", "end_date")
				values = append(values, &ss.CourseID, &ss.StartDate, &ss.EndDate)
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
					[]pgtype.Text{database.Text("student-id-1")},
					[]pgtype.Text{database.Text("sp-id-1")},
					[]pgtype.Text{database.Text("lm-id-1")},
				)
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			expectedResp: []*StudentSubmissionInfo{
				{StudentSubmission: entities.StudentSubmission{ID: database.Text("ss-id-1")}},
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		submissions, err := StudentSubmissionRepo.RetrieveByStudyPlanIdentities(ctx, mockDB.DB, testCase.req.([]*StudyPlanItemIdentity))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
		if testCase.expectedResp != nil {
			assert.Equal(t, testCase.expectedResp.([]*StudentSubmissionInfo), submissions)
		}
	}
}

type mockDB struct {
	QueryFn func(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	ExecFn  func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

func (m mockDB) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return m.QueryFn(ctx, query, args...)
}

func (mockDB) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return nil
}

func (m mockDB) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return m.ExecFn(ctx, sql, args...)
}
func (mockDB) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults { return nil }

func TestStudentSubmissionRepo_ListV2(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := &StudentSubmissionRepo{}
	mockDB := testutil.NewMockDB()
	e := &entities.StudentSubmission{}
	fields, values := e.FieldMap()
	results := make(entities.StudentSubmissions, 0, int(100))
	results = append(results, &entities.StudentSubmission{})

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockDB.MockScanFields(nil, fields, values)
			},
			req: &StudentSubmissionFilter{
				Limit: 100,
			},
			expectedResp: results,
			expectedErr:  nil,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockDB.MockScanFields(pgx.ErrNoRows, fields, values)
			},
			req: &StudentSubmissionFilter{
				Limit: 100,
			},
			expectedResp: nil,
			expectedErr:  fmt.Errorf("database.Select: %w", fmt.Errorf("rows.Scan: %w", pgx.ErrNoRows)),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.ListV2(ctx, mockDB.DB, testCase.req.(*StudentSubmissionFilter))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestStudentSubmissionRepo_GetListV2Statement(t *testing.T) {
	t.Parallel()

	pattern := regexp.MustCompile(`\s+|\t+`)
	now := time.Now()

	filter := &StudentSubmissionFilter{
		Limit: uint(100),
	}
	_ = multierr.Combine(
		filter.OffsetID.Set(nil),
		filter.CreatedAt.Set(nil),
		filter.StudentIDs.Set(nil),
		filter.Statuses.Set(nil),
		filter.StartDate.Set(now),
		filter.EndDate.Set(now),
		filter.AssignmentName.Set(nil),
		filter.CourseID.Set("course_id"),
		filter.ClassIDs.Set([]string{"class_id_1", "class_id_2"}),
		filter.LocationIDs.Set([]string{"location_id_1", "location_id_2"}),
		filter.StudentName.Set("student_name"),
	)

	testCases := []struct {
		Name         string
		Input        *StudentSubmissionFilter
		ExpectedStmt string
		ExpectedArgs []interface{}
	}{
		{
			Name: "Search all",
			Input: &StudentSubmissionFilter{
				Limit:          uint(100),
				OffsetID:       pgtype.Text{Status: pgtype.Null},
				StudentIDs:     pgtype.TextArray{Status: pgtype.Null},
				Statuses:       pgtype.TextArray{Status: pgtype.Null},
				CreatedAt:      pgtype.Timestamptz{Status: pgtype.Null},
				AssignmentName: pgtype.Text{Status: pgtype.Null},
				StartDate:      pgtype.Timestamptz{Status: pgtype.Null},
				EndDate:        pgtype.Timestamptz{Status: pgtype.Null},
				CourseID:       pgtype.Text{Status: pgtype.Null},
				LocationIDs:    pgtype.TextArray{Status: pgtype.Null},
				StudentName:    pgtype.Text{Status: pgtype.Null},
				ClassIDs:       pgtype.TextArray{Status: pgtype.Null},
			},
			ExpectedStmt: `
            SELECT sls.student_submission_id,
                   sls.study_plan_item_id,
                   sls.assignment_id,
                   sls.student_id,
                   sls.submission_content,
                   sls.check_list,
                   sls.note,
                   sls.student_submission_grade_id,
                   sls.status,
                   sls.created_at,
                   sls.updated_at,
                   sls.deleted_at,
                   sls.deleted_by,
                   sls.editor_id,
                   sls.complete_date,
                   sls.duration,
                   sls.correct_score,
                   sls.total_score,
                   sls.understanding_level,
                   sls.study_plan_id,
                   sls.learning_material_id
            FROM student_latest_submissions sls
                     JOIN assignments a ON a.assignment_id = sls.assignment_id
                     JOIN study_plans sp ON sp.study_plan_id = sls.study_plan_id
                     JOIN course_students cs ON cs.course_id = sp.course_id AND cs.student_id = sls.student_id
            WHERE 1 = 1
              AND sls.deleted_at IS NULL
              AND ($1::text IS NULL OR sls.student_submission_id < $1)
              AND ($2::_text IS NULL OR sls.student_id = ANY ($2))
              AND ($3::_text IS NULL OR sls.status = ANY ($3))
              AND ($4::timestamp IS NULL OR sls.created_at < $4)
              AND a.deleted_at IS NULL
              AND a.type <> $5
              AND ($6::text IS NULL OR a.name ILIKE '%' || $6 || '%')
              AND sp.deleted_at IS NULL
              AND cs.deleted_at IS NULL
            ORDER BY sls.student_submission_id DESC
            LIMIT 100;
			`,
			ExpectedArgs: []interface{}{
				&filter.OffsetID,
				&filter.StudentIDs,
				&filter.Statuses,
				&filter.CreatedAt,
				epb.AssignmentType_ASSIGNMENT_TYPE_TASK.String(),
				&filter.AssignmentName,
			},
		},
		{
			Name: "Search by start_date",
			Input: &StudentSubmissionFilter{
				Limit:          uint(100),
				OffsetID:       pgtype.Text{Status: pgtype.Null},
				StudentIDs:     pgtype.TextArray{Status: pgtype.Null},
				Statuses:       pgtype.TextArray{Status: pgtype.Null},
				CreatedAt:      pgtype.Timestamptz{Status: pgtype.Null},
				AssignmentName: pgtype.Text{Status: pgtype.Null},
				StartDate:      database.Timestamptz(now),
				EndDate:        database.Timestamptz(now),
				CourseID:       pgtype.Text{Status: pgtype.Null},
				LocationIDs:    pgtype.TextArray{Status: pgtype.Null},
				StudentName:    pgtype.Text{Status: pgtype.Null},
				ClassIDs:       pgtype.TextArray{Status: pgtype.Null},
			},
			ExpectedStmt: `
            SELECT sls.student_submission_id,
                   sls.study_plan_item_id,
                   sls.assignment_id,
                   sls.student_id,
                   sls.submission_content,
                   sls.check_list,
                   sls.note,
                   sls.student_submission_grade_id,
                   sls.status,
                   sls.created_at,
                   sls.updated_at,
                   sls.deleted_at,
                   sls.deleted_by,
                   sls.editor_id,
                   sls.complete_date,
                   sls.duration,
                   sls.correct_score,
                   sls.total_score,
                   sls.understanding_level,
                   sls.study_plan_id,
                   sls.learning_material_id
            FROM student_latest_submissions sls
                     JOIN assignments a ON a.assignment_id = sls.assignment_id
                     JOIN study_plans sp ON sp.study_plan_id = sls.study_plan_id
                     JOIN course_students cs ON cs.course_id = sp.course_id AND cs.student_id = sls.student_id
            WHERE 1 = 1
              AND sls.deleted_at IS NULL
              AND ($1::text IS NULL OR sls.student_submission_id < $1)
              AND ($2::_text IS NULL OR sls.student_id = ANY ($2))
              AND ($3::_text IS NULL OR sls.status = ANY ($3))
              AND ($4::timestamp IS NULL OR sls.created_at < $4)
              AND a.deleted_at IS NULL
              AND a.type <> $5
              AND ($6::text IS NULL OR a.name ILIKE '%' || $6 || '%')
              AND sp.deleted_at IS NULL
              AND cs.deleted_at IS NULL
              AND EXISTS(SELECT 1
                         FROM study_plan_items spi
                         WHERE spi.deleted_at IS NULL
                           AND spi.study_plan_item_id = sls.study_plan_item_id
                           AND spi.start_date BETWEEN $7 AND $8)
            ORDER BY sls.student_submission_id DESC
            LIMIT 100;
			`,
			ExpectedArgs: []interface{}{
				&filter.OffsetID,
				&filter.StudentIDs,
				&filter.Statuses,
				&filter.CreatedAt,
				epb.AssignmentType_ASSIGNMENT_TYPE_TASK.String(),
				&filter.AssignmentName,
				&filter.StartDate,
				&filter.EndDate,
			},
		},
		{
			Name: "Search by course_id",
			Input: &StudentSubmissionFilter{
				Limit:          uint(100),
				OffsetID:       pgtype.Text{Status: pgtype.Null},
				StudentIDs:     pgtype.TextArray{Status: pgtype.Null},
				Statuses:       pgtype.TextArray{Status: pgtype.Null},
				CreatedAt:      pgtype.Timestamptz{Status: pgtype.Null},
				AssignmentName: pgtype.Text{Status: pgtype.Null},
				StartDate:      pgtype.Timestamptz{Status: pgtype.Null},
				EndDate:        pgtype.Timestamptz{Status: pgtype.Null},
				CourseID:       database.Text("course_id"),
				LocationIDs:    pgtype.TextArray{Status: pgtype.Null},
				StudentName:    pgtype.Text{Status: pgtype.Null},
				ClassIDs:       pgtype.TextArray{Status: pgtype.Null},
			},
			ExpectedStmt: `
            SELECT sls.student_submission_id,
                   sls.study_plan_item_id,
                   sls.assignment_id,
                   sls.student_id,
                   sls.submission_content,
                   sls.check_list,
                   sls.note,
                   sls.student_submission_grade_id,
                   sls.status,
                   sls.created_at,
                   sls.updated_at,
                   sls.deleted_at,
                   sls.deleted_by,
                   sls.editor_id,
                   sls.complete_date,
                   sls.duration,
                   sls.correct_score,
                   sls.total_score,
                   sls.understanding_level,
                   sls.study_plan_id,
                   sls.learning_material_id
            FROM student_latest_submissions sls
                     JOIN assignments a ON a.assignment_id = sls.assignment_id
                     JOIN study_plans sp ON sp.study_plan_id = sls.study_plan_id
                     JOIN course_students cs ON cs.course_id = sp.course_id AND cs.student_id = sls.student_id
            WHERE 1 = 1
              AND sls.deleted_at IS NULL
              AND ($1::text IS NULL OR sls.student_submission_id < $1)
              AND ($2::_text IS NULL OR sls.student_id = ANY ($2))
              AND ($3::_text IS NULL OR sls.status = ANY ($3))
              AND ($4::timestamp IS NULL OR sls.created_at < $4)
              AND a.deleted_at IS NULL
              AND a.type <> $5
              AND ($6::text IS NULL OR a.name ILIKE '%' || $6 || '%')
              AND sp.deleted_at IS NULL
              AND cs.deleted_at IS NULL
              AND cs.course_id = $7
            ORDER BY sls.student_submission_id DESC
            LIMIT 100;
			`,
			ExpectedArgs: []interface{}{
				&filter.OffsetID,
				&filter.StudentIDs,
				&filter.Statuses,
				&filter.CreatedAt,
				epb.AssignmentType_ASSIGNMENT_TYPE_TASK.String(),
				&filter.AssignmentName,
				&filter.CourseID,
			},
		},
		{
			Name: "Search by course_id and location_ids",
			Input: &StudentSubmissionFilter{
				Limit:          uint(100),
				OffsetID:       pgtype.Text{Status: pgtype.Null},
				StudentIDs:     pgtype.TextArray{Status: pgtype.Null},
				Statuses:       pgtype.TextArray{Status: pgtype.Null},
				CreatedAt:      pgtype.Timestamptz{Status: pgtype.Null},
				AssignmentName: pgtype.Text{Status: pgtype.Null},
				StartDate:      pgtype.Timestamptz{Status: pgtype.Null},
				EndDate:        pgtype.Timestamptz{Status: pgtype.Null},
				CourseID:       database.Text("course_id"),
				LocationIDs:    database.TextArray([]string{"location_id_1", "location_id_2"}),
				StudentName:    pgtype.Text{Status: pgtype.Null},
				ClassIDs:       pgtype.TextArray{Status: pgtype.Null},
			},
			ExpectedStmt: `
            SELECT sls.student_submission_id,
                   sls.study_plan_item_id,
                   sls.assignment_id,
                   sls.student_id,
                   sls.submission_content,
                   sls.check_list,
                   sls.note,
                   sls.student_submission_grade_id,
                   sls.status,
                   sls.created_at,
                   sls.updated_at,
                   sls.deleted_at,
                   sls.deleted_by,
                   sls.editor_id,
                   sls.complete_date,
                   sls.duration,
                   sls.correct_score,
                   sls.total_score,
                   sls.understanding_level,
                   sls.study_plan_id,
                   sls.learning_material_id
            FROM student_latest_submissions sls
                     JOIN assignments a ON a.assignment_id = sls.assignment_id
                     JOIN study_plans sp ON sp.study_plan_id = sls.study_plan_id
                     JOIN course_students cs ON cs.course_id = sp.course_id AND cs.student_id = sls.student_id
            WHERE 1 = 1
              AND sls.deleted_at IS NULL
              AND ($1::text IS NULL OR sls.student_submission_id < $1)
              AND ($2::_text IS NULL OR sls.student_id = ANY ($2))
              AND ($3::_text IS NULL OR sls.status = ANY ($3))
              AND ($4::timestamp IS NULL OR sls.created_at < $4)
              AND a.deleted_at IS NULL
              AND a.type <> $5
              AND ($6::text IS NULL OR a.name ILIKE '%' || $6 || '%')
              AND sp.deleted_at IS NULL
              AND cs.deleted_at IS NULL
              AND EXISTS(SELECT 1
                         FROM course_students_access_paths csap
                         WHERE csap.deleted_at IS NULL
                           AND csap.course_student_id = cs.course_student_id
                           AND csap.course_id = $7
                           AND csap.location_id = ANY ($8))
            ORDER BY sls.student_submission_id DESC
            LIMIT 100;
			`,
			ExpectedArgs: []interface{}{
				&filter.OffsetID,
				&filter.StudentIDs,
				&filter.Statuses,
				&filter.CreatedAt,
				epb.AssignmentType_ASSIGNMENT_TYPE_TASK.String(),
				&filter.AssignmentName,
				&filter.CourseID,
				&filter.LocationIDs,
			},
		},
		{
			Name: "Search by course_id and class_ids",
			Input: &StudentSubmissionFilter{
				Limit:          uint(100),
				OffsetID:       pgtype.Text{Status: pgtype.Null},
				StudentIDs:     pgtype.TextArray{Status: pgtype.Null},
				Statuses:       pgtype.TextArray{Status: pgtype.Null},
				CreatedAt:      pgtype.Timestamptz{Status: pgtype.Null},
				AssignmentName: pgtype.Text{Status: pgtype.Null},
				StartDate:      pgtype.Timestamptz{Status: pgtype.Null},
				EndDate:        pgtype.Timestamptz{Status: pgtype.Null},
				CourseID:       database.Text("course_id"),
				LocationIDs:    pgtype.TextArray{Status: pgtype.Null},
				StudentName:    pgtype.Text{Status: pgtype.Null},
				ClassIDs:       database.TextArray([]string{"class_id_1", "class_id_2"}),
			},
			ExpectedStmt: `
            SELECT sls.student_submission_id,
                   sls.study_plan_item_id,
                   sls.assignment_id,
                   sls.student_id,
                   sls.submission_content,
                   sls.check_list,
                   sls.note,
                   sls.student_submission_grade_id,
                   sls.status,
                   sls.created_at,
                   sls.updated_at,
                   sls.deleted_at,
                   sls.deleted_by,
                   sls.editor_id,
                   sls.complete_date,
                   sls.duration,
                   sls.correct_score,
                   sls.total_score,
                   sls.understanding_level,
                   sls.study_plan_id,
                   sls.learning_material_id
            FROM student_latest_submissions sls
                     JOIN assignments a ON a.assignment_id = sls.assignment_id
                     JOIN study_plans sp ON sp.study_plan_id = sls.study_plan_id
                     JOIN course_students cs ON cs.course_id = sp.course_id AND cs.student_id = sls.student_id
            WHERE 1 = 1
              AND sls.deleted_at IS NULL
              AND ($1::text IS NULL OR sls.student_submission_id < $1)
              AND ($2::_text IS NULL OR sls.student_id = ANY ($2))
              AND ($3::_text IS NULL OR sls.status = ANY ($3))
              AND ($4::timestamp IS NULL OR sls.created_at < $4)
              AND a.deleted_at IS NULL
              AND a.type <> $5
              AND ($6::text IS NULL OR a.name ILIKE '%' || $6 || '%')
              AND sp.deleted_at IS NULL
              AND cs.deleted_at IS NULL
              AND cs.course_id = $7
              AND EXISTS(SELECT 1
                         FROM class_students cls
                         WHERE cls.deleted_at IS NULL
                           AND cls.student_id = sls.student_id
                           AND cls.class_id = ANY ($8))
            ORDER BY sls.student_submission_id DESC
            LIMIT 100;
			`,
			ExpectedArgs: []interface{}{
				&filter.OffsetID,
				&filter.StudentIDs,
				&filter.Statuses,
				&filter.CreatedAt,
				epb.AssignmentType_ASSIGNMENT_TYPE_TASK.String(),
				&filter.AssignmentName,
				&filter.CourseID,
				&filter.ClassIDs,
			},
		},
		{
			Name: "Search by course_id, location_ids and class_ids",
			Input: &StudentSubmissionFilter{
				Limit:          uint(100),
				OffsetID:       pgtype.Text{Status: pgtype.Null},
				StudentIDs:     pgtype.TextArray{Status: pgtype.Null},
				Statuses:       pgtype.TextArray{Status: pgtype.Null},
				CreatedAt:      pgtype.Timestamptz{Status: pgtype.Null},
				AssignmentName: pgtype.Text{Status: pgtype.Null},
				StartDate:      pgtype.Timestamptz{Status: pgtype.Null},
				EndDate:        pgtype.Timestamptz{Status: pgtype.Null},
				CourseID:       database.Text("course_id"),
				LocationIDs:    database.TextArray([]string{"location_id_1", "location_id_2"}),
				StudentName:    pgtype.Text{Status: pgtype.Null},
				ClassIDs:       database.TextArray([]string{"class_id_1", "class_id_2"}),
			},
			ExpectedStmt: `
            SELECT sls.student_submission_id,
                   sls.study_plan_item_id,
                   sls.assignment_id,
                   sls.student_id,
                   sls.submission_content,
                   sls.check_list,
                   sls.note,
                   sls.student_submission_grade_id,
                   sls.status,
                   sls.created_at,
                   sls.updated_at,
                   sls.deleted_at,
                   sls.deleted_by,
                   sls.editor_id,
                   sls.complete_date,
                   sls.duration,
                   sls.correct_score,
                   sls.total_score,
                   sls.understanding_level,
                   sls.study_plan_id,
                   sls.learning_material_id
            FROM student_latest_submissions sls
                     JOIN assignments a ON a.assignment_id = sls.assignment_id
                     JOIN study_plans sp ON sp.study_plan_id = sls.study_plan_id
                     JOIN course_students cs ON cs.course_id = sp.course_id AND cs.student_id = sls.student_id
            WHERE 1 = 1
              AND sls.deleted_at IS NULL
              AND ($1::text IS NULL OR sls.student_submission_id < $1)
              AND ($2::_text IS NULL OR sls.student_id = ANY ($2))
              AND ($3::_text IS NULL OR sls.status = ANY ($3))
              AND ($4::timestamp IS NULL OR sls.created_at < $4)
              AND a.deleted_at IS NULL
              AND a.type <> $5
              AND ($6::text IS NULL OR a.name ILIKE '%' || $6 || '%')
              AND sp.deleted_at IS NULL
              AND cs.deleted_at IS NULL
              AND EXISTS(SELECT 1
                         FROM course_students_access_paths csap
                         WHERE csap.deleted_at IS NULL
                           AND csap.course_student_id = cs.course_student_id
                           AND csap.course_id = $7
                           AND csap.location_id = ANY ($8))
              AND EXISTS(SELECT 1
                         FROM class_students cls
                         WHERE cls.deleted_at IS NULL
                           AND cls.student_id = sls.student_id
                           AND cls.class_id = ANY ($9))
            ORDER BY sls.student_submission_id DESC
            LIMIT 100;
			`,
			ExpectedArgs: []interface{}{
				&filter.OffsetID,
				&filter.StudentIDs,
				&filter.Statuses,
				&filter.CreatedAt,
				epb.AssignmentType_ASSIGNMENT_TYPE_TASK.String(),
				&filter.AssignmentName,
				&filter.CourseID,
				&filter.LocationIDs,
				&filter.ClassIDs,
			},
		},
		{
			Name: "Search by location_ids",
			Input: &StudentSubmissionFilter{
				Limit:          uint(100),
				OffsetID:       pgtype.Text{Status: pgtype.Null},
				StudentIDs:     pgtype.TextArray{Status: pgtype.Null},
				Statuses:       pgtype.TextArray{Status: pgtype.Null},
				CreatedAt:      pgtype.Timestamptz{Status: pgtype.Null},
				AssignmentName: pgtype.Text{Status: pgtype.Null},
				StartDate:      pgtype.Timestamptz{Status: pgtype.Null},
				EndDate:        pgtype.Timestamptz{Status: pgtype.Null},
				CourseID:       pgtype.Text{Status: pgtype.Null},
				LocationIDs:    database.TextArray([]string{"location_id_1", "location_id_2"}),
				StudentName:    pgtype.Text{Status: pgtype.Null},
				ClassIDs:       pgtype.TextArray{Status: pgtype.Null},
			},
			ExpectedStmt: `
            SELECT sls.student_submission_id,
                   sls.study_plan_item_id,
                   sls.assignment_id,
                   sls.student_id,
                   sls.submission_content,
                   sls.check_list,
                   sls.note,
                   sls.student_submission_grade_id,
                   sls.status,
                   sls.created_at,
                   sls.updated_at,
                   sls.deleted_at,
                   sls.deleted_by,
                   sls.editor_id,
                   sls.complete_date,
                   sls.duration,
                   sls.correct_score,
                   sls.total_score,
                   sls.understanding_level,
                   sls.study_plan_id,
                   sls.learning_material_id
            FROM student_latest_submissions sls
                     JOIN assignments a ON a.assignment_id = sls.assignment_id
                     JOIN study_plans sp ON sp.study_plan_id = sls.study_plan_id
                     JOIN course_students cs ON cs.course_id = sp.course_id AND cs.student_id = sls.student_id
            WHERE 1 = 1
              AND sls.deleted_at IS NULL
              AND ($1::text IS NULL OR sls.student_submission_id < $1)
              AND ($2::_text IS NULL OR sls.student_id = ANY ($2))
              AND ($3::_text IS NULL OR sls.status = ANY ($3))
              AND ($4::timestamp IS NULL OR sls.created_at < $4)
              AND a.deleted_at IS NULL
              AND a.type <> $5
              AND ($6::text IS NULL OR a.name ILIKE '%' || $6 || '%')
              AND sp.deleted_at IS NULL
              AND cs.deleted_at IS NULL
              AND EXISTS(SELECT 1
                         FROM course_students_access_paths csap
                         WHERE csap.deleted_at IS NULL
                           AND csap.course_student_id = cs.course_student_id
                           AND csap.location_id = ANY ($7))
            ORDER BY sls.student_submission_id DESC
            LIMIT 100;
			`,
			ExpectedArgs: []interface{}{
				&filter.OffsetID,
				&filter.StudentIDs,
				&filter.Statuses,
				&filter.CreatedAt,
				epb.AssignmentType_ASSIGNMENT_TYPE_TASK.String(),
				&filter.AssignmentName,
				&filter.LocationIDs,
			},
		},
		{
			Name: "Search by student name",
			Input: &StudentSubmissionFilter{
				Limit:          uint(100),
				OffsetID:       pgtype.Text{Status: pgtype.Null},
				StudentIDs:     pgtype.TextArray{Status: pgtype.Null},
				Statuses:       pgtype.TextArray{Status: pgtype.Null},
				CreatedAt:      pgtype.Timestamptz{Status: pgtype.Null},
				AssignmentName: pgtype.Text{Status: pgtype.Null},
				StartDate:      pgtype.Timestamptz{Status: pgtype.Null},
				EndDate:        pgtype.Timestamptz{Status: pgtype.Null},
				CourseID:       pgtype.Text{Status: pgtype.Null},
				LocationIDs:    pgtype.TextArray{Status: pgtype.Null},
				StudentName:    database.Text("student_name"),
				ClassIDs:       pgtype.TextArray{Status: pgtype.Null},
			},
			ExpectedStmt: `
            SELECT sls.student_submission_id,
                   sls.study_plan_item_id,
                   sls.assignment_id,
                   sls.student_id,
                   sls.submission_content,
                   sls.check_list,
                   sls.note,
                   sls.student_submission_grade_id,
                   sls.status,
                   sls.created_at,
                   sls.updated_at,
                   sls.deleted_at,
                   sls.deleted_by,
                   sls.editor_id,
                   sls.complete_date,
                   sls.duration,
                   sls.correct_score,
                   sls.total_score,
                   sls.understanding_level,
                   sls.study_plan_id,
                   sls.learning_material_id
            FROM student_latest_submissions sls
                     JOIN assignments a ON a.assignment_id = sls.assignment_id
                     JOIN study_plans sp ON sp.study_plan_id = sls.study_plan_id
                     JOIN course_students cs ON cs.course_id = sp.course_id AND cs.student_id = sls.student_id
            WHERE 1 = 1
              AND sls.deleted_at IS NULL
              AND ($1::text IS NULL OR sls.student_submission_id < $1)
              AND ($2::_text IS NULL OR sls.student_id = ANY ($2))
              AND ($3::_text IS NULL OR sls.status = ANY ($3))
              AND ($4::timestamp IS NULL OR sls.created_at < $4)
              AND a.deleted_at IS NULL
              AND a.type <> $5
              AND ($6::text IS NULL OR a.name ILIKE '%' || $6 || '%')
              AND sp.deleted_at IS NULL
              AND cs.deleted_at IS NULL
              AND EXISTS(SELECT 1
                         FROM users u
                         WHERE u.deleted_at IS NULL AND u.user_id = sls.student_id AND u.name ILIKE '%' || $7 || '%')
            ORDER BY sls.student_submission_id DESC
            LIMIT 100;
			`,
			ExpectedArgs: []interface{}{
				&filter.OffsetID,
				&filter.StudentIDs,
				&filter.Statuses,
				&filter.CreatedAt,
				epb.AssignmentType_ASSIGNMENT_TYPE_TASK.String(),
				&filter.AssignmentName,
				&filter.StudentName,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			stmt, args := GetListV2Statement(testCase.Input)
			assert.Equal(t, strings.TrimSpace(pattern.ReplaceAllString(testCase.ExpectedStmt, " ")), pattern.ReplaceAllString(stmt, " "))
			assert.Equal(t, testCase.ExpectedArgs, args)
		})
	}
}
