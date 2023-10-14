package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func UserAccessPathRepoWithSqlMock() (*UserAccessPathRepo, *testutil.MockDB) {
	sp := &UserAccessPathRepo{}
	return sp, testutil.NewMockDB()
}

func TestUserAccessPath_Upsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	userAccessPathRepo := &UserAccessPathRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entity.UserAccessPath{
				{
					UserID:       pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					LocationID:   pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
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
			name: "happy case: upsert multiple user_access_paths",
			req: []*entity.UserAccessPath{
				{
					UserID:       pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					LocationID:   pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
				},
				{
					UserID:       pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					LocationID:   pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
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
			req: []*entity.UserAccessPath{
				{
					UserID:       pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					LocationID:   pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
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
		err := userAccessPathRepo.Upsert(ctx, db, testCase.req.([]*entity.UserAccessPath))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestUserAccessPath_FindLocationIDsFromUserID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := UserAccessPathRepoWithSqlMock()

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
			studentParents, err := r.FindLocationIDsFromUserID(ctx, mockDB.DB, testCase.req.(string))
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

func TestUserAccessPath_Delete(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	userAccessPathRepo := &UserAccessPathRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []string{
				"user_id_1",
				"user_id_2",
			},
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`2`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
			},
		},
		{
			name: "user id empty",
			req:  []string{},
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`0`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
			},
			expectedErr: fmt.Errorf("cannot delete user_access_path"),
		},
		{
			name: "execute fail",
			req:  []string{},
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`0`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, pgx.ErrTxClosed)
			},
			expectedErr: pgx.ErrTxClosed,
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)

		err := userAccessPathRepo.Delete(ctx, db, database.TextArray(testCase.req.([]string)))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
