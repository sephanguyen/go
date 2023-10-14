package learning

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/bob"
	"github.com/manabie-com/backend/internal/golibs/constants"
)

func (s *suite) teacherGetNewConversationForThisLesson(ctx context.Context, lessonName string) (context.Context, error) {
	// called := make(chan bool)
	// errorChan := make(chan error)
	// defer func() {
	// 	close(called)
	// 	close(errorChan)
	// }()
	bobState := bob.StepStateFromContext(ctx)
	// subscription, err := s.Bus.Subscribe(constants.SubjectLessonEventNats, func(msg *stan.Msg) {
	// 	// waiting some seconds to other Subscribers finish
	// 	time.Sleep(2 * time.Second)
	// 	req := bpb.EvtLesson{}
	// 	if err := req.Unmarshal(msg.Data); err != nil {
	// 		err = multierr.Append(err, msg.Ack())
	// 		errorChan <- fmt.Errorf("proto.Unmarshal %v", zap.Error(err))
	// 		return
	// 	}
	// 	evtCreateLesson := req.GetCreateLessons()
	// 	if evtCreateLesson == nil {
	// 		errorChan <- fmt.Errorf("could not get evtCreateLesson")
	// 		return
	// 	}
	// 	lessonID := database.Text(evtCreateLesson.Lessons[0].LessonId)
	// 	name := database.Text(evtCreateLesson.Lessons[0].Name)
	// 	if name.String != lessonName {
	// 		errorChan <- fmt.Errorf("expected lesson name %s in nats message but got %s", lessonName, name.String)
	// 		return
	// 	}
	// 	// check conversion
	// 	clRepo := tomeRepo.ConversationLessonRepo{}
	// 	res, err := clRepo.FindByLessonID(ctx, s.tomDB, lessonID)
	// 	if err != nil {
	// 		errorChan <- fmt.Errorf("could not ConversationLessonRepo.FindByLessonID: %v", err)
	// 		return
	// 	}
	// 	cRepo := tomeRepo.ConversationRepo{}
	// 	conversation, err := cRepo.FindByID(ctx, s.tomDB, res.ConversationID)
	// 	if err != nil {
	// 		errorChan <- fmt.Errorf("could not ConversationRepo.FindByID: %v", err)
	// 		return
	// 	}
	// 	if conversation.Name.String != name.String {
	// 		errorChan <- fmt.Errorf("expected conversation name %s but got %s", name.String, conversation.Name.String)
	// 		return
	// 	}
	// 	// check room id
	// 	lessonRepo := bobRepo.LessonRepo{}
	// 	_, err = lessonRepo.FindByID(ctx, s.bobDB, lessonID)
	// 	if err != nil {
	// 		errorChan <- fmt.Errorf("could not LessonRepo.FindByID: %v", err)
	// 		return
	// 	}
	// 	// if len(lesson.RoomID.String) == 0 {
	// 	// 	// TODO: enable return error checking room id soon
	// 	// 	errorChan <- fmt.Errorf("expected lesson's room id but got empty")
	// 	// 	return
	// 	// }
	// 	called <- true
	// }, stan.StartAtTime(bobState.RequestSentAt), stan.SetManualAckMode())
	// if err != nil {
	// 	return ctx, fmt.Errorf("cannot subscribe to NATS: %v", err)
	// }
	// defer func() {
	// 	_ = subscription.Unsubscribe()
	// 	subscription.Close()
	// }()
	select {
	// case err := <-errorChan:
	// 	return ctx, err
	case <-bobState.FoundChanForJetStream:
		return ctx, nil
	case <-time.After(10 * time.Second):
		return ctx, fmt.Errorf("time out when subscribe topic %s", constants.SubjectLessonCreated)
	}
}
