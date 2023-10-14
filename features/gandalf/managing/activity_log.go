package managing

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/lestrrat/go-jwx/jwt"
	"github.com/segmentio/ksuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/manabie-com/backend/features/bob"
	bobpb "github.com/manabie-com/backend/pkg/genproto/bob"
	tompb "github.com/manabie-com/backend/pkg/genproto/tom"
	yasuopb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	bobV1 "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	commomV1 "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	eurekapb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	fatimapb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	shamirpb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"
)

func (s *suite) dataInTableActivityLogIsEmpty(ctx context.Context) (context.Context, error) {
	_, err := s.zeusDB.Exec(ctx, "delete from activity_logs")
	return ctx, err
}

func (s *suite) requestUpdateUserProfileWithUserGroupNamePhoneEmailSchool(ctx context.Context, number int, userGroup, userName, phone, email string, school int) (context.Context, error) {
	var err error
	bobReqs := make([]*bobpb.UpdateUserProfileRequest, 0, number)
	for i := 0; i < number; i++ {
		ctx, err = s.bobSuite.UserUpdatedProfileWithUserGroupNamePhoneEmailSchool(ctx, "his own", userGroup, userName, phone, email, school)
		if err != nil {
			return ctx, err
		}

		bobState := bob.StepStateFromContext(ctx)
		bobReqs = append(bobReqs, bobState.Request.(*bobpb.UpdateUserProfileRequest))
	}

	zeusState := GandalfStepStateFromContext(ctx)
	zeusState.ZeusStepState.UpdateUserProfileRequests = bobReqs
	zeusState.ZeusStepState.BobAuthToken = bob.StepStateFromContext(ctx).AuthToken
	return GandalfStepStateToContext(ctx, zeusState), nil
}

func (s *suite) requestListClassByCourse(ctx context.Context, number int) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	for i := 0; i < number; i++ {
		stepState.ZeusStepState.GetClassByCourseRequests = append(stepState.ZeusStepState.GetClassByCourseRequests, &eurekapb.ListClassByCourseRequest{CourseId: stepState.EurekaStepState.CourseID})
	}
	return GandalfStepStateToContext(ctx, stepState), nil
}

func (s *suite) requestCreatePackage(ctx context.Context, number int) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	for i := 0; i < number; i++ {
		now := time.Now()
		startAt := timestamppb.Now()
		endAt := timestamppb.New(now.Add(7 * 24 * time.Hour))
		request := &fatimapb.CreatePackageRequest{
			Name:    "package-" + ksuid.New().String(),
			Country: commomV1.Country_COUNTRY_VN,
			Descriptions: []string{
				"you can " + ksuid.New().String(),
				"you can " + ksuid.New().String(),
				"you can " + ksuid.New().String(),
			},
			Price:           uint32(rand.Int31n(999999999)),
			DiscountedPrice: uint32(rand.Int31n(999999999)),
			StartAt:         startAt,
			EndAt:           endAt,
			Duration:        0,
			PrioritizeLevel: rand.Int31(),
			Properties: &fatimapb.PackageProperties{
				CanWatchVideo:      []string{"course_" + ksuid.New().String(), "course_" + ksuid.New().String()},
				CanViewStudyGuide:  []string{"course_" + ksuid.New().String(), "course_" + ksuid.New().String()},
				CanDoQuiz:          []string{"course_" + ksuid.New().String(), "course_" + ksuid.New().String()},
				LimitOnlineLession: rand.Int31(),
				AskTutor:           &fatimapb.PackageProperties_AskTutorCfg{TotalQuestionLimit: 30, LimitDuration: bobV1.AskDuration_ASK_DURATION_WEEK},
			},
			IsRecommended: false,
		}
		stepState.ZeusStepState.CreatePackageRequests = append(stepState.ZeusStepState.CreatePackageRequests, request)
	}
	return GandalfStepStateToContext(ctx, stepState), nil
}

func (s *suite) requestGetConversationByID(ctx context.Context, number int) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	var err error
	lessonID := s.tomSuite.Request.(*bobpb.EvtLesson).GetJoinLesson().LessonId
	conversationID := s.tomSuite.LessonConversationMap[lessonID]
	for i := 0; i < number; i++ {
		ctx, err = s.tomSuite.AGetConversationRequest(ctx)
		if err != nil {
			return GandalfStepStateToContext(ctx, stepState), err
		}
		s.tomSuite.Request.(*tompb.GetConversationRequest).ConversationId = conversationID
		stepState.ZeusStepState.GetConversationRequests = append(stepState.ZeusStepState.GetConversationRequests, s.tomSuite.Request.(*tompb.GetConversationRequest))
	}
	stepState.ZeusStepState.TomAuthToken = s.tomSuite.TeacherToken
	return GandalfStepStateToContext(ctx, stepState), nil
}

