package services

import (
	"context"
	"fmt"
	"sort"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AssignmentReaderService struct {
	DB  database.Ext
	Env string

	StudentStudyPlanRepo interface {
		ListStudentAvailableContents(ctx context.Context, db database.QueryExecer, q *repositories.ListStudentAvailableContentsArgs) ([]*entities.StudyPlanItem, error)
		ListStudyPlans(ctx context.Context, db database.QueryExecer, query *repositories.ListStudyPlansArgs) ([]*entities.StudyPlan, error)
		ListActiveStudyPlanItems(ctx context.Context, db database.QueryExecer, q *repositories.ListStudyPlanItemsArgs) ([]*entities.StudyPlanItem, error)
		ListCompletedStudyPlanItems(ctx context.Context, db database.QueryExecer, q *repositories.ListStudyPlanItemsArgs) ([]*entities.StudyPlanItem, error)
		ListOverdueStudyPlanItems(ctx context.Context, db database.QueryExecer, q *repositories.ListStudyPlanItemsArgs) ([]*entities.StudyPlanItem, error)
		ListUpcomingStudyPlanItems(ctx context.Context, db database.QueryExecer, q *repositories.ListStudyPlanItemsArgs) ([]*entities.StudyPlanItem, error)
		CountStudentStudyPlanItems(ctx context.Context, db database.QueryExecer, studentID, studyPlanID pgtype.Text, now pgtype.Timestamptz, onlyCompleted pgtype.Bool) (int, error)
		ListStudyPlanItems(ctx context.Context, db database.QueryExecer, q *repositories.ListStudyPlanItemsArgs) ([]*entities.StudyPlanItem, error)
	}
	LoStudyPlanItemRepo interface {
		FindByStudyPlanItemIDs(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) ([]*entities.LoStudyPlanItem, error)
	}
	AssignmentStudyPlanItemRepo interface {
		FindByStudyPlanItemIDs(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) ([]*entities.AssignmentStudyPlanItem, error)
	}
	AssignmentRepo interface {
		RetrieveAssignments(ctx context.Context, db database.QueryExecer, assignmentIDs pgtype.TextArray) ([]*entities.Assignment, error)
		RetrieveAssignmentsByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.Assignment, error)
	}

	StudyPlanItemRepo interface {
		CountStudentInStudyPlanItem(ctx context.Context, db database.QueryExecer, masterStudentStudyPlanID pgtype.Text, onlyCompleted pgtype.Bool) (map[string]int, []string, error)
		FindByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.StudyPlanItem, error)
		FindAndSortByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.StudyPlanItem, error)
		RetrieveChildStudyPlanItem(ctx context.Context, db database.Ext, studyPlanItemID pgtype.Text, userIDs pgtype.TextArray) (map[string]*entities.StudyPlanItem, error)
		CountStudentStudyPlanItemsInClass(ctx context.Context, db database.Ext, filter *repositories.CountStudentStudyPlanItemsInClassFilter) (int, error)
	}

	TopicsAssignmentsRepo interface {
		RetrieveByAssignmentIDs(ctx context.Context, db database.QueryExecer, assignmentIDs []string) ([]*entities.TopicsAssignments, error)
	}

	TopicRepo interface {
		FindByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs, topicIDs pgtype.TextArray, limit, offset pgtype.Int4) ([]*entities.Topic, error)
	}

	TopicsLearningObjectivesRepo interface {
		RetrieveByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) ([]*repositories.TopicLearningObjective, error)
	}
}

func (s *AssignmentReaderService) findAllStudyPlanItemID(ctx context.Context, args *repositories.ListStudyPlanItemsArgs) ([]*entities.StudyPlanItem, *cpb.Paging, error) {
	items, err := s.StudentStudyPlanRepo.ListStudyPlanItems(ctx, s.DB, args)
	if err != nil {
		return nil, nil, fmt.Errorf("StudentStudyPlanRepo.ListStudyPlanItems: %w", err)
	}
	nextPage := &cpb.Paging{
		Limit: args.Limit,
	}
	if len(items) > 0 {
		lastItem := items[len(items)-1]
		nextPage.Offset = &cpb.Paging_OffsetCombined{
			OffsetCombined: &cpb.Paging_Combined{
				OffsetString:  lastItem.ID.String,
				OffsetInteger: int64(lastItem.DisplayOrder.Int),
			},
		}
	}

	return items, nextPage, nil
}

