package bob

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (s *suite) userSignedAsStudentWhoBelongToLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error
	studentID := stepState.StudentIds[0]
	stepState.AuthToken, err = s.generateExchangeToken(studentID, constant.UserGroupStudent)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentUserID = studentID
	stepState.CurrentStudentID = studentID
	stepState.CurrentUserGroup = constant.UserGroupStudent

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) InsertStudentSubscription(ctx context.Context, startAt, endAt time.Time, studentIDWithCourseID ...string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	queueFn := func(b *pgx.Batch, studentID, courseID string) string {
		id := idutil.ULIDNow()
		query := `INSERT INTO lesson_student_subscriptions (student_subscription_id, subscription_id, student_id, course_id, start_at, end_at) VALUES ($1, $2, $3, $4, $5, $6)`
		b.Queue(query, id, id, studentID, courseID, startAt, endAt)
		return id
	}

	b := &pgx.Batch{}
	ids := make([]string, 0, len(studentIDWithCourseID))
	for i := 0; i < len(studentIDWithCourseID); i += 2 {
		ids = append(ids, queueFn(b, studentIDWithCourseID[i], studentIDWithCourseID[i+1]))
	}
	result := s.DB.SendBatch(ctx, b)
	defer result.Close()

	for i, iEnd := 0, b.Len(); i < iEnd; i++ {
		_, err := result.Exec()
		if err != nil {
			return nil, fmt.Errorf("result.Exec[%d]: %w", i, err)
		}
	}
	return ids, nil
}

