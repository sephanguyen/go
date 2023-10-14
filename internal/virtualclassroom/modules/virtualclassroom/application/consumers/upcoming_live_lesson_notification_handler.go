package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/notification/consts"
	consumer "github.com/manabie-com/backend/internal/notification/transports/nats"
	serviceConstants "github.com/manabie-com/backend/internal/virtualclassroom/constants"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
	bpb "github.com/manabie-com/backend/pkg/genproto/bob"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type UpcomingLiveLessonNotificationHandler struct {
	Logger                         *zap.Logger
	JSM                            nats.JetStreamManagement
	BobDB                          database.Ext
	WrapperConnection              *support.WrapperDBConnection
	VirtualLessonRepo              infrastructure.VirtualLessonRepo
	LiveLessonSentNotificationRepo infrastructure.LiveLessonSentNotificationRepo
	LessonMemberRepo               infrastructure.LessonMemberRepo
	StudentParentRepo              infrastructure.StudentParentRepo
	UserRepo                       infrastructure.UserRepo
}

type LiveLessonParticipant struct {
	StudentID string
	ParentIDs []string
	SchoolID  string
	Country   string
}

type LiveLessonNotificationDetails struct {
	LessonID        string
	LessonStartTime time.Time
	Interval        string
	Recipients      []string
	Country         string
}

type LocalizedTextData struct {
	Title      string
	Message15M string
	Message24H string
}

// represents notification intervals
const (
	H24 string = "24h"
	M15 string = "15m"
)

var countryToLocalizedTextMap = map[string]LocalizedTextData{
	"COUNTRY_JP": {
		Title:      "ライブ授業の通知",
		Message15M: "15分後の %s %s にライブ授業が始まります。",
		Message24H: "24時間後の %s %s にライブ授業が始まります。",
	},
}

func (u *UpcomingLiveLessonNotificationHandler) Handle(ctx context.Context, msg []byte) (bool, error) {
	conn, err := u.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return false, nil
	}

	u.Logger.Info("[UpcomingLiveLessonNotificationEvent]: Received message on",
		zap.String("subject", constants.SubjectUpcomingLiveLessonNotification),
		zap.String("queue", constants.QueueUpcomingLiveLessonNotification),
	)
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	upcomingLiveLessonNotificationEventData := &vpb.UpcomingLiveLessonNotificationRequest{}
	if err := proto.Unmarshal(msg, upcomingLiveLessonNotificationEventData); err != nil {
		u.logError("Failed to parse vpb.UpcomingLiveLessonNotificationRequest", err)
		return false, err
	}

	lessonIDs := upcomingLiveLessonNotificationEventData.GetLessonIds()
	virtualLessons, err := u.VirtualLessonRepo.GetVirtualLessonsByLessonIDs(ctx, conn, lessonIDs)
	if err != nil {
		u.logError("Query failed for GetVirtualLessonsByLessonIDs()", err)
		return false, err
	}

	now := time.Now()
	errStrings := []string{}
	for _, virtualLesson := range virtualLessons {
		startTime := virtualLesson.StartTime
		lessonID := virtualLesson.LessonID
		interval := u.getInterval(now, startTime)

		// could not determine interval, we skip this lesson for now
		if len(interval) == 0 {
			errStrings = append(errStrings, fmt.Sprintf("Could not determine interval for lesson %s on start time %s, skipping", lessonID, startTime.String()))
			continue
		}

		sentNotificationCount, err := u.LiveLessonSentNotificationRepo.GetLiveLessonSentNotificationCount(ctx, conn, lessonID, interval)
		if err != nil {
			errStrings = append(errStrings, fmt.Sprintf("u.LiveLessonSentNotificationRepo.GetLiveLessonSentNotificationCount: %s", err.Error()))
			continue
		}

		if sentNotificationCount <= 0 {
			if err = database.ExecInTx(ctx, conn, func(ctx context.Context, tx pgx.Tx) error {
				err = u.LiveLessonSentNotificationRepo.CreateLiveLessonSentNotificationRecord(ctx, tx, lessonID, interval, now)
				if err != nil {
					return err
				}
				participants, err := u.getNotificationParticipants(ctx, tx, lessonID)
				if err != nil {
					return err
				}
				if len(participants) == 0 {
					return nil
				}
				err = u.sendNotificationsToParticipants(ctx, virtualLesson, participants, interval)
				if err != nil {
					return err
				}
				return nil
			}); err != nil {
				// combine the errors and return them later
				errStrings = append(errStrings, err.Error())
			}
		}
	}

	if len(errStrings) > 0 {
		return false, errors.New(strings.Join(errStrings, "\n"))
	}

	return true, nil
}

func (u *UpcomingLiveLessonNotificationHandler) getInterval(from, to time.Time) string {
	diff := to.Sub(from)
	interval := ""

	if diff.Hours() <= 24 && diff.Minutes() > 15 {
		// notification hasn't been sent within 24 hours
		interval = H24
	} else if diff.Minutes() <= 15 && diff.Minutes() > 0 {
		// notification hasn't been sent within 15 minutes
		interval = M15
	}

	return interval
}

