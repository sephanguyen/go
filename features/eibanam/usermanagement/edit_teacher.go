package usermanagement

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/yasuo/constant"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	tom_pb "github.com/manabie-com/backend/pkg/genproto/tom"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/hasura/go-graphql-client"
	"github.com/pkg/errors"
)

type editTeacherFeature struct{}

func (e *editTeacherFeature) createStudentRespFromStack(s *suite) (*ypb.CreateStudentResponse, error) {
	if len(s.ResponseStack.Responses) < 3 {
		return nil, errors.New("can't get create student response")
	}
	createStudentResp := s.ResponseStack.Responses[2].(*ypb.CreateStudentResponse)
	return createStudentResp, nil
}

func (e *editTeacherFeature) updateTeacherReqFromStack(s *suite) (*bpb.UpdateUserProfileRequest, error) {
	updateTeacherReq, err := s.RequestStack.Peek()
	if err != nil {
		return nil, errors.Wrap(err, "Peek()")
	}
	return updateTeacherReq.(*bpb.UpdateUserProfileRequest), nil
}

func (s *suite) schoolAdminHasCreatedATeacher() error {
	err := s.schoolAdminCreatesATeacher()
	if err != nil {
		return errors.Wrap(err, "schoolAdminCreatesATeacher()")
	}

	resp, err := s.ResponseStack.Peek()
	if err != nil {
		return errors.Wrap(err, "Peek()")
	}

	createdTeacherID := resp.(*ypb.CreateUserResponse).Users[0].UserId
	err = s.saveCredential(createdTeacherID, constant.UserGroupTeacher, s.getSchoolId())
	if err != nil {
		return errors.Wrap(err, "saveCredential()")
	}

	return nil
}

func (s *suite) schoolAdminHasCreatedAStudentWithParentInfoAndVisibleCourse() error {
	if err := s.schoolAdminCreatesANewStudentWithParentInfo(); err != nil {
		return err
	}
	if err := s.schoolAdminCreatesANewStudentWithNewParentExistedParentAndVisibleCourse(); err != nil {
		return err
	}

	createStudentResp, err := new(editTeacherFeature).createStudentRespFromStack(s)
	if err != nil {
		return errors.Wrap(err, "createStudentRespFromStack()")
	}
	createdStudentID := createStudentResp.StudentProfile.Student.UserProfile.UserId
	createdParentID := createStudentResp.ParentProfiles[0].Parent.UserProfile.UserId

	if err := s.saveCredential(createdStudentID, constant.UserGroupStudent, s.getSchoolId()); err != nil {
		return errors.Wrap(err, "saveCredential()")
	}
	if err := s.saveCredential(createdParentID, constant.UserGroupParent, s.getSchoolId()); err != nil {
		return errors.Wrap(err, "saveCredential()")
	}

	return nil
}

func (s *suite) schoolAdminEditsTeacherName() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	createTeacherResp := s.ResponseStack.Responses[0].(*ypb.CreateUserResponse)
	createdTeacher := createTeacherResp.Users[0]

	req := &bpb.UpdateUserProfileRequest{
		Profile: &bpb.UserProfile{
			Id:        createdTeacher.UserId,
			Name:      "updated-" + createdTeacher.Name,
			Avatar:    createdTeacher.Avatar,
			UserGroup: createdTeacher.Group.String(),
		},
	}

	s.UserGroupInContext = constant.UserGroupSchoolAdmin
	ctx = contextWithTokenForGrpcCall(s, ctx)

	resp, err :=
		bpb.
			NewUserModifierServiceClient(s.bobConn).
			UpdateUserProfile(ctx, req)
	if err != nil {
		return errors.Wrap(err, "UpdateUserProfile()")
	}
	s.ResponseStack.Push(resp)
	s.RequestStack.Push(req)

	return nil
}

