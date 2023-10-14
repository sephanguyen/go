package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type BobStudentReaderServiceClient interface {
	RetrieveStudentProfile(ctx context.Context, in *bpb.RetrieveStudentProfileRequest, opts ...grpc.CallOption) (*bpb.RetrieveStudentProfileResponse, error)
	RetrieveStudentSchoolHistory(ctx context.Context, in *bpb.RetrieveStudentSchoolHistoryRequest, opts ...grpc.CallOption) (*bpb.RetrieveStudentSchoolHistoryResponse, error)
}

type AssignmentModifierService struct {
	DB  database.Ext
	JSM nats.JetStreamManagement

	StudyPlanRepo interface {
		Insert(ctx context.Context, db database.QueryExecer, studyPlan *entities.StudyPlan) (pgtype.Text, error)
		FindByID(ctx context.Context, db database.QueryExecer, studyPlanID pgtype.Text) (*entities.StudyPlan, error)
		FindByIDs(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) ([]*entities.StudyPlan, error)
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlan) error
		BulkCopy(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) ([]string, []string, error)
		BulkUpdateBook(ctx context.Context, db database.QueryExecer, spbs []*repositories.StudyPlanBook) error
		RetrieveStudyPlanItemInfo(ctx context.Context, db database.QueryExecer, args repositories.StudyPlanItemInfoArgs) ([]*repositories.StudyPlanItemInfo, error)
	}
	StudyPlanItemRepo interface {
		BulkInsert(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlanItem) error
		BulkSync(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlanItem) ([]*entities.StudyPlanItem, error)
		FindByStudyPlanID(ctx context.Context, db database.QueryExecer, studyPlanID pgtype.Text) ([]*entities.StudyPlanItem, error)
		BulkCopy(ctx context.Context, db database.QueryExecer, originalStudyPlanIDs pgtype.TextArray, newStudyPlanIDs pgtype.TextArray) error
		UpdateWithCopiedFromItem(ctx context.Context, db database.QueryExecer, studyPlanItems []*entities.StudyPlanItem) error
		UpdateCompletedAtByID(ctx context.Context, db database.QueryExecer, id pgtype.Text, completedAt pgtype.Timestamptz) error
		SoftDeleteByStudyPlanItemIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) error
		FindByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.StudyPlanItem, error)
		DeleteStudyPlanItemsByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) error
		DeleteStudyPlanItemsByStudyPlans(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
		UpdateSchoolDate(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, studentID pgtype.Text, schoolDate pgtype.Timestamptz) error
		UpdateStatus(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, studentID pgtype.Text, status pgtype.Text) error
	}
	AssignmentRepo AssignmentRepo

	AssignmentStudyPlanItemRepo interface {
		BulkInsert(ctx context.Context, db database.QueryExecer, assignmentStudyPlanItems []*entities.AssignmentStudyPlanItem) error
		FindByStudyPlanItemIDs(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) ([]*entities.AssignmentStudyPlanItem, error)
		CopyFromStudyPlan(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
		EditAssignmentTime(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, studyPlanItemIDs pgtype.TextArray, startDate, endDate pgtype.Timestamptz) error
		SoftDeleteByAssigmentIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) (pgtype.TextArray, error)
		BulkEditAssignmentTime(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, ens []*entities.StudyPlanItem) error
		BulkUpsertByStudyPlanItem(ctx context.Context, db database.QueryExecer, assignmentStudyPlanItems []*entities.AssignmentStudyPlanItem) error
	}
	LoStudyPlanItemRepo interface {
		BulkInsert(ctx context.Context, db database.QueryExecer, assignmentStudyPlanItems []*entities.LoStudyPlanItem) error
		CopyFromStudyPlan(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
		FindByStudyPlanItemIDs(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) ([]*entities.LoStudyPlanItem, error)
		DeleteLoStudyPlanItemsByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) error
		DeleteLoStudyPlanItemsAndStudyPlanItemByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) error
	}
	CourseStudyPlanRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, courseStudyPlans []*entities.CourseStudyPlan) error
	}
	StudentRepo interface {
		FindStudentsByCourseID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text) (*pgtype.TextArray, error)
	}
	ClassStudyPlanRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, classStudyPlans []*entities.ClassStudyPlan) error
	}
	StudentStudyPlanRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, studentStudyPlans []*entities.StudentStudyPlan) error
	}
	TopicsAssignmentsRepo TopicsAssignmentsRepo

	TopicRepo interface {
		RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.Topic, error)
		RetrieveByID(ctx context.Context, db database.QueryExecer, id pgtype.Text, enhancers ...repositories.QueryEnhancer) (*entities.Topic, error)
		FindByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs, topicIDs pgtype.TextArray, limit, offset pgtype.Int4) ([]*entities.Topic, error)
		FindByIDsV2(ctx context.Context, db database.QueryExecer, ids []string, isAll bool) (map[string]*entities.Topic, error)
		BulkImport(ctx context.Context, db database.QueryExecer, cc []*entities.Topic) error
		UpdateTotalLOs(ctx context.Context, db database.QueryExecer, topicID pgtype.Text) error
		UpdateLODisplayOrderCounter(ctx context.Context, db database.QueryExecer, topicID pgtype.Text, number pgtype.Int4) error
		BulkUpsertWithoutDisplayOrder(ctx context.Context, db database.QueryExecer, topics []*entities.Topic) error
		UpdateStatus(ctx context.Context, db database.Ext, ids pgtype.TextArray, topicStatus pgtype.Text) error
		SoftDelete(ctx context.Context, db database.QueryExecer, topicIDs []string) (int, error)
	}

	BookRepo interface {
		FindByIDs(ctx context.Context, db database.QueryExecer, bookIDs []string) (map[string]*entities.Book, error)
		FindByID(ctx context.Context, db database.QueryExecer, bookID pgtype.Text, enhancers ...repositories.QueryEnhancer) (*entities.Book, error)
		RetrieveBookTreeByBookID(ctx context.Context, db database.QueryExecer, bookID pgtype.Text) ([]*repositories.BookTreeInfo, error)
		RetrieveAdHocBookByCourseIDAndStudentID(ctx context.Context, db database.QueryExecer, courseID, studentID pgtype.Text) (*entities.Book, error)
		UpdateCurrentChapterDisplayOrder(ctx context.Context, db database.QueryExecer, totalGeneratedChapterDisplayOrder pgtype.Int4, bookID pgtype.Text) error
		Upsert(ctx context.Context, db database.Ext, cc []*entities.Book) error
	}

	ChapterRepo interface {
		FindByID(ctx context.Context, db database.QueryExecer, chapterID pgtype.Text, enhancers ...repositories.QueryEnhancer) (*entities.Chapter, error)
		FindByIDs(ctx context.Context, db database.QueryExecer, chapterIDs []string) (map[string]*entities.Chapter, error)
		Upsert(ctx context.Context, db database.QueryExecer, cc []*entities.Chapter) error
		UpsertWithoutDisplayOrderWhenUpdate(ctx context.Context, db database.QueryExecer, cc []*entities.Chapter) error
		UpdateCurrentTopicDisplayOrder(ctx context.Context, db database.QueryExecer, totalGeneratedTopicDisplayOrder pgtype.Int4, chapterID pgtype.Text) error
	}

	BookChapterRepo interface {
		Upsert(ctx context.Context, db database.Ext, cc []*entities.BookChapter) error
		SoftDelete(ctx context.Context, db database.QueryExecer, chapterIDs, bookIDs pgtype.TextArray) error
		SoftDeleteByChapterIDs(ctx context.Context, db database.QueryExecer, chapterIDs []string) error
		RetrieveContentStructuresByTopics(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) (map[string][]entities.ContentStructure, error)
	}

	CourseBookRepo interface {
		FindByCourseIDAndBookID(ctx context.Context, db database.QueryExecer, bookID, courseID pgtype.Text) (*entities.CoursesBooks, error)
		Upsert(ctx context.Context, db database.Ext, cc []*entities.CoursesBooks) error
	}

	LearningObjectiveRepo interface {
		RetrieveLearningObjectivesByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.LearningObjective, error)
	}

	BobStudentReaderSvc BobStudentReaderServiceClient
}

type AssignmentRepo interface {
	BulkUpsert(ctx context.Context, db database.QueryExecer, assignments []*entities.Assignment) error
	SoftDelete(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) error
	RetrieveAssignments(ctx context.Context, db database.QueryExecer, assignmentIDs pgtype.TextArray) ([]*entities.Assignment, error)
	RetrieveAssignmentsByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.Assignment, error)
}

type TopicsAssignmentsRepo interface {
	Upsert(ctx context.Context, db database.QueryExecer, topicsAssignments *entities.TopicsAssignments) error
	BulkUpsert(ctx context.Context, db database.QueryExecer, topicsAssignmentsList []*entities.TopicsAssignments) error
	SoftDeleteByAssignmentIDs(ctx context.Context, db database.QueryExecer, assignmentIDs pgtype.TextArray) error
}