func (s *AssignmentReaderService) findStudyPlanItemID(ctx context.Context, args *repositories.ListStudyPlanItemsArgs, paging *cpb.Paging, status pb.ToDoStatus) ([]*entities.StudyPlanItem, *cpb.Paging, error) {
	args.DisplayOrder.Set(nil)
	args.StudyPlanItemID.Set(nil)

	if paging != nil {
		if limit := paging.Limit; 1 <= limit && limit <= 100 {
			args.Limit = limit
		}
		if c := paging.GetOffsetCombined(); c != nil {
			if c.OffsetString != "" {
				args.StudyPlanItemID.Set(c.OffsetString)
			}

			if c.OffsetTime != nil && c.OffsetTime.AsTime().Unix() > 0 {
				args.Offset.Set(c.OffsetTime.AsTime())
			} else {
				args.Offset.Set(nil)
			}

			if c.OffsetInteger >= 0 {
				args.DisplayOrder.Set(c.OffsetInteger)
			}
		}
	}

	var (
		items []*entities.StudyPlanItem
		err   error
	)
	switch status {
	case pb.ToDoStatus_TO_DO_STATUS_ACTIVE:
		items, err = s.StudentStudyPlanRepo.ListActiveStudyPlanItems(ctx, s.DB, args)
	case pb.ToDoStatus_TO_DO_STATUS_COMPLETED:
		items, err = s.StudentStudyPlanRepo.ListCompletedStudyPlanItems(ctx, s.DB, args)
	case pb.ToDoStatus_TO_DO_STATUS_OVERDUE:
		items, err = s.StudentStudyPlanRepo.ListOverdueStudyPlanItems(ctx, s.DB, args)
	case pb.ToDoStatus_TO_DO_STATUS_UPCOMING:
		items, err = s.StudentStudyPlanRepo.ListUpcomingStudyPlanItems(ctx, s.DB, args)
	case pb.ToDoStatus_TO_DO_STATUS_NONE:
		return s.findAllStudyPlanItemID(ctx, args)
	default:
		err = fmt.Errorf("unknown todo status: %v", status)
	}
	if err != nil {
		return nil, nil, err
	}

	nextPage := &cpb.Paging{
		Limit: args.Limit,
	}

	if len(items) > 0 {
		lastItem := items[len(items)-1]
		switch status {
		case pb.ToDoStatus_TO_DO_STATUS_ACTIVE, pb.ToDoStatus_TO_DO_STATUS_COMPLETED:
			nextPage.Offset = &cpb.Paging_OffsetCombined{
				OffsetCombined: &cpb.Paging_Combined{
					OffsetTime:    timestamppb.New(lastItem.StartDate.Time),
					OffsetInteger: int64(lastItem.DisplayOrder.Int),
					OffsetString:  lastItem.ID.String,
				},
			}
		default:
			nextPage.Offset = &cpb.Paging_OffsetCombined{
				OffsetCombined: &cpb.Paging_Combined{
					OffsetTime:   timestamppb.New(lastItem.StartDate.Time),
					OffsetString: lastItem.ID.String,
				},
			}
		}
	}

	return items, nextPage, nil
}

func (s *AssignmentReaderService) toDoItems(ctx context.Context, items []*entities.StudyPlanItem, status pb.ToDoStatus) ([]*pb.ToDoItem, error) {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID.String)
	}
	pgIDs := database.TextArray(ids)

	loItems, err := s.LoStudyPlanItemRepo.FindByStudyPlanItemIDs(ctx, s.DB, pgIDs)
	if err != nil {
		return nil, err
	}

	assignmentItems, err := s.AssignmentStudyPlanItemRepo.FindByStudyPlanItemIDs(ctx, s.DB, pgIDs)
	if err != nil {
		return nil, err
	}

	pbItems := make([]*pb.ToDoItem, 0, len(items))
	for _, item := range items {
		resourceID, _, toDoItemType := getResourceID(item, loItems, assignmentItems)
		studyPlanItem := toStudyPlanItemPb(item)
		pbItems = append(pbItems, &pb.ToDoItem{
			StudyPlanItem: studyPlanItem,
			ResourceId:    resourceID,
			Status:        status,
			Type:          toDoItemType,
		})
	}
	return pbItems, nil
}

