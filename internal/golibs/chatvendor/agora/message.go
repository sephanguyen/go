package agora

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/chatvendor/agora/dto"
	abstract_dto "github.com/manabie-com/backend/internal/golibs/chatvendor/dto"
)

func (a *agoraClientImpl) DeleteMessage(req *abstract_dto.DeleteMessageRequest) (*abstract_dto.DeleteMessageResponse, error) {
	if req.ConversationID == "" || req.VendorMessageID == "" || req.VendorUserID == "" {
		return nil, fmt.Errorf("missing conversation_id | vendor_message_id | vendor_user_id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	reqBody, err := json.Marshal(&dto.RecallMessageRequest{
		MessageID: req.VendorMessageID,
		To:        req.ConversationID,
		From:      req.VendorUserID,
		ChatType:  "groupchat",
		Force:     true,
	})
	if err != nil {
		return nil, err
	}

	// Create chatgroup endpoint: POST /messages/msg_recall
	endpoint := string(RecallMessage)

	recallMessageResponse := &dto.RecallMessageResponse{}
	err = a.doRequest(ctx, MethodPost, endpoint, GetAgoraCommonHeader(), bytes.NewBuffer(reqBody), recallMessageResponse)
	if err != nil {
		return nil, err
	}

	if recallMessageResponse.Data.Recalled {
		return &abstract_dto.DeleteMessageResponse{
			IsSuccess: true,
		}, nil
	}

	return nil, fmt.Errorf("[agora]: cannot recall message")
}
