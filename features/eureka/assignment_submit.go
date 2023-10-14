package eureka

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	bob_repository "github.com/manabie-com/backend/internal/bob/repositories"
	consta "github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	bpb "github.com/manabie-com/backend/pkg/genproto/bob"
	common "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) ourSystemMustRecordsAllTheSubmissionsFromStudent(ctx context.Context, times string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	repo := repositories.StudentSubmissionRepo{}
	for _, v := range stepState.Submissions {
		e, err := repo.Get(ctx, s.DB, database.Text(v.Submission.SubmissionId))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if v.Submission.AssignmentId != e.AssignmentID.String {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecting %s assignment-id, got %s", v.Submission.AssignmentId, e.AssignmentID.String)
		}

		actualContents := []*pb.SubmissionContent{}
		if err := e.SubmissionContent.AssignTo(&actualContents); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		for k, c := range v.Submission.SubmissionContent {
			if !proto.Equal(c, actualContents[k]) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expecting %v , got %v", c, actualContents[k])
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) allRelatedStudyPlanItemsMarkAsCompleted(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studyPlanItems := make([]string, 0, len(stepState.Submissions))
	for _, s := range stepState.Submissions {
		studyPlanItems = append(studyPlanItems, s.Submission.StudyPlanItemId)
	}

	studyPlanItems = golibs.Uniq(studyPlanItems)
	var nNotCompleted pgtype.Int8
	err := database.Select(ctx, s.DB, `
		SELECT COUNT(*) 
		FROM study_plan_items 
		WHERE study_plan_item_id = ANY($1) AND completed_at IS NULL`,
		&studyPlanItems).ScanFields(&nNotCompleted)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if nNotCompleted.Int != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting all item completed, get some null (items: %s)", strings.Join(studyPlanItems, ","))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentSubmitTheirAssignment(ctx context.Context, contentStatus, times string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studentID := stepState.StudentIDs[0]
	stepState.CurrentStudentID = studentID
	token, err := s.generateExchangeToken(studentID, entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("generateValidAuthenticationToken: %w", err)
	}
	stepState.AuthToken = token
	ctx, assignments, err := s.getOneStudentAssignedAssignment(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.getOneStudentAssignedAssignment: %w", err)
	}

	stepState.StudentsSubmittedAssignments[studentID] = append(stepState.StudentsSubmittedAssignments[studentID], assignments...)
	ctx, err = s.submitAssignment(ctx, contentStatus, times, assignments)
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) studentSubmitTheirAssignmentTimesForDifferentAssignments(ctx context.Context, contentStatus, times string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	token, err := s.generateExchangeToken(stepState.StudentIDs[0], entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.generateExchangeToken: %w", err)
	}
	stepState.AuthToken = token
	ctx = contextWithToken(s, ctx)
	ctx, assignments, err := s.getSomeStudentAssignedAssignments(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.getSomeStudentAssignedAssignments: %w", err)
	}

	ctx, err = s.submitAssignment(ctx, contentStatus, times, assignments)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.submitAssignment: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) givenStudentSubmitTheirAssignmentInCurrentStudyPlan(ctx context.Context, studentID, contentStatus, times string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentStudentID = studentID
	token, err := s.generateExchangeToken(studentID, entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.generateExchangeToken: %w", err)
	}
	stepState.AuthToken = token
	ctx, assignments, err := s.getOneStudentAssignedAssignmentWithCurrentStudyPlan(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.getOneStudentAssignedWithCurrentStudyPlan: %w", err)
	}

	stepState.StudentsSubmittedAssignments[studentID] = append(stepState.StudentsSubmittedAssignments[studentID], assignments...)

	ctx, err = s.submitAssignment(ctx, contentStatus, times, assignments)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.submitAssignment: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentSubmitTheirAssignmentInCurrentStudyPlan(ctx context.Context, contentStatus, times string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.givenStudentSubmitTheirAssignmentInCurrentStudyPlan(ctx, stepState.StudentIDs[0], contentStatus, times)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.givenStudentSubmitTheirAssignmentInCurrentStudyPlan")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getSomeAssignmentSubmission(contentStatus, times string, assignments []*pb.Content) []*pb.SubmitAssignmentRequest {
	reqs := make([]*pb.SubmitAssignmentRequest, 0)

	for _, assignment := range assignments {
		n := 1
		if times != "single" {
			// random "multiple" times
			n = genRand(10)
		}

		for i := 0; i < n; i++ {
			submissionContent := []*epb.SubmissionContent{}
			if contentStatus == "existed" {
				submissionContent = []*epb.SubmissionContent{
					{
						SubmitMediaId:     fmt.Sprintf("%s media 1 %d", assignment.StudyPlanItem.StudyPlanItemId, i),
						AttachmentMediaId: fmt.Sprintf("%s attachment 1 %d", assignment.StudyPlanItem.StudyPlanItemId, i),
					},
					{
						SubmitMediaId:     fmt.Sprintf("%s media 2 %d", assignment.StudyPlanItem.StudyPlanItemId, i),
						AttachmentMediaId: fmt.Sprintf("%s attachment 2 %d", assignment.StudyPlanItem.StudyPlanItemId, i),
					},
				}
			}
			req := &pb.SubmitAssignmentRequest{
				Submission: &pb.StudentSubmission{
					StudyPlanItemId:    assignment.StudyPlanItem.StudyPlanItemId,
					AssignmentId:       assignment.ResourceId,
					SubmissionContent:  submissionContent,
					Note:               fmt.Sprintf("random note: %s", assignment.StudyPlanItem.StudyPlanItemId),
					CompleteDate:       timestamppb.New(time.Now().Add(time.Duration(genRand(24)) * time.Hour)),
					Duration:           int32(genRand(100)),
					CorrectScore:       rand.Float32() * 10,
					TotalScore:         rand.Float32() * 100,
					UnderstandingLevel: pb.SubmissionUnderstandingLevel(rand.Intn(len(pb.SubmissionUnderstandingLevel_value))),
				},
			}
			reqs = append(reqs, req)
		}
	}
	return reqs
}

func (s *suite) submitAssignment(ctx context.Context, contentStatus, times string, assignments []*pb.Content) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	reqs := s.getSomeAssignmentSubmission(contentStatus, times, assignments)

	for _, req := range reqs {
		resp, err := pb.NewStudentAssignmentWriteServiceClient(s.Conn).
			SubmitAssignment(contextWithToken(s, ctx), req)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to submit assignment: %w", err)
		}

		if resp.SubmissionId == "" {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecting SubmissionId")
		}
		req.Submission.SubmissionId = resp.SubmissionId
		stepState.Submissions = append(stepState.Submissions, req)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustUpdateNullContentForEachSubmission(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("listing return StepStateToContext(ctx, stepState), error: %w", stepState.ResponseErr)
	}

	ids := make([]string, 0, len(stepState.Submissions))
	for _, submission := range stepState.Submissions {
		ids = append(ids, submission.Submission.SubmissionId)
	}
	stmt := `SELECT COUNT(*) FROM student_submissions WHERE submission_content IS NULL AND student_submission_id=ANY($1::_TEXT)`
	var total int
	err := s.DB.QueryRow(ctx, stmt, database.TextArray(ids)).Scan(&total)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if total != len(ids) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total null submission content, expected %d, got %d", len(ids), total)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getOneStudentAssignedAssignment(ctx context.Context) (context.Context, []*pb.Content, error) {
	stepState := StepStateFromContext(ctx)
	ctx, contents, err := s.getSomeStudentAssignedAssignments(ctx)
	if len(contents) > 0 {
		stepState.StudyPlanID = contents[0].StudyPlanItem.StudyPlanId
	}
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, fmt.Errorf("s.getSomeStudentAssignedAssignments: %w", err)
	}
	return StepStateToContext(ctx, stepState), contents[:1], nil
}

