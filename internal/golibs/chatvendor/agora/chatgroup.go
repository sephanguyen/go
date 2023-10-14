package agora

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/chatvendor/agora/dto"
	abstract_dto "github.com/manabie-com/backend/internal/golibs/chatvendor/dto"
	"github.com/manabie-com/backend/internal/golibs/idutil"
)

func (a *agoraClientImpl) CreateConversation(req *abstract_dto.CreateConversationRequest) (*abstract_dto.CreateConversationResponse, error) {
	if req.OwnerVendorID == "" || len(req.MemberVendorIDs) == 0 {
		return nil, fmt.Errorf("missing owner of chatgroup or missing members")
	}

	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	reqBody, err := json.Marshal(&dto.CreateChatGroupRequest{
		GroupName: idutil.ULIDNow(),
		Desc:      "Conversation",
		Public:    false,
		MaxUsers:  MaxUsersOnChatGroup,
		Owner:     req.OwnerVendorID,
		Members:   req.MemberVendorIDs,
	})
	if err != nil {
		return nil, err
	}

	// Create chatgroup endpoint: POST /chatgroups
	endpoint := string(ChatGroups)

	createChatGroupResponse := &dto.CreateChatGroupResponse{}
	err = a.doRequest(ctx, MethodPost, endpoint, GetAgoraCommonHeader(), bytes.NewBuffer(reqBody), createChatGroupResponse)
	if err != nil {
		return nil, err
	}

	if createChatGroupResponse.Data.GroupID != "" {
		return &abstract_dto.CreateConversationResponse{
			ConversationID: createChatGroupResponse.Data.GroupID,
		}, nil
	}

	return nil, fmt.Errorf("[agora]: cannot create agora chat group")
}

func (a *agoraClientImpl) AddConversationMembers(req *abstract_dto.AddConversationMembersRequest) (*abstract_dto.AddConversationMembersResponse, error) {
	if req.ConversationID == "" || len(req.MemberVendorIDs) == 0 {
		return nil, fmt.Errorf("invalid request AddConversationMembers")
	}

	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	reqBody, err := json.Marshal(&dto.AddMemberToChatGroupRequest{
		Usernames: req.MemberVendorIDs,
	})
	if err != nil {
		return nil, err
	}

	// endpoint: POST /chatgroups/{chatgroupid}/users
	endpoint := string(ChatGroups) + "/" + req.ConversationID + "/users"

	addMemberToChatGroupResponse := &dto.AddMemberToChatGroupResponse{}
	err = a.doRequest(ctx, MethodPost, endpoint, GetAgoraCommonHeader(), bytes.NewBuffer(reqBody), addMemberToChatGroupResponse)
	if err != nil {
		return nil, err
	}

	if addMemberToChatGroupResponse.Data.GroupID != "" && len(addMemberToChatGroupResponse.Data.NewMembers) != 0 {
		return &abstract_dto.AddConversationMembersResponse{
			ConversationID: addMemberToChatGroupResponse.Data.GroupID,
		}, nil
	}
	return nil, fmt.Errorf("[agora]: failed add member to chat group")
}

func (a *agoraClientImpl) RemoveConversationMembers(req *abstract_dto.RemoveConversationMembersRequest) (*abstract_dto.RemoveConversationMembersResponse, error) {
	if req.ConversationID == "" || len(req.MemberVendorIDs) == 0 {
		return nil, fmt.Errorf("invalid request RemoveConversationMembers")
	}

	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	// endpoint: DELETE /chatgroups/{chatgroupid}/users/{member usernames separated by commas}
	endpoint := string(ChatGroups) + "/" + req.ConversationID + "/users/" + strings.Join(req.MemberVendorIDs, ",")

	removeMemberFromChatGroupResponse := &dto.RemoveMemberFromChatGroupResponse{}
	err := a.doRequest(ctx, MethodDel, endpoint, GetAgoraCommonHeader(), bytes.NewBuffer([]byte(`{}`)), removeMemberFromChatGroupResponse)
	if err != nil {
		return nil, err
	}

	removedMembers, err := removeMemberFromChatGroupResponse.GetRemoveMembers()
	if err != nil {
		return nil, fmt.Errorf("failed GetRemoveMember: %+v", err)
	}

	if len(removedMembers) > 0 {
		conversationID := removedMembers[0].GroupID
		failedRemoveMembers := []abstract_dto.FailedRemoveMember{}
		for _, removedMember := range removedMembers {
			if !removedMember.Result {
				failedRemoveMembers = append(failedRemoveMembers, abstract_dto.FailedRemoveMember{
					MemberVendorID: removedMember.User,
					Reason:         removedMember.Reason,
				})
			}
		}
		return &abstract_dto.RemoveConversationMembersResponse{
			ConversationID: conversationID,
			FailedMembers:  failedRemoveMembers,
		}, nil
	}

	return nil, fmt.Errorf("[agora]: failed remove member from chat group")
}