func UpsertStudyPlansInTx(ctx context.Context, req *pb.UpsertStudyPlansRequest, tx pgx.Tx,
	bulkUpsertStudyPlans func(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlan) error) ([]string, error) {
	studyPlans := make([]*entities.StudyPlan, len(req.StudyPlans))
	studyPlanIDs := make([]string, len(req.StudyPlans))
	now := timeutil.Now()
	for i, studyPlan := range req.StudyPlans {
		if studyPlan.Type == pb.StudyPlanType_STUDY_PLAN_TYPE_NONE {
			studyPlan.Type = pb.StudyPlanType_STUDY_PLAN_TYPE_COURSE
		}
		e := &entities.StudyPlan{}
		database.AllNullEntity(e)
		err := multierr.Combine(
			e.ID.Set(idutil.ULIDNow()),
			e.Name.Set(studyPlan.Name),
			e.SchoolID.Set(studyPlan.SchoolId),
			e.CourseID.Set(studyPlan.CourseId),
			e.StudyPlanType.Set(studyPlan.Type.String()),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
			e.Status.Set("STUDY_PLAN_STATUS_ACTIVE"),
			e.TrackSchoolProgress.Set(false),
			e.Grades.Set("{}"),
			e.BookID.Set(studyPlan.BookId),
		)
		if err != nil {
			return nil, fmt.Errorf("error create study plan: %w", err)
		}

		if studyPlan.StudyPlanId != nil {
			_ = e.ID.Set(studyPlan.StudyPlanId.Value)
		}

		studyPlans[i] = e
		studyPlanIDs[i] = e.ID.String
	}
	err := bulkUpsertStudyPlans(ctx, tx, studyPlans)
	if err != nil {
		return nil, fmt.Errorf("bulkUpsertStudyPlans: %w", err)
	}
	return studyPlanIDs, nil
}

func toAssignmentConfig(src *pb.AssignmentSetting) *entities.AssignmentSetting {
	if src == nil {
		return &entities.AssignmentSetting{}
	}
	return &entities.AssignmentSetting{
		AllowLateSubmission:       src.AllowLateSubmission,
		AllowResubmission:         src.AllowResubmission,
		RequireAssignmentNote:     src.RequireAssignmentNote,
		RequireAttachment:         src.RequireAttachment,
		RequireVideoSubmission:    src.RequireVideoSubmission,
		RequireCompleteDate:       src.RequireCompleteDate,
		RequireDuration:           src.RequireDuration,
		RequireCorrectness:        src.RequireCorrectness,
		RequireUnderstandingLevel: src.RequireUnderstandingLevel,
	}
}

func toAssignmentToDoList(src *pb.CheckList) *entities.AssignmentCheckList {
	e := &entities.AssignmentCheckList{}
	e.CheckList = make(map[string]bool)
	if src == nil {
		return e
	}
	for _, item := range src.Items {
		e.CheckList[item.Content] = item.IsChecked
	}
	return e
}

func toAssignmentContent(src *pb.AssignmentContent) *entities.AssignmentContent {
	return &entities.AssignmentContent{
		TopicID: src.TopicId,
		LoIDs:   src.LoId,
	}
}

func toAssignmentEn(src *pb.Assignment) (*entities.Assignment, bool, error) {
	isAutoGenID := false
	var dst entities.Assignment
	database.AllNullEntity(&dst)
	assignmentConfig := toAssignmentConfig(src.Setting)
	checkList := toAssignmentToDoList(src.CheckList)
	now := timeutil.Now()
	if src.AssignmentId == "" {
		src.AssignmentId = idutil.ULIDNow()
		isAutoGenID = true
	}
	assignmentContent := toAssignmentContent(src.Content)
	err := multierr.Combine(
		dst.ID.Set(src.AssignmentId),
		dst.Settings.Set(assignmentConfig),
		dst.Attachment.Set(src.Attachments),
		dst.CheckList.Set(checkList.CheckList),
		dst.Type.Set(src.AssignmentType.String()),
		dst.MaxGrade.Set(src.MaxGrade),
		dst.CreatedAt.Set(now),
		dst.UpdatedAt.Set(now),
		dst.Instruction.Set(src.Instruction),
		dst.Name.Set(src.Name),
		dst.Content.Set(assignmentContent),
		dst.IsRequiredGrade.Set(src.RequiredGrade),
		dst.DisplayOrder.Set(src.DisplayOrder),
		dst.OriginalTopic.Set(assignmentContent.TopicID),
		dst.TopicID.Set(assignmentContent.TopicID),
	)
	if len(checkList.CheckList) == 0 {
		err = multierr.Append(err, dst.CheckList.Set(nil))
	}
	if src.Instruction == "" {
		err = multierr.Append(err, dst.Instruction.Set(nil))
	}
	return &dst, isAutoGenID, err
}

func toGeneralAssignmentEnt(src *sspb.AssignmentBase) (*entities.GeneralAssignment, error) {
	var dst entities.GeneralAssignment
	database.AllNullEntity(&dst)
	id := idutil.ULIDNow()
	if src.Base.LearningMaterialId != "" {
		id = src.Base.LearningMaterialId
	}
	now := timeutil.Now()
	err := multierr.Combine(
		dst.LearningMaterial.ID.Set(id),
		dst.LearningMaterial.TopicID.Set(src.Base.TopicId),
		dst.LearningMaterial.Name.Set(src.Base.Name),
		dst.LearningMaterial.Type.Set(sspb.LearningMaterialType_LEARNING_MATERIAL_GENERAL_ASSIGNMENT.String()),
		dst.IsPublished.Set(false),
		dst.SetDefaultVendorType(),
		dst.LearningMaterial.CreatedAt.Set(now),
		dst.LearningMaterial.UpdatedAt.Set(now),
		dst.Attachments.Set(src.Attachments),
		dst.MaxGrade.Set(src.MaxGrade),
		dst.Instruction.Set(src.Instruction),
		dst.IsRequiredGrade.Set(src.IsRequiredGrade),
		dst.AllowResubmission.Set(src.AllowResubmission),
		dst.RequireAttachment.Set(src.RequireAttachment),
		dst.AllowLateSubmission.Set(src.AllowLateSubmission),
		dst.RequireAssignmentNote.Set(src.RequireAssignmentNote),
		dst.RequireVideoSubmission.Set(src.RequireVideoSubmission),
	)
	return &dst, err
}

func (s *AssignmentModifierService) UpsertAssignment(ctx context.Context, req *pb.UpsertAssignmentsRequest) (*pb.UpsertAssignmentsResponse, error) {
	b, _ := json.Marshal(req)
	r := &pb.UpsertAssignmentsRequest{}
	json.Unmarshal(b, r)
	resp, err := s.UpsertAssignments(ctx, r)
	if err != nil {
		return nil, err
	}
	b, _ = json.Marshal(resp)
	response := &pb.UpsertAssignmentsResponse{}
	json.Unmarshal(b, response)
	return response, err
	// return upsertAssignment(ctx, req, s.DB, s.AssignmentRepo, s.BobInternalAssignmentModifier)
}

func upsertAssignment(
	ctx context.Context,
	req *pb.UpsertAssignmentsRequest,
	db database.Ext,
	assignmentRepo AssignmentRepo,
) (*pb.UpsertAssignmentsResponse, error) {
	ids := make([]string, 0, len(req.Assignments))
	for _, assignment := range req.Assignments {
		if assignment.Name == "" {
			return nil, status.Error(codes.InvalidArgument, "empty assignment name")
		}
		en, _, err := toAssignmentEn(assignment)
		if err != nil {
			return nil, err
		}

		ids = append(ids, en.ID.String)
	}
	assignments, err := assignmentRepo.RetrieveAssignments(
		ctx,
		db,
		database.TextArray(ids),
	)
	if err != nil && err != pgx.ErrNoRows {
		return nil, status.Errorf(codes.FailedPrecondition, fmt.Errorf("unable to retrieve assignment by ids: %w", err).Error())
	}
	assignmentsReq := make([]*bpb.Assignment, 0, len(assignments))
	for _, assignment := range assignments {
		assignmentsReq = append(assignmentsReq, &bpb.Assignment{
			Id:           assignment.ID.String,
			DisplayOrder: assignment.DisplayOrder.Int,
		})
	}

	return &pb.UpsertAssignmentsResponse{
		AssignmentIds: ids,
	}, nil
}

