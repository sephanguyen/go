package communication

import (
	"context"
	//nolint:gosec
	"crypto/md5"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/godogutil"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/entities"
	bobPb "github.com/manabie-com/backend/pkg/genproto/bob"
	yasuoPb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) hasCreatedCourse(ctx context.Context, role string, numCourse int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	reqs := make([]*yasuoPb.UpsertCoursesRequest, 0)

	for i := 0; i < numCourse; i++ {
		courseID := idutil.ULIDNow()
		stepState.courseIDs = append(stepState.courseIDs, courseID)
		schoolID, err := strconv.ParseInt(stepState.schoolID, 10, 64)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		id := rand.Int31()

		country := bobPb.COUNTRY_VN
		grade, _ := i18n.ConvertIntGradeToString(country, 7)
		req := &yasuoPb.UpsertCoursesRequest{
			Courses: []*yasuoPb.UpsertCoursesRequest_Course{
				{
					Id:       courseID,
					Name:     fmt.Sprintf("course-%d", id),
					Country:  country,
					Subject:  bobPb.SUBJECT_BIOLOGY,
					SchoolId: int32(schoolID),
					Grade:    grade,
				},
			},
		}
		reqs = append(reqs, req)
	}
	switch role {
	case schoolAdmin:
		token := s.getToken(schoolAdmin)
		for _, req := range reqs {
			_, err := yasuoPb.NewCourseServiceClient(s.yasuoConn).UpsertCourses(contextWithToken(ctx, token), req)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}

		return StepStateToContext(ctx, stepState), nil
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("not supported for role: %s", role)
	}
}

func (s *suite) hasCreatedAStudentWithGradeAndParentInfo(ctx context.Context, role string) (context.Context, error) {
	switch role {
	case schoolAdmin:
		return s.schoolAdminHasCreatedStudentWithParentInfo(ctx)
	default:
		return ctx, fmt.Errorf("not supported role")
	}
}

func (s *suite) hasAddedCreatedCourseForStudent(ctx context.Context, role string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch role {
	case schoolAdmin:
		student := stepState.profile.defaultStudent
		parent := stepState.profile.defaultParent
		pInfo := []parentInfo{
			{
				id:    parent.id,
				name:  parent.name,
				email: parent.email,
			},
		}

		sInfo := studentInfo{
			id:        student.id,
			name:      student.name,
			email:     student.email,
			parents:   pInfo,
			courseIDs: stepState.courseIDs,
		}
		return s.schoolAdminAddCoursesForStudent(ctx, sInfo)
	default:
		return ctx, fmt.Errorf("not supported role")
	}
}

func (s *suite) schoolAdminAddCoursesForStudent(ctx context.Context, student studentInfo) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	token := s.getToken(schoolAdmin)
	schoolID, err := strconv.ParseInt(stepState.schoolID, 10, 64)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	parentProfiles := make([]*ypb.UpdateStudentRequest_ParentProfile, 0)
	for _, p := range student.parents {
		parentProfiles = append(parentProfiles, &ypb.UpdateStudentRequest_ParentProfile{
			Id:    p.id,
			Email: p.email,
			Name:  p.name,
		})
	}
	for _, courseID := range student.courseIDs {
		_, err = ypb.NewUserModifierServiceClient(s.yasuoConn).UpdateStudent(contextWithToken(ctx, token), &ypb.UpdateStudentRequest{
			SchoolId: int32(schoolID),
			StudentProfile: &ypb.UpdateStudentRequest_StudentProfile{
				Id:               student.id,
				Email:            student.email,
				Name:             fmt.Sprintf("student-name+%s", student.name),
				Grade:            7,
				EnrollmentStatus: cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			},
			ParentProfiles: parentProfiles,
			StudentPackageProfiles: []*ypb.UpdateStudentRequest_StudentPackageProfile{
				{
					Id: &ypb.UpdateStudentRequest_StudentPackageProfile_CourseId{
						CourseId: courseID,
					},
					StartTime: timestamppb.New(time.Now().Add(time.Hour * -20)),
					EndTime:   timestamppb.New(time.Now().Add(time.Hour * 24 * 10)),
				},
			},
		})
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("school admin unable to add course to student: %w", err)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) isAtNotificationPageOnCMS(arg1 string) error {
	return nil
}

