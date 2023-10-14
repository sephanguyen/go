package agora

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/chatvendor/agora/dto"
	abstract_dto "github.com/manabie-com/backend/internal/golibs/chatvendor/dto"
	"github.com/stretchr/testify/assert"
)

func Test_RecallMessage(t *testing.T) {
	t.Run("happy case", func(t *testing.T) {
		t.Parallel()
		userID := "example-username"
		msgID := "example-message-id"
		chatgroupID := "example-chatgroup-id"
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			req := &dto.RecallMessageRequest{}
			err := json.NewDecoder(r.Body).Decode(&req)
			w.WriteHeader(http.StatusOK)
			if err == nil {
				json.NewEncoder(w).Encode(dto.RecallMessageResponse{
					Timestamp:       1,
					Application:     "app-test",
					ApplicationName: "app-test-name",
					Organization:    "org-test",
					Action:          "get",
					URI:             "https://example.com",
					Path:            "/messages/msg_recall",
					Duration:        0,
					Data: struct {
						Recalled  bool   `json:"recalled"`
						ChatType  string `json:"chattype"`
						To        string `json:"to"`
						From      string `json:"from"`
						MessageID string `json:"msg_id"`
					}{
						Recalled:  true,
						ChatType:  "groupchat",
						From:      req.From,
						To:        req.To,
						MessageID: msgID,
					},
				})
			}

		}))
		defer ts.Close()

		agoraClient := newAgoraClientForUnitTest(ts.URL)
		resRecall, err := agoraClient.DeleteMessage(&abstract_dto.DeleteMessageRequest{
			ConversationID:  chatgroupID,
			VendorUserID:    userID,
			VendorMessageID: msgID,
		})
		assert.Equal(t, nil, err)
		assert.Equal(t, true, resRecall.IsSuccess)
	})
}