func (s *AssignmentModifierService) UpsertAssignmentsData(ctx context.Context, req *pb.UpsertAssignmentsDataRequest) (*emptypb.Empty, error) {
	assignmentEntities := make([]*entities.Assignment, 0, len(req.Assignments))
	topicAssignmentEntities := make([]*entities.TopicsAssignments, 0, len(req.Assignments))
	for _, assignment := range req.Assignments {
		en, _, err := toAssignmentEn(assignment)
		if err != nil {
			return nil, err
		}
		assignmentEntities = append(assignmentEntities, en)
		var assignmentContent entities.AssignmentContent
		_ = en.Content.AssignTo(&assignmentContent)
		topicAssignmentEntities = append(topicAssignmentEntities, &entities.TopicsAssignments{
			TopicID:      database.Text(assignmentContent.TopicID),
			AssignmentID: database.Text(en.ID.String),
			DisplayOrder: database.Int2(int16(en.DisplayOrder.Int)),
		})
	}
	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.AssignmentRepo.BulkUpsert(ctx, tx, assignmentEntities); err != nil {
			return fmt.Errorf("unable to bulk upsert assignment: %w", err)
		}
		// if in future we the feature `link_lo/assignment back, we have to check on topics_assignments too`
		if err := s.TopicsAssignmentsRepo.BulkUpsert(ctx, tx, topicAssignmentEntities); err != nil {
			return fmt.Errorf("unable to bulk upsert topic assignment: %w", err)
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	data := &npb.EventAssignmentsCreated{
		Assignments: req.Assignments,
	}
	msg, _ := proto.Marshal(data)
	_, err := s.JSM.PublishContext(ctx, constants.SubjectAssignmentsCreated, msg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.JSM.PublishContext: subject: %q, %v", constants.SubjectAssignmentsCreated, err).Error())
	}

	return &emptypb.Empty{}, nil
}

func toCourseStudyPlanEn(courseID string, studyPlanID string) (*entities.CourseStudyPlan, error) {
	e := &entities.CourseStudyPlan{}
	now := timeutil.Now()
	database.AllNullEntity(e)
	err := multierr.Combine(
		e.StudyPlanID.Set(studyPlanID),
		e.CourseID.Set(courseID),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	)
	return e, err
}

func ToClassStudyPlanEn(classID int32, studyPlanID string) (*entities.ClassStudyPlan, error) {
	e := &entities.ClassStudyPlan{}
	now := timeutil.Now()
	database.AllNullEntity(e)
	err := multierr.Combine(
		e.StudyPlanID.Set(studyPlanID),
		e.ClassID.Set(classID),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	)
	return e, err
}

// TO-DO uncomment when implement class level study plan
// func (s *AssignmentModifierService) makeClassStudyPlanEn(classIDs []int32, studyPlanIDs []string) ([]*entities.ClassStudyPlan, error) {
// 	classStudyPlans := make([]*entities.ClassStudyPlan, 0, len(classIDs))

// 	for i, classID := range classIDs {
// 		studyPlanID := studyPlanIDs[i]
// 		enClassStudyPlan, err := ToClassStudyPlanEn(classID, studyPlanID)
// 		if err != nil {
// 			return nil, fmt.Errorf("ToClassStudyPlanEn: %w", err)
// 		}
// 		classStudyPlans = append(classStudyPlans, enClassStudyPlan)
// 	}
// 	return classStudyPlans, nil
// }

func toStudentStudyPlanEn(studentID pgtype.Text, studyPlanID string) (*entities.StudentStudyPlan, error) {
	e := &entities.StudentStudyPlan{}
	database.AllNullEntity(e)
	e.Now()
	e.StudentID = studentID

	err := e.StudyPlanID.Set(studyPlanID)
	return e, err
}

func makeStudentStudyPlanEn(studentIDs pgtype.TextArray, studyPlanIDs []string) ([]*entities.StudentStudyPlan, error) {
	studentStudyPlans := make([]*entities.StudentStudyPlan, 0, len(studentIDs.Elements))
	if len(studentIDs.Elements) != len(studyPlanIDs) {
		return nil, fmt.Errorf("number of student doesn't match number of study plan")
	}
	for i, studentID := range studentIDs.Elements {
		studentStudyPlan, err := toStudentStudyPlanEn(studentID, studyPlanIDs[i])
		if err != nil {
			return nil, fmt.Errorf("ToStudentStudyPlanEn: %w", err)
		}
		studentStudyPlans = append(studentStudyPlans, studentStudyPlan)
	}
	return studentStudyPlans, nil
}

func appendMultiple(ss []string, times int, item string) []string {
	for i := 0; i < times; i++ {
		ss = append(ss, item)
	}
	return ss
}

// IAssignStudyPlan is use for creating and assign study plan to course and student
type IAssignStudyPlan struct {
	StudyPlanRepo interface {
		BulkCopy(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) ([]string, []string, error)
		BulkUpdateBook(ctx context.Context, db database.QueryExecer, spbs []*repositories.StudyPlanBook) error
		FindByIDs(ctx context.Context, db database.QueryExecer, studyPlanID pgtype.TextArray) ([]*entities.StudyPlan, error)
	}
	StudentRepo interface {
		FindStudentsByCourseID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text) (*pgtype.TextArray, error)
	}
	CourseStudyPlanRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, courseStudyPlans []*entities.CourseStudyPlan) error
	}
	StudentStudyPlan interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, studentStudyPlans []*entities.StudentStudyPlan) error
	}
	StudyPlanItemRepo interface {
		BulkCopy(ctx context.Context, db database.QueryExecer, originalStudyPlanIDs pgtype.TextArray, newStudyPlanIDs pgtype.TextArray) error
		BulkInsert(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlanItem) error
		UpdateWithCopiedFromItem(ctx context.Context, db database.QueryExecer, studyPlanItems []*entities.StudyPlanItem) error
	}
	AssignmentStudyPlanItemRepo interface {
		CopyFromStudyPlan(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
		BulkInsert(ctx context.Context, db database.QueryExecer, assignmentStudyPlanItems []*entities.AssignmentStudyPlanItem) error
	}
	LoStudyPlanItemRepo interface {
		CopyFromStudyPlan(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
		BulkInsert(ctx context.Context, db database.QueryExecer, loStudyPlanItems []*entities.LoStudyPlanItem) error
	}
}

func CreateStudyPlanForStudents(ctx context.Context, courseID string, studyPlanID string, studentList *pgtype.TextArray,
	tx pgx.Tx, r *IAssignStudyPlan) error {
	n := len(studentList.Elements)
	orgStudentStudyPlanIDs := make([]string, 0, n)
	orgStudentStudyPlanIDs = appendMultiple(orgStudentStudyPlanIDs, n, studyPlanID)

	masterStudentStudyPlanIDs, createdStudentStudyPlansIDs, err := r.StudyPlanRepo.BulkCopy(ctx, tx, database.TextArray(orgStudentStudyPlanIDs))
	if err != nil {
		return fmt.Errorf("s.StudyPlanRepo.BulkCopy: %w", err)
	}

	studentStudyPlans, err := makeStudentStudyPlanEn(*studentList, createdStudentStudyPlansIDs)
	if err != nil {
		return fmt.Errorf("error convert student study plan: %w", err)
	}

	err = r.StudentStudyPlan.BulkUpsert(ctx, tx, studentStudyPlans)
	if err != nil {
		return fmt.Errorf("s.StudentStudyPlan.BulkUpsert: %w", err)
	}

	// TO-DO copy class study plan item
	err = r.StudyPlanItemRepo.BulkCopy(ctx, tx, database.TextArray(masterStudentStudyPlanIDs), database.TextArray(createdStudentStudyPlansIDs))
	if err != nil {
		return fmt.Errorf("s.StudyPlanItemRepo.BulkCopy: %w", err)
	}

	err = r.AssignmentStudyPlanItemRepo.CopyFromStudyPlan(ctx, tx, database.TextArray(createdStudentStudyPlansIDs))
	if err != nil {
		return fmt.Errorf("s.AssignmentStudyPlanItemRepo.CopyFromStudyPlan: %w", err)
	}
	err = r.LoStudyPlanItemRepo.CopyFromStudyPlan(ctx, tx, database.TextArray(createdStudentStudyPlansIDs))
	if err != nil {
		return fmt.Errorf("s.LoStudyPlanItemRepo.CopyFromStudyPlan: %w", err)
	}
	return nil
}

func HandleAssignCourseStudyPlan(ctx context.Context, courseID string, studyPlanID string,
	tx pgx.Tx, r *IAssignStudyPlan) error {
	studentList, err := r.StudentRepo.FindStudentsByCourseID(ctx, tx, database.Text(courseID))
	if err != nil {
		return fmt.Errorf("s.StudentRepo.FindStudentsByCourseID: %w", err)
	}
	if len(studentList.Elements) == 0 {
		return nil
	}
	errTx := CreateStudyPlanForStudents(ctx, courseID, studyPlanID, studentList, tx, r)
	return errTx
}

func HandleClassStudyPlan(ctx context.Context, classID int32, studyPlanID string) error {
	return fmt.Errorf("not supported")
}

func HandleStudentStudyPlan(ctx context.Context, studentIDS []string, studyPlanID []string, tx database.QueryExecer, r *IAssignStudyPlan) error {
	studentStudyPlans, err := makeStudentStudyPlanEn(database.TextArray(studentIDS), studyPlanID)
	if err != nil {
		return fmt.Errorf("makeStudentStudyPlanEn: %w", err)
	}
	err = r.StudentStudyPlan.BulkUpsert(ctx, tx, studentStudyPlans)
	if err != nil {
		return fmt.Errorf("StudentStudyPlan.BulkUpsert: %w", err)
	}
	return nil
}

func UpsertCourseStudyPlan(ctx context.Context, courseID string, studyPlanID string,
	tx database.QueryExecer, r *IAssignStudyPlan) error {
	csp, err := toCourseStudyPlanEn(courseID, studyPlanID)
	if err != nil {
		return fmt.Errorf("toCourseStudyPlanEn: %w", err)
	}
	err = r.CourseStudyPlanRepo.BulkUpsert(ctx, tx, []*entities.CourseStudyPlan{csp})
	if err != nil {
		return fmt.Errorf("s.CourseStudyPlanRepo.BulkUpsert: %w", err)
	}

	return nil
}

func AssignStudyPlanWithTx(ctx context.Context, req *pb.AssignStudyPlanRequest, tx pgx.Tx, r *IAssignStudyPlan) error {
	switch v := req.Data.(type) {
	case *pb.AssignStudyPlanRequest_CourseId:
		err := UpsertCourseStudyPlan(ctx, v.CourseId, req.StudyPlanId, tx, r)
		if err != nil {
			return fmt.Errorf("UpsertCourseStudyPlan: %w", err)
		}
		err = HandleAssignCourseStudyPlan(ctx, v.CourseId, req.StudyPlanId, tx, r)
		if err != nil {
			return fmt.Errorf("HandleAssignCourseStudyPlan: %w", err)
		}
	case *pb.AssignStudyPlanRequest_ClassId:
		err := HandleClassStudyPlan(ctx, v.ClassId, req.StudyPlanId)
		if err != nil {
			return fmt.Errorf("handleClassStudyPlan: %w", err)
		}
	case *pb.AssignStudyPlanRequest_StudentId:
		err := HandleStudentStudyPlan(ctx, []string{v.StudentId}, []string{req.StudyPlanId}, tx, r)
		if err != nil {
			return fmt.Errorf("handleStudentStudyPlan: %w", err)
		}
	}
	return nil
}

func (s *AssignmentModifierService) AssignStudyPlan(ctx context.Context, req *pb.AssignStudyPlanRequest) (*pb.AssignStudyPlanResponse, error) {
	r := &IAssignStudyPlan{
		StudyPlanRepo:               s.StudyPlanRepo,
		CourseStudyPlanRepo:         s.CourseStudyPlanRepo,
		StudentRepo:                 s.StudentRepo,
		StudentStudyPlan:            s.StudentStudyPlanRepo,
		StudyPlanItemRepo:           s.StudyPlanItemRepo,
		AssignmentStudyPlanItemRepo: s.AssignmentStudyPlanItemRepo,
		LoStudyPlanItemRepo:         s.LoStudyPlanItemRepo,
	}
	err := database.ExecInTxWithRetry(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		return AssignStudyPlanWithTx(ctx, req, tx, r)
	})
	if err != nil {
		return nil, fmt.Errorf("AssignStudyPlan: %w", err)
	}
	return &pb.AssignStudyPlanResponse{}, nil
}

