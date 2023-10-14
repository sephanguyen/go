package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestCase struct {
	name         string
	req          interface{}
	expectedErr  error
	expectedResp interface{}
	setup        func(ctx context.Context)
}

func UserRepoWithSqlMock() (*UserRepo, *testutil.MockDB) {
	r := &UserRepo{}
	return r, testutil.NewMockDB()
}

func TestUserRepo_Retrieve(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := UserRepoWithSqlMock()
	ids := pgtype.TextArray{}
	_ = ids.Set([]string{"id"})
	pgTextNull := pgtype.Text{Status: pgtype.Null}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything,
			mock.AnythingOfType("string"),
			&ids.Elements,
			&pgTextNull,
			&pgTextNull,
			&pgTextNull,
		)

		users, err := r.Retrieve(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, users)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			&ids.Elements,
			&pgTextNull,
			&pgTextNull,
			&pgTextNull,
		)

		e := &entity.LegacyUser{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		_ = e.FullName.Set(ksuid.New().String())
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		users, err := r.Retrieve(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.Equal(t, []*entity.LegacyUser{e}, users)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"user_id":      {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"email":        {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 2}},
			"phone_number": {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 3}},
			"user_group":   {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 4}},
			"deleted_at":   {HasNullTest: true},
		})
	})

	t.Run("success with select field user_id", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			&ids.Elements,
			&pgTextNull,
			&pgTextNull,
			&pgTextNull,
		)

		e := &entity.LegacyUser{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		_ = e.FullName.Set(ksuid.New().String())
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		users, err := r.Retrieve(ctx, mockDB.DB, ids, "id", "name")
		assert.Nil(t, err)
		assert.Equal(t, []*entity.LegacyUser{e}, users)

		mockDB.RawStmt.AssertSelectedFields(t, "id", "name")
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"user_id":      {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"email":        {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 2}},
			"phone_number": {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 3}},
			"user_group":   {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 4}},
			"deleted_at":   {HasNullTest: true},
		})
	})
}

func TestUserRepo_UserGroup(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := UserRepoWithSqlMock()
	ids := pgtype.TextArray{}
	_ = ids.Set([]string{"id"})
	pgTextNull := pgtype.Text{Status: pgtype.Null}

	t.Run("success get user_group", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			&ids.Elements,
			&pgTextNull,
			&pgTextNull,
			&pgTextNull,
		)

		e := &entity.LegacyUser{}

		_ = e.Group.Set("USER_GROUP_STUDENT")
		mockDB.MockScanArray(nil, []string{"user_group"}, [][]interface{}{
			{&e.Group},
		})

		group, err := r.UserGroup(ctx, mockDB.DB, pgtype.Text{String: "id", Status: pgtype.Present})
		assert.Nil(t, err)
		assert.Equal(t, "USER_GROUP_STUDENT", group)

		mockDB.RawStmt.AssertSelectedFields(t, "user_group")
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"user_id":      {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"email":        {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 2}},
			"phone_number": {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 3}},
			"user_group":   {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 4}},
			"deleted_at":   {HasNullTest: true},
		})
	})
}

func TestUserRepo_ResourcePath(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := UserRepoWithSqlMock()
	ids := pgtype.TextArray{}
	_ = ids.Set([]string{"id"})
	pgTextNull := pgtype.Text{Status: pgtype.Null}

	t.Run("success get resource_path", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			&ids.Elements,
			&pgTextNull,
			&pgTextNull,
			&pgTextNull,
		)

		e := &entity.LegacyUser{}

		_ = e.ResourcePath.Set("1")
		mockDB.MockScanArray(nil, []string{"resource_path"}, [][]interface{}{
			{&e.ResourcePath},
		})

		resourcePath, err := r.ResourcePath(ctx, mockDB.DB, pgtype.Text{String: "id", Status: pgtype.Present})
		assert.Nil(t, err)
		assert.Equal(t, "1", resourcePath)

		mockDB.RawStmt.AssertSelectedFields(t, "resource_path")
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"user_id":      {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"email":        {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 2}},
			"phone_number": {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 3}},
			"user_group":   {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 4}},
			"deleted_at":   {HasNullTest: true},
		})
	})

	t.Run("fail to get resource_path", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything,
			mock.AnythingOfType("string"),
			&ids.Elements,
			&pgTextNull,
			&pgTextNull,
			&pgTextNull,
		)

		e := &entity.LegacyUser{}

		_ = e.ResourcePath.Set("1")
		mockDB.MockScanArray(nil, []string{"resource_path"}, [][]interface{}{
			{&e.ResourcePath},
		})

		resourcePath, err := r.ResourcePath(ctx, mockDB.DB, pgtype.Text{String: "id", Status: pgtype.Present})
		assert.Equal(t, "", resourcePath)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
	})
}

