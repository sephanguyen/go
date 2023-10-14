package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) addAdHocAssignment(ctx context.Context, arg string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var times int

	switch arg {
	case "one time":
		times = 1
	case "multiple times":
		times = 5 + int(rand.Int31n(10))
	}

	now := time.Now()

	for i := 0; i < times; i++ {
		ctx, ass1 := s.generateAssignment(ctx, "", false, false, true)

		req := &pb.UpsertAdHocAssignmentRequest{
			CourseId:    stepState.CourseID,
			StudentId:   stepState.StudentID,
			ChapterName: "chapter example",
			TopicName:   "topic example",
			StartDate:   timestamppb.New(now.Add(-24 * time.Hour)),
			EndDate:     timestamppb.New(now.Add(24 * time.Hour)),
			Assignment:  ass1,
		}

		stepState.Request = req
		stepState.Response, stepState.ResponseErr = pb.NewAssignmentModifierServiceClient(s.Conn).UpsertAdHocAssignment(s.signedCtx(ctx), req)
		if stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), stepState.ResponseErr
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) addAdHocAssignmentWith(ctx context.Context, req string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var upsertAdHocAssignmentReq *pb.UpsertAdHocAssignmentRequest

	switch req {
	case "missing course_id":
		upsertAdHocAssignmentReq = &pb.UpsertAdHocAssignmentRequest{}
	case "missing student_id":
		upsertAdHocAssignmentReq = &pb.UpsertAdHocAssignmentRequest{
			StudentId: stepState.StudentID,
		}
	case "valid":
		now := time.Now()
		_, ass1 := s.generateAssignment(ctx, "", false, false, true)
		upsertAdHocAssignmentReq = &pb.UpsertAdHocAssignmentRequest{
			CourseId:    stepState.CourseID,
			StudentId:   stepState.StudentID,
			ChapterName: "chapter example",
			TopicName:   "topic example",
			StartDate:   timestamppb.New(now.Add(-24 * time.Hour)),
			EndDate:     timestamppb.New(now.Add(24 * time.Hour)),
			Assignment:  ass1,
		}
	}

	stepState.Request = upsertAdHocAssignmentReq
	stepState.Response, stepState.ResponseErr = pb.NewAssignmentModifierServiceClient(s.Conn).UpsertAdHocAssignment(s.signedCtx(ctx), upsertAdHocAssignmentReq)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustAddAdhocAssignment(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.UpsertAdHocAssignmentRequest)

	// query addHoc assignments
	var assignmentIDs pgtype.TextArray
	query := `SELECT array_agg(distinct assignment_id) 
	FROM assignments 
	WHERE assignment_id = ANY($1) AND 
	deleted_at IS NULL`
	if err := s.DB.QueryRow(ctx, query, &stepState.AssignmentIDs).Scan(&assignmentIDs); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if len(assignmentIDs.Elements) != len(stepState.AssignmentIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("number of ad hoc assignment: expect %v, got %v", len(stepState.AssignmentIDs), len(assignmentIDs.Elements))
	}

	// query adHoc book
	var bookIDs pgtype.TextArray
	query = `SELECT array_agg(distinct b.book_id)
	FROM books as b
	JOIN books_chapters as bc ON b.book_id = bc.book_id
	JOIN chapters as c ON bc.chapter_id = c.chapter_id
	JOIN topics as t ON t.chapter_id = c.chapter_id
	JOIN assignments as a ON a.original_topic = t.topic_id
	WHERE b.deleted_at IS NULL AND
	b.book_type = 'BOOK_TYPE_ADHOC'::TEXT AND
	bc.deleted_at IS NULL AND
	c.deleted_at IS NULL AND
	t.deleted_at IS NULL AND
	a.deleted_at IS NULL AND
	a.assignment_id = ANY($1)
	`
	if err := s.DB.QueryRow(ctx, query, &stepState.AssignmentIDs).Scan(&bookIDs); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if len(bookIDs.Elements) != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("number of ad hoc book: expect %v, got %v", len(bookIDs.Elements), 1)
	}

	// query studyPlan of adHoc book
	var studyPlanIDs pgtype.TextArray
	query = `SELECT array_agg(distinct study_plan_id)
	FROM study_plans
	WHERE book_id = ANY($1) AND
	deleted_at IS NULL`

	if err := s.DB.QueryRow(ctx, query, &bookIDs).Scan(&studyPlanIDs); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if len(studyPlanIDs.Elements) != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("number of studyPlans of adHoc book: expect %v, got %v", len(studyPlanIDs.Elements), 1)
	}

	spl := &entities.StudyPlanItem{}
	splFields, _ := spl.FieldMap()
	spls := entities.StudyPlanItems{}
	query = fmt.Sprintf(`SELECT %s
	FROM study_plan_items
	WHERE study_plan_id = ANY($1) AND
	content_structure ->> 'assignment_id' = ANY($2) AND
	deleted_at IS NULL`, strings.Join(splFields, ", "))

	if err := database.Select(ctx, s.DB, query, &studyPlanIDs, &assignmentIDs).ScanAll(&spls); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	t, _ := time.Parse("2006/02/01 15:04", "2300/01/01 23:59")
	studyPlanItemIDs := make([]string, 0, len(spls))
	for _, studyPlanItem := range spls {
		studyPlanItemIDs = append(studyPlanItemIDs, studyPlanItem.ID.String)
		if req.StartDate != nil {
			if !compareTimesWithoutNsec(studyPlanItem.AvailableFrom.Time, req.StartDate.AsTime()) || !compareTimesWithoutNsec(studyPlanItem.StartDate.Time, req.StartDate.AsTime()) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("availableFrom or startDate must be equal to req.StartDate %v %v %v", studyPlanItem.AvailableFrom.Time, studyPlanItem.StartDate.Time, req.StartDate.AsTime())
			}
		}
		if req.EndDate != nil {
			if !compareTimesWithoutNsec(studyPlanItem.AvailableTo.Time, t) || !compareTimesWithoutNsec(studyPlanItem.EndDate.Time, req.EndDate.AsTime()) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("availableTo or endDate must be equal to req.EndDate %v %v %v", studyPlanItem.AvailableTo.Time, studyPlanItem.EndDate.Time, req.EndDate.AsTime())
			}
		} else {
			if !compareTimesWithoutNsec(studyPlanItem.AvailableTo.Time, t) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("availableTo must be equal 2030/11/11 23:59:00")
			}
		}
	}
	if len(studyPlanItemIDs) != len(assignmentIDs.Elements) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("number of studyPlanItems of adHoc book: expect %v, got %v", len(studyPlanItemIDs), len(assignmentIDs.Elements))
	}

	resp, err := pb.NewAssignmentReaderServiceClient(s.Conn).ListStudentAvailableContents(s.signedCtx(ctx), &pb.ListStudentAvailableContentsRequest{})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(resp.Contents) != len(assignmentIDs.Elements) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("number of studentAvailableContents must be equal to number of assignments")
	}

	return StepStateToContext(ctx, stepState), nil
}

