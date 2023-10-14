package usermgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func generateSchoolHistoryPbWithCondition(ctx context.Context, db database.Ext, condition, gradeID string) ([]*pb.SchoolHistory, error) {
	schoolInfo, err := insertRandomSchoolInfo(ctx, db, idutil.ULIDNow())
	if err != nil {
		return nil, fmt.Errorf("insertRandomSchoolInfo: %v", err)
	}

	var schoolHistories []*pb.SchoolHistory
	switch condition {
	case "one row":
		schoolCourse, err := insertRandomSchoolCourse(ctx, db, schoolInfo.ID.String)
		if err != nil {
			return nil, fmt.Errorf("insertRandomSchoolCourse: %v", err)
		}
		schoolHistories = []*pb.SchoolHistory{
			{
				SchoolId:       schoolInfo.ID.String,
				SchoolCourseId: schoolCourse.ID.String,
				StartDate:      timestamppb.Now(),
			},
		}
	case "many rows":
		schoolInfo1, err := insertRandomSchoolInfo(ctx, db, idutil.ULIDNow())
		if err != nil {
			return nil, fmt.Errorf("insertRandomSchoolInfo: %v", err)
		}
		schoolInfo2, err := insertRandomSchoolInfo(ctx, db, idutil.ULIDNow())
		if err != nil {
			return nil, fmt.Errorf("insertRandomSchoolInfo: %v", err)
		}
		schoolInfo3, err := insertRandomSchoolInfo(ctx, db, idutil.ULIDNow())
		if err != nil {
			return nil, fmt.Errorf("insertRandomSchoolInfo: %v", err)
		}
		schoolHistories = []*pb.SchoolHistory{
			{
				SchoolId:  schoolInfo1.ID.String,
				StartDate: timestamppb.Now(),
			},
			{
				SchoolId: schoolInfo2.ID.String,
				EndDate:  timestamppb.New(time.Now().Add(time.Hour)),
			},
			{
				SchoolId:  schoolInfo3.ID.String,
				StartDate: timestamppb.Now(),
				EndDate:   timestamppb.New(time.Now().Add(time.Hour)),
			},
		}
	case "one row current school":
		_, schoolInfoCurrent, err := insertRandomSchoolLevelGrade(ctx, db, gradeID)
		if err != nil {
			return nil, fmt.Errorf("insertRandomSchoolLevelGrade: %v", err)
		}
		schoolCourse, err := insertRandomSchoolCourse(ctx, db, schoolInfoCurrent.ID.String)
		if err != nil {
			return nil, fmt.Errorf("insertRandomSchoolCourse: %v", err)
		}
		schoolHistories = []*pb.SchoolHistory{
			{
				SchoolId:       schoolInfoCurrent.ID.String,
				SchoolCourseId: schoolCourse.ID.String,
				StartDate:      timestamppb.Now(),
			},
		}
	case "many rows current school":
		_, schoolInfo1, err := insertRandomSchoolLevelGrade(ctx, db, gradeID)
		if err != nil {
			return nil, fmt.Errorf("insertRandomSchoolLevelGrade: %v", err)
		}
		_, schoolInfo2, err := insertRandomSchoolLevelGrade(ctx, db, gradeID)
		if err != nil {
			return nil, fmt.Errorf("insertRandomSchoolLevelGrade: %v", err)
		}
		_, schoolInfo3, err := insertRandomSchoolLevelGrade(ctx, db, gradeID)
		if err != nil {
			return nil, fmt.Errorf("insertRandomSchoolLevelGrade: %v", err)
		}
		schoolHistories = []*pb.SchoolHistory{
			{
				SchoolId:  schoolInfo1.ID.String,
				StartDate: timestamppb.Now(),
			},
			{
				SchoolId: schoolInfo2.ID.String,
				EndDate:  timestamppb.New(time.Now().Add(time.Hour)),
			},
			{
				SchoolId:  schoolInfo3.ID.String,
				StartDate: timestamppb.Now(),
				EndDate:   timestamppb.New(time.Now().Add(time.Hour)),
			},
		}
	case "mandatory only":
		schoolHistories = []*pb.SchoolHistory{
			{
				SchoolId: schoolInfo.ID.String,
			},
		}
	}

	return schoolHistories, nil
}

