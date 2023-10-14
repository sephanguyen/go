package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
)

func TestAllocateMarkerRepo_BulkUpsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	allocateMarkerRepo := &AllocateMarkerRepo{}
	testCases := []TestCase{
		{
			name: "happy Case",
			req: []*entities.AllocateMarker{
				&entities.AllocateMarker{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag(`1`)
				db.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
			},
		},
		{
			name: "error exec error",
			req: []*entities.AllocateMarker{
				&entities.AllocateMarker{},
			},
			expectedErr: fmt.Errorf("database.BulkUpsert error: error exec error"),
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag(`1`)
				db.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, fmt.Errorf("error exec error"))
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := allocateMarkerRepo.BulkUpsert(ctx, db, testCase.req.([]*entities.AllocateMarker))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, nil, err)
		}
	}
}

func TestExamLOSubmission_GetTeacherID(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &AllocateMarkerRepo{}
	args := &StudyPlanItemIdentity{
		StudentID:          database.Text("student_id"),
		StudyPlanID:        database.Text("study_plan_id"),
		LearningMaterialID: database.Text("learning_material_id"),
	}
	result := pgtype.Text{String: "", Status: pgtype.Present}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, args.StudentID, args.StudyPlanID, args.LearningMaterialID)
				mockDB.MockScanFields(nil, []string{"teacher_id"}, []interface{}{&result})
			},
			req:          args,
			expectedResp: result,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, args.StudentID, args.StudyPlanID, args.LearningMaterialID)
				mockDB.MockScanFields(pgx.ErrNoRows, []string{"teacher_id"}, []interface{}{&result})
			},
			req:          args,
			expectedResp: pgtype.Text{Status: pgtype.Undefined},
			expectedErr:  pgx.ErrNoRows,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.GetTeacherID(ctx, mockDB.DB, testCase.req.(*StudyPlanItemIdentity))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestAllocateMarkerRepo_ListAllocateTeacher(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &AllocateMarkerRepo{}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
			req: database.TextArray([]string{"course-id-1"}),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.GetAllocateTeacherByCourseAccess(ctx, mockDB.DB, testCase.req.(pgtype.TextArray))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp.([]*entities.AllocateTeacherItem), resp)
			}
		})
	}
}
