package dto

import (
	"testing"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/stretchr/testify/assert"
)

func Test_ToInternalAdminUserDomain(t *testing.T) {
	t.Parallel()

	t.Run("should convert successful", func(t *testing.T) {
		dto := &InternalAdminUserPgDTO{
			UserID:       database.Text("user-id"),
			VendorUserID: database.Text("vendor-user-id"),
			IsSystem:     database.Bool(false),
		}

		domainResult := dto.ToInternalAdminUserDomain()
		expectedDomain := &domain.InternalAdminUser{
			UserID:       "user-id",
			VendorUserID: "vendor-user-id",
			IsSystem:     false,
		}

		assert.Equal(t, expectedDomain, domainResult)
	})
	t.Run("nil", func(t *testing.T) {
		dto := &MessagePgDTO{}
		dto = nil
		domainResult := dto.ToMessageDomain()
		assert.Nil(t, domainResult)
	})
}
