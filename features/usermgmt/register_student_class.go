package usermgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	fatima_entities "github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) signedInUserRegisterClassForAStudent(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	newClassId := idutil.ULIDNow()
	req := &fpb.RegisterStudentClassRequest{
		ClassesInformation: []*fpb.RegisterStudentClassRequest_ClassInformation{
			{
				StudentPackageId: stepState.StudentPackageID,
				StudentId:        stepState.ExistingStudents[0].ID.String,
				StartTime:        timestamppb.New(time.Now()),
				EndTime:          timestamppb.New(time.Now().Add(30 * 24 * time.Hour)),
				ClassId:          newClassId,
			},
		},
	}
	stepState.ClassIds = append(stepState.ClassIds, newClassId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	resp, err := fpb.NewSubscriptionModifierServiceClient(s.FatimaConn).RegisterStudentClass(contextWithToken(ctx), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Request = req
	stepState.Response = resp

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentClassMustStoreInDatabase(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studentPackageClass := &fatima_entities.StudentPackageClass{}
	fields, _ := studentPackageClass.FieldMap()

	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE student_package_id = $1 AND student_id = $2 AND deleted_at IS NULL", strings.Join(fields, ","), studentPackageClass.TableName())

	err := database.Select(ctx, s.FatimaDB, stmt, stepState.StudentPackageID, stepState.ExistingStudents[0].ID).ScanOne(studentPackageClass)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if !golibs.InArrayString(studentPackageClass.ClassID.String, stepState.ClassIds) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("new class not in database expect: %s, actual: %s", studentPackageClass.ClassID.String, stepState.ClassIds)
	}
	return StepStateToContext(ctx, stepState), nil
}