func TestUserRepo_UpdateLastLoginDate(t *testing.T) {
	t.Parallel()
	r, mockDB := UserRepoWithSqlMock()

	now := time.Now()
	e := &entity.LegacyUser{}
	_ = e.ID.Set("id")
	_ = e.LastLoginDate.Set(now)
	userID := &pgtype.Text{String: "id", Status: pgtype.Present}
	lastLoginDate := &pgtype.Timestamptz{Time: now, Status: pgtype.Present}

	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag([]byte(`1`)), nil, mock.Anything, mock.Anything, lastLoginDate, userID)
			},
		},
		{
			name:        "error due to no rows",
			expectedErr: errors.New("cannot update user last_login_date"),
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag([]byte(`0`)), nil, mock.Anything, mock.Anything, lastLoginDate, userID)
			},
		},
		{
			name:        "error due to tx error",
			expectedErr: fmt.Errorf("tx mock error"),
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag(""), fmt.Errorf("tx mock error"), mock.Anything, mock.Anything, lastLoginDate, userID)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := r.UpdateLastLoginDate(ctx, mockDB.DB, e)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestUserRepo_CreateMultiple(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	userRepo := &UserRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entity.LegacyUser{
				{
					ID: pgtype.Text{String: "1", Status: pgtype.Present},
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
			name: "happy case: ResourcePath status nil",
			req: []*entity.LegacyUser{
				{
					ID:           pgtype.Text{String: "1", Status: pgtype.Present},
					ResourcePath: pgtype.Text{Status: pgtype.Null},
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
			name: "happy case: create multiple teachers",
			req: []*entity.LegacyUser{
				{
					ID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				},
				{
					ID: pgtype.Text{String: "2", Status: pgtype.Present},
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
			req: []*entity.LegacyUser{
				{
					ID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
			},
			expectedErr: errors.New("batchResults.Exec: closed pool"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "send batch return ",
			req: []*entity.LegacyUser{
				{
					ID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
			},
			expectedErr: errors.New("user not inserted"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`0`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := userRepo.CreateMultiple(ctx, db, testCase.req.([]*entity.LegacyUser))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestUserRepo_Get(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ids := pgtype.TextArray{}
	_ = ids.Set([]string{"id"})
	pgTextNull := pgtype.Text{Status: pgtype.Null}

	t.Run("err select", func(t *testing.T) {
		r, mockDB := UserRepoWithSqlMock()
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything,
			mock.AnythingOfType("string"),
			&ids.Elements,
			&pgTextNull,
			&pgTextNull,
			&pgTextNull,
		)

		user, err := r.Get(ctx, mockDB.DB, ids.Elements[0])
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, user)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		r, mockDB := UserRepoWithSqlMock()
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			&ids.Elements,
			&pgTextNull,
			&pgTextNull,
			&pgTextNull,
		)

		e := &entity.LegacyUser{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		_ = e.FullName.Set(ksuid.New().String())
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		user, err := r.Get(ctx, mockDB.DB, ids.Elements[0])
		assert.Nil(t, err)
		assert.Equal(t, e, user)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"user_id":      {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"email":        {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 2}},
			"phone_number": {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 3}},
			"user_group":   {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 4}},
			"deleted_at":   {HasNullTest: true},
		})
	})
	t.Run("scan error", func(t *testing.T) {
		r, mockDB := UserRepoWithSqlMock()
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			&ids.Elements,
			&pgTextNull,
			&pgTextNull,
			&pgTextNull,
		)

		e := &entity.LegacyUser{}
		fields, _ := e.FieldMap()
		_ = e.ID.Set(ksuid.New().String())
		_ = e.FullName.Set(ksuid.New().String())

		mockDB.MockScanArray(pgx.ErrNoRows, []string{}, [][]interface{}{
			nil,
		})

		_, err := r.Get(ctx, mockDB.DB, ids.Elements[0])
		assert.Equal(t, fmt.Errorf("database.Select: rows.Scan: %w", pgx.ErrNoRows).Error(), err.Error())

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"user_id":      {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"email":        {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 2}},
			"phone_number": {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 3}},
			"user_group":   {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 4}},
			"deleted_at":   {HasNullTest: true},
		})
	})
}