func (s *suite) getSomeStudentAssignedAssignmentsWithCurrentStudyPlan(ctx context.Context) (context.Context, []*pb.Content, error) {
	stepState := StepStateFromContext(ctx)

	resp, err := pb.NewAssignmentReaderServiceClient(s.Conn).
		ListStudentAvailableContents(contextWithToken(s, ctx), &pb.ListStudentAvailableContentsRequest{})
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, fmt.Errorf("ListStudentAvailableContents: %w", err)
	}
	return StepStateToContext(ctx, stepState), resp.Contents[:len(resp.Contents)/2], nil
}

func (s *suite) getOneStudentAssignedAssignmentWithCurrentStudyPlan(ctx context.Context) (context.Context, []*pb.Content, error) {
	stepState := StepStateFromContext(ctx)
	ctx, contents, err := s.getSomeStudentAssignedAssignmentsWithCurrentStudyPlan(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, fmt.Errorf("unable to get some student assigned assignment: %w", err)
	}
	if contents == nil {
		return StepStateToContext(ctx, stepState), nil, fmt.Errorf("the content of student is nil")
	}
	if len(contents) > 0 {
		stepState.StudyPlanID = contents[0].StudyPlanItem.StudyPlanId
	}

	return StepStateToContext(ctx, stepState), contents[:1], nil
}

func (s *suite) getSomeStudentAssignedAssignments(ctx context.Context) (context.Context, []*pb.Content, error) {
	stepState := StepStateFromContext(ctx)

	token, err := s.generateExchangeToken(stepState.StudentIDs[0], entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, fmt.Errorf("s.generateExchangeToken: %w", err)
	}

	stepState.AuthToken = token
	resp, err := pb.NewAssignmentReaderServiceClient(s.Conn).
		ListStudentAvailableContents(contextWithToken(s, ctx), &pb.ListStudentAvailableContentsRequest{
			StudyPlanId: []string{},
		})
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, fmt.Errorf("ListStudentAvailableContents: %w", err)
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(resp.Contents), func(i, j int) {
		resp.Contents[i], resp.Contents[j] = resp.Contents[j], resp.Contents[i]
	})

	if len(resp.Contents) == 0 {
		return StepStateToContext(ctx, stepState), nil, fmt.Errorf("cannot find any sample assignments")
	}

	if len(resp.Contents) == 1 {
		return StepStateToContext(ctx, stepState), resp.Contents, nil
	}
	return StepStateToContext(ctx, stepState), resp.Contents[:len(resp.Contents)/2], nil
}