func (s *AssignmentReaderService) ListStudentToDoItems(ctx context.Context, req *pb.ListStudentToDoItemsRequest) (*pb.ListStudentToDoItemsResponse, error) {
	now := timeutil.Now().UTC()
	pgNow := database.Timestamptz(now)

	args := &repositories.ListStudyPlanItemsArgs{
		StudentID:        database.Text(req.StudentId),
		Now:              pgNow,
		Limit:            10,
		CourseIDs:        database.TextArray(req.CourseIds),
		StudyPlanID:      pgtype.Text{Status: pgtype.Null},
		StudyPlanItemID:  pgtype.Text{Status: pgtype.Null},
		IncludeCompleted: false,
	}

	switch req.Status {
	case pb.ToDoStatus_TO_DO_STATUS_OVERDUE, pb.ToDoStatus_TO_DO_STATUS_UPCOMING, pb.ToDoStatus_TO_DO_STATUS_NONE:
		// offset is start_date, and start_date must be exist in 1st page.
		args.Offset.Set(now)
	case pb.ToDoStatus_TO_DO_STATUS_COMPLETED:
		// offset is start_date, and start_date can be null in 1st page.
		args.Offset.Set(nil)
	default:
		args.Offset.Set(nil)
	}

	items, nextPage, err := s.findStudyPlanItemID(ctx, args, req.Paging, req.Status)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return &pb.ListStudentToDoItemsResponse{}, nil
	}
	pbItems, err := s.toDoItems(ctx, items, req.Status)
	if err != nil {
		return nil, err
	}

	return &pb.ListStudentToDoItemsResponse{
		Items:    pbItems,
		NextPage: nextPage,
	}, nil
}

func (s *AssignmentReaderService) ListStudyPlans(ctx context.Context, req *pb.ListStudyPlansRequest) (*pb.ListStudyPlansResponse, error) {
	query := &repositories.ListStudyPlansArgs{
		StudentID: database.Text(req.StudentId),
		CourseID:  pgtype.Text{Status: pgtype.Null},
		SchoolID:  pgtype.Int4{Status: pgtype.Null},
		Limit:     10,
		Offset:    pgtype.Text{Status: pgtype.Null},
	}
	if req.CourseId != "" {
		query.CourseID.Set(req.CourseId)
	}
	if req.SchoolId != 0 {
		query.SchoolID.Set(req.SchoolId)
	}
	if req.Paging != nil {
		if limit := req.Paging.Limit; 1 <= limit && limit <= 100 {
			query.Limit = limit
		}
		if o := req.Paging.GetOffsetString(); o != "" {
			query.Offset = database.Text(o)
		}
	}

	plans, err := s.StudentStudyPlanRepo.ListStudyPlans(ctx, s.DB, query)
	if err != nil {
		return nil, err
	}
	if len(plans) == 0 {
		return &pb.ListStudyPlansResponse{}, nil
	}

	plansPb := make([]*pb.StudyPlan, 0, len(plans))
	for _, plan := range plans {
		plansPb = append(plansPb, &pb.StudyPlan{
			StudyPlanId:         plan.ID.String,
			Name:                plan.Name.String,
			BookId:              plan.BookID.String,
			Status:              pb.StudyPlanStatus(pb.StudyPlanStatus_value[plan.Status.String]),
			Grades:              database.FromInt4Array(plan.Grades),
			TrackSchoolProgress: plan.TrackSchoolProgress.Bool,
		})
	}

	return &pb.ListStudyPlansResponse{
		Items: plansPb,
		NextPage: &cpb.Paging{
			Limit: query.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: plans[len(plans)-1].ID.String,
			},
		},
	}, nil
}

//nolint
/* 1: Get topics were sorted by display_order (through by book_ids)
 * 2: Get topics_assignments and topics_learning_objectives
 * 3: Replace display_order of study_plan_items by display_order of topics_assignments and topics_learning_objectives
 * 4: Group study_plan_items by topic_id then sort it by display_order (were replaced in previous step)
 * 5: Return study_plan_items depends on order of topics (step 1)
 */
