package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

		users, err := r.retrieve(ctx, mockDB.DB, ids)
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

		e := &User{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		_ = e.LastName.Set(ksuid.New().String())
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		users, err := r.retrieve(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.Equal(t, []*User{e}, users)

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

		e := &User{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		_ = e.LastName.Set(ksuid.New().String())
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		users, err := r.retrieve(ctx, mockDB.DB, ids, "id", "name")
		assert.Nil(t, err)
		assert.Equal(t, []*User{e}, users)

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
func TestUserRepo_GetUserByUserID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := UserRepoWithSqlMock()
	ids := pgtype.TextArray{}
	_ = ids.Set([]string{"id"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything,
			mock.AnythingOfType("string"),
			mock.Anything, mock.Anything,
		)

		users, err := r.GetUserByUserID(ctx, mockDB.DB, "id")
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, users)
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

		e := &User{}

		_ = e.Group.Set("USER_GROUP_STUDENT")
		mockDB.MockScanArray(nil, []string{"user_group"}, [][]interface{}{
			{&e.Group},
		})

		group, err := r.GetUserGroupByUserID(ctx, mockDB.DB, "id")
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
func TestUserRepo_GetStudentCurrentGradeByUserIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockUserRepo, mockDB := UserRepoWithSqlMock()
	userIDs := []string{"test-user-id-1", "test-user-id-2", "test-user-id-3"}

	fields := []string{
		"user_id",
		"student_grade",
	}

	var (
		userID       pgtype.Text
		studentGrade pgtype.Text
	)
	values := make([]interface{}, 0, 2)
	values = append(values, &userID)
	values = append(values, &studentGrade)

	t.Run("failed to get student grades", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), userIDs)

		studentGradeMap, err := mockUserRepo.GetStudentCurrentGradeByUserIDs(ctx, mockDB.DB, userIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, studentGradeMap)
	})

	t.Run("successfully fetched student grades", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), userIDs)

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		studentGradeMap, err := mockUserRepo.GetStudentCurrentGradeByUserIDs(ctx, mockDB.DB, userIDs)
		assert.Nil(t, err)
		assert.NotNil(t, studentGradeMap)
	})

	t.Run("successfully fetched student grades using user basic info", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), userIDs)

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		studentGradeMap, err := mockUserRepo.GetStudentCurrentGradeByUserIDs(ctx, mockDB.DB, userIDs)
		assert.Nil(t, err)
		assert.NotNil(t, studentGradeMap)
	})
}
func TestUserRepo_GetStudentsManyReferenceByNameOrEmail(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockUserRepo, mockDB := UserRepoWithSqlMock()
	keyword := "name"
	limit := uint32(30)
	offset := uint32(0)

	student := Student{}
	fields, _ := student.FieldMap()
	var (
		ID    pgtype.Text
		Name  pgtype.Text
		Email pgtype.Text
	)
	values := make([]interface{}, 0, 4)
	values = append(values, &ID)
	values = append(values, &Name)
	values = append(values, &Email)

	t.Run("failed to get student", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, limit, offset)

		students, err := mockUserRepo.GetStudentsManyReferenceByNameOrEmail(ctx, mockDB.DB, keyword, limit, offset)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, students)
	})

	t.Run("successfully fetched student", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, limit, offset)

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		studentGradeMap, err := mockUserRepo.GetStudentsManyReferenceByNameOrEmail(ctx, mockDB.DB, keyword, limit, offset)
		assert.Nil(t, err)
		assert.NotNil(t, studentGradeMap)
	})

	t.Run("successfully fetched student without keyword", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, limit, offset)

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		studentGradeMap, err := mockUserRepo.GetStudentsManyReferenceByNameOrEmail(ctx, mockDB.DB, "", limit, offset)
		assert.Nil(t, err)
		assert.NotNil(t, studentGradeMap)
	})
}
