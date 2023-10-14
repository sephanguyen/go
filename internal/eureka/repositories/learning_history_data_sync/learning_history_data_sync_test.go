package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgconn"
	entities "github.com/manabie-com/backend/internal/eureka/entities/learning_history_data_sync"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/stretchr/testify/mock"
	"gotest.tools/assert"
)

type TestCase struct {
	name         string
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestRetrieveMappingCourseID(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	rows := &mock_database.Rows{}
	repo := &LearningHistoryDataSyncRepo{}

	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "query error",
			expectedErr: fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveMappingCourseID.Query: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "scan error",
			expectedErr: fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveMappingCourseID.Scan: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
		{
			name:        "rows error",
			expectedErr: fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveMappingCourseID.Err: %w", fmt.Errorf("rows error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(fmt.Errorf("rows error"))
				rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := repo.RetrieveMappingCourseID(ctx, db)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		}
	}
}

func TestRetrieveMappingExamLoID(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	rows := &mock_database.Rows{}
	repo := &LearningHistoryDataSyncRepo{}

	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "query error",
			expectedErr: fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveMappingExamLoID.Query: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "scan error",
			expectedErr: fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveMappingExamLoID.Scan: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
		{
			name:        "rows error",
			expectedErr: fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveMappingExamLoID.Err: %w", fmt.Errorf("rows error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(fmt.Errorf("rows error"))
				rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := repo.RetrieveMappingExamLoID(ctx, db)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		}
	}
}

func TestRetrieveMappingQuestionTag(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	rows := &mock_database.Rows{}
	repo := &LearningHistoryDataSyncRepo{}

	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "query error",
			expectedErr: fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveMappingQuestionTag.Query: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "scan error",
			expectedErr: fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveMappingQuestionTag.Scan: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
		{
			name:        "rows error",
			expectedErr: fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveMappingQuestionTag.Err: %w", fmt.Errorf("rows error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(fmt.Errorf("rows error"))
				rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := repo.RetrieveMappingQuestionTag(ctx, db)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		}
	}
}

func TestRetrieveFailedSyncEmailRecipient(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	rows := &mock_database.Rows{}
	repo := &LearningHistoryDataSyncRepo{}

	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "query error",
			expectedErr: fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveFailedSyncEmailRecipient.Query: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "scan error",
			expectedErr: fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveFailedSyncEmailRecipient.Scan: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
		{
			name:        "rows error",
			expectedErr: fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveFailedSyncEmailRecipient.Err: %w", fmt.Errorf("rows error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(fmt.Errorf("rows error"))
				rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := repo.RetrieveFailedSyncEmailRecipient(ctx, db)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		}
	}
}

func TestBulkUpsertMappingCourseID(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	repo := &LearningHistoryDataSyncRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.MappingCourseID{
				{
					ManabieCourseID: database.Text("manabie-course-id-1"),
					WithusCourseID:  database.Text("withus-course-id-1"),
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockFields := []interface{}{mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything}
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mockFields...)
				db.On("Exec", args...).Once().Return(cmdTag, nil)
			},
		},
		{
			name: "error exec error",
			req: []*entities.MappingCourseID{
				{
					ManabieCourseID: database.Text("manabie-course-id-1"),
					WithusCourseID:  database.Text("withus-course-id-1"),
				},
			},
			expectedErr: fmt.Errorf("LearningHistoryDataSyncRepo.BulkUpsertMappingCourseID: error exec error"),
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockFields := []interface{}{mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything}
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mockFields...)
				db.On("Exec", args...).Once().Return(cmdTag, fmt.Errorf("error exec error"))
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := repo.BulkUpsertMappingCourseID(ctx, db, testCase.req.([]*entities.MappingCourseID))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, nil, err)
		}
	}
}

