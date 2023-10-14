package eureka

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/bob/constants"
	consta "github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) generateLearningObjective1(ctx context.Context) *cpb.LearningObjective {
	stepState := StepStateFromContext(ctx)
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
		TopicId: stepState.TopicID,
		Prerequisites: []string{
			"AL-PH3.1", "AL-PH3.2",
		},
		StudyGuide:     "https://guides/1/master",
		Video:          "https://videos/1/master",
		Instruction:    "instruction-2",
		GradeToPass:    wrapperspb.Int32(1),
		ManualGrading:  false,
		TimeLimit:      wrapperspb.Int32(1),
		MaximumAttempt: wrapperspb.Int32(1),
		ApproveGrading: true,
		GradeCapping:   false,
		ReviewOption:   cpb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_IMMEDIATELY,
		VendorType:     cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_LEARNOSITY,
	}
}

func (s *suite) userCreateNewLosAndAssignments(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	_, assignment1 := s.generateAssignment(ctx, "", false, false, true)
	_, assignment2 := s.generateAssignment(ctx, "", false, false, true)
	_, assignment3 := s.generateAssignment(ctx, "", false, false, true)

	assignments := []*pb.Assignment{
		assignment1,
		assignment2,
		assignment3,
	}
	los := []*cpb.LearningObjective{
		s.generateLearningObjective1(ctx),
		s.generateLearningObjective1(ctx),
	}
	for i := 0; i < len(los); i++ {
		los[i].TopicId = stepState.TopicID
	}

	// if _, err := s.aValidUser(ctx, stepState.SchoolAdminID, consta.RoleSchoolAdmin); err != nil {
	// 	return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create student: %w", err)
	// }

	token, err := s.generateExchangeToken(stepState.SchoolAdminID, consta.RoleSchoolAdmin)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error creating user: %w", err)
	}

	// ctx, _, token, err = s.signedInAs(ctx, consta.RoleSchoolAdmin)
	// if err != nil {
	// 	return StepStateToContext(ctx, stepState), err
	// }
	stepState.AuthToken = token

	req := &pb.UpsertLOsAndAssignmentsRequest{
		Assignments:        assignments,
		LearningObjectives: los,
	}
	stepState.Response, stepState.ResponseErr = pb.NewCourseModifierServiceClient(s.Conn).UpsertLOsAndAssignments(contextWithToken(s, ctx), req)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("UpsertLOsAndAssignments: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aRandomNumber(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Random = strconv.Itoa(rand.Int())
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aSchoolName(ctx context.Context, schoolName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch schoolName {
	case "Manabie":
		stepState.SchoolIDInt = constants.ManabieSchool
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("unsupported school name %v", schoolName)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) newLOsAndAssignmentsMustBeCreated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.UpsertLOsAndAssignmentsResponse)
	assignmentIDs := resp.AssignmentIds
	loIDs := resp.LoIds

	mainProcess := func() error {
		queryAssignments := `SELECT count(assignment_id)
			FROM assignments
		    WHERE assignment_id = ANY($1)`

		rows, err := s.DB.Query(ctx, queryAssignments, assignmentIDs)
		if err != nil {
			return err
		}
		defer rows.Close()

		var countAssignments int
		for rows.Next() {
			err = rows.Scan(&countAssignments)
			if err != nil {
				return err
			}
		}

		if countAssignments != len(assignmentIDs) {
			return errors.New(fmt.Sprintf("expected: %v assignments but got: %v assignments", countAssignments, len(assignmentIDs)))
		}

		queryResourcePathAssignments := `SELECT distinct(resource_path), count(assignment_id)
			FROM assignments
			WHERE assignment_id = ANY($1)
			GROUP BY resource_path`

		rows, err = s.DB.Query(ctx, queryResourcePathAssignments, assignmentIDs)
		if err != nil {
			return err
		}
		defer rows.Close()

		var resourcePath string
		for rows.Next() {
			err = rows.Scan(&resourcePath, &countAssignments)
			if err != nil {
				return err
			}
		}

		if resourcePath != fmt.Sprintf("%d", stepState.SchoolIDInt) {
			return errors.New(fmt.Sprintf("expected: resource_path(%s) but got: %s", fmt.Sprintf("%d", stepState.SchoolIDInt), resourcePath))
		}

		if countAssignments != len(assignmentIDs) {
			return errors.New(fmt.Sprintf("expected: %v assignments but got: %v assignments", countAssignments, len(assignmentIDs)))
		}

		queryLOs := `SELECT count(lo_id)
			FROM learning_objectives
		    WHERE lo_id = ANY($1)`

		rows, err = s.DB.Query(ctx, queryLOs, loIDs)
		if err != nil {
			return err
		}
		defer rows.Close()
		if err != nil {
			return err
		}

		var countLOs int
		for rows.Next() {
			err = rows.Scan(&countLOs)
			if err != nil {
				return err
			}
		}

		if countLOs != len(loIDs) {
			return errors.New(fmt.Sprintf("expected: %v los but got: %v los", countLOs, len(loIDs)))
		}

		queryResourcePathLOs := `SELECT distinct(resource_path), count(lo_id)
			FROM learning_objectives
			WHERE lo_id = ANY($1)
			GROUP BY resource_path`

		rows, err = s.DB.Query(ctx, queryResourcePathLOs, loIDs)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			err = rows.Scan(&resourcePath, &countLOs)
			if err != nil {
				return err
			}
		}

		if resourcePath != fmt.Sprintf("%d", stepState.SchoolIDInt) {
			return errors.New(fmt.Sprintf("expected: resource_path(%s) but got: %s", fmt.Sprintf("%d", stepState.SchoolIDInt), resourcePath))
		}

		if countLOs != len(loIDs) {
			return errors.New(fmt.Sprintf("expected: %v los but got: %v los", countLOs, len(loIDs)))
		}

		return nil
	}

	return s.ExecuteWithRetry(ctx, mainProcess, 2*time.Second, 10)
}

func (s *suite) generateLOsReq(ctx context.Context) *pb.UpsertLOsRequest {
	n := rand.Intn(5) + 3
	los := make([]*cpb.LearningObjective, 0, n)
	for i := 0; i < n; i++ {
		lo := s.generateLearningObjective1(ctx)
		lo.Info.Id = idutil.ULIDNow()
		los = append(los, lo)
	}

	return &pb.UpsertLOsRequest{
		LearningObjectives: los,
	}
}

const (
	LOType = "learning objective"
)

type LOAssignmentDisplayOrder struct {
	Type         string
	ID           string
	DisplayOrder int32
}
