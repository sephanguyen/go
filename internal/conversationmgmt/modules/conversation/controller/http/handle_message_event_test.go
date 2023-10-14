package http

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/utils"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_service "github.com/manabie-com/backend/mock/conversationmgmt/modules/conversation/core/port/service"
	mock_repositories "github.com/manabie-com/backend/mock/conversationmgmt/modules/conversation/infrastructure/postgres"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gotest.tools/assert"
)

func TestConversationModifierHTTP_HandleMessageEvent_UpdateLatestMessage(t *testing.T) {
	t.Parallel()
	mockDB := &mock_database.Ext{}
	mockPortService := mock_service.NewConversationModifierService(t)
	mockConversationRepo := &mock_repositories.MockConversationRepo{}
	mockChatVendorUserRepo := &mock_repositories.MockAgoraUserRepo{}

	svc := &ConversationModifierHTTP{
		DB:                              mockDB,
		ConversationModifierServicePort: mockPortService,
		ConversationRepo:                mockConversationRepo,
		ChatVendorUserRepo:              mockChatVendorUserRepo,
	}

	t.Run("missing resource_path", func(t *testing.T) {
		core, log := observer.New(zap.InfoLevel)
		logger := zap.New(core)
		svc.Logger = logger
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		// dummy body data, expected to fail to marshalling this
		rawReqEvent := map[string]interface{}{
			"eventType": "chat",
			"chat_type": "groupchat",
			"group_id":  "convo-id",
		}

		ctx.Request, _ = utils.NewMockRequest("POST", rawReqEvent, nil)

		svc.HandleMessageEvent(ctx)

		entry := log.All()[0]
		assert.Equal(t, "cannot verify mesage and conversation: [empty resource_path]", entry.Message)
	})

	t.Run("failed verify resource_path - not found user", func(t *testing.T) {
		core, log := observer.New(zap.InfoLevel)
		logger := zap.New(core)
		svc.Logger = logger
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		// dummy body data, expected to fail to marshalling this
		rawReqEvent := map[string]interface{}{
			"eventType": "chat",
			"chat_type": "groupchat",
			"from":      "from-id",
			"payload": map[string]interface{}{
				"ext": map[string]interface{}{
					"resource_path": "resource_path",
				},
			},
			"group_id": "convo-id",
		}

		ctx.Request, _ = utils.NewMockRequest("POST", rawReqEvent, nil)
		tenantCtx := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{ResourcePath: "resource_path"},
		})
		// Return empty conversation
		mockChatVendorUserRepo.On("GetByVendorUserIDs", tenantCtx, mock.Anything, []string{"from-id"}).Once().Return([]*domain.ChatVendorUser{}, nil)

		svc.HandleMessageEvent(ctx)

		entry := log.All()[0]
		assert.Equal(t, "cannot verify mesage and conversation: [not found vendor user]", entry.Message)
	})

	t.Run("failed verify resource_path - not found conversation", func(t *testing.T) {
		core, log := observer.New(zap.InfoLevel)
		logger := zap.New(core)
		svc.Logger = logger
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		// dummy body data, expected to fail to marshalling this
		rawReqEvent := map[string]interface{}{
			"eventType": "chat",
			"chat_type": "groupchat",
			"from":      "from-id",
			"payload": map[string]interface{}{
				"ext": map[string]interface{}{
					"resource_path": "resource_path",
				},
			},
			"group_id": "convo-id",
		}

		ctx.Request, _ = utils.NewMockRequest("POST", rawReqEvent, nil)
		tenantCtx := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{ResourcePath: "resource_path"},
		})
		// Return empty conversation
		mockChatVendorUserRepo.On("GetByVendorUserIDs", tenantCtx, mock.Anything, []string{"from-id"}).Once().Return([]*domain.ChatVendorUser{{UserID: "user-id"}}, nil)
		mockConversationRepo.On("FindByIDsAndUserID", tenantCtx, mock.Anything, "user-id", []string{"convo-id"}).Once().Return([]*domain.Conversation{}, nil)

		svc.HandleMessageEvent(ctx)

		entry := log.All()[0]
		assert.Equal(t, "cannot verify mesage and conversation: [not found conversation]", entry.Message)
	})

	t.Run("happy case - new message", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		// dummy body data, expected to fail to marshalling this
		rawReqEvent := map[string]interface{}{
			"callId":    "call-id",
			"eventType": "chat",
			"chat_type": "groupchat",
			"from":      "from-id",
			"payload": map[string]interface{}{
				"ext": map[string]interface{}{
					"resource_path":    "resource_path",
					"manabie_msg_type": "text",
					"paths": []map[string]interface{}{
						{
							"name": "file.ext",
							"url":  "example.com",
						},
					},
				},
				"bodies": []map[string]interface{}{
					{
						"msg":  "msg",
						"type": "txt",
					},
				},
				"type": "groupchat",
			},
			"group_id":  "convo-id",
			"to":        "convo-id",
			"msg_id":    "msg-id",
			"timestamp": 1646101800000,
		}

		ctx.Request, _ = utils.NewMockRequest("POST", rawReqEvent, nil)
		tenantCtx := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{ResourcePath: "resource_path"},
		})
		mockChatVendorUserRepo.On("GetByVendorUserIDs", tenantCtx, mock.Anything, []string{"from-id"}).Once().Return([]*domain.ChatVendorUser{{UserID: "user-id"}}, nil)
		mockConversationRepo.On("FindByIDsAndUserID", tenantCtx, mock.Anything, "user-id", []string{"convo-id"}).Once().Return([]*domain.Conversation{{ID: "convo-id"}}, nil)

		expectedSentTime, _ := time.Parse(time.RFC3339, "2022-03-01T02:30:00.00Z")
		mockPortService.On("UpdateLatestMessage", tenantCtx, &domain.Message{
			ConversationID:  "convo-id",
			VendorMessageID: "msg-id",
			Message:         "msg",
			VendorUserID:    "from-id",
			UserID:          "user-id",
			Type:            domain.MessageTypeText,
			SentTime:        expectedSentTime.Local(),
			IsDeleted:       false,
			Media: []domain.MessageMedia{
				{
					Name: "file.ext",
					URL:  "example.com",
				},
			},
		}).Once().Return(nil)

		svc.HandleMessageEvent(ctx)
	})

}
