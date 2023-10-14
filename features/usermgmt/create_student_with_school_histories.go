package usermgmt

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func generateSchoolHistoryPbWithInvalidCondition(ctx context.Context, db database.Ext, condition string) ([]*pb.SchoolHistory, error) {
	levelID := idutil.ULIDNow()
	schoolInfo, err := insertRandomSchoolInfo(ctx, db, levelID)
	if err != nil {
		return nil, fmt.Errorf("insertRandomSchoolInfo: %v", err)
	}

	schoolCourse, err := insertRandomSchoolCourse(ctx, db, schoolInfo.ID.String)
	if err != nil {
		return nil, fmt.Errorf("insertRandomSchoolInfo: %v", err)
	}

	var schoolHistories []*pb.SchoolHistory
	switch condition {
	case "missing mandatory":
		schoolHistories = []*pb.SchoolHistory{
			{
				SchoolCourseId: schoolCourse.ID.String,
				StartDate:      timestamppb.Now(),
				EndDate:        timestamppb.New(time.Now().Add(-time.Hour)),
			},
		}
	case "invalid start_date and end_date":
		schoolHistories = []*pb.SchoolHistory{
			{
				SchoolId:  schoolInfo.ID.String,
				StartDate: timestamppb.Now(),
				EndDate:   timestamppb.New(time.Now().Add(-time.Hour)),
			},
		}
	case "duplicate school_info":
		schoolHistories = []*pb.SchoolHistory{
			{
				SchoolId: schoolInfo.ID.String,
			},
			{
				SchoolId: schoolInfo.ID.String,
			},
		}
	case "duplicate school_course":
		anotherSchoolInfo, err := insertRandomSchoolInfo(ctx, db, idutil.ULIDNow())
		if err != nil {
			return nil, fmt.Errorf("insertRandomSchoolInfo: %v", err)
		}
		anotherSchoolCourse, err := insertRandomSchoolCourse(ctx, db, anotherSchoolInfo.ID.String)
		if err != nil {
			return nil, fmt.Errorf("insertRandomSchoolCourse: %v", err)
		}
		schoolHistories = []*pb.SchoolHistory{
			{
				SchoolId:       schoolInfo.ID.String,
				SchoolCourseId: schoolInfo.ID.String,
			},
			{
				SchoolId:       anotherSchoolInfo.ID.String,
				SchoolCourseId: anotherSchoolCourse.ID.String,
			},
		}
	case "duplicate school_level":
		schoolInfoWithSameLevelID, err := insertRandomSchoolInfo(ctx, db, levelID)
		if err != nil {
			return nil, fmt.Errorf("insertRandomSchoolInfo: %v", err)
		}
		schoolHistories = []*pb.SchoolHistory{
			{
				SchoolId: schoolInfo.ID.String,
			},
			{
				SchoolId: schoolInfoWithSameLevelID.ID.String,
			},
		}
	}
	return schoolHistories, nil
}

func insertRandomSchoolInfo(ctx context.Context, db database.Ext, levelID string) (*entity.SchoolInfo, error) {
	sequence := rand.Intn(999999999)
	isArchived := false
	stmt := `INSERT INTO public.school_level (school_level_id, school_level_name, sequence, is_archived) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`
	_, err := db.Exec(ctx, stmt, levelID, levelID, sequence, isArchived)
	if err != nil {
		return nil, fmt.Errorf("db.Exec: %v", err)
	}

	schoolInfoRepo := &repository.SchoolInfoRepo{}

	schoolInfo := &entity.SchoolInfo{}
	database.AllNullEntity(schoolInfo)
	err = multierr.Combine(
		schoolInfo.ID.Set(idutil.ULIDNow()),
		schoolInfo.Name.Set("random-school_info"),
		schoolInfo.PartnerID.Set(idutil.ULIDNow()),
		schoolInfo.LevelID.Set(levelID),
		schoolInfo.IsArchived.Set(isArchived),
	)
	if err != nil {
		return nil, fmt.Errorf("multierr.Combine: %v", err)
	}
	err = schoolInfoRepo.Create(ctx, db, schoolInfo)
	if err != nil {
		return nil, fmt.Errorf("schoolInfoRepo.Create: %v", err)
	}

	return schoolInfo, nil
}

func insertRandomSchoolCourse(ctx context.Context, db database.Ext, schoolInfoID string) (*entity.SchoolCourse, error) {
	schoolCourseRepo := &repository.SchoolCourseRepo{}

	schoolCourse := &entity.SchoolCourse{}
	database.AllNullEntity(schoolCourse)
	err := multierr.Combine(
		schoolCourse.ID.Set(idutil.ULIDNow()),
		schoolCourse.Name.Set("random-school_info"),
		schoolCourse.PartnerID.Set(idutil.ULIDNow()),
		schoolCourse.SchoolID.Set(schoolInfoID),
		schoolCourse.IsArchived.Set(false),
	)
	if err != nil {
		return nil, fmt.Errorf("multierr.Combine: %v", err)
	}
	err = schoolCourseRepo.Create(ctx, db, schoolCourse)
	if err != nil {
		return nil, fmt.Errorf("schoolCourseRepo.Create: %v", err)
	}

	return schoolCourse, nil
}

