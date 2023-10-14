package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
	"github.com/pkg/errors"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func DomainTaggedUserRepoWithSqlMock() (*DomainTaggedUserRepo, *testutil.MockDB) {
	r := &DomainTaggedUserRepo{}
	return r, testutil.NewMockDB()
}

type mockDomainTaggedUser struct {
	userID field.String
	tagID  field.String
	entity.EmptyDomainTaggedUser
}

func createMockDomainTaggedUser(userID string, tagID string) entity.DomainTaggedUser {
	return &mockDomainTaggedUser{
		userID: field.NewString(userID),
		tagID:  field.NewString(tagID),
	}
}

func (m *mockDomainTaggedUser) UserID() field.String {
	return m.userID
}

func (m *mockDomainTaggedUser) TagID() field.String {
	return m.tagID
}

func TestDomainTaggedUserRepo_GetByUserID(t *testing.T) {
	ctx := auth.InjectFakeJwtToken(context.Background(), fmt.Sprint(constants.ManabieSchool))
	db := new(mock_database.QueryExecer)
	userIDs := []string{idutil.ULIDNow()}

	var attrs []interface{}
	fieldMap, _ := NewTaggedUser(&entity.EmptyDomainTaggedUser{}).FieldMap()
	for range fieldMap {
		attrs = append(attrs, mock.Anything)
	}

	tests := []struct {
		name    string
		userIDs []string
		wantErr error
		setup   func()
	}{
		{
			name:    "happy case",
			wantErr: nil,
			setup: func() {
				rows := &mock_database.Rows{}
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Times(len(userIDs)).Return(true)
				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return()
				rows.On("Err").Once().Return(nil)
				rows.On("Scan", attrs...).Times(len(userIDs)).Return(nil)
			},
		},
		{
			name:    "error: db.Query error",
			wantErr: InternalError{RawError: errors.Wrap(fmt.Errorf("error"), "db.Query")},
			setup: func() {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
		{
			name:    "error: rows.Err error",
			wantErr: InternalError{RawError: errors.Wrap(fmt.Errorf("error"), "rows.Err")},
			setup: func() {
				rows := &mock_database.Rows{}
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return()
				rows.On("Err").Once().Return(fmt.Errorf("error"))
			},
		},
		{
			name:    "error: rows.Scan error",
			wantErr: InternalError{RawError: errors.Wrap(fmt.Errorf("error"), "rows.Scan")},
			setup: func() {
				rows := &mock_database.Rows{}
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return()
				rows.On("Err").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", attrs...).Once().Return(fmt.Errorf("error"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ut := new(DomainTaggedUserRepo)
			if tt.setup != nil {
				tt.setup()
			}
			_, err := ut.GetByUserIDs(ctx, db, userIDs)
			if err != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.wantErr)
			}
		})
	}
}

func TestDomainTaggedUserRepo_SoftDelete(t *testing.T) {
	ctx := context.Background()
	db := new(mock_database.Ext)

	taggedUsers := []entity.DomainTaggedUser{
		NewTaggedUser(createMockDomainTaggedUser(idutil.ULIDNow(), idutil.ULIDNow())),
		NewTaggedUser(createMockDomainTaggedUser(idutil.ULIDNow(), idutil.ULIDNow())),
	}

	tests := []struct {
		name    string
		wantErr error
		setup   func()
	}{
		{
			"happy case",
			nil,
			func() {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			"error: db.Exec error",
			InternalError{RawError: errors.Wrap(fmt.Errorf("error"), "db.Exec")},
			func() {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ut := &DomainTaggedUserRepo{}
			if tt.setup != nil {
				tt.setup()
			}
			err := ut.SoftDelete(ctx, db, taggedUsers...)
			if err != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.wantErr)
			}
		})
	}
}

func TestDomainTaggedUserRepo_UpsertBatch(t *testing.T) {
	ctx := context.Background()
	db := new(mock_database.Ext)

	taggedUsers := []entity.DomainTaggedUser{
		NewTaggedUser(createMockDomainTaggedUser(idutil.ULIDNow(), idutil.ULIDNow())),
		NewTaggedUser(createMockDomainTaggedUser(idutil.ULIDNow(), idutil.ULIDNow())),
	}

	tests := []struct {
		name    string
		wantErr error
		setup   func()
	}{
		{
			name:    "happy case",
			wantErr: nil,
			setup: func() {
				fieldNames := database.GetFieldNames(new(Tag))
				batchResults := new(mock_database.BatchResults)

				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Times(len(fieldNames)).Return(nil, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name:    "error: batchResults.Exec error",
			wantErr: InternalError{RawError: errors.Wrap(fmt.Errorf("error"), "batchResults.Exec")},
			setup: func() {
				batchResults := new(mock_database.BatchResults)

				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, fmt.Errorf("error"))
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ut := &DomainTaggedUserRepo{}
			if tt.setup != nil {
				tt.setup()
			}
			err := ut.UpsertBatch(ctx, db, taggedUsers...)
			if err != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.wantErr)
			}
		})
	}
}

func TestDomainTaggedUserRepo_SoftDeleteByUserIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := DomainTaggedUserRepoWithSqlMock()
	userIDs := []string{"userID-1", "userID-2"}
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, database.TextArray(userIDs))

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
		},
		{
			name:        "err update",
			expectedErr: InternalError{RawError: errors.Wrap(puddle.ErrClosedPool, "db.Exec")},
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(ctx)
			}
			err := repo.SoftDeleteByUserIDs(ctx, mockDB.DB, userIDs)
			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.expectedErr)
				mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
				mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
					"user_id":    {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
					"deleted_at": {HasNullTest: true},
				})
			}
		})
	}
}
