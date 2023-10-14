package usermgmt

import (
	"context"
	"fmt"
	"time"

	golibs_auth "github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) aSignedInUserWith(ctx context.Context, kindOfUser, role string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, role)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "signedAsAccount")
	}

	token, err := jwt.ParseString(stepState.AuthToken)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	schoolID := int64(stepState.CurrentSchoolID)
	if schoolID == 0 {
		schoolID = constants.ManabieSchool
	}

	userID := token.Subject()
	switch kindOfUser {
	case "has signed in before":
		query := "UPDATE users SET last_login_date = $1 WHERE user_id = $2"
		if _, err := s.BobDBTrace.DB.Exec(golibs_auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID)), query, time.Now().UTC().Add(-time.Hour), &userID); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "newly created":
		break
	default:
		return StepStateToContext(ctx, stepState), errors.New("not supported scenario step")
	}

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) userLastLoginDate(ctx context.Context, updatedResult string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	token, err := jwt.ParseString(stepState.AuthToken)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	userID := token.Subject()
	reqLastLoginDate := stepState.Request.(*pb.UpdateUserLastLoginDateRequest).LastLoginDate
	updatedLastLoginDate := &time.Time{}

	schoolID := int64(stepState.CurrentSchoolID)
	if schoolID == 0 {
		schoolID = constants.ManabieSchool
	}

	query := "SELECT last_login_date FROM users WHERE user_id = $1"
	err = s.BobDBTrace.DB.QueryRow(golibs_auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID)), query, &userID).Scan(&updatedLastLoginDate)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	switch updatedResult {
	case "is updated":
		if !reqLastLoginDate.AsTime().Equal(*updatedLastLoginDate) {
			return StepStateToContext(ctx, stepState), errors.New("expect user last login date updated, but it's not")
		}
	case "is not updated":
		if (updatedLastLoginDate == nil && reqLastLoginDate != nil && !reqLastLoginDate.AsTime().IsZero()) ||
			(updatedLastLoginDate != nil && reqLastLoginDate.AsTime().Equal(*updatedLastLoginDate)) {
			return StepStateToContext(ctx, stepState), errors.New("expect user last login date not updated, but it's updated")
		}
	default:
		return StepStateToContext(ctx, stepState), errors.New("not supported scenario step")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpdateLastLoginDateWithValue(ctx context.Context, valueCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch valueCondition {
	case "valid":
		stepState.Request = &pb.UpdateUserLastLoginDateRequest{
			LastLoginDate: timestamppb.New(time.Now().UTC().Round(time.Millisecond)),
		}
	case "missing":
		stepState.Request = &pb.UpdateUserLastLoginDateRequest{}
	case "zero":
		stepState.Request = &pb.UpdateUserLastLoginDateRequest{
			LastLoginDate: timestamppb.New(time.Time{}),
		}
	default:
		return StepStateToContext(ctx, stepState), errors.New("not supported scenario step")
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewUserModifierServiceClient(s.UserMgmtConn).
		UpdateUserLastLoginDate(contextWithToken(ctx), stepState.Request.(*pb.UpdateUserLastLoginDateRequest))

	return StepStateToContext(ctx, stepState), nil
}
