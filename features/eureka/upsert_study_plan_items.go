package eureka

import (
	"context"
	"fmt"
	"time"

	consta "github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) aStudyPlanNameInDb(ctx context.Context, name string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var err error
	_, _, stepState.AuthToken, err = s.signedInAs(ctx, consta.RoleSchoolAdmin)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to login as school admin: %w", err)
	}

	if stepState.CourseID == "" {
		ctx, err := s.aValidCourseBackground(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.aValidCourseBackground: %w", err)
		}
	}

	rsp, err := pb.NewStudyPlanModifierServiceClient(s.Conn).UpsertStudyPlan(contextWithToken(s, ctx), &pb.UpsertStudyPlanRequest{
		SchoolId: constants.ManabieSchool,
		BookId:   stepState.BookID,
		Name:     idutil.ULIDNow(),
		CourseId: stepState.CourseID,
		Status:   pb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("UpsertStudyPlan: %w", err)
	}
	stepState.StudyPlanID = rsp.StudyPlanId
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eurekaMustStoreCorrectStudyPlanItem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	rsp := stepState.Response.(*pb.UpsertStudyPlanItemResponse)
	query := `SELECT count(*) FROM study_plan_items WHERE study_plan_item_id = ANY($1) AND study_plan_id = $2`
	var count int
	if err := s.DB.QueryRow(ctx, query, rsp.StudyPlanItemIds, stepState.StudyPlanID).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if count != len(rsp.StudyPlanItemIds) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Expect %d study plan item but got %d", len(rsp.StudyPlanItemIds), count)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateStudyPlanItem(ctx context.Context, studyPlanItemID string, studyPlanID string) (context.Context, *pb.StudyPlanItem) {
	stepState := StepStateFromContext(ctx)
	now := time.Now()
	availableFrom := timestamppb.New(now.Add(-17 * time.Minute))
	startDate := timestamppb.New(now.Add(-7 * time.Second))
	endDate := timestamppb.New(now.Add(17 * time.Second))
	availableTo := timestamppb.New(now.Add(24 * time.Hour))
	schoolDate := timestamppb.New(now.Add(1 * time.Second))

	contentStructure := &pb.ContentStructure{
		BookId:    "book-id-1",
		ChapterId: "chapter-id-1",
		CourseId:  "course-id-1",
		TopicId:   "topic-id-1",
	}

	if stepState.BookID != "" {
		contentStructure.BookId = stepState.BookID
	}
	if stepState.ChapterID != "" {
		contentStructure.ChapterId = stepState.ChapterID
	}
	if stepState.CourseID != "" {
		contentStructure.CourseId = stepState.CourseID
	}
	if stepState.TopicID != "" {
		contentStructure.TopicId = stepState.TopicID
	}

	return StepStateToContext(ctx, stepState), &pb.StudyPlanItem{
		StudyPlanId:      studyPlanID,
		AvailableFrom:    availableFrom,
		AvailableTo:      availableTo,
		StartDate:        startDate,
		EndDate:          endDate,
		StudyPlanItemId:  studyPlanItemID,
		ContentStructure: contentStructure,
		SchoolDate:       schoolDate,
		Status:           pb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
	}
}

func (s *suite) convertStudyPlanItemEntitiesToPb(studyPlanItem *entities.StudyPlanItem) *pb.StudyPlanItem {
	now := time.Now()
	availableFrom := timestamppb.New(now.Add(-17 * time.Minute))
	startDate := timestamppb.New(now.Add(-7 * time.Second))
	endDate := timestamppb.New(now.Add(17 * time.Second))
	availableTo := timestamppb.New(now.Add(24 * time.Hour))
	schoolDate := timestamppb.New(now.Add(1 * time.Second))

	type ContentStructure struct {
		CourseID     string `json:"course_id,omitempty"`
		BookID       string `json:"book_id,omitempty"`
		ChapterID    string `json:"chapter_id,omitempty"`
		TopicID      string `json:"topic_id,omitempty"`
		LoID         string `json:"lo_id,omitempty"`
		AssignmentID string `json:"assignment_id,omitempty"`
	}

	var content ContentStructure
	studyPlanItem.ContentStructure.AssignTo(&content)

	contentStructure := &pb.ContentStructure{
		BookId:    content.BookID,
		ChapterId: content.ChapterID,
		CourseId:  content.CourseID,
		TopicId:   content.TopicID,
	}

	if content.LoID != "" {
		contentStructure.ItemId = &pb.ContentStructure_LoId{
			LoId: wrapperspb.String(content.LoID),
		}
	}
	if content.AssignmentID != "" {
		contentStructure.ItemId = &pb.ContentStructure_AssignmentId{
			AssignmentId: wrapperspb.String(content.AssignmentID),
		}
	}

	return &pb.StudyPlanItem{
		StudyPlanId:             studyPlanItem.StudyPlanID.String,
		AvailableFrom:           availableFrom,
		AvailableTo:             availableTo,
		StartDate:               startDate,
		EndDate:                 endDate,
		StudyPlanItemId:         studyPlanItem.ID.String,
		ContentStructure:        contentStructure,
		ContentStructureFlatten: studyPlanItem.ContentStructureFlatten.String,
		SchoolDate:              schoolDate,
		DisplayOrder:            studyPlanItem.DisplayOrder.Int,
		Status:                  pb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
	}
}

