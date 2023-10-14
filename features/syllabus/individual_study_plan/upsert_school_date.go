package individual_study_plan

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/bob/constants"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	ypb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) hasCreatedAStudyPlanExactMatchWithTheBookContentForMultipleStudent(ctx context.Context, user string, numStudent int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.Token = stepState.SchoolAdmin.Token

	stepState.CourseID = idutil.ULIDNow()
	if _, err := ypb.NewCourseServiceClient(s.YasuoConn).UpsertCourses(s.AuthHelper.SignedCtx(ctx, stepState.Token), &ypb.UpsertCoursesRequest{
		Courses: []*ypb.UpsertCoursesRequest_Course{
			{
				Id:           stepState.CourseID,
				Name:         fmt.Sprintf("course-name+%s", stepState.CourseID),
				Country:      bob_pb.COUNTRY_VN,
				Subject:      bob_pb.SUBJECT_MATHS,
				Grade:        i18n.OutGradeMap[bob_pb.COUNTRY_VN][int(1)],
				SchoolId:     constants.ManabieSchool,
				DisplayOrder: 1,
			},
		},
	}); err != nil {
		if e, ok := status.FromError(err); ok && e.Code() == codes.PermissionDenied {
			stepState.ResponseErr = err
			return utils.StepStateToContext(ctx, stepState), nil
		}
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create course: %w", err)
	}

	if resp, err := epb.NewCourseModifierServiceClient(s.EurekaConn).AddBooks(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.AddBooksRequest{
		CourseId: stepState.CourseID,
		BookIds:  []string{stepState.BookID},
	}); err != nil || !resp.Successful {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert course book: %w", err)
	}
	req := &epb.UpsertStudyPlanRequest{
		Name:                fmt.Sprintf("studyplan-%s", stepState.StudyPlanID),
		SchoolId:            constants.ManabieSchool,
		TrackSchoolProgress: true,
		Grades:              []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Status:              epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
		BookId:              stepState.BookID,
		CourseId:            stepState.CourseID,
	}
	resp, err := epb.NewStudyPlanModifierServiceClient(s.EurekaConn).UpsertStudyPlan(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert study plan: %w", err)
	}

	studyPlanItems, err := (&repositories.StudyPlanItemRepo{}).FindByStudyPlanID(ctx, s.EurekaDB, database.Text(resp.StudyPlanId))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve study plan items: %w", err)
	}

	stepState.LoIDs = nil

	upsertSpiReq := &epb.UpsertStudyPlanItemV2Request{}
	for _, item := range studyPlanItems {
		cse := &entities.ContentStructure{}
		err := item.ContentStructure.AssignTo(cse)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unmarshal ContentStructure: %w", err)
		}

		cs := &epb.ContentStructure{}
		_ = item.ContentStructure.AssignTo(cs)

		if len(cse.LoID) != 0 {
			stepState.LoIDs = append(stepState.LoIDs, cse.LoID)
			cs.ItemId = &epb.ContentStructure_LoId{LoId: wrapperspb.String(cse.LoID)}
		} else if len(cse.AssignmentID) != 0 {
			cs.ItemId = &epb.ContentStructure_AssignmentId{AssignmentId: wrapperspb.String(cse.AssignmentID)}
		}

		upsertSpiReq.StudyPlanItems = append(upsertSpiReq.StudyPlanItems, &epb.StudyPlanItem{
			StudyPlanId:             item.StudyPlanID.String,
			StudyPlanItemId:         item.ID.String,
			AvailableFrom:           timestamppb.New(time.Now().Add(-24 * time.Hour)),
			AvailableTo:             timestamppb.New(time.Now().AddDate(0, 0, 10)),
			StartDate:               timestamppb.New(time.Now().Add(-23 * time.Hour)),
			EndDate:                 timestamppb.New(time.Now().AddDate(0, 0, 1)),
			ContentStructure:        cs,
			ContentStructureFlatten: item.ContentStructureFlatten.String,
			Status:                  epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
		})
	}

	_, err = epb.NewStudyPlanModifierServiceClient(s.EurekaConn).UpsertStudyPlanItemV2(s.AuthHelper.SignedCtx(ctx, stepState.Token), upsertSpiReq)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert study plan item: %w", err)
	}

	stepState.StudyPlanID = resp.StudyPlanId
	stepState.StudyPlanIDs = append(stepState.StudyPlanIDs, resp.StudyPlanId)

	for i := 0; i < numStudent; i++ {
		ctx, _ := s.aSignedIn(ctx, "student")
		stepState.StudentIDs = append(stepState.StudentIDs, stepState.Student.ID)
		stmt := `SELECT email FROM users WHERE user_id = $1`
		var studentEmail string
		err = s.BobDB.QueryRow(ctx, stmt, stepState.Student.ID).Scan(&studentEmail)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}

		stepState.Token = stepState.SchoolAdmin.Token
		ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)
		locationID := idutil.ULIDNow()
		e := &bob_entities.Location{}
		database.AllNullEntity(e)
		if err := multierr.Combine(
			e.LocationID.Set(locationID),
			e.Name.Set(fmt.Sprintf("location-%s", locationID)),
			e.IsArchived.Set(false),
			e.CreatedAt.Set(time.Now()),
			e.UpdatedAt.Set(time.Now()),
		); err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}

		if _, err := database.Insert(ctx, e, s.BobDB.Exec); err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}

		_, err = upb.NewUserModifierServiceClient(s.UserMgmtConn).UpdateStudent(
			ctx,
			&upb.UpdateStudentRequest{
				StudentProfile: &upb.UpdateStudentRequest_StudentProfile{
					Id:               stepState.Student.ID,
					Name:             "test-name",
					Grade:            5,
					EnrollmentStatus: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Email:            studentEmail,
					LocationIds:      []string{locationID},
				},

				SchoolId: constants.ManabieSchool,
			},
		)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to update student: %w", err)
		}

		if _, err := upb.NewUserModifierServiceClient(s.UserMgmtConn).UpsertStudentCoursePackage(ctx, &upb.UpsertStudentCoursePackageRequest{
			StudentId: stepState.Student.ID,
			StudentPackageProfiles: []*upb.UpsertStudentCoursePackageRequest_StudentPackageProfile{{
				Id: &upb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
					CourseId: stepState.CourseID,
				},
				StartTime: timestamppb.New(time.Now().Add(time.Hour * -20)),
				EndTime:   timestamppb.New(time.Now().Add(time.Hour * 24 * 10)),
			}},
		}); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to update student course package: %w", err)
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemTriggersDataToIndividualStudyPlanTableCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	query := `
	select isp.school_date from individual_study_plan isp
	where isp.study_plan_id = $1::TEXT
	and isp.learning_material_id = $2::TEXT
	and isp.student_id  = ANY($3::TEXT[])
	`

	rows, err := s.EurekaDB.Query(ctx, query, database.Text(stepState.StudyPlanID), database.Text(stepState.LoIDs[0]), database.TextArray(stepState.StudentIDs))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can not query: %w", err)
	}
	defer rows.Close()

	resp := make([]*entities.IndividualStudyPlan, 0)
	for rows.Next() {
		individualStudyPlan := new(entities.IndividualStudyPlan)
		err := rows.Scan(&individualStudyPlan.SchoolDate)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can not scan: %w", err)
		}

		resp = append(resp, individualStudyPlan)
	}

	if rows.Err() != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("rows Err: %v", rows.Err().Error())
	}

	for _, i := range resp {
		if i.SchoolDate.Time.GoString() == stepState.SchoolDate.AsTime().GoString() {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("trigger data not correct")
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userUpdateSchoolDateV2(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.SchoolDate = timestamppb.Now()
	var studyPlanItemIdentities []*sspb.StudyPlanItemIdentity
	for _, studentID := range stepState.StudentIDs {
		studyPlanItemIdentities = append(studyPlanItemIdentities, &sspb.StudyPlanItemIdentity{
			StudyPlanId:        stepState.StudyPlanID,
			LearningMaterialId: stepState.LoIDs[0],
			StudentId:          wrapperspb.String(studentID),
		})
	}
	stepState.Response, stepState.ResponseErr = sspb.NewStudyPlanClient(s.EurekaConn).UpsertSchoolDate(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.UpsertSchoolDateRequest{
		SchoolDate:              stepState.SchoolDate,
		StudyPlanItemIdentities: studyPlanItemIdentities,
	})
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemStoresSchoolDateCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	ctx, spis, err := s.getStudyPlanItemByStudyPlanItemIdentity(ctx,
		database.Text(stepState.StudyPlanID),
		database.Text(stepState.LoIDs[0]),
		database.TextArray(stepState.StudentIDs))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("not found study plan items")
	}

	for _, i := range spis {
		if i.SchoolDate.Time.GoString() == stepState.SchoolDate.AsTime().GoString() {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("update school date not correct")
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) getStudyPlanItemByStudyPlanItemIdentity(ctx context.Context, masterStudyPlanID, lmID pgtype.Text, studentIDs pgtype.TextArray) (context.Context, []*entities.StudyPlanItem, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	query := `
	select spi.study_plan_item_id,spi.school_date from study_plan_items spi join student_study_plans ssp 
	on spi.study_plan_id = ssp.study_plan_id 
	where ssp.master_study_plan_id = $1::TEXT
	and (spi.content_structure->>'lo_id' = $2::TEXT or spi.content_structure ->> 'assignment_id' = $2::TEXT)
	and ssp.student_id  = ANY($3::TEXT[])
	`

	rows, err := s.EurekaDB.Query(ctx, query, &masterStudyPlanID, &lmID, &studentIDs)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), nil, fmt.Errorf("can not query: %w", err)
	}
	defer rows.Close()

	resp := make([]*entities.StudyPlanItem, 0)
	for rows.Next() {
		studyPlanItem := new(entities.StudyPlanItem)
		err := rows.Scan(&studyPlanItem.ID, &studyPlanItem.SchoolDate)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), nil, fmt.Errorf("can not scant: %w", err)
		}

		resp = append(resp, studyPlanItem)
	}

	if rows.Err() != nil {
		return utils.StepStateToContext(ctx, stepState), nil, fmt.Errorf("rows Err: %v", rows.Err().Error())
	}

	return utils.StepStateToContext(ctx, stepState), resp, nil
}
