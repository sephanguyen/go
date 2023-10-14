package usermgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	pkg_unleash "github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
)

func (s *suite) getUserAuthInfoByLoginEmailAndDomainName(ctx context.Context, usernameCondition, domainNameCondition string) (context.Context, error) {
	username, domainName, err := s.initUsernameAndDomainNameByConditions(ctx, usernameCondition, domainNameCondition)
	if err != nil {
		return ctx, errors.Wrap(err, "failed to init username and domain name")
	}

	req := &pb.GetAuthInfoRequest{Username: username, DomainName: domainName}
	// use empty context to make sure that the request is sent by the user who is not signed in
	emptyContext := context.TODO()
	resp, err := pb.NewAuthServiceClient(s.UserMgmtConn).GetAuthInfo(emptyContext, req)

	s.Request = req
	s.Response = resp
	s.ResponseErr = err

	return ctx, nil
}

func (s *suite) userReceivesLoginEmailAndTenantIDSuccessfully(ctx context.Context) (context.Context, error) {
	request := s.Request.(*pb.GetAuthInfoRequest)
	response := s.Response.(*pb.GetAuthInfoResponse)

	signedInCtx, err := s.signedAsAccount(ctx, StaffRoleSchoolAdmin)
	if err != nil {
		return ctx, errors.Wrap(err, "failed to sign in as school admin")
	}

	isUserNameEnable, err := s.isUserNameEnable(signedInCtx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check username enable")
	}

	loginEmail := field.NewNullString()
	tenantID := field.NewNullString()

	var query string

	// double check if feature toggle and config are enabled
	if isUserNameEnable {
		// if feature toggle is enabled and config is enabled, we gonna use username to get auth info
		query = `
			SELECT
				organizations.tenant_id,
				users.login_email

			FROM organizations
			JOIN users
				ON organizations.resource_path = users.resource_path

			WHERE
				  username = $1
			  AND domain_name = $2
		`
	} else {
		// if feature toggle is disabled or config is disabled, we gonna use email to get auth info
		query = `
			SELECT
				organizations.tenant_id,
				users.login_email

			FROM organizations
			JOIN users
				ON organizations.resource_path = users.resource_path

			WHERE
				  email = $1
			  AND domain_name = $2
		`
	}

	err = s.BobDB.
		QueryRow(signedInCtx, query, request.GetUsername(), request.GetDomainName()).
		Scan(&tenantID, &loginEmail)
	if err != nil {
		return ctx, errors.Wrap(err, "failed to query user auth info")
	}

	switch {
	case tenantID.String() != response.GetTenantId():
		return ctx, fmt.Errorf("tenantID not match, expected: %s, got: %s", tenantID.String(), response.GetTenantId())

	case loginEmail.String() != response.GetLoginEmail():
		return ctx, fmt.Errorf("loginEmail not match, expected: %s, got: %s", loginEmail.String(), response.GetLoginEmail())
	}

	return ctx, nil
}

func (s *suite) isUserNameEnable(ctx context.Context) (bool, error) {
	// check feature toggle of username student parent
	enabledFeatureToggleUsernameStudentParent, err := isFeatureToggleEnabled(ctx, s.UnleashSuite, pkg_unleash.FeatureToggleUserNameStudentParent)
	if err != nil {
		return false, errors.Wrap(err, "failed to check feature toggle")
	}

	// check feature toggle of username staff
	enabledFeatureToggleUsernameStaff, err := isFeatureToggleEnabled(ctx, s.UnleashSuite, pkg_unleash.FeatureToggleStaffUsername)
	if err != nil {
		return false, errors.Wrap(err, "failed to check feature toggle")
	}

	// check config of username
	config, err := new(repository.DomainInternalConfigurationRepo).GetByKey(ctx, s.BobDB, constant.KeyAuthUsernameConfig)
	if err != nil {
		return false, errors.Wrap(err, "failed to get config")
	}

	enableConfigUsername := config.ConfigValue().String() == constant.ConfigValueOn
	return enabledFeatureToggleUsernameStaff && enabledFeatureToggleUsernameStudentParent && enableConfigUsername, nil
}