func (s *suite) someStudents(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// create random student ID and token
	stepState.StudentID = idutil.ULIDNow()
	if _, err := s.aValidUser(ctx, stepState.StudentID, consta.RoleStudent); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create student: %w", err)
	}
	token, err := s.generateExchangeToken(stepState.StudentID, entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.generateExchangeToken: %w", err)
	}
	stepState.AuthToken = token
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) unrelatedAssignments(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.someStudentsAreAssignedSomeValidStudyPlans(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.someStudentsAreAssignedSomeValidStudyPlans: %w", err)
	}
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) studentSubmitRandomAssignment(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	selectRandStmt := "SELECT assignment_id FROM assignments WHERE random() < 0.01 LIMIT 1"
	var assignmentID string
	var studyPlanItemID string
	for i := 0; i < 100; i++ {
		err := database.Select(ctx, s.DB, selectRandStmt).ScanFields(&assignmentID)
		if err != nil && errors.Is(err, pgx.ErrNoRows) {
			continue
		}

		selectStudyPlanItemStmt := "SELECT study_plan_item_id FROM assignment_study_plan_items WHERE assignment_id = $1"
		err = database.Select(ctx, s.DB, selectStudyPlanItemStmt, &assignmentID).ScanFields(&studyPlanItemID)
		if err != nil && errors.Is(err, pgx.ErrNoRows) {
			continue
		}

		break
	}

	req := &pb.SubmitAssignmentRequest{
		Submission: &pb.StudentSubmission{
			StudyPlanItemId: studyPlanItemID,
			AssignmentId:    assignmentID,
			SubmissionContent: []*epb.SubmissionContent{
				{
					SubmitMediaId:     "%d media 1 %d",
					AttachmentMediaId: "%d attachment 1 %d",
				},
				{
					SubmitMediaId:     "%d media 2 %d",
					AttachmentMediaId: "%d attachment 2 %d",
				},
			},
			StartDate: timestamppb.Now(),
			EndDate:   timestamppb.New(time.Now().Add(2 * time.Hour)),
		},
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewStudentAssignmentWriteServiceClient(s.Conn).
		SubmitAssignment(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustRejectThat(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.returnsStatusCode(ctx, codes.PermissionDenied.String())
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) listTheSubmissions(ctx context.Context, actor string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch actor {
	case "student":
		ctx, err := s.studentListTheSubmissions(ctx)
		return StepStateToContext(ctx, stepState), err
	case "teacher":
		ctx, err := s.teacherListTheSubmissions(ctx)
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentListTheSubmissions(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var count int
	query := "SELECT COUNT(*) FROM users where user_id = $1 AND deleted_at IS NULL"
	if err := s.DB.QueryRow(ctx, query, database.Text(stepState.StudentIDs[0])).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if count == 0 {
		if _, err := s.aValidUser(ctx, stepState.StudentIDs[0], consta.RoleStudent); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create student: %w", err)
		}
	}

	token, err := s.generateExchangeToken(stepState.StudentIDs[0], entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error creating user: %w", err)
	}
	stepState.AuthToken = token

	ctx = contextWithToken(s, ctx)
	resp, err := pb.NewStudentAssignmentReaderServiceClient(s.Conn).ListSubmissions(ctx, &pb.ListSubmissionsRequest{
		Paging: &common.Paging{
			Limit: 10,
		},
		CourseId: wrapperspb.String(stepState.CourseID),
		Start:    timestamppb.New(time.Now().Add(-1 * time.Hour)),
		End:      timestamppb.New(time.Now().Add(1 * time.Hour)),
	})

	stepState.Response = resp
	stepState.ResponseErr = err

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) modifyNotMarkedStatusToMultiStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	token, err := s.generateExchangeToken(stepState.StudentIDs[0], entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.generateExchangeToken: %w", err)
	}
	stepState.AuthToken = token
	resp, err := pb.NewStudentAssignmentReaderServiceClient(s.Conn).
		ListSubmissions(contextWithToken(s, ctx), &pb.ListSubmissionsRequest{
			Paging: &common.Paging{
				Limit: 10,
			},
			CourseId: wrapperspb.String(stepState.CourseID),
			ClassIds: []string{},
			Start:    timestamppb.New(time.Now().Add(-1 * time.Hour)),
			End:      timestamppb.New(time.Now().Add(1 * time.Hour)),
		})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("ListSubmissions: %w", err)
	}

	for i := 0; i < len(resp.Items); i++ {
		status := rand.Intn(len(epb.SubmissionStatus_name)) + 1
		// multierr.AppendInto(&err, UpdateStatus(ctx, updateCmd, , resp.Items[i].SubmissionId))
		_, err := pb.NewStudentAssignmentWriteServiceClient(s.Conn).UpdateStudentSubmissionsStatus(ctx, &pb.UpdateStudentSubmissionsStatusRequest{
			SubmissionIds: []string{},
			Status:        epb.SubmissionStatus(int32(status)),
		})
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("UpdateStudentSubmissionsStatus: %w", err)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) listTheSubmissionsWithMultiStatusFilter(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	token, err := s.generateExchangeToken(idutil.ULIDNow(), entities.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.generateExchangeToken: %w", err)
	}
	stepState.AuthToken = token
	ctx = contextWithToken(s, ctx)
	FilteredStatus := []pb.SubmissionStatus{1, 2, 3}
	stepState.ListStatus = FilteredStatus

	resp, err := pb.NewStudentAssignmentReaderServiceClient(s.Conn).
		ListSubmissions(ctx, &pb.ListSubmissionsRequest{
			Paging: &common.Paging{
				Limit: 10,
			},
			CourseId: wrapperspb.String(stepState.CourseID),
			ClassIds: []string{},
			Statuses: FilteredStatus,
			Start:    timestamppb.New(time.Now().Add(-1 * time.Hour)),
			End:      timestamppb.New(time.Now().Add(1 * time.Hour)),
		})
	stepState.Response = resp
	stepState.ResponseErr = err

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) ourSystemMustReturnsOnlyLatestSubmissionForSpecificStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("listing return StepStateToContext(ctx, stepState), error: %w", stepState.ResponseErr)
	}

	resp := stepState.Response.(interface {
		GetItems() []*pb.StudentSubmission
	})
	if len(resp.GetItems()) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected 0 submission returned")
	}

	stmt := `
		SELECT student_submission_id 
		FROM student_submissions 
		WHERE 
			study_plan_item_id = $1 
			AND status=ANY($2) 
		ORDER BY created_at DESC, student_submission_id DESC 
		LIMIT 1`
	for _, i := range resp.GetItems()[:len(resp.GetItems())-1] {
		id := pgtype.Text{}
		err := database.Select(ctx, s.DB, stmt, &i.StudyPlanItemId, database.TextArray(convertToValueMapAssignmentSubmission(stepState.ListStatus))).ScanFields(&id)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if id.String != i.SubmissionId {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecting returned submissions (%s) is latest one (but got: %s) -- %s",
				i.SubmissionId, id.String, i.StudyPlanItemId)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validateStudyPlanInfo(ctx context.Context, sub *pb.StudentSubmission) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.CourseID != "" && sub.CourseId != stepState.CourseID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("course id in student submission mismatch: got: %q, want: %q", sub.CourseId, stepState.CourseID)
	}

	var startDate, endDate pgtype.Timestamptz
	query := "SELECT start_date, end_date FROM study_plan_items WHERE study_plan_item_id = $1"
	if err := database.Select(ctx, s.DB, query, database.Text(sub.StudyPlanItemId)).ScanFields(&startDate, &endDate); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if !startDate.Time.Equal(sub.StartDate.AsTime()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("start_date of study plan item is not equal: got: %v, want: %v", startDate.Time, sub.StartDate.AsTime())
	}
	if !endDate.Time.Equal(sub.EndDate.AsTime()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("end_date of study plan item is not equal: got: %v, want: %v", endDate.Time, sub.EndDate.AsTime())
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) teacherListTheSubmissions(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.TeacherID == "" {
		stepState.TeacherID = idutil.ULIDNow()
		if _, err := s.aValidUser(ctx, stepState.TeacherID, consta.RoleTeacher); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create student: %w", err)
		}
	}

	token, err := s.generateExchangeToken(stepState.TeacherID, entities.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken = token

	ctx = contextWithToken(s, ctx)
	resp, err := pb.NewStudentAssignmentReaderServiceClient(s.Conn).
		ListSubmissions(ctx, &pb.ListSubmissionsRequest{
			Paging: &common.Paging{
				Limit: 2,
			},
			CourseId: wrapperspb.String(stepState.CourseID),
			Start:    timestamppb.New(time.Now().Add(-1 * time.Hour)),
			End:      timestamppb.New(time.Now().Add(1 * time.Hour)),
		})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("ListSubmissions: %w", err)
	}

	resp2, err := pb.NewStudentAssignmentReaderServiceClient(s.Conn).
		ListSubmissions(ctx, &pb.ListSubmissionsRequest{
			Paging: &common.Paging{
				Limit: 5,
				Offset: &common.Paging_OffsetCombined{
					OffsetCombined: &common.Paging_Combined{
						OffsetString: resp.NextPage.GetOffsetCombined().OffsetString,
						OffsetTime:   resp.NextPage.GetOffsetCombined().OffsetTime,
					},
				},
			},
			CourseId: wrapperspb.String(stepState.CourseID),
			ClassIds: []string{},
			Start:    timestamppb.New(time.Now().Add(-1 * time.Hour)),
			End:      timestamppb.New(time.Now().Add(1 * time.Hour)),
		})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("ListSubmissions: %w", err)
	}

	if resp2.Items != nil {
		if resp.NextPage.GetOffsetString() == resp2.Items[0].SubmissionId {
			return StepStateToContext(ctx, stepState), errors.New("The first submissionId in resp2 have to difference with next page offset.")
		}
	}

	stepState.Response = resp
	// check total submission ids
	SubIDs := make(map[string]struct{}, 0)
	for i := 0; i < len(resp.Items); i++ {
		SubIDs[resp.Items[i].SubmissionId] = struct{}{}
	}
	for i := 0; i < len(resp2.Items); i++ {
		SubIDs[resp2.Items[i].SubmissionId] = struct{}{}
	}
	if len(SubIDs) != (len(resp.Items) + len(resp2.Items)) {
		return StepStateToContext(ctx, stepState), errors.New("Total real submissions have to equal total expect submissions")
	}
	stepState.ResponseErr = err

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) retrieveSomeElseAssignment(ctx context.Context) (context.Context, []*pb.Content, error) {
	stepState := StepStateFromContext(ctx)
	token, err := s.generateExchangeToken(stepState.StudentIDs[1], entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, fmt.Errorf("s.generateExchangeToken: %w", err)
	}
	stepState.AuthToken = token

	ctx, assignments, err := s.getSomeStudentAssignedAssignments(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, fmt.Errorf("getSomeStudentAssignedAssignments: %w", err)
	}

	return StepStateToContext(ctx, stepState), assignments, nil
}

