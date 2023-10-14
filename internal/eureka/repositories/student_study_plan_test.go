package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func studentStudyPlanRepoWithMock() (*StudentStudyPlanRepo, *testutil.MockDB) {
	r := &StudentStudyPlanRepo{}
	return r, testutil.NewMockDB()
}

func TestListStudentAvailableContents(t *testing.T) {
	t.Run("err select", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, mockDB := studentStudyPlanRepoWithMock()

		studentID := database.Text("id")
		studyPlanIDs := database.TextArray([]string{})
		offset := database.Timestamptz(time.Now())
		bookID := database.Text("book-id")
		chapterID := database.Text("chapter-id")
		topicID := database.Text("topic-id")
		courseID := database.Text("course-id")
		q := &ListStudentAvailableContentsArgs{
			StudentID:    studentID,
			StudyPlanIDs: studyPlanIDs,
			Offset:       offset,
			BookID:       bookID,
			ChapterID:    chapterID,
			TopicID:      topicID,
			CourseID:     courseID,
		}
		mockDB.MockQueryArgs(
			t,
			puddle.ErrClosedPool,
			mock.Anything,
			mock.AnythingOfType("string"),
			&q.StudentID,
			&q.StudyPlanIDs,
			&q.Offset,
			&q.BookID,
			&q.ChapterID,
			&q.TopicID,
			&q.CourseID,
		)

		contents, err := r.ListStudentAvailableContents(ctx, mockDB.DB, q)

		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, contents)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, mockDB := studentStudyPlanRepoWithMock()

		studentID := database.Text("id")
		studyPlanIDs := database.TextArray([]string{})
		offset := database.Timestamptz(time.Now())
		q := &ListStudentAvailableContentsArgs{
			StudentID:    studentID,
			StudyPlanIDs: studyPlanIDs,
			Offset:       offset,
		}
		mockDB.MockQueryArgs(
			t,
			nil,
			mock.Anything,
			mock.AnythingOfType("string"),
			&q.StudentID,
			&q.StudyPlanIDs,
			&q.Offset,
			&q.BookID,
			&q.ChapterID,
			&q.TopicID,
			&q.CourseID,
		)

		e := &entities.StudyPlanItem{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		contents, err := r.ListStudentAvailableContents(ctx, mockDB.DB, q)

		assert.Nil(t, err)
		assert.Equal(t, []*entities.StudyPlanItem{e}, contents)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})
}

func TestListStudyPlans(t *testing.T) {
	t.Run("err select", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		studentID := database.Text("id")
		courseID := database.Text("cid")
		schoolID := database.Int4(constant.ManabieSchool)

		limit := uint32(5)
		offset := database.Text("offset")

		r, mockDB := studentStudyPlanRepoWithMock()
		mockDB.MockQueryArgs(
			t,
			puddle.ErrClosedPool,
			mock.Anything,
			mock.AnythingOfType("string"),
			&studentID,
			&courseID,
			&schoolID,
			&offset,
			&limit,
		)

		plans, err := r.ListStudyPlans(ctx, mockDB.DB, &ListStudyPlansArgs{
			StudentID: studentID,
			CourseID:  courseID,
			SchoolID:  schoolID,
			Limit:     limit,
			Offset:    offset,
		})

		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, plans)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		studentID := database.Text("id")
		courseID := database.Text("cid")
		schoolID := database.Int4(constant.ManabieSchool)

		limit := uint32(5)
		offset := database.Text("offset")

		r, mockDB := studentStudyPlanRepoWithMock()
		mockDB.MockQueryArgs(
			t,
			nil,
			mock.Anything,
			mock.AnythingOfType("string"),
			&studentID,
			&courseID,
			&schoolID,
			&offset,
			&limit,
		)

		e := &entities.StudyPlan{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		plans, err := r.ListStudyPlans(ctx, mockDB.DB, &ListStudyPlansArgs{
			StudentID: studentID,
			CourseID:  courseID,
			SchoolID:  schoolID,
			Limit:     limit,
			Offset:    offset,
		})

		assert.Nil(t, err)
		assert.Equal(t, []*entities.StudyPlan{e}, plans)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})

	return
}

