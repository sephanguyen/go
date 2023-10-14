package agora

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/chatvendor"
	"github.com/manabie-com/backend/internal/golibs/chatvendor/dto"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"go.uber.org/zap"
)

const (
	DataMockAppID          = "970CA35de60c44645bbae8a215061b33"
	DataMockAppCertificate = "5CFd2fd1755d40ecb72977518be15d3b"
	DataMockAppName        = "1234567"
	DataMockOrgName        = "87654321"
	DataMockUserUUID       = "2882341273"
)

// nolint
func newAgoraClientForUnitTest(mockRestAPI string) *agoraClientImpl {
	return &agoraClientImpl{
		AgoraConfig: configs.AgoraConfig{
			AppID:              DataMockAppID,
			PrimaryCertificate: DataMockAppCertificate,
			AppName:            DataMockAppName,
			OrgName:            DataMockOrgName,
			RestAPI:            mockRestAPI,
		},
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).Dial,
				MaxIdleConnsPerHost: 5,
			},
		},
		logger: zap.NewExample(),
	}
}

// Mock client for integration
type Mock struct{}

func NewMockAgoraClient() chatvendor.ChatVendorClient {
	return &Mock{}
}

func (*Mock) GetAppToken() (string, error) {
	return "agora-app-token", nil
}

func (*Mock) GetUserToken(userID string) (string, uint64, error) {
	return userID + "-user-token", 1000, nil
}

func (*Mock) GetAppKey() string {
	return "agora-app-key"
}

// TODO: implement more mock logic
func (*Mock) GetUser(_ *dto.GetUserRequest) (*dto.GetUserResponse, error) {
	return &dto.GetUserResponse{}, nil
}

// TODO: implement more mock logic
func (*Mock) CreateUser(_ *dto.CreateUserRequest) (*dto.CreateUserResponse, error) {
	return &dto.CreateUserResponse{}, nil
}

// TODO: implement more mock logic
func (*Mock) CreateConversation(_ *dto.CreateConversationRequest) (*dto.CreateConversationResponse, error) {
	return &dto.CreateConversationResponse{
		ConversationID: idutil.ULIDNow(),
	}, nil
}

// TODO: implement more mock logic
func (*Mock) AddConversationMembers(req *dto.AddConversationMembersRequest) (*dto.AddConversationMembersResponse, error) {
	return &dto.AddConversationMembersResponse{
		ConversationID: req.ConversationID,
	}, nil
}

func (*Mock) RemoveConversationMembers(req *dto.RemoveConversationMembersRequest) (*dto.RemoveConversationMembersResponse, error) {
	failedMembers := []dto.FailedRemoveMember{}
	// check if the member ID has keyword "failed" and return it as failed member
	for _, reqMember := range req.MemberVendorIDs {
		if strings.Contains(reqMember, "failed") {
			failedMembers = append(failedMembers, dto.FailedRemoveMember{
				MemberVendorID: reqMember,
				Reason:         "some reason",
			})
		}
	}
	return &dto.RemoveConversationMembersResponse{
		ConversationID: req.ConversationID,
		FailedMembers:  failedMembers,
	}, nil
}

func (*Mock) DeleteMessage(_ *dto.DeleteMessageRequest) (*dto.DeleteMessageResponse, error) {
	return &dto.DeleteMessageResponse{
		IsSuccess: true,
	}, nil
}