func (s *suite) schoolAdminSeesTheEditedTeacherNameOnCMS() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
		query ($user_id: String!) {
			users(where: {user_id: {_eq: $user_id}}) {
					name
					email
					avatar
					phone_number
					user_group
					country
				}
		}
		`

	if err := addQueryToAllowListForHasuraQuery(bobHasuraAdminUrl+"/v1/query", query); err != nil {
		return errors.Wrap(err, "addQueryToAllowListForHasuraQuery()")
	}

	// Query newly created teacher from hasura
	var profileQuery struct {
		Users []struct {
			Name        string `graphql:"name"`
			Email       string `graphql:"email"`
			Avatar      string `graphql:"avatar"`
			PhoneNumber string `graphql:"phone_number"`
			UserGroup   string `graphql:"user_group"`
			Country     string `graphql:"country"`
		} `graphql:"users(where: {user_id: {_eq: $user_id}})"`
	}

	createTeacherResp := s.ResponseStack.Responses[0].(*ypb.CreateUserResponse)
	createdTeacher := createTeacherResp.Users[0]

	variables := map[string]interface{}{
		"user_id": graphql.String(createdTeacher.UserId),
	}
	err := queryHasura(ctx, &profileQuery, variables, bobHasuraAdminUrl+"/v1/graphql")
	if err != nil {
		return errors.Wrap(err, "queryHasura")
	}

	if len(profileQuery.Users) != 1 {
		return errors.New("failed to query teacher")
	}

	editNameReq, err := s.RequestStack.Peek()
	if err != nil {
		return errors.Wrap(err, "")
	}
	req := editNameReq.(*bpb.UpdateUserProfileRequest)
	if profileQuery.Users[0].Name != req.Profile.Name {
		return errors.New("failed to change teacher name")
	}

	return nil
}

func (s *suite) teacherSeesTheEditedTeacherNameOnTeacherApp() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.UserGroupInContext = constant.UserGroupTeacher
	ctx = contextWithTokenForGrpcCall(s, ctx)

	chatSvc := tom_pb.NewChatServiceClient(s.tomConn)
	streamV2, err := chatSvc.SubscribeV2(ctx, &tom_pb.SubscribeV2Request{})
	if err != nil {
		return err
	}

	sessionID := ""
	resp, err := streamV2.Recv()
	if err != nil {
		return err
	}

	if resp.Event.GetEventPing() == nil {
		return fmt.Errorf("stream must receive pingEvent first")
	}
	sessionID = resp.Event.GetEventPing().SessionId

	firstPing := make(chan error, 1)
	pingFunc := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		s.UserGroupInContext = constant.UserGroupTeacher
		ctx = contextWithTokenForGrpcCall(s, ctx)
		_, err := chatSvc.PingSubscribeV2(ctx, &tom_pb.PingSubscribeV2Request{SessionId: sessionID})
		if err != nil {
			return errors.Wrap(err, "PingSubscribeV2()")
		}
		return nil
	}
	go func(ctx context.Context) {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				select {
				case firstPing <- pingFunc():
				default:
				}
			}
		}
	}(ctx)

	if err := <-firstPing; err != nil {
		return err
	}

	listConversationsInSchoolReq := &tpb.ListConversationsInSchoolRequest{
		Paging: &cpb.Paging{
			Limit: 100,
		},
	}

	res, err := tpb.NewChatReaderServiceClient(s.tomConn).ListConversationsInSchool(ctx, listConversationsInSchoolReq)
	if err != nil {
		return errors.Wrap(err, "ListConversationsInSchool()")
	}

	createStudentResp, err := new(editTeacherFeature).createStudentRespFromStack(s)
	if err != nil {
		return errors.Wrap(err, "createStudentRespFromStack()")
	}
	createdStudentId := createStudentResp.StudentProfile.Student.UserProfile.UserId
	createdParentId := createStudentResp.ParentProfiles[0].Parent.UserProfile.UserId

	for _, item := range res.Items {
		if item.ConversationType.String() == "MESSAGE_TYPE_SYSTEM" {
			continue
		}
		if createdStudentId == item.StudentId || createdParentId == item.StudentId {
			joinReq := &tpb.JoinConversationsRequest{
				ConversationIds: []string{item.ConversationId},
			}
			_, err := tpb.NewChatModifierServiceClient(s.tomConn).JoinConversations(ctx, joinReq)
			if err != nil {
				return errors.Wrap(err, "JoinConversations()")
			}

			req := &tom_pb.SendMessageRequest{
				ConversationId: item.ConversationId,
				Type:           tom_pb.MESSAGE_TYPE_TEXT,
				Message:        "test message",
			}
			_, err =
				tom_pb.
					NewChatServiceClient(s.tomConn).
					SendMessage(ctx, req)
			if err != nil {
				return errors.Wrap(err, "SendMessage()")
			}
		}
	}

	return nil
}

func (s *suite) getTeacherInfoInConversation(ctx context.Context, searchUserId string) (*bob_pb.BasicProfile, error) {
	req := &tom_pb.ConversationListRequest{
		Limit: 100,
	}
	resp, err :=
		tom_pb.
			NewChatServiceClient(s.tomConn).
			ConversationList(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "ConversationList()")
	}

	for _, conversation := range resp.Conversations {
		for _, user := range conversation.Users {
			if user.Id != searchUserId {
				continue
			}
			req := &bob_pb.GetBasicProfileRequest{
				UserIds: []string{user.Id},
			}
			resp, err := bob_pb.NewUserServiceClient(s.bobConn).GetBasicProfile(ctx, req)
			if err != nil {
				return nil, errors.Wrap(err, "GetBasicProfile()")
			}
			return resp.Profiles[0], nil
		}
	}
	return nil, fmt.Errorf("can not found user with id: %s", searchUserId)
}

func (s *suite) studentSeesTheEditedTeacherNameOnLearnerApp() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.UserGroupInContext = constant.UserGroupStudent
	ctx = contextWithTokenForGrpcCall(s, ctx)

	updateTeacherReq, err := new(editTeacherFeature).updateTeacherReqFromStack(s)
	if err != nil {
		return errors.Wrap(err, "updateTeacherReqFromStack()")
	}
	requestedUpdateTeacherId := updateTeacherReq.Profile.Id
	requestedUpdateTeacherName := updateTeacherReq.Profile.Name

	profile, err := s.getTeacherInfoInConversation(ctx, requestedUpdateTeacherId)
	if err != nil {
		return errors.Wrap(err, "getTeacherInfoInConversation()")
	}

	if profile.Name != requestedUpdateTeacherName {
		return fmt.Errorf(`expected teacher name: "%v", actual is "%v"`, requestedUpdateTeacherName, profile.Name)
	}

	return nil
}

func (s *suite) parentSeesTheEditedTeacherNameOnLearnerApp() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.UserGroupInContext = constant.UserGroupParent
	ctx = contextWithTokenForGrpcCall(s, ctx)

	updateTeacherReq, err := new(editTeacherFeature).updateTeacherReqFromStack(s)
	if err != nil {
		return errors.Wrap(err, "updateTeacherReqFromStack()")
	}
	requestedUpdateTeacherId := updateTeacherReq.Profile.Id
	requestedUpdateTeacherName := updateTeacherReq.Profile.Name

	profile, err := s.getTeacherInfoInConversation(ctx, requestedUpdateTeacherId)
	if err != nil {
		return errors.Wrap(err, "getTeacherInfoInConversation()")
	}

	if profile.Name != requestedUpdateTeacherName {
		return fmt.Errorf(`expected teacher name: "%v", actual is "%v"`, requestedUpdateTeacherName, profile.Name)
	}

	return nil
}
