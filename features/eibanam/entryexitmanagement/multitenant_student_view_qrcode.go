package entryexitmanagement

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"
)

func (s *suite) withResourcePathFrom(ctx context.Context, role, organization string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	roleStrSplit := strings.Split(role, " ")[0]
	organizationStrSplit := strings.Split(organization, " ")

	resourcePathOrdinalStr := organizationStrSplit[len(organizationStrSplit)-1]
	s.stepState.ResourcePath = resourcePathOrdinalStr
	resourcePathOrdinal, err := strconv.Atoi(resourcePathOrdinalStr)
	if err != nil {
		return err
	}
	s.CurrentSchoolID = int32(resourcePathOrdinal)

	err = s.signedInAsAccountWithResourcePath(ctx, "school admin", resourcePathOrdinalStr)
	if err != nil {
		return err
	}

	s.UserGroupInContext = constant.UserGroupSchoolAdmin
	if roleStrSplit == "student" {

		studentID := idutil.ULIDNow()
		err := s.aValidStudentInDBWithResourcePath(ctx, studentID, resourcePathOrdinalStr)
		if err != nil {
			return fmt.Errorf("err s.aValidStudentInDBWithResourcePath: %w", err)
		}

		s.stepState.ResponseStack.Push(studentID)
	}
	return nil
}

func (s *suite) loginsLearnerApp(ctx context.Context, student string) error {
	studentNum, err := getStudentNumber(student)
	if err != nil {
		return err
	}

	// get the student id to be used for generate qr code
	studentID := s.ResponseStack.Responses[studentNum-1].(string)
	err = s.saveCredential(studentID, constant.UserGroupStudent, int64(s.getSchoolId()))
	if err != nil {
		return err
	}

	return nil
}

func (s *suite) seeQrCode(ctx context.Context, signedInUser, result, student string) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	studentNum, err := getStudentNumber(student)
	if err != nil {
		return err
	}

	s.UserGroupInContext = constant.UserGroupStudent
	studentID := s.ResponseStack.Responses[studentNum-1].(string)

	ctx = contextWithTokenForGrpcCall(s, ctx)
	req := &eepb.RetrieveStudentQRCodeRequest{
		StudentId: studentID,
	}
	s.Response, s.ResponseErr = eepb.NewEntryExitServiceClient(s.entryExitMgmtConn).RetrieveStudentQRCode(ctx, req)

	var qrURL string
	if s.Response != nil && s.ResponseErr == nil {
		qrURL = strings.TrimSpace(s.Response.(*eepb.RetrieveStudentQRCodeResponse).QrUrl)
	}

	switch result {
	case "can":
		if qrURL == "" {
			return errors.New("unable to retrieve student qr code")
		}
	case "cannot":
		if qrURL != "" {
			return errors.New("student expected not to see qr code")
		}
	}

	return nil
}

func (s *suite) hasExistingQrCode(ctx context.Context, role string) error {
	// Setup context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	studentNum, err := getStudentNumber(role)
	if err != nil {
		return err
	}

	s.UserGroupInContext = constant.UserGroupSchoolAdmin
	ctx = contextWithTokenForGrpcCall(s, ctx)
	// get the student id to be used for generate qr code
	studentID := s.ResponseStack.Responses[studentNum-1].(string)

	reqGenerateQR := &eepb.GenerateBatchQRCodesRequest{
		StudentIds: []string{studentID},
	}
	respGenerateQR, err := eepb.NewEntryExitServiceClient(s.entryExitMgmtConn).GenerateBatchQRCodes(ctx, reqGenerateQR)
	if err != nil {
		return err
	}

	if len(respGenerateQR.Errors) > 0 {
		return fmt.Errorf("response errors: %v", respGenerateQR.Errors)
	}
	return nil
}

func getStudentNumber(studentStr string) (int, error) {
	// get student S1, S2
	studentSplit := strings.Split(studentStr, " ")[1]
	// get the number of student from S1, S2
	studentNumStr := studentSplit[len(studentSplit)-1:]
	studentNum, err := strconv.Atoi(studentNumStr)
	if err != nil {
		return 0, err
	}

	return studentNum, nil
}
