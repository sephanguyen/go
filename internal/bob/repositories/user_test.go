package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
)

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

		e := &entities_bob.User{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		_ = e.LastName.Set(ksuid.New().String())
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		users, err := r.Retrieve(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.Equal(t, []*entities_bob.User{e}, users)

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

		e := &entities_bob.User{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		_ = e.LastName.Set(ksuid.New().String())
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		users, err := r.Retrieve(ctx, mockDB.DB, ids, "id", "name")
		assert.Nil(t, err)
		assert.Equal(t, []*entities_bob.User{e}, users)

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

		e := &entities_bob.User{}

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

		e := &entities_bob.User{}

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

		e := &entities_bob.User{}

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
	e := &entities_bob.User{}
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

type mockCacheCtx struct {
	hit       bool
	callTrack map[string]int
}

func (m *mockCacheCtx) Set(ctx context.Context, group, key string, value interface{}, ttl time.Duration) bool {
	m.callTrack["Set"]++
	return true
}

func (m *mockCacheCtx) Get(ctx context.Context, group, key string) (interface{}, bool) {
	m.callTrack["Get"]++
	if m.hit {
		return "USER_GROUP_STUDENT", true
	}
	return nil, false
}

func (m *mockCacheCtx) Del(ctx context.Context, group, key string) bool {
	m.callTrack["Del"]++
	return true
}

func TestUserRepoWrapper_UserGroup_NoCache(t *testing.T) {
	t.Parallel()
	r, mockDB := UserRepoWithSqlMock()
	mc := &mockCacheCtx{
		callTrack: map[string]int{},
	}
	w := &UserRepoWrapper{
		UserRepository: r,
		LocalCacher:    mc,
	}

	pgTextNull := pgtype.Text{Status: pgtype.Null}
	ids := database.TextArray([]string{"a"})
	mockDB.MockQueryArgs(t, nil, mock.Anything,
		mock.AnythingOfType("string"),
		&ids.Elements,
		&pgTextNull,
		&pgTextNull,
		&pgTextNull,
	)

	e := &entities_bob.User{}
	mockDB.MockScanArray(nil, []string{"user_group"}, [][]interface{}{
		{&e.Group},
	})

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{"no-cache": []string{"1"}})
	w.UserGroup(ctx, mockDB.DB, ids.Elements[0])

	for k, v := range mc.callTrack {
		assert.Equal(t, 0, v, "expecting 0 call to cache repo, got %d call to method %s", v, k)
	}
}

func TestUserRepoWrapper_UserGroup_WithCache_CacheHit(t *testing.T) {
	t.Parallel()
	r, mockDB := UserRepoWithSqlMock()
	mc := &mockCacheCtx{
		hit:       true,
		callTrack: map[string]int{},
	}
	w := &UserRepoWrapper{
		UserRepository: r,
		LocalCacher:    mc,
	}

	pgTextNull := pgtype.Text{Status: pgtype.Null}
	ids := database.TextArray([]string{"a"})
	mockDB.MockQueryArgs(t, nil, mock.Anything,
		mock.AnythingOfType("string"),
		&ids.Elements,
		&pgTextNull,
		&pgTextNull,
		&pgTextNull,
	)

	e := &entities_bob.User{}
	mockDB.MockScanArray(nil, []string{"user_group"}, [][]interface{}{
		{&e.Group},
	})

	ctx := context.Background()
	group, err := w.UserGroup(ctx, mockDB.DB, ids.Elements[0])
	assert.NoError(t, err, "expecting no error returned")
	assert.Equal(t, "USER_GROUP_STUDENT", group, "unexpected group returned")
	assert.Equal(t, 1, mc.callTrack["Get"], "expecting 1 call to cache repo.Get")
	assert.Equal(t, 0, mc.callTrack["Set"], "expecting 0 call to cache repo.Set")
	assert.Equal(t, 0, mc.callTrack["Del"], "expecting 0 call to cache repo.Del")
}

func TestUserRepoWrapper_UserGroup_WithCache_CacheMiss(t *testing.T) {
	t.Parallel()
	r, mockDB := UserRepoWithSqlMock()
	mc := &mockCacheCtx{
		hit:       false,
		callTrack: map[string]int{},
	}
	w := &UserRepoWrapper{
		UserRepository: r,
		LocalCacher:    mc,
	}

	pgTextNull := pgtype.Text{Status: pgtype.Null}
	ids := database.TextArray([]string{"a"})
	mockDB.MockQueryArgs(t, nil, mock.Anything,
		mock.AnythingOfType("string"),
		&ids.Elements,
		&pgTextNull,
		&pgTextNull,
		&pgTextNull,
	)

	e := &entities_bob.User{
		Group: database.Text("USER_GROUP_STUDENT"),
	}
	mockDB.MockScanArray(nil, []string{"user_group"}, [][]interface{}{
		{&e.Group},
	})

	ctx := context.Background()
	group, err := w.UserGroup(ctx, mockDB.DB, ids.Elements[0])
	assert.NoError(t, err, "expecting no error returned")
	assert.Equal(t, "USER_GROUP_STUDENT", group, "expecting STUDENT group")
	assert.Equal(t, 1, mc.callTrack["Get"], "expecting 1 call to cache repo.Get")
	assert.Equal(t, 1, mc.callTrack["Set"], "expecting 0 call to cache repo.Set")
	assert.Equal(t, 0, mc.callTrack["Del"], "expecting 0 call to cache repo.Del")
}

func TestUserRepo_SearchProfile(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := UserRepoWithSqlMock()
	ids := pgtype.TextArray{}
	_ = ids.Set([]string{"id"})
	pgTextNull := pgtype.Text{Status: pgtype.Null}
	UintNull := uint(0)
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything,
			mock.AnythingOfType("string"),
			ids,
			pgTextNull,
			UintNull,
			UintNull,
		)
		filter := &SearchProfileFilter{
			StudentIDs:  ids,
			StudentName: pgTextNull,
		}
		users, err := r.SearchProfile(ctx, mockDB.DB, filter)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, users)
	})
	t.Run("success get", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			ids,
			pgTextNull,
			UintNull,
			UintNull,
		)
		filter := &SearchProfileFilter{
			StudentIDs:  ids,
			StudentName: pgTextNull,
		}
		e := &entities_bob.User{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		users, err := r.SearchProfile(ctx, mockDB.DB, filter)
		assert.True(t, errors.Is(err, nil))
		assert.NotNil(t, users)
	})
}

func TestUserRepo_GetUsernameByUserID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := UserRepoWithSqlMock()
	t.Run("find error", func(t *testing.T) {
		id := "id"
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything,
			mock.AnythingOfType("string"),
			&id,
		)

		user, err := r.GetUsernameByUserID(ctx, mockDB.DB, id)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, user)
	})

	t.Run("find success", func(tt *testing.T) {
		id := "id"
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			&id,
		)

		e := &entities_bob.Username{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		user, err := r.GetUsernameByUserID(ctx, mockDB.DB, id)
		assert.NoError(tt, err)
		assert.Equal(tt, e, user)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}
