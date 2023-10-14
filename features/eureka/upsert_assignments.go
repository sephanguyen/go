package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	eureka_entities "github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"golang.org/x/sync/errgroup"
)

func (s *suite) eurekaMustStoreCorrectAssignments(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.UpsertAssignmentsResponse)
	stepState.AssignmentIDs = rsp.AssignmentIds

	query := "SELECT count(*) FROM assignments WHERE assignment_id = ANY($1)"
	var count int
	if err := s.DB.QueryRow(ctx, query, &rsp.AssignmentIds).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if count != len(rsp.AssignmentIds) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Eureka not return StepStateToContext(ctx, stepState), study plan id")
	}
	req := stepState.Request.(*pb.UpsertAssignmentsRequest)
	for _, assignment := range req.Assignments {
		assignmentQuery := `SELECT count(*)
			FROM assignments
			WHERE attachment = $1::text[]
			AND max_grade = $2
			AND instruction = $3
			AND type = $4
			AND name = $5
			AND is_required_grade = $6`

		var count int
		aErr := s.DB.QueryRow(ctx, assignmentQuery, assignment.Attachments, assignment.MaxGrade,
			assignment.Instruction, assignment.AssignmentType.String(), assignment.Name, assignment.RequiredGrade).Scan(&count)
		if aErr != nil {
			return StepStateToContext(ctx, stepState), aErr
		}
		if count != 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error storing assignment: %v", assignment)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateAssignment(ctx context.Context, assignmentID string, allowLate, allowResubmit, gradingMethod bool) (context.Context, *pb.Assignment) {
	stepState := StepStateFromContext(ctx)
	id := assignmentID
	if id == "" {
		id = idutil.ULIDNow()
	}
	if stepState.TopicID == "" {
		stepState.TopicID = idutil.ULIDNow()
	}
	stepState.AssignmentIDs = append(stepState.AssignmentIDs, id)
	return StepStateToContext(ctx, stepState), &pb.Assignment{
		AssignmentId: id,
		Name:         fmt.Sprintf("assignment-%s", idutil.ULIDNow()),
		Content: &pb.AssignmentContent{
			TopicId: stepState.TopicID,
			LoId:    []string{"lo-id-1", "lo-id-2"},
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
			AllowLateSubmission: allowLate,
			AllowResubmission:   allowResubmit,
		},
		RequiredGrade: gradingMethod,
		DisplayOrder:  0,
	}
}