func (s *AssignmentReaderService) ListStudentAvailableContents(ctx context.Context, req *pb.ListStudentAvailableContentsRequest) (*pb.ListStudentAvailableContentsResponse, error) {
	args := convertListStudentAvailableContentsRequestToArgs(ctx, req)
	studyPlanItems, err := s.StudentStudyPlanRepo.ListStudentAvailableContents(
		ctx,
		s.DB,
		args,
	)
	if err != nil {
		return nil, err
	}
	if len(studyPlanItems) == 0 {
		return &pb.ListStudentAvailableContentsResponse{}, nil
	}

	// get bookIDs to listTopicsByBookIDs (topics will be sorted by display_order)
	// get loIDs to get topics_learning_objectives (get display_order to replace display_order of above study_plan_items)
	// get assignmentIDs to get topics_assignments (get display_order to replace display_order of above study_plan_items)
	var (
		bookIDs       []string
		topicIDs      []string
		loIDs         []string
		assignmentIDs []string
	)
	bookIDsMap := make(map[string]bool)
	topicIDsMap := make(map[string]bool)
	studyPlanItemIDs := make([]string, 0, len(studyPlanItems))
	studyPlanItemsMap := make(map[string]*entities.StudyPlanItem)
	loStudyPlanItemsMap := make(map[string][]*entities.StudyPlanItem)
	assignmentStudyPlanItemsMap := make(map[string][]*entities.StudyPlanItem)
	for _, studyPlanItem := range studyPlanItems {
		var contentStructure *entities.ContentStructure

		studyPlanItemIDs = append(studyPlanItemIDs, studyPlanItem.ID.String)
		studyPlanItemsMap[studyPlanItem.ID.String] = studyPlanItem

		if err := studyPlanItem.ContentStructure.AssignTo(&contentStructure); err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("item.ContentStructure.AssignTo: %w", err).Error())
		}

		if contentStructure.LoID != "" {
			loID := contentStructure.LoID
			loIDs = append(loIDs, loID)
			loStudyPlanItemsMap[loID] = append(loStudyPlanItemsMap[loID], studyPlanItem)
		}

		if contentStructure.AssignmentID != "" {
			assignmentID := contentStructure.AssignmentID
			assignmentIDs = append(assignmentIDs, assignmentID)
			assignmentStudyPlanItemsMap[assignmentID] = append(assignmentStudyPlanItemsMap[assignmentID], studyPlanItem)
		}

		if ok := bookIDsMap[contentStructure.BookID]; !ok {
			bookIDsMap[contentStructure.BookID] = true
			bookIDs = append(bookIDs, contentStructure.BookID)
		}
		if ok := topicIDsMap[contentStructure.TopicID]; !ok {
			topicIDsMap[contentStructure.TopicID] = true
			topicIDs = append(topicIDs, contentStructure.TopicID)
		}
	}

	// listTopicsByBookIDs
	// list loStudyPlanItems and assignmentStudyPlanItems to map type of content (response)
	topicLearningObjectivesMap := make(map[string][]*repositories.TopicLearningObjective)
	topics, err := s.retrieveTopics(ctx, nil, bookIDs, topicIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.retrieveTopics: %v", err).Error())
	}

	if len(loIDs) != 0 {
		topicLearningObjectives, err := s.TopicsLearningObjectivesRepo.RetrieveByLoIDs(ctx, s.DB, database.TextArray(loIDs))
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("s.TopicLearningObjectivesRepo.RetrieveByLOIDs: %v", err).Error())
		}
		if err != nil {
			return nil, err
		}

		if topicLearningObjectives != nil {
			for _, topicLearningObjective := range topicLearningObjectives {
				topicLearningObjectivesMap[topicLearningObjective.Topic.ID.String] = append(topicLearningObjectivesMap[topicLearningObjective.Topic.ID.String], topicLearningObjective)
			}
		}
	}

	topicAssignmentsMap := make(map[string][]*entities.TopicsAssignments)

	// get topicAssignments, topicLearningObjectivs and replace display_order of studyPlayItems (above)
	if len(assignmentIDs) != 0 {
		topicAssignments, err := s.TopicsAssignmentsRepo.RetrieveByAssignmentIDs(ctx, s.DB, assignmentIDs)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("s.TopicsAssignmentsRepo.RetrieveByAssignmentIDs: %v", err).Error())
		}
		for _, topicAssignment := range topicAssignments {
			topicAssignmentsMap[topicAssignment.TopicID.String] = append(topicAssignmentsMap[topicAssignment.TopicID.String], topicAssignment)
		}
	}

	type StudyPlanItemWithResourceIDAndType struct {
		*entities.StudyPlanItem
		ResourceID string
		Type       pb.ContentType
	}

	studyPlanItemsSorted := make([]StudyPlanItemWithResourceIDAndType, 0, len(studyPlanItems))
	for _, topic := range topics {
		var studyPlanItemsTmp []StudyPlanItemWithResourceIDAndType

		for _, topicAssignment := range topicAssignmentsMap[topic.ID.String] {
			items := assignmentStudyPlanItemsMap[topicAssignment.AssignmentID.String]
			for _, item := range items {
				item.DisplayOrder.Set(topicAssignment.DisplayOrder.Int)
				studyPlanItemsTmp = append(studyPlanItemsTmp, StudyPlanItemWithResourceIDAndType{
					StudyPlanItem: item,
					ResourceID:    topicAssignment.AssignmentID.String,
					Type:          pb.ContentType_CONTENT_TYPE_ASSIGNMENT,
				})
			}
		}

		for _, topicLO := range topicLearningObjectivesMap[topic.ID.String] {
			items := loStudyPlanItemsMap[topicLO.LearningObjective.ID.String]
			for _, item := range items {
				item.DisplayOrder.Set(topicLO.DisplayOrder)
				studyPlanItemsTmp = append(studyPlanItemsTmp, StudyPlanItemWithResourceIDAndType{
					StudyPlanItem: item,
					ResourceID:    topicLO.LearningObjective.ID.String,
					Type:          pb.ContentType_CONTENT_TYPE_LO,
				})
			}
		}

		sort.SliceStable(studyPlanItemsTmp, func(i, j int) bool {
			return studyPlanItemsTmp[i].DisplayOrder.Int < studyPlanItemsTmp[j].DisplayOrder.Int
		})

		studyPlanItemsSorted = append(studyPlanItemsSorted, studyPlanItemsTmp...)
	}

	// generate contents and map type of contents
	contents := make([]*pb.Content, 0, len(studyPlanItemsSorted))
	for _, item := range studyPlanItemsSorted {
		contents = append(contents, &pb.Content{
			StudyPlanItem: toStudyPlanItemPb(item.StudyPlanItem),
			ResourceId:    item.ResourceID,
			Type:          item.Type,
		})
	}

	return &pb.ListStudentAvailableContentsResponse{Contents: contents}, nil
}

