package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/errorx"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func StudentParentRepoWithSqlMock() (*StudentParentRepo, *testutil.MockDB) {
	sp := &StudentParentRepo{}
	return sp, testutil.NewMockDB()
}

func TestStudentParentRepo_Upsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	studentParentRepo := &StudentParentRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entity.StudentParent{
				{
					StudentID:    pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ParentID:     pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					Relationship: pgtype.Text{String: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER.String(), Status: pgtype.Present},
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
			name: "happy case: upsert multiple parents",
			req: []*entity.StudentParent{
				{
					StudentID:    pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ParentID:     pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					Relationship: pgtype.Text{String: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER.String(), Status: pgtype.Present},
				},
				{
					StudentID:    pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ParentID:     pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					Relationship: pgtype.Text{String: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER.String(), Status: pgtype.Present},
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
			req: []*entity.StudentParent{
				{
					StudentID:    pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ParentID:     pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					Relationship: pgtype.Text{String: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER.String(), Status: pgtype.Present},
				},
			},
			expectedErr: errors.Wrap(puddle.ErrClosedPool, "batchResults.Exec"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := studentParentRepo.Upsert(ctx, db, testCase.req.([]*entity.StudentParent))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentParentRepo_FindParentIDsFromStudentID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := StudentParentRepoWithSqlMock()

	studentID := "id"

	testCases := []TestCase{
		{
			name:         "happy case",
			req:          studentID,
			expectedErr:  nil,
			expectedResp: []string{"id"},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, &studentID).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", mock.Anything).Once().Return(nil)
				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "fail case: database query fail",
			req:         studentID,
			expectedErr: errors.New("fail in database query"),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, &studentID).Once().Return(mockDB.Rows, errors.New("fail in database query"))
			},
		},
		{
			name:        "fail case: database rows Scan fail",
			req:         studentID,
			expectedErr: errors.New("fail in database rows Scan"),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, &studentID).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", mock.Anything).Once().Return(errors.New("fail in database rows Scan"))
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "fail case: database rows err fail",
			req:         studentID,
			expectedErr: errors.New("fail in database rows err"),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, &studentID).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", mock.Anything).Once().Return(nil)
				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(errors.New("fail in database rows err"))
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			testCase.setup(ctx)
			studentParents, err := r.FindParentIDsFromStudentID(ctx, mockDB.DB, testCase.req.(string))
			if err != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
				return
			}

			assert.Nil(t, err)
			expectedResp := testCase.expectedResp.([]string)
			assert.Equal(t, len(expectedResp), len(studentParents))

		})
	}
}

func TestStudentParentRepo_RemoveParent(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	studentParentRepo := &StudentParentRepo{}
	unRelationshipParentId := uuid.NewString()
	unRelationshipStudentId := uuid.NewString()
	testCases := []TestCase{
		{
			name: "happy case",
			req: &entity.StudentParent{
				StudentID:    pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				ParentID:     pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				Relationship: pgtype.Text{String: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER.String(), Status: pgtype.Present},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
			},
		},
		{
			name: "Some error when query to database",
			req: &entity.StudentParent{
				StudentID:    pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				ParentID:     pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				Relationship: pgtype.Text{String: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER.String(), Status: pgtype.Present},
			},
			expectedErr: errorx.ToStatusError(status.Error(codes.Internal, "something wrong")),
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, fmt.Errorf("something wrong"))
			},
		},
		{
			name: "Can't remove patent without relationship with student",
			req: &entity.StudentParent{
				StudentID:    pgtype.Text{String: unRelationshipStudentId, Status: pgtype.Present},
				ParentID:     pgtype.Text{String: unRelationshipParentId, Status: pgtype.Present},
				Relationship: pgtype.Text{String: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER.String(), Status: pgtype.Present},
			},
			expectedErr: errorx.ToStatusError(status.Error(codes.InvalidArgument, fmt.Sprintf("student with id %v don't have relationship with parent with id %v", unRelationshipStudentId, unRelationshipParentId))),
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`0`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.(*entity.StudentParent)
		err := studentParentRepo.RemoveParentFromStudent(ctx, db, req.ParentID, req.StudentID)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentParentRepo_UpsertParentAccessPathByStudentIDs(t *testing.T) {
	t.Parallel()
	repo, mockDB := StudentParentRepoWithSqlMock()
	studentID := uuid.NewString()
	testCases := []TestCase{
		{
			name:        "error send batch",
			req:         []string{studentID},
			expectedErr: errors.Wrap(pgx.ErrTxClosed, "batchResults.Exec"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, pgx.ErrTxClosed)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "happy case",
			req:  []string{studentID},
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := repo.UpsertParentAccessPathByStudentIDs(ctx, mockDB.DB, testCase.req.([]string))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentParentRepo_UpsertParentAccessPathByID(t *testing.T) {
	t.Parallel()
	repo, mockDB := StudentParentRepoWithSqlMock()
	parentID := uuid.NewString()
	testCases := []TestCase{
		{
			name:        "error send batch",
			req:         []string{parentID},
			expectedErr: errors.Wrap(pgx.ErrTxClosed, "batchResults.Exec"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, pgx.ErrTxClosed)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "happy case",
			req:  []string{parentID},
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := repo.UpsertParentAccessPathByID(ctx, mockDB.DB, testCase.req.([]string))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentParentRepo_FindStudentParentsByParentID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, studentParentValues := new(entity.StudentParent).FieldMap()
	argsStudentParent := append([]interface{}{}, genSliceMock(len(studentParentValues))...)

	r, mockDB := StudentParentRepoWithSqlMock()

	parentID := "parentID"

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         parentID,
			expectedErr: nil,
			expectedResp: []*entity.StudentParent{
				{
					StudentID: database.Text("StudentID"),
					ParentID:  database.Text(parentID),
				},
			},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, &parentID).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsStudentParent...).Once().Return(nil)
				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "fail case: database query fail",
			req:         parentID,
			expectedErr: errors.New("fail in database query"),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, &parentID).Once().Return(mockDB.Rows, errors.New("fail in database query"))
			},
		},
		{
			name:        "fail case: database rows Scan fail",
			req:         parentID,
			expectedErr: errors.New("fail in database rows Scan"),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, &parentID).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsStudentParent...).Once().Return(errors.New("fail in database rows Scan"))
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "fail case: database rows err fail",
			req:         parentID,
			expectedErr: errors.New("fail in database rows err"),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, &parentID).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsStudentParent...).Once().Return(nil)
				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(errors.New("fail in database rows err"))
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			testCase.setup(ctx)
			studentParents, err := r.FindStudentParentsByParentID(ctx, mockDB.DB, testCase.req.(string))
			if err != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
				return
			}

			assert.Nil(t, err)
			expectedResp := testCase.expectedResp.([]*entity.StudentParent)
			assert.Equal(t, len(expectedResp), len(studentParents))

		})
	}
}