func insertRandomSchoolLevelGrade(ctx context.Context, db database.Ext, gradeID string) (context.Context, *entity.SchoolInfo, error) {
	stepState := StepStateFromContext(ctx)
	levelID := idutil.ULIDNow()

	schoolInfo, err := insertRandomSchoolInfo(ctx, db, levelID)
	if err != nil {
		return nil, nil, fmt.Errorf("insertRandomSchoolInfo: %v", err)
	}

	stmtSchoolLevelGrade := `INSERT INTO school_level_grade (school_level_id, grade_id, created_at, updated_at) VALUES ($1, $2, now(), now())`
	cmd, err := db.Exec(ctx, stmtSchoolLevelGrade, levelID, gradeID)
	if err != nil {
		return ctx, nil, fmt.Errorf("db.Exec err: %v", err)
	}
	if cmd.RowsAffected() == 0 {
		return ctx, nil, fmt.Errorf("db.Exec err: no row effect")
	}

	return StepStateToContext(ctx, stepState), schoolInfo, nil
}

func (s *suite) studentInfoWithSchoolHistoriesRequest(ctx context.Context, condition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, OrgIDFromCtx(ctx), StaffRoleSchoolAdmin)
	randomID := newID()

	domainGradeRepo := &repository.DomainGradeRepo{}
	grades, err := domainGradeRepo.GetByPartnerInternalIDs(ctx, s.BobDBTrace, stepState.PartnerInternalIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("domainGradeRepo.GetByPartnerInternalIDs err: %v", err)
	}
	gradeID := grades[0].GradeID().String()

	req := &pb.CreateStudentRequest{
		SchoolId: int32(OrgIDFromCtx(ctx)),
		StudentProfile: &pb.CreateStudentRequest_StudentProfile{
			Email:             fmt.Sprintf("%v@example.com", randomID),
			Password:          fmt.Sprintf("password-%v", randomID),
			Name:              fmt.Sprintf("user-%v", randomID),
			CountryCode:       cpb.Country_COUNTRY_VN,
			EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			PhoneNumber:       fmt.Sprintf("phone-number-%v", randomID),
			StudentExternalId: fmt.Sprintf("student-external-id-%v", randomID),
			StudentNote:       fmt.Sprintf("some random student note %v", randomID),
			Grade:             5,
			Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
			Gender:            pb.Gender_MALE,
			LocationIds:       []string{s.ExistingLocations[0].LocationID.String},
			GradeId:           gradeID,
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

func (s *suite) studentInfoWithSchoolHistoriesInvalidRequest(ctx context.Context, condition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	randomID := newID()

	orgID := OrgIDFromCtx(ctx)
	req := &pb.CreateStudentRequest{
		SchoolId: int32(orgID),
		StudentProfile: &pb.CreateStudentRequest_StudentProfile{
			Email:             fmt.Sprintf("%v@example.com", randomID),
			Password:          fmt.Sprintf("password-%v", randomID),
			Name:              fmt.Sprintf("user-%v", randomID),
			CountryCode:       cpb.Country_COUNTRY_VN,
			EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			PhoneNumber:       fmt.Sprintf("phone-number-%v", randomID),
			StudentExternalId: fmt.Sprintf("student-external-id-%v", randomID),
			StudentNote:       fmt.Sprintf("some random student note %v", randomID),
			Grade:             5,
			Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
			Gender:            pb.Gender_MALE,
			LocationIds:       []string{s.ExistingLocations[0].LocationID.String},
		},
	}
	ctx = s.signedIn(ctx, orgID, StaffRoleSchoolAdmin)
	schoolHistories, err := generateSchoolHistoryPbWithInvalidCondition(ctx, s.BobDBTrace, condition)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("generateSchoolHistoryPbWithInvalidCondition: %v", err)
	}
	req.SchoolHistories = schoolHistories
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) newStudentAccountCreatedSuccessWithSchoolHistories(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	req := stepState.Request.(*pb.CreateStudentRequest)
	resp := stepState.Response.(*pb.CreateStudentResponse)

	if err := validateSchoolHistoriesInDB(ctx, s.BobDBTrace, req.SchoolHistories, resp.StudentProfile.Student.UserProfile.UserId, fmt.Sprint(OrgIDFromCtx(ctx)), entity.DomainSchools{}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("validateSchoolHistoriesInDB: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) newStudentAccountCreatedSuccessWithSchoolHistoriesHaveCurrentSchool(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	req := stepState.Request.(*pb.CreateStudentRequest)
	resp := stepState.Response.(*pb.CreateStudentResponse)

	if err := validateSchoolHistoriesInDBWithCurrentSchool(ctx, s.BobDBTrace, req.SchoolHistories, resp.StudentProfile.Student.UserProfile.UserId, fmt.Sprint(OrgIDFromCtx(ctx)), true); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("validateSchoolHistoriesInDB: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}
