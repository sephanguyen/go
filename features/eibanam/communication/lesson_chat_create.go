package communication

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/try"
	legacytpb "github.com/manabie-com/backend/pkg/genproto/tom"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
)

func (s *suite) lessonChatSteps() map[string]interface{} {
	return map[string]interface{}{
		`^school admin has created live lesson including student and teacher$`: s.schoolAdminHasCreatedLiveLessonIncludingStudentAndTeacher,
		`^"([^"]*)" has joined lesson$`:                                        s.hasJoinedLesson,
		`^"([^"]*)" sees lesson chat group is created$`:                        s.seesLessonChatGroupIsCreated,
		`^jprep school admin has created live lesson including student$`:       s.jprepSchoolAdminHasCreatedLiveLessonIncludingStudent,
	}
}

func (s *suite) jprepSchoolAdminHasCreatedLiveLessonIncludingStudent(ctx context.Context) (context.Context, error) {
	st := StepStateFromContext(ctx)
	student := st.profile.defaultStudent.id
	lesson, err := s.helper.JprepCreateLessonForStudents([]string{student})
	if err != nil {
		return ctx, err
	}
	st.lesson.id = s.helper.TranslateJprepLessonID(lesson.LessonID)
	st.lesson.name = lesson.ClassName
	return StepStateToContext(ctx, st), nil
}

func (s *suite) seesLessonChatGroupIsCreated(ctx context.Context, person string) (context.Context, error) {
	st := StepStateFromContext(ctx)
	err := try.Do(func(attempt int) (bool, error) {
		req := tpb.LiveLessonConversationDetailRequest{
			LessonId: st.lesson.id,
		}
		res, err := tpb.NewLessonChatReaderServiceClient(s.tomConn).LiveLessonConversationDetail(
			contextWithToken(ctx, st.getToken(person)),
			&req,
		)
		if err != nil {
			time.Sleep(2 * time.Second)
			return attempt < 5, err
		}
		conv := res.Conversation
		if conv.GetConversationName() != st.lesson.name {
			return false, fmt.Errorf("want conversation to have name %s, has %s", st.lesson.name, conv.ConversationName)
		}
		var foundUser bool
		usr := st.getProfile(person).id
		for _, u := range conv.Users {
			if u.Id == usr {
				foundUser = true
			}
		}
		if !foundUser {
			time.Sleep(2 * time.Second)
			return attempt < 5, fmt.Errorf("%s is not member of conversation", person)
		}
		st.chat.id = conv.GetConversationId()
		return false, nil
	})
	return StepStateToContext(ctx, st), err
}

func (s *suite) connectV2Stream(ctx context.Context, person string) (context.Context, error) {
	st := StepStateFromContext(ctx)
	chatSvc := legacytpb.NewChatServiceClient(s.tomConn)
	streamCtx, cancelstream := context.WithCancel(ctx)
	sessionID := ""

	err := try.Do(func(attempt int) (bool, error) {
		time.Sleep(1 * time.Second)

		streamV2, err := chatSvc.SubscribeV2(contextWithToken(streamCtx, st.getToken(person)), &legacytpb.SubscribeV2Request{})

		if err != nil {
			return attempt < 5, s.ResponseErr
		}
		id := s.getProfile(person).id
		// to prevent multiple streams opened per user
		oldstream, hasOldstream := st.chat.streams[id]
		if hasOldstream {
			oldstream.cancel()
		}
		st.chat.streams[id] = &stream{
			cancel: cancelstream,
			stream: streamV2,
		}

		for try := 0; try < 5; try++ {
			resp, err := streamV2.Recv()
			if err != nil {
				return true, err
			}
			if resp.Event.GetEventPing() != nil {
				sessionID = resp.Event.GetEventPing().SessionId
				break
			}
		}
		if sessionID == "" {
			cancelstream()
			return true, fmt.Errorf("did not receive first ping event from stream to get sessionid")
		}

		return false, nil
	})

	if err != nil {
		cancelstream()
		return ctx, err
	}

	token := s.getToken(person)

	go func() {
		for {
			_, err := chatSvc.PingSubscribeV2(contextWithToken(streamCtx, token), &legacytpb.PingSubscribeV2Request{SessionId: sessionID})
			if err != nil {
				return
			}
			time.Sleep(2 * time.Second)
		}
	}()
	return StepStateToContext(ctx, st), nil
}

// subscribe v2
func (s *suite) hasJoinedLesson(ctx context.Context, person string) (context.Context, error) {
	st := StepStateFromContext(ctx)
	ctx, err := s.connectV2Stream(ctx, person)
	if err != nil {
		return ctx, err
	}

	err = try.Do(func(attempt int) (bool, error) {
		req := bpb.JoinLessonRequest{
			LessonId: st.lesson.id,
		}

		_, err := bpb.NewClassModifierServiceClient(s.bobConn).JoinLesson(
			contextWithToken(ctx, st.getToken(person)),
			&req,
		)
		if err != nil {
			time.Sleep(2 * time.Second)
			return attempt < 5, err
		}
		return false, nil
	})
	if err != nil {
		return ctx, err
	}

	// make this user call refresh session api
	if st.lesson.usersInLesson == 0 {
		refresh := &tpb.RefreshLiveLessonSessionRequest{
			LessonId: st.lesson.id,
		}
		_, err := tpb.NewLessonChatReaderServiceClient(s.tomConn).RefreshLiveLessonSession(
			contextWithToken(ctx, st.getToken(person)),
			refresh,
		)
		if err != nil {
			return ctx, err
		}
		s.chat.sessionOffset = len(s.chat.sentMessages)
	}
	st.lesson.usersInLesson++
	return StepStateToContext(ctx, st), nil
}

func (s *suite) schoolAdminHasCreatedLiveLessonIncludingStudentAndTeacher(ctx context.Context) (context.Context, error) {
	st := StepStateFromContext(ctx)
	token := st.getToken(schoolAdmin)
	student := st.profile.defaultStudent.id
	teacher := st.profile.defaultTeacher.id
	intSchool, _ := strconv.ParseInt(st.SchoolID, 10, 64)
	req, res, err := s.helper.CreateLesson(token, []string{teacher}, []string{student}, int32(intSchool))
	st.lesson.id = res.Id
	st.lesson.name = req.GetName()
	if err != nil {
		return ctx, err
	}
	return StepStateToContext(ctx, st), nil
}