func TestUserRepo_GetByEmail(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	emails := pgtype.TextArray{}
	_ = emails.Set([]string{"id"})

	t.Run("err select", func(t *testing.T) {
		r, mockDB := UserRepoWithSqlMock()
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything,
			mock.AnythingOfType("string"),
			&emails,
			mock.Anything,
		)

		user, err := r.GetByEmail(ctx, mockDB.DB, emails)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, user)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		r, mockDB := UserRepoWithSqlMock()
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			&emails,
			mock.Anything,
		)

		e := &entity.LegacyUser{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		_ = e.FullName.Set(ksuid.New().String())
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		users, err := r.GetByEmail(ctx, mockDB.DB, emails)
		assert.Nil(t, err)
		assert.Equal(t, []*entity.LegacyUser{e}, users)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"email": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestUserRepo_GetByPhone(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	phones := pgtype.TextArray{}
	_ = phones.Set([]string{"id"})

	t.Run("err select", func(t *testing.T) {
		r, mockDB := UserRepoWithSqlMock()
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything,
			mock.AnythingOfType("string"),
			&phones,
			mock.Anything,
		)

		user, err := r.GetByPhone(ctx, mockDB.DB, phones)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, user)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		r, mockDB := UserRepoWithSqlMock()
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			&phones,
			mock.Anything,
		)

		e := &entity.LegacyUser{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		_ = e.FullName.Set(ksuid.New().String())
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		users, err := r.GetByPhone(ctx, mockDB.DB, phones)
		assert.Nil(t, err)
		assert.Equal(t, []*entity.LegacyUser{e}, users)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"phone_number": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})

		t.Run("success with select all fields", func(t *testing.T) {
			r, mockDB := UserRepoWithSqlMock()
			mockDB.MockQueryArgs(t, nil, mock.Anything,
				mock.AnythingOfType("string"),
				&phones,
				mock.Anything,
			)

			e := &entity.LegacyUser{}
			fields, values := e.FieldMap()

			_ = e.ID.Set(ksuid.New().String())
			_ = e.FullName.Set(ksuid.New().String())
			mockDB.MockScanArray(puddle.ErrClosedPool, fields, [][]interface{}{
				values,
			})

			_, err := r.GetByPhone(ctx, mockDB.DB, phones)
			assert.Equal(t, fmt.Errorf("database.Select: rows.Scan: %w", puddle.ErrClosedPool).Error(), err.Error())
		})
	})
}

func TestUserRepo_Create(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := UserRepoWithSqlMock()
		user := entity.LegacyUser{}

		_, userValues := user.FieldMap()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, &user)
		assert.Nil(t, err)
	})
	t.Run("create success whn resource path is nil", func(t *testing.T) {
		repo, mockDB := UserRepoWithSqlMock()
		user := entity.LegacyUser{
			ResourcePath: pgtype.Text{Status: pgtype.Null},
		}

		_, userValues := user.FieldMap()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, &user)
		assert.Nil(t, err)
	})
	t.Run("create fail", func(t *testing.T) {
		repo, mockDB := UserRepoWithSqlMock()
		user := entity.LegacyUser{
			ResourcePath: pgtype.Text{Status: pgtype.Null},
		}

		_, userValues := user.FieldMap()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		mockDB.DB.On("Exec", args...).Return(nil, puddle.ErrClosedPool)

		err := repo.Create(ctx, mockDB.DB, &user)
		assert.Equal(t, fmt.Errorf("user not inserted: %w", puddle.ErrClosedPool), err)
	})
}