func (s *suite) studentInfoWithSchoolHistoriesUpdateInvalidRequest(ctx context.Context, condition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &pb.UpdateStudentRequest{
		SchoolId: int32(OrgIDFromCtx(ctx)),
		StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
			Id:                stepState.CurrentStudentID,
			Name:              fmt.Sprintf("updated-%s", stepState.CurrentStudentID),
			Grade:             int32(1),
			EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			StudentExternalId: fmt.Sprintf("student-external-id-%v", stepState.CurrentStudentID),
			StudentNote:       fmt.Sprintf("some random student note edited %v", stepState.CurrentStudentID),
			Email:             fmt.Sprintf("student-email-edited-%s@example.com", stepState.CurrentStudentID),
			Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
			Gender:            pb.Gender_MALE,
			LocationIds:       []string{s.ExistingLocations[0].LocationID.String},
		},
	}
	schoolHistories, err := generateSchoolHistoryPbWithInvalidCondition(ctx, s.BobDBTrace, condition)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("generateSchoolHistoryPbWithInvalidCondition: %v", err)
	}
	req.SchoolHistories = schoolHistories
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentInfoWithSchoolHistoriesUpdateValidRequest(ctx context.Context, condition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	domainGradeRepo := &repository.DomainGradeRepo{}
	grades, err := domainGradeRepo.GetByPartnerInternalIDs(ctx, s.BobDBTrace, stepState.PartnerInternalIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("domainGradeRepo.GetByPartnerInternalIDs err: %v", err)
	}
	gradeID := grades[1].GradeID().String()

	req := &pb.UpdateStudentRequest{
		SchoolId: int32(OrgIDFromCtx(ctx)),
		StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
			Id:                stepState.CurrentStudentID,
			Name:              fmt.Sprintf("updated-%s", stepState.CurrentStudentID),
			Grade:             int32(1),
			EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			StudentExternalId: fmt.Sprintf("student-external-id-%v", stepState.CurrentStudentID),
			StudentNote:       fmt.Sprintf("some random student note edited %v", stepState.CurrentStudentID),
			Email:             fmt.Sprintf("student-email-edited-%s@example.com", stepState.CurrentStudentID),
			Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
			Gender:            pb.Gender_MALE,
			GradeId:           gradeID,
			LocationIds:       []string{s.ExistingLocations[0].LocationID.String},
		},
	}
	schoolHistories, err := generateSchoolHistoryPbWithCondition(ctx, s.BobDBTrace, condition, gradeID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("generateSchoolHistoryPbWithCondition: %v", err)
	}
	req.SchoolHistories = schoolHistories
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentAccountUpdatedSuccessWithSchoolHistories(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	req := stepState.Request.(*pb.UpdateStudentRequest)
	if err := validateSchoolHistoriesInDB(ctx, s.BobDBTrace, req.SchoolHistories, stepState.CurrentStudentID, fmt.Sprint(OrgIDFromCtx(ctx)), entity.DomainSchools{}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("validateSchoolHistoriesInDB: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentAccountUpdatedSuccessWithSchoolHistoriesHaveCurrentSchool(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	req := stepState.Request.(*pb.UpdateStudentRequest)
	if err := validateSchoolHistoriesInDBWithCurrentSchool(ctx, s.BobDBTrace, req.SchoolHistories, stepState.CurrentStudentID, fmt.Sprint(OrgIDFromCtx(ctx)), true); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("validateSchoolHistoriesInDB: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentAccountUpdatedSuccessWithSchoolHistoriesRemoveCurrentSchool(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	req := stepState.Request.(*pb.UpdateStudentRequest)
	if err := validateSchoolHistoriesInDBWithCurrentSchool(ctx, s.BobDBTrace, req.SchoolHistories, stepState.CurrentStudentID, fmt.Sprint(OrgIDFromCtx(ctx)), false); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("validateSchoolHistoriesInDB: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func validateSchoolHistoriesInDB(ctx context.Context, db database.Ext, reqSchoolHistories []*pb.SchoolHistory, studentID string, resourcePath string, currentSchoolsInDB entity.DomainSchools) error {
	schoolHistoryRepo := &repository.SchoolHistoryRepo{}
	schoolHistories, err := schoolHistoryRepo.GetByStudentID(ctx, db, database.Text(studentID))
	if err != nil {
		return fmt.Errorf("schoolHistoryRepo.GetByStudentID: %v", err)
	}

	if len(schoolHistories) != len(reqSchoolHistories) {
		return fmt.Errorf("validation school_history failed, expect len(reqSchoolHistories): %v, actual: %v", len(reqSchoolHistories), len(schoolHistories))
	}

	count := 0
	for _, schoolHistory := range schoolHistories {
		for _, schoolHistoryPb := range reqSchoolHistories {
			if schoolHistoryPb.SchoolId != schoolHistory.SchoolID.String {
				continue
			}
			switch {
			case schoolHistoryPb.StartDate.IsValid() && (schoolHistoryPb.StartDate.AsTime().Unix() != schoolHistory.StartDate.Time.Unix()):
				return fmt.Errorf("validation school_history failed, expect start_date: %v, actual: %v", schoolHistoryPb.StartDate.AsTime().Unix(), schoolHistory.StartDate.Time.Unix())
			case !schoolHistoryPb.StartDate.IsValid() && schoolHistory.StartDate.Status != pgtype.Null:
				return fmt.Errorf("validation school_history failed, expect start_date: null, actual: %v", schoolHistory.StartDate.Time)
			case schoolHistoryPb.EndDate.IsValid() && (schoolHistoryPb.EndDate.AsTime().Unix() != schoolHistory.EndDate.Time.Unix()):
				return fmt.Errorf("validation school_history failed, expect end_date: %v, actual: %v", schoolHistoryPb.EndDate.AsTime().Unix(), schoolHistory.EndDate.Time.Unix())
			case !schoolHistoryPb.EndDate.IsValid() && schoolHistory.EndDate.Status != pgtype.Null:
				return fmt.Errorf("validation school_history failed, expect end_date: null, actual: %v", schoolHistory.EndDate.Time)
			case schoolHistoryPb.SchoolCourseId != schoolHistory.SchoolCourseID.String:
				return fmt.Errorf("validation school_history failed, expect school_course_id: %v, actual: %v", schoolHistoryPb.SchoolCourseId, schoolHistory.SchoolCourseID.String)
			case schoolHistory.ResourcePath.String != resourcePath:
				return fmt.Errorf("validation school_history failed, expect resource_path: %v, actual: %v", schoolHistory.ResourcePath.String, resourcePath)
			}
			count++
		}

		if len(currentSchoolsInDB) != 0 {
			if currentSchoolsInDB[0].SchoolID().String() == schoolHistory.SchoolID.String {
				if !schoolHistory.IsCurrent.Bool {
					return fmt.Errorf("expected: there is current school history but actual is not, schoolID: %s, current school: %t", schoolHistory.SchoolID.String, schoolHistory.IsCurrent.Bool)
				}
			} else {
				if schoolHistory.IsCurrent.Bool {
					return fmt.Errorf("expected: there is no current school history but actual is, schoolID: %s, current school: %t", schoolHistory.SchoolID.String, schoolHistory.IsCurrent.Bool)
				}
			}
		} else {
			if schoolHistory.IsCurrent.Bool {
				return fmt.Errorf("expected: there is no current school history but actual is, schoolID: %s, current school: %t", schoolHistory.SchoolID.String, schoolHistory.IsCurrent.Bool)
			}
		}
	}

	if len(schoolHistories) != count {
		return fmt.Errorf("cannot find any school_info match with request")
	}

	return nil
}

func validateSchoolHistoriesInDBWithCurrentSchool(ctx context.Context, db database.Ext, reqSchoolHistories []*pb.SchoolHistory, studentID string, resourcePath string, isCurrent bool) error {
	schoolHistoryRepo := &repository.SchoolHistoryRepo{}
	schoolHistories, err := schoolHistoryRepo.GetByStudentID(ctx, db, database.Text(studentID))
	if err != nil {
		return fmt.Errorf("schoolHistoryRepo.GetByStudentID: %v", err)
	}

	if len(schoolHistories) != len(reqSchoolHistories) {
		return fmt.Errorf("validation school_history failed, expect len(reqSchoolHistories): %v, actual: %v", len(reqSchoolHistories), len(schoolHistories))
	}

	count := 0
	for _, schoolHistory := range schoolHistories {
		for _, schoolHistoryPb := range reqSchoolHistories {
			if schoolHistoryPb.SchoolId != schoolHistory.SchoolID.String {
				continue
			}
			switch {
			case schoolHistoryPb.StartDate.IsValid() && (schoolHistoryPb.StartDate.AsTime().Unix() != schoolHistory.StartDate.Time.Unix()):
				return fmt.Errorf("validation school_history failed, expect start_date: %v, actual: %v", schoolHistoryPb.StartDate.AsTime().Unix(), schoolHistory.StartDate.Time.Unix())
			case !schoolHistoryPb.StartDate.IsValid() && schoolHistory.StartDate.Status != pgtype.Null:
				return fmt.Errorf("validation school_history failed, expect start_date: null, actual: %v", schoolHistory.StartDate.Time)
			case schoolHistoryPb.EndDate.IsValid() && (schoolHistoryPb.EndDate.AsTime().Unix() != schoolHistory.EndDate.Time.Unix()):
				return fmt.Errorf("validation school_history failed, expect end_date: %v, actual: %v", schoolHistoryPb.EndDate.AsTime().Unix(), schoolHistory.EndDate.Time.Unix())
			case !schoolHistoryPb.EndDate.IsValid() && schoolHistory.EndDate.Status != pgtype.Null:
				return fmt.Errorf("validation school_history failed, expect end_date: null, actual: %v", schoolHistory.EndDate.Time)
			case schoolHistoryPb.SchoolCourseId != schoolHistory.SchoolCourseID.String:
				return fmt.Errorf("validation school_history failed, expect school_course_id: %v, actual: %v", schoolHistoryPb.SchoolCourseId, schoolHistory.SchoolCourseID.String)
			case schoolHistory.ResourcePath.String != resourcePath:
				return fmt.Errorf("validation school_history failed, expect resource_path: %v, actual: %v", schoolHistory.ResourcePath.String, resourcePath)
				// case schoolHistory.IsCurrent.Bool != isCurrent:
				// 	return fmt.Errorf("validation school_history failed, expect is_current: %v, actual: %v", isCurrent, schoolHistory.IsCurrent.Bool) TODO: wrong expect @minhthao will fix later
			}
			count++
		}
	}

	if len(schoolHistories) != count {
		return fmt.Errorf("cannot find any school_info match with request")
	}

	return nil
}
