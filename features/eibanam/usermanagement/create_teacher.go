package usermanagement

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/features/helper"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	common_pbv1 "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	usermgmt_pbv1 "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
	yasuo_pbv1 "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/hasura/go-graphql-client"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type createTeacherFeature struct{}

func (createTeacherFeature *createTeacherFeature) reqWithTeacherInfo(schoolId int64) *yasuo_pbv1.CreateUserRequest {
	randomId := idutil.ULIDNow()
	return &yasuo_pbv1.CreateUserRequest{
		Users: []*yasuo_pbv1.CreateUserProfile{
			{
				Name:        fmt.Sprintf("new-teacher-%s", randomId),
				Country:     common_pbv1.Country_COUNTRY_VN,
				PhoneNumber: fmt.Sprintf("new-teacher-phone-number-%s", randomId),
				Email:       fmt.Sprintf("new-teacher-email-%s@example.com", randomId),
				Avatar:      fmt.Sprintf("new-teacher-avatar-url-%s", randomId),
			},
		},
		UserGroup:    common_pbv1.UserGroup_USER_GROUP_TEACHER,
		SchoolId:     schoolId,
		Organization: strconv.Itoa(int(schoolId)),
	}
}

func (s *suite) schoolAdminCreatesATeacher() error {
	// Setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.UserGroupInContext = constant.UserGroupSchoolAdmin
	ctx = contextWithTokenForGrpcCall(s, ctx)

	req := new(createTeacherFeature).reqWithTeacherInfo(s.getSchoolId())

	// Create new teacher using
	// yasuo v1 CreateUser api
	resp, err :=
		yasuo_pbv1.
			NewUserModifierServiceClient(s.yasuoConn).
			CreateUser(ctx, req)
	if err != nil {
		return err
	}
	s.ResponseStack.Push(resp)
	s.RequestStack.Push(req)

	return nil
}

