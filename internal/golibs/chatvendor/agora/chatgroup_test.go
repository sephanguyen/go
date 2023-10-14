package agora

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/chatvendor/agora/dto"
	abstract_dto "github.com/manabie-com/backend/internal/golibs/chatvendor/dto"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/stretchr/testify/assert"
)

func Test_CreateChatGroup(t *testing.T) {
	t.Run("happy case", func(t *testing.T) {
		t.Parallel()
		ownerUserID := "example-username"
		chatgroupID := "example-chatgroup-id"
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			req := &dto.CreateChatGroupRequest{}
			err := json.NewDecoder(r.Body).Decode(&req)
			if err == nil && req.Owner == ownerUserID {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(dto.CreateChatGroupResponse{
					Timestamp:       1,
					Application:     "app-test",
					ApplicationName: "app-test-name",
					Organization:    "org-test",
					Action:          "get",
					URI:             "https://example.com",
					Data: struct {
						GroupID string "json:\"groupid\""
					}{
						GroupID: chatgroupID,
					},
				})
			} else {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(dto.ErrorResponse{
					Duration:         1,
					Timestamp:        1,
					Exception:        "failed",
					Error:            "bad_request",
					ErrorDescription: "bad_request",
				})
			}
		}))
		defer ts.Close()

		agoraClient := newAgoraClientForUnitTest(ts.URL)
		chatgroup, err := agoraClient.CreateConversation(&abstract_dto.CreateConversationRequest{
			OwnerVendorID:   ownerUserID,
			MemberVendorIDs: []string{"mem-1", "mem-2"},
		})
		assert.Equal(t, nil, err)
		assert.Equal(t, chatgroupID, chatgroup.ConversationID)
	})
}

func Test_AddConversationMembers(t *testing.T) {
	t.Parallel()

	t.Run("happy case", func(t *testing.T) {
		chatgroupID := "example-chatgroup-id"
		memberIDs := []string{"member-1", "member-2"}
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			req := &dto.AddMemberToChatGroupRequest{}
			err := json.NewDecoder(r.Body).Decode(&req)
			if err == nil && stringutil.SliceEqual(req.Usernames, memberIDs) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(dto.AddMemberToChatGroupResponse{
					Timestamp:       1,
					Application:     "app-test",
					ApplicationName: "app-test-name",
					Organization:    "org-test",
					Action:          "post",
					URI:             "https://example.com",
					Data: struct {
						NewMembers []string "json:\"newmembers\""
						GroupID    string   "json:\"groupid\""
						Action     string   "json:\"action\""
					}{
						GroupID:    chatgroupID,
						NewMembers: memberIDs,
						Action:     "add_members",
					},
				})
			} else {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(dto.ErrorResponse{
					Duration:         1,
					Timestamp:        1,
					Exception:        "failed",
					Error:            "bad_request",
					ErrorDescription: "bad_request",
				})
			}
		}))
		defer ts.Close()

		agoraClient := newAgoraClientForUnitTest(ts.URL)
		chatgroup, err := agoraClient.AddConversationMembers(&abstract_dto.AddConversationMembersRequest{
			ConversationID:  chatgroupID,
			MemberVendorIDs: memberIDs,
		})
		assert.Equal(t, nil, err)
		assert.Equal(t, chatgroupID, chatgroup.ConversationID)
	})
}

func Test_RemoveConversationMembers(t *testing.T) {
	t.Parallel()

	t.Run("happy case", func(t *testing.T) {
		chatgroupID := "example-chatgroup-id"
		memberIDs := []string{"member-1", "member-2"}
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(dto.RemoveMemberFromChatGroupResponse{
				Timestamp:       1,
				Application:     "app-test",
				ApplicationName: "app-test-name",
				Organization:    "org-test",
				Action:          "post",
				URI:             "https://example.com",
				Data: []struct {
					Result  bool   "json:\"result\""
					GroupID string "json:\"groupid\""
					Reason  string "json:\"reason\""
					Action  string "json:\"action\""
					User    string "json:\"user\""
				}{
					{
						Result:  true,
						GroupID: chatgroupID,
						Action:  "remove_member",
						User:    memberIDs[0],
					},
					{
						Result:  true,
						GroupID: chatgroupID,
						Action:  "remove_member",
						User:    memberIDs[1],
					},
				},
			})
		}))
		defer ts.Close()

		agoraClient := newAgoraClientForUnitTest(ts.URL)
		chatgroup, err := agoraClient.RemoveConversationMembers(&abstract_dto.RemoveConversationMembersRequest{
			ConversationID:  chatgroupID,
			MemberVendorIDs: memberIDs,
		})
		assert.Equal(t, nil, err)
		assert.Equal(t, chatgroupID, chatgroup.ConversationID)
	})

	t.Run("failed some members", func(t *testing.T) {
		chatgroupID := "example-chatgroup-id"
		memberIDs := []string{"member-1", "member-2"}
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(dto.RemoveMemberFromChatGroupResponse{
				Timestamp:       1,
				Application:     "app-test",
				ApplicationName: "app-test-name",
				Organization:    "org-test",
				Action:          "post",
				URI:             "https://example.com",
				Data: []struct {
					Result  bool   "json:\"result\""
					GroupID string "json:\"groupid\""
					Reason  string "json:\"reason\""
					Action  string "json:\"action\""
					User    string "json:\"user\""
				}{
					{
						Result:  true,
						GroupID: chatgroupID,
						Action:  "remove_member",
						User:    memberIDs[0],
					},
					{
						Result:  false,
						GroupID: chatgroupID,
						Action:  "remove_member",
						User:    memberIDs[1],
						Reason:  "This member is not participate in chat group",
					},
				},
			})
		}))
		defer ts.Close()

		expectFailedMembers := []abstract_dto.FailedRemoveMember{
			{
				MemberVendorID: memberIDs[1],
				Reason:         "This member is not participate in chat group",
			},
		}

		agoraClient := newAgoraClientForUnitTest(ts.URL)
		resp, err := agoraClient.RemoveConversationMembers(&abstract_dto.RemoveConversationMembersRequest{
			ConversationID:  chatgroupID,
			MemberVendorIDs: memberIDs,
		})
		assert.Equal(t, nil, err)
		assert.Equal(t, chatgroupID, resp.ConversationID)
		assert.Equal(t, expectFailedMembers, resp.FailedMembers)
	})
}