func toAssignmentStudyPlanItems(studyPlanItemID string, assignmentID string) (*entities.AssignmentStudyPlanItem, error) {
	now := timeutil.Now()
	item := &entities.AssignmentStudyPlanItem{}
	database.AllNullEntity(item)
	err := multierr.Combine(
		item.AssignmentID.Set(assignmentID),
		item.StudyPlanItemID.Set(studyPlanItemID),
		item.CreatedAt.Set(now),
		item.UpdatedAt.Set(now),
	)
	return item, err
}

func toLoStudyPlanItems(studyPlanID string, loID string) (*entities.LoStudyPlanItem, error) {
	now := timeutil.Now()
	item := &entities.LoStudyPlanItem{}
	database.AllNullEntity(item)
	err := multierr.Combine(
		item.LoID.Set(loID),
		item.StudyPlanItemID.Set(studyPlanID),
		item.CreatedAt.Set(now),
		item.UpdatedAt.Set(now),
	)
	return item, err
}

func ScheduleStudyPlanWithTx(ctx context.Context, tx pgx.Tx, req *pb.ScheduleStudyPlanRequest,
	insertAssignmentStudyPlanItems func(ctx context.Context, db database.QueryExecer, assignmentStudyPlanItems []*entities.AssignmentStudyPlanItem) error,
	insertLOStudyPlanItems func(ctx context.Context, db database.QueryExecer, assignmentStudyPlanItems []*entities.LoStudyPlanItem) error) error {
	var assignmentStudyPlanItems []*entities.AssignmentStudyPlanItem
	var loStudyPlanItems []*entities.LoStudyPlanItem
	for _, schedule := range req.Schedule {
		switch v := schedule.Item.(type) {
		case *pb.ScheduleStudyPlan_AssignmentId:
			item, err := toAssignmentStudyPlanItems(schedule.StudyPlanItemId, v.AssignmentId)
			if err != nil {
				return fmt.Errorf("s.toAssignmentStudyPlanItems: %w", err)
			}
			assignmentStudyPlanItems = append(assignmentStudyPlanItems, item)
		case *pb.ScheduleStudyPlan_LoId:
			item, err := toLoStudyPlanItems(schedule.StudyPlanItemId, v.LoId)
			if err != nil {
				return fmt.Errorf("s.toLoStudyPlanItems: %w", err)
			}
			loStudyPlanItems = append(loStudyPlanItems, item)
		}
	}

	if len(assignmentStudyPlanItems) > 0 {
		err := insertAssignmentStudyPlanItems(ctx, tx, assignmentStudyPlanItems)
		if err != nil {
			return fmt.Errorf("s.AssignmentStudyPlanItemRepo.BulkInsert: %w", err)
		}
	}

	if len(loStudyPlanItems) > 0 {
		err := insertLOStudyPlanItems(ctx, tx, loStudyPlanItems)
		if err != nil {
			return fmt.Errorf("s.LoStudyPlanItemRepo.BulkInsert: %w", err)
		}
	}
	return nil
}

func (s *AssignmentModifierService) ScheduleStudyPlan(ctx context.Context, req *pb.ScheduleStudyPlanRequest) (*pb.ScheduleStudyPlanResponse, error) {
	err := database.ExecInTxWithRetry(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		return ScheduleStudyPlanWithTx(ctx, tx, req, s.AssignmentStudyPlanItemRepo.BulkInsert, s.LoStudyPlanItemRepo.BulkInsert)
	})
	if err != nil {
		return nil, fmt.Errorf("ScheduleStudyPlan: database.ExecInTxWithRetry: %w", err)
	}

	return &pb.ScheduleStudyPlanResponse{}, nil
}

func toContentStructure(src *pb.ContentStructure) (output *entities.ContentStructure) {
	if src == nil {
		return nil
	}

	output = &entities.ContentStructure{
		BookID:    src.BookId,
		ChapterID: src.ChapterId,
		CourseID:  src.CourseId,
		TopicID:   src.TopicId,
	}
	if src.GetLoId() != nil {
		output.LoID = src.GetLoId().GetValue()
	}
	if src.GetAssignmentId() != nil {
		output.AssignmentID = src.GetAssignmentId().GetValue()
	}

	return output
}

func toContentStructureFlattenLO(cs *pb.ContentStructure, loID string) string {
	// contentStructureFlatten format:
	return fmt.Sprintf("book::%stopic::%schapter::%scourse::%slo::%s", cs.BookId, cs.TopicId, cs.ChapterId, cs.CourseId, loID)
}

func toContentStructureFlattenAssignment(cs *pb.ContentStructure, assignmentID string) string {
	// contentStructureFlatten format:
	return fmt.Sprintf("book::%stopic::%schapter::%scourse::%sassignment::%s", cs.BookId, cs.TopicId, cs.ChapterId, cs.CourseId, assignmentID)
}

func ToStudyPlanItem(src *pb.StudyPlanItem) (*entities.StudyPlanItem, error) {
	dst := &entities.StudyPlanItem{}
	database.AllNullEntity(dst)
	id := src.StudyPlanItemId
	if id == "" {
		id = idutil.ULIDNow()
	}

	startDate := database.TimestamptzFromPb(src.StartDate)
	endDate := database.TimestamptzFromPb(src.EndDate)
	availableFrom := database.TimestamptzFromPb(src.AvailableFrom)
	availableTo := database.TimestamptzFromPb(src.AvailableTo)
	completedAt := database.TimestamptzFromPb(src.CompletedAt)
	schoolDate := database.TimestamptzFromPb(src.SchoolDate)

	contentStructure := toContentStructure(src.ContentStructure)
	now := timeutil.Now()
	dst.StartDate = startDate
	dst.EndDate = endDate
	dst.AvailableFrom = availableFrom
	dst.AvailableTo = availableTo
	dst.CompletedAt = completedAt
	dst.SchoolDate = schoolDate

	if src.ContentStructureFlatten != "" {
		dst.ContentStructureFlatten.Set(src.ContentStructureFlatten)
	}

	setErr := multierr.Combine(
		dst.ID.Set(id),
		dst.StudyPlanID.Set(src.StudyPlanId),
		dst.CreatedAt.Set(now),
		dst.UpdatedAt.Set(now),
		dst.ContentStructure.Set(contentStructure),
		dst.DisplayOrder.Set(src.DisplayOrder),
		dst.Status.Set(src.Status.String()),
	)
	return dst, setErr
}