func (s *AssignmentReaderService) retrieveTopics(ctx context.Context, paging *cpb.Paging, bookIDs, topicIDs []string) (topics []*entities.Topic, err error) {
	if paging != nil {
		if paging.GetOffsetInteger() < 0 {
			return nil, status.Error(codes.InvalidArgument, "offset must be positive")
		}

		if paging.Limit <= 0 {
			paging.Limit = 100
		}

		offset := paging.GetOffsetInteger()
		limit := paging.Limit

		topics, err = s.TopicRepo.FindByBookIDs(ctx, s.DB, database.TextArray(bookIDs), database.TextArray(topicIDs), database.Int4(int32(limit)), database.Int4(int32(offset)))
	} else {
		topics, err = s.TopicRepo.FindByBookIDs(ctx, s.DB, database.TextArray(bookIDs), database.TextArray(topicIDs), pgtype.Int4{Status: pgtype.Null}, pgtype.Int4{Status: pgtype.Null})
	}

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("crs.TopicRepo.FindByBookIDs: %w", err).Error())
	}

	return topics, nil
}

func ToTopicPbV1(p *entities.Topic) *cpb.Topic {
	topic := &cpb.Topic{
		Info: &cpb.ContentBasicInfo{
			Id:           p.ID.String,
			Name:         p.Name.String,
			Country:      cpb.Country(cpb.Country_value[p.Country.String]),
			Subject:      cpb.Subject(cpb.Subject_value[p.Subject.String]),
			Grade:        int32(p.Grade.Int),
			SchoolId:     p.SchoolID.Int,
			DisplayOrder: int32(p.DisplayOrder.Int),
			IconUrl:      p.IconURL.String,
			CreatedAt:    &timestamppb.Timestamp{Seconds: p.CreatedAt.Time.Unix()},
			UpdatedAt:    &timestamppb.Timestamp{Seconds: p.UpdatedAt.Time.Unix()},
		},
		Type:        cpb.TopicType(cpb.TopicType_value[p.TopicType.String]),
		Status:      cpb.TopicStatus(cpb.TopicStatus_value[p.Status.String]),
		ChapterId:   p.ChapterID.String,
		Instruction: p.Instruction.String,
	}

	numAttachment := len(p.AttachmentNames.Elements)
	if n := len(p.AttachmentURLs.Elements); n < numAttachment {
		numAttachment = n
	}

	for i := 0; i < numAttachment; i++ {
		topic.Attachments = append(topic.Attachments, &cpb.Attachment{
			Name: p.AttachmentNames.Elements[i].String,
			Url:  p.AttachmentURLs.Elements[i].String,
		})
	}

	return topic
}

