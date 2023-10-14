package managing

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"time"

	"github.com/jackc/pgtype"
	"github.com/lestrrat-go/jwx/jwt"
	"go.uber.org/multierr"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	bobEnt "github.com/manabie-com/backend/internal/bob/entities"
	bobRepo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	tomRepo "github.com/manabie-com/backend/internal/tom/repositories"
	bobPb "github.com/manabie-com/backend/pkg/genproto/bob"
	tomPb "github.com/manabie-com/backend/pkg/genproto/tom"
)

func (s *suite) anAuthenticatedAdmin(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	adminID, adminToken, err := s.signedInAs(ctx, bobEnt.UserGroupAdmin)
	if err != nil {
		return GandalfStepStateToContext(ctx, stepState), err

	}
	stepState.GandalfStateCurrentUserID = adminID
	stepState.GandalfStateAuthToken = adminToken
	return GandalfStepStateToContext(ctx, stepState), nil
}

func (s *suite) aSchool(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	stepState.GandalfStateSchool = s.createSchool("S1", "COUNTRY_VN", "Ho Chi Minh", "2")
	return s.adminInsertsSchools(ctx)
}

func (s *suite) aValidClassInformation(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	numberOfTeacher := 1
	request := s.createClassRequest(stepState.GandalfStateSchool.ID.Int)
	ctx, err := s.assignOwnersToCreateClassRequest(ctx, request, numberOfTeacher)
	stepState.GandalfStateRequest = request
	return GandalfStepStateToContext(ctx, stepState), err
}

func (s *suite) anAdminCreatesANewClass(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	stepState.GandalfStateRequestSentAt = time.Now()
	stepState.GandalfStateResponse, stepState.GandalfStateResponseErr = bobPb.NewClassClient(s.bobConn).CreateClass(contextWithToken(ctx, stepState.GandalfStateAuthToken), stepState.GandalfStateRequest.(*bobPb.CreateClassRequest))
	time.Sleep(time.Second)
	return GandalfStepStateToContext(ctx, stepState), nil
}

func (s *suite) thatClassMustBeCreatedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	if ctx, err := s.validResponse(ctx, codes.OK); err != nil {
		return ctx, err
	}
	stepState.GandalfStateCurrentClassID = stepState.GandalfStateResponse.(*bobPb.CreateClassResponse).ClassId
	ctx, err := s.classMustHaveMembers(ctx,
		len(stepState.GandalfStateTeacherIDsMap),
		bobEnt.UserGroupTeacher,
		"true",
		"CLASS_MEMBER_STATUS_ACTIVE")
	return GandalfStepStateToContext(ctx, stepState), err
}

func (s *suite) aChatRoomAssociatedWithThatClassMustBeCreated(ctx context.Context) (context.Context, error) {
	err := try.Do(func(attempt int) (retry bool, err error) {
		ctx, err1 := s.getConversationByClass(ctx)
		ctx, err2 := s.validResponse(ctx, codes.OK)
		err = multierr.Combine(err1, err2)
		if err != nil {
			time.Sleep(2 * time.Second)
			return true, err
		}
		return false, nil
	})
	if err != nil {
		return ctx, err

	}

	stepState := GandalfStepStateFromContext(ctx)
	resp := stepState.GandalfStateResponse.(*tomPb.ConversationByClassResponse)
	if len(resp.Conversations) == 0 {
		return ctx, fmt.Errorf("response have no conversation")
	}
	if resp.Conversations[0].ClassId != uint32(stepState.GandalfStateCurrentClassID) {
		return ctx, fmt.Errorf("expected classID %v but get classID %v from response", stepState.GandalfStateCurrentClassID, resp.Conversations[0].ClassId)
	}
	return ctx, nil
}

func (s *suite) aValidClassInformationWithMultipleTeachersAssigned(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	numberOfTeacher := 4
	request := s.createClassRequest(stepState.GandalfStateSchool.ID.Int)
	ctx, err := s.assignOwnersToCreateClassRequest(ctx, request, numberOfTeacher)
	stepState.GandalfStateRequest = request
	return GandalfStepStateToContext(ctx, stepState), err
}

func (s *suite) allTeachersInThatClassCanChatWithEachOther(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	if stepState.GandalfStateResponse != nil {
		conversationResp := stepState.GandalfStateResponse.(*tomPb.ConversationByClassResponse)
		stepState.GandalfStateConversationID = conversationResp.Conversations[0].ConversationId
	}
	ctx, err := s.allTeachersInThisClassSubscribeToEndpointStreamingEvent(ctx)
	if err != nil {
		return GandalfStepStateToContext(ctx, stepState), err

	}
	teacherIDs := []string{}
	for teacherID := range stepState.GandalfStateTeacherIDsMap {
		teacherIDs = append(teacherIDs, teacherID)
	}
	for teacherID, teacherToken := range stepState.GandalfStateTeacherIDsMap {
		ctx, err = s.userSendMessageAndBroadcastEventSendMessageToAllMembersInThisClass(ctx, teacherID, teacherToken, teacherIDs)
		if err != nil {
			return GandalfStepStateToContext(ctx, stepState), err

		}
	}
	return ctx, nil
}

