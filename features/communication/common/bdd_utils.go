package common

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/communication/common/helpers"
	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/consts"
	noti_repos "github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	"github.com/manabie-com/backend/internal/notification/services/utils"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	studentType         = "student"
	teacherType         = "teacher"
	parentType          = "parent"
	schoolAdminType     = "school admin"
	organizationType    = "organization manager"
	unauthenticatedType = "unauthenticated"

	JPREPSchool     = "JPREPSchool"
	SynersiaSchool  = "SynersiaSchool"
	RenseikaiSchool = "RenseikaiSchool"
	TestingSchool   = "TestingSchool"
	GASchool        = "GASchool"
	KECSchool       = "KECSchool"
	AICSchool       = "AICSchool"
	NSGSchool       = "NSGSchool"
)

func (s *NotificationSuite) CheckReturnStatusCode(ctx context.Context, code string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stt, ok := status.FromError(stepState.ResponseErr)
	if !ok {
		return ctx, fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}
	if stt.Code().String() != code {
		return ctx, fmt.Errorf("expecting %s, got %s status code, message: %s", code, stt.Code().String(), stt.Message())
	}
	return ctx, nil
}

func (s *NotificationSuite) WaitingForFCMIsSent(ctx context.Context) (context.Context, error) {
	time.Sleep(10 * time.Second)
	return ctx, nil
}

func (s *NotificationSuite) CheckReturnStatusCodeAndContainMsg(ctx context.Context, code string, errMsg string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stt, ok := status.FromError(stepState.ResponseErr)
	if !ok {
		return ctx, fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}
	if stt.Code().String() != code {
		return ctx, fmt.Errorf("expecting %s, got %s status code, message: %s", code, stt.Code().String(), stt.Message())
	}
	if !strings.Contains(stt.Message(), errMsg) {
		return ctx, fmt.Errorf("expect error message \"%v\" but got \"%v\" ", errMsg, stt.Message())
	}
	return ctx, nil
}

func (s *NotificationSuite) CheckReturnsErrorMessage(ctx context.Context, errMsg string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	st, ok := status.FromError(stepState.ResponseErr)
	if !ok {
		return ctx, fmt.Errorf("cannot get status error")
	}
	if st.Message() != errMsg {
		return ctx, fmt.Errorf("expect error message \"%v\" but got \"%v\" ", errMsg, st.Message())
	}
	return ctx, nil
}

func ContextWithToken(ctx context.Context, token string) context.Context {
	return contextWithToken(ctx, token)
}

func ContextWithTokenAndTimeOut(ctx context.Context, token string) (context.Context, context.CancelFunc) {
	newCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	return contextWithToken(newCtx, token), cancel
}

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

func contextWithToken(ctx context.Context, token string) context.Context {
	// clear old auth info in context
	ctx = metadata.NewOutgoingContext(ctx, metadata.New(make(map[string]string)))
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", token)
}

func contextWithResourcePath(ctx context.Context, rp string) context.Context {
	claim := interceptors.JWTClaimsFromContext(ctx)
	if claim == nil {
		claim = &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{},
		}
	}
	claim.Manabie.ResourcePath = rp
	return interceptors.ContextWithJWTClaims(ctx, claim)
}

func contextWithUserID(ctx context.Context, userID string) context.Context {
	claim := interceptors.JWTClaimsFromContext(ctx)
	if claim == nil {
		claim = &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{},
		}
	}
	claim.Manabie.UserID = userID
	return interceptors.ContextWithJWTClaims(ctx, claim)
}

func newUserEntity() (*entity.LegacyUser, error) {
	userID := idutil.ULIDNow()
	now := time.Now()
	user := new(entity.LegacyUser)
	firstName := fmt.Sprintf("user-first-name-%s", userID)
	lastName := fmt.Sprintf("user-last-name-%s", userID)
	fullName := helper.CombineFirstNameAndLastNameToFullName(firstName, lastName)
	database.AllNullEntity(user)
	database.AllNullEntity(&user.AppleUser)
	if err := multierr.Combine(
		user.ID.Set(userID),
		user.Email.Set(fmt.Sprintf("valid-user-%s@email.com", userID)),
		user.Avatar.Set(fmt.Sprintf("http://valid-user-%s", userID)),
		user.IsTester.Set(false),
		user.FacebookID.Set(userID),
		user.PhoneVerified.Set(false),
		user.AllowNotification.Set(true),
		user.EmailVerified.Set(false),
		user.FullName.Set(fullName),
		user.FirstName.Set(firstName),
		user.LastName.Set(lastName),
		user.Country.Set(cpb.Country_COUNTRY_VN.String()),
		user.Group.Set(entity.UserGroupStudent),
		user.Birthday.Set(now),
		user.Gender.Set(upb.Gender_FEMALE.String()),
		user.ResourcePath.Set(nil),
		user.CreatedAt.Set(now),
		user.UpdatedAt.Set(now),
		user.DeletedAt.Set(nil),
	); err != nil {
		return nil, errors.Wrap(err, "set value user")
	}

	user.UserAdditionalInfo = entity.UserAdditionalInfo{
		CustomClaims: map[string]interface{}{
			"external-info": "example-info",
		},
	}
	return user, nil
}