func TestUserRepo_UpdateEmail(t *testing.T) {
	t.Parallel()
	r, mockDB := UserRepoWithSqlMock()
	now := time.Now()
	e := &entity.LegacyUser{}
	_ = e.ID.Set("id")
	_ = e.Email.Set("email@example.com")
	_ = e.LoginEmail.Set("login_email@example.com")
	_ = e.UpdatedAt.Set(now)
	userID := &pgtype.Text{String: "id", Status: pgtype.Present}
	email := &pgtype.Text{String: "email@example.com", Status: pgtype.Present}
	loginEmail := &pgtype.Text{String: "login_email@example.com", Status: pgtype.Present}

	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag([]byte(`1`)), nil, mock.Anything, mock.Anything, email, loginEmail, mock.Anything, userID)
			},
		},
		{
			name:        "error due to no rows",
			expectedErr: errors.New("cannot update user email"),
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag([]byte(`0`)), nil, mock.Anything, mock.Anything, email, loginEmail, mock.Anything, userID)
			},
		},
		{
			name:        "error due to tx error",
			expectedErr: fmt.Errorf("tx mock error"),
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag(""), fmt.Errorf("tx mock error"), mock.Anything, mock.Anything, email, loginEmail, mock.Anything, userID)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := r.UpdateEmail(ctx, mockDB.DB, e)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestUserRepo_UpdateProfileV1(t *testing.T) {
	t.Parallel()
	r, mockDB := UserRepoWithSqlMock()

	now := time.Now()
	user := &entity.LegacyUser{}

	user.ID = database.Text("id")
	user.FullName = database.Text("test full name")
	user.UserName = database.Text("username")
	user.Avatar = database.Text("test avatar")
	user.Group = database.Text("test user Group")
	user.PhoneNumber = database.Text("098765432")
	user.UpdatedAt = pgtype.Timestamptz{Time: now, Status: pgtype.Present}

	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag([]byte(`1`)), nil, mock.Anything, mock.Anything, &user.UserName, &user.FullName, &user.Avatar, &user.Group, &user.PhoneNumber, &user.Birthday, &user.Gender, &user.Remarks, mock.Anything, &user.ID)
			},
		},
		{
			name:        "happy case without birthday",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				user.Birthday = pgtype.Date{}
				mockDB.MockExecArgs(t, pgconn.CommandTag([]byte(`1`)), nil, mock.Anything, mock.Anything, &user.UserName, &user.FullName, &user.Avatar, &user.Group, &user.PhoneNumber, &user.Gender, &user.Remarks, mock.Anything, &user.ID)
			},
		},
		{
			name:        "happy case without gender",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				user.Gender = pgtype.Text{}
				mockDB.MockExecArgs(t, pgconn.CommandTag([]byte(`1`)), nil, mock.Anything, mock.Anything, &user.UserName, &user.FullName, &user.Avatar, &user.Group, &user.PhoneNumber, &user.Birthday, &user.Remarks, mock.Anything, &user.ID)
			},
		},
		{
			name:        "happy case without remarks",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				user.Remarks = pgtype.Text{}
				mockDB.MockExecArgs(t, pgconn.CommandTag([]byte(`1`)), nil, mock.Anything, mock.Anything, &user.UserName, &user.FullName, &user.Avatar, &user.Group, &user.PhoneNumber, &user.Birthday, &user.Gender, mock.Anything, &user.ID)
			},
		},
		{
			name:        "error due to no rows",
			expectedErr: errors.New("cannot update profile"),
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag([]byte(`0`)), nil, mock.Anything, mock.Anything, &user.UserName, &user.FullName, &user.Avatar, &user.Group, &user.PhoneNumber, &user.Birthday, &user.Gender, &user.Remarks, mock.Anything, &user.ID)
			},
		},
		{
			name:        "error due to tx error",
			expectedErr: fmt.Errorf("tx mock error"),
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag(""), fmt.Errorf("tx mock error"), mock.Anything, mock.Anything, &user.UserName, &user.FullName, &user.Avatar, &user.Group, &user.PhoneNumber, &user.Birthday, &user.Gender, &user.Remarks, mock.Anything, &user.ID)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			user.Gender = pgtype.Text{String: "MALE", Status: pgtype.Present}
			user.Birthday = pgtype.Date{Time: now, Status: pgtype.Present}
			user.Remarks = pgtype.Text{String: "remarks data", Status: pgtype.Present}

			ctx := context.Background()
			testCase.setup(ctx)
			err := r.UpdateProfileV1(ctx, mockDB.DB, user)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestUserRepo_SearchProfile(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := UserRepoWithSqlMock()
	studentIDs := pgtype.TextArray{}
	_ = studentIDs.Set([]string{"studentID 1", "studentID 2", "studentID 3"})
	locationIDs := pgtype.TextArray{}
	_ = locationIDs.Set([]string{"locationID 1", "locationID 2", "locationID 3"})
	pgTextNull := pgtype.Text{Status: pgtype.Null}
	UintNull := uint(0)
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything,
			mock.AnythingOfType("string"),
			studentIDs,
			pgTextNull,
			locationIDs,
			UintNull,
			UintNull,
		)
		filter := &SearchProfileFilter{
			StudentIDs:  studentIDs,
			StudentName: pgTextNull,
			LocationIDs: locationIDs,
		}
		users, err := r.SearchProfile(ctx, mockDB.DB, filter)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, users)
	})
	t.Run("success get", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			studentIDs,
			pgTextNull,
			locationIDs,
			UintNull,
			UintNull,
		)
		filter := &SearchProfileFilter{
			StudentIDs:  studentIDs,
			StudentName: pgTextNull,
			LocationIDs: locationIDs,
		}
		e := &entity.LegacyUser{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		users, err := r.SearchProfile(ctx, mockDB.DB, filter)
		assert.True(t, errors.Is(err, nil))
		assert.NotNil(t, users)
	})
}

