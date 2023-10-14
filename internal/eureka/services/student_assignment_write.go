package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/notification/consts"
	consumer "github.com/manabie-com/backend/internal/notification/transports/nats"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// StudentAssignmentWriteService implementation
type StudentAssignmentWriteService struct {
	DB  database.Ext
	JSM nats.JetStreamManagement

	StudentLearningTimeDailyRepo interface {
		UpsertTaskAssignment(ctx context.Context, db database.QueryExecer, s *entities.StudentLearningTimeDaily) error
		Retrieve(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz, queryEnhancers ...repositories.QueryEnhancer) ([]*entities.StudentLearningTimeDaily, error)
	}

	SubmissionRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.StudentSubmission) error
		UpdateGradeStatus(ctx context.Context, db database.QueryExecer,
			id, gradeID, userChangeStatusID, status pgtype.Text) error
		BulkUpdateStatus(ctx context.Context, db database.QueryExecer, editorID, status pgtype.Text, grades []*entities.StudentSubmissionGrade) error
		FindBySubmissionIDs(ctx context.Context, db database.QueryExecer, submissionIDs pgtype.TextArray) (*entities.StudentSubmissions, error)
	}

	StudentLatestSubmissionRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, e *entities.StudentLatestSubmission) error
		BulkUpserts(ctx context.Context, db database.QueryExecer, StudentLatestSubmission []*entities.StudentLatestSubmission) error
	}

	SubmissionGradeRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.StudentSubmissionGrade) error
		FindBySubmissionIDs(ctx context.Context, db database.QueryExecer, submissionIDs pgtype.TextArray) (*entities.StudentSubmissionGrades, error)
		BulkImport(ctx context.Context, db database.QueryExecer, grades []*entities.StudentSubmissionGrade) error
		RetrieveInfoByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*repositories.StudentSubmissionGradeInfo, error)
	}

	StudyPlanItemRepo interface {
		MarkItemCompleted(context.Context, database.QueryExecer, pgtype.Text) error
	}
	StudentReaderClient interface {
		RetrieveStudentProfile(ctx context.Context, in *bpb.RetrieveStudentProfileRequest, opts ...grpc.CallOption) (*bpb.RetrieveStudentProfileResponse, error)
	}

	UsermgmtUserReaderService interface {
		SearchBasicProfile(ctx context.Context, in *upb.SearchBasicProfileRequest, opts ...grpc.CallOption) (*upb.SearchBasicProfileResponse, error)
	}
}

// SubmitAssignment is a gRPC method
func (s *StudentAssignmentWriteService) SubmitAssignment(ctx context.Context, req *pb.SubmitAssignmentRequest) (*pb.SubmitAssignmentResponse, error) {
	submission := submitAssignmentRequestToEnt(ctx, req)

	err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.calculateAssignmentLearningTime(ctx, tx, &entities.StudentAssignmentLearningTime{
			StudentID:    submission.StudentID.String,
			AssignmentID: req.GetSubmission().GetAssignmentId(),
			CompleteDate: req.GetSubmission().GetCompleteDate(),
			Duration:     req.GetSubmission().GetDuration(),
		}); err != nil {
			return err
		}

		if err := multierr.Combine(
			s.SubmissionRepo.Create(ctx, tx, submission),
			s.StudyPlanItemRepo.MarkItemCompleted(ctx, tx, submission.StudyPlanItemID),
			s.StudentLatestSubmissionRepo.Upsert(ctx, tx, &entities.StudentLatestSubmission{StudentSubmission: *submission}),
		); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("database.ExecInTx: %w", err)
	}
	return &pb.SubmitAssignmentResponse{
		SubmissionId: submission.ID.String,
	}, nil
}