func convertListStudentAvailableContentsRequestToArgs(ctx context.Context, req *pb.ListStudentAvailableContentsRequest) *repositories.ListStudentAvailableContentsArgs {
	userID := interceptors.UserIDFromContext(ctx)
	now := database.Timestamptz(timeutil.Now().UTC())
	q := &repositories.ListStudentAvailableContentsArgs{
		StudentID:    database.Text(userID),
		StudyPlanIDs: database.TextArray(req.StudyPlanId),
		Offset:       now,
		BookID:       database.Text(req.BookId),
		ChapterID:    database.Text(req.ChapterId),
		TopicID:      database.Text(req.TopicId),
		CourseID:     database.Text(req.CourseId),
	}

	if req.BookId == "" {
		q.BookID.Set(nil)
	}
	if req.ChapterId == "" {
		q.ChapterID.Set(nil)
	}
	if req.TopicId == "" {
		q.TopicID.Set(nil)
	}
	if req.CourseId == "" {
		q.CourseID.Set(nil)
	}

	return q
}

func toAssignmentPb(src *entities.Assignment) (*pb.Assignment, error) {
	var content *entities.AssignmentContent
	var attachment []string
	var settings *entities.AssignmentSetting
	err := multierr.Combine(
		src.Content.AssignTo(&content),
		src.Attachment.AssignTo(&attachment),
		src.Settings.AssignTo(&settings),
	)
	if err != nil {
		return nil, err
	}
	return &pb.Assignment{
		AssignmentId: src.ID.String,
		Name:         src.Name.String,
		MaxGrade:     int64(src.MaxGrade.Int),
		Content: &pb.AssignmentContent{
			LoId:    content.LoIDs,
			TopicId: content.TopicID,
		},
		AssignmentType:   pb.AssignmentType(pb.AssignmentType_value[src.Type.String]),
		Attachments:      attachment,
		Instruction:      src.Instruction.String,
		AssignmentStatus: pb.AssignmentStatus(pb.AssignmentStatus_value[src.Status.String]),
		Setting: &pb.AssignmentSetting{
			AllowLateSubmission:       settings.AllowLateSubmission,
			AllowResubmission:         settings.AllowResubmission,
			RequireAssignmentNote:     settings.RequireAssignmentNote,
			RequireAttachment:         settings.RequireAttachment,
			RequireVideoSubmission:    settings.RequireVideoSubmission,
			RequireDuration:           settings.RequireDuration,
			RequireCorrectness:        settings.RequireCorrectness,
			RequireCompleteDate:       settings.RequireCompleteDate,
			RequireUnderstandingLevel: settings.RequireUnderstandingLevel,
		},
		RequiredGrade: src.IsRequiredGrade.Bool,
		DisplayOrder:  src.DisplayOrder.Int,
	}, nil
}

func (s *AssignmentReaderService) RetrieveAssignments(ctx context.Context, req *pb.RetrieveAssignmentsRequest) (*pb.RetrieveAssignmentsResponse, error) {
	assignments, err := s.AssignmentRepo.RetrieveAssignments(ctx, s.DB, database.TextArray(req.Ids))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "s.AssignmentRepo.RetrieveAssignments: %v", err)
	}

	assignmentsPb := make([]*pb.Assignment, len(assignments))
	for i, assignment := range assignments {
		assignmentsPb[i], err = toAssignmentPb(assignment)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "error convert assignment to pb: %v", err)
		}
	}
	return &pb.RetrieveAssignmentsResponse{
		Items: assignmentsPb,
	}, nil
}

// ListCourseTodo list statistic per study plan item in course
func (s *AssignmentReaderService) ListCourseTodo(ctx context.Context, req *pb.ListCourseTodoRequest) (*pb.ListCourseTodoResponse, error) {
	if req.StudyPlanId == "" {
		return &pb.ListCourseTodoResponse{}, nil
	}

	// count total student
	studyPlanItemTotalStudentMap, studyPlanItemIDs, err := s.StudyPlanItemRepo.CountStudentInStudyPlanItem(ctx, s.DB, database.Text(req.StudyPlanId), database.Bool(false))
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("ListCourseTodo.StudyPlanItemRepo.CountStudentInStudyPlanItem with total student err: %v", err))
	}
	if len(studyPlanItemTotalStudentMap) == 0 {
		return &pb.ListCourseTodoResponse{}, nil
	}

	// count completed student
	studyPlanItemCompletedStudentMap, _, err := s.StudyPlanItemRepo.CountStudentInStudyPlanItem(ctx, s.DB, database.Text(req.StudyPlanId), database.Bool(true))
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("ListCourseTodo.StudyPlanItemRepo.CountStudentInStudyPlanItem with completed student err: %v", err))
	}

	studyPlanItems, err := s.StudyPlanItemRepo.FindAndSortByIDs(ctx, s.DB, database.TextArray(studyPlanItemIDs))
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("ListCourseTodo.StudyPlanItemRepo.FindByIDs err: %v", err))
	}

	pbToDoItems, err := s.toDoItems(ctx, studyPlanItems, pb.ToDoStatus_TO_DO_STATUS_NONE)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("ListCourseTodo.StudyPlanItemRepo.toDoItems err: %v", err))
	}

	respItems := make([]*pb.StatisticTodoItem, 0, len(studyPlanItemIDs))
	for _, studyPlanItem := range pbToDoItems {
		total := studyPlanItemTotalStudentMap[studyPlanItem.StudyPlanItem.StudyPlanItemId]
		completed := studyPlanItemCompletedStudentMap[studyPlanItem.StudyPlanItem.StudyPlanItemId]
		respItems = append(respItems, &pb.StatisticTodoItem{
			Item:                 studyPlanItem,
			CompletedStudent:     int32(completed),
			TotalAssignedStudent: int32(total),
		})
	}

	return &pb.ListCourseTodoResponse{StatisticItems: respItems}, nil
}

