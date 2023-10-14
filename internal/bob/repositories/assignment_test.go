package repositories

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/bob/entities"
	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestCase struct {
	name         string
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func AssignmentRepoWithSqlMock() (*AssignmentRepo, *testutil.MockDB) {
	r := &AssignmentRepo{}
	return r, testutil.NewMockDB()
}

func TestExecQueueAssignment_Batch(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	assignmentRepo := &AssignmentRepo{}
	// query := "INSERT INTO VALUES"
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities_bob.Assignment{
				{
					AssignmentID: pgtype.Text{String: "1", Status: pgtype.Present},
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
			req: []*entities_bob.Assignment{
				{
					AssignmentID: pgtype.Text{String: "1", Status: pgtype.Present},
				}, {
					AssignmentID: pgtype.Text{String: "2", Status: pgtype.Present},
				},
			},
			expectedErr: errors.New("batchResults.Exec: closed pool"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := assignmentRepo.ExecQueueAssignment(ctx, db, testCase.req.([]*entities_bob.Assignment))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestExecQueueStudentAssignment_Batch(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	assignmentRepo := &AssignmentRepo{}

	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities_bob.StudentAssignment{
				{
					AssignmentID: pgtype.Text{String: "1", Status: pgtype.Present},
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
			req: []*entities_bob.StudentAssignment{
				{
					AssignmentID: pgtype.Text{String: "1", Status: pgtype.Present},
				}, {
					AssignmentID: pgtype.Text{String: "2", Status: pgtype.Present},
				},
			},
			expectedErr: errors.New("batchResults.Exec: closed pool"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := assignmentRepo.ExecQueueStudentAssignment(ctx, db, testCase.req.([]*entities_bob.StudentAssignment))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestAssignmentRepo_FindAssignmentByIDs(t *testing.T) {
	t.Parallel()
	r, mockDB := AssignmentRepoWithSqlMock()

	ID := idutil.ULIDNow()
	assignmentIDs := database.TextArray([]string{ID})
	rows := mockDB.Rows

	assignment := &entities.Assignment{}
	fields, _ := assignment.FieldMap()
	scanFields := database.GetScanFields(assignment, fields)
	// scanFields = append(scanFields, &assignmentIDs)

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         assignmentIDs,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &assignmentIDs)
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
			req:         assignmentIDs,
			expectedErr: errors.Wrap(pgx.ErrNoRows, "rows.Err"),
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &assignmentIDs)
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
			_, err := r.FindAssignmentByIDs(ctx, mockDB.DB, tc.req.(pgtype.TextArray))
			if err != nil {
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			}
		})
	}
}

func TestAssignmentRepo_FindStudentAssignmentWithStudyPlan(t *testing.T) {
	t.Parallel()
	r, mockDB := AssignmentRepoWithSqlMock()

	ID := idutil.ULIDNow()
	studentID := database.Text(ID)
	rows := mockDB.Rows

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         studentID,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &studentID)
				mockDB.DB.On("Query").Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "err query",
			req:         studentID,
			expectedErr: errors.Wrap(pgx.ErrNoRows, "rows.Err"),
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &studentID)
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
			_, err := r.FindStudentAssignmentWithStudyPlan(ctx, mockDB.DB, tc.req.(pgtype.Text))
			if err != nil {
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			}
		})
	}
}

func TestAssignmentRepo_FindStudentOverdueAssignment(t *testing.T) {
	t.Parallel()
	r, mockDB := AssignmentRepoWithSqlMock()

	ID := idutil.ULIDNow()
	studentID := database.Text(ID)
	rows := mockDB.Rows

	a := &entities_bob.Assignment{}
	assignmentFields := database.GetFieldNames(a)
	topic := new(entities_bob.Topic)
	topicFields := database.GetFieldNames(topic)
	user := new(entities_bob.User)
	fields := database.GetScanFields(a, assignmentFields)
	fields = append(fields, database.GetScanFields(topic, topicFields)...)
	fields = append(fields, &user.ID, &user.LastName, &user.Group, &user.Avatar)

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         studentID,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &studentID)
				mockDB.DB.On("Query").Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", fields...).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "err query",
			req:         studentID,
			expectedErr: errors.Wrap(pgx.ErrNoRows, "rows.Err"),
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &studentID)
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
			_, err := r.FindStudentOverdueAssignment(ctx, mockDB.DB, tc.req.(pgtype.Text))
			if err != nil {
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			}
		})
	}
}