func TestUserRepo_GetUserGroups(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r := &UserRepo{}
	id := database.Text(idutil.ULIDNow())
	userGroupV2 := &entity.UserGroupV2{}
	fields, values := userGroupV2.FieldMap()

	tests := []struct {
		name         string
		ctx          context.Context
		expectedErr  error
		expectedResp bool
		setup        func(context.Context) *mock_database.Ext
	}{
		{
			name:        "happy case",
			ctx:         ctx,
			expectedErr: nil,
			setup: func(ctx context.Context) *mock_database.Ext {
				mockDB := testutil.NewMockDB()
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), &id)
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
				return mockDB.DB
			},
		},
		{
			name:        "error when execute query",
			ctx:         ctx,
			expectedErr: fmt.Errorf("database.Select: %w", fmt.Errorf("err db.Query: %w", puddle.ErrClosedPool)),
			setup: func(ctx context.Context) *mock_database.Ext {
				mockDB := testutil.NewMockDB()
				mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), &id)
				return mockDB.DB
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			db := testCase.setup(testCase.ctx)

			userGroups, err := r.GetUserGroups(testCase.ctx, db, id)

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NotNil(t, userGroups)
			}
		})
	}
}

func TestRepo_GetUserRoles(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo := &UserRepo{}
	id := database.Text(idutil.ULIDNow())
	role := &entity.Role{}
	fields, values := role.FieldMap()

	tests := []struct {
		name        string
		ctx         context.Context
		expectedErr error
		setup       func(context.Context) *mock_database.Ext
	}{
		{
			name:        "happy case",
			ctx:         ctx,
			expectedErr: nil,
			setup: func(ctx context.Context) *mock_database.Ext {
				mockDB := testutil.NewMockDB()
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), &id)
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
				return mockDB.DB
			},
		},
		{
			name:        "error when execute query",
			ctx:         ctx,
			expectedErr: fmt.Errorf("database.Select: %w", fmt.Errorf("err db.Query: %w", puddle.ErrClosedPool)),
			setup: func(ctx context.Context) *mock_database.Ext {
				mockDB := testutil.NewMockDB()
				mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), &id)
				return mockDB.DB
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			db := testCase.setup(testCase.ctx)
			userGroups, err := repo.GetUserRoles(testCase.ctx, db, id)

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NotNil(t, userGroups)
			}
		})
	}
}