func (s *suite) sendsNotificationWithRequiredFieldsToStudentAndParent(ctx context.Context, role string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch role {
	case schoolAdmin:
		noti, err := s.newNotification(stepState.schoolID, stepState.profile.schoolAdmin.id)
		if err != nil {
			return ctx, err
		}
		receiverIDs := make([]string, 0)
		receiverIDs = append(receiverIDs, stepState.profile.defaultStudent.id)
		noti = s.notificationWithReceiver(receiverIDs, noti)
		ctx, err = godogutil.MultiErrChain(ctx,
			s.upsertNotification, noti,
			s.sendNotification, noti,
		)
		if err != nil {
			return ctx, err
		}
		stepState.notification = noti
		return StepStateToContext(ctx, stepState), nil
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("not supported role")
	}
}

func (s *suite) newNotification(schoolIDStr string, schoolAdminID string) (*cpb.Notification, error) {
	schoolID, err := strconv.ParseInt(schoolIDStr, 10, 64)
	if err != nil {
		return nil, err
	}
	infoNotification := &cpb.Notification{
		Data:          `{"promote": [{"code": "ABC123", "type": "Prime", "amount": "30", "expired_at": "2020-02-13 05:35:44.657508"}, {"code": "ABC123", "type": "Basic", "amount": "20", "expired_at": "2020-03-23 05:35:44.657508"}], "image_url": "https://manabie.com/f50cffe1a8068b04a1b05d1a13b60642.png"}`,
		EditorId:      schoolAdminID,
		CreatedUserId: schoolAdminID,
		ReceiverIds:   nil,
		Message: &cpb.NotificationMessage{
			Title: "🎁 Em được gửi tặng 1 món quà đặc biệt! 🎁",
			Content: &cpb.RichText{
				Raw:      "raw richtext: Em được gửi tặng 1 món quà đặc biệt さくら さくら やよいの空は 見わたす限り かすみか雲か ",
				Rendered: "rendered html: Em được gửi tặng 1 món quà đặc biệt さくら さくら やよいの空は 見わたす限り かすみか雲か ",
			},
		},
		// this is the default value for fields status, target_group
		Status: cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT,
		TargetGroup: &cpb.NotificationTargetGroup{
			CourseFilter: &cpb.NotificationTargetGroup_CourseFilter{},
			GradeFilter:  &cpb.NotificationTargetGroup_GradeFilter{},
			UserGroupFilter: &cpb.NotificationTargetGroup_UserGroupFilter{
				UserGroups: []cpb.UserGroup{cpb.UserGroup_USER_GROUP_PARENT, cpb.UserGroup_USER_GROUP_STUDENT},
			},
		},
		SchoolId: int32(schoolID),
	}

	return infoNotification, nil
}

func (s *suite) notificationWithScheduledAt(scheduledAt time.Time, noti *cpb.Notification) *cpb.Notification {
	noti.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED
	noti.ScheduledAt = timestamppb.New(scheduledAt)
	return noti
}

func (s *suite) notificationWithCourse(selectType cpb.NotificationTargetGroupSelect, courseIDs []string, noti *cpb.Notification) *cpb.Notification {
	noti.TargetGroup.CourseFilter = &cpb.NotificationTargetGroup_CourseFilter{
		Type:      selectType,
		CourseIds: courseIDs,
	}
	return noti
}

func (s *suite) notificationWithGrade(selectType cpb.NotificationTargetGroupSelect, grades []int32, noti *cpb.Notification) *cpb.Notification {
	noti.TargetGroup.GradeFilter = &cpb.NotificationTargetGroup_GradeFilter{
		Type:   selectType,
	}
	return noti
}

func (s *suite) notificationWithRecipientType(recipientType string, noti *cpb.Notification) *cpb.Notification {
	recipient := make([]cpb.UserGroup, 0)
	switch recipientType {
	case "student and parent":
		recipient = append(recipient, cpb.UserGroup_USER_GROUP_STUDENT)
		recipient = append(recipient, cpb.UserGroup_USER_GROUP_PARENT)
	case student:
		recipient = append(recipient, cpb.UserGroup_USER_GROUP_STUDENT)
	case parent:
		recipient = append(recipient, cpb.UserGroup_USER_GROUP_PARENT)
	default:
		return nil
	}

	noti.TargetGroup.UserGroupFilter = &cpb.NotificationTargetGroup_UserGroupFilter{
		UserGroups: recipient,
	}
	return noti
}

