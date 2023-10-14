package http

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/constants"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/controller/http/payload"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/gin-gonic/gin"
)

func (s *ConversationModifierHTTP) HandleMessageEvent(ctx *gin.Context) {
	rawReq, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("failed io.ReadAl: [%+v]", err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, payload.NewMessageEventResponse(false, err))
		return
	}

	req, err := payload.NewMessageEventRequestFromJSONBytes(rawReq)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("NewMessageEventRequestFromJSONBytes: [%+v]", err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, payload.NewMessageEventResponse(false, err))
		return
	}

	handlerType, conversationID, vendorUserID := req.GetWebhookHandlerTypeAndConversationIDAndVendorUserID()
	if handlerType == constants.WebhookHandlerType("") || conversationID == "" {
		s.Logger.Error("missing conversation_id in request webhook | or cannot specify handler type")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, payload.NewMessageEventResponse(false, err))
		return
	}

	// IMPORTANT: For multi-tenancy/member checking, the validation of webhook and our data
	tenantCtx := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{ResourcePath: req.Payload.Extension.ResourcePath},
	})
	manabieUserID, err := s.verifyWebhookEvtWithMultiTenancyAndConversationMember(tenantCtx, req.Payload.Extension.ResourcePath, conversationID, vendorUserID)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("cannot verify mesage and conversation: [%+v]", err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, payload.NewMessageEventResponse(false, err))
		return
	}

	switch handlerType {
	case constants.WebhookHandlerTypeNewMessage:
		// Handle new message - latest message
		msg := req.ToMessageDomain(manabieUserID)
		err = s.ConversationModifierServicePort.UpdateLatestMessage(tenantCtx, msg)
		if err != nil {
			s.Logger.Error(fmt.Sprintf("cannot handle UpdateLatestMessage: [%+v]", err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, payload.NewMessageEventResponse(false, err))
			return
		}
	case constants.WebhookHandlerTypeOfflineMessage:
		// Handle offline message - push offline notification
		offlineMsg := req.ToOfflineMessageDomain(manabieUserID)
		err = s.NotificationHandlerServicePort.PushNotification(tenantCtx, offlineMsg)
		if err != nil {
			s.Logger.Error(fmt.Sprintf("cannot handle PushNotification: [%+v]", err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, payload.NewMessageEventResponse(false, err))
			return
		}
	case constants.WebhookHandlerTypeDeleteMessage:
		// Currently, do nothing on this, maybe need it in the future
	default:
		// Do nothing
	}

	ctx.JSON(http.StatusOK, payload.NewMessageEventResponse(true, nil))
}

func (s *ConversationModifierHTTP) verifyWebhookEvtWithMultiTenancyAndConversationMember(ctx context.Context, resourcePath, conversationID, vendorUserID string) (string, error) {
	if resourcePath == "" {
		return "", fmt.Errorf("empty resource_path")
	}

	vendorUsers, err := s.ChatVendorUserRepo.GetByVendorUserIDs(ctx, s.DB, []string{vendorUserID})
	if err != nil {
		return "", err
	}
	if len(vendorUsers) == 0 {
		return "", fmt.Errorf("not found vendor user")
	}

	// Need to check this action is from a user that is a member of current conversation.
	userID := vendorUsers[0].UserID
	conversations, err := s.ConversationRepo.FindByIDsAndUserID(ctx, s.DB, userID, []string{conversationID})
	if err != nil {
		return "", err
	}
	if len(conversations) == 0 {
		return "", fmt.Errorf("not found conversation")
	}

	return vendorUsers[0].UserID, nil
}