func (s *AssignmentReaderService) GetChildStudyPlanItems(ctx context.Context, req *pb.GetChildStudyPlanItemsRequest) (*pb.GetChildStudyPlanItemsResponse, error) {
	if req.StudyPlanItemId == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid argument: study plan id have to not empty")
	}

	studyPlanItems, err := s.StudyPlanItemRepo.RetrieveChildStudyPlanItem(ctx, s.DB, database.Text(req.StudyPlanItemId), database.TextArray(req.UserIds))
	if err != nil {
		return nil, fmt.Errorf("StudyPlanItemRepo.RetrieveChildStudyPlanItem: %w", err)
	}

	return &pb.GetChildStudyPlanItemsResponse{
		Items: toUserStudyPlanItem(studyPlanItems),
	}, nil
}

func (s *AssignmentReaderService) RetrieveStatisticAssignmentClass(ctx context.Context, req *pb.RetrieveStatisticAssignmentClassRequest) (*pb.RetrieveStatisticAssignmentClassResponse, error) {
	if req.ClassId == "" || req.StudyPlanItemId == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid arguments: the params have to not empty")
	}
	getTotalFilter := &repositories.CountStudentStudyPlanItemsInClassFilter{
		ClassID:         database.Text(req.ClassId),
		StudyPlanItemID: database.Text(req.StudyPlanItemId),
		IsCompleted:     database.Bool(false),
	}

	getCompletedFilter := &repositories.CountStudentStudyPlanItemsInClassFilter{
		ClassID:         database.Text(req.ClassId),
		StudyPlanItemID: database.Text(req.StudyPlanItemId),
		IsCompleted:     database.Bool(true),
	}

	studyPlanItems, err := s.StudyPlanItemRepo.FindByIDs(ctx, s.DB, database.TextArray([]string{req.StudyPlanItemId}))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("StudyPlanItemRepo.FindByIDs: %w", err).Error())
	}
	if len(studyPlanItems) == 0 {
		return &pb.RetrieveStatisticAssignmentClassResponse{}, nil
	}

	totalAssigned, err := s.StudyPlanItemRepo.CountStudentStudyPlanItemsInClass(ctx, s.DB, getTotalFilter)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("StudyPlanItemRepo.CountStudentStudyPlanItemsInClass: %w", err).Error())
	}

	// reuse filter
	totalCompleted, err := s.StudyPlanItemRepo.CountStudentStudyPlanItemsInClass(ctx, s.DB, getCompletedFilter)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("StudyPlanItemRepo.CountStudentStudyPlanItemsInClass: %w", err).Error())
	}

	return &pb.RetrieveStatisticAssignmentClassResponse{
		StatisticItem: &pb.StatisticTodoItem{
			Item: &pb.ToDoItem{
				StudyPlanItem: &pb.StudyPlanItem{
					StudyPlanId:     studyPlanItems[0].StudyPlanID.String,
					StudyPlanItemId: studyPlanItems[0].ID.String,
					AvailableFrom:   timestamppb.New(studyPlanItems[0].AvailableFrom.Time),
					AvailableTo:     timestamppb.New(studyPlanItems[0].AvailableTo.Time),
					StartDate:       timestamppb.New(studyPlanItems[0].StartDate.Time),
					EndDate:         timestamppb.New(studyPlanItems[0].EndDate.Time),
					DisplayOrder:    studyPlanItems[0].DisplayOrder.Int,
				},
			},
			CompletedStudent:     int32(totalCompleted),
			TotalAssignedStudent: int32(totalAssigned),
		},
	}, nil
}

