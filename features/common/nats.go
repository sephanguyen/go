package common

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	golibs_constants "github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

//nolint:gocyclo
func (s *suite) BobMustPushMsgSubjectToNats(ctx context.Context, msg, subject string) (context.Context, error) {
	time.Sleep(500 * time.Millisecond)
	stepState := StepStateFromContext(ctx)
	foundChn := make(chan struct{}, 1)

	switch subject {
	case golibs_constants.SubjectClassUpserted:
		timer := time.NewTimer(time.Minute * 1)
		defer timer.Stop()
		for {
			select {
			case message := <-stepState.FoundChanForJetStream:
				switch message := message.(type) {
				case *pb.EvtClassRoom_CreateClass_:
					if msg == "CreateClass" {
						if message.CreateClass.ClassId == stepState.CurrentClassID && message.CreateClass.ClassName != "" {
							return StepStateToContext(ctx, stepState), nil
						}
					}
				case *pb.EvtClassRoom_EditClass_:
					if msg == "EditClass" {
						if message.EditClass.ClassId == stepState.CurrentClassID && message.EditClass.ClassName != "" {
							return StepStateToContext(ctx, stepState), nil
						}
					}
				case *pb.EvtClassRoom_JoinClass_:
					if msg == "JoinClass" {
						if message.JoinClass.ClassId == stepState.CurrentClassID {
							return StepStateToContext(ctx, stepState), nil
						}
					}
				case *pb.EvtClassRoom_LeaveClass_:
					if msg == "LeaveClass" {
						return StepStateToContext(ctx, stepState), nil
					}
					if strings.Contains(msg, "LeaveClass") {
						if message.LeaveClass.ClassId == stepState.CurrentClassID && len(message.LeaveClass.UserIds) != 0 {
							if strings.Contains(msg, fmt.Sprintf("-is_kicked=%v", message.LeaveClass.IsKicked)) {
								return StepStateToContext(ctx, stepState), nil
							}
						}
					}
				case *pb.EvtClassRoom_ActiveConversation_:
					active := msg == "ActiveConversation"
					if message.ActiveConversation.ClassId == stepState.CurrentClassID && message.ActiveConversation.Active == active {
						return StepStateToContext(ctx, stepState), nil
					}
				}
			case <-timer.C:
				return StepStateToContext(ctx, stepState), errors.New("time out")
			}
		}

	case golibs_constants.SubjectLessonCreated:
		timer := time.NewTimer(time.Minute * 1)
		defer timer.Stop()
		select {
		case message := <-stepState.FoundChanForJetStream:
			switch v := message.(type) {
			case *bpb.EvtLesson_CreateLessons_:
				if msg == "CreateLessons" {
					res := stepState.Response.(*bpb.CreateLessonResponse)
					l := v.CreateLessons.Lessons[0]
					if l.LessonId == res.Id {
						if err := s.isLessonCreatedEventCorrectly(ctx, l); err != nil {
							return StepStateToContext(ctx, stepState), err
						}
						return StepStateToContext(ctx, stepState), nil
					}
				}
			}

		case <-timer.C:
			return StepStateToContext(ctx, stepState), errors.New("time out")
		}
	case golibs_constants.SubjectLessonUpdated:
		timer := time.NewTimer(time.Minute * 1)
		defer timer.Stop()
		select {
		case message := <-stepState.FoundChanForJetStream:
			switch v := message.(type) {
			case *bpb.EvtLesson_UpdateLesson_:
				if msg == "UpdateLesson" {
					l := v.UpdateLesson
					req := stepState.Request.(*bpb.UpdateLessonRequest)
					if l.LessonId == req.LessonId {
						if err := s.checkEventLessonUpdated(ctx, v.UpdateLesson); err != nil {
							return StepStateToContext(ctx, stepState), err
						}
						return StepStateToContext(ctx, stepState), nil
					}
				}
			}
		case <-timer.C:
			return StepStateToContext(ctx, stepState), errors.New("time out")
		}
	case golibs_constants.SubjectLearningObjectivesCreated:
		timer := time.NewTimer(time.Minute)
		defer timer.Stop()

		select {
		case <-stepState.FoundChanForJetStream:
			return StepStateToContext(ctx, stepState), nil
		case <-timer.C:
			return StepStateToContext(ctx, stepState), errors.New("time out")
		}

	case golibs_constants.SubjectSyncLocationUpserted:
		if msg == "UpsertLocation" {
			timer := time.NewTimer(time.Minute)
			defer timer.Stop()

			select {
			case <-stepState.FoundChanForJetStream:
				return StepStateToContext(ctx, stepState), nil
			case <-timer.C:
				return StepStateToContext(ctx, stepState), errors.New("time out")
			}
		}
	case golibs_constants.SubjectSyncLocationTypeUpserted:
		if msg == "UpsertLocationType" {
			timer := time.NewTimer(time.Minute)
			defer timer.Stop()

			select {
			case <-stepState.FoundChanForJetStream:
				return StepStateToContext(ctx, stepState), nil
			case <-timer.C:
				return StepStateToContext(ctx, stepState), errors.New("time out")
			}
		}
	}

	timer := time.NewTimer(time.Minute * 6)
	defer timer.Stop()

	select {
	case <-foundChn:
		return StepStateToContext(ctx, stepState), nil
	case <-timer.C:
		return StepStateToContext(ctx, stepState), errors.New("time out")
	}
}

