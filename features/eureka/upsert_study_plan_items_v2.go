package eureka

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) userUpsertAListOfStudyPlanItemV2(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CourseID = s.newID()
	ctx = StepStateToContext(ctx, stepState)
	ctx, studyPlanItems, err := s.generateAListOfStudyPlanItemV2(ctx)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), nil
	}
	req := &pb.UpsertStudyPlanItemV2Request{
		StudyPlanItems: studyPlanItems,
	}

	stepState.StudyPlanItems = studyPlanItems
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewStudyPlanModifierServiceClient(s.Conn).UpsertStudyPlanItemV2(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eurekaMustStoreCorrectStudyPlanItemV2(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	rsp := stepState.Response.(*pb.UpsertStudyPlanItemV2Response)

	// check number of study_plan_items
	{
		query := `SELECT count(*) FROM study_plan_items WHERE study_plan_item_id = ANY($1) AND study_plan_id = $2`
		var count int
		if err := s.DB.QueryRow(ctx, query, rsp.StudyPlanItemIds, stepState.StudyPlanID).Scan(&count); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if count != len(rsp.StudyPlanItemIds) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect %d study plan item but got %d", len(rsp.StudyPlanItemIds), count)
		}
	}

	// check fields status, content_structure, school_date
	{
		query := `SELECT study_plan_item_id, content_structure, school_date, status FROM study_plan_items WHERE study_plan_item_id = ANY($1) AND study_plan_id = $2`
		rows, err := s.DB.Query(ctx, query, rsp.StudyPlanItemIds, stepState.StudyPlanID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		defer rows.Close()
		studyPlanItemsMap := make(map[string]*entities.StudyPlanItem)
		for rows.Next() {
			var (
				id, status       pgtype.Text
				contentStructure pgtype.JSONB
				schoolDate       pgtype.Timestamptz
			)

			if err := rows.Scan(&id, &contentStructure, &schoolDate, &status); err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			studyPlanItemsMap[id.String] = &entities.StudyPlanItem{
				ID:               id,
				ContentStructure: contentStructure,
				SchoolDate:       schoolDate,
				Status:           status,
			}
		}

		for _, studyPlanItem := range stepState.StudyPlanItems {
			enStudyPlanItem := studyPlanItemsMap[studyPlanItem.StudyPlanItemId]
			if enStudyPlanItem == nil {
				continue
			}

			if studyPlanItem.Status.String() != enStudyPlanItem.Status.String {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expect status %s study plan item but got %s", studyPlanItem.Status.String(), enStudyPlanItem.Status.String)
			}

			if studyPlanItem.SchoolDate.AsTime().Second() != enStudyPlanItem.SchoolDate.Time.Second() {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expect school_date %s study plan item but got %s", studyPlanItem.SchoolDate.AsTime().String(), enStudyPlanItem.SchoolDate.Time.String())
			}

			var contentStructure *entities.ContentStructure
			enStudyPlanItem.ContentStructure.AssignTo(&contentStructure)

			if studyPlanItem.ContentStructure.GetLoId() != nil {
				if studyPlanItem.ContentStructure.GetLoId().Value != contentStructure.LoID {
					return StepStateToContext(ctx, stepState), fmt.Errorf("expect content_structure -> lo_id %s study plan item but got %s", studyPlanItem.ContentStructure.GetLoId().String(), contentStructure.LoID)
				}
			}

			if studyPlanItem.ContentStructure.GetAssignmentId() != nil {
				if studyPlanItem.ContentStructure.GetAssignmentId().Value != contentStructure.AssignmentID {
					return StepStateToContext(ctx, stepState), fmt.Errorf("expect content_structure -> assignment_id %s study plan item but got %s", studyPlanItem.ContentStructure.GetAssignmentId().String(), contentStructure.AssignmentID)
				}
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateStudyPlanItemV2(ctx context.Context, studyPlanItemID string, studyPlanID string) (context.Context, *pb.StudyPlanItem) {
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
		ItemId: &pb.ContentStructure_LoId{
			LoId: &wrapperspb.StringValue{
				Value: idutil.ULIDNow(),
			},
		},
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

func (s *suite) generateAListOfStudyPlanItemV2(ctx context.Context) (context.Context, []*pb.StudyPlanItem, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.StudyPlanID == "" {
		ctx, err := s.aStudyPlanNameInDb(ctx, idutil.ULIDNow())
		if err != nil {
			return StepStateToContext(ctx, stepState), nil, err
		}
	}

	ctx, nullStartDate := s.generateStudyPlanItemV2(ctx, idutil.ULIDNow(), stepState.StudyPlanID)
	nullStartDate.StartDate = nil

	ctx, nullStartDate2 := s.generateStudyPlanItemV2(ctx, idutil.ULIDNow(), stepState.StudyPlanID)
	nullStartDate2.StartDate = nil

	ctx, nullStartDate3 := s.generateStudyPlanItemV2(ctx, idutil.ULIDNow(), stepState.StudyPlanID)
	nullStartDate3.StartDate = nil

	ctx, nullStartDate4 := s.generateStudyPlanItemV2(ctx, idutil.ULIDNow(), stepState.StudyPlanID)
	nullStartDate4.StartDate = nil

	ctx, nullAvailableTo := s.generateStudyPlanItemV2(ctx, idutil.ULIDNow(), stepState.StudyPlanID)
	nullAvailableTo.AvailableTo = nil
	ctx, sp1 := s.generateStudyPlanItemV2(ctx, idutil.ULIDNow(), stepState.StudyPlanID)
	ctx, sp2 := s.generateStudyPlanItemV2(ctx, idutil.ULIDNow(), stepState.StudyPlanID)
	ctx, sp3 := s.generateStudyPlanItemV2(ctx, idutil.ULIDNow(), stepState.StudyPlanID)
	ctx, sp4 := s.generateStudyPlanItemV2(ctx, idutil.ULIDNow(), stepState.StudyPlanID)
	ctx, sp5 := s.generateStudyPlanItemV2(ctx, idutil.ULIDNow(), stepState.StudyPlanID)
	ctx, sp6 := s.generateStudyPlanItemV2(ctx, idutil.ULIDNow(), stepState.StudyPlanID)
	ctx, sp7 := s.generateStudyPlanItemV2(ctx, idutil.ULIDNow(), stepState.StudyPlanID)

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

func (s *suite) generateAStudyPlanItemV2(ctx context.Context) (context.Context, []*pb.StudyPlanItem, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.StudyPlanID == "" {
		ctx, err := s.aStudyPlanNameInDb(ctx, idutil.ULIDNow())
		if err != nil {
			return StepStateToContext(ctx, stepState), nil, err
		}
	}

	stepState.StudyPlanItemID = idutil.ULIDNow()
	ctx, sp1 := s.generateStudyPlanItemV2(ctx, stepState.StudyPlanItemID, stepState.StudyPlanID)
	studyPlanItems := []*pb.StudyPlanItem{
		sp1,
	}

	return StepStateToContext(ctx, stepState), studyPlanItems, nil
}
