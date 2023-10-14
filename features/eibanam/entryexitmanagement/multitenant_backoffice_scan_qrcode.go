package entryexitmanagement

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) scannerIsSetupOn(ctx context.Context, resourcePath string) error {
	s.ScannerResourcePath = resourcePath
	return nil
}

func (s *suite) thereIsAnExistingStudentWithQrCodeFrom(ctx context.Context, organization string) error {
	// Setup context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ctx, err := s.loginsWithResourcePathFrom(ctx, "school admin", organization)
	if err != nil {
		return err
	}

	s.UserGroupInContext = constant.UserGroupSchoolAdmin
	err = s.createStudentWithResourcePath(ctx)
	if err != nil {
		return err
	}
	ctx = contextWithTokenForGrpcCall(s, ctx)
	// get the student id to be used for generate qr code
	reqGenerateQR := &eepb.GenerateBatchQRCodesRequest{
		StudentIds: []string{s.Response.(*upb.CreateStudentResponse).StudentProfile.Student.UserProfile.UserId},
	}
	respGenerateQR, err := eepb.NewEntryExitServiceClient(s.entryExitMgmtConn).
		GenerateBatchQRCodes(ctx, reqGenerateQR)
	if err != nil {
		return err
	}

	if len(respGenerateQR.Errors) > 0 {
		return fmt.Errorf("response errors: %v", respGenerateQR.Errors)
	}

	return nil
}

func (s *suite) thisStudentScansQrCode(ctx context.Context) error {
	studentID := s.Response.(*upb.CreateStudentResponse).StudentProfile.Student.UserProfile.UserId
	qrcodeContent := base64.URLEncoding.EncodeToString([]byte(studentID))

	s.Request = &eepb.ScanRequest{
		QrcodeContent: qrcodeContent,
		TouchTime:     timestamppb.New(time.Now()),
	}
	// get the auth scanner setup
	_, err := s.loginsWithResourcePathFrom(ctx, "school admin", s.ScannerResourcePath)
	if err != nil {
		return err
	}

	return nil
}

func (s *suite) scannerShouldReturn(ctx context.Context, result string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	// calll grpc with permission
	s.stepState.UserGroupInContext = constant.UserGroupSchoolAdmin
	ctx = contextWithTokenForGrpcCall(s, ctx)
	s.Response, s.ResponseErr = eepb.NewEntryExitServiceClient(s.entryExitMgmtConn).Scan(ctx, s.Request.(*eepb.ScanRequest))

	switch result {
	case "successfully":
		if !s.Response.(*eepb.ScanResponse).Successful || s.ResponseErr != nil {
			return errors.New("student expected to scan qr code successfully")
		}
	case "unsuccessfully":
		if s.ResponseErr == nil {
			return errors.New("student expected to scan qr code unsuccessfully")
		}
	}
	return nil
}

func reqWithOnlyStudentInfo(schoolID int32) *upb.CreateStudentRequest {
	randomId := idutil.ULIDNow()
	return &upb.CreateStudentRequest{
		SchoolId: schoolID,
		StudentProfile: &upb.CreateStudentRequest_StudentProfile{
			Email:             fmt.Sprintf("%v@example.com", randomId),
			Password:          fmt.Sprintf("password-%v", randomId),
			Name:              fmt.Sprintf("user-%v", randomId),
			CountryCode:       cpb.Country_COUNTRY_VN,
			EnrollmentStatus:  upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			PhoneNumber:       fmt.Sprintf("phone-number-%v", randomId),
			StudentExternalId: fmt.Sprintf("student-external-id-%v", randomId),
			StudentNote:       fmt.Sprintf("some random student note %v", randomId),
			Grade:             5,
		},
	}
}