//nolint:gocyclo
func (s *suite) LessonmgmtMustPushMsgSubjectToNats(ctx context.Context, msg, subject string) (context.Context, error) {
	time.Sleep(500 * time.Millisecond)
	stepState := StepStateFromContext(ctx)
	foundChn := make(chan struct{}, 1)

	switch subject {
	case golibs_constants.SubjectLessonDeleted:
		timer := time.NewTimer(time.Minute * 1)
		defer timer.Stop()
		select {
		case message := <-stepState.FoundChanForJetStream:
			switch v := message.(type) {
			case *bpb.EvtLesson_DeletedLessons_:
				if msg == "DeleteLesson" {
					// check logic
					lessonIDs := v.DeletedLessons.LessonIds
					if err := s.isLessonDeletedCorrectly(ctx, lessonIDs); err != nil {
						return StepStateToContext(ctx, stepState), err
					}
					return StepStateToContext(ctx, stepState), nil
				}
			}

		case <-timer.C:
			return StepStateToContext(ctx, stepState), errors.New("NATS with subject Lesson.Deleted error: timeout")
		}

	case golibs_constants.SubjectLessonCreated:
		timer := time.NewTimer(time.Minute * 1)
		defer timer.Stop()
		select {
		case message := <-stepState.FoundChanForJetStream:
			switch v := message.(type) {
			case *bpb.EvtLesson_CreateLessons_:
				if msg == "CreateLessons" {
					res := stepState.Response.(*lpb.CreateLessonResponse)
					l := v.CreateLessons.Lessons[0]
					if l.LessonId == res.Id {
						if err := s.isLessonCreatedEventCorrectly(ctx, l); err != nil {
							return StepStateToContext(ctx, stepState), err
						}
						return StepStateToContext(ctx, stepState), nil
					}
				}
			}

		case <-timer.C:
			return StepStateToContext(ctx, stepState), errors.New("time out")
		}
	case golibs_constants.SubjectLessonUpdated:
		timer := time.NewTimer(time.Minute * 1)
		defer timer.Stop()
		select {
		case message := <-stepState.FoundChanForJetStream:
			switch v := message.(type) {
			case *bpb.EvtLesson_UpdateLesson_:
				if msg == "UpdateLesson" {
					var lessonId string
					switch req := stepState.Request.(type) {
					case *lpb.UpdateLessonRequest:
						lessonId = req.LessonId
					case *lpb.UpdateLessonSchedulingStatusRequest:
						lessonId = req.LessonId
					}
					if lessonId == v.UpdateLesson.GetLessonId() {
						if err := s.checkEventLessonUpdated(ctx, v.UpdateLesson); err != nil {
							return StepStateToContext(ctx, stepState), err
						}
						return StepStateToContext(ctx, stepState), nil
					}
				}
			}
		case <-timer.C:
			return StepStateToContext(ctx, stepState), errors.New("time out")
		}
	}

	timer := time.NewTimer(time.Minute * 6)
	defer timer.Stop()

	select {
	case <-foundChn:
		return StepStateToContext(ctx, stepState), nil
	case <-timer.C:
		return StepStateToContext(ctx, stepState), errors.New("time out")
	}
}

