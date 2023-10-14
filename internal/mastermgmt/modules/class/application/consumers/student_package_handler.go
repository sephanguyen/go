package consumers

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type SubscriberHandler interface {
	Handle(ctx context.Context, msg []byte) (bool, error)
}

type StudentPackageHandler struct {
	Logger *zap.Logger
	DB     database.Ext
	JSM    nats.JetStreamManagement

	ClassMemberRepo infrastructure.ClassMemberRepo
}

func (s *StudentPackageHandler) Handle(ctx context.Context, msg []byte) (bool, error) {
	s.Logger.Info("[StudentPackageHandler]: Received message on",
		zap.String("data", string(msg)),
		zap.String("subject", constants.SubjectStudentPackageV2EventNats),
	)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var studentPackageEvt npb.EventStudentPackageV2
	err := proto.Unmarshal(msg, &studentPackageEvt)
	if err != nil {
		s.Logger.Error("Failed to parse npb.EventStudentPackageV2: ", zap.Error(err))
		return false, fmt.Errorf("Failed to parse npb.EventStudentPackageV2 :%w", err)
	}
	sp := studentPackageEvt.GetStudentPackage()
	if sp.GetIsActive() {
		now := time.Now()
		messageJoinClass := &mpb.EvtClass_JoinClass{
			ClassId:    "",
			UserId:     sp.StudentId,
			OldClassId: "",
		}
		messageLeaveClass := &mpb.EvtClass_LeaveClass{
			ClassId: "",
			UserId:  sp.StudentId,
		}
		err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
			classMembers, err := s.ClassMemberRepo.GetByUserAndCourse(ctx, s.DB, sp.StudentId, sp.Package.CourseId)
			if err != nil {
				return err
			}
			classMemberID := idutil.ULIDNow()
			if cm, ok := classMembers[sp.StudentId]; ok {
				if len(sp.Package.ClassId) == 0 || sp.Package.ClassId != cm.ClassID {
					messageLeaveClass.ClassId = cm.ClassID
					messageJoinClass.OldClassId = cm.ClassID
				} else {
					classMemberID = cm.ClassMemberID
				}
				messageJoinClass.ClassId = sp.Package.ClassId
				err := s.ClassMemberRepo.DeleteByUserIDAndClassID(ctx, s.DB, cm.UserID, cm.ClassID)
				if err != nil {
					return err
				}
			}

			if len(sp.Package.ClassId) > 0 {
				if len(classMembers) == 0 {
					messageJoinClass.ClassId = sp.Package.ClassId
				}
				classMember := &domain.ClassMember{
					ClassMemberID: classMemberID,
					ClassID:       sp.Package.ClassId,
					UserID:        sp.StudentId,
					CreatedAt:     now,
					UpdatedAt:     now,
					StartDate:     sp.Package.StartDate.AsTime(),
					EndDate:       sp.Package.EndDate.AsTime(),
				}
				err = s.ClassMemberRepo.UpsertClassMember(ctx, s.DB, classMember)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return true, fmt.Errorf("s.Handle: %w", err)
		}
		if len(messageJoinClass.ClassId) > 0 {
			err = s.PublishClassEvt(ctx, &mpb.EvtClass{
				Message: &mpb.EvtClass_JoinClass_{
					JoinClass: messageJoinClass,
				},
			})
			if err != nil {
				return false, fmt.Errorf("PublishClassEvt err: %w", err)
			}
		}
		if len(messageLeaveClass.ClassId) > 0 {
			err = s.PublishClassEvt(ctx, &mpb.EvtClass{
				Message: &mpb.EvtClass_LeaveClass_{
					LeaveClass: messageLeaveClass,
				},
			})
			if err != nil {
				return false, fmt.Errorf("PublishClassEvt err: %w", err)
			}
		}

	} else {
		err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
			classMembers, err := s.ClassMemberRepo.GetByUserAndCourse(ctx, s.DB, sp.StudentId, sp.Package.CourseId)
			if err != nil {
				return err
			}
			// now, only 1 class of course belong 1 user
			if cm, ok := classMembers[sp.StudentId]; ok {
				err := s.ClassMemberRepo.DeleteByUserIDAndClassID(ctx, s.DB, cm.UserID, cm.ClassID)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return true, fmt.Errorf("s.Handle: %w", err)
		}
		err = s.PublishClassEvt(ctx, &mpb.EvtClass{
			Message: &mpb.EvtClass_LeaveClass_{
				LeaveClass: &mpb.EvtClass_LeaveClass{
					ClassId: sp.Package.ClassId,
					UserId:  sp.StudentId,
				},
			},
		})
		if err != nil {
			return false, fmt.Errorf("PublishClassEvt err: %w", err)
		}
	}
	return true, nil
}

func (s *StudentPackageHandler) PublishClassEvt(ctx context.Context, msg *mpb.EvtClass) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	msgID, err := s.JSM.PublishAsyncContext(ctx, constants.SubjectMasterMgmtClassUpserted, data)
	if err != nil {
		return nats.HandlePushMsgFail(ctx, fmt.Errorf("PublishClassEvent JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, err))
	}
	return nil
}
