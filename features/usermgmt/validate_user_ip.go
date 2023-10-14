package usermgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"google.golang.org/grpc/metadata"
)

func (s *suite) setUpUserIPAndFeatureConfig(ctx context.Context, ipType, featureAction string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if featureAction == "off" {
		// change resource path to avoid all scenarios access to 1 row in db
		ctx = s.signedIn(ctx, constants.TestingSchool, StaffRoleSchoolAdmin)
	}
	userIP := "user-ip"
	whitelistIP := "whitelist-ip"
	if ipType == "in" {
		userIP = whitelistIP
	}
	ctx = metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "cf-connecting-ip", userIP)
	err := generateExternalConfiguration(ctx, s.BobDBTrace, whitelistIP, featureAction)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func generateExternalConfiguration(ctx context.Context, db database.Ext, whitelistIP, featureAction string) error {
	stmt := `UPDATE external_configuration_value
	SET config_value = $1
	WHERE config_key = $2`
	whitelistConfigValue := fmt.Sprintf(`{"ipv4": [], "ipv6": ["%s"]}`, whitelistIP)
	_, err := db.Exec(ctx, stmt, whitelistConfigValue, constant.KeyIPRestrictionWhitelistConfig)
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, stmt, featureAction, constant.KeyIPRestrictionFeatureConfig)
	if err != nil {
		return err
	}
	return nil
}

func (s *suite) validateUserIPAddress(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &pb.ValidateUserIPRequest{}
	resp, err := pb.NewAuthServiceClient(s.UserMgmtConn).ValidateUserIP(ctx, req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Response = resp
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) assertUserIPValidation(ctx context.Context, permission string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.ValidateUserIPResponse)
	if permission != "allowed" && permission != "not allowed" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not match any condition")
	}
	if permission == "allowed" && !resp.Allow {
		return StepStateToContext(ctx, stepState), fmt.Errorf("user IP should be allowed: %v", resp.Allow)
	}
	if permission == "not allowed" && resp.Allow {
		return StepStateToContext(ctx, stepState), fmt.Errorf("user IP should not be allowed: %v", resp.Allow)
	}
	return StepStateToContext(ctx, stepState), nil
}