func (s *suite) createClassRequest(schoolID int32) *bobPb.CreateClassRequest {
	request := &bobPb.CreateClassRequest{
		SchoolId:  schoolID,
		ClassName: "Class" + strconv.Itoa(rand.Int()),
		Grades:    []string{"G10", "G11"},
		Subjects: []bobPb.Subject{bobPb.SUBJECT_BIOLOGY,
			bobPb.SUBJECT_CHEMISTRY,
			bobPb.SUBJECT_LITERATURE},
		OwnerId:  "",
		OwnerIds: []string{}}
	return request
}

func (s *suite) createSchool(schoolName, countryName, cityName, districtName string) *bobEnt.School {
	city := &bobEnt.City{Name: database.Text(cityName), Country: database.Text(countryName)}
	district := &bobEnt.District{Name: database.Text(districtName), Country: database.Text(countryName), City: city}
	school := &bobEnt.School{Name: database.Text(schoolName + strconv.Itoa(rand.Int())), Country: database.Text(countryName), City: city, District: district, IsSystemSchool: pgtype.Bool{Bool: true, Status: pgtype.Present}, Point: pgtype.Point{Status: pgtype.Null}}
	return school
}

func (s *suite) adminInsertsSchools(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	r := &bobRepo.SchoolRepo{}
	if err := r.Import(ctx, s.bobDB, []*bobEnt.School{stepState.GandalfStateSchool}); err != nil {
		return ctx, err
	}

	//Init auth info
	stmt :=
		`
		INSERT INTO organization_auths
			(organization_id, auth_project_id, auth_tenant_id)
		VALUES
			($1, 'fake_aud', ''), ($1, 'dev-manabie-online', '')
		ON CONFLICT 
			DO NOTHING
		;
		`
	_, err := s.bobDB.Exec(ctx, stmt, stepState.GandalfStateSchool.ID)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func (s *suite) assignOwnersToCreateClassRequest(ctx context.Context, request *bobPb.CreateClassRequest, numberOfOwners int) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	ownerIDs := request.OwnerIds
	stepState.GandalfStateTeacherIDsMap = make(map[string]string)
	for i := 0; i < numberOfOwners; i++ {
		teacherID, teacherToken, err := s.signedInAs(ctx, bobEnt.UserGroupTeacher)
		if err != nil {
			return GandalfStepStateToContext(ctx, stepState), err

		}
		stepState.GandalfStateTeacherIDsMap[teacherID] = teacherToken
		t, _ := jwt.ParseString(teacherToken)
		ownerIDs = append(ownerIDs, t.Subject())
	}
	request.OwnerIds = ownerIDs
	return ctx, nil
}

func (s *suite) classMustHaveMembers(ctx context.Context, total int, group, isOwner, status string) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	count := 0
	err := s.bobDB.QueryRow(ctx, "SELECT COUNT(*) FROM class_members WHERE class_id = $1 AND user_group = $2 AND is_owner = $3 AND status = $4", stepState.GandalfStateCurrentClassID, group, isOwner, status).Scan(&count)
	if err != nil {
		return GandalfStepStateToContext(ctx, stepState), err

	}
	if total != count {
		return ctx, fmt.Errorf("class member not match result expect %d, got %d", total, count)
	}
	return ctx, nil
}

func (s *suite) getConversationByClass(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	request := &tomPb.ConversationByClassRequest{
		Limit:   uint32(1),
		ClassId: (uint32)(stepState.GandalfStateCurrentClassID)}
	stepState.GandalfStateResponse, stepState.GandalfStateResponseErr = tomPb.NewChatServiceClient(s.tomConn).ConversationByClass(contextWithToken(ctx, stepState.GandalfStateAuthToken), request)
	return GandalfStepStateToContext(ctx, stepState), nil
}

func (s *suite) validResponse(ctx context.Context, expectedCode codes.Code) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	stt, ok := status.FromError(stepState.GandalfStateResponseErr)
	if !ok {
		return ctx, fmt.Errorf("returned error is not status.Status, err: %s", stepState.GandalfStateResponseErr.Error())
	}
	if stt.Code() != expectedCode {
		return ctx, fmt.Errorf("expecting %d, got %d status code, message: %s", expectedCode, stt.Code(), stt.Message())
	}
	return ctx, nil
}

func (s *suite) allTeachersInThisClassSubscribeToEndpointStreamingEvent(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	stepState.GandalfStateSubV2Clients = make(map[string]tomPb.ChatService_SubscribeV2Client)
	for teacherID, teacherToken := range stepState.GandalfStateTeacherIDsMap {
		ctx := ctx
		subClient, err := tomPb.NewChatServiceClient(s.tomConn).SubscribeV2(contextWithToken(ctx, teacherToken),
			&tomPb.SubscribeV2Request{})
		if err != nil {
			return GandalfStepStateToContext(ctx, stepState), err

		}
		stepState.GandalfStateSubV2Clients[teacherID] = subClient
		time.Sleep(time.Second)
		ctx, err = s.userSendPingEventToStreamEverySeconds(ctx, 2, teacherToken)
		if err != nil {
			return GandalfStepStateToContext(ctx, stepState), err
		}
	}
	return GandalfStepStateToContext(ctx, stepState), nil
}