func (s *suite) isLessonDeletedCorrectly(ctx context.Context, lessonIDs []string) error {
	if _, err := s.getCountDeletedLessonByLessonID(ctx, lessonIDs); err != nil {
		return err
	}
	if s.DeletedLessonCount != len(lessonIDs) {
		return fmt.Errorf("mismatch deleted lesson count: expected %d, got %d", len(lessonIDs), s.DeletedLessonCount)
	}
	return nil
}

func (s *suite) isLessonCreatedEventCorrectly(ctx context.Context, l *bpb.EvtLesson_Lesson) error {
	if _, err := s.getLessonByID(ctx); err != nil {
		return err
	}

	aLesson := s.Lesson

	if domain.LessonSchedulingStatus(l.SchedulingStatus.String()) != aLesson.SchedulingStatus {
		return fmt.Errorf("SchedulingStatus in message from event lesson.created doesn't match")
	}

	if !aLesson.StartTime.Equal(l.StartAt.AsTime()) {
		return fmt.Errorf("StartAt in message from event lesson.created doesn't match")
	}

	if !aLesson.EndTime.Equal(l.EndAt.AsTime()) {
		return fmt.Errorf("EndAt in message from event lesson.created doesn't match")
	}

	if !stringutil.SliceElementsMatch(l.GetLearnerIds(), aLesson.GetLearnersIDs()) {
		return fmt.Errorf("LearnerIds in message from event lesson.created doesn't match")
	}

	if !stringutil.SliceElementsMatch(l.GetTeacherIds(), aLesson.GetTeacherIDs()) {
		return fmt.Errorf("TeacherIds in message from event lesson.created doesn't match")
	}

	if l.LocationId != aLesson.LocationID {
		return fmt.Errorf("LocationId in message from event lesson.created doesn't match")
	}
	return nil
}

