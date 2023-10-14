package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"

	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Scan handles reading qrcode
func (s *EntryExitModifierService) Scan(ctx context.Context, req *eepb.ScanRequest) (*eepb.ScanResponse, error) {
	if err := validateScanRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	studentID, err := s.getStudentIDFromQr(ctx, req.QrcodeContent)
	if err != nil {
		return nil, err
	}

	_, err = s.StudentQRRepo.FindByID(ctx, s.DB, studentID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	student, err := s.getStudentWithEEAccess(ctx, studentID, true)
	if err != nil {
		return nil, err
	}

	lockAcquired, err := s.StudentEntryExitRecordsRepo.LockAdvisoryByStudentID(ctx, s.DB, studentID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !lockAcquired {
		return nil, status.Error(codes.Aborted, fmt.Sprintf("%s already processing. skipped", studentID))
	}
	defer s.releaseAdvisoryLockByStudentID(ctx, studentID)

	latestRecord, err := s.StudentEntryExitRecordsRepo.GetLatestRecordByID(ctx, s.DB, studentID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	location, err := time.LoadLocation(req.Timezone)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Error on initializing timezone location: err %v", err))
	}

	if err := s.validateTouchTime(req.TouchTime.AsTime(), latestRecord, location); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	entryExitQueued, err := genEntryExitQueued(studentID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.EntryExitQueueRepo.Create(ctx, s.DB, entryExitQueued)
	if err != nil {
		if strings.Contains(err.Error(), constant.EntryExitQueueAbort) {
			return nil, status.Error(codes.Aborted, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.processScanAndNotify(ctx, req, student, latestRecord, location)
}

// Validate the payload of scan request
func validateScanRequest(req *eepb.ScanRequest) error {
	switch {
	case req.QrcodeContent == "":
		return errors.New("qrcode content cannot be empty")
	case req.TouchTime == nil:
		return errors.New("touch time cannot be empty")
	case req.TouchTime.AsTime().UTC().Truncate(constant.DayDuration).Before(time.Now().UTC().Truncate(constant.DayDuration)):
		return errors.New("touch time should not be past date")
	}
	return nil
}

func isNewDay(current, existing time.Time) bool {
	return !(current.Day() == existing.Day() && current.Month() == existing.Month() && current.Year() == existing.Year())
}

func getLocation(country cpb.Country) (time.Location, error) {
	var zone string
	switch country {
	case cpb.Country_COUNTRY_JP:
		zone = "Asia/Tokyo"
	case cpb.Country_COUNTRY_VN:
		zone = "Asia/Ho_Chi_Minh"
	default:
		zone = "UTC"
	}

	location, err := time.LoadLocation(zone)
	if err != nil {
		return time.Location{}, err
	}

	return *location, err
}

func getEntryExitRecordByTouchEvent(
	touchEvent eepb.TouchEvent,
	latestRecord *entities.StudentEntryExitRecords,
	studentID string,
	touchTime *timestamppb.Timestamp,
) (record *entities.StudentEntryExitRecords, err error) {
	switch touchEvent {
	case eepb.TouchEvent_TOUCH_ENTRY:
		record, err = generateEntryExitCreateRecord(studentID, touchTime, nil)
	case eepb.TouchEvent_TOUCH_EXIT:
		record, err = generateEntryExitUpdateRecord(latestRecord.ID.Int, timestamppb.New(latestRecord.EntryAt.Time), touchTime)
	}

	return record, err
}

func (s *EntryExitModifierService) processScanAndNotify(
	ctx context.Context,
	req *eepb.ScanRequest,
	student *entities.Student,
	latestRecord *entities.StudentEntryExitRecords,
	location *time.Location,
) (*eepb.ScanResponse, error) {
	touchEvent := s.getTouchEvent(latestRecord, req.TouchTime.AsTime(), location)

	entryExitRecord, err := getEntryExitRecordByTouchEvent(touchEvent, latestRecord, student.ID.String, req.TouchTime)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.mutateEntryExitRecordByTouchEvent(ctx, touchEvent, entryExitRecord)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	isNotified, err := s.scanTouchNotify(ctx, student, req.TouchTime, touchEvent)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &eepb.ScanResponse{
		Successful:     true,
		TouchEvent:     touchEvent,
		ParentNotified: isNotified,
		StudentName:    student.GetName(),
	}, nil
}

// Get touch event
func (s *EntryExitModifierService) getTouchEvent(studentEntryExitRecord *entities.StudentEntryExitRecords, touchTimeNow time.Time, location *time.Location) eepb.TouchEvent {
	if studentEntryExitRecord == nil {
		return eepb.TouchEvent_TOUCH_ENTRY
	}

	existingDate := studentEntryExitRecord.EntryAt.Time.In(location)
	currentDate := touchTimeNow.In(location)

	if isNewDay(currentDate, existingDate) {
		return eepb.TouchEvent_TOUCH_ENTRY
	}

	if studentEntryExitRecord.ExitAt.Time.IsZero() {
		return eepb.TouchEvent_TOUCH_EXIT
	}

	return eepb.TouchEvent_TOUCH_ENTRY
}

// Validate touch time is greater than limit
func (s *EntryExitModifierService) validateTouchTime(touchTimeNow time.Time, studentEntryExitRecord *entities.StudentEntryExitRecords, location *time.Location) error {
	var diff float64
	if studentEntryExitRecord == nil {
		return nil
	}

	if studentEntryExitRecord.ExitAt.Time.IsZero() {
		diff = touchTimeNow.Sub(studentEntryExitRecord.EntryAt.Time.In(location)).Minutes()
	} else {
		diff = touchTimeNow.Sub(studentEntryExitRecord.ExitAt.Time.In(location)).Minutes()
	}

	if diff < float64(constant.TouchIntervalInMinutes) {
		return fmt.Errorf("please wait after %v min to scan again", constant.TouchIntervalInMinutes)
	}
	return nil
}

func (s *EntryExitModifierService) mutateEntryExitRecordByTouchEvent(ctx context.Context, touchEvent eepb.TouchEvent, e *entities.StudentEntryExitRecords) (err error) {
	switch touchEvent {
	case eepb.TouchEvent_TOUCH_ENTRY:
		err = s.StudentEntryExitRecordsRepo.Create(ctx, s.DB, e)
	case eepb.TouchEvent_TOUCH_EXIT:
		err = s.StudentEntryExitRecordsRepo.Update(ctx, s.DB, e)
	}

	return err
}

func (s *EntryExitModifierService) scanTouchNotify(
	ctx context.Context,
	student *entities.Student,
	touchTime *timestamppb.Timestamp,
	touchEvent eepb.TouchEvent,
) (bool, error) {
	location, err := getLocation(cpb.Country(cpb.Country_value[student.Country.String]))
	if err != nil {
		return false, err
	}

	return s.notifyParents(ctx, student, &EntryExitNotifyDetails{
		Student:    student,
		TouchEvent: touchEvent,
		TouchTime:  touchTime.AsTime().In(&location),
		RecordType: eepb.RecordType_QR_CODE_SCAN,
	})
}

func (s *EntryExitModifierService) releaseAdvisoryLockByStudentID(ctx context.Context, studentID string) {
	if err := s.StudentEntryExitRecordsRepo.UnLockAdvisoryByStudentID(ctx, s.DB, studentID); err != nil {
		s.logger.Warnf("%v unable to release lock", err.Error())
	}
}

func genEntryExitQueued(studentID string) (*entities.EntryExitQueue, error) {
	e := new(entities.EntryExitQueue)
	database.AllNullEntity(e)

	id := idutil.ULIDNow()
	if err := multierr.Combine(
		e.EntryExitQueueID.Set(id),
		e.StudentID.Set(studentID),
	); err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}

	return e, nil
}