func (s *suite) generateAListOfStudyPlanItem(ctx context.Context) (context.Context, []*pb.StudyPlanItem, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.StudyPlanID == "" {
		ctx, err := s.aStudyPlanNameInDb(ctx, idutil.ULIDNow())
		if err != nil {
			return StepStateToContext(ctx, stepState), nil, err
		}
	}
	now := time.Now()

	ctx, nullStartDate := s.generateStudyPlanItem(ctx, "", stepState.StudyPlanID)
	nullStartDate.StartDate = nil

	ctx, nullStartDate2 := s.generateStudyPlanItem(ctx, "", stepState.StudyPlanID)
	nullStartDate2.StartDate = nil

	ctx, nullStartDate3 := s.generateStudyPlanItem(ctx, "", stepState.StudyPlanID)
	nullStartDate3.StartDate = nil

	ctx, nullStartDate4 := s.generateStudyPlanItem(ctx, "", stepState.StudyPlanID)
	nullStartDate4.StartDate = nil

	ctx, nullAvailableTo := s.generateStudyPlanItem(ctx, "", stepState.StudyPlanID)
	nullAvailableTo.AvailableTo = nil

	ctx, upCommingItem := s.generateStudyPlanItem(ctx, "", stepState.StudyPlanID)
	upCommingItem.StartDate = timestamppb.New(now.Add(time.Hour))

	ctx, sp1 := s.generateStudyPlanItem(ctx, idutil.ULIDNow(), stepState.StudyPlanID)
	ctx, sp2 := s.generateStudyPlanItem(ctx, idutil.ULIDNow(), stepState.StudyPlanID)
	ctx, sp3 := s.generateStudyPlanItem(ctx, idutil.ULIDNow(), stepState.StudyPlanID)
	ctx, sp4 := s.generateStudyPlanItem(ctx, idutil.ULIDNow(), stepState.StudyPlanID)
	ctx, sp5 := s.generateStudyPlanItem(ctx, idutil.ULIDNow(), stepState.StudyPlanID)
	ctx, sp6 := s.generateStudyPlanItem(ctx, idutil.ULIDNow(), stepState.StudyPlanID)
	ctx, sp7 := s.generateStudyPlanItem(ctx, idutil.ULIDNow(), stepState.StudyPlanID)

	studyPlanItems := []*pb.StudyPlanItem{
		sp1,
		nullStartDate,
		sp2,
		sp3,
		sp4,
		nullStartDate2,
		nullStartDate3,
		sp5,
		sp6,
		sp7,
		upCommingItem,
		nullStartDate4,
		nullAvailableTo,
	}

	studyPlanItems[2].StartDate = studyPlanItems[0].StartDate
	studyPlanItems[3].StartDate = studyPlanItems[0].StartDate
	studyPlanItems[7].StartDate = studyPlanItems[0].StartDate
	studyPlanItems[8].StartDate = studyPlanItems[0].StartDate

	for i, item := range studyPlanItems {
		item.DisplayOrder = int32(i) + 1
	}

	studyPlanItems[3].DisplayOrder, studyPlanItems[2].DisplayOrder = studyPlanItems[2].DisplayOrder, studyPlanItems[3].DisplayOrder // change order

	return StepStateToContext(ctx, stepState), studyPlanItems, nil
}

func (s *suite) generateAListOfStudyPlanItemWithExistStudyPlanItems(ctx context.Context) (context.Context, []*pb.StudyPlanItem) {
	stepState := StepStateFromContext(ctx)
	now := time.Now()

	for i := 0; i < 4; i++ {
		stepState.StudyPlanItems[i].StartDate = nil
	}
	stepState.StudyPlanItems[4].AvailableTo = nil
	stepState.StudyPlanItems[5].StartDate = timestamppb.New(now.Add(time.Hour))

	for i := 7; i < 11; i++ {
		stepState.StudyPlanItems[i].StartDate = stepState.StudyPlanItems[6].StartDate
	}

	for i, item := range stepState.StudyPlanItems {
		item.DisplayOrder = int32(i) + 1
	}

	stepState.StudyPlanItems[3].DisplayOrder, stepState.StudyPlanItems[2].DisplayOrder = stepState.StudyPlanItems[2].DisplayOrder, stepState.StudyPlanItems[3].DisplayOrder // change order

	return StepStateToContext(ctx, stepState), stepState.StudyPlanItems
}

func (s *suite) userUpsertAListOfStudyPlanItemWithExistStudyPlanItems(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.UserId = idutil.ULIDNow()
	ctx, err := s.aValidUser(ctx, stepState.UserId, consta.RoleSchoolAdmin)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, studyPlanItems := s.generateAListOfStudyPlanItemWithExistStudyPlanItems(ctx)

	req := &pb.UpsertStudyPlanItemRequest{
		StudyPlanItems: studyPlanItems,
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewAssignmentModifierServiceClient(s.Conn).UpsertStudyPlanItem(contextWithToken(s, ctx), req)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpsertAListOfStudyPlanItem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.UserId = idutil.ULIDNow()
	ctx, err := s.aValidUser(ctx, stepState.UserId, consta.RoleSchoolAdmin)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, studyPlanItems, err := s.generateAListOfStudyPlanItem(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := &pb.UpsertStudyPlanItemRequest{
		StudyPlanItems: studyPlanItems,
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewAssignmentModifierServiceClient(s.Conn).UpsertStudyPlanItem(contextWithToken(s, ctx), req)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aStudyPlanDoesNotHaveAnyStudyPlanItems(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if ctx, err := s.aStudyPlanNameInDb(ctx, "study_plan_name"); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.BookID = idutil.ULIDNow()
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) bookOfStudyPlanHasStoredCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studyPlanRepo := &repositories.StudyPlanRepo{}
	studyPlan, err := studyPlanRepo.FindByID(ctx, s.DB, database.Text(stepState.StudyPlanID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if studyPlan.BookID.String != stepState.BookID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("book of study plan has stored wrong, expect %v but got %v", stepState.BookID, studyPlan.BookID.String)
	}
	return StepStateToContext(ctx, stepState), nil
}
