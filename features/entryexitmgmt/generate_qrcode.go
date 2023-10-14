package entryexitmgmt

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	userConstant "github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"

	"go.uber.org/multierr"
)

func (s *suite) aQrcodeRequestPayloadWithStudentIds(ctx context.Context, numOfStudents string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.SchoolID = strconv.Itoa(int(stepState.CurrentSchoolID))

	numOfStudentsInt, err := strconv.Atoi(numOfStudents)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for i := 0; i < numOfStudentsInt; i++ {
		stepState.StudentID = idutil.ULIDNow()
		stepState.StudentIds = append(stepState.StudentIds, stepState.StudentID)
		_, err := s.aValidUser(StepStateToContext(ctx, stepState), s.BobDBTrace, withID(stepState.StudentID), withUserGroup(cpb.UserGroup_USER_GROUP_STUDENT.String()), withRole(userConstant.RoleStudent))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	time.Sleep(3 * time.Second) // added for kafka sync delay

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generatesQrcodeForTheseStudentIds(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &eepb.GenerateBatchQRCodesRequest{
		StudentIds: stepState.StudentIds,
	}

	ctx, err := s.signedAsAccount(ctx, user)
	if err != nil {
		return ctx, err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = eepb.NewEntryExitServiceClient(s.EntryExitMgmtConn).GenerateBatchQRCodes(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) responseHasNoErrors(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.Response != nil {
		if len(stepState.Response.(*eepb.GenerateBatchQRCodesResponse).Errors) > 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("response errors: %v", stepState.Response.(*eepb.GenerateBatchQRCodesResponse).Errors)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentHasQrVersion(ctx context.Context, version string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentQrRepo := s.StudentQrRepo

	if version == "v1" {
		version = ""
	}

	if len(stepState.StudentIds) == 0 {
		stepState.StudentIds = []string{stepState.StudentID}
	}

	for _, studentID := range stepState.StudentIds {

		u := &entities.StudentQR{}
		database.AllNullEntity(u)

		err := multierr.Combine(
			u.StudentID.Set(studentID),
			u.QRURL.Set(DownloadURLPrefix+base64.URLEncoding.EncodeToString([]byte(studentID))+".png"),
			u.Version.Set(version),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		err = studentQrRepo.Upsert(StepStateToContext(ctx, stepState), s.EntryExitMgmtDBTrace, u)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentShouldHaveUpdatedQrVersion(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(stepState.StudentIds) == 0 {
		stepState.StudentIds = []string{stepState.StudentID}
	}

	studentQrRepo := s.StudentQrRepo

	for _, studentID := range stepState.StudentIds {

		studentQR, err := studentQrRepo.FindByID(StepStateToContext(ctx, stepState), s.EntryExitMgmtDBTrace, studentID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if studentQR.Version.String != CurrentUpdatedVersion {
			return StepStateToContext(ctx, stepState), fmt.Errorf("version is not updated student_id: %s version: %s", studentID, studentQR.Version.String)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
