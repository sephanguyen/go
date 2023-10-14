package mastermgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	fatima_entities "github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/infrastructure/repo"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/utils"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) hasScheduledClassToReserveClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now()
	studentID := stepState.StudentID
	studentPackageID := stepState.Response.(*fpb.AddStudentPackageCourseResponse).StudentPackageId
	courseID := stepState.CourseIDs[0]
	classID := stepState.ClassIds[1]

	req := &mpb.ScheduleStudentClassRequest{
		StudentId:        studentID,
		StudentPackageId: studentPackageID,
		CourseId:         courseID,
		ClassId:          classID,
		StartTime:        timestamppb.Now(),
		EndTime:          timestamppb.New(now.Add(100 * 24 * time.Hour)),
		EffectiveDate:    timestamppb.New(now.Add(30 * 24 * time.Hour)),
	}

	stepState.RequestSentAt = time.Now()
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = mpb.NewScheduleClassServiceClient(s.MasterMgmtConn).
		ScheduleStudentClass(contextWithToken(s, utils.SignCtx(ctx)), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) scheduleClassToReserveClassAgain(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now()
	preReq := stepState.Request.(*mpb.ScheduleStudentClassRequest)
	classID := stepState.ClassIds[2]

	req := &mpb.ScheduleStudentClassRequest{
		StudentId:        preReq.StudentId,
		StudentPackageId: preReq.StudentPackageId,
		CourseId:         preReq.CourseId,
		ClassId:          classID,
		StartTime:        timestamppb.New(now.Add(-10 * 24 * time.Hour)),
		EndTime:          timestamppb.New(now.Add(100 * 24 * time.Hour)),
		EffectiveDate:    timestamppb.New(now.Add(30 * 24 * time.Hour)),
	}

	stepState.RequestSentAt = time.Now()
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = mpb.NewScheduleClassServiceClient(s.MasterMgmtConn).
		ScheduleStudentClass(contextWithToken(s, utils.SignCtx(ctx)), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) reserveClassMustStoreInDatabaseCorrect(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	reserveClassDTO := &repo.ReserveClassDTO{}
	reserveClassDTOs := repo.ReserveClassDTOs{}
	fields, _ := reserveClassDTO.FieldMap()
	preReq := stepState.Request.(*mpb.ScheduleStudentClassRequest)
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE student_package_id = $1 AND student_id = $2 AND deleted_at IS NULL", strings.Join(fields, ","), reserveClassDTO.TableName())

	err := database.Select(ctx, s.BobDB, stmt, preReq.StudentPackageId, preReq.StudentId).ScanAll(&reserveClassDTOs)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if len(reserveClassDTOs) != 1 || reserveClassDTOs[0].ClassID.String != stepState.ClassIds[2] || reserveClassDTOs[0].CourseID.String != preReq.CourseId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("upsert record on reserve class incorrect")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) callFuncWrapperRegisterStudentClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	reserveClassDTO := &repo.ReserveClassDTO{}
	fields, _ := reserveClassDTO.FieldMap()
	preReq := stepState.Request.(*mpb.ScheduleStudentClassRequest)
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE student_package_id = $1 AND student_id = $2 AND deleted_at IS NULL", strings.Join(fields, ","), reserveClassDTO.TableName())

	err := database.Select(ctx, s.BobDB, stmt, preReq.StudentPackageId, preReq.StudentId).ScanOne(reserveClassDTO)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := &fpb.WrapperRegisterStudentClassRequest{
		ReserveClassesInformation: []*fpb.WrapperRegisterStudentClassRequest_ReserveClassInformation{
			{
				StudentId:        reserveClassDTO.StudentID.String,
				StudentPackageId: reserveClassDTO.StudentPackageID.String,
				CourseId:         reserveClassDTO.CourseID.String,
				ClassId:          reserveClassDTO.ClassID.String,
			},
		},
	}
	stepState.Response, stepState.ResponseErr = fpb.NewSubscriptionModifierServiceClient(s.FatimaConn).
		WrapperRegisterStudentClass(contextWithToken(s, utils.SignCtx(ctx)), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentClassMustStoreInDatabase(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studentPackageClass := &fatima_entities.StudentPackageClass{}
	fields, _ := studentPackageClass.FieldMap()
	preReq := stepState.Request.(*mpb.ScheduleStudentClassRequest)

	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE student_package_id = $1 AND student_id = $2 AND deleted_at IS NULL", strings.Join(fields, ","), studentPackageClass.TableName())

	err := database.Select(ctx, s.FatimaDB, stmt, preReq.StudentPackageId, preReq.StudentId).ScanOne(studentPackageClass)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if studentPackageClass.ClassID.String != preReq.ClassId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("new class not in database expect: %s, actual: %s", studentPackageClass.ClassID.String, preReq.ClassId)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) retrieveScheduledClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studentID := stepState.StudentID

	time.Sleep(time.Second) // wait for student package class sync from fatima to bob
	req := &mpb.RetrieveScheduledStudentClassRequest{
		StudentId: studentID,
	}

	stepState.RequestSentAt = time.Now()
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = mpb.NewScheduleClassServiceClient(s.MasterMgmtConn).
		RetrieveScheduledStudentClass(contextWithToken(s, utils.SignCtx(ctx)), req)
	return StepStateToContext(ctx, stepState), nil
}