func UpsertStudyPlanItemWithTx(ctx context.Context, req *pb.UpsertStudyPlanItemRequest, tx pgx.Tx, r *IAssignStudyPlan) ([]string, error) {
	now := time.Now()
	enStudyPlanItems := make([]*entities.StudyPlanItem, 0, len(req.StudyPlanItems))
	studyPlanItemIDs := make([]string, 0, len(req.StudyPlanItems))
	assignmentStudyPlanItems := make([]*entities.AssignmentStudyPlanItem, 0, len(req.StudyPlanItems))
	loStudyPlanItems := make([]*entities.LoStudyPlanItem, 0, len(req.StudyPlanItems))
	studyPlanMap := make(map[string]bool)
	var studyPlanIDs []string
	for _, item := range req.StudyPlanItems {
		if _, ok := studyPlanMap[item.StudyPlanId]; !ok {
			studyPlanMap[item.StudyPlanId] = true
			studyPlanIDs = append(studyPlanIDs, item.StudyPlanId)
		}
		enItem, err := ToStudyPlanItem(item)
		if err != nil {
			return nil, fmt.Errorf("error convert study plan item: %w", err)
		}
		enStudyPlanItems = append(enStudyPlanItems, enItem)
		studyPlanItemIDs = append(studyPlanItemIDs, item.StudyPlanItemId)
		contentStructure := item.ContentStructure
		if contentStructure == nil {
			continue
		}

		if contentStructure.GetAssignmentId() != nil {
			assignmentStudyPlanItems = append(assignmentStudyPlanItems, &entities.AssignmentStudyPlanItem{
				BaseEntity: entities.BaseEntity{
					CreatedAt: database.Timestamptz(now),
					UpdatedAt: database.Timestamptz(now),
					DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
				},
				AssignmentID:    database.Text(contentStructure.GetAssignmentId().GetValue()),
				StudyPlanItemID: enItem.ID,
			})
		}

		if contentStructure.GetLoId() != nil {
			loStudyPlanItems = append(loStudyPlanItems, &entities.LoStudyPlanItem{
				BaseEntity: entities.BaseEntity{
					CreatedAt: database.Timestamptz(now),
					UpdatedAt: database.Timestamptz(now),
					DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
				},
				LoID:            database.Text(contentStructure.GetLoId().GetValue()),
				StudyPlanItemID: enItem.ID,
			})
		}
	}
	studyPlans, err := r.StudyPlanRepo.FindByIDs(ctx, tx, database.TextArray(studyPlanIDs))
	if err != nil {
		return nil, fmt.Errorf("s..StudyPlanRepo.FindByIDs: %w", err)
	}
	studyPlanBookMap := make(map[string]string)
	var studyPlanBooks []*repositories.StudyPlanBook
	for _, studyPlan := range studyPlans {
		studyPlanBookMap[studyPlan.ID.String] = studyPlan.BookID.String
	}
	for _, item := range enStudyPlanItems {
		var content entities.ContentStructure
		item.ContentStructure.AssignTo(&content)
		bookID := studyPlanBookMap[item.ID.String]
		if bookID != content.BookID && bookID != "" {
			switch {
			case content.AssignmentID != "":
				return nil, fmt.Errorf("study plan item of assignment %v has book is wrong, expect %v but got %v", content.AssignmentID, bookID, content.BookID)
			case content.LoID != "":
				return nil, fmt.Errorf("study plan item of lo %v has book is wrong, expect %v but got %v", content.LoID, bookID, content.BookID)
			}
		}
		if bookID == "" {
			studyPlanBooks = append(studyPlanBooks, &repositories.StudyPlanBook{
				StudyPlanID: database.Text(item.StudyPlanID.String),
				BookID:      database.Text(content.BookID),
			})
		}
	}

	if err := r.StudyPlanItemRepo.BulkInsert(ctx, tx, enStudyPlanItems); err != nil {
		return nil, fmt.Errorf("s.StudyPlanItemRepo.BulkInsert: %w", err)
	}

	if err := r.StudyPlanItemRepo.UpdateWithCopiedFromItem(ctx, tx, enStudyPlanItems); err != nil {
		return nil, fmt.Errorf("s.StudyPlanItemRepo.UpdateWithCopiedFromItem: %w", err)
	}

	if len(assignmentStudyPlanItems) != 0 {
		if err := r.AssignmentStudyPlanItemRepo.BulkInsert(ctx, tx, assignmentStudyPlanItems); err != nil {
			return nil, fmt.Errorf("s.AssignmentStudyPlanItemRepo.BulkInsert: %w", err)
		}
	}

	if len(loStudyPlanItems) != 0 {
		if err := r.LoStudyPlanItemRepo.BulkInsert(ctx, tx, loStudyPlanItems); err != nil {
			return nil, fmt.Errorf("s.LoStudyPlanItemRepo.BulkInsert: %w", err)
		}
	}
	if len(studyPlanBooks) > 0 {
		if err := r.StudyPlanRepo.BulkUpdateBook(ctx, tx, studyPlanBooks); err != nil {
			return nil, fmt.Errorf("s.StudyPlanRepo.BulkUpdateBook: %w", err)
		}
	}

	ids := make([]string, 0, len(enStudyPlanItems))
	for _, item := range enStudyPlanItems {
		ids = append(ids, item.ID.String)
	}

	return ids, nil
}

func (s *AssignmentModifierService) UpsertStudyPlanItem(ctx context.Context, req *pb.UpsertStudyPlanItemRequest) (*pb.UpsertStudyPlanItemResponse, error) {
	var ids []string
	r := &IAssignStudyPlan{
		StudyPlanRepo:               s.StudyPlanRepo,
		CourseStudyPlanRepo:         s.CourseStudyPlanRepo,
		StudentRepo:                 s.StudentRepo,
		StudentStudyPlan:            s.StudentStudyPlanRepo,
		StudyPlanItemRepo:           s.StudyPlanItemRepo,
		AssignmentStudyPlanItemRepo: s.AssignmentStudyPlanItemRepo,
		LoStudyPlanItemRepo:         s.LoStudyPlanItemRepo,
	}
	err := database.ExecInTxWithRetry(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		var errTx error
		ids, errTx = UpsertStudyPlanItemWithTx(ctx, req, tx, r)
		if errTx != nil {
			return errTx
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("UpsertStudyPlanItem: %w", err)
	}
	return &pb.UpsertStudyPlanItemResponse{
		StudyPlanItemIds: ids,
	}, nil
}

func (s *AssignmentModifierService) DeleteAssignments(
	ctx context.Context, req *pb.DeleteAssignmentsRequest,
) (*pb.DeleteAssignmentsResponse, error) {
	err := database.ExecInTxWithRetry(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.AssignmentRepo.SoftDelete(ctx, tx, database.TextArray(req.AssignmentIds)); err != nil {
			return fmt.Errorf("s.AssignmentRepo.SoftDelete: %w", err)
		}

		if err := s.TopicsAssignmentsRepo.SoftDeleteByAssignmentIDs(ctx, tx, database.TextArray(req.AssignmentIds)); err != nil {
			return fmt.Errorf("TopicsAssignmentsRepo.SoftDeleteByAssignmentIDs: %w", err)
		}

		studyPlanItemIDs, err := s.AssignmentStudyPlanItemRepo.SoftDeleteByAssigmentIDs(ctx, tx, database.TextArray(req.AssignmentIds))
		if err != nil {
			return fmt.Errorf("AssignmentStudyPlanItemRepo.SoftDeleteByAssigmentIDs: %w", err)
		}

		if err := s.StudyPlanItemRepo.SoftDeleteByStudyPlanItemIDs(ctx, tx, studyPlanItemIDs); err != nil {
			return fmt.Errorf("s.SoftDeleteByStudyPlanItemIDs: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &pb.DeleteAssignmentsResponse{}, nil
}

func (s *AssignmentModifierService) EditAssignmentTime(ctx context.Context, req *pb.EditAssignmentTimeRequest) (*pb.EditAssignmentTimeResponse, error) {
	if len(req.StudyPlanItemIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("AssignmentModifierService.EditAssignmentTime: study plan item ids are empty").Error())
	}

	if req.StudentId == "" {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("AssignmentModifierService.EditAssignmentTime: student id is empty").Error())
	}

	studyPlanItems, err := s.StudyPlanItemRepo.FindByIDs(ctx, s.DB, database.TextArray(req.StudyPlanItemIds))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("AssignmentModifierService.EditAssignmentTime: %w", err).Error())
	}

	for i := range studyPlanItems {
		switch req.UpdateType {
		case pb.UpdateType_UPDATE_END_DATE:
			studyPlanItems[i].EndDate = database.TimestamptzFromPb(req.EndDate)
		case pb.UpdateType_UPDATE_START_DATE:
			studyPlanItems[i].StartDate = database.TimestamptzFromPb(req.StartDate)
		case pb.UpdateType_UPDATE_START_DATE_END_DATE:
			if req.StartDate != nil && req.EndDate != nil {
				if req.StartDate.AsTime().After(req.EndDate.AsTime()) {
					return nil, status.Error(codes.InvalidArgument, fmt.Errorf("AssignmentModifierService.EditAssignmentTime: start date after end date").Error())
				}
			}
			studyPlanItems[i].EndDate = database.TimestamptzFromPb(req.EndDate)
			studyPlanItems[i].StartDate = database.TimestamptzFromPb(req.StartDate)
		}
	}

	if err := s.AssignmentStudyPlanItemRepo.BulkEditAssignmentTime(ctx, s.DB, database.Text(req.StudentId), studyPlanItems); err != nil {
		if err.Error() == "cannot update all study plan items" {
			return nil, status.Error(codes.InvalidArgument, "invalid time")
		}
		return nil, status.Error(codes.Internal, fmt.Errorf("AssignmentModifierService.EditAssignmentTime: %w", err).Error())
	}

	return &pb.EditAssignmentTimeResponse{}, nil
}

