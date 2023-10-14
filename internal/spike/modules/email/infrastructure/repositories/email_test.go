package repositories

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/model"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEmailRepo_UpsertEmail(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	testCases := []struct {
		Name  string
		Ent   *model.Email
		Err   error
		SetUp func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Ent:  &model.Email{},
			SetUp: func(_ context.Context) {
				e := &model.Email{}
				fields := database.GetFieldNames(e)
				values := database.GetScanFields(e, fields)
				mockValues := make([]interface{}, 0, len(values)+2)
				ctx, span := interceptors.StartSpan(context.Background(), "EmailRepo.UpsertEmail")
				defer span.End()
				mockValues = append(mockValues, ctx)
				mockValues = append(mockValues, mock.AnythingOfType("string"))
				for range values {
					mockValues = append(mockValues, mock.Anything)
				}
				db.On("Exec", mockValues...).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
		},
	}

	repo := &EmailRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.UpsertEmail(ctx, db, testCase.Ent)
			if testCase.Err == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func TestEmailRepo_UpdateEmail(t *testing.T) {
	t.Parallel()
	db := testutil.NewMockDB()
	type args struct {
		ctx        context.Context
		db         *testutil.MockDB
		emailID    string
		attributes map[string]interface{}
	}

	sgMessageID := idutil.ULIDNow()
	emailIDSample := idutil.ULIDNow()

	tests := []struct {
		name     string
		args     args
		mockFunc func(bd *testutil.MockDB)
		wantErr  bool
	}{
		{
			name: "no update case",
			args: args{
				ctx:        context.Background(),
				db:         db,
				emailID:    emailIDSample,
				attributes: make(map[string]interface{}),
			},
			wantErr:  false,
			mockFunc: func(mockDB *testutil.MockDB) {},
		},
		{
			name: "update case",
			args: args{
				ctx:     context.Background(),
				db:      db,
				emailID: emailIDSample,
				attributes: map[string]interface{}{
					"status":        "EMAIL_STATUS_PROCESSED",
					"sg_message_id": sgMessageID,
				},
			},
			wantErr: false,
			mockFunc: func(mockDB *testutil.MockDB) {
				params := []interface{}{mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything}
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, params...)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &EmailRepo{}
			tt.mockFunc(tt.args.db)
			if e := r.UpdateEmail(tt.args.ctx, tt.args.db.DB, tt.args.emailID, tt.args.attributes); tt.wantErr && e != nil {
				t.Errorf("want error but have no")
			}
		})
	}
}