func (s *suite) SomeStudentSubscriptions(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courseID := stepState.CourseIDs[len(stepState.CourseIDs)-1]
	studentIDWithCourseID := make([]string, 0, len(stepState.StudentIds)*2)
	for _, studentID := range stepState.StudentIds {
		studentIDWithCourseID = append(studentIDWithCourseID, studentID, courseID)
	}
	stepState.StartDate = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	stepState.EndDate = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	ids, err := s.InsertStudentSubscription(ctx, stepState.StartDate, stepState.EndDate, studentIDWithCourseID...)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not insert student subscription: %w", err)
	}
	stepState.StudentIDWithCourseID = studentIDWithCourseID

	// create access path for above list student subscriptions
	for _, l := range stepState.LocationIDs {
		for _, id := range ids {
			stmt := `INSERT INTO lesson_student_subscription_access_path (student_subscription_id,location_id) VALUES($1,$2)`
			_, err := s.DB.Exec(ctx, stmt, id, l)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lesson_student_subscription_access_path with student_subscription_id:%s, location_id:%s, err:%v", id, l, err)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aLiveLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if ctx, err := s.signedAsAccountV2(ctx, "school admin"); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if ctx, err := s.SomeStudentSubscriptions(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if ctx, err := s.UserCreateLiveLessonWithMissing(ctx, ""); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userShareAMaterialWithTypeIsPdfInLiveLessonRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &bpb.ModifyLiveLessonStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_ShareAMaterial{
			ShareAMaterial: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand{
				MediaId: stepState.MediaIDs[0],
			},
		},
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.Conn).ModifyLiveLessonState(s.signedCtx(ctx), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userShareAMaterialWithTypeIsVideoInLiveLessonRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &bpb.ModifyLiveLessonStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_ShareAMaterial{
			ShareAMaterial: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand{
				MediaId: stepState.MediaIDs[0],
				State: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand_VideoState{
					VideoState: &bpb.LiveLessonState_CurrentMaterial_VideoState{
						CurrentTime: durationpb.New(2 * time.Minute),
						PlayerState: bpb.PlayerState_PLAYER_STATE_PLAYING,
					},
				},
			},
		},
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.Conn).ModifyLiveLessonState(s.signedCtx(ctx), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetCurrentStateOfLiveLessonRoom(ctx context.Context, lessonID string) (*bpb.LiveLessonStateResponse, error) {
	req := &bpb.LiveLessonStateRequest{Id: lessonID}

	res, err := bpb.NewLessonReaderServiceClient(s.Conn).GetLiveLessonState(s.signedCtx(ctx), req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *suite) userGetCurrentMaterialStateOfLiveLessonRoomIsEmpty(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	res, err := s.userGetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.Id != stepState.CurrentLessonID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson %s but got %s", stepState.CurrentLessonID, res.Id)
	}

	if res.CurrentTime.AsTime().IsZero() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson's current time but got empty")
	}

	if res.CurrentMaterial != nil &&
		(len(res.CurrentMaterial.MediaId) != 0 || res.CurrentMaterial.State != nil || res.CurrentMaterial.Data != nil) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected current material be empty but got %v", res.CurrentMaterial)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetCurrentMaterialStateOfLiveLessonRoomIsPdf(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	res, err := s.userGetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.Id != stepState.CurrentLessonID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson %s but got %s", stepState.CurrentLessonID, res.Id)
	}

	if res.CurrentTime.AsTime().IsZero() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson's current time but got empty")
	}

	req := stepState.Request.(*bpb.ModifyLiveLessonStateRequest)
	if res.CurrentMaterial == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected current material but got empty")
	}

	if res.CurrentMaterial.MediaId != req.GetShareAMaterial().MediaId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected media %s but got %s", req.GetShareAMaterial().MediaId, res.CurrentMaterial.MediaId)
	}

	if res.CurrentMaterial.State != nil {
		if _, ok := res.CurrentMaterial.State.(*bpb.LiveLessonState_CurrentMaterial_PdfState); !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("current meterial is not pdf type")
		}
	}

	if res.CurrentMaterial.Data == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected media's data but got empty")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetCurrentMaterialStateOfLiveLessonRoomIsVideo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	res, err := s.userGetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.Id != stepState.CurrentLessonID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson %s but got %s", stepState.CurrentLessonID, res.Id)
	}

	if res.CurrentTime.AsTime().IsZero() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson's current time but got empty")
	}

	req := stepState.Request.(*bpb.ModifyLiveLessonStateRequest)
	if res.CurrentMaterial == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected current material but got empty")
	}

	if res.CurrentMaterial.MediaId != req.GetShareAMaterial().MediaId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected media %s but got %s", req.GetShareAMaterial().MediaId, res.CurrentMaterial.MediaId)
	}

	if res.CurrentMaterial.Data == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected media's data but got empty")
	}

	if res.CurrentMaterial.GetVideoState() == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected video state but got empty")
	}

	actualCurrentT := res.CurrentMaterial.GetVideoState().CurrentTime.AsDuration()
	expectedCurrentT := req.GetShareAMaterial().GetVideoState().CurrentTime.AsDuration()
	if actualCurrentT != expectedCurrentT {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected video's current time %v but got %v", expectedCurrentT, actualCurrentT)
	}

	actualPlayerSt := res.CurrentMaterial.GetVideoState().PlayerState.String()
	expectedPlayerSt := req.GetShareAMaterial().GetVideoState().PlayerState.String()
	if actualPlayerSt != expectedPlayerSt {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected video's player state %v but got %v", expectedPlayerSt, actualPlayerSt)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userRaiseHandInLiveLessonRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &bpb.ModifyLiveLessonStateRequest{
		Id:      stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_RaiseHand{},
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.Conn).ModifyLiveLessonState(s.signedCtx(ctx), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func filterHandsUpState(userStateLearners []*bpb.LiveLessonStateResponse_UsersState_LearnerState) []*bpb.LiveLessonStateResponse_UsersState_LearnerState {
	handsUpStates := []*bpb.LiveLessonStateResponse_UsersState_LearnerState{}

	for _, i := range userStateLearners {
		if i.HandsUp.Value {
			handsUpStates = append(handsUpStates, i)
		}
	}
	return handsUpStates
}

func (s *suite) userGetHandsUpState(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	res, err := s.userGetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.Id != stepState.CurrentLessonID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson %s but got %s", stepState.CurrentLessonID, res.Id)
	}

	// only get hands up leaner instead check directly on all state
	handUpLearners := filterHandsUpState(res.UsersState.Learners)

	if len(handUpLearners) != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected get 1 learner's state but got %d", len(res.UsersState.Learners))
	}

	expectedHandsUp := false
	req := stepState.Request.(*bpb.ModifyLiveLessonStateRequest)
	switch req.Command.(type) {
	case *bpb.ModifyLiveLessonStateRequest_RaiseHand:
		expectedHandsUp = true
	case *bpb.ModifyLiveLessonStateRequest_HandOff:
		expectedHandsUp = false
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("before request %T is invalid command", req.Command)
	}

	actualLearnerSt := handUpLearners[0]
	if actualLearnerSt.UserId != stepState.CurrentStudentID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected learner id %s but got %s", stepState.CurrentStudentID, actualLearnerSt.UserId)
	}

	if actualLearnerSt.HandsUp.Value != expectedHandsUp {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected learner's hands up state %v but got %v", expectedHandsUp, actualLearnerSt.HandsUp.Value)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetAllLearnersHandsUpStatesWhoAllHaveValueIsOff(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	res, err := s.userGetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.Id != stepState.CurrentLessonID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson %s but got %s", stepState.CurrentLessonID, res.Id)
	}

	for _, learner := range res.UsersState.Learners {
		if learner.HandsUp.Value {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected all learner's hands up is off but %s is not", learner.UserId)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userRequestRecordingLiveLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	currentTime := strconv.FormatInt(time.Now().Unix(), 10)
	req := &vpb.StartRecordingRequest{
		LessonId:           stepState.CurrentLessonID,
		SubscribeVideoUids: []string{"#allstream#"},
		SubscribeAudioUids: []string{"#allstream#"},
		FileNamePrefix:     []string{stepState.CurrentLessonID, currentTime},
		TranscodingConfig: &vpb.StartRecordingRequest_TranscodingConfig{
			Height:           720,
			Width:            1280,
			Bitrate:          2000,
			Fps:              30,
			MixedVideoLayout: 0,
			BackgroundColor:  "#FF0000",
		},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLessonRecordingServiceClient(s.VirtualClassroomConn).StartRecording(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetCurrentRecordingLiveLessonPermissionToStartRecording(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	res, err := s.userGetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.Id != stepState.CurrentLessonID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson %s but got %s", stepState.CurrentLessonID, res.Id)
	}

	if res.Recording == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson have recording status but got nil")
	}
	if !res.Recording.IsRecording {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected live lesson is recording but it's not")
	}
	if res.Recording.Creator != stepState.CurrentUserID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected user %s get recording permission but creator is %s", stepState.CurrentUserID, res.Recording.Creator)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) liveLessonIsNotRecording(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if ctx, err := s.aSignedInSchoolAdmin(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	res, err := s.userGetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.Id != stepState.CurrentLessonID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson %s but got %s", stepState.CurrentLessonID, res.Id)
	}

	if res.Recording != nil && res.Recording.IsRecording {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected live lesson %s is not recording but it do", stepState.CurrentLessonID)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpdatesChatOfLearnersInLiveLessonRoom(ctx context.Context, state string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var request *bpb.ModifyLiveLessonStateRequest
	switch state {
	case "enables":
		request = &bpb.ModifyLiveLessonStateRequest{
			Id: stepState.CurrentLessonID,
			Command: &bpb.ModifyLiveLessonStateRequest_ChatEnable{
				ChatEnable: &bpb.ModifyLiveLessonStateRequest_Learners{
					Learners: stepState.StudentIds,
				},
			},
		}
	case "disables":
		request = &bpb.ModifyLiveLessonStateRequest{
			Id: stepState.CurrentLessonID,
			Command: &bpb.ModifyLiveLessonStateRequest_ChatDisable{
				ChatDisable: &bpb.ModifyLiveLessonStateRequest_Learners{
					Learners: stepState.StudentIds,
				},
			},
		}
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("state entered is not supported")
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.Conn).ModifyLiveLessonState(s.signedCtx(ctx), request)
	stepState.Request = request

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsLearnersChatPermission(ctx context.Context, state string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	res, err := s.userGetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if stepState.CurrentLessonID != res.Id {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson %s but got %s", stepState.CurrentLessonID, res.Id)
	}

	var expectedChatPermission bool
	switch state {
	case "enabled":
		expectedChatPermission = true
	case "disabled":
		expectedChatPermission = false
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("unsupported permission state")
	}

	for _, learner := range res.UsersState.Learners {
		if learner.Chat.Value != expectedChatPermission {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected learner's chat permission %v but got %v", expectedChatPermission, learner.Chat.Value)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