func toUserStudyPlanItem(userStudyPlanItems map[string]*entities.StudyPlanItem) []*pb.GetChildStudyPlanItemsResponse_UserStudyPlanItem {
	res := make([]*pb.GetChildStudyPlanItemsResponse_UserStudyPlanItem, 0, len(userStudyPlanItems))
	for uID, item := range userStudyPlanItems {
		res = append(res, &pb.GetChildStudyPlanItemsResponse_UserStudyPlanItem{
			UserId:        uID,
			StudyPlanItem: toStudyPlanItemPb(item),
		})
	}
	return res
}

func getResourceID(e *entities.StudyPlanItem, loItems []*entities.LoStudyPlanItem, assignmentItems []*entities.AssignmentStudyPlanItem) (string, pb.ContentType, pb.ToDoItemType) {
	for _, item := range loItems {
		if item.StudyPlanItemID.String == e.ID.String {
			return item.LoID.String, pb.ContentType_CONTENT_TYPE_LO, pb.ToDoItemType_TO_DO_ITEM_TYPE_LO
		}
	}
	for _, item := range assignmentItems {
		if item.StudyPlanItemID.String == e.ID.String {
			return item.AssignmentID.String, pb.ContentType_CONTENT_TYPE_ASSIGNMENT, pb.ToDoItemType_TO_DO_ITEM_TYPE_ASSIGNMENT
		}
	}
	return "", pb.ContentType_CONTENT_TYPE_NONE, pb.ToDoItemType_TO_DO_ITEM_TYPE_NONE
}

func toStudyPlanItemPb(e *entities.StudyPlanItem) *pb.StudyPlanItem {
	var cs entities.ContentStructure
	if err := e.ContentStructure.AssignTo(&cs); err != nil {
		return nil
	}

	var startDate *timestamppb.Timestamp
	if e.StartDate.Status == pgtype.Present {
		startDate = timestamppb.New(e.StartDate.Time)
	}
	var endDate *timestamppb.Timestamp
	if e.EndDate.Status == pgtype.Present {
		endDate = timestamppb.New(e.EndDate.Time)
	}
	var completedAt *timestamppb.Timestamp
	if e.CompletedAt.Status == pgtype.Present {
		completedAt = timestamppb.New(e.CompletedAt.Time)
	}

	var schoolDate *timestamppb.Timestamp
	if e.SchoolDate.Status == pgtype.Present {
		schoolDate = timestamppb.New(e.SchoolDate.Time)
	}

	return &pb.StudyPlanItem{
		StudyPlanId:     e.StudyPlanID.String,
		StudyPlanItemId: e.ID.String,
		AvailableFrom:   timestamppb.New(e.AvailableFrom.Time),
		AvailableTo:     timestamppb.New(e.AvailableTo.Time),
		StartDate:       startDate,
		EndDate:         endDate,
		CompletedAt:     completedAt,
		ContentStructure: &pb.ContentStructure{
			CourseId:  cs.CourseID,
			BookId:    cs.BookID,
			ChapterId: cs.ChapterID,
			TopicId:   cs.TopicID,
		},
		DisplayOrder: e.DisplayOrder.Int,
		SchoolDate:   schoolDate,
		Status:       pb.StudyPlanItemStatus(pb.StudyPlanItemStatus_value[e.Status.String]),
	}
}

func (s *AssignmentReaderService) RetrieveStudyPlanProgress(ctx context.Context, req *pb.RetrieveStudyPlanProgressRequest) (*pb.RetrieveStudyPlanProgressResponse, error) {
	pgStudentID := database.Text(req.StudentId)
	pgStudyPlanID := database.Text(req.StudyPlanId)
	now := database.Timestamptz(timeutil.Now())

	totalCompleted, err := s.StudentStudyPlanRepo.CountStudentStudyPlanItems(ctx, s.DB, pgStudentID, pgStudyPlanID, now, database.Bool(true))
	if err != nil {
		return nil, err
	}

	totalAssigned, err := s.StudentStudyPlanRepo.CountStudentStudyPlanItems(ctx, s.DB, pgStudentID, pgStudyPlanID, now, database.Bool(false))
	if err != nil {
		return nil, err
	}

	return &pb.RetrieveStudyPlanProgressResponse{
		CompletedAssignments: int32(totalCompleted),
		TotalAssignments:     int32(totalAssigned),
	}, nil
}