func (s *suite) checkEventLessonUpdated(ctx context.Context, l *bpb.EvtLesson_UpdateLesson) error {
	stepState := StepStateFromContext(ctx)
	bLesson := s.Lesson

	if _, err := s.getLessonByID(ctx); err != nil {
		return err
	}

	aLesson := s.Lesson

	switch stepState.Request.(type) {
	case *bpb.UpdateLessonRequest, *lpb.UpdateLessonRequest:
		switch {
		case domain.LessonSchedulingStatus(l.SchedulingStatusAfter.String()) != bLesson.SchedulingStatus:
			return fmt.Errorf("expected %s for SchedulingStatusAfter, got %s", bLesson.SchedulingStatus, l.SchedulingStatusAfter)

		case domain.LessonSchedulingStatus(l.SchedulingStatusBefore.String()) != bLesson.SchedulingStatus:
			return fmt.Errorf("expected %s for SchedulingStatusBefore, got %s", bLesson.SchedulingStatus, l.SchedulingStatusBefore)

		case !bLesson.StartTime.Equal(l.StartAtBefore.AsTime()):
			return fmt.Errorf("expected %s for StartAtBefore, got %s", bLesson.StartTime, l.StartAtBefore.AsTime())

		case !aLesson.StartTime.Equal(l.StartAtAfter.AsTime()):
			return fmt.Errorf("expected %s for StartAtAfter, got %s", aLesson.StartTime, l.StartAtAfter.AsTime())

		case !bLesson.EndTime.Equal(l.EndAtBefore.AsTime()):
			return fmt.Errorf("expected %s for EndAtBefore, got %s", bLesson.EndTime, l.EndAtBefore.AsTime())

		case !aLesson.EndTime.Equal(l.EndAtAfter.AsTime()):
			return fmt.Errorf("expected %s for EndAtAfter, got %s", aLesson.EndTime, l.EndAtAfter.AsTime())

		case l.LocationIdBefore != bLesson.LocationID:
			return fmt.Errorf("expected %s for LocationIdBefore, got %s", bLesson.LocationID, l.LocationIdAfter)

		case l.LocationIdAfter != aLesson.LocationID:
			return fmt.Errorf("expected %s for LocationIdAfter, got %s", aLesson.LocationID, l.LocationIdAfter)

		case !stringutil.SliceElementsMatch(l.TeacherIdsBefore, bLesson.GetTeacherIDs()):
			return fmt.Errorf("expected %s for TeacherIdsBefore, got %s", bLesson.GetTeacherIDs(), l.TeacherIdsBefore)

		case !stringutil.SliceElementsMatch(l.TeacherIdsAfter, aLesson.GetTeacherIDs()):
			return fmt.Errorf("expected %s for TeacherIdsAfter, got %s", aLesson.GetTeacherIDs(), l.TeacherIdsAfter)

		case !stringutil.SliceElementsMatch(l.LearnerIds, aLesson.GetLearnersIDs()):
			return fmt.Errorf("expected %s for LearnerIds, got %s", aLesson.GetLearnersIDs(), l.LearnerIds)
		}

	case *lpb.UpdateLessonSchedulingStatusRequest:
		switch {
		case domain.LessonSchedulingStatus(l.SchedulingStatusAfter.String()) != aLesson.SchedulingStatus:
			return fmt.Errorf("expected %s for SchedulingStatusAfter, got %s", aLesson.SchedulingStatus, l.SchedulingStatusAfter)

		case domain.LessonSchedulingStatus(l.SchedulingStatusBefore.String()) != bLesson.SchedulingStatus:
			return fmt.Errorf("expected %s for SchedulingStatusBefore, got %s", bLesson.SchedulingStatus, l.SchedulingStatusBefore)

		case l.ClassName != bLesson.Name:
			return fmt.Errorf("expected %s for ClassName, got %s", bLesson.Name, l.ClassName)

		case !bLesson.StartTime.Equal(l.StartAtBefore.AsTime()):
			return fmt.Errorf("expected %s for StartAtBefore, got %s", bLesson.StartTime, l.StartAtBefore.AsTime())

		case !bLesson.StartTime.Equal(l.StartAtAfter.AsTime()):
			return fmt.Errorf("expected %s for StartAtAfter, got %s", bLesson.StartTime, l.StartAtAfter.AsTime())

		case !bLesson.EndTime.Equal(l.EndAtBefore.AsTime()):
			return fmt.Errorf("expected %s for EndAtBefore, got %s", bLesson.EndTime, l.EndAtBefore.AsTime())

		case !bLesson.EndTime.Equal(l.EndAtAfter.AsTime()):
			return fmt.Errorf("expected %s for EndAtAfter, got %s", bLesson.EndTime, l.EndAtAfter.AsTime())

		case l.LocationIdBefore != bLesson.LocationID:
			return fmt.Errorf("expected %s for LocationIdBefore, got %s", bLesson.LocationID, l.LocationIdBefore)

		case l.LocationIdAfter != bLesson.LocationID:
			return fmt.Errorf("expected %s for LocationIdAfter, got %s", bLesson.LocationID, l.LocationIdAfter)

		case !stringutil.SliceElementsMatch(l.TeacherIdsAfter, bLesson.GetTeacherIDs()):
			return fmt.Errorf("expected %s for TeacherIdsAfter, got %s", bLesson.GetTeacherIDs(), l.TeacherIdsAfter)

		case !stringutil.SliceElementsMatch(l.TeacherIdsBefore, bLesson.GetTeacherIDs()):
			return fmt.Errorf("expected %s for TeacherIdsBefore, got %s", bLesson.GetTeacherIDs(), l.TeacherIdsBefore)

		case !stringutil.SliceElementsMatch(l.LearnerIds, bLesson.GetLearnersIDs()):
			return fmt.Errorf("expected %s for LearnerIds, got %s", bLesson.GetLearnersIDs(), l.LearnerIds)
		}
	default:
		return fmt.Errorf("expected type of request is *bpb.UpdateLessonRequest, *lpb.UpdateLessonRequest  or *lpb.UpdateLessonSchedulingStatusRequest")
	}
	return nil
}
