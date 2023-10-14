package service

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
)

func TestEcho(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	s := &EchoService{}
	testCases := []utils.TestCase{
		{
			Name: "happy case: userGroupAdmin update user password",
			Ctx:  ctx,
			Req: &pb.EchoRequest{
				Message: "some message",
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
			ExpectedErr: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)
			request := testCase.Req.(*pb.EchoRequest)
			response, err := s.Echo(testCase.Ctx, request)

			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			assert.Equal(t, request.Message, response.Message)
		})
	}
}
