package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/domain/models"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserBasicInfoRepo_GetUsers(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	userBasicInfoRepo := UserBasicInfoRepo{}
	mockDB := testutil.NewMockDB()
	args := []interface{}{mock.Anything, mock.Anything, []string{"user-id-1", "user-id-2"}}

	t.Run("success", func(t *testing.T) {
		u := &models.UserBasicInfo{}
		fields, values := u.FieldMap()
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		userBasicInfo, err := userBasicInfoRepo.GetUsers(ctx, mockDB.DB, []string{"user-id-1", "user-id-2"})
		assert.Nil(t, err)
		assert.NotNil(t, userBasicInfo)
	})
}