func (s *suite) requestVerifyToken(ctx context.Context, number int) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	for i := 0; i < number; i++ {
		stepState.ZeusStepState.VerifyTokenRequests = append(stepState.ZeusStepState.VerifyTokenRequests, &shamirpb.VerifyTokenRequest{
			OriginalToken: bob.StepStateFromContext(ctx).AuthToken,
		})
	}
	return GandalfStepStateToContext(ctx, stepState), nil
}

func (s *suite) allOfAboveRequestAreSent(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	userServiceClient := bobpb.NewUserServiceClient(s.bobConn)
	schoolServiceClient := yasuopb.NewSchoolServiceClient(s.yasuoConn)
	courseReaderServiceClient := eurekapb.NewCourseReaderServiceClient(s.eurekaConn)
	subscriptionModifierServiceClient := fatimapb.NewSubscriptionModifierServiceClient(s.fatimaConn)
	chatServiceClient := tompb.NewChatServiceClient(s.tomConn)
	stepState.GandalfStateAuthToken = stepState.ZeusStepState.BobAuthToken
	t, err := jwt.ParseString(stepState.GandalfStateAuthToken)
	if err != nil {
		return GandalfStepStateToContext(ctx, stepState), err
	}
	stepState.GandalfStateUserIDs = append(stepState.GandalfStateUserIDs, t.Subject())
	callUpdateUserProfile := func(req *bobpb.UpdateUserProfileRequest) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_, err := userServiceClient.UpdateUserProfile(s.signedCtx(ctx), req)
		return err
	}
	for _, v := range stepState.ZeusStepState.UpdateUserProfileRequests {
		err = callUpdateUserProfile(v)
		if err != nil {
			return GandalfStepStateToContext(ctx, stepState), fmt.Errorf("callUpdateUserProfile: %v", err)
		}
	}
	callUpdateSchool := func(req *yasuopb.UpdateSchoolRequest) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_, err := schoolServiceClient.UpdateSchool(s.signedCtx(ctx), req)
		return err
	}
	for _, v := range stepState.ZeusStepState.UpdateSchoolRequests {
		err = callUpdateSchool(v)
		if err != nil {
			return GandalfStepStateToContext(ctx, stepState), fmt.Errorf("callUpdateSchool: %v", err)
		}
	}
	callGetClassByCourse := func(req *eurekapb.ListClassByCourseRequest) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_, err := courseReaderServiceClient.ListClassByCourse(s.signedCtx(ctx), req)
		return err
	}
	for _, v := range stepState.ZeusStepState.GetClassByCourseRequests {
		err = callGetClassByCourse(v)
		if err != nil {
			return GandalfStepStateToContext(ctx, stepState), fmt.Errorf("callGetClassByCourse: %v", err)
		}
	}
	callCreatePackage := func(req *fatimapb.CreatePackageRequest) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_, err := subscriptionModifierServiceClient.CreatePackage(s.signedCtx(ctx), req)
		return err
	}
	for _, v := range stepState.ZeusStepState.CreatePackageRequests {
		err = callCreatePackage(v)
		if err != nil {
			return GandalfStepStateToContext(ctx, stepState), fmt.Errorf("callCreatePackage: %v", err)
		}
	}
	stepState.GandalfStateAuthToken = stepState.ZeusStepState.TomAuthToken
	t, err = jwt.ParseString(stepState.GandalfStateAuthToken)
	if err != nil {
		return ctx, fmt.Errorf("jwt.ParseString: %v", err)
	}

	stepState.GandalfStateUserIDs = append(stepState.GandalfStateUserIDs, t.Subject())
	callGetConversation := func(req *tompb.GetConversationRequest) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_, err := chatServiceClient.GetConversation(s.signedCtx(ctx), req)
		return err
	}
	for _, v := range stepState.ZeusStepState.GetConversationRequests {
		err = callGetConversation(v)
		if err != nil {
			return GandalfStepStateToContext(ctx, stepState), fmt.Errorf("callGetConversation: %v", err)
		}
	}
	return ctx, nil
}

func (s *suite) numberOfRecordInTableActivityLog(ctx context.Context, number int) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	mainProcess := func() error {
		rows, err := s.zeusDB.Query(ctx,
			"SELECT COUNT(activity_log_id) FROM activity_logs WHERE user_id = ANY($1) OR action_type = $2",
			stepState.GandalfStateUserIDs,
			"/shamir.v1.TokenReaderService/VerifyToken")
		if err != nil {
			return err
		}
		defer rows.Close()

		var count int
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				return err
			}
		}

		if count != number {
			return fmt.Errorf("expected rows is %d, but the fact is %d", number, count)
		}

		return nil
	}
	return ctx, s.ExecuteWithRetry(mainProcess, 2*time.Second, 20)
}