func (s *suite) retrieveSomeElseSubmissions(ctx context.Context, actor string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, assignments, err := s.retrieveSomeElseAssignment(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.retrieveSomeElseAssignment: %w", err)
	}

	studyPlanItemIDs := make([]string, 0, len(assignments))
	for _, a := range assignments {
		studyPlanItemIDs = append(studyPlanItemIDs, a.StudyPlanItem.StudyPlanItemId)
	}

	token, err := s.generateExchangeToken(stepState.StudentIDs[0], entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.generateExchangeToken: %w", err)
	}
	stepState.AuthToken = token
	resp, err := pb.NewStudentAssignmentReaderServiceClient(s.Conn).
		RetrieveSubmissions(contextWithToken(s, ctx), &pb.RetrieveSubmissionsRequest{
			StudyPlanItemIds: studyPlanItemIDs,
		})

	stepState.Response = resp
	stepState.ResponseErr = err

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) retrieveTheirOwnSubmissions(ctx context.Context, actor string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	token, err := s.generateExchangeToken(stepState.StudentIDs[0], entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.generateExchangeToken: %w", err)
	}
	stepState.AuthToken = token

	studyPlanItemIDs := make([]string, 0, len(stepState.Submissions))
	for _, a := range stepState.Submissions {
		studyPlanItemIDs = append(studyPlanItemIDs, a.Submission.StudyPlanItemId)
	}

	resp, err := pb.NewStudentAssignmentReaderServiceClient(s.Conn).
		RetrieveSubmissions(contextWithToken(s, ctx), &pb.RetrieveSubmissionsRequest{
			StudyPlanItemIds: studyPlanItemIDs,
		})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("RetrieveSubmissions: %w", err)
	}

	stepState.Response = resp
	return StepStateToContext(ctx, stepState), err
}