func compareTimesWithoutNsec(a, b time.Time) bool {
	if a.Year() != b.Year() {
		return false
	}
	if a.Month() != b.Month() {
		return false
	}
	if a.Day() != b.Day() {
		return false
	}
	if a.Hour() != b.Hour() {
		return false
	}
	if a.Minute() != b.Minute() {
		return false
	}
	if a.Second() != b.Second() {
		return false
	}

	return true
}

func (s *suite) ourSystemMustAddAdhocAssignmentCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	spl := &entities.StudyPlanItem{}
	splFields, _ := spl.FieldMap()
	query := fmt.Sprintf(`SELECT %s
	FROM study_plan_items
	WHERE study_plan_id = ANY($1) 
	AND content_structure ->> 'assignment_id' = ANY($2) 
	AND deleted_at IS NULL`, strings.Join(splFields, ", "))
	spls := &entities.StudyPlanItems{}
	if err := database.Select(ctx, s.DB, query, &stepState.StudyPlanIDs, &stepState.AssignmentIDs).ScanAll(spls); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	t, _ := time.Parse("2006/02/01 15:04", "2300/01/01 23:59")
	for _, spl := range *spls {
		if !compareTimesWithoutNsec(spl.AvailableTo.Time, t) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("available to of %s store wrong, expect %s but got %s", spl.ID.String, t.String(), spl.AvailableTo.Time.String())
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