func TestStudentStudyPlanItem(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	studentStudyPlanRepo := &StudentStudyPlanRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.StudentStudyPlan{
				{
					StudentID:   database.Text("student-id"),
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
			req: []*entities.StudentStudyPlan{
				{
					StudentID:   database.Text("student-id-2"),
					StudyPlanID: database.Text("study-plan-id-2"),
				},
				{
					StudentID:   database.Text("student-id-3"),
					StudyPlanID: database.Text("study-plan-id-3"),
				},
				{
					StudentID:   database.Text("student-id-4"),
					StudyPlanID: database.Text("study-plan-id-4"),
				},
				{
					StudentID:   database.Text("student-id-5"),
					StudyPlanID: database.Text("study-plan-id-5"),
				},
			},
			expectedErr: fmt.Errorf("batchResults.Exec: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, pgx.ErrTxClosed)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := studentStudyPlanRepo.BulkUpsert(ctx, db, testCase.req.([]*entities.StudentStudyPlan))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}

	return
}

func TestListActiveStudyPlanItems(t *testing.T) {
	t.Parallel()
	t.Run("err select", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		studentID := database.Text("id")

		now := timeutil.Now().UTC()

		pgNow := database.Timestamptz(now)

		limit := uint32(5)
		offset := database.Timestamptz(now)

		includeComplete := false

		r, mockDB := studentStudyPlanRepoWithMock()
		mockDB.MockQueryArgs(
			t,
			puddle.ErrClosedPool,
			mock.Anything,
			mock.AnythingOfType("string"),
			&studentID,
			&offset,
			&pgNow,
			&pgtype.TextArray{Status: pgtype.Null},
			&pgtype.Text{Status: pgtype.Null},
			&includeComplete,
			&pgtype.Int4{Status: pgtype.Null},
			&pgtype.Text{Status: pgtype.Null},
			&limit,
		)

		plans, err := r.ListActiveStudyPlanItems(ctx, mockDB.DB, &ListStudyPlanItemsArgs{
			StudentID:        studentID,
			Offset:           offset,
			Limit:            limit,
			Now:              pgNow,
			CourseIDs:        pgtype.TextArray{Status: pgtype.Null},
			StudyPlanID:      pgtype.Text{Status: pgtype.Null},
			StudyPlanItemID:  pgtype.Text{Status: pgtype.Null},
			IncludeCompleted: includeComplete,
			DisplayOrder:     pgtype.Int4{Status: pgtype.Null},
		})

		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, plans)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		studentID := database.Text("id")

		now := timeutil.Now().UTC()

		pgNow := database.Timestamptz(now)

		limit := uint32(5)
		offset := database.Timestamptz(now)

		includeComplete := false

		r, mockDB := studentStudyPlanRepoWithMock()
		mockDB.MockQueryArgs(
			t,
			nil,
			mock.Anything,
			mock.AnythingOfType("string"),
			&studentID,
			&offset,
			&pgNow,
			&pgtype.TextArray{Status: pgtype.Null},
			&pgtype.Text{Status: pgtype.Null},
			&includeComplete,
			&pgtype.Int4{Status: pgtype.Null},
			&pgtype.Text{Status: pgtype.Null},
			&limit,
		)

		e := &entities.StudyPlanItem{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		plans, err := r.ListActiveStudyPlanItems(ctx, mockDB.DB, &ListStudyPlanItemsArgs{
			StudentID:        studentID,
			Offset:           offset,
			Limit:            limit,
			Now:              pgNow,
			CourseIDs:        pgtype.TextArray{Status: pgtype.Null},
			StudyPlanID:      pgtype.Text{Status: pgtype.Null},
			StudyPlanItemID:  pgtype.Text{Status: pgtype.Null},
			IncludeCompleted: includeComplete,
			DisplayOrder:     pgtype.Int4{Status: pgtype.Null},
		})

		assert.Nil(t, err)
		assert.Equal(t, []*entities.StudyPlanItem{e}, plans)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})
}

func TestListCompletedStudyPlanItems(t *testing.T) {
	t.Parallel()
	t.Run("err select", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		studentID := database.Text("id")

		now := timeutil.Now().UTC()

		pgNow := database.Timestamptz(now)

		limit := uint32(5)
		offset := database.Timestamptz(now)
		courseIds := database.TextArray([]string{})
		r, mockDB := studentStudyPlanRepoWithMock()
		mockDB.MockQueryArgs(
			t,
			puddle.ErrClosedPool,
			mock.Anything,
			mock.AnythingOfType("string"),
			&studentID,
			&offset,
			&pgNow,
			&pgtype.Int4{Status: pgtype.Null},
			&courseIds,
			&pgtype.Text{Status: pgtype.Null},
			&limit,
		)

		plans, err := r.ListCompletedStudyPlanItems(ctx, mockDB.DB, &ListStudyPlanItemsArgs{
			StudentID:       studentID,
			Offset:          offset,
			Limit:           limit,
			StudyPlanItemID: pgtype.Text{Status: pgtype.Null},
			Now:             pgNow,
			CourseIDs:       courseIds,
			StudyPlanID:     pgtype.Text{Status: pgtype.Null},
			DisplayOrder:    pgtype.Int4{Status: pgtype.Null},
		})

		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, plans)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		studentID := database.Text("id")

		now := timeutil.Now().UTC()

		pgNow := database.Timestamptz(now)

		limit := uint32(5)
		courseIds := database.TextArray([]string{})

		offset := database.Timestamptz(now)

		displayOrder := database.Int4(0)

		r, mockDB := studentStudyPlanRepoWithMock()
		mockDB.MockQueryArgs(
			t,
			nil,
			mock.Anything,
			mock.AnythingOfType("string"),
			&studentID,
			&offset,
			&pgNow,
			&displayOrder,
			&courseIds,
			&pgtype.Text{Status: pgtype.Null},
			&limit,
		)

		e := &entities.StudyPlanItem{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		plans, err := r.ListCompletedStudyPlanItems(ctx, mockDB.DB, &ListStudyPlanItemsArgs{
			StudentID:       studentID,
			Offset:          offset,
			Limit:           limit,
			StudyPlanItemID: pgtype.Text{Status: pgtype.Null},
			Now:             pgNow,
			CourseIDs:       courseIds,
			DisplayOrder:    displayOrder,
		})

		assert.Nil(t, err)
		assert.Equal(t, []*entities.StudyPlanItem{e}, plans)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})
}

func TestListOverdueStudyPlanItems(t *testing.T) {
	t.Parallel()
	t.Run("err select", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		studentID := database.Text("id")

		now := timeutil.Now().UTC()

		pgNow := database.Timestamptz(now)

		limit := uint32(5)
		offset := database.Timestamptz(now)
		courseIds := database.TextArray([]string{})

		r, mockDB := studentStudyPlanRepoWithMock()
		mockDB.MockQueryArgs(
			t,
			puddle.ErrClosedPool,
			mock.Anything,
			mock.AnythingOfType("string"),
			&studentID,
			&offset,
			&pgNow,
			&pgtype.Text{Status: pgtype.Null},
			&courseIds,
			&limit,
		)

		plans, err := r.ListOverdueStudyPlanItems(ctx, mockDB.DB, &ListStudyPlanItemsArgs{
			StudentID:       studentID,
			Offset:          offset,
			Limit:           limit,
			StudyPlanItemID: pgtype.Text{Status: pgtype.Null},
			Now:             pgNow,
			CourseIDs:       courseIds,
		})

		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, plans)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		studentID := database.Text("id")

		now := timeutil.Now().UTC()

		pgNow := database.Timestamptz(now)
		courseIds := database.TextArray([]string{})

		limit := uint32(5)
		offset := database.Timestamptz(now)

		r, mockDB := studentStudyPlanRepoWithMock()
		mockDB.MockQueryArgs(
			t,
			nil,
			mock.Anything,
			mock.AnythingOfType("string"),
			&studentID,
			&offset,
			&pgNow,
			&pgtype.Text{Status: pgtype.Null},
			&courseIds,
			&limit,
		)

		e := &entities.StudyPlanItem{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		plans, err := r.ListOverdueStudyPlanItems(ctx, mockDB.DB, &ListStudyPlanItemsArgs{
			StudentID:       studentID,
			Offset:          offset,
			Limit:           limit,
			StudyPlanItemID: pgtype.Text{Status: pgtype.Null},
			Now:             pgNow,
			CourseIDs:       courseIds,
		})

		assert.Nil(t, err)
		assert.Equal(t, []*entities.StudyPlanItem{e}, plans)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})
}