func (s *suite) userSendPingEventToStreamEverySeconds(ctx context.Context, durationSec int, token string) (context.Context, error) {
	c := tomPb.NewChatServiceClient(s.tomConn)
	// make sure user online
	t, _ := jwt.ParseString(token)
	r := &tomRepo.OnlineUserRepo{}
	try.Do(func(attempt int) (retry bool, err error) {
		users, err := r.OnlineUserDBRepo.Find(ctx, s.tomDB,
			database.TextArray([]string{t.Subject()}),
			pgtype.Timestamptz{Time: time.Now().Add(-5 * time.Second), Status: 2})
		if err != nil {
			return false, err
		}

		if len(users) == 0 {
			return true, fmt.Errorf("not found user online")
		}

		return false, nil
	})

	stepState := GandalfStepStateFromContext(ctx)
	streamV2, ok := stepState.GandalfStateSubV2Clients[t.Subject()]
	if !ok {
		return ctx, errors.New("user did not subscribe before")
	}

	sessionID := ""
	resp, err := streamV2.Recv()
	if err != nil {
		return ctx, err

	}

	if resp.Event.GetEventPing() == nil {
		return ctx, fmt.Errorf("stream must receive pingEvent first")
	}

	sessionID = resp.Event.GetEventPing().SessionId

	go func() {
		for {
			_, err := c.PingSubscribeV2(
				metadata.AppendToOutgoingContext(metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", token)),
				&tomPb.PingSubscribeV2Request{
					SessionId: sessionID,
				})
			if err != nil {
				return
			}

			time.Sleep(time.Duration(durationSec) * time.Second)
		}
	}()

	return ctx, nil
}

func (s *suite) createSendMessageRequestBy(userID, conversationId string) *tomPb.SendMessageRequest {
	request := &tomPb.SendMessageRequest{
		ConversationId: conversationId,
		Message:        fmt.Sprintf("Hello from %s", userID),
		UrlMedia:       "",
		Type:           tomPb.MESSAGE_TYPE_TEXT,
		LocalMessageId: idutil.ULIDNow()}
	return request
}

func (s *suite) teacherSendMessage(ctx context.Context, teacherID, teacherToken string) (context.Context, *tomPb.SendMessageResponse, error) {
	stepState := GandalfStepStateFromContext(ctx)
	request := s.createSendMessageRequestBy(teacherID, stepState.GandalfStateConversationID)
	stepState.GandalfStateRequestSentAt = time.Now()
	stepState.GandalfStateResponse, stepState.GandalfStateResponseErr = tomPb.NewChatServiceClient(s.tomConn).SendMessage(contextWithToken(ctx, teacherToken), request)
	ctx, err := s.validResponse(ctx, codes.OK)
	if err != nil {
		return GandalfStepStateToContext(ctx, stepState), nil, err
	}
	messageResponse := stepState.GandalfStateResponse.(*tomPb.SendMessageResponse)
	return GandalfStepStateToContext(ctx, stepState), messageResponse, nil
}

func (s *suite) userSendMessageAndBroadcastEventSendMessageToAllMembersInThisClass(ctx context.Context, userID, userToken string, members []string) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	ctx, messageResponse, err := s.teacherSendMessage(ctx, userID, userToken)
	if err != nil {
		return GandalfStepStateToContext(ctx, stepState), err

	}

	mRepo := &tomRepo.MessageRepo{}
	m, err := mRepo.FindByID(ctx, s.tomDB, database.Text(messageResponse.MessageId))
	if err != nil {
		return ctx, err
	}

	g, _ := errgroup.WithContext(ctx)
	// check all stream client already received event new message

	for _, mem := range members {
		subClient, ok := stepState.GandalfStateSubV2Clients[mem]
		if !ok {
			return ctx, fmt.Errorf("missing subscribeV2 for userID: %s", mem)
		}

		g.Go(func() error {
			for {
				resp, err := subClient.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					s.ZapLogger.Fatal(fmt.Sprintf("can not receive %v", err))
				}

				msg := resp.Event.GetEventNewMessage()
				if msg == nil {
					continue
				}

				if msg.GetUserId() != m.UserID.String {
					err = fmt.Errorf("userID from message does not match: expected %s, got: %s", m.UserID.String, msg.GetUserId())
					return err
				}

				if msg.GetMessageId() != m.ID.String {
					err = fmt.Errorf("messageID does not match: expected %s, got: %s", m.ID.String, msg.GetMessageId())
					return err
				}

				if msg.GetConversationId() != m.ConversationID.String {
					err = fmt.Errorf("conversationID does not match: expected %s, got: %s", m.ConversationID.String, msg.GetConversationId())
					return err
				}

				break
			}

			err := subClient.CloseSend()

			return err
		})
	}

	if err := g.Wait(); err != nil {
		return ctx, err
	}

	return ctx, nil
}