func (u *UpcomingLiveLessonNotificationHandler) sendNotificationsToParticipants(ctx context.Context, lesson *domain.VirtualLesson, participants []*LiveLessonParticipant, interval string) error {
	lessonID := lesson.LessonID
	startTime := lesson.StartTime

	u.logInfo(fmt.Sprintf("Sending notifications to recipients for %s lesson", lessonID))

	// each participant data contains a student to parent pair and their country id
	// we group them in such a way so that we'll be able to send localized messages according to their country id
	// and not the generalized country id retrieved from a random student on a lesson
	recipientCountryMap := u.createCountryParticipantMap(participants)

	for country, recipients := range recipientCountryMap {
		if err := u.notify(ctx, &LiveLessonNotificationDetails{
			LessonID:        lessonID,
			LessonStartTime: startTime,
			Recipients:      recipients,
			Country:         country,
			Interval:        interval,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (u *UpcomingLiveLessonNotificationHandler) createCountryParticipantMap(participants []*LiveLessonParticipant) map[string][]string {
	recipientCountryMap := make(map[string][]string)

	for _, participant := range participants {
		recipientIDs := append([]string{participant.StudentID}, participant.ParentIDs...)

		recipientCountryMap[participant.Country] = append(recipientCountryMap[participant.Country],
			recipientIDs...,
		)
	}

	return recipientCountryMap
}

func (u *UpcomingLiveLessonNotificationHandler) notify(ctx context.Context, details *LiveLessonNotificationDetails) error {
	resourcePath := golibs.ResourcePathFromCtx(ctx)
	schoolID, err := strconv.ParseInt(resourcePath, 10, 32)
	if err != nil {
		return err
	}
	customData := map[string]string{
		"lesson_id": details.LessonID,
	}
	customDataJSON, _ := json.Marshal(&customData)
	notification := u.generateMessageDetails(details)
	data := &ypb.NatsCreateNotificationRequest{
		ClientId:       serviceConstants.ClientIDNatsVirtualClassroomService,
		SendingMethods: []string{consts.SendingMethodPushNotification},
		Target: &ypb.NatsNotificationTarget{
			GenericUserIds: details.Recipients,
		},
		NotificationConfig: &ypb.NatsPushNotificationConfig{
			Mode:             consts.NotificationModeNotify,
			PermanentStorage: false,
			Notification:     notification,
			Data: map[string]string{
				"custom_data_type":  "lesson",
				"custom_data_value": string(customDataJSON),
			},
		},
		SendTime: &ypb.NatsNotificationSendTime{
			Type: consts.NotificationTypeImmediate,
		},
		TracingId: uuid.New().String(),
		SchoolId:  int32(schoolID),
	}
	msg, err := proto.Marshal(data)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	if err := try.Do(func(attempt int) (bool, error) {
		if _, err := u.JSM.PublishContext(ctx, consumer.SubjectNotificationCreated, msg); err == nil {
			return false, nil
		}
		time.Sleep(1 * time.Second)
		return attempt < 4, fmt.Errorf("publish error")
	}); err != nil {
		return err
	}

	return nil
}

func (u *UpcomingLiveLessonNotificationHandler) getNotificationParticipants(ctx context.Context, tx database.QueryExecer, lessonID string) ([]*LiveLessonParticipant, error) {
	notificationParticipants := []*LiveLessonParticipant{}

	learnerIDs, err := u.LessonMemberRepo.GetLearnerIDsByLessonID(ctx, tx, lessonID)
	if err != nil {
		return nil, fmt.Errorf("error in LessonMemberRepo.GetLearnerIDsByLessonID, lesson %s: %w", lessonID, err)
	}
	if len(learnerIDs) == 0 {
		return nil, nil
	}

	lessonMemberUsers, err := u.UserRepo.GetUsersByIDs(ctx, u.BobDB, learnerIDs)
	if err != nil {
		return nil, fmt.Errorf("error in UserRepo.GetUsersByIDs, user IDs from lesson %s: %w", lessonID, err)
	}
	studentIDs := []string{}
	for _, member := range lessonMemberUsers {
		studentIDs = append(studentIDs, member.ID)
	}

	studentToParentMap := map[string][]string{}
	studentParents, err := u.StudentParentRepo.GetStudentParents(ctx, tx, studentIDs)
	if err != nil {
		return nil, err
	}
	for _, studentParent := range studentParents {
		studentToParentMap[studentParent.StudentID] = append(studentToParentMap[studentParent.StudentID], studentParent.ParentID)
	}

	for _, member := range lessonMemberUsers {
		notificationParticipants = append(notificationParticipants, &LiveLessonParticipant{
			StudentID: member.ID,
			ParentIDs: studentToParentMap[member.ID],
			Country:   member.Country,
		})
	}

	return notificationParticipants, nil
}

func (u *UpcomingLiveLessonNotificationHandler) generateMessageDetails(details *LiveLessonNotificationDetails) *ypb.NatsNotification {
	// convert to local time based on country
	localTimestamp := details.LessonStartTime.In(timeutil.Timezone(bpb.Country(bpb.Country_value[details.Country])))
	startTime := localTimestamp.Format("15:04")
	startDate := localTimestamp.Format("2006/01/02")

	// default texts
	title := "Live lesson reminder"
	message := "Live lesson will start in 15 minutes from %s on %s."
	if details.Interval == H24 {
		message = "Live lesson will start in 24 hours from %s on %s."
	}
	message = fmt.Sprintf(message, startTime, startDate)

	localizedText, ok := countryToLocalizedTextMap[details.Country]
	if ok {
		title = localizedText.Title

		message = localizedText.Message15M
		if details.Interval == H24 {
			message = localizedText.Message24H
		}

		message = fmt.Sprintf(message, startTime, startDate)
	}

	return &ypb.NatsNotification{
		Title:   title,
		Message: message,
		Content: "<h1>" + message + "</h1>",
	}
}

func (u *UpcomingLiveLessonNotificationHandler) logError(msg string, err error) {
	u.Logger.Sugar().Errorf("[UpcomingLiveLessonNotificationEvent]: %s %w", msg, err)
}

func (u *UpcomingLiveLessonNotificationHandler) logInfo(msg string) {
	u.Logger.Sugar().Infof("[UpcomingLiveLessonNotificationEvent]: %s", msg)
}