func (s *suite) schoolAdminSeesNewlyCreatedTeacherOnCMS() error {
	// Setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.UserGroupInContext = constant.UserGroupSchoolAdmin
	ctx = contextWithToken(s, ctx)

	// Pre-setup for hasura query using admin secret
	if err := trackTableForHasuraQuery(bobHasuraAdminUrl+"/v1/query", "teachers", "users", "users_groups"); err != nil {
		return errors.Wrap(err, "trackTableForHasuraQuery()")
	}
	if err := createSelectPermissionForHasuraQuery(bobHasuraAdminUrl+"/v1/query", "teachers", "users", "users_groups"); err != nil {
		return errors.Wrap(err, "createSelectPermissionForHasuraQuery()")
	}

	query :=
		`
		query ($userID: String!){
			teachers(where: {teacher_id: {_eq: $userID}}) {
				school_ids
			}
			users(where: {user_id: {_eq: $userID}}) {
				user_id
				email
				name
				country
				phone_number
				avatar
				user_group
			}
			users_groups(where: {user_id: {_eq: $userID}}) {
				user_id
				group_id
			}
		  }
		`
	if err := addQueryToAllowListForHasuraQuery(bobHasuraAdminUrl+"/v1/query", query); err != nil {
		return errors.Wrap(err, "addQueryToAllowListForHasuraQuery()")
	}

	// Query newly created teacher from hasura
	var profileQuery struct {
		Teachers []struct {
			SchoolIDs []int32 `graphql:"school_ids"`
		} `graphql:"teachers(where: {teacher_id: {_eq: $userID}})"`
		Users []struct {
			UserID      string `graphql:"user_id"`
			Email       string `graphql:"email"`
			Name        string `graphql:"name"`
			Country     string `graphql:"country"`
			PhoneNumber string `graphql:"phone_number"`
			Avatar      string `graphql:"avatar"`
			UserGroup   string `graphql:"user_group"`
		} `graphql:"users(where: {user_id: {_eq: $userID}})"`
		UserGroups []struct {
			UserID  string `graphql:"user_id"`
			GroupID string `graphql:"group_id"`
		} `graphql:"users_groups(where: {user_id: {_eq: $userID}})"`
	}
	resp, err := s.ResponseStack.Peek()
	if err != nil {
		return errors.Wrap(err, "s.ResponseStack.Peek()")
	}

	if len(resp.(*yasuo_pbv1.CreateUserResponse).Users) <= 0 {
		return errors.New("failed to create teacher")
	}
	userID := resp.(*yasuo_pbv1.CreateUserResponse).Users[0].UserId

	variables := map[string]interface{}{
		"userID": graphql.String(userID),
	}
	err = queryHasura(ctx, &profileQuery, variables, bobHasuraAdminUrl+"/v1/graphql")
	if err != nil {
		return errors.Wrap(err, "queryHasura")
	}

	if len(profileQuery.Users) <= 0 {
		return errors.New("failed to query teacher")
	}
	if len(profileQuery.UserGroups) <= 0 {
		return errors.New("failed to query user group")
	}
	createdTeacher := profileQuery.Users[0]

	iReq, err := s.RequestStack.Peek()
	if err != nil {
		return errors.Wrap(err, "s.RequestStack.Peek()")
	}
	req := iReq.(*yasuo_pbv1.CreateUserRequest)
	requestedTeacher := req.Users[0]
	switch {
	case createdTeacher.Name != requestedTeacher.Name:
		return fmt.Errorf(`expect created teacher has "name": %v but actual is %v`, requestedTeacher.Name, createdTeacher.Name)
	case createdTeacher.Country != requestedTeacher.Country.String():
		return fmt.Errorf(`expect created teacher has "country": %v but actual is %v`, requestedTeacher.Country.String(), createdTeacher.Country)
	case createdTeacher.PhoneNumber != requestedTeacher.PhoneNumber:
		return fmt.Errorf(`expect created teacher has "phone number": %v but actual is %v`, requestedTeacher.PhoneNumber, createdTeacher.PhoneNumber)
	case createdTeacher.Email != requestedTeacher.Email:
		return fmt.Errorf(`expect created teacher has "email": %v but actual is %v`, requestedTeacher.Email, createdTeacher.Email)
	case createdTeacher.Avatar != requestedTeacher.Avatar:
		return fmt.Errorf(`expect created teacher has "avatar": %v but actual is %v`, requestedTeacher.Avatar, createdTeacher.Avatar)
	case createdTeacher.UserGroup != common_pbv1.UserGroup_USER_GROUP_TEACHER.String():
		return fmt.Errorf(`expect created teacher has "avatar": %v but actual is %v`, common_pbv1.UserGroup_USER_GROUP_TEACHER.String(), createdTeacher.UserGroup)
	}

	if len(resp.(*yasuo_pbv1.CreateUserResponse).Users) <= 0 {
		return errors.New("failed to create teacher")
	}
	respondedTeacher := resp.(*yasuo_pbv1.CreateUserResponse).Users[0]
	switch {
	case respondedTeacher.UserId == "":
		return fmt.Errorf(`expect responded teacher has "id": %v but actual is nil`, respondedTeacher.UserId)
	case createdTeacher.Name != respondedTeacher.Name:
		return fmt.Errorf(`expect responded teacher has "name": %v but actual is %v`, createdTeacher.Name, respondedTeacher.Name)
	case createdTeacher.Avatar != respondedTeacher.Avatar:
		return fmt.Errorf(`expect responded teacher has "avatar": %v but actual is %v`, createdTeacher.Avatar, respondedTeacher.Avatar)
	case createdTeacher.UserGroup != respondedTeacher.Group.String():
		return fmt.Errorf(`expect responded teacher has "user group": %v but actual is %v`, createdTeacher.UserGroup, respondedTeacher.Group.String())
	}

	user := new(bob_entities.User)
	database.AllNullEntity(user)
	err = multierr.Combine(
		user.ID.Set(createdTeacher.UserID),
		user.Country.Set(createdTeacher.Country),
		user.PhoneNumber.Set(createdTeacher.PhoneNumber),
		user.Email.Set(createdTeacher.Email),
		user.Avatar.Set(createdTeacher.Avatar),
	)
	if err != nil {
		return err
	}
	s.User = user

	return nil
}

func (s *suite) teacherLoginsTeacherAppSuccessfullyAfterForgotPassword() error {
	// Setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.UserGroupInContext = constant.UserGroupSchoolAdmin
	ctx = contextWithTokenForGrpcCall(s, ctx)

	createdTeacher := s.User
	newPassword := idutil.ULIDNow()
	req := &usermgmt_pbv1.ReissueUserPasswordRequest{
		UserId:      createdTeacher.ID.String,
		NewPassword: newPassword,
	}

	_, err := usermgmt_pbv1.NewUserModifierServiceClient(s.userMgmtConn).ReissueUserPassword(ctx, req)
	if err != nil {
		return errors.Wrap(err, "ReissueUserPassword()")
	}

	token, err := loginFirebaseAccount(ctx, s.Config.FirebaseAPIKey, createdTeacher.Email.String, newPassword)
	if err != nil {
		return errors.Wrap(err, "loginFirebaseAccount()")
	}
	//fmt.Println(token)

	// Exchange teacher token
	token, err = helper.ExchangeToken(token, createdTeacher.ID.String, constant.UserGroupTeacher, applicantID, s.getSchoolId(), shamirConn)
	if err != nil {
		return errors.Wrap(err, "helper.ExchangeToken()")
	}
	s.UserGroupCredentials[constant.UserGroupTeacher] = &userCredential{
		UserID:    createdTeacher.ID.String,
		AuthToken: token,
		UserGroup: constant.UserGroupTeacher,
	}
	//fmt.Println(token)

	return nil
}