func TestBulkUpsertMappingExamLoID(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	repo := &LearningHistoryDataSyncRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.MappingExamLoID{
				{
					ExamLoID:     database.Text("exam-lo-id-1"),
					MaterialCode: database.Text("material-code-1"),
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockFields := []interface{}{mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything}
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mockFields...)
				db.On("Exec", args...).Once().Return(cmdTag, nil)
			},
		},
		{
			name: "error exec error",
			req: []*entities.MappingExamLoID{
				{
					ExamLoID:     database.Text("exam-lo-id-1"),
					MaterialCode: database.Text("material-code-1"),
				},
			},
			expectedErr: fmt.Errorf("LearningHistoryDataSyncRepo.BulkUpsertMappingExamLoID: error exec error"),
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockFields := []interface{}{mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything}
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mockFields...)
				db.On("Exec", args...).Once().Return(cmdTag, fmt.Errorf("error exec error"))
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := repo.BulkUpsertMappingExamLoID(ctx, db, testCase.req.([]*entities.MappingExamLoID))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, nil, err)
		}
	}
}

func TestBulkUpsertMappingQuestionTag(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	repo := &LearningHistoryDataSyncRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.MappingQuestionTag{
				{
					ManabieTagID:   database.Text("manabie-tag-id-1"),
					ManabieTagName: database.Text("manabie-tag-name-1"),
					WithusTagName:  database.Text("withus-tag-name-1"),
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockFields := []interface{}{mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything}
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mockFields...)
				db.On("Exec", args...).Once().Return(cmdTag, nil)
			},
		},
		{
			name: "error exec error",
			req: []*entities.MappingQuestionTag{
				{
					ManabieTagID:   database.Text("manabie-tag-id-1"),
					ManabieTagName: database.Text("manabie-tag-name-1"),
					WithusTagName:  database.Text("withus-tag-name-1"),
				},
			},
			expectedErr: fmt.Errorf("LearningHistoryDataSyncRepo.BulkUpsertMappingQuestionTag: error exec error"),
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockFields := []interface{}{mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything}
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mockFields...)
				db.On("Exec", args...).Once().Return(cmdTag, fmt.Errorf("error exec error"))
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := repo.BulkUpsertMappingQuestionTag(ctx, db, testCase.req.([]*entities.MappingQuestionTag))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, nil, err)
		}
	}
}

func TestBulkUpsertFailedSyncEmailRecipient(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	repo := &LearningHistoryDataSyncRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.FailedSyncEmailRecipient{
				{
					RecipientID:  database.Text("recipient-id-1"),
					EmailAddress: database.Text("email-address-1"),
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockFields := []interface{}{mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything}
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mockFields...)
				db.On("Exec", args...).Once().Return(cmdTag, nil)
			},
		},
		{
			name: "error exec error",
			req: []*entities.FailedSyncEmailRecipient{
				{
					RecipientID:  database.Text("recipient-id-1"),
					EmailAddress: database.Text("email-address-1"),
				},
			},
			expectedErr: fmt.Errorf("LearningHistoryDataSyncRepo.BulkUpsertFailedSyncEmailRecipient: error exec error"),
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockFields := []interface{}{mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything}
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mockFields...)
				db.On("Exec", args...).Once().Return(cmdTag, fmt.Errorf("error exec error"))
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := repo.BulkUpsertFailedSyncEmailRecipient(ctx, db, testCase.req.([]*entities.FailedSyncEmailRecipient))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, nil, err)
		}
	}
}

func TestRetrieveWithusData(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	rows := &mock_database.Rows{}
	repo := &LearningHistoryDataSyncRepo{}

	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "query error",
			expectedErr: fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveWithusData.Query: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "scan error",
			expectedErr: fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveWithusData.Scan: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
		{
			name:        "rows error",
			expectedErr: fmt.Errorf("LearningHistoryDataSyncRepo.RetrieveWithusData.Err: %w", fmt.Errorf("rows error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(fmt.Errorf("rows error"))
				rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := repo.RetrieveWithusData(ctx, db)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, nil, err)
		}
	}
}
