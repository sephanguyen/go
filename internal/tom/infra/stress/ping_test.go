package stress

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/tom/configurations"
	stressmock "github.com/manabie-com/backend/mock/tom/infra/stress"
	legacytpb "github.com/manabie-com/backend/pkg/genproto/tom"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestPing(t *testing.T) {
	t.Parallel()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	mockSvc := &stressmock.GrpcClient{}
	mockIdentityUrl := "https://localhost:8081"
	mockFirebaseToken := "123"
	mockToken := "manabie123"
	mockFirebaseResp := `{"idToken":"123","localId":"456"}`
	mockStream := &stressmock.ClientStream{}
	mockSession := idutil.ULIDNow()
	mockStreamResp := &legacytpb.SubscribeV2Response{Event: &legacytpb.Event{
		Event: &legacytpb.Event_EventPing{
			EventPing: &legacytpb.EventPing{SessionId: mockSession},
		},
	}}
	mockStream.On("Recv").Return(mockStreamResp, nil)
	mockStream.On("CloseSend").Return(nil)
	httpmock.RegisterResponder("POST", mockIdentityUrl,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, mockFirebaseResp), nil
		},
	)

	suite := StagingStress{
		// logger: zap.NewExample().Sugar(),
		logger:          zap.NewNop().Sugar(),
		userModifierSvc: mockSvc,
		bobSvc:          mockSvc,
		chatSvc:         mockSvc,
		runtimeConfig: &StagingStressConfig{
			TotalStudent:               5,
			FirebaseIdentityToolkitURL: mockIdentityUrl,
			ConPerUser:                 1,
		},
	}
	email := idutil.ULIDNow()
	userID := idutil.ULIDNow()
	password := idutil.ULIDNow()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	sampleRes := &upb.CreateStudentResponse{
		StudentProfile: &upb.CreateStudentResponse_StudentProfile{
			Student: &upb.Student{
				Email:       email,
				UserProfile: &upb.UserProfile{UserId: userID},
			},
			StudentPassword: password,
		},
	}
	mockSvc.On("SubscribeV2", mock.Anything, mock.Anything).Return(mockStream, nil)

	mockSvc.On("ExchangeToken", mock.Anything, &bpb.ExchangeTokenRequest{Token: mockFirebaseToken}).Return(&bpb.ExchangeTokenResponse{
		Token: mockToken,
	}, nil)
	mockSvc.On("CreateStudent", mock.Anything, mock.Anything).Times(5).Return(sampleRes, nil)
	mockSvc.On("PingSubscribeV2", mock.Anything, &legacytpb.PingSubscribeV2Request{SessionId: mockSession}).Return(&legacytpb.PingSubscribeV2Response{}, nil)

	fmt.Println("here")
	suite.StartPingStressTest(ctx, &configurations.Config{})
	fmt.Println("here")
	mock.AssertExpectationsForObjects(t, mockSvc)
	fmt.Println("here")
}