func convertToValueMapAssignmentSubmission(listStatus []pb.SubmissionStatus) []string {
	detailStatus := make([]string, 0, 5)
	for _, status := range listStatus {
		detailStatus = append(detailStatus, pb.SubmissionStatus_name[int32(status)])
	}
	return detailStatus
}

func (s *suite) ourSystemMustReturnsOnlyLatestSubmissionForEachAssignment(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("listing return error: %w", stepState.ResponseErr)
	}

	resp := stepState.Response.(interface {
		GetItems() []*pb.StudentSubmission
	})
	if len(resp.GetItems()) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected 0 submission returned")
	}

	stmt := `
		SELECT %s 
		FROM student_submissions 
		WHERE 
			study_plan_item_id = $1 
			AND deleted_at IS NULL 
		ORDER BY created_at DESC, student_submission_id DESC LIMIT 1`
	for idx, i := range resp.GetItems() {
		e := &entities.StudentSubmission{}
		fieldNames := database.GetFieldNames(e)
		scanFields := database.GetScanFields(e, fieldNames)
		query := fmt.Sprintf(stmt, strings.Join(fieldNames, " ,"))
		err := database.Select(ctx, s.DB, query, &i.StudyPlanItemId).ScanFields(scanFields...)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if e.ID.String != i.SubmissionId {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecting returned submissions (%s) is latest one (but got: %s) -step %d/%d",
				i.SubmissionId, e.ID.String, idx+1, len(resp.GetItems()))
		}
		if e.Status.String != i.Status.String() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecting returned submissions (%s) is with status %s but got %s -step %d/%d",
				i.SubmissionId, i.Status.String(), e.Status.String, idx+1, len(resp.GetItems()))
		}

		if e.SubmissionGradeID.String != i.SubmissionGradeId.GetValue() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecting returned submissions (%s) is with submission grade id %s but got %s -step %d/%d",
				i.SubmissionId, i.SubmissionGradeId.GetValue(), e.SubmissionGradeID.String, idx+1, len(resp.GetItems()))
		}

		if ctx, err := s.validateStudyPlanInfo(ctx, i); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustReturnsNullGradeContent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("listing return StepStateToContext(ctx, stepState), error: %w", stepState.ResponseErr)
	}

	ids := make([]string, 0, len(stepState.Submissions))
	for _, submission := range stepState.Submissions {
		ids = append(ids, submission.Submission.SubmissionId)
	}
	stmt := `
		SELECT COUNT(*) 
		FROM student_submission_grades ssg 
		JOIN student_submissions ss 
		ON ssg.student_submission_grade_id = ss.student_submission_grade_id 
		WHERE ssg.grade_content IS NULL AND ss.student_submission_id=ANY($1::_TEXT)`
	var total int
	err := s.DB.QueryRow(ctx, stmt, database.TextArray(ids)).Scan(&total)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if total != len(ids) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total null grade content, expected %d, got %d", len(ids), total)
	}
	return StepStateToContext(ctx, stepState), nil
}

func genRand(max int) int {
	return rand.Intn(10) + 1
}