// GradeStudentSubmission is a gRPC method
func (s *StudentAssignmentWriteService) GradeStudentSubmission(ctx context.Context, req *pb.GradeStudentSubmissionRequest) (*pb.GradeStudentSubmissionResponse, error) {
	grade, err := reqGradeToEnt(ctx, req)
	if err != nil {
		return nil, err
	}

	err = database.ExecInTxWithRetry(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.SubmissionGradeRepo.Create(ctx, tx, grade); err != nil {
			return fmt.Errorf("SubmissionGradeRepo.Create: %w", err)
		}

		if err := s.SubmissionRepo.UpdateGradeStatus(ctx, tx,
			grade.StudentSubmissionID,
			grade.ID,
			grade.GraderID,
			database.Text(req.Status.String())); err != nil {
			return fmt.Errorf("SubmissionRepo.UpdateGradeStatus: %w", err)
		}
		if err := s.updateStudentLatestSubmission(ctx, tx, []string{grade.StudentSubmissionID.String}); err != nil {
			return err
		}
		if req.Status == pb.SubmissionStatus_SUBMISSION_STATUS_RETURNED {
			s.publishGradeStudentSubmissions(ctx, tx, []*entities.StudentSubmissionGrade{grade})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &pb.GradeStudentSubmissionResponse{
		SubmissionGradeId: grade.ID.String,
	}, nil
}

func reqGradeToEnt(ctx context.Context, req *pb.GradeStudentSubmissionRequest) (*entities.StudentSubmissionGrade, error) {
	if req.Grade == nil {
		return nil, status.Error(codes.InvalidArgument, "Grade cannot be null")
	}

	graderID := interceptors.UserIDFromContext(ctx)
	e := &entities.StudentSubmissionGrade{}
	database.AllNullEntity(e)

	e.StudentSubmissionID = database.Text(req.Grade.SubmissionId)
	e.GraderID = database.Text(graderID)
	e.GraderComment = database.Text(req.Grade.Note)
	e.Status = database.Text(req.Status.String())

	grade := math.Round(req.Grade.Grade*100) / 100
	if err := e.Grade.Set(grade); err != nil {
		return nil, err
	}
	if req.Grade.GradeContent != nil {
		e.GradeContent.Set(req.Grade.GradeContent)
	}

	return e, nil
}

// UpdateStudentSubmissionsStatus is a gRPC method
func (s *StudentAssignmentWriteService) UpdateStudentSubmissionsStatus(ctx context.Context, req *pb.UpdateStudentSubmissionsStatusRequest) (*pb.UpdateStudentSubmissionsStatusResponse, error) {
	submissionIds := golibs.Uniq(req.SubmissionIds)
	userChangeStatusID := interceptors.UserIDFromContext(ctx)
	err := database.ExecInTxWithRetry(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		esRaw, err := s.SubmissionGradeRepo.FindBySubmissionIDs(ctx, tx, database.TextArray(submissionIds))
		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("SubmissionGradeRepo.FindBySubmissionIDs: %w", err)
		}
		es := make([]*entities.StudentSubmissionGrade, 0, len(submissionIds))
		es = append(es, *esRaw...)

		grades := defaultGrades(es, submissionIds, userChangeStatusID, req.Status.String())
		err = s.SubmissionGradeRepo.BulkImport(ctx, tx, grades)
		if err != nil {
			return fmt.Errorf("SubmissionGradeRepo.BulkImport: %w", err)
		}
		if err := s.SubmissionRepo.BulkUpdateStatus(ctx, tx, database.Text(userChangeStatusID), database.Text(req.Status.String()), grades); err != nil {
			return fmt.Errorf("SubmissionRepo.BulkUpdateStatus: %w", err)
		}
		if err := s.updateStudentLatestSubmission(ctx, tx, submissionIds); err != nil {
			return err
		}
		if req.Status == pb.SubmissionStatus_SUBMISSION_STATUS_RETURNED {
			s.publishGradeStudentSubmissions(ctx, tx, es)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("database.ExecInTx: %w", err)
	}

	return &pb.UpdateStudentSubmissionsStatusResponse{
		Successfully: true,
	}, nil
}

func submitAssignmentRequestToEnt(ctx context.Context, req *pb.SubmitAssignmentRequest) *entities.StudentSubmission {
	e := &entities.StudentSubmission{}
	database.AllNullEntity(e)

	e.StudyPlanItemID = database.Text(req.Submission.StudyPlanItemId)
	e.AssignmentID = database.Text(req.Submission.AssignmentId)
	if req.Submission.SubmissionContent != nil {
		e.SubmissionContent.Set(req.Submission.SubmissionContent)
	}
	e.Note = database.Text(req.Submission.Note)
	e.Status = database.Text(pb.SubmissionStatus_SUBMISSION_STATUS_NOT_MARKED.String())

	groupID := interceptors.UserGroupFromContext(ctx)

	switch groupID {
	case constant.RoleStudent:
		e.StudentID = database.Text(interceptors.UserIDFromContext(ctx))
	case constant.RoleTeacher:
		e.StudentID = database.Text(req.Submission.StudentId)
	}

	if req.Submission.CompleteDate != nil && req.Submission.CompleteDate.IsValid() {
		e.CompleteDate = database.TimestamptzFromPb(req.Submission.CompleteDate)
	} else {
		e.CompleteDate = database.Timestamptz(time.Now())
	}

	e.Duration = database.Int4(req.Submission.Duration)
	e.CorrectScore = database.Float4(req.Submission.CorrectScore)
	e.TotalScore = database.Float4(req.Submission.TotalScore)
	e.UnderstandingLevel = database.Text(req.Submission.UnderstandingLevel.String())

	return e
}

// defaultGrades create from old student_submission_grades record with user change status id and status given
func defaultGrades(es []*entities.StudentSubmissionGrade, submissionIds []string, userChangeStatusID, status string) []*entities.StudentSubmissionGrade {
	// init temp grades with some existed grades
	tmp := append([]*entities.StudentSubmissionGrade{}, es...)
	// some submissions haven't graded
	for _, id := range submissionIds {
		if !contains(es, id) {
			e := defaultGrade(id)
			tmp = append(tmp, e)
		}
	}
	// set status and editor for all grades
	for _, e := range tmp {
		e.Status = database.Text(status)
		e.EditorID = database.Text(userChangeStatusID)
	}

	return tmp
}

func contains(s []*entities.StudentSubmissionGrade, target string) bool {
	for _, val := range s {
		if target == val.StudentSubmissionID.String {
			return true
		}
	}
	return false
}

func defaultGrade(submissionID string) *entities.StudentSubmissionGrade {
	ent := &entities.StudentSubmissionGrade{}
	database.AllNullEntity(ent)
	ent.StudentSubmissionID = database.Text(submissionID)

	// -1 is a default grade when the teacher haven't grade submission
	_ = ent.Grade.Set(float64(-1))
	return ent
}

func (s *StudentAssignmentWriteService) updateStudentLatestSubmission(ctx context.Context, db database.Ext, submissionIDs []string) error {
	submissions, err := s.SubmissionRepo.FindBySubmissionIDs(ctx, db, database.TextArray(submissionIDs))
	if err != nil {
		return fmt.Errorf("SubmissionRepo.FindBySubmissionIDs: %w", err)
	}
	latestSubmissions := make([]*entities.StudentLatestSubmission, len(*submissions))

	for i, submission := range *submissions {
		latestSubmissions[i] = &entities.StudentLatestSubmission{StudentSubmission: *submission}
	}

	err = s.StudentLatestSubmissionRepo.BulkUpserts(ctx, db, latestSubmissions)
	if err != nil {
		return fmt.Errorf("StudentLatestSubmissionRepo.BulkUpserts: %w", err)
	}
	return nil
}

func (s *StudentAssignmentWriteService) publishGradeStudentSubmissions(ctx context.Context, db database.Ext, es []*entities.StudentSubmissionGrade) {
	var (
		studentIDs []string
	)
	logger := ctxzap.Extract(ctx)
	studentMap := make(map[string]bool)
	ssgMap := make(map[string]*repositories.StudentSubmissionGradeInfo)
	ssgIDs := make([]string, 0, len(es))
	for _, ssg := range es {
		ssgIDs = append(ssgIDs, ssg.ID.String)
	}
	infos, err := s.SubmissionGradeRepo.RetrieveInfoByIDs(ctx, db, database.TextArray(ssgIDs))
	if err != nil {
		logger.Error("publishGradeStudentSubmissions: unable to retrieve info by ids", zap.Error(err))
		return
	}
	for _, info := range infos {
		if ok := studentMap[info.StudentID.String]; !ok {
			studentMap[info.StudentID.String] = true
			studentIDs = append(studentIDs, info.StudentID.String)
		}
		ssgMap[info.StudentSubmissionGradeID.String] = info
	}

	mdCtx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		logger.Error("publishGradeStudentSubmissions: interceptors.GetOutgoingContext:", zap.Error(err))
		return
	}
	resp, err := s.StudentReaderClient.RetrieveStudentProfile(mdCtx, &bpb.RetrieveStudentProfileRequest{
		StudentIds: studentIDs,
	})
	if err != nil {
		logger.Error("unable to retrieve student profile", zap.Error(err))
		return
	}
	studentInfoMap := make(map[string]*bpb.StudentProfile)
	for _, profile := range resp.Items {
		studentInfoMap[profile.Profile.Id] = profile.Profile
	}
	var (
		wg         sync.WaitGroup
		publishErr error
	)
	for _, info := range infos {
		student, ok := studentInfoMap[info.StudentID.String]
		if !ok {
			logger.Error("student not exists in map", zap.String("studentID", info.StudentID.String))
			continue
		}
		wg.Add(1)
		go func(info *repositories.StudentSubmissionGradeInfo, student *bpb.StudentProfile) {
			defer wg.Done()
			customData := map[string]any{
				"study_plan_item_id": info.StudyPlanItemID.String,
				"course_id":          info.CourseID.String,
				"assignment_id":      info.AssignmentID.String,
				"student_plan_item_identity": map[string]string{
					"student_id":           info.StudentID.String,
					"study_plan_id":        info.StudyPlanID.String,
					"learning_material_id": info.LearningMaterialID.String,
				},
			}
			stringData, _ := json.Marshal(&customData)
			var (
				content string
				title   string
				message string
			)
			switch student.Country {
			case cpb.Country_COUNTRY_JP:
				title = "課題が返却されました"
				content = fmt.Sprintf("課題 <b>%s</b> が返却されました", info.AssignmentName.String)
				message = fmt.Sprintf("課題 %s が返却されました", info.AssignmentName.String)
			default:
				title = "Assignment returned"
				content = fmt.Sprintf("You have received grading for your assignment <b>%s</b>.", info.AssignmentName.String)
				message = fmt.Sprintf("You have received grading for your assignment %s.", info.AssignmentName.String)
			}

			notification := &ypb.NatsCreateNotificationRequest{
				ClientId:       "syllabus_assigment_client_id",
				SendingMethods: []string{consts.SendingMethodPushNotification},
				Target: &ypb.NatsNotificationTarget{
					ReceivedUserIds: []string{info.StudentID.String},
				},
				TargetGroup: &ypb.NatsNotificationTargetGroup{
					UserGroupFilter: &ypb.NatsNotificationTargetGroup_UserGroupFilter{
						UserGroups: []string{consts.TargetUserGroupStudent},
					},
				},
				NotificationConfig: &ypb.NatsPushNotificationConfig{
					Mode:             consts.NotificationModeNotify,
					PermanentStorage: true,
					Notification: &ypb.NatsNotification{
						Title:   title,
						Message: message,
						Content: content,
					},
					Data: map[string]string{
						"custom_data_type":  "assignment_returned",
						"custom_data_value": string(stringData),
					},
				},
				SendTime: &ypb.NatsNotificationSendTime{
					Type: consts.NotificationTypeImmediate,
				},
				TracingId: uuid.New().String(),
			}

			data, _ := proto.Marshal(notification)
			if _, err := s.JSM.PublishContext(ctx, consumer.SubjectNotificationCreated, data); err != nil {
				publishErr = multierr.Append(publishErr, err)
			}
		}(info, student)
	}
	wg.Wait()
	if publishErr != nil {
		logger.Error("publishGradeStudentSubmissions: JSM.PublishContext:", zap.Error(publishErr))
	}
}

func (s *StudentAssignmentWriteService) calculateAssignmentLearningTime(ctx context.Context, tx database.QueryExecer, req *entities.StudentAssignmentLearningTime) error {
	country, err := s.getStudentCountry(ctx, req.StudentID)
	if err != nil {
		return err
	}

	if req.Duration != 0 && req.CompleteDate != nil {
		day := timeutil.MidnightIn(bob_pb.Country(bob_pb.Country_value[country]), req.CompleteDate.AsTime())

		pgDay := database.Timestamptz(day.UTC())
		var assignmentSubmissionIDs []string

		dailies, err := s.StudentLearningTimeDailyRepo.Retrieve(ctx, tx, database.Text(req.StudentID), &pgDay, &pgDay, repositories.WithUpdateLock())
		if err != nil {
			return fmt.Errorf("Retrieve: %w", err)
		}

		if len(dailies) > 0 {
			assignmentSubmissionIDs = database.FromTextArray(dailies[0].AssignmentSubmissionIDs)
		}

		if req.AssignmentID != "" && !golibs.InArrayString(req.AssignmentID, assignmentSubmissionIDs) {
			assignmentSubmissionIDs = append(assignmentSubmissionIDs, req.AssignmentID)
		}

		studentLearningTimeDaily := &entities.StudentLearningTimeDaily{
			StudentID:               database.Text(req.StudentID),
			LearningTime:            database.Int4(req.Duration),
			AssignmentLearningTime:  database.Int4(req.Duration),
			AssignmentSubmissionIDs: database.TextArray(assignmentSubmissionIDs),
			Day:                     pgDay, // DB always store UTC time
		}
		if err := s.StudentLearningTimeDailyRepo.UpsertTaskAssignment(ctx, tx, studentLearningTimeDaily); err != nil {
			return fmt.Errorf("Upsert: %w", err)
		}
	}
	return nil
}

func (s *StudentAssignmentWriteService) getStudentCountry(ctx context.Context, studentID string) (string, error) {
	upbReq := &upb.SearchBasicProfileRequest{
		UserIds: []string{studentID},
		Paging:  &cpb.Paging{Limit: uint32(1)},
	}

	mdCtx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return "", status.Errorf(codes.Unauthenticated, "GetOutgoingContext: %v", err)
	}
	resp, err := s.UsermgmtUserReaderService.SearchBasicProfile(mdCtx, upbReq)
	if err != nil {
		return "", status.Errorf(codes.Internal, "UsermgmtUserReaderService.SearchBasicProfile: %v", err)
	}

	if len(resp.Profiles) == 0 {
		return "", status.Errorf(codes.NotFound, "UsermgmtUserReaderService.SearchBasicProfile: user %s not found", studentID)
	}

	return resp.Profiles[0].Country.String(), nil
}
