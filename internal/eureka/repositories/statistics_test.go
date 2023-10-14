package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetStudentTopicProgress(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	rows := &mock_database.Rows{}
	repo := &StatisticsRepo{}

	type Req struct {
		studyPlanID pgtype.Text
		studentID   pgtype.Text
	}

	testCases := []TestCase{
		{
			name: "happy case",
			req: &Req{
				studyPlanID: database.Text("study_plan_id"),
				studentID:   database.Text("student_id"),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name: "query error",
			req: &Req{
				studyPlanID: database.Text("study_plan_id"),
				studentID:   database.Text("student_id"),
			},
			expectedErr: fmt.Errorf("StatisticsRepo.GetStudentTopicProgress.Query: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			name: "scan error",
			req: &Req{
				studyPlanID: database.Text("study_plan_id"),
				studentID:   database.Text("student_id"),
			},
			expectedErr: fmt.Errorf("StatisticsRepo.GetStudentTopicProgress.Scan: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
		{
			name: "rows error",
			req: &Req{
				studyPlanID: database.Text("study_plan_id"),
				studentID:   database.Text("student_id"),
			},
			expectedErr: fmt.Errorf("StatisticsRepo.GetStudentTopicProgress.Err: %w", fmt.Errorf("rows error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(fmt.Errorf("rows error"))
				rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.(*Req)
		_, err := repo.GetStudentTopicProgress(ctx, db, req.studyPlanID, req.studentID)
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestGetStudentChapterProgress(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	rows := &mock_database.Rows{}
	repo := &StatisticsRepo{}

	type Req struct {
		studyPlanID pgtype.Text
		studentID   pgtype.Text
	}

	testCases := []TestCase{
		{
			name: "happy case",
			req: &Req{
				studyPlanID: database.Text("study_plan_id"),
				studentID:   database.Text("student_id"),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name: "query error",
			req: &Req{
				studyPlanID: database.Text("study_plan_id"),
				studentID:   database.Text("student_id"),
			},
			expectedErr: fmt.Errorf("StatisticsRepo.GetStudentChapterProgress.Query: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			name: "scan error",
			req: &Req{
				studyPlanID: database.Text("study_plan_id"),
				studentID:   database.Text("student_id"),
			},
			expectedErr: fmt.Errorf("StatisticsRepo.GetStudentChapterProgress.Scan: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
		{
			name: "rows error",
			req: &Req{
				studyPlanID: database.Text("study_plan_id"),
				studentID:   database.Text("student_id"),
			},
			expectedErr: fmt.Errorf("StatisticsRepo.GetStudentChapterProgress.Err: %w", fmt.Errorf("rows error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(fmt.Errorf("rows error"))
				rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.(*Req)
		_, err := repo.GetStudentChapterProgress(ctx, db, req.studyPlanID, req.studentID)
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestGetStudentProgress(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	bookTreeRows := &mock_database.Rows{}
	topicRows := &mock_database.Rows{}
	chapterRows := &mock_database.Rows{}
	repo := &StatisticsRepo{}

	type Req struct {
		studyPlanID pgtype.Text
		studentID   pgtype.Text
		courseID    pgtype.Text
	}

	req := &Req{
		studyPlanID: database.Text("study_plan_id"),
		studentID:   database.Text("student_id"),
		courseID:    database.Text("course_id"),
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         req,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmdTag, nil)
				db.On("Query", mock.Anything, mock.Anything).Once().Return(bookTreeRows, nil)
				bookTreeRows.On("Close").Once().Return(nil)
				bookTreeRows.On("Next").Once().Return(true)

				bookTreeRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				bookTreeRows.On("Next").Once().Return(false)
				bookTreeRows.On("Err").Once().Return(nil)

				db.On("Query", mock.Anything, mock.Anything).Once().Return(topicRows, nil)
				topicRows.On("Close").Once().Return(nil)
				topicRows.On("Next").Once().Return(true)

				topicRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				topicRows.On("Next").Once().Return(false)
				topicRows.On("Err").Once().Return(nil)

				db.On("Query", mock.Anything, mock.Anything).Once().Return(chapterRows, nil)
				chapterRows.On("Close").Once().Return(nil)
				chapterRows.On("Next").Once().Return(true)

				chapterRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				chapterRows.On("Next").Once().Return(false)
				chapterRows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "query error",
			req:         req,
			expectedErr: fmt.Errorf("StatisticsRepo.GetStudentProgress.BookTreeQuery: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmdTag, nil)
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(bookTreeRows, fmt.Errorf("error"))
			},
		},
		{
			name:        "scan error",
			req:         req,
			expectedErr: fmt.Errorf("StatisticsRepo.GetStudentProgress.BookTreeScan: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmdTag, nil)
				db.On("Query", mock.Anything, mock.Anything).Once().Return(bookTreeRows, nil)
				bookTreeRows.On("Close").Once().Return(nil)
				bookTreeRows.On("Next").Once().Return(true)

				bookTreeRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
		{
			name:        "rows error",
			req:         req,
			expectedErr: fmt.Errorf("StatisticsRepo.GetStudentProgress.BookTreeRowsErr: %w", fmt.Errorf("rows error")),
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmdTag, nil)
				db.On("Query", mock.Anything, mock.Anything).Once().Return(bookTreeRows, nil)
				bookTreeRows.On("Next").Once().Return(false)
				bookTreeRows.On("Err").Once().Return(fmt.Errorf("rows error"))
				bookTreeRows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.(*Req)
		_, _, _, err := repo.GetStudentProgress(ctx, db, req.studyPlanID, req.studentID, req.courseID)
		assert.Equal(t, testCase.expectedErr, err)
	}
}
