package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserDeviceTokenRepo_UpsertUserDeviceToken(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}

	testCases := []struct {
		Name  string
		Ent   *entities.UserDeviceToken
		Err   error
		Setup func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Ent:  &entities.UserDeviceToken{},
			Err:  nil,
			Setup: func(ctx context.Context) {
				e := &entities.UserDeviceToken{}
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
			Name: "unique constrain",
			Ent:  &entities.UserDeviceToken{},
			Err:  errors.Wrap(&pgconn.PgError{Code: pgerrcode.UniqueViolation}, "r.DB.ExecEx"),
			Setup: func(ctx context.Context) {
				e := &entities.UserDeviceToken{}
				fields := database.GetFieldNames(e)
				values := database.GetScanFields(e, fields)
				mockValues := make([]interface{}, 0, len(values)+2)
				mockValues = append(mockValues, mock.Anything)
				mockValues = append(mockValues, mock.AnythingOfType("string"))
				for range values {
					mockValues = append(mockValues, mock.Anything)
				}
				db.On("Exec", mockValues...).Once().Return(pgconn.CommandTag([]byte(`0`)), &pgconn.PgError{Code: pgerrcode.UniqueViolation})
			},
		},
	}

	repo := &UserDeviceTokenRepo{}
	ctx := context.Background()

	for _, testcase := range testCases {
		t.Run(testcase.Name, func(t *testing.T) {
			testcase.Setup(ctx)
			err := repo.UpsertUserDeviceToken(ctx, db, testcase.Ent)
			if testcase.Err == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, testcase.Err.Error(), err.Error())
			}
		})
	}
}

func Test_FindByUserIDs(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	userIDs := []string{"user-id-1", "user-id-2"}

	userDeviceToken1 := &entities.UserDeviceToken{}
	userDeviceToken2 := &entities.UserDeviceToken{}
	database.AllNullEntity(userDeviceToken1)
	database.AllNullEntity(userDeviceToken2)
	testCases := []struct {
		Name    string
		UserIDs pgtype.TextArray
		Err     error
		Setup   func(ctx context.Context)
	}{
		{
			Name:    "happy case",
			UserIDs: database.TextArray(userIDs),
			Err:     nil,
			Setup: func(ctx context.Context) {
				fields, vals1 := userDeviceToken1.FieldMap()
				_, vals2 := userDeviceToken2.FieldMap()
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.TextArray(userIDs))
				mockDB.MockScanArray(nil, fields, [][]interface{}{
					vals1,
					vals2,
				})
			},
		},
		{
			Name:    "erro scan",
			UserIDs: database.TextArray(userIDs),
			Err:     fmt.Errorf("rows.Scan: %w", pgx.ErrNoRows),
			Setup: func(ctx context.Context) {
				fields, vals1 := userDeviceToken1.FieldMap()
				_, vals2 := userDeviceToken2.FieldMap()
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.TextArray(userIDs))
				mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{
					vals1,
					vals2,
				})
			},
		},
	}

	repo := &UserDeviceTokenRepo{}
	ctx := context.Background()
	for _, testcase := range testCases {
		t.Run(testcase.Name, func(t *testing.T) {
			testcase.Setup(ctx)
			res, err := repo.FindByUserIDs(ctx, mockDB.DB, testcase.UserIDs)
			if testcase.Err == nil {
				assert.Nil(t, err)
				assert.Equal(t, userDeviceToken1, res[0])
				assert.Equal(t, userDeviceToken2, res[1])
			} else {
				assert.Equal(t, testcase.Err.Error(), err.Error())
			}
		})
	}
}