func (s *suite) teacherGradeTheSubmissionsMultipleTimes(ctx context.Context, gradeContentStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	n := genRand(10)
	ctx, err := s.teacherGradeTheSubmissions(ctx, stepState.Submissions, n, gradeContentStatus)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) teacherGradeTheSubmissions(ctx context.Context, submissions []*pb.SubmitAssignmentRequest, times int, gradeContentStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.TeacherID = idutil.ULIDNow()
	if _, err := s.aValidUser(ctx, stepState.TeacherID, consta.RoleTeacher); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create teacher: %w", err)
	}
	token, err := s.generateExchangeToken(stepState.TeacherID, entities.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken = token

	stepState.LatestGrade = make(map[string]*pb.SubmissionGrade)

	client := pb.NewStudentAssignmentWriteServiceClient(s.Conn)
	for submissionPos, submission := range submissions {
		n := times
		for i := 0; i < n; i++ {
			gradeContent := []*pb.SubmissionContent{}
			if gradeContentStatus != "none" {
				gradeContent = []*pb.SubmissionContent{
					{
						SubmitMediaId:     fmt.Sprintf("%s media 1 %d", submission.Submission.SubmissionId, i),
						AttachmentMediaId: fmt.Sprintf("%s attachment 1 %d", submission.Submission.SubmissionId, i),
					},
					{
						SubmitMediaId:     fmt.Sprintf("%s media 2 %d", submission.Submission.SubmissionId, i),
						AttachmentMediaId: fmt.Sprintf("%s attachment 2 %d", submission.Submission.SubmissionId, i),
					},
				}
			}
			grade := &pb.SubmissionGrade{
				SubmissionId: submission.Submission.SubmissionId,
				Note:         fmt.Sprintf("%s note %d", submission.Submission.SubmissionId, i),
				Grade:        float64(i) * 3.14,
				GradeContent: gradeContent,
			}
			resp, err := client.GradeStudentSubmission(contextWithToken(s, ctx), &pb.GradeStudentSubmissionRequest{
				Grade:  grade,
				Status: pb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS,
			})
			submission.Submission.SubmissionGradeId = wrapperspb.String(resp.SubmissionGradeId)
			stepState.Submissions[submissionPos] = submission
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected error when making testing grade: %w", err)
			}

			stepState.LatestGrade[resp.SubmissionGradeId] = grade
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustUpdateTheSubmissionsWithLatestResult(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	studyPlanItemIDs := make([]string, len(stepState.Submissions))
	for _, sub := range stepState.Submissions {
		studyPlanItemIDs = append(studyPlanItemIDs, sub.Submission.StudyPlanItemId)
	}

	client := pb.NewStudentAssignmentReaderServiceClient(s.Conn)
	resp, err := client.RetrieveSubmissions(ctx, &pb.RetrieveSubmissionsRequest{
		StudyPlanItemIds: studyPlanItemIDs,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected error when fecthing test submission: %w", err)
	}

	gradeIDs := make([]string, 0, len(resp.Items))
	for _, ss := range resp.Items {
		if _, ok := stepState.LatestGrade[ss.SubmissionGradeId.Value]; !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("submission grade ID not found: %s", ss.SubmissionGradeId.Value)
		}
		gradeIDs = append(gradeIDs, ss.SubmissionGradeId.Value)
	}

	rGrades, err := client.RetrieveSubmissionGrades(ctx, &pb.RetrieveSubmissionGradesRequest{
		SubmissionGradeIds: gradeIDs,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, g := range rGrades.Grades {
		if !isGradeEqual(stepState.LatestGrade[g.SubmissionGradeId].Grade, g.Grade.Grade) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecting same grade (%.2f) for %s (got %.2f)",
				stepState.LatestGrade[g.SubmissionGradeId].Grade,
				g.SubmissionGradeId,
				g.Grade.Grade,
			)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) someStudentHasTheirSubmissionGraded(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err1 := s.someStudentsAreAssignedSomeValidStudyPlans(ctx)
	if err1 != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.someStudentsAreAssignedSomeValidStudyPlans: %w", err1)
	}
	ctx, err2 := s.studentSubmitTheirAssignmentTimesForDifferentAssignments(ctx, "existed", "single")
	if err2 != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.studentSubmitTheirAssignmentTimesForDifferentAssignments: %w", err2)
	}
	ctx, err3 := s.teacherGradeTheSubmissionsMultipleTimes(ctx, "existed")
	if err3 != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.teacherGradeTheSubmissionsMultipleTimes: %w", err3)
	}
	stepState.NumberOfSubmissionGraded = len(stepState.Submissions)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) someStudentHasTheirSubmissionHaventGraded(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err1 := s.someStudentsAreAssignedSomeValidStudyPlans(ctx)
	ctx, err2 := s.studentSubmitTheirAssignmentTimesForDifferentAssignments(ctx, "existed", "single")
	err := multierr.Combine(err1, err2)
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) teacherChangeStudentsSubmissionStatusTo(ctx context.Context, stringStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	status := pb.SubmissionStatus(pb.SubmissionStatus_value[stringStatus])

	var count int
	query := "SELECT COUNT(*) FROM users where user_id = $1 AND deleted_at IS NULL"
	if err := s.DB.QueryRow(ctx, query, database.Text(stepState.TeacherID)).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if count == 0 {
		if _, err := s.aValidUser(ctx, stepState.TeacherID, consta.RoleTeacher); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create teacher: %w", err)
		}
	}
	token, err := s.generateExchangeToken(stepState.TeacherID, entities.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.generateExchangeToken: %w", err)
	}

	stepState.AuthToken = token

	submissionIDs := make([]string, 0, len(stepState.Submissions))
	client := pb.NewStudentAssignmentWriteServiceClient(s.Conn)
	for _, submission := range stepState.Submissions {
		submissionIDs = append(submissionIDs, submission.Submission.SubmissionId)
	}
	resp, err := client.UpdateStudentSubmissionsStatus(contextWithToken(s, ctx), &pb.UpdateStudentSubmissionsStatusRequest{
		SubmissionIds: submissionIDs,
		Status:        status,
	})

	stepState.Response = resp

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected error when making testing grade: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustUpdateTheSubmissionsStatusTo(ctx context.Context, stringStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.StatusSubmission = stringStatus

	submissionIDs := make([]string, 0, len(stepState.Submissions))
	for _, submission := range stepState.Submissions {
		submissionIDs = append(submissionIDs, submission.Submission.SubmissionId)
	}
	submissionIDs = golibs.Uniq(submissionIDs)

	studentsubmissiongradeRepo := repositories.StudentSubmissionGradeRepo{}
	esRaw, err := studentsubmissiongradeRepo.FindBySubmissionIDs(ctx, s.DB, database.TextArray(submissionIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("studentsubmissiongradeRepo.FindBySubmissionIDs: %w", err)
	}
	es := make([]*entities.StudentSubmissionGrade, 0, len(submissionIDs))
	es = append(es, *esRaw...)
	for _, e := range es {
		if e.EditorID.Status == pgtype.Status(pgtype.Null) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("invalid value: editor id have to not null")
		}
	}
	if len(es) != len(submissionIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not all submission changed status. Expected %d got %d", len(es), len(submissionIDs))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) gradeInfomationHaveToIncludedToSubmissions(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	submissionIDs := make([]string, 0, len(stepState.Submissions))
	for _, submission := range stepState.Submissions {
		submissionIDs = append(submissionIDs, submission.Submission.SubmissionId)
	}
	submissionIDs = golibs.Uniq(submissionIDs)
	query := `
		SELECT count(*)
		FROM student_submissions
		WHERE student_submission_id = ANY($1) AND status = $2 AND student_submission_grade_id IS NOT NULL
	`

	var count int64
	if err := s.DB.QueryRow(ctx, query, &submissionIDs, &stepState.StatusSubmission).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if int(count) != len(submissionIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("submission changed status have to equal total student_submission_grade_id included. Expected %d got %d", count, len(submissionIDs))
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) listTheSubmissionsWithAssignmentName(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.TeacherID = idutil.ULIDNow()
	if _, err := s.aValidUser(ctx, stepState.TeacherID, consta.RoleTeacher); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create teacher: %w", err)
	}
	token, err := s.generateExchangeToken(stepState.TeacherID, entities.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken = token

	ctx = contextWithToken(s, ctx)
	stepState.Response, stepState.ResponseErr = pb.NewStudentAssignmentReaderServiceClient(s.Conn).
		ListSubmissions(ctx, &pb.ListSubmissionsRequest{
			Paging: &common.Paging{
				Limit: 100,
			},
			CourseId:   wrapperspb.String(stepState.CourseID),
			ClassIds:   []string{},
			Start:      timestamppb.New(time.Now().Add(-1 * time.Hour)),
			End:        timestamppb.New(time.Now().Add(1 * time.Hour)),
			SearchText: wrapperspb.String("assignment"),
			SearchType: epb.SearchType_SEARCH_TYPE_ASSIGNMENT_NAME,
		})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) listTheSubmissionsWithInvalidAssignmentName(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.TeacherID = idutil.ULIDNow()
	if _, err := s.aValidUser(ctx, stepState.TeacherID, consta.RoleTeacher); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create teacher: %w", err)
	}
	token, err := s.generateExchangeToken(stepState.TeacherID, entities.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken = token

	ctx = contextWithToken(s, ctx)
	stepState.Response, stepState.ResponseErr = pb.NewStudentAssignmentReaderServiceClient(s.Conn).
		ListSubmissions(ctx, &pb.ListSubmissionsRequest{
			Paging: &common.Paging{
				Limit: 2,
			},
			CourseId:   wrapperspb.String(stepState.CourseID),
			ClassIds:   []string{},
			Start:      timestamppb.New(time.Now().Add(-1 * time.Hour)),
			End:        timestamppb.New(time.Now().Add(1 * time.Hour)),
			SearchText: wrapperspb.String("wrongassignment search"),
			SearchType: epb.SearchType_SEARCH_TYPE_ASSIGNMENT_NAME,
		})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustReturnsEmptySubmission(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.ListSubmissionsResponse)
	if len(rsp.Items) != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error format")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustReturnsSubmissionWithValidAssignmentName(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	rsp := stepState.Response.(*pb.ListSubmissionsResponse)
	var assignmentIDs []string

	if len(rsp.Items) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("system does not return StepStateToContext(ctx, stepState), any submission")
	}
	for _, submission := range rsp.Items {
		assignmentIDs = append(assignmentIDs, submission.AssignmentId)
	}

	query := `SELECT count(*) FROM assignments WHERE assignment_id =ANY($1) AND name LIKE $2`
	var count int64
	err := s.DB.QueryRow(ctx, query, &assignmentIDs, "%assignment%").Scan(&count)

	return StepStateToContext(ctx, stepState), err
}

func isGradeEqual(a, b float64) bool {
	var eps float64 = 0.001
	return math.Abs(a-b) < eps
}

func (s *suite) listTheSubmissionsWithCourseId(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.TeacherID = idutil.ULIDNow()
	if _, err := s.aValidUser(ctx, stepState.TeacherID, consta.RoleTeacher); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create teacher: %w", err)
	}
	token, err := s.generateExchangeToken(stepState.TeacherID, entities.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken = token

	ctx = contextWithToken(s, ctx)
	stepState.Response, stepState.ResponseErr = pb.NewStudentAssignmentReaderServiceClient(s.Conn).
		ListSubmissions(ctx, &pb.ListSubmissionsRequest{
			Paging: &common.Paging{
				Limit: 100,
			},
			CourseId:   wrapperspb.String(stepState.CourseID),
			ClassIds:   []string{},
			Start:      timestamppb.New(time.Now().Add(-1 * time.Hour)),
			End:        timestamppb.New(time.Now().Add(1 * time.Hour)),
			SearchType: epb.SearchType_SEARCH_TYPE_ASSIGNMENT_NAME,
		})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustReturnsSubmissionWithValidCourses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.ListSubmissionsResponse)

	if len(rsp.Items) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("system does not return StepStateToContext(ctx, stepState), any submission")
	}
	for _, submission := range rsp.Items {
		if submission.CourseId != stepState.CourseID {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect course id %s got course id %s", stepState.CourseID, submission.CourseId)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustStoresCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	NumberOfSubmissionNotGraded := len(stepState.Submissions) - stepState.NumberOfSubmissionGraded
	ids := make([]string, 0, len(stepState.Submissions))
	for _, s := range stepState.Submissions {
		ids = append(ids, s.Submission.SubmissionId)
	}
	actualNumberOfSubmissionGraded := 0
	stmtCountGraded := `
		SELECT COUNT(*) 
		FROM student_submission_grades asg 
		JOIN student_submissions ss ON asg.student_submission_grade_id=ss.student_submission_grade_id 
		WHERE ss.student_submission_id=ANY($1::_TEXT) AND asg.grade_content IS NOT NULL`
	err := s.DB.QueryRow(ctx, stmtCountGraded, ids).Scan(&actualNumberOfSubmissionGraded)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if actualNumberOfSubmissionGraded != stepState.NumberOfSubmissionGraded {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total submission grade not null, expected: %d, got: %d ", stepState.NumberOfSubmissionGraded, actualNumberOfSubmissionGraded)
	}
	stmtCountNotGraded := `
		SELECT COUNT(*) 
		FROM student_submission_grades ssg 
		JOIN student_submissions ss ON ssg.student_submission_grade_id = ss.student_submission_grade_id 
		WHERE ssg.grade_content IS NULL AND ss.student_submission_id=ANY($1::_TEXT)`
	var actualNumberOfSubmissionNotGraded int
	err = s.DB.QueryRow(ctx, stmtCountNotGraded, database.TextArray(ids)).Scan(&actualNumberOfSubmissionNotGraded)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if actualNumberOfSubmissionNotGraded != NumberOfSubmissionNotGraded {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total submission grade null, expected: %d, got: %d ", NumberOfSubmissionNotGraded, actualNumberOfSubmissionNotGraded)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentIsRemoveFromAClassAfterTheySubmitTheirSubmission(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.someStudentsAreAssignedSomeValidStudyPlans(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	ctx, err1 := s.studentSubmitTheirAssignment(ctx, "existed", "single")
	ctx, err2 := s.ourSystemMustRecordsAllTheSubmissionsFromStudent(ctx, "single")
	err = multierr.Combine(err1, err2)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for i := range stepState.StudentIDs {
		studentID := stepState.StudentIDs[i]
		token, err := s.generateExchangeToken(studentID, entities.UserGroupStudent)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.AuthToken = token
		ctx, assignments, err := s.getOneStudentAssignedAssignment(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.StudentsSubmittedAssignments[studentID] = append(stepState.StudentsSubmittedAssignments[studentID], assignments...)

		ctx, err = s.submitAssignment(ctx, "existed", "multiple", assignments)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	now := time.Now()
	stepState.Event = &npb.EventSyncStudentPackage{
		StudentPackages: []*npb.EventSyncStudentPackage_StudentPackage{
			{
				ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
				StudentId:  stepState.StudentIDs[0],
				Packages: []*npb.EventSyncStudentPackage_Package{
					{
						CourseIds: []string{stepState.CourseID},
						StartDate: timestamppb.New(now.Add(-7 * time.Second)),
						EndDate:   timestamppb.New(now.Add(7 * time.Second)),
					},
				},
			},
		},
	}
	ctx, err = s.sendEventToNatsJS(ctx, "SyncStudentPackageEvent", constants.SubjectSyncStudentPackage)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	time.Sleep(200 * time.Millisecond)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theResponseSubmissionsDontContainSubmissionOfStudentWhoIsRemovedFromClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.ListSubmissionsResponse)
	for _, item := range resp.Items {
		if item.StudentId == stepState.StudentIDs[0] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect no submission of student %v in the list submission response", stepState.StudentIDs[0])
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustUpdateCreatedAtForEachLatestSubmission(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp := stepState.Response.(*pb.ListSubmissionsResponse)
	for _, submission := range resp.Items {
		var createdAt pgtype.Timestamptz
		query := "SELECT created_at FROM student_submissions WHERE student_submission_id = $1"
		if err := database.Select(ctx, s.DB, query, database.Text(submission.SubmissionId)).ScanFields(&createdAt); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if !createdAt.Time.Equal(submission.CreatedAt.AsTime()) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected created_at %v but got %v", submission.CreatedAt.AsTime(), createdAt.Time)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) teacherSubmitContentAssignmentTimes(ctx context.Context, contentStatus, times string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studentID := stepState.StudentIDs[0]
	stepState.CurrentStudentID = studentID
	token, err := s.generateExchangeToken(studentID, entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.generateExchangeToken (student): %w", err)
	}
	stepState.AuthToken = token

	ctx, assignments, err := s.getOneStudentAssignedAssignment(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.getOneStudentAssignedAssignment: %w", err)
	}

	// teacherID := idutil.ULIDNow()
	// _, err = s.generateExchangeToken(teacherID, consta.RoleTeacher)
	// if err != nil {
	// 	return StepStateToContext(ctx, stepState), fmt.Errorf("s.generateExchangeToken (teacher): %w", err)
	// }

	stepState.StudentsSubmittedAssignments[studentID] = append(stepState.StudentsSubmittedAssignments[studentID], assignments...)
	ctx, err = s.submitAssignment(ctx, contentStatus, times, assignments)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.submitAssignment: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustUpdateDailyLearningTimeCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	learningTimes := make(map[time.Time]int32)

	for _, sub := range stepState.Submissions {
		if sub.GetSubmission().GetCompleteDate().IsValid() {
			dayMidnight := timeutil.MidnightIn(bpb.COUNTRY_VN, sub.GetSubmission().GetCompleteDate().AsTime())
			learningTimes[dayMidnight] += sub.GetSubmission().GetDuration()
		}
	}

	learningTimeRepo := &bob_repository.StudentLearningTimeDailyRepo{}
	for day, duration := range learningTimes {
		pgDay := database.Timestamptz(day)
		result, err := learningTimeRepo.Retrieve(ctx, s.DB, database.Text(stepState.CurrentStudentID), &pgDay, &pgDay)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if len(result) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect learning time '%v' for '%v' but empty", duration, day)
		}
		if result[0].AssignmentLearningTime.Int != duration ||
			result[0].LearningTime.Int != duration {
			return StepStateToContext(ctx, stepState),
				fmt.Errorf("expect learning time '%v' for '%v' but got AssignmentLearningTime=%v LearningTime=%v",
					duration, day, result[0].AssignmentLearningTime.Int, result[0].LearningTime.Int)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