func TestRepo_GetUserGroupMembers(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo := &UserRepo{}
	id := database.Text(idutil.ULIDNow())
	userGroupMember := &entity.UserGroupMember{}
	fields, values := userGroupMember.FieldMap()

	tests := []struct {
		name         string
		ctx          context.Context
		expectedErr  error
		expectedResp []*entity.UserGroupMember
		setup        func(context.Context) *mock_database.Ext
	}{
		{
			name:         "happy case",
			ctx:          ctx,
			expectedErr:  nil,
			expectedResp: []*entity.UserGroupMember{{}},
			setup: func(ctx context.Context) *mock_database.Ext {
				mockDB := testutil.NewMockDB()
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), &id)
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
				return mockDB.DB
			},
		},
		{
			name:        "error when execute query",
			ctx:         ctx,
			expectedErr: fmt.Errorf("database.Select: %w", fmt.Errorf("err db.Query: %w", puddle.ErrClosedPool)),
			setup: func(ctx context.Context) *mock_database.Ext {
				mockDB := testutil.NewMockDB()
				mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), &id)
				return mockDB.DB
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			db := testCase.setup(testCase.ctx)
			userGroups, err := repo.GetUserGroupMembers(testCase.ctx, db, id)
			assert.Equal(t, testCase.expectedResp, userGroups)
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestUserRepo_GetByUserID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := UserRepoWithSqlMock()
	id := database.Text(idutil.ULIDNow())
	user := &entity.LegacyUser{}
	fields, values := user.FieldMap()

	tests := []struct {
		name        string
		ctx         context.Context
		expectedErr error
		setup       func(context.Context) *mock_database.Ext
	}{
		{
			name:        "happy case",
			ctx:         ctx,
			expectedErr: nil,
			setup: func(ctx context.Context) *mock_database.Ext {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), &id)
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
				return mockDB.DB
			},
		},
		{
			name:        "error when execute query",
			ctx:         ctx,
			expectedErr: fmt.Errorf("err db.Query: %w", puddle.ErrClosedPool),
			setup: func(ctx context.Context) *mock_database.Ext {
				mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), &id)
				return mockDB.DB
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			db := testCase.setup(testCase.ctx)
			_, err := repo.FindByIDUnscope(testCase.ctx, db, id)
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestUserRepo_SoftDelete(t *testing.T) {
	t.Parallel()
	repo, mockDB := UserRepoWithSqlMock()
	userIDs := database.TextArray([]string{"user-1", "user-2"})

	testCases := []TestCase{
		{
			name:        "error cannot delete user",
			expectedErr: fmt.Errorf("cannot delete user"),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), &userIDs).Once().Return(nil, fmt.Errorf("cannot delete user"))
			},
		},
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`2`))
				mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), &userIDs).Once().Return(cmdTag, nil)
				mockDB.DB.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := repo.SoftDelete(ctx, mockDB.DB, userIDs)
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestUserRepo_GetUsersByUserGroupID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo := &UserRepo{}
	id := database.Text(idutil.ULIDNow())
	userGroupMember := &entity.LegacyUser{}
	fields, values := userGroupMember.FieldMap()

	tests := []struct {
		name        string
		ctx         context.Context
		expectedErr error
		setup       func(context.Context) *mock_database.Ext
	}{
		{
			name:        "happy case",
			ctx:         ctx,
			expectedErr: nil,
			setup: func(ctx context.Context) *mock_database.Ext {
				mockDB := testutil.NewMockDB()
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), &id)
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
				return mockDB.DB
			},
		},
		{
			name:        "error when execute query",
			ctx:         ctx,
			expectedErr: fmt.Errorf("database.Select: %w", fmt.Errorf("err db.Query: %w", puddle.ErrClosedPool)),
			setup: func(ctx context.Context) *mock_database.Ext {
				mockDB := testutil.NewMockDB()
				mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), &id)
				return mockDB.DB
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			db := testCase.setup(testCase.ctx)
			_, err := repo.GetUsersByUserGroupID(testCase.ctx, db, id)
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestUserRepo_UpdateManyUserGroup(t *testing.T) {
	t.Parallel()
	usrEmailRepo := &UserRepo{}
	db := testutil.NewMockDB()

	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(successTag))
				db.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
			},
		},
		{
			name:        "connection closed",
			expectedErr: fmt.Errorf("db.Exec: %w", puddle.ErrClosedPool),
			setup: func(ctx context.Context) {
				db.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, puddle.ErrClosedPool)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := usrEmailRepo.UpdateManyUserGroup(ctx, db.DB, database.TextArray([]string{idutil.ULIDNow()}), database.Text(idutil.ULIDNow()))
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestUserRepo_GetBasicInfo(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	UserRepo := &UserRepo{}
	mockDB := testutil.NewMockDB()
	row := &mock_database.Row{}
	var (
		userID, userGroup string
	)
	scanFields := []interface{}{&userID, &userGroup}

	testCases := []struct {
		name      string
		setup     func()
		expectErr error
	}{
		{
			name:      "happy case",
			expectErr: nil,
			setup: func() {
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Once().Return(row, nil)
				row.On("Scan", scanFields...).Once().Return(nil)
			},
		},
		{
			name:      "query return no row err",
			expectErr: fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows),
			setup: func() {
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Once().Return(row, nil)
				row.On("Scan", scanFields...).Once().Return(pgx.ErrNoRows)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup()
			_, err := UserRepo.GetBasicInfo(ctx, mockDB.DB, database.Text("123"))
			assert.Equal(t, err, testCase.expectErr)
			mock.AssertExpectationsForObjects(t, mockDB.DB)
		})
	}
}