func (s *AssignmentModifierService) AssignAssignmentsToTopic(
	ctx context.Context, req *pb.AssignAssignmentsToTopicRequest,
) (*pb.AssignAssignmentsToTopicResponse, error) {
	topicsAssignmentsList := make([]*entities.TopicsAssignments, 0, len(req.GetAssignment()))

	assignmentIDs := make([]string, 0, len(req.GetAssignment()))
	mDisplayOrder := make(map[string]int32)
	for _, assignment := range req.GetAssignment() {
		topicsAssignmentsList = append(topicsAssignmentsList, &entities.TopicsAssignments{
			TopicID:      database.Text(req.TopicId),
			AssignmentID: database.Text(assignment.AssignmentId),
			DisplayOrder: database.Int2(int16(assignment.DisplayOrder)),
		})

		assignmentIDs = append(assignmentIDs, assignment.AssignmentId)
		mDisplayOrder[assignment.AssignmentId] = assignment.DisplayOrder
	}
	if err := s.TopicsAssignmentsRepo.BulkUpsert(ctx, s.DB, topicsAssignmentsList); err != nil {
		return nil, fmt.Errorf("s.TopicsAssignmentsRepo.Upsert: %w", err)
	}

	assignments, err := s.AssignmentRepo.RetrieveAssignments(ctx, s.DB, database.TextArray(assignmentIDs))
	if err != nil {
		return nil, fmt.Errorf("s.AssignmentRepo.RetrieveAssignments: %w", err)
	}

	assignmentsPb := make([]*pb.Assignment, 0, len(assignments))
	for _, assignment := range assignments {
		assignmentPb, err := toAssignmentPb(assignment)
		if err != nil {
			return nil, err
		}
		assignmentPb.DisplayOrder = mDisplayOrder[assignment.ID.String]
		assignmentsPb = append(assignmentsPb, assignmentPb)
	}

	data := &npb.EventAssignmentsCreated{
		Assignments: assignmentsPb,
	}
	msg, _ := proto.Marshal(data)
	_, err = s.JSM.PublishContext(ctx, constants.SubjectAssignmentsCreated, msg)
	if err != nil {
		return nil, fmt.Errorf("s.JSM.PublishContext: subject: %q, %v", constants.SubjectAssignmentsCreated, err)
	}

	return &pb.AssignAssignmentsToTopicResponse{}, nil
}

func (s *AssignmentModifierService) UpsertAssignments(ctx context.Context, req *pb.UpsertAssignmentsRequest) (*pb.UpsertAssignmentsResponse, error) {
	resp, isInserted, err := UpsertAssignmentsWithoutPublishEvent(ctx, s.DB, req, s.TopicRepo, s.AssignmentRepo, s.TopicsAssignmentsRepo)
	if err != nil {
		return nil, err
	}

	if isInserted {
		data := &npb.EventAssignmentsCreated{
			Assignments: req.Assignments,
		}
		msg, _ := proto.Marshal(data)

		if _, err := s.JSM.PublishContext(ctx, constants.SubjectAssignmentsCreated, msg); err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("s.JSM.PublishContext: subject: %q, %v", constants.SubjectAssignmentsCreated, err).Error())
		}
	}

	return &pb.UpsertAssignmentsResponse{
		AssignmentIds: resp.AssignmentIds,
	}, nil
}

func UpsertAssignmentsWithoutPublishEvent(
	ctx context.Context,
	db database.Ext,
	req *pb.UpsertAssignmentsRequest,
	topicRepo ITopicRepository,
	assignmentRepo IAssignmentRepository,
	topicsAssignmentsRepo ITopicsAssignmentsRepository,
) (*pb.UpsertAssignmentsResponse, bool, error) {
	var topicIDs []string
	isInserted := false

	topicMap := make(map[string][]*entities.Assignment)
	ids := make([]string, 0, len(req.Assignments))
	assignmentMap := make(map[string]*pb.Assignment)
	for _, assignment := range req.Assignments {
		if assignment.Name == "" {
			return nil, isInserted, status.Error(codes.InvalidArgument, "empty assignment name")
		}
		en, isAutoGenID, err := toAssignmentEn(assignment)
		if err != nil {
			return nil, isInserted, err
		}

		if isAutoGenID {
			isInserted = true
		}

		assignmentMap[assignment.AssignmentId] = assignment
		ids = append(ids, en.ID.String)
		topicID := assignment.Content.TopicId
		if _, ok := topicMap[topicID]; !ok {
			topicIDs = append(topicIDs, topicID)
		}
		topicMap[topicID] = append(topicMap[topicID], en)
	}

	topics, err := topicRepo.RetrieveByIDs(ctx, db, database.TextArray(topicIDs))
	if err != nil {
		return nil, isInserted, status.Errorf(codes.InvalidArgument, fmt.Errorf("unable to retrieve topics by ids: %w", err).Error())
	}

	if !isTopicsExisted(topicIDs, topics) {
		return nil, isInserted, status.Errorf(codes.InvalidArgument, "some topics does not exists")
	}
	assignmentsExisted, err := assignmentRepo.RetrieveAssignments(
		ctx,
		db,
		database.TextArray(ids),
	)
	if err != nil && err != pgx.ErrNoRows {
		return nil, isInserted, status.Errorf(codes.FailedPrecondition, fmt.Errorf("unable to retrieve assignment by ids: %w", err).Error())
	}

	doMap := make(map[string]int32)
	for _, assignment := range assignmentsExisted {
		doMap[assignment.ID.String] = assignment.DisplayOrder.Int
	}
	if err := database.ExecInTx(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
		for topicID, assignments := range topicMap {
			var (
				insertNum int32
				count     int32
			)
			topic, err := topicRepo.RetrieveByID(ctx, tx, database.Text(topicID), repositories.WithUpdateLock())
			if err != nil {
				return fmt.Errorf("unable to retrieve topic by id: %w", err)
			}
			topicAssignmentEntities := make([]*entities.TopicsAssignments, 0, len(assignments))
			if isAutoGenAssignmentDisplayOrder(assignments) {
				total := topic.LODisplayOrderCounter.Int
				for _, assignment := range assignments {
					if do, ok := doMap[assignment.ID.String]; !ok {
						assignment.DisplayOrder.Set(total + count + 1)
						count++
					} else {
						assignment.DisplayOrder.Set(do)
					}
				}
				insertNum = count
			}
			for _, assignment := range assignments {
				topicAssignmentEntities = append(topicAssignmentEntities, &entities.TopicsAssignments{
					TopicID:      database.Text(topicID),
					AssignmentID: database.Text(assignment.ID.String),
					DisplayOrder: database.Int2(int16(assignment.DisplayOrder.Int)),
				})
				assignmentMap[assignment.ID.String].DisplayOrder = assignment.DisplayOrder.Int
			}
			if err := assignmentRepo.BulkUpsert(ctx, tx, assignments); err != nil {
				return fmt.Errorf("unable to bulk upsert assignment: %w", err)
			}
			// if in future we the feature `link_lo/assignment back, we have to check on topics_assignments too`
			if err := topicsAssignmentsRepo.BulkUpsert(ctx, tx, topicAssignmentEntities); err != nil {
				return fmt.Errorf("unable to bulk upsert topic assignment: %w", err)
			}
			if err := topicRepo.UpdateLODisplayOrderCounter(ctx, tx, database.Text(topicID), database.Int4(insertNum)); err != nil {
				return fmt.Errorf("unable to update lo display order counter: %w", err)
			}

			if err := topicRepo.UpdateTotalLOs(ctx, tx, database.Text(topicID)); err != nil {
				return fmt.Errorf("unable to update total learing objectives: %w", err)
			}
		}
		return nil
	}); err != nil {
		return nil, isInserted, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpsertAssignmentsResponse{
		AssignmentIds: ids,
	}, isInserted, nil
}

