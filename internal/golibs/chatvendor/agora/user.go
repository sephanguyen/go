package agora

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/chatvendor/agora/dto"
	abstract_dto "github.com/manabie-com/backend/internal/golibs/chatvendor/dto"
	"github.com/manabie-com/backend/internal/golibs/chatvendor/entities"
)

// Ref: https://docs.agora.io/en/agora-chat/restful-api/user-system-registration

func (a *agoraClientImpl) GetUser(req *abstract_dto.GetUserRequest) (*abstract_dto.GetUserResponse, error) {
	if req.VendorUserID == "" {
		return nil, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	// Get User endpoint: GET /users/{{username}}
	endpoint := fmt.Sprintf("%s/%s", Users, req.VendorUserID)

	getUserResponse := &dto.GetUserResponse{}
	err := a.doRequest(ctx, MethodGet, endpoint, GetAgoraCommonHeader(), nil, getUserResponse)
	if err != nil {
		return nil, err
	}

	if len(getUserResponse.Entities) > 0 {
		firstUser := getUserResponse.Entities[0]
		return &abstract_dto.GetUserResponse{
			User: entities.User{
				UserID:       firstUser.NickName,
				VendorUserID: firstUser.UserName,
				Activated:    firstUser.Activated,
				UpdatedAt:    firstUser.Modified,
				CreatedAt:    firstUser.Created,
			},
		}, nil
	}

	return nil, nil
}

func (a *agoraClientImpl) CreateUser(req *abstract_dto.CreateUserRequest) (*abstract_dto.CreateUserResponse, error) {
	if req.VendorUserID == "" || req.UserID == "" {
		return nil, fmt.Errorf("missing manabie_user_id and agora_user_id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	reqBody, err := json.Marshal(&dto.CreateUserRequest{
		UserName: req.VendorUserID,
		UserID:   req.UserID,
		Password: GetAgoraUserPassword(req.UserID, req.VendorUserID),
	})
	if err != nil {
		return nil, err
	}

	// Create User endpoint: POST /users
	endpoint := string(Users)

	createUserResponse := &dto.CreateUserResponse{}
	err = a.doRequest(ctx, MethodPost, endpoint, GetAgoraCommonHeader(), bytes.NewBuffer(reqBody), createUserResponse)
	if err != nil {
		return nil, err
	}

	if len(createUserResponse.Entities) > 0 {
		firstUser := createUserResponse.Entities[0]
		return &abstract_dto.CreateUserResponse{
			User: entities.User{
				UserID:       firstUser.NickName,
				VendorUserID: firstUser.UserName,
				Activated:    firstUser.Activated,
				UpdatedAt:    firstUser.Modified,
				CreatedAt:    firstUser.Created,
			},
		}, nil
	}

	return nil, fmt.Errorf("[agora]: cannot create agora user")
}
