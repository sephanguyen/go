package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/domain/models"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAgoraUserRepo_Create(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	testCases := []struct {
		Name  string
		Ent   *models.AgoraUser
		Err   error
		SetUp func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Ent: &models.AgoraUser{
				UserID: database.Text(idutil.ULIDNow()),
			},
			SetUp: func(ctx context.Context) {
				e := &models.AgoraUser{}
				fields := database.GetFieldNames(e)
				values := database.GetScanFields(e, fields)
				mockValues := make([]interface{}, 0, len(values)+2)
				mockValues = append(mockValues, mock.Anything)
				mockValues = append(mockValues, mock.AnythingOfType("string"))
				for range values {
					mockValues = append(mockValues, mock.Anything)
				}
				db.On("Exec", mockValues...).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
		},
		{
			Name: "missing manabie_user_id",
			Ent:  &models.AgoraUser{},
			SetUp: func(ctx context.Context) {
				e := &models.AgoraUser{}
				fields := database.GetFieldNames(e)
				values := database.GetScanFields(e, fields)
				mockValues := make([]interface{}, 0, len(values)+2)
				mockValues = append(mockValues, ctx)
				mockValues = append(mockValues, mock.AnythingOfType("string"))
				for range values {
					mockValues = append(mockValues, mock.Anything)
				}
				db.On("Exec", mockValues...).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
			Err: fmt.Errorf("missing manabie_user_id when creating agora user"),
		},
	}

	repo := &AgoraUserRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.Create(ctx, db, testCase.Ent)
			if testCase.Err == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func TestAgoraUserRepo_CreateAgoraUserFailure(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	testCases := []struct {
		Name  string
		Ent   *models.AgoraUserFailure
		Err   error
		SetUp func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Ent: &models.AgoraUserFailure{
				UserID: database.Text(idutil.ULIDNow()),
			},
			SetUp: func(ctx context.Context) {
				e := &models.AgoraUserFailure{}
				fields := database.GetFieldNames(e)
				values := database.GetScanFields(e, fields)
				mockValues := make([]interface{}, 0, len(values)+2)
				mockValues = append(mockValues, mock.Anything)
				mockValues = append(mockValues, mock.AnythingOfType("string"))
				for range values {
					mockValues = append(mockValues, mock.Anything)
				}
				db.On("Exec", mockValues...).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
		},
		{
			Name: "missing manabie_user_id",
			Ent:  &models.AgoraUserFailure{},
			SetUp: func(ctx context.Context) {
				e := &models.AgoraUserFailure{}
				fields := database.GetFieldNames(e)
				values := database.GetScanFields(e, fields)
				mockValues := make([]interface{}, 0, len(values)+2)
				mockValues = append(mockValues, ctx)
				mockValues = append(mockValues, mock.AnythingOfType("string"))
				for range values {
					mockValues = append(mockValues, mock.Anything)
				}
				db.On("Exec", mockValues...).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
			Err: fmt.Errorf("missing manabie_user_id when creating agora user"),
		},
	}

	repo := &AgoraUserRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.CreateAgoraUserFailure(ctx, db, testCase.Ent)
			if testCase.Err == nil {
				assert.Nil(t, err)
			}
		})
	}
}