func isTopicsExisted(topicIDs []string, topics []*entities.Topic) bool {
	m := make(map[string]bool)
	for _, topic := range topics {
		m[topic.ID.String] = true
	}

	for _, id := range topicIDs {
		if _, ok := m[id]; !ok {
			return false
		}
	}
	return true
}

func isAutoGenAssignmentDisplayOrder(assignments []*entities.Assignment) bool {
	for _, assignment := range assignments {
		if assignment.DisplayOrder.Int != 0 {
			return false
		}
	}
	return true
}

func (s *AssignmentModifierService) UpsertAdHocAssignment(ctx context.Context, req *pb.UpsertAdHocAssignmentRequest) (*pb.UpsertAdHocAssignmentResponse, error) {
	if err := s.verifyUpsertAdHocAssignmentRequest(req); err != nil {
		return nil, err
	}

	cctx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return nil, err
	}
	retrieveStudentProfileResp, err := s.BobStudentReaderSvc.RetrieveStudentProfile(cctx, &bpb.RetrieveStudentProfileRequest{
		StudentIds: []string{req.StudentId},
	})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.StudentReaderSvc.RetrieveStudentProfile: %w", err).Error())
	}

	if retrieveStudentProfileResp == nil || len(retrieveStudentProfileResp.Items) != 1 ||
		retrieveStudentProfileResp.Items[0].Profile == nil || retrieveStudentProfileResp.Items[0].Profile.School == nil {
		return nil, status.Error(codes.Internal, "student must belongs to a school")
	}

	adHocBook, err := s.BookRepo.RetrieveAdHocBookByCourseIDAndStudentID(ctx, s.DB, database.Text(req.CourseId), database.Text(req.StudentId))
	if err != nil && err != pgx.ErrNoRows {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.BookRepo.RetrieveAdHocBookByCourseIDAndStudentID: %w", err).Error())
	}

	var assignmentID string
	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		var topicID string
		// create ad_hoc book and studyPlan when adHoc doesn't exist
		if adHocBook == nil {
			var _err error
			topicID, _err = s.createAdHocBookAndStudyPlan(ctx, tx, retrieveStudentProfileResp, req)
			if _err != nil {
				return _err
			}
		} else { // get topic when adHoc exists
			topics, err := s.TopicRepo.FindByBookIDs(ctx, tx, database.TextArray([]string{adHocBook.ID.String}), pgtype.TextArray{Status: pgtype.Null}, pgtype.Int4{Status: pgtype.Null}, pgtype.Int4{Status: pgtype.Null})
			if err != nil {
				return status.Error(codes.Internal, fmt.Errorf("s.TopicRepo.FindByBookIDs: %w", err).Error())
			}

			if len(topics) == 0 {
				return status.Error(codes.Internal, fmt.Errorf("book (%s) must have a topic", adHocBook.ID.String).Error())
			}
			topicID = topics[0].ID.String
		}

		if req.Assignment.Content == nil {
			req.Assignment.Content = &pb.AssignmentContent{}
		}
		req.Assignment.Content.TopicId = topicID

		upsertAssignmentsReq := &pb.UpsertAssignmentsRequest{
			Assignments: []*pb.Assignment{req.Assignment},
		}

		resp, _, err := UpsertAssignmentsWithoutPublishEvent(ctx, tx, upsertAssignmentsReq, s.TopicRepo, s.AssignmentRepo, s.TopicsAssignmentsRepo)
		if err != nil {
			return status.Error(codes.Internal, fmt.Errorf("UpsertAssignmentsWithoutPublishEvent: %w", err).Error())
		}

		var availableFrom, availableTo, startDate, endDate *time.Time
		if req.StartDate != nil {
			t := req.StartDate.AsTime()
			availableFrom = &t
			startDate = &t
		}
		if req.EndDate != nil {
			t := req.EndDate.AsTime()
			endDate = &t
		}
		t := time.Date(2300, time.January, 1, 23, 59, 0, 0, time.UTC)
		availableTo = &t
		assignmentsWithTimesMap := map[*pb.Assignment]*StudyPlanItemTimes{
			req.Assignment: {
				AvailableFrom: availableFrom,
				AvailableTo:   availableTo,
				StartDate:     startDate,
				EndDate:       endDate,
			},
		}

		if err := HandedStudyPlanItemsWithTimesOnAssignmentsCreated(ctx, tx, assignmentsWithTimesMap,
			&ImportService{
				BookChapterRepo:             &repositories.BookChapterRepo{},
				StudyPlanRepo:               &repositories.StudyPlanRepo{},
				StudyPlanItemRepo:           &repositories.StudyPlanItemRepo{},
				AssignmentStudyPlanItemRepo: &repositories.AssignmentStudyPlanItemRepo{},
			}); err != nil {
			return status.Error(codes.Internal, fmt.Errorf("HandedStudyPlanItemsWithTimesOnAssignmentsCreated: %w", err).Error())
		}

		assignmentID = resp.AssignmentIds[0]

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pb.UpsertAdHocAssignmentResponse{
		AssignmentId: assignmentID,
	}, nil
}

func (s *AssignmentModifierService) createAdHocBookAndStudyPlan(
	ctx context.Context, tx pgx.Tx,
	retrieveStudentProfileResp *bpb.RetrieveStudentProfileResponse,
	req *pb.UpsertAdHocAssignmentRequest) (topicID string, _ error) {
	schoolID := retrieveStudentProfileResp.Items[0].Profile.School.Id
	gradeStr := retrieveStudentProfileResp.Items[0].Profile.Grade
	country := retrieveStudentProfileResp.Items[0].Profile.Country
	studentName := retrieveStudentProfileResp.Items[0].Profile.Name
	grade, err := i18n.ConvertStringGradeToInt(bob_pb.Country(int32(country)), gradeStr)
	if err != nil {
		return "", err
	}

	chapterName := req.ChapterName
	topicName := req.TopicName
	bookName := req.BookName
	studyPlanName := req.StudyPlanName
	if chapterName == "" {
		chapterName = studentName + "'s To-do"
	}
	if topicName == "" {
		topicName = studentName + "'s To-do"
	}
	if bookName == "" {
		bookName = studentName + "'s To-do"
	}
	if studyPlanName == "" {
		studyPlanName = studentName + "'s To-do"
	}

	bookID := idutil.ULIDNow()
	book := &entities.Book{}
	database.AllNullEntity(book)
	if err := multierr.Combine(
		book.ID.Set(bookID),
		book.Name.Set(bookName),
		book.SchoolID.Set(schoolID),
		book.Country.Set(country.String()),
		book.Subject.Set(cpb.Subject_SUBJECT_NONE.String()),
		book.Grade.Set(grade),
		book.BookType.Set(cpb.BookType_BOOK_TYPE_ADHOC.String()),
		book.CurrentChapterDisplayOrder.Set(0),
	); err != nil {
		return "", err
	}

	if err := s.BookRepo.Upsert(ctx, tx, []*entities.Book{book}); err != nil {
		return "", status.Error(codes.Internal, fmt.Errorf("s.BookRepo.Upsert: %w", err).Error())
	}

	chapterID := idutil.ULIDNow()
	chapter := &entities.Chapter{}
	database.AllNullEntity(chapter)
	if err := multierr.Combine(
		chapter.ID.Set(chapterID),
		chapter.Name.Set(chapterName),
		chapter.Country.Set(country.String()),
		chapter.Subject.Set(cpb.Subject_SUBJECT_NONE.String()),
		chapter.Grade.Set(int16(grade)),
		chapter.DisplayOrder.Set(0),
		chapter.SchoolID.Set(schoolID),
	); err != nil {
		return "", err
	}
	if err := s.ChapterRepo.Upsert(ctx, tx, []*entities.Chapter{chapter}); err != nil {
		return "", status.Error(codes.Internal, fmt.Errorf("s.ChapterRepo.Upsert: %w", err).Error())
	}

	bookChapter := &entities.BookChapter{}
	database.AllNullEntity(bookChapter)
	if err := multierr.Combine(
		bookChapter.BookID.Set(bookID),
		bookChapter.ChapterID.Set(chapterID),
	); err != nil {
		return "", err
	}
	if err := s.BookChapterRepo.Upsert(ctx, tx, []*entities.BookChapter{bookChapter}); err != nil {
		return "", status.Error(codes.Internal, fmt.Errorf("s.BookChapterRepo.Upsert: %w", err).Error())
	}

	topicID = idutil.ULIDNow()
	topic := &entities.Topic{}
	database.AllNullEntity(topic)
	if err := multierr.Combine(
		topic.ID.Set(topicID),
		topic.Name.Set(topicName),
		topic.Country.Set(country.String()),
		topic.Grade.Set(grade),
		topic.Subject.Set(cpb.Subject_SUBJECT_NONE.String()),
		topic.TopicType.Set(pb.TopicType_TOPIC_TYPE_NONE.String()),
		topic.Status.Set(pb.TopicStatus_TOPIC_STATUS_NONE.String()),
		topic.DisplayOrder.Set(0),
		topic.ChapterID.Set(chapterID),
		topic.SchoolID.Set(schoolID),
		topic.TotalLOs.Set(0),
		topic.EssayRequired.Set(false),
	); err != nil {
		return "", err
	}
	if err := s.TopicRepo.BulkImport(ctx, tx, []*entities.Topic{topic}); err != nil {
		return "", status.Error(codes.Internal, fmt.Errorf("s.TopicRepo.BulkImport: %w", err).Error())
	}

	courseBook := &entities.CoursesBooks{}
	database.AllNullEntity(courseBook)
	if err := multierr.Combine(
		courseBook.BookID.Set(bookID),
		courseBook.CourseID.Set(req.CourseId),
	); err != nil {
		return "", err
	}
	if err := s.CourseBookRepo.Upsert(ctx, tx, []*entities.CoursesBooks{courseBook}); err != nil {
		return "", status.Error(codes.Internal, fmt.Errorf("s.CourseBookRepo.Upsert: %w", err).Error())
	}

	upsertAdHocIndividualStudyPlanReq := &pb.UpsertAdHocIndividualStudyPlanRequest{
		SchoolId:  schoolID,
		Name:      studyPlanName,
		CourseId:  req.CourseId,
		StudentId: req.StudentId,
		BookId:    bookID,
		Status:    pb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
		Grades:    []int32{int32(grade)},
	}

	if _, err := UpsertAdHocIndividualStudyPlan(ctx, tx, upsertAdHocIndividualStudyPlanReq, &InternalModifierService{
		StudyPlanRepo:               s.StudyPlanRepo,
		CourseBookRepo:              s.CourseBookRepo,
		BookRepo:                    s.BookRepo,
		StudentStudyPlanRepo:        s.StudentStudyPlanRepo,
		AssignmentRepo:              s.AssignmentRepo,
		StudyPlanItemRepo:           s.StudyPlanItemRepo,
		AssignmentStudyPlanItemRepo: s.AssignmentStudyPlanItemRepo,
		LoStudyPlanItemRepo:         s.LoStudyPlanItemRepo,
		LearningObjectiveRepo:       s.LearningObjectiveRepo,
	}); err != nil {
		return topicID, status.Error(codes.Internal, fmt.Errorf("UpsertAdHocIndividualStudyPlan: %w", err).Error())
	}

	return topicID, nil
}

