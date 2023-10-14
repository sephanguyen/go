package entryexitmgmt

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/entryexitmgmt/repositories"
	"github.com/manabie-com/backend/internal/entryexitmgmt/services"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	user_repo "github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	entities_user "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	userConstant "github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	bpb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) thereIsAnExistingStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.SchoolID = strconv.Itoa(int(stepState.CurrentSchoolID))
	stepState.StudentID = idutil.ULIDNow()

	_, err := s.aValidUser(StepStateToContext(ctx, stepState),
		s.BobDBTrace,
		withID(stepState.StudentID),
		withUserGroup(cpb.UserGroup_USER_GROUP_STUDENT.String()),
		withResourcePath(stepState.ResourcePath),
		withRole(userConstant.RoleStudent),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	time.Sleep(5 * time.Second) // added for kafka sync delay

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentHasRecord(ctx context.Context, existingTouchEvent string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	entryexitRepo := s.StudentEntryExitRecordsRepo
	var entryAt time.Time
	var exitAt time.Time
	// if err := s.deleteRecords(StepStateToContext(ctx, stepState), stepState.StudentID); err != nil {
	// 	return StepStateToContext(ctx, stepState), err
	// }

	if existingTouchEvent == "no entry and exit" {
		return StepStateToContext(ctx, stepState), nil
	}

	now := time.Now().UTC()

	switch existingTouchEvent {
	case "entry":
		entryAt = time.Now()
		exitAt = time.Time{}
	case "entry date equivalent to previous date in UTC":
		entryAt = time.Date(now.Year(), now.Month(), now.Day(), 23, 0, 0, 0, time.UTC)
		entryAt = entryAt.AddDate(0, 0, -1)
		exitAt = time.Time{}
	case "exit":
		entryAt = time.Now().Add(-14 * time.Minute)
		exitAt = time.Now().Add(-7 * time.Minute)
	case "past entry":
		entryAt = time.Now().AddDate(0, 0, -1)
		exitAt = time.Time{}
	case "past completed":
		entryAt = time.Now().AddDate(0, 0, -1)
		exitAt = entryAt.Add(2 * time.Hour)
	}

	entryexitEntity, err := generateEntryExitRecord(stepState.StudentID, entryAt, exitAt)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if err := entryexitRepo.Create(ctx, s.EntryExitMgmtDBTrace, entryexitEntity); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentParentHasExistingDevice(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Get Parent
	var err error
	studentParentRepo := &repositories.StudentParentRepo{}

	stepState.ParentIDs, err = studentParentRepo.GetParentIDsByStudentID(ctx, s.BobDBTrace, stepState.StudentID)
	if err != nil {
		return nil, err
	}

	// Update the Device Token of Parents
	if len(stepState.ParentIDs) != 0 {
		for _, userID := range stepState.ParentIDs {
			_, err = s.updateDeviceToken(ctx, userID)
			if err != nil {
				return nil, err
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func generateV2QRContent(studentID string, encryptionKey string) (string, error) {
	secretByte, err := hex.DecodeString(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("hex.DecodeString: %v", err)
	}

	// encrypt student ID
	crypt := &services.CryptV2{}
	encryptedStudentID, err := crypt.Encrypt(studentID, secretByte)
	if err != nil {
		return "", fmt.Errorf("encrypt: %v", err)
	}

	qrCodeByte, _ := json.Marshal(services.QrCodeContent{QrCode: encryptedStudentID, Version: constant.V2})
	return base64.URLEncoding.EncodeToString(qrCodeByte), nil
}

func (s *suite) studentScansQrcodeWithDateInTimeZoneWithInvalidEncryption(ctx context.Context, timePeriod string, timeZone string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	qrCodeContent, err := generateV2QRContent(stepState.StudentID, "b88ddefd7e8b70386aafa6b8b500feaa")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.TimeZone = timeZone

	location, err := time.LoadLocation(timeZone)
	if err != nil {
		return ctx, err
	}

	var touchTime time.Time
	switch timePeriod {
	case "Present":
		touchTime = time.Now().Add(7 * time.Minute)
	case "Past":
		touchTime = time.Now().AddDate(0, -1, -10)
	}

	req := &eepb.ScanRequest{
		QrcodeContent: qrCodeContent,
		TouchTime:     timestamppb.New(touchTime.In(location)),
		Timezone:      timeZone,
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = eepb.NewEntryExitServiceClient(s.EntryExitMgmtConn).Scan(contextWithToken(ctx), req)

	stepState.ParentNotified = false
	if stepState.ResponseErr == nil && stepState.Response != nil {
		stepState.ParentNotified = stepState.Response.(*eepb.ScanResponse).ParentNotified
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentScansQrcodeWithDateInTimeZone(ctx context.Context, timePeriod string, timeZone string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	qrCodeContent, err := generateV2QRContent(stepState.StudentID, "93db70a8de02328365f69a3f3eb9d2ed")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.TimeZone = timeZone

	location, err := time.LoadLocation(timeZone)
	if err != nil {
		return ctx, err
	}

	var touchTime time.Time
	now := time.Now()
	switch timePeriod {
	case "Present":
		touchTime = time.Now().Add(7 * time.Minute)
	case "Past":
		touchTime = time.Now().AddDate(0, -1, -10)
	case "Fixed Time":
		// There is a step where the existing entry date is the past UTC date (fixed date time)
		// Using the time.Now() can be flaky in some of the test so use this prevent it
		// For example, the test ran at 10-18-2022 17:00:00 UTC (10-19-2022 2AM JST) the previous UTC date is set to 10-17-2022 23:00:00 (10-18-2022 8AM JST)
		// The date is different, there will cause a flaky test
		// The value here is equivalent to 3PM JST
		touchTime = time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, time.UTC)
	}

	req := &eepb.ScanRequest{
		QrcodeContent: qrCodeContent,
		TouchTime:     timestamppb.New(touchTime.In(location)),
		Timezone:      timeZone,
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = eepb.NewEntryExitServiceClient(s.EntryExitMgmtConn).Scan(contextWithToken(ctx), req)
	// declared initially as when we scan again qr code scenario
	stepState.ParentNotified = false
	if stepState.ResponseErr == nil && stepState.Response != nil {
		stepState.ParentNotified = stepState.Response.(*eepb.ScanResponse).ParentNotified
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentScansAgain(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	_, err := s.studentScansQrcodeWithDateInTimeZone(ctx, "Present", stepState.TimeZone)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) scanReturnsStatusCode(ctx context.Context, expectedCode string) (context.Context, error) {
	return s.receivesStatusCode(ctx, expectedCode)
}

func (s *suite) touchTypeIsRecorded(ctx context.Context, latestTouchEvent string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if err := s.getTouchEvent(ctx, latestTouchEvent, stepState.StudentID); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) parentReceivesNotificationStatus(ctx context.Context, notifStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if notifStatus != "Successfully" {
		// no response for student without parents and parents that are not notified
		if !stepState.ParentNotified || len(stepState.ParentIDs) == 0 {
			return StepStateToContext(ctx, stepState), nil
		}
		return StepStateToContext(ctx, stepState), fmt.Errorf("not expecting any notification but parent was notified")
	}

	if !stepState.ParentNotified {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting notification to be successful but parent was not notified")
	}

	// parent notification should not be saved
	query := `SELECT count(notification_msg_id) FROM info_notification_msgs WHERE info_notification_msgs.title IN ('Entry & Exit Activity', '入退室記録')`
	var count int

	if err := s.BobDBTrace.QueryRow(ctx, query).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if count >= 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected %d info notification message created", count)
	}

	// wait for notification to be sent
	time.Sleep(10 * time.Second)

	ctx, err := s.checkUserNotification(ctx, stepState.ParentIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error in checking notification: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentHasParent(ctx context.Context, parentInfo string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studentParentRepo := &repositories.StudentParentRepo{}
	if parentInfo == "Existing" {
		parent, err := s.createParent(ctx, 1)
		if err != nil {
			return ctx, err
		}
		err = s.createStudentParentRelationship(
			ctx,
			stepState.StudentID,
			[]string{parent.LegacyUser.ID.String},
			upb.FamilyRelationship_name[int32(upb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER)],
		)
		if err != nil {
			return ctx, err
		}
	}

	getParentIDs, err := studentParentRepo.GetParentIDsByStudentID(ctx, s.BobDBTrace, stepState.StudentID)
	if err != nil {
		return nil, err
	}

	stepState.ParentIDs = getParentIDs

	time.Sleep(3 * time.Second) // added for kafka sync delay

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) nameOfTheStudentIsDisplayedOnWelcomeScreen(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if strings.TrimSpace(stepState.Response.(*eepb.ScanResponse).StudentName) == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student name cannot be retrieved")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getTouchEvent(ctx context.Context, touchEvent string, studentID string) error {
	entryexitRepo := s.StudentEntryExitRecordsRepo

	latestRecord, err := entryexitRepo.GetLatestRecordByID(ctx, s.EntryExitMgmtDBTrace, studentID)
	if err != nil {
		return err
	}

	if latestRecord == nil {
		return fmt.Errorf("unexpected record: latest record is nil")
	}

	switch touchEvent {
	case "TOUCH_ENTRY":
		if latestRecord.EntryAt.Time.IsZero() {
			return fmt.Errorf("touch entry should have saved entry date")
		}

		if !latestRecord.ExitAt.Time.IsZero() {
			return fmt.Errorf("touch entry should not have exit date")
		}

	case "TOUCH_EXIT":
		if latestRecord.ExitAt.Time.IsZero() {
			return fmt.Errorf("touch exit should have exit date")
		}

		if latestRecord.EntryAt.Time.IsZero() {
			return fmt.Errorf("touch exit should also have entry date")
		}
	default:
		return fmt.Errorf("touch event %s is not supported", touchEvent)
	}

	return nil
}

func (s *suite) deleteRecords(ctx context.Context, studentID string) error {
	query := `UPDATE student_entryexit_records SET deleted_at = $1 WHERE student_id = $2`

	_, err := s.EntryExitMgmtDBTrace.Exec(ctx, query, database.Timestamptz(time.Now()), studentID)
	if err != nil {
		return fmt.Errorf("unable update deleted_at for student entryexit %s: %w", studentID, err)
	}
	return nil
}

func (s *suite) createParent(ctx context.Context, schoolID int32) (*entities_user.Parent, error) {
	stepState := StepStateFromContext(ctx)

	id := idutil.ULIDNow()

	parent := &entities_user.Parent{}
	database.AllNullEntity(parent)
	database.AllNullEntity(&parent.LegacyUser)
	err := multierr.Combine(
		parent.LegacyUser.ID.Set(id),
		parent.LegacyUser.Email.Set(fmt.Sprintf("email-%s@example.com", id)),
		parent.LegacyUser.LastName.Set(fmt.Sprintf("name-%s", id)),
		parent.LegacyUser.Country.Set(bpb.COUNTRY_VN.String()),
		parent.LegacyUser.PhoneNumber.Set(fmt.Sprintf("phone-number-%s", id)),
		parent.LegacyUser.ResourcePath.Set(stepState.ResourcePath),
		parent.LegacyUser.FullName.Set(fmt.Sprintf("parent-%v", id)),
		parent.LegacyUser.FirstName.Set(fmt.Sprintf("parent-first-name-%v", id)),
		parent.LegacyUser.LastName.Set(fmt.Sprintf("parent-first-name-%v", id)),
		parent.ID.Set(id),
		parent.SchoolID.Set(schoolID),
		parent.ResourcePath.Set(stepState.ResourcePath),
	)
	if err != nil {
		return nil, err
	}
	parent.LegacyUser.UserAdditionalInfo.Password = id

	parentRepo := &user_repo.ParentRepo{}
	if err := parentRepo.Create(ctx, s.BobDBTrace, parent); err != nil {
		return nil, err
	}

	return parent, nil
}

func (s *suite) loginsToBackofficeApp(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, user)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentScansQrcodeRequestCountWithDateInTimeZone(ctx context.Context, requestCount string, timePeriod string, timeZone string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	qrCodeContent, err := generateV2QRContent(stepState.StudentID, "93db70a8de02328365f69a3f3eb9d2ed")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.TimeZone = timeZone

	location, err := time.LoadLocation(timeZone)
	if err != nil {
		return ctx, err
	}

	var touchTime time.Time
	now := time.Now()
	switch timePeriod {
	case "Present":
		touchTime = time.Now().Add(7 * time.Minute)
	case "Past":
		touchTime = time.Now().AddDate(0, -1, -10)
	case "Fixed Time":
		// There is a step where the existing entry date is the past UTC date (fixed date time)
		// Using the time.Now() can be flaky in some of the test so use this prevent it
		// For example, the test ran at 10-18-2022 17:00:00 UTC (10-19-2022 2AM JST) the previous UTC date is set to 10-17-2022 23:00:00 (10-18-2022 8AM JST)
		// The date is different, there will cause a flaky test
		// The value here is equivalent to 3PM JST
		touchTime = time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, time.UTC)
	}

	req := &eepb.ScanRequest{
		QrcodeContent: qrCodeContent,
		TouchTime:     timestamppb.New(touchTime.In(location)),
		Timezone:      timeZone,
	}

	stepState.RequestSentAt = time.Now()

	requestCountInt, _ := strconv.Atoi(requestCount)

	var wg sync.WaitGroup
	wg.Add(requestCountInt)

	// replicate a duplicate sent by goroutine
	for i := 0; i < requestCountInt; i++ {
		go func(data int) {
			defer wg.Done()
			fmt.Printf("\n%s sent: %v, time: %v", stepState.StudentID, data, time.Now())
			stepState.Response, stepState.ResponseErr = eepb.NewEntryExitServiceClient(s.EntryExitMgmtConn).Scan(contextWithToken(ctx), req)
		}(i + 1)
	}

	wg.Wait()

	// declared initially as when we scan again qr code scenario
	stepState.ParentNotified = false
	if stepState.ResponseErr == nil && stepState.Response != nil {
		stepState.ParentNotified = stepState.Response.(*eepb.ScanResponse).ParentNotified
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentHasNoMultipleRecord(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if err := try.Do(func(attempt int) (bool, error) {
		query := `SELECT count(student_id) FROM student_entryexit_records s WHERE s.student_id = $1`
		var count int

		if err := s.EntryExitMgmtDBTrace.QueryRow(ctx, query, stepState.StudentID).Scan(&count); err != nil {
			return false, err
		}
		if count == 1 {
			return false, nil
		}
		if count > 1 {
			return false, fmt.Errorf("unexpected %d record created for student", count)
		}
		time.Sleep(1 * time.Second)
		return attempt < 10, fmt.Errorf("record not created")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
