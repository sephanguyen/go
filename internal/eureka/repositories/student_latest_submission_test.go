package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStudentLatestSubmissionRepo_Upsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	studentLatestSubmissionRepo := &StudentLatestSubmissionRepo{}
	studentLatestSubmission := &entities.StudentLatestSubmission{
		StudentSubmission: entities.StudentSubmission{
			BaseEntity:        entities.BaseEntity{},
			ID:                pgtype.Text{},
			StudyPlanItemID:   pgtype.Text{},
			AssignmentID:      pgtype.Text{},
			StudentID:         pgtype.Text{},
			SubmissionContent: pgtype.JSONB{},
			CheckList:         pgtype.JSONB{},
			SubmissionGradeID: pgtype.Text{},
			Note:              pgtype.Text{},
			Status:            pgtype.Text{},
			EditorID:          pgtype.Text{},
			DeletedBy:         pgtype.Text{},
		},
	}
	fieldNames := database.GetFieldNames(studentLatestSubmission)
	scanFields := database.GetScanFields(studentLatestSubmission, fieldNames)
	args := []interface{}{mock.Anything, mock.AnythingOfType("string")}
	args = append(args, scanFields...)

	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Exec", args...).Once().Return(nil, nil)
			},
		},
		{
			name:        "error no rows",
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				db.On("Exec", args...).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := studentLatestSubmissionRepo.Upsert(ctx, db, studentLatestSubmission)
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestStudentLatestSubmissionRepo_UpsertV2(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	studentLatestSubmissionRepo := &StudentLatestSubmissionRepo{}
	studentLatestSubmission := &entities.StudentLatestSubmission{
		StudentSubmission: entities.StudentSubmission{
			BaseEntity:         entities.BaseEntity{},
			ID:                 pgtype.Text{},
			StudyPlanItemID:    pgtype.Text{},
			AssignmentID:       pgtype.Text{},
			StudentID:          pgtype.Text{},
			SubmissionContent:  pgtype.JSONB{},
			CheckList:          pgtype.JSONB{},
			SubmissionGradeID:  pgtype.Text{},
			Note:               pgtype.Text{},
			Status:             pgtype.Text{},
			EditorID:           pgtype.Text{},
			DeletedBy:          pgtype.Text{},
			StudyPlanID:        database.Text("study-plan-id"),
			LearningMaterialID: database.Text("lm-id"),
		},
	}
	fieldNames := database.GetFieldNames(studentLatestSubmission)
	scanFields := database.GetScanFields(studentLatestSubmission, fieldNames)
	args := []interface{}{mock.Anything, mock.AnythingOfType("string")}
	args = append(args, scanFields...)

	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Exec", args...).Once().Return(nil, nil)
			},
		},
		{
			name:        "error no rows",
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				db.On("Exec", args...).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := studentLatestSubmissionRepo.UpsertV2(ctx, db, studentLatestSubmission)
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestStudentLatestSubmissionRepo_DeleteByStudyPlanItemID(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	studentLatestSubmissionRepo := &StudentLatestSubmissionRepo{}

	studyPlanItemID := database.Text("study-plan-item-id")
	deletedBy := database.Text("deleted-by")

	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag(`1`)
				db.On("Exec", mock.Anything, mock.Anything, &deletedBy, &studyPlanItemID).Once().Return(cmdTag, nil)
			},
		},
		{
			name:        "error exec database",
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, &deletedBy, &studyPlanItemID).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:        "error no raw affected",
			expectedErr: fmt.Errorf("no raw affected, failed delete study plan item"),
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag(`0`)
				db.On("Exec", mock.Anything, mock.Anything, &deletedBy, &studyPlanItemID).Once().Return(cmdTag, nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := studentLatestSubmissionRepo.DeleteByStudyPlanItemID(ctx, db, studyPlanItemID, deletedBy)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}

	return
}