func (s *AssignmentModifierService) verifyUpsertAdHocAssignmentRequest(req *pb.UpsertAdHocAssignmentRequest) error {
	if req.CourseId == "" {
		return status.Errorf(codes.InvalidArgument, "req must have course_id")
	}

	if req.StudentId == "" {
		return status.Errorf(codes.InvalidArgument, "req must have student_id")
	}

	if req.StartDate == nil {
		return status.Errorf(codes.InvalidArgument, "req must have start_date")
	}

	return nil
}

type AssignmentService struct {
	sspb.UnimplementedAssignmentServer
	DB database.Ext

	BookRepo interface {
		RetrieveAdHocBookByCourseIDAndStudentID(ctx context.Context, db database.QueryExecer, courseID, studentID pgtype.Text) (*entities.Book, error)
	}

	TopicRepo interface {
		RetrieveByID(ctx context.Context, db database.QueryExecer, id pgtype.Text, enhancers ...repositories.QueryEnhancer) (*entities.Topic, error)
		UpdateTotalLOs(ctx context.Context, db database.QueryExecer, topicID pgtype.Text) error
		UpdateLODisplayOrderCounter(ctx context.Context, db database.QueryExecer, topicID pgtype.Text, number pgtype.Int4) error
		FindByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs, topicIDs pgtype.TextArray, limit, offset pgtype.Int4) ([]*entities.Topic, error)
	}

	GeneralAssignmentRepo interface {
		Insert(ctx context.Context, db database.QueryExecer, m *entities.GeneralAssignment) error
		Update(ctx context.Context, db database.QueryExecer, m *entities.GeneralAssignment) error
		List(ctx context.Context, db database.QueryExecer, learningMaterialIds pgtype.TextArray) ([]*entities.GeneralAssignment, error)
	}

	AssignmentRepo interface {
		IsStudentAssignedV2(ctx context.Context, db database.QueryExecer, studyPlanID, studentID pgtype.Text) (bool, error)
	}

	SubmissionRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.StudentSubmission) error
	}

	StudentLatestSubmissionRepo interface {
		UpsertV2(ctx context.Context, db database.QueryExecer, e *entities.StudentLatestSubmission) error
	}

	StudentLearningTimeDailyRepo interface {
		Retrieve(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz, queryEnhancers ...repositories.QueryEnhancer) ([]*entities.StudentLearningTimeDaily, error)
		UpsertTaskAssignment(ctx context.Context, db database.QueryExecer, s *entities.StudentLearningTimeDaily) error
	}

	UsermgmtUserReaderService interface {
		SearchBasicProfile(ctx context.Context, in *upb.SearchBasicProfileRequest, opts ...grpc.CallOption) (*upb.SearchBasicProfileResponse, error)
	}
}

func (s *AssignmentService) validateInsertAssignmentReq(req *sspb.InsertAssignmentRequest) error {
	if req.Assignment.Base.LearningMaterialId != "" {
		return status.Error(codes.InvalidArgument, "learning_material_id must be empty")
	}
	if req.Assignment.Base.Type != "" {
		return status.Error(codes.InvalidArgument, "type must be empty")
	}
	if req.Assignment.Base.Name == "" {
		return status.Error(codes.InvalidArgument, "empty assignment name")
	}
	if req.Assignment.Base.TopicId == "" {
		return status.Error(codes.InvalidArgument, "empty topic_id")
	}

	return nil
}

func (s *AssignmentService) InsertAssignment(ctx context.Context, req *sspb.InsertAssignmentRequest) (*sspb.InsertAssignmentResponse, error) {
	if err := s.validateInsertAssignmentReq(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("validateInsertAssignmentReq: %w", err).Error())
	}

	_, err := s.TopicRepo.RetrieveByID(ctx, s.DB, database.Text(req.Assignment.Base.TopicId))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Error(codes.InvalidArgument, fmt.Errorf("topic %s doesn't exists", req.Assignment.Base.TopicId).Error())
		}
		return nil, status.Error(codes.Internal, fmt.Errorf("s.TopicRepo.RetrieveByID: %w", err).Error())
	}

	generalAssignment, err := toGeneralAssignmentEnt(req.Assignment)
	if err != nil {
		return nil, err
	}

	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		topic, err := s.TopicRepo.RetrieveByID(ctx, tx, database.Text(req.Assignment.Base.TopicId), repositories.WithUpdateLock())
		if err != nil {
			return fmt.Errorf("unable to retrieve topic by id: %w", err)
		}
		total := topic.LODisplayOrderCounter.Int
		if err := generalAssignment.LearningMaterial.DisplayOrder.Set(total + 1); err != nil {
			return err
		}

		if err := s.GeneralAssignmentRepo.Insert(ctx, tx, generalAssignment); err != nil {
			return fmt.Errorf("unable to insert general assignment: %w", err)
		}

		if err := s.TopicRepo.UpdateLODisplayOrderCounter(ctx, tx, topic.ID, database.Int4(1)); err != nil {
			return fmt.Errorf("unable to update lo display order counter: %w", err)
		}

		return nil
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &sspb.InsertAssignmentResponse{
		LearningMaterialId: generalAssignment.ID.String,
	}, nil
}

func (s *AssignmentService) validateUpdateAssignmentReq(req *sspb.UpdateAssignmentRequest) error {
	if req.Assignment.Base.LearningMaterialId == "" {
		return status.Error(codes.InvalidArgument, "empty learning_material_id")
	}
	if req.Assignment.Base.Type != "" {
		return status.Error(codes.InvalidArgument, "type must be empty")
	}
	if req.Assignment.Base.TopicId != "" {
		return status.Error(codes.InvalidArgument, "topic_id must be empty")
	}

	return nil
}

func (s *AssignmentService) UpdateAssignment(ctx context.Context, req *sspb.UpdateAssignmentRequest) (*sspb.UpdateAssignmentResponse, error) {
	if err := s.validateUpdateAssignmentReq(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("validateUpdateAssignmentReq: %w", err).Error())
	}

	generalAssignment, err := toGeneralAssignmentEnt(req.Assignment)
	if err != nil {
		return nil, err
	}

	if err := s.GeneralAssignmentRepo.Update(ctx, s.DB, generalAssignment); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("unable to update general assignment: %w", err).Error())
	}

	return &sspb.UpdateAssignmentResponse{}, nil
}
