package repo

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func UserBasicInfoRepoSqlMock() (*UserBasicInfoRepo, *testutil.MockDB) {
	r := &UserBasicInfoRepo{}
	return r, testutil.NewMockDB()
}

func TestUserBasicInfoRepo_GetTeachersSameGrantedLocation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := UserBasicInfoRepoSqlMock()
	var (
		UserID            pgtype.Text
		Name              pgtype.Text
		FirstName         pgtype.Text
		LastName          pgtype.Text
		FullNamePhonetic  pgtype.Text
		FirstNamePhonetic pgtype.Text
		LastNamePhonetic  pgtype.Text
		GradeID           pgtype.Text
		CreatedAt         pgtype.Timestamptz
		UpdatedAt         pgtype.Timestamptz
	)
	t.Run("success", func(t *testing.T) {
		query := domain.UserBasicInfoQuery{
			KeyWord:    "123",
			LocationID: "11213",
			Offset:     1,
			Limit:      2,
		}

		fields := []string{"user_id", "name", "first_name", "last_name", "full_name_phonetic",
			"first_name_phonetic", "last_name_phonetic", "grade_id", "created_at", "updated_at"}

		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		values := []interface{}{&UserID, &Name, &FirstName, &LastName, &FullNamePhonetic,
			&FirstNamePhonetic, &LastNamePhonetic, &GradeID, &CreatedAt, &UpdatedAt}

		mockDB.MockScanFields(nil, fields, values)

		rs, err := r.GetTeachersSameGrantedLocation(ctx, mockDB.DB, query)
		assert.NoError(t, err)
		assert.NotNil(t, rs)
	})
}

func TestUserBasicInfoRepo_GetUser(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	userBasicInfoRepo, mockDB := UserBasicInfoRepoSqlMock()

	args := []interface{}{mock.Anything, mock.Anything, []string{"user-id-1", "user-id-2"}}

	t.Run("success", func(t *testing.T) {
		u := &UserBasicInfo{}
		fields, values := u.FieldMap()
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		userBasicInfo, err := userBasicInfoRepo.GetUser(ctx, mockDB.DB, []string{"user-id-1", "user-id-2"})
		assert.Nil(t, err)
		assert.NotNil(t, userBasicInfo)
	})
}
