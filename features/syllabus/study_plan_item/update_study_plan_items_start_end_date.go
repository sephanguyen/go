package study_plan_item

import (
	"context"
	"fmt"
	"strings"
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
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/sirupsen/logrus"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// nolint
func (s *Suite) adminUpdateStudyPlanItemsStartEndDateWith(ctx context.Context, updateType string) (context.Context, error) {
	time.Sleep(2 * time.Second)
	stepState := utils.StepStateFromContext[StepState](ctx)
	var studyPlanItemIdentities []*sspb.StudyPlanItemIdentity
	for _, studentID := range stepState.StudentIDs {
		studyPlanItemIdentities = append(studyPlanItemIdentities, &sspb.StudyPlanItemIdentity{
			StudyPlanId:        stepState.StudyPlanID,
			LearningMaterialId: stepState.LoIDs[0],
			StudentId:          wrapperspb.String(studentID),
		})
	}
	req := &sspb.UpdateStudyPlanItemsStartEndDateRequest{
		StudyPlanItemIdentities: studyPlanItemIdentities,
	}
	date := time.Now().AddDate(0, 0, 1)
	switch updateType {
	case "start":
		req.Fields = sspb.UpdateStudyPlanItemsStartEndDateFields_START_DATE
		stepState.StudyPlanItemStartDate = date
		req.StartDate = timestamppb.New(stepState.StudyPlanItemStartDate)
	case "end":
		req.Fields = sspb.UpdateStudyPlanItemsStartEndDateFields_END_DATE
		stepState.StudyPlanItemEndDate = date.AddDate(0, 0, 1)
		req.EndDate = timestamppb.New(stepState.StudyPlanItemEndDate)
	case "start_end":
		if updateType == "start_end" {
			req.Fields = sspb.UpdateStudyPlanItemsStartEndDateFields_ALL
		}
		stepState.StudyPlanItemStartDate = date.AddDate(0, 0, 1)
		stepState.StudyPlanItemEndDate = date.AddDate(0, 0, 3)
		req.StartDate = timestamppb.New(stepState.StudyPlanItemStartDate)
		req.EndDate = timestamppb.New(stepState.StudyPlanItemEndDate)
	}

	logrus.Debug("REQ", req)
	_, err := sspb.NewStudyPlanClient(s.EurekaConn).UpdateStudyPlanItemsStartEndDate(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable edit study plan items time: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

// nolint
func (s *Suite) adminUpdateStudyPlanItemsStartEndDateWithNullData(ctx context.Context, updateType string) (context.Context, error) {
	time.Sleep(2 * time.Second)
	stepState := utils.StepStateFromContext[StepState](ctx)

	var studyPlanItemIdentities []*sspb.StudyPlanItemIdentity
	for _, studentID := range stepState.StudentIDs {
		studyPlanItemIdentities = append(studyPlanItemIdentities, &sspb.StudyPlanItemIdentity{
			StudyPlanId:        stepState.StudyPlanID,
			LearningMaterialId: stepState.LoIDs[0],
			StudentId:          wrapperspb.String(studentID),
		})
	}
	req := &sspb.UpdateStudyPlanItemsStartEndDateRequest{
		StudyPlanItemIdentities: studyPlanItemIdentities,
	}
	switch updateType {
	case "start":
		req.Fields = sspb.UpdateStudyPlanItemsStartEndDateFields_START_DATE
		req.StartDate = nil
	case "end":
		req.Fields = sspb.UpdateStudyPlanItemsStartEndDateFields_END_DATE
		req.EndDate = nil
	case "start_end":
		req.Fields = sspb.UpdateStudyPlanItemsStartEndDateFields_ALL
		req.StartDate = nil
		req.EndDate = nil
	}

	_, err := sspb.NewStudyPlanClient(s.EurekaConn).UpdateStudyPlanItemsStartEndDate(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable edit study plan items time: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

// nolint
func (s *Suite) studyPlanItemsTimeWasUpdatedWithAccordingUpdateType(ctx context.Context, updateType string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	spi := &entities.StudyPlanItem{}
	fields, _ := spi.FieldMap()
	selectFields := make([]string, 0)
	for _, v := range fields {
		selectFields = append(selectFields, "spi."+v)
	}
	spis := &entities.StudyPlanItems{}
	query := fmt.Sprintf(`SELECT %s
FROM student_study_plans ssp
JOIN LATERAL (
    SELECT *
    FROM study_plan_items
    WHERE (study_plan_id = ssp.study_plan_id OR study_plan_id = ssp.master_study_plan_id)
      AND coalesce(content_structure ->> 'lo_id', content_structure ->> 'assignment_id') = $2
) spi ON TRUE
WHERE ssp.master_study_plan_id = $1
        AND (array_length($3::TEXT[], 1) = 0
            OR (ssp.student_id = ANY ($3::TEXT[])
                AND spi.study_plan_id != $1));`, strings.Join(selectFields, ","))
	if err := database.Select(ctx, s.EurekaDB, query, stepState.StudyPlanID, stepState.LoIDs[0], stepState.StudentIDs).ScanAll(spis); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	for _, item := range *spis {
		switch updateType {
		case "start":
			if item.AvailableFrom.Time.Before(stepState.StudyPlanItemStartDate) && !isEqual(item.StartDate.Time, stepState.StudyPlanItemStartDate) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("edit time fail with %s", updateType)
			}
		case "end":
			if item.AvailableTo.Time.After(stepState.StudyPlanItemEndDate) && !isEqual(item.EndDate.Time, stepState.StudyPlanItemEndDate) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("edit time fail with %s", updateType)
			}
		case "start_end":
			if item.AvailableFrom.Time.Before(stepState.StudyPlanItemStartDate) && item.AvailableTo.Time.After(stepState.StudyPlanItemEndDate) && !isEqual(item.StartDate.Time, stepState.StudyPlanItemStartDate) && !isEqual(item.EndDate.Time, stepState.StudyPlanItemEndDate) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("edit time fail with %s", updateType)
			}
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

// nolint
func (s *Suite) assignmentTimeWasUpdatedWithNullDataAndAccordingUpdate_type(ctx context.Context, updateType string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	spi := &entities.StudyPlanItem{}
	fields, _ := spi.FieldMap()
	selectFields := make([]string, 0)
	for _, v := range fields {
		selectFields = append(selectFields, "spi."+v)
	}
	spis := &entities.StudyPlanItems{}
	query := fmt.Sprintf(`SELECT %s
FROM student_study_plans ssp
JOIN LATERAL (
    SELECT *
    FROM study_plan_items
    WHERE (study_plan_id = ssp.study_plan_id OR study_plan_id = ssp.master_study_plan_id)
      AND coalesce(content_structure ->> 'lo_id', content_structure ->> 'assignment_id') = $2
) spi ON TRUE
WHERE ssp.master_study_plan_id = $1
        AND (array_length($3::TEXT[], 1) = 0
            OR (ssp.student_id = ANY ($3::TEXT[])
                AND spi.study_plan_id != $1));`, strings.Join(selectFields, ","))
	if err := database.Select(ctx, s.EurekaDB, query, stepState.StudyPlanID, stepState.LoIDs[0], stepState.StudentIDs).ScanAll(spis); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	for _, item := range *spis {
		switch updateType {
		case "start":
			if item.StartDate.Status != pgtype.Null {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("edit time fail with %s", updateType)
			}
		case "end":
			if item.EndDate.Status != pgtype.Null {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("edit time fail with %s", updateType)
			}
		case "start_end":
			if item.StartDate.Status != pgtype.Null && item.EndDate.Status == pgtype.Null {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("edit time fail with %s", updateType)
			}
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func isEqual(src, dest time.Time) bool {
	return src.Format(time.RFC3339) == dest.Format(time.RFC3339)
}

func (s *Suite) userSendUpdateStudyPlanItemsStartEndDateRequest(ctx context.Context) (context.Context, error) {
	date := time.Now().AddDate(0, 0, 1)

	stepState := utils.StepStateFromContext[StepState](ctx)

	var studyPlanItemIdentities []*sspb.StudyPlanItemIdentity
	for _, studentID := range stepState.StudentIDs {
		studyPlanItemIdentities = append(studyPlanItemIdentities, &sspb.StudyPlanItemIdentity{
			StudyPlanId:        stepState.StudyPlanID,
			LearningMaterialId: stepState.LoIDs[0],
			StudentId:          wrapperspb.String(studentID),
		})
	}
	req := &sspb.UpdateStudyPlanItemsStartEndDateRequest{
		StudyPlanItemIdentities: studyPlanItemIdentities,
		Fields:                  sspb.UpdateStudyPlanItemsStartEndDateFields_ALL,
		StartDate:               timestamppb.New(date),
		EndDate:                 timestamppb.New(date.AddDate(0, 0, 2)),
	}
	_, err := sspb.NewStudyPlanClient(s.EurekaConn).UpdateStudyPlanItemsStartEndDate(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable edit assignment time: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) generateLOsReq(ctx context.Context) *pb.UpsertLOsRequest {
	los := make([]*cpb.LearningObjective, 0, 3)
	for i := 0; i < 3; i++ {
		lo := s.generateLearningObjective1(ctx)
		lo.Info.Id = idutil.ULIDNow()
		los = append(los, lo)
	}

	return &pb.UpsertLOsRequest{
		LearningObjectives: los,
	}
}

func (s *Suite) generateLearningObjective1(ctx context.Context) *cpb.LearningObjective {
	stepState := utils.StepStateFromContext[StepState](ctx)
	id := idutil.ULIDNow()

	return &cpb.LearningObjective{
		Info: &cpb.ContentBasicInfo{
			Id:        id,
			Name:      "learning",
			Country:   cpb.Country_COUNTRY_VN,
			Grade:     12,
			Subject:   cpb.Subject_SUBJECT_MATHS,
			MasterId:  "",
			SchoolId:  constants.ManabieSchool,
			CreatedAt: nil,
			UpdatedAt: nil,
		},
		TopicId: stepState.TopicIDs[0],
		Prerequisites: []string{
			"AL-PH3.1", "AL-PH3.2",
		},
		StudyGuide: "https://guides/1/master",
		Video:      "https://videos/1/master",
	}
}

func (s *Suite) hasCreatedAContentBook(ctx context.Context, user string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)
	ctx, err := s.aValidBookContent(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	loReq := s.generateLOsReq(ctx)
	if _, err := pb.NewLearningObjectiveModifierServiceClient(s.EurekaConn).UpsertLOs(ctx, loReq); err != nil {
		if e, ok := status.FromError(err); ok && e.Code() == codes.PermissionDenied {
			stepState.ResponseErr = err
			return utils.StepStateToContext(ctx, stepState), nil
		}
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable create los: %w", err)
	}

	for _, lo := range loReq.LearningObjectives {
		stepState.LoIDs = append(stepState.LoIDs, lo.Info.Id)
	}

	assignmentID := idutil.ULIDNow()
	stepState.AssignmentID = assignmentID
	stepState.AssignmentIDs = append(stepState.AssignmentIDs, stepState.AssignmentID)
	if _, err := pb.NewAssignmentModifierServiceClient(s.EurekaConn).UpsertAssignments(ctx, &pb.UpsertAssignmentsRequest{
		Assignments: []*pb.Assignment{
			{
				AssignmentId: assignmentID,
				Name:         fmt.Sprintf("assignment-%s", assignmentID),
				Content: &pb.AssignmentContent{
					TopicId: stepState.TopicIDs[0],
					LoId:    stepState.LoIDs,
				},
				CheckList: &pb.CheckList{
					Items: []*pb.CheckListItem{
						{
							Content:   "Complete all learning objectives",
							IsChecked: true,
						},
						{
							Content:   "Submitted required videos",
							IsChecked: false,
						},
					},
				},
				Instruction:    "teacher's instruction",
				MaxGrade:       100,
				Attachments:    []string{"media-id-1", "media-id-2"},
				AssignmentType: pb.AssignmentType_ASSIGNMENT_TYPE_LEARNING_OBJECTIVE,
				Setting: &pb.AssignmentSetting{
					AllowLateSubmission: true,
					AllowResubmission:   true,
				},
				RequiredGrade: true,
				DisplayOrder:  0,
			},
		},
	}); err != nil {
		if e, ok := status.FromError(err); ok && e.Code() == codes.PermissionDenied {
			stepState.ResponseErr = err
			return utils.StepStateToContext(ctx, stepState), nil
		}
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable create a assignment %v", err)
	}
	stepState.CourseID = idutil.ULIDNow()
	if _, err := ypb.NewCourseServiceClient(s.YasuoConn).UpsertCourses(ctx, &ypb.UpsertCoursesRequest{
		Courses: []*ypb.UpsertCoursesRequest_Course{
			{
				Id:           stepState.CourseID,
				Name:         fmt.Sprintf("course-name+%s", stepState.CourseID),
				Country:      bob_pb.COUNTRY_VN,
				Subject:      bob_pb.SUBJECT_MATHS,
				Grade:        i18n.OutGradeMap[bob_pb.COUNTRY_VN][int(stepState.Grade)],
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

	if _, err := pb.NewCourseModifierServiceClient(s.EurekaConn).AddBooks(ctx, &pb.AddBooksRequest{
		BookIds:  []string{stepState.BookID},
		CourseId: stepState.CourseID,
	}); err != nil {
		if e, ok := status.FromError(err); ok && e.Code() == codes.PermissionDenied {
			stepState.ResponseErr = err
			return utils.StepStateToContext(ctx, stepState), nil
		}
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to add books: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

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

	if resp, err := pb.NewCourseModifierServiceClient(s.EurekaConn).AddBooks(s.AuthHelper.SignedCtx(ctx, stepState.Token), &pb.AddBooksRequest{
		CourseId: stepState.CourseID,
		BookIds:  []string{stepState.BookID},
	}); err != nil || !resp.Successful {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert course book: %w", err)
	}
	req := &pb.UpsertStudyPlanRequest{
		Name:                fmt.Sprintf("studyplan-%s", stepState.StudyPlanID),
		SchoolId:            constants.ManabieSchool,
		TrackSchoolProgress: true,
		Grades:              []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Status:              pb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
		BookId:              stepState.BookID,
		CourseId:            stepState.CourseID,
	}
	resp, err := pb.NewStudyPlanModifierServiceClient(s.EurekaConn).UpsertStudyPlan(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert study plan: %w", err)
	}

	studyPlanItems, err := (&repositories.StudyPlanItemRepo{}).FindByStudyPlanID(ctx, s.EurekaDB, database.Text(resp.StudyPlanId))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve study plan items: %w", err)
	}

	stepState.LoIDs = nil

	upsertSpiReq := &pb.UpsertStudyPlanItemV2Request{}
	for _, item := range studyPlanItems {
		cse := &entities.ContentStructure{}
		err := item.ContentStructure.AssignTo(cse)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unmarshal ContentStructure: %w", err)
		}

		cs := &pb.ContentStructure{}
		_ = item.ContentStructure.AssignTo(cs)

		if len(cse.LoID) != 0 {
			stepState.LoIDs = append(stepState.LoIDs, cse.LoID)
			cs.ItemId = &pb.ContentStructure_LoId{LoId: wrapperspb.String(cse.LoID)}
		} else if len(cse.AssignmentID) != 0 {
			cs.ItemId = &pb.ContentStructure_AssignmentId{AssignmentId: wrapperspb.String(cse.AssignmentID)}
		}

		upsertSpiReq.StudyPlanItems = append(upsertSpiReq.StudyPlanItems, &pb.StudyPlanItem{
			StudyPlanId:             item.StudyPlanID.String,
			StudyPlanItemId:         item.ID.String,
			AvailableFrom:           timestamppb.New(time.Now().Add(-24 * time.Hour)),
			AvailableTo:             timestamppb.New(time.Now().AddDate(0, 0, 10)),
			StartDate:               timestamppb.New(time.Now().Add(-23 * time.Hour)),
			EndDate:                 timestamppb.New(time.Now().AddDate(0, 0, 1)),
			ContentStructure:        cs,
			ContentStructureFlatten: item.ContentStructureFlatten.String,
			Status:                  pb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
		})
	}

	_, err = pb.NewStudyPlanModifierServiceClient(s.EurekaConn).UpsertStudyPlanItemV2(s.AuthHelper.SignedCtx(ctx, stepState.Token), upsertSpiReq)
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