func TestListStudyPlanItems(t *testing.T) {
	t.Parallel()
	t.Run("err select", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		studentID := database.Text("id")

		now := timeutil.Now().UTC()

		pgNow := database.Timestamptz(now)

		limit := uint32(5)
		displayorder := database.Int4(0)
		studyplanID := database.Text("study-plan-id")
		offset := database.Timestamptz(now)
		courseIds := database.TextArray([]string{})

		r, mockDB := studentStudyPlanRepoWithMock()
		mockDB.MockQueryArgs(
			t,
			puddle.ErrClosedPool,
			mock.Anything,
			mock.AnythingOfType("string"),
			&studentID,
			&displayorder,
			&pgNow,
			&pgtype.Text{Status: pgtype.Null},
			&studyplanID,
			&courseIds,
			&limit,
		)

		plans, err := r.ListStudyPlanItems(ctx, mockDB.DB, &ListStudyPlanItemsArgs{
			StudentID:       studentID,
			Offset:          offset,
			Limit:           limit,
			StudyPlanItemID: pgtype.Text{Status: pgtype.Null},
			Now:             pgNow,
			CourseIDs:       courseIds,
			DisplayOrder:    displayorder,
			StudyPlanID:     studyplanID,
		})

		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, plans)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		studentID := database.Text("id")

		now := timeutil.Now().UTC()

		pgNow := database.Timestamptz(now)
		courseIds := database.TextArray([]string{})

		limit := uint32(5)
		displayorder := database.Int4(0)
		studyplanID := database.Text("study-plan-id")
		offset := database.Timestamptz(now)

		r, mockDB := studentStudyPlanRepoWithMock()
		mockDB.MockQueryArgs(
			t,
			nil,
			mock.Anything,
			mock.AnythingOfType("string"),
			&studentID,
			&displayorder,
			&pgNow,
			&pgtype.Text{Status: pgtype.Null},
			&studyplanID,
			&courseIds,
			&limit)

		e := &entities.StudyPlanItem{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		plans, err := r.ListStudyPlanItems(ctx, mockDB.DB, &ListStudyPlanItemsArgs{
			StudentID:       studentID,
			Offset:          offset,
			Limit:           limit,
			StudyPlanItemID: pgtype.Text{Status: pgtype.Null},
			Now:             pgNow,
			CourseIDs:       courseIds,
			DisplayOrder:    displayorder,
			StudyPlanID:     studyplanID,
		})

		assert.Nil(t, err)
		assert.Equal(t, []*entities.StudyPlanItem{e}, plans)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})
}

