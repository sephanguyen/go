package virtualclassroom

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) userGetsUserInformation(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	time.Sleep(5 * time.Second)

	userIDs := stepState.TeacherIDs
	userIDs = append(userIDs, stepState.StudentIds...)
	req := &vpb.GetUserInformationRequest{
		UserIds: userIDs,
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomReaderServiceClient(s.VirtualClassroomConn).
		GetUserInformation(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userReceivesExpectedUserInformation(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	request := stepState.Request.(*vpb.GetUserInformationRequest)
	response := stepState.Response.(*vpb.GetUserInformationResponse)

	if len(response.UserInfos) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting user IDs but got 0")
	}
	expectedUserIDs := request.GetUserIds()
	actualUserIDs := make([]string, 0, len(response.UserInfos))

	for _, userInfo := range response.UserInfos {
		actualUserIDs = append(actualUserIDs, userInfo.UserId)
	}

	if !sliceutils.UnorderedEqual(expectedUserIDs, actualUserIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the user IDs received %s is not equal to the expected user IDs %s", actualUserIDs, expectedUserIDs)
	}

	return StepStateToContext(ctx, stepState), nil
}
