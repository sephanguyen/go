package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (svc *NotificationModifierService) SyncJprepClassMember(ctx context.Context, req *pb.EvtClassRoom) error {
	now := time.Now()
	switch req.Message.(type) {
	case *pb.EvtClassRoom_JoinClass_:
		msg := req.GetJoinClass()
		strClassID := strconv.Itoa(int(msg.ClassId))

		// get related course data
		courseID, err := svc.ClassRepo.FindCourseIDByClassID(ctx, svc.DB, strClassID)
		if err != nil {
			return fmt.Errorf("failed get CourseID: %v", err)
		}

		notificationClassMember, err := mappers.EventJoinClassRoomToNotificationClassMemberEnt(msg, courseID, constants.JPREPOrgLocation)
		if err != nil {
			return fmt.Errorf("failed EventJoinClassRoomToNotificationClassMemberEnt: %v", err)
		}

		err = svc.NotificationClassMemberRepo.Upsert(ctx, svc.DB, notificationClassMember)
		if err != nil {
			return fmt.Errorf("failed NotificationClassMemberRepo.Upsert: %v", err)
		}
	case *pb.EvtClassRoom_LeaveClass_:
		msg := req.GetLeaveClass()
		strClassID := strconv.Itoa(int(msg.ClassId))
		// get related course data
		courseID, err := svc.ClassRepo.FindCourseIDByClassID(ctx, svc.DB, strClassID)
		if err != nil {
			return fmt.Errorf("failed get CourseID: %v", err)
		}

		classMembers, err := mappers.EventLeaveClassRoomToNotificationClassMemberEnts(msg, courseID, constants.JPREPOrgLocation, now)
		if err != nil {
			return fmt.Errorf("failed EventLeaveClassRoomToNotificationClassMemberEnts: %v", err)
		}
		err = svc.NotificationClassMemberRepo.BulkUpsert(ctx, svc.DB, classMembers)
		if err != nil {
			return fmt.Errorf("failed NotificationClassMemberRepo.BulkUpsert: %v", err)
		}
	}
	return nil
}