func newStudentEntity() (*entity.LegacyStudent, error) {
	now := time.Now()
	student := new(entity.LegacyStudent)
	database.AllNullEntity(student)
	database.AllNullEntity(&student.LegacyUser)
	user, err := newUserEntity()
	if err != nil {
		return nil, errors.Wrap(err, "newUserEntity")
	}
	student.LegacyUser = *user
	schoolID, err := strconv.ParseInt(student.LegacyUser.ResourcePath.String, 10, 32)
	if err != nil {
		return nil, errors.Wrap(err, "strconv.ParseInt")
	}

	if err := multierr.Combine(
		student.ID.Set(student.LegacyUser.ID),
		student.SchoolID.Set(schoolID),
		student.EnrollmentStatus.Set(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED),
		student.StudentExternalID.Set(student.LegacyUser.ID),
		student.StudentNote.Set("this is the note"),
		student.CurrentGrade.Set(1),
		student.TargetUniversity.Set("HUST"),
		student.TotalQuestionLimit.Set(32),
		student.OnTrial.Set(false),
		student.BillingDate.Set(now),
		student.CreatedAt.Set(student.LegacyUser.CreatedAt),
		student.UpdatedAt.Set(student.LegacyUser.UpdatedAt),
		student.DeletedAt.Set(student.LegacyUser.DeletedAt),
		student.PreviousGrade.Set(12),
	); err != nil {
		return nil, errors.Wrap(err, "set value student")
	}

	return student, nil
}

func assignUserGroupToUser(ctx context.Context, dbBob database.QueryExecer, userID string, userGroupIDs []string, resourcePath string) error {
	userGroupMembers := make([]*entity.UserGroupMember, 0)
	for _, userGroupID := range userGroupIDs {
		userGroupMem := &entity.UserGroupMember{}
		database.AllNullEntity(userGroupMem)
		if err := multierr.Combine(
			userGroupMem.UserID.Set(userID),
			userGroupMem.UserGroupID.Set(userGroupID),
			userGroupMem.ResourcePath.Set(fmt.Sprint(resourcePath)),
		); err != nil {
			return err
		}
		userGroupMembers = append(userGroupMembers, userGroupMem)
	}

	if err := (&repository.UserGroupsMemberRepo{}).UpsertBatch(ctx, dbBob, userGroupMembers); err != nil {
		return errors.Wrapf(err, "assignUserGroupToUser")
	}
	return nil
}

func makeSampleNotification(createdUserID, editorID string, schoolID int32) *cpb.Notification {
	infoNotification := &cpb.Notification{
		Data:                       `{"promote": [{"code": "ABC123", "type": "Prime", "amount": "30", "expired_at": "2020-02-13 05:35:44.657508"}, {"code": "ABC123", "type": "Basic", "amount": "20", "expired_at": "2020-03-23 05:35:44.657508"}], "image_url": "https://manabie.com/f50cffe1a8068b04a1b05d1a13b60642.png"}`,
		EditorId:                   editorID,
		CreatedUserId:              createdUserID,
		ReceiverIds:                []string{},
		GenericReceiverIds:         []string{},
		ExcludedGenericReceiverIds: []string{},
		Message: &cpb.NotificationMessage{
			Title: "🎁 Em được gửi tặng 1 món quà đặc biệt! 🎁",
			Content: &cpb.RichText{
				Raw:      "raw richtext: Em được gửi tặng 1 món quà đặc biệt さくら さくら やよいの空は 見わたす限り かすみか雲か " + createdUserID,
				Rendered: "rendered html: Em được gửi tặng 1 món quà đặc biệt さくら さくら やよいの空は 見わたす限り かすみか雲か " + createdUserID,
			},
		},
		Type:   cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED,
		Event:  cpb.NotificationEvent_NOTIFICATION_EVENT_NONE,
		Status: cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED,
		TargetGroup: &cpb.NotificationTargetGroup{
			CourseFilter:    &cpb.NotificationTargetGroup_CourseFilter{Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE},
			GradeFilter:     &cpb.NotificationTargetGroup_GradeFilter{Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE},
			LocationFilter:  &cpb.NotificationTargetGroup_LocationFilter{Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE},
			ClassFilter:     &cpb.NotificationTargetGroup_ClassFilter{Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE},
			SchoolFilter:    &cpb.NotificationTargetGroup_SchoolFilter{Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE},
			UserGroupFilter: &cpb.NotificationTargetGroup_UserGroupFilter{},
		},
		SchoolId:    schoolID,
		IsImportant: false,
	}

	return infoNotification
}