func (s *suite) notificationWithCourseType(ctx context.Context, courseType string, noti *cpb.Notification) *cpb.Notification {
	stepState := StepStateFromContext(ctx)
	switch courseType {
	case "empty":
		return s.notificationWithCourse(cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE, nil, noti)
	case "All courses":
		return s.notificationWithCourse(cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL, nil, noti)
	default:
		// current ly this course type will be the list of course's name separated by &
		courseNames := strings.Split(courseType, "&")
		courseIDs := make([]string, 0)
		for _, name := range courseNames {
			if strings.TrimSpace(name) == "All" {
				return s.notificationWithCourseType(ctx, "All courses", noti)
			}
			course, ok := stepState.courseInfos[strings.TrimSpace(name)]
			if !ok {
				return nil
			}
			courseIDs = append(courseIDs, course.id)
		}
		return s.notificationWithCourse(cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST, courseIDs, noti)
	}
}

func (s *suite) notificationWithGradeType(ctx context.Context, gradeType string, noti *cpb.Notification) *cpb.Notification {
	stepState := StepStateFromContext(ctx)
	switch gradeType {
	case "empty":
		return s.notificationWithGrade(cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE, nil, noti)
	case "All grade":
		return s.notificationWithGrade(cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL, nil, noti)
	default:
		// current ly this grade type will be the list of grade's name separated by &
		gradeArgs := strings.Split(gradeType, "&")
		grades := make([]int32, 0)
		for _, grade := range gradeArgs {
			if strings.TrimSpace(grade) == "All" {
				return s.notificationWithGradeType(ctx, "All grade", noti)
			}

			valid, userName := userNameOfGradeIs(strings.TrimSpace(grade))
			if !valid {
				return nil
			}

			user, ok := stepState.studentInfos[userName]
			if !ok {
				return nil
			}
			grades = append(grades, user.grade)

		}
		return s.notificationWithGrade(cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST, grades, noti)
	}
}

func (s *suite) notificationWithIndividualType(ctx context.Context, individualType string, noti *cpb.Notification) *cpb.Notification {
	stepState := StepStateFromContext(ctx)
	switch individualType {
	case "empty":
		return noti
	default:
		valid, userName := userNameIs(individualType)
		if !valid {
			return nil
		}
		user, ok := stepState.users[userName]
		if !ok {
			return nil
		}
		return s.notificationWithReceiver([]string{user.id}, noti)
	}

}

func (s *suite) notificationWithReceiver(receiverIDs []string, noti *cpb.Notification) *cpb.Notification {
	noti.ReceiverIds = append(noti.ReceiverIds, receiverIDs...)
	return noti
}

func (s *suite) notificationWithMediaIDs(mediaIDs []string, noti *cpb.Notification) *cpb.Notification {
	noti.Message.MediaIds = mediaIDs
	return noti
}