func (s *suite) userCreateNewAssignments(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error

	ass1, ass2, ass3, ass := &pb.Assignment{}, &pb.Assignment{}, &pb.Assignment{}, &pb.Assignment{}
	stepState.SchoolID = strconv.Itoa(constants.ManabieSchool)

	if stepState.BookID == "" {
		ctx, _ = s.aSignedIn(ctx, "school admin")
		// ctx = s.setFakeClaimToContext(ctx, stepState.SchoolID, cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String())
		bookResp, err := pb.NewBookModifierServiceClient(s.Conn).UpsertBooks(s.signedCtx(ctx), &pb.UpsertBooksRequest{
			Books: s.generateBooks(1, nil),
		})

		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create book: %w", err)
		}

		stepState.BookID = bookResp.BookIds[0]
		resp, err := pb.NewChapterModifierServiceClient(s.Conn).UpsertChapters(s.signedCtx(ctx), &pb.UpsertChaptersRequest{
			Chapters: s.generateChapters(ctx, 1, nil),
			BookId:   stepState.BookID,
		})

		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create a chapter: %w", err)
		}

		stepState.ChapterID = resp.ChapterIds[0]

		topic := s.generateValidTopic(stepState.ChapterID)
		topics := []*pb.Topic{&topic}
		ctx = s.setFakeClaimToContext(ctx, stepState.SchoolID, cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String())

		if _, err := pb.NewTopicModifierServiceClient(s.Conn).Upsert(s.signedCtx(ctx), &pb.UpsertTopicsRequest{
			Topics: topics,
		}); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create topics: %w", err)
		}
		stepState.TopicID = topic.Id
	}

	var assignments []*pb.Assignment
	if len(stepState.StudyPlanItemIDs) > 0 {
		// assignment - study plan item has 1-1 relationship, make sure they're equal.
		for i := range stepState.StudyPlanItemIDs {
			ctx, ass = s.generateAssignment(ctx, "", false, false, i == 0)
			assignments = append(assignments, ass)
		}
	} else {
		ctx, ass1 = s.generateAssignment(ctx, "", false, false, true)
		ctx, ass2 = s.generateAssignment(ctx, "", false, false, false)
		ctx, ass3 = s.generateAssignment(ctx, "", false, false, false)
		assignments = []*pb.Assignment{
			ass1, ass2, ass3,
		}
	}

	stepState.Assignments = assignments

	req := &pb.UpsertAssignmentsRequest{
		Assignments: assignments,
	}
	stepState.Request = req
	stepState.Assignments = assignments
	// ctx = s.setFakeClaimToContext(ctx, strconv.Itoa(constants.ManabieSchool), cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String())
	ctx = contextWithToken(s, ctx)
	_, _, stepState.AuthToken, err = s.signedInAs(ctx, constant.RoleSchoolAdmin)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Response, stepState.ResponseErr = pb.NewAssignmentModifierServiceClient(s.Conn).UpsertAssignments(s.signedCtx(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpdateAssignments(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if ctx, err := s.userCreateNewAssignments(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	assignment := stepState.Request.(*pb.UpsertAssignmentsRequest).Assignments[0]
	ctx, assignment = s.generateAssignment(ctx, assignment.AssignmentId, false, false, true)
	assignment.DisplayOrder = 1
	req := &pb.UpsertAssignmentsRequest{
		Assignments: []*pb.Assignment{
			assignment,
		},
	}
	stepState.Request = req

	stepState.Response, stepState.ResponseErr = pb.NewAssignmentModifierServiceClient(s.Conn).UpsertAssignments(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCreateAssignmentWithEmptyAssignmentID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, _ = s.aSignedIn(ctx, "school admin")
	bookResp, err := pb.NewBookModifierServiceClient(s.Conn).UpsertBooks(s.signedCtx(ctx), &pb.UpsertBooksRequest{
		Books: s.generateBooks(1, nil),
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create book: %w", err)
	}
	stepState.BookID = bookResp.BookIds[0]
	resp, err := pb.NewChapterModifierServiceClient(s.Conn).UpsertChapters(s.signedCtx(ctx), &pb.UpsertChaptersRequest{
		Chapters: s.generateChapters(ctx, 1, nil),
		BookId:   stepState.BookID,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create a chapter: %w", err)
	}
	stepState.ChapterID = resp.ChapterIds[0]

	topic := s.generateValidTopic(stepState.ChapterID)
	topics := []*pb.Topic{&topic}
	if _, err := pb.NewTopicModifierServiceClient(s.Conn).Upsert(s.signedCtx(ctx), &pb.UpsertTopicsRequest{
		Topics: topics,
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create topics: %w", err)
	}
	stepState.TopicID = topic.Id

	ctx, assignment := s.generateAssignment(ctx, "", false, false, true)
	assignment.Name = idutil.ULIDNow()
	assignment.DisplayOrder = 1

	req := &pb.UpsertAssignmentsRequest{
		Assignments: []*pb.Assignment{
			assignment,
		},
	}
	stepState.Request = req

	stepState.Response, stepState.ResponseErr = pb.NewAssignmentModifierServiceClient(s.Conn).UpsertAssignments(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eurekaMustStoreCorrectAssignmentWhenCreateAssignment(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	assignment := stepState.Request.(*pb.UpsertAssignmentsRequest).Assignments[0]
	var originalTopic string
	sql := "SELECT original_topic FROM assignments WHERE name = $1"
	if err := s.DB.QueryRow(ctx, sql, &assignment.Name).Scan(&originalTopic); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if originalTopic != assignment.Content.TopicId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("topic_id: expected %v, got %v", originalTopic, assignment.Content.TopicId)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) retrieveAssignmentsWithThatTopic(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = s.setFakeClaimToContext(ctx, "1", "USER_GROUP_SCHOOL_ADMIN")
	e := eureka_entities.Assignment{}
	fields := database.GetFieldNames(&e)
	placeholders := make([]string, len(stepState.AssignmentIDs))
	for i, v := range stepState.AssignmentIDs {
		placeholders[i] = fmt.Sprintf("'%v'", v)
	}
	sql := fmt.Sprintf(
		"SELECT %s FROM %s WHERE assignment_id IN(%s)",
		strings.Join(fields, ","),
		e.TableName(),
		strings.Join(placeholders, ","),
	)
	rows, err := s.DB.Query(ctx, sql)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve assignment by topic: %w", err)
	}
	defer rows.Close()
	var assignments []*eureka_entities.Assignment
	for rows.Next() {
		e := &eureka_entities.Assignment{}

		if err := rows.Scan(database.GetScanFields(e, database.GetFieldNames(e))...); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to scan row: %w", err)
		}

		if e.DisplayOrder.Status != pgtype.Present {
			return StepStateToContext(ctx, stepState), fmt.Errorf("can't create display order for assignment %s", e.ID.String)
		}

		assignments = append(assignments, e)
	}

	stepState.Response = assignments
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsAAssignmentListWithDifferentDisplayOrder(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	assignments := stepState.Response.([]*eureka_entities.Assignment)
	m := make(map[int32]string)
	for _, assignment := range assignments {
		do := assignment.DisplayOrder.Int
		if id, ok := m[do]; ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("learning objective %v and %v have the same display order (%v)", assignment.ID.String, id, do)
		}
		m[do] = assignment.ID.String
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCreateSomeAssignmentsSameTopicAndTime(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, _ = s.aSignedIn(ctx, "school admin")
	bookResp, err := pb.NewBookModifierServiceClient(s.Conn).UpsertBooks(s.signedCtx(ctx), &pb.UpsertBooksRequest{
		Books: s.generateBooks(1, nil),
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create book: %w", err)
	}
	stepState.BookID = bookResp.BookIds[0]
	resp, err := pb.NewChapterModifierServiceClient(s.Conn).UpsertChapters(s.signedCtx(ctx), &pb.UpsertChaptersRequest{
		Chapters: s.generateChapters(ctx, 1, nil),
		BookId:   stepState.BookID,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create a chapter: %w", err)
	}
	stepState.ChapterID = resp.ChapterIds[0]

	topic := s.generateValidTopic(stepState.ChapterID)
	topics := []*pb.Topic{&topic}
	if _, err := pb.NewTopicModifierServiceClient(s.Conn).Upsert(s.signedCtx(ctx), &pb.UpsertTopicsRequest{
		Topics: topics,
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create topics: %w", err)
	}
	stepState.TopicID = topic.Id

	n := rand.Int()%4 + 2
	eg, _ := errgroup.WithContext(ctx)
	for i := 1; i <= n; i++ {
		eg.Go(func() error {
			ctx, ass := s.generateAssignment(ctx, "", false, false, false)
			req := &pb.UpsertAssignmentsRequest{
				Assignments: []*pb.Assignment{ass},
			}
			if _, err := pb.NewAssignmentModifierServiceClient(s.Conn).UpsertAssignments(contextWithToken(s, ctx), req); err != nil {
				return fmt.Errorf("unable to upsert assignemt: %w", err)
			}
			return nil
		})
	}
	return StepStateToContext(ctx, stepState), eg.Wait()
}

type AssignmentDIsplayOrder struct {
	AssignmentID string
	DisplayOrder int32
}

func (s *suite) retrieveAssignments(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ados := stepState.Request.([]*AssignmentDIsplayOrder)
	e := eureka_entities.Assignment{}
	fields := database.GetFieldNames(&e)
	placeholders := make([]string, len(ados))
	for i, v := range ados {
		placeholders[i] = fmt.Sprintf("'%v'", v.AssignmentID)
	}
	sql := fmt.Sprintf(
		"SELECT %s FROM %s WHERE assignment_id IN(%s)",
		strings.Join(fields, ","),
		e.TableName(),
		strings.Join(placeholders, ","),
	)
	rows, err := s.DB.Query(ctx, sql)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve assignment by topic: %w", err)
	}
	defer rows.Close()
	var assignments []*eureka_entities.Assignment
	for rows.Next() {
		e := &eureka_entities.Assignment{}

		if err := rows.Scan(database.GetScanFields(e, database.GetFieldNames(e))...); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to scan row: %w", err)
		}

		if e.DisplayOrder.Status != pgtype.Present {
			return StepStateToContext(ctx, stepState), fmt.Errorf("can't create display order for assignment %s", e.ID.String)
		}

		assignments = append(assignments, e)
	}

	if len(assignments) < 2 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("can't create double assignments")
	}
	stepState.Response = assignments
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsAssignmentListWithDisplayOrderCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.([]*AssignmentDIsplayOrder)
	res := stepState.Response.([]*eureka_entities.Assignment)

	m := make(map[string]int32)
	for _, v := range req {
		m[v.AssignmentID] = v.DisplayOrder
	}

	for _, v := range res {
		do, ok := m[v.ID.String]
		if !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("assignment %s in response is not in request", v.ID.String)
		}
		if int32(v.DisplayOrder.Int) != do {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong when store display order, expect %d but got %v", do, v.DisplayOrder)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCreateAssignmentsWithDisplayOrder(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	n := rand.Int()%5 + 1

	ctx, _ = s.aSignedIn(ctx, "school admin")
	bookResp, err := pb.NewBookModifierServiceClient(s.Conn).UpsertBooks(s.signedCtx(ctx), &pb.UpsertBooksRequest{
		Books: s.generateBooks(1, nil),
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create book: %w", err)
	}
	stepState.BookID = bookResp.BookIds[0]
	resp, err := pb.NewChapterModifierServiceClient(s.Conn).UpsertChapters(s.signedCtx(ctx), &pb.UpsertChaptersRequest{
		Chapters: s.generateChapters(ctx, 1, nil),
		BookId:   stepState.BookID,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create a chapter: %w", err)
	}
	stepState.ChapterID = resp.ChapterIds[0]

	topic := s.generateValidTopic(stepState.ChapterID)
	topics := []*pb.Topic{&topic}
	if _, err := pb.NewTopicModifierServiceClient(s.Conn).Upsert(s.signedCtx(ctx), &pb.UpsertTopicsRequest{
		Topics: topics,
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create topics: %w", err)
	}
	stepState.TopicID = topic.Id

	assignments := make([]*pb.Assignment, 0, n)
	ados := make([]*AssignmentDIsplayOrder, 0, n)

	for i := 0; i <= n; i++ {
		assignment := &pb.Assignment{}
		ctx, assignment = s.generateAssignment(ctx, "", false, false, false)
		assignment.DisplayOrder = int32(i)
		assignments = append(assignments, assignment)
		ados = append(ados, &AssignmentDIsplayOrder{
			AssignmentID: assignment.AssignmentId,
			DisplayOrder: assignment.DisplayOrder,
		})
	}
	if _, err := pb.NewAssignmentModifierServiceClient(s.Conn).UpsertAssignments(contextWithToken(s, ctx), &pb.UpsertAssignmentsRequest{
		Assignments: assignments,
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert assignemt: %w", err)
	}
	stepState.Request = ados
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCreateAssignmentsWithoutDisplayOrder(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	n := rand.Int()%5 + 1

	ctx, _ = s.aSignedIn(ctx, "school admin")
	bookResp, err := pb.NewBookModifierServiceClient(s.Conn).UpsertBooks(s.signedCtx(ctx), &pb.UpsertBooksRequest{
		Books: s.generateBooks(1, nil),
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create book: %w", err)
	}
	stepState.BookID = bookResp.BookIds[0]
	resp, err := pb.NewChapterModifierServiceClient(s.Conn).UpsertChapters(s.signedCtx(ctx), &pb.UpsertChaptersRequest{
		Chapters: s.generateChapters(ctx, 1, nil),
		BookId:   stepState.BookID,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create a chapter: %w", err)
	}
	stepState.ChapterID = resp.ChapterIds[0]

	topic := s.generateValidTopic(stepState.ChapterID)
	topics := []*pb.Topic{&topic}
	if _, err := pb.NewTopicModifierServiceClient(s.Conn).Upsert(s.signedCtx(ctx), &pb.UpsertTopicsRequest{
		Topics: topics,
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create topics: %w", err)
	}
	stepState.TopicID = topic.Id

	assignments := make([]*pb.Assignment, 0, n)
	ados := make([]*AssignmentDIsplayOrder, 0, n)
	for i := 0; i <= n; i++ {
		assignment := &pb.Assignment{}
		ctx, assignment = s.generateAssignment(ctx, "", false, false, false)
		assignments = append(assignments, assignment)
		ados = append(ados, &AssignmentDIsplayOrder{
			AssignmentID: assignment.AssignmentId,
			DisplayOrder: assignment.DisplayOrder,
		})
	}
	if _, err := pb.NewAssignmentModifierServiceClient(s.Conn).UpsertAssignments(contextWithToken(s, ctx), &pb.UpsertAssignmentsRequest{
		Assignments: assignments,
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert assignemt: %w", err)
	}
	stepState.Request = ados
	return StepStateToContext(ctx, stepState), nil
}