func TestBulkSoftDeleteStudentStudyPlan(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	studentStudyPlanRepo := &StudentStudyPlanRepo{}

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
		err := studentStudyPlanRepo.DeleteStudentStudyPlans(ctx, db, testCase.req.(pgtype.TextArray))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}

	return
}

func Test_RetrieveByStudentCourse(t *testing.T) {
	t.Parallel()
	type TestCase struct {
		name         string
		intput1      pgtype.TextArray
		intput2      pgtype.TextArray
		expectedResp interface{}
		expectedErr  error
		setup        func(ctx context.Context)
	}
	r, mockDB := studentStudyPlanRepoWithMock()
	studentIDs := database.TextArray([]string{"student-id-1", "student-id-2"})
	courseIDs := database.TextArray([]string{"course-id-1", "course-id-2"})
	testCases := []TestCase{
		{
			name:        "happy case",
			intput1:     studentIDs,
			intput2:     courseIDs,
			expectedErr: nil, //fmt.Errorf("err db.Query: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				e := &entities.StudentStudyPlan{}
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &studentIDs, &courseIDs)
				fields, values := e.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
		{
			name:        "err query",
			intput1:     studentIDs,
			intput2:     courseIDs,
			expectedErr: fmt.Errorf("err db.Query: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				e := &entities.StudentStudyPlan{}
				mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, &studentIDs, &courseIDs)
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
			_, err := r.RetrieveByStudentCourse(ctx, mockDB.DB, tc.intput1, tc.intput2)
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, tc.expectedErr, err)
			}
			if tc.expectedErr == nil {
				e := &entities.StudentStudyPlan{}
				fields, _ := e.FieldMap()
				mockDB.RawStmt.AssertSelectedFields(t, fields...)
			}
		})
	}
}