func TestAssignmentRepo_FindStudentCompletedAssignmentWeeklies(t *testing.T) {
	t.Parallel()
	r, mockDB := AssignmentRepoWithSqlMock()

	ID := idutil.ULIDNow()
	studentID := database.Text(ID)
	rows := mockDB.Rows

	studentAssignment := &entities.StudentAssignment{}
	a := &entities_bob.Assignment{}
	assignmentFields := database.GetFieldNames(a)
	topic := new(entities_bob.Topic)
	topicFields := database.GetFieldNames(topic)
	user := new(entities_bob.User)
	fields := database.GetScanFields(a, assignmentFields)
	fields = append(fields, database.GetScanFields(topic, topicFields)...)
	fields = append(fields, &user.ID, &user.LastName, &user.Group, &user.Avatar, &studentAssignment.CompletedAt)

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         studentID,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &studentID)
				mockDB.DB.On("Query").Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", fields...).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "err query",
			req:         studentID,
			expectedErr: errors.Wrap(pgx.ErrNoRows, "rows.Err"),
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &studentID)
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
			_, err := r.FindStudentCompletedAssignmentWeeklies(ctx, mockDB.DB, tc.req.(pgtype.Text), nil, nil)
			if err != nil {
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			}
		})
	}
}

func TestAssignmentRepo_RetrieveStudentAssignmentByTopic(t *testing.T) {
	t.Parallel()
	r, mockDB := AssignmentRepoWithSqlMock()

	rows := mockDB.Rows

	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockDB.DB.On("Query").Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "err query",
			expectedErr: errors.Wrap(pgx.ErrNoRows, "rows.Err"),
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
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
			_, err := r.RetrieveStudentAssignmentByTopic(ctx, mockDB.DB, database.Text("topic-id"), database.TextArray([]string{"topic-1", "topic-2"}))
			if err != nil {
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			}
		})
	}
}

func TestAssignmentRepo_FindByTopicID(t *testing.T) {
	t.Parallel()
	r, mockDB := AssignmentRepoWithSqlMock()

	ID := idutil.ULIDNow()
	topicID := database.Text(ID)
	rows := mockDB.Rows

	assignment := &entities.Assignment{}
	fields, _ := assignment.FieldMap()
	scanFields := database.GetScanFields(assignment, fields)
	// scanFields = append(scanFields, &assignmentIDs)

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         topicID,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &topicID)
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Scan", scanFields...).Once().Return(nil)
			},
		},
		{
			name:        "err query",
			req:         topicID,
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &topicID)
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Scan", scanFields...).Once().Return(pgx.ErrNoRows)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.setup(ctx)
			_, err := r.FindByTopicID(ctx, mockDB.DB, tc.req.(pgtype.Text))
			if err != nil {
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			}
		})
	}
}

func TestAssignmentRepo_RetrieveByTopicIDs(t *testing.T) {
	t.Parallel()
	r, mockDB := AssignmentRepoWithSqlMock()

	rows := mockDB.Rows

	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
				mockDB.DB.On("Query").Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "err query",
			expectedErr: errors.Wrap(pgx.ErrNoRows, "rows.Err"),
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
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
			_, err := r.RetrieveByTopicIDs(ctx, mockDB.DB, database.TextArray([]string{"topic-id"}))
			if err != nil {
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			}
		})
	}
}