func (s *suite) checkStoreNotification(ctx context.Context, expectedNoti *cpb.Notification) error {
	curNoti := &entities.InfoNotification{}
	fields := database.GetFieldNames(curNoti)
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE notification_id = $1 AND deleted_at IS NULL;`, strings.Join(fields, ","), curNoti.TableName())

	err := database.Select(ctx, s.bobDB, query, expectedNoti.NotificationId).ScanOne(curNoti)
	if err != nil {
		return err
	}

	err = s.checkInfoNotificationResponse(ctx, expectedNoti, curNoti)
	if err != nil {
		return err
	}

	notiMsg := &entities.InfoNotificationMsg{}
	fields = database.GetFieldNames(notiMsg)
	query = fmt.Sprintf(`SELECT %s FROM %s WHERE notification_msg_id = $1 AND deleted_at IS NULL;`, strings.Join(fields, ","), notiMsg.TableName())

	err = database.Select(ctx, s.bobDB, query, curNoti.NotificationMsgID).ScanOne(notiMsg)
	if err != nil {
		return err
	}

	err = s.checkInfoNotificationMsgResponse(s.toinfoNotificationMessageEnt(expectedNoti.Message), notiMsg)
	if err != nil {
		return err
	}

	return nil
}

func (s *suite) storeNotificationSuccessfully(ctx context.Context) (context.Context, error) {
	st := StepStateFromContext(ctx)
	err := s.checkStoreNotification(ctx, st.notification)
	if err != nil {
		return ctx, err
	}
	return StepStateToContext(ctx, st), nil
}

func (s *suite) checkInfoNotificationResponse(ctx context.Context, expect *cpb.Notification, cur *entities.InfoNotification) error {
	if string(cur.Data.Bytes) != expect.Data {
		return fmt.Errorf("expect notification data %v but got %v", expect.Data, string(cur.Data.Bytes))
	}

	if cur.EditorID.String != expect.EditorId {
		return fmt.Errorf("expect notification editor id %v but got %v", expect.EditorId, cur.EditorID.String)
	}

	if cur.CreatedUserID.String != expect.CreatedUserId {
		return fmt.Errorf("expect notification editor id %v but got %v", expect.CreatedUserId, cur.CreatedUserID.String)
	}

	if cur.Type.String != expect.Type.String() {
		return fmt.Errorf("expect notification type %v but got %v", cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String(), cur.Type.String)
	}

	if cur.Status.String != expect.Status.String() {
		return fmt.Errorf("expect notification status %v but got %v", expect.Status.String(), cur.Status.String)
	}

	targetGroup, err := cur.GetTargetGroup()
	if err != nil {
		return err
	}

	err = s.checkTargetGroupFilter(expect.TargetGroup, targetGroup)
	if err != nil {
		return err
	}

	if cur.Event.String != expect.Event.String() {
		return fmt.Errorf("expect notification event %v but got %v", expect.Event, cur.Event.String)
	}

	if cur.Status.String != expect.Status.String() {
		return fmt.Errorf("expect notification status %v but got %v", expect.Status, cur.Status.String)
	}

	if expect.Status == cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED && !expect.ScheduledAt.AsTime().Round(time.Second).Equal(cur.ScheduledAt.Time.Round(time.Second)) {
		return fmt.Errorf("expect notification scheduled at %v but got %v", expect.ScheduledAt.AsTime(), cur.ScheduledAt.Time)
	}

	schoolAdminRepo := &repositories.SchoolAdminRepo{}
	schoolAdmin, err := schoolAdminRepo.Get(ctx, s.bobDB, database.Text(expect.EditorId))
	if err != nil {
		return err
	}

	if cur.Owner.Int != schoolAdmin.SchoolID.Int {
		return fmt.Errorf("expect notification school id %v but got %v", schoolAdmin.SchoolID, cur.Owner.Int)
	}

	return nil
}

func (s *suite) checkInfoNotificationMsgResponse(expect *entities.InfoNotificationMsg, cur *entities.InfoNotificationMsg) error {
	if expect.Title.String != cur.Title.String {
		return fmt.Errorf("expect notification message title %v but got %v", expect.Title, cur.Title.String)
	}

	expectContent, _ := expect.GetContent()

	currentContent, err := cur.GetContent()
	if err != nil {
		return err
	}

	if expectContent.Raw != currentContent.Raw {
		return fmt.Errorf("expect notification message content raw %v but got %v", expect.Content, cur.Content)
	}

	url, _ := generateUploadURL(s.Cfg.Storage.Endpoint, s.Cfg.Storage.Bucket, expectContent.RenderedURL)
	if url != currentContent.RenderedURL {
		return fmt.Errorf("expect notification message content rendered url %v but got %v", url, currentContent.RenderedURL)
	}
	if len(expect.MediaIDs.Elements) != len(cur.MediaIDs.Elements) {
		return fmt.Errorf("expect notification message medias %v but got %v", expect.MediaIDs, cur.MediaIDs)
	}

	for i := range expect.MediaIDs.Elements {
		if expect.MediaIDs.Elements[i].String != cur.MediaIDs.Elements[i].String {
			return fmt.Errorf("expect notification message medias %v but got %v", expect.MediaIDs, cur.MediaIDs)
		}
	}

	return nil
}

func (s *suite) checkTargetGroupFilter(expectPb *cpb.NotificationTargetGroup, cur *entities.InfoNotificationTarget) error {
	if expectPb.CourseFilter.Type.String() != cur.CourseFilter.Type {
		return fmt.Errorf("expect course filter select type %v but got %v", expectPb.CourseFilter.Type.String(), cur.CourseFilter.Type)
	}
	if len(expectPb.CourseFilter.CourseIds) != len(cur.CourseFilter.CourseIDs) {
		return fmt.Errorf("expect number of target course_ids %v but got %v", len(expectPb.CourseFilter.CourseIds), len(cur.CourseFilter.CourseIDs))
	}

	if expectPb.GradeFilter.Type.String() != cur.GradeFilter.Type {
		return fmt.Errorf("expect grade filter select type %v but got %v", expectPb.GradeFilter.Type.String(), cur.GradeFilter.Type)
	}

	if len(expectPb.UserGroupFilter.UserGroups) != len(cur.UserGroupFilter.UserGroups) {
		return fmt.Errorf("expect number of target user group %v but got %v", len(expectPb.UserGroupFilter.UserGroups), len(cur.UserGroupFilter.UserGroups))
	}

	return nil
}

func (s *suite) toinfoNotificationMessageEnt(msg *cpb.NotificationMessage) *entities.InfoNotificationMsg {
	e := &entities.InfoNotificationMsg{}

	err := multierr.Combine(
		e.NotificationMsgID.Set(msg.NotificationMsgId),
		e.Title.Set(msg.Title),
		e.Content.Set(&entities.RichText{
			Raw:         msg.Content.Raw,
			RenderedURL: msg.Content.Rendered,
		}),
		e.MediaIDs.Set(msg.MediaIds),
		e.UpdatedAt.Set(msg.CreatedAt.AsTime()),
		e.CreatedAt.Set(msg.UpdatedAt.AsTime()),
	)

	if err != nil {
		return nil
	}
	return e
}

func generateUploadURL(endpoint, bucket, content string) (url, fileName string) {
	h := md5.New()
	io.WriteString(h, content)
	fileName = "/content/" + fmt.Sprintf("%x.html", h.Sum(nil))

	return endpoint + "/" + bucket + fileName, fileName
}

func (s *suite) receivesTheNotificationInTheirDevice(ctx context.Context, role string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	token := ""
	id := ""
	switch role {
	case student:
		token = s.getToken(student)
		id = stepState.profile.defaultStudent.id
	case parent, "parent P1":
		token = s.getToken(parent)
		id = stepState.profile.defaultParent.id
	default:
		return ctx, fmt.Errorf("not supported role")
	}

	return s.RetrieveNotificationDetail(ctx, stepState.notification.NotificationId, token, id, cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW)
}

func (s *suite) RetrieveNotificationDetail(ctx context.Context, notificationID string, token string, userID string, expectStatus cpb.UserNotificationStatus) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &bpb.RetrieveNotificationDetailRequest{
		NotificationId: stepState.notification.NotificationId,
	}
	resp, err := bpb.NewNotificationReaderServiceClient(s.bobConn).RetrieveNotificationDetail(contextWithToken(ctx, token), req)
	if err != nil {
		return ctx, err
	}

	if resp.Item == nil {
		return ctx, fmt.Errorf("expect get notification with id %s but not found", stepState.notification.NotificationId)
	}

	if resp.Item.NotificationId != stepState.notification.NotificationId {
		return ctx, fmt.Errorf("expect notification id %s but got %v", stepState.notification.NotificationId, resp.Item.NotificationId)
	}

	if resp.UserNotification.UserId != userID {
		return ctx, fmt.Errorf("expect user id %s but got %s", userID, resp.UserNotification.UserId)
	}
	if resp.UserNotification.Status != expectStatus {
		return ctx, fmt.Errorf("expect user noti status %s but got %s", expectStatus, resp.UserNotification.Status)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) schoolAdminHasSavedADraftNotificationWithRequiredFields(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	noti, err := s.newNotification(stepState.schoolID, stepState.profile.schoolAdmin.id)
	if err != nil {
		return ctx, err
	}
	noti.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT
	receiverIDs := make([]string, 0)
	receiverIDs = append(receiverIDs, stepState.profile.defaultStudent.id)
	noti = s.notificationWithReceiver(receiverIDs, noti)

	ctx, err = s.upsertNotification(ctx, noti)
	if err != nil {
		return ctx, err
	}
	stepState.notification = noti
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) schoolAdminSendsThatDraftNotificationForStudentAndParent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.sendNotification(ctx, stepState.notification)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertNotification(ctx context.Context, noti *cpb.Notification) (context.Context, error) {
	token := s.getToken(schoolAdmin)
	req := &ypb.UpsertNotificationRequest{
		Notification: noti,
	}
	resp, err := ypb.NewNotificationModifierServiceClient(s.yasuoConn).UpsertNotification(contextWithToken(ctx, token), req)
	if err != nil {
		return ctx, err
	}
	noti.NotificationId = resp.NotificationId
	return ctx, nil
}

func (s *suite) sendNotification(ctx context.Context, noti *cpb.Notification) (context.Context, error) {
	token := s.getToken(schoolAdmin)
	_, err := ypb.NewNotificationModifierServiceClient(s.yasuoConn).SendNotification(contextWithToken(ctx, token), &ypb.SendNotificationRequest{
		NotificationId: noti.NotificationId,
	})
	if err != nil {
		return ctx, err
	}
	noti.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_SENT
	return ctx, nil
}