func (s *NotificationSuite) MakeSampleNatsNotification(clientID string, traceID string, schoolID int32, notiType string) *ypb.NatsCreateNotificationRequest {
	notification := &ypb.NatsCreateNotificationRequest{
		ClientId:       clientID,
		SendingMethods: []string{consts.SendingMethodPushNotification},
		NotificationConfig: &ypb.NatsPushNotificationConfig{
			Mode:             consts.NotificationModeNotify,
			PermanentStorage: false,
			Notification: &ypb.NatsNotification{
				Title:   fmt.Sprintf("nats notify %v", traceID),
				Message: "popup message",
				Content: "<h1>hello world</h1>",
			},
			Data: map[string]string{
				"custom_data_type": "eibanam",
			},
		},
		SendTime: &ypb.NatsNotificationSendTime{
			Type: notiType,
		},
		TracingId: traceID,
		SchoolId:  schoolID,
		Target:    &ypb.NatsNotificationTarget{},
	}
	return notification
}

func (s *NotificationSuite) CheckNoneSelectTargetGroup(targetGroup *cpb.NotificationTargetGroup) bool {
	return utils.CheckNoneSelectTargetGroup(mappers.PbToNotificationTargetEnt(targetGroup))
}

func (s *NotificationSuite) GetGrantedLocations(ctx context.Context, notificationID string) ([]string, error) {
	infoNotiAccessPathRepo := &noti_repos.InfoNotificationAccessPathRepo{}
	locationRepo := &noti_repos.LocationRepo{}
	notificationLocationIDs, err := infoNotiAccessPathRepo.GetLocationIDsByNotificationID(ctx, s.BobDBConn, notificationID)
	if err != nil {
		return nil, fmt.Errorf("failed GetNotificationID: %v", err)
	}
	locationIDs, err := locationRepo.GetLowestLocationIDsByIDs(ctx, s.BobDBConn, notificationLocationIDs)
	if err != nil {
		return nil, fmt.Errorf("failed GetLowestLocationIDsByIDs: %v", err)
	}
	return locationIDs, nil
}

func (s *NotificationSuite) GetLocationIDsFromString(ctx context.Context, locationFilter string) ([]string, cpb.NotificationTargetGroupSelect, error) {
	stepState := StepStateFromContext(ctx)
	locations := make([]string, 0)
	switch locationFilter {
	case "random":
		for i, location := range stepState.Organization.DescendantLocations {
			// nolint
			if i == 0 || rand.Intn(2) == 1 {
				locations = append(locations, location.ID)
			}
		}
		return locations, cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST, nil
	case "none":
		return nil, cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE, nil
	case "all":
		return nil, cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL, nil
	default:
		locations := make([]string, 0)
		idxsLocsStr := strings.Split(locationFilter, ",")
		for _, idxLocStr := range idxsLocsStr {
			if idxLocStr == "default" {
				locations = append(locations, stepState.Organization.DefaultLocation.ID)
				continue
			}

			idxLoc, err := strconv.Atoi(idxLocStr)
			if err != nil {
				return nil, cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE, fmt.Errorf("can't convert descendant location index: %v", err)
			}
			if idxLoc <= 0 || idxLoc > helpers.NumberOfNewCenterLocationCreated {
				return nil, cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE, fmt.Errorf("index descendant location out of range")
			}
			locations = append(locations, stepState.Organization.DescendantLocations[idxLoc-1].ID)
		}
		return locations, cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST, nil
	}
}
