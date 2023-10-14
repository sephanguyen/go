package consumers

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/consumers/handler"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	masterMgmt "github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type StudentChangeClassHandler struct {
	Logger            *zap.Logger
	DB                database.Ext
	WrapperConnection *support.WrapperDBConnection
	JSM               nats.JetStreamManagement
	LessonRepo        infrastructure.LessonRepo
	LessonMemberRepo  infrastructure.LessonMemberRepo
	ClassMemberRepo   masterMgmt.ClassMemberRepo
	LessonReportRepo  infrastructure.LessonReportRepo
}

func (s *StudentChangeClassHandler) getHandler(ctx context.Context, event *mpb.EvtClass) handler.IChangeClassHandler {
	if event.GetJoinClass() != nil {
		eventJoinClass := event.GetJoinClass()
		classId := eventJoinClass.GetClassId()
		studentId := eventJoinClass.GetUserId()
		oldClassId := eventJoinClass.GetOldClassId()
		return &handler.StudentJoinClass{
			ClassID:          classId,
			StudentID:        studentId,
			OldClassID:       oldClassId,
			Ctx:              ctx,
			DB:               s.DB,
			ClassMemberRepo:  s.ClassMemberRepo,
			LessonMemberRepo: s.LessonMemberRepo,
			LessonReportRepo: s.LessonReportRepo,
		}
	}
	if event.GetLeaveClass() != nil {
		eventLeaveClass := event.GetLeaveClass()
		classId := eventLeaveClass.GetClassId()
		studentId := eventLeaveClass.GetUserId()

		return &handler.StudentLeaveClass{
			ClassID:          classId,
			StudentID:        studentId,
			Ctx:              ctx,
			LessonMemberRepo: s.LessonMemberRepo,
		}
	}
	return nil
}

func (s *StudentChangeClassHandler) Handle(ctx context.Context, msg []byte) (bool, error) {
	s.Logger.Info("[StudentChangeClassHandler]: Received message on",
		zap.String("data", string(msg)),
		zap.String("subject", constants.SubjectMasterMgmtClassUpserted),
	)
	return s.handleStudentChangeEvent(ctx, msg)
}

func (s *StudentChangeClassHandler) handleStudentChangeEvent(ctx context.Context, msg []byte) (bool, error) {
	conn, err := s.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return false, err
	}
	classEvent := &mpb.EvtClass{}

	if err := proto.Unmarshal(msg, classEvent); err != nil {
		return false, err
	}

	handler := s.getHandler(ctx, classEvent)
	if handler != nil {
		queryLesson, err := handler.GetQueryLesson()
		if err != nil {
			return false, err
		}
		lessons, err := s.LessonRepo.GetLessonsTeachingModelGroupByClassIdWithDuration(ctx, conn, queryLesson)
		if err != nil {
			return false, err
		}

		lessonMembers := make([]*domain.LessonMember, 0, len(lessons))

		for _, lesson := range lessons {
			lessonMember := handler.GetLessonMember(lesson)
			lessonMembers = append(lessonMembers, lessonMember)
		}

		err = handler.UpdateLessonMember(conn, lessonMembers)

		if err != nil {
			return false, err
		}
	}
	return true, nil
}
