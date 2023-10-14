package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type CourseStudentPackageService struct {
	DB                database.Ext
	CourseStudentRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.CourseStudent) (map[repositories.CourseStudentKey]string, error)
		BulkUpsertV2(ctx context.Context, db database.QueryExecer, items []*entities.CourseStudent) error
		SoftDelete(ctx context.Context, db database.QueryExecer, studentIDs, courseIDs []string) error
		SoftDeleteByStudentID(ctx context.Context, db database.QueryExecer, studentID string) error
		SoftDeleteByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) error
		GetByCourseStudents(ctx context.Context, db database.QueryExecer, courseStudents entities.CourseStudents) (entities.CourseStudents, error)
	}
	CourseStudentAccessPathRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.CourseStudentsAccessPath) error
		DeleteLatestCourseStudentAccessPathsByCourseStudentIDs(ctx context.Context, db database.QueryExecer, courseStudentIDs pgtype.TextArray) error
	}
}

type CourseStudentService struct {
	DB          database.Ext
	JSM         nats.JetStreamManagement
	Logger      *zap.Logger
	StudentRepo interface {
		FindStudentsByCourseID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text) (*pgtype.TextArray, error)
	}
	CourseStudentRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.CourseStudent) (map[repositories.CourseStudentKey]string, error)
		BulkUpsertV2(ctx context.Context, db database.QueryExecer, items []*entities.CourseStudent) error
		SoftDelete(ctx context.Context, db database.QueryExecer, studentIDs, courseIDs []string) error
		SoftDeleteByStudentID(ctx context.Context, db database.QueryExecer, studentID string) error
		SoftDeleteByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) error
		GetByCourseStudents(ctx context.Context, db database.QueryExecer, courseStudents entities.CourseStudents) (entities.CourseStudents, error)
	}

	CourseStudentAccessPathRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.CourseStudentsAccessPath) error
		DeleteLatestCourseStudentAccessPathsByCourseStudentIDs(ctx context.Context, db database.QueryExecer, courseStudentIDs pgtype.TextArray) error
	}
	StudentStudyPlanRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, studentStudyPlans []*entities.StudentStudyPlan) error
		FindStudentStudyPlanWithCourseIDs(ctx context.Context, db database.QueryExecer, studentIDs, courseIDs []string) ([]string, error)
		SoftDelete(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
		FindByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) ([]string, error)
		FindAllStudentStudyPlan(ctx context.Context, db database.QueryExecer, masterStudentStudyPlanIDs pgtype.TextArray, studentID pgtype.Text) ([]*entities.StudyPlan, error)
	}
	StudyPlanRepo interface {
		FindByIDs(ctx context.Context, db database.QueryExecer, studyPlanID pgtype.TextArray) ([]*entities.StudyPlan, error)
		SoftDelete(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) error
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlan) error
		BulkCopy(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) ([]string, []string, error)
		BulkUpdateBook(ctx context.Context, db database.QueryExecer, spbs []*repositories.StudyPlanBook) error
	}
	StudyPlanItemRepo interface {
		BulkInsert(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlanItem) error
		SoftDeleteWithStudyPlanIDs(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
		BulkCopy(ctx context.Context, db database.QueryExecer, originalStudyPlanIDs pgtype.TextArray, newStudyPlanIDs pgtype.TextArray) error
		UpdateWithCopiedFromItem(ctx context.Context, db database.QueryExecer, studyPlanItems []*entities.StudyPlanItem) error
	}
	CourseStudyPlanRepo interface {
		FindByCourseIDs(ctx context.Context, db database.QueryExecer, courseIDs pgtype.TextArray) ([]*entities.CourseStudyPlan, error)
		BulkUpsert(ctx context.Context, db database.QueryExecer, courseStudyPlans []*entities.CourseStudyPlan) error
	}
	StudentStudyPlan interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, studentStudyPlans []*entities.StudentStudyPlan) error
	}
	AssignmentStudyPlanItemRepo interface {
		BulkInsert(ctx context.Context, db database.QueryExecer, assignmentStudyPlanItems []*entities.AssignmentStudyPlanItem) error
		CopyFromStudyPlan(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
	}
	LoStudyPlanItemRepo interface {
		BulkInsert(ctx context.Context, db database.QueryExecer, assignmentStudyPlanItems []*entities.LoStudyPlanItem) error
		CopyFromStudyPlan(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
	}
}

func getCourseStudentFromReq(req *npb.EventSyncStudentPackage_StudentPackage) ([]*entities.CourseStudent, error) {
	courseStudents := make([]*entities.CourseStudent, 0, len(req.Packages))
	courseStds := make(map[string]bool)
	for index := 0; index < len(req.Packages); index++ {
		value := req.Packages[index]
		for _, courseID := range value.CourseIds {
			if isExist := courseStds[fmt.Sprintf("%s-%s", req.StudentId, courseID)]; !isExist {
				courseStds[fmt.Sprintf("%s-%s", req.StudentId, courseID)] = true
				courseStudent := &entities.CourseStudent{}
				database.AllNullEntity(courseStudent)
				err := multierr.Combine(
					courseStudent.ID.Set(idutil.ULIDNow()),
					courseStudent.StudentID.Set(req.StudentId),
					courseStudent.CourseID.Set(courseID),
					courseStudent.CreatedAt.Set(timeutil.Now()),
					courseStudent.UpdatedAt.Set(timeutil.Now()),
					courseStudent.StartAt.Set(value.StartDate.AsTime()),
					courseStudent.EndAt.Set(value.EndDate.AsTime()),
				)
				if err != nil {
					return nil, fmt.Errorf("err set CourseStudent: %w", err)
				}
				courseStudents = append(courseStudents, courseStudent)
			}
		}
	}
	return courseStudents, nil
}

func (s *CourseStudentService) upsertCourseStudent(ctx context.Context, db database.Ext, courseStudents []*entities.CourseStudent) (map[repositories.CourseStudentKey]string, error) {
	if len(courseStudents) == 0 {
		return nil, nil
	}

	courseStudentMap, err := s.CourseStudentRepo.BulkUpsert(ctx, db, courseStudents)
	if err != nil {
		return nil, fmt.Errorf("err s.CourseStudentRepo.BulkUpsert: %w", err)
	}

	return courseStudentMap, nil
}

func retrieveStudentIDsFromCourseStudents(css []*entities.CourseStudent) []string {
	res := make([]string, 0, len(css))
	for _, cs := range css {
		res = append(res, cs.StudentID.String)
	}
	return golibs.GetUniqueElementStringArray(res)
}

func retrieveStudentIDsFromEventSyncStudentPackage(evt *npb.EventSyncStudentPackage) []string {
	res := make([]string, 0)
	for _, e := range evt.GetStudentPackages() {
		res = append(res, e.GetStudentId())
	}
	return golibs.GetUniqueElementStringArray(res)
}

func (s *CourseStudentService) removeStudyPlan(ctx context.Context, db database.QueryExecer, studyPlanIDs []string) error {
	err := s.StudentStudyPlanRepo.SoftDelete(ctx, db, database.TextArray(studyPlanIDs))
	if err != nil {
		return fmt.Errorf("s.StudentStudyPlanRepo.SoftDelete: %w", err)
	}
	err = s.StudyPlanRepo.SoftDelete(ctx, db, database.TextArray(studyPlanIDs))
	if err != nil {
		return fmt.Errorf("StudyPlanRepo.SoftDelete: %w", err)
	}
	return err
}

func (s *CourseStudentService) softDeleteStudentStudyPlanByCourseStudent(ctx context.Context, db database.QueryExecer, studentIDs, courseIDs []string) error {
	if len(courseIDs) == 0 || len(studentIDs) == 0 {
		return nil
	}

	studentStudyPlanIDs, err := s.StudentStudyPlanRepo.FindStudentStudyPlanWithCourseIDs(ctx, db, studentIDs, courseIDs)
	if err != nil {
		return fmt.Errorf("s.StudentStudyPlanRepo.FindStudentStudyPlan: %w", err)
	}

	if len(studentStudyPlanIDs) == 0 {
		return nil
	}

	err = s.removeStudyPlan(ctx, db, studentStudyPlanIDs)
	if err != nil {
		return err
	}

	return nil
}

func (s *CourseStudentService) softDeleteStudentStudyPlanByStudent(ctx context.Context, db database.QueryExecer, studentIDs []string) error {
	if len(studentIDs) == 0 {
		return nil
	}

	studentStudyPlanIDs, err := s.StudentStudyPlanRepo.FindByStudentIDs(ctx, db, database.TextArray(studentIDs))
	if err != nil {
		return fmt.Errorf("s.StudentStudyPlanRepo.FindByStudentIDs: %w", err)
	}

	if len(studentStudyPlanIDs) == 0 {
		return nil
	}

	err = s.removeStudyPlan(ctx, db, studentStudyPlanIDs)
	if err != nil {
		return err
	}
	return nil
}

func (s *CourseStudentService) softDeleteCourseStudent(ctx context.Context, db database.QueryExecer, courseStudents []*entities.CourseStudent) error {
	if len(courseStudents) == 0 {
		return nil
	}

	studentIds := make([]string, 0, len(courseStudents))
	courseIds := make([]string, 0, len(courseStudents))
	for _, courseStu := range courseStudents {
		item := courseStu
		studentIds = append(studentIds, item.StudentID.String)
		courseIds = append(courseIds, item.CourseID.String)
	}

	err := s.CourseStudentRepo.SoftDelete(ctx, db, studentIds, courseIds)
	if err != nil {
		return fmt.Errorf("err s.CourseStudentRepo.SoftDelete: %w", err)
	}
	return nil
}

func (s *CourseStudentService) upsertStudyPlanForStudent(ctx context.Context, studentID string, courseIDs []string, tx pgx.Tx) error {
	courseStudyPlans, err := s.CourseStudyPlanRepo.FindByCourseIDs(ctx, tx, database.TextArray(courseIDs))
	if err != nil {
		return fmt.Errorf("s.CourseStudyPlanRepo.FindByCourseIDs: %w", err)
	}

	studyPlanIDs := make([]string, 0, len(courseStudyPlans))
	for _, csp := range courseStudyPlans {
		studyPlanIDs = append(studyPlanIDs, csp.StudyPlanID.String)
	}
	studyPlans, err := s.StudentStudyPlanRepo.FindAllStudentStudyPlan(ctx, tx, database.TextArray(studyPlanIDs), database.Text(studentID))
	if err != nil {
		return fmt.Errorf("s.StudentStudyPlanRepo.FindAllStudentStudyPlan: %w", err)
	}

	masterStudyPlanMap := make(map[string]*entities.StudyPlan)
	for _, studyPlan := range studyPlans {
		masterStudyPlanMap[studyPlan.MasterStudyPlan.String] = studyPlan
	}
	r := &IAssignStudyPlan{
		StudyPlanRepo:               s.StudyPlanRepo,
		CourseStudyPlanRepo:         s.CourseStudyPlanRepo,
		StudentRepo:                 s.StudentRepo,
		StudentStudyPlan:            s.StudentStudyPlan,
		StudyPlanItemRepo:           s.StudyPlanItemRepo,
		AssignmentStudyPlanItemRepo: s.AssignmentStudyPlanItemRepo,
		LoStudyPlanItemRepo:         s.LoStudyPlanItemRepo,
	}

	upsertStudyPlans := make([]*entities.StudyPlan, 0, len(courseStudyPlans))
	studentStudyPlans := make([]*entities.StudentStudyPlan, 0, len(courseStudyPlans))
	for _, csp := range courseStudyPlans {
		studyPlan, ok := masterStudyPlanMap[csp.StudyPlanID.String]
		if !ok {
			// Create student study plan if not exist in course
			studentList := database.TextArray([]string{studentID})
			err = CreateStudyPlanForStudents(ctx, csp.CourseID.String, csp.StudyPlanID.String, &studentList, tx, r)
			if err != nil {
				return err
			}
			continue
		}
		upsertStudyPlans = append(upsertStudyPlans, studyPlan)
		studentStudyPlan := &entities.StudentStudyPlan{}
		database.AllNullEntity(studentStudyPlan)
		studentStudyPlan.StudentID.Set(studentID)
		studentStudyPlan.StudyPlanID.Set(studyPlan.ID)
		studentStudyPlan.UpdatedAt.Set(timeutil.Now())
		studentStudyPlan.CreatedAt.Set(timeutil.Now())
		studentStudyPlans = append(studentStudyPlans, studentStudyPlan)
	}

	// Upsert student study plan again
	err = s.StudentStudyPlanRepo.BulkUpsert(ctx, tx, studentStudyPlans)
	if err != nil {
		return err
	}

	err = s.StudyPlanRepo.BulkUpsert(ctx, tx, upsertStudyPlans)
	if err != nil {
		return err
	}

	return nil
}

func (s *CourseStudentService) upsertStudyPlanForCourseStudent(ctx context.Context, tx pgx.Tx, courseStudents []*entities.CourseStudent) error {
	studentCourseMap := make(map[string][]string)
	for _, cs := range courseStudents {
		studentCourseMap[cs.StudentID.String] = append(studentCourseMap[cs.StudentID.String], cs.CourseID.String)
	}
	//TODO: optimize with go-routine
	for studentID, courseIDs := range studentCourseMap {
		err := s.upsertStudyPlanForStudent(ctx, studentID, courseIDs, tx)
		if err != nil {
			return err
		}
	}
	return nil
}

// SyncCourseStudent handle EventSyncStudentPackage event, upsert CourseStudent if ActionKind=UPSERTED and softDelete if ActionKind=DELETED.
func (s *CourseStudentService) SyncCourseStudent(ctx context.Context, req *npb.EventSyncStudentPackage) error {
	err := database.ExecInTxWithRetry(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		for _, request := range req.StudentPackages {
			courseStudents, err := getCourseStudentFromReq(request)
			if err != nil {
				return err
			}

			studentIDs := make([]string, 0, len(courseStudents))
			courseIDs := make([]string, 0, len(courseStudents))
			for _, courseStu := range courseStudents {
				item := courseStu
				studentIDs = append(studentIDs, item.StudentID.String)
				courseIDs = append(courseIDs, item.CourseID.String)
			}
			switch request.ActionKind {
			case npb.ActionKind_ACTION_KIND_UPSERTED:
				err = s.CourseStudentRepo.SoftDeleteByStudentID(ctx, tx, request.StudentId)
				if err != nil {
					return fmt.Errorf("err soft delete student course of studentID %v: %w", request.StudentId, err)
				}

				if err := s.softDeleteStudentStudyPlanByStudent(ctx, tx, []string{request.StudentId}); err != nil {
					return fmt.Errorf("error removing student study plan %w", err)
				}

				if len(courseStudents) != 0 {
					// upsert course student
					if _, err := s.upsertCourseStudent(ctx, tx, courseStudents); err != nil {
						return fmt.Errorf("err upsert student course of studentID %v: %w", request.StudentId, err)
					}

					if err := s.upsertStudyPlanForCourseStudent(ctx, tx, courseStudents); err != nil {
						return fmt.Errorf("error create new study plan for course student: %w", err)
					}
				}
			case npb.ActionKind_ACTION_KIND_DELETED:
				// softdelete course student
				if err := s.softDeleteCourseStudent(ctx, tx, courseStudents); err != nil {
					return fmt.Errorf("err soft delete student course of studentID %v: %w", request.StudentId, err)
				}
				if err := s.softDeleteStudentStudyPlanByCourseStudent(ctx, tx, studentIDs, courseIDs); err != nil {
					return fmt.Errorf("error removing student study plan %w", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	// publish to course student evt
	event := &npb.EventCourseStudent{
		StudentIds: retrieveStudentIDsFromEventSyncStudentPackage(req),
	}
	data, err := proto.Marshal(event)
	if err != nil {
		s.Logger.Warn("upsertCourseStudent: error to marshal", zap.Error(err))
	}

	_, err = s.JSM.PublishAsyncContext(ctx, constants.SubjectCourseStudentEventNats, data)
	if err != nil {
		s.Logger.Warn("upsertCourseStudent: error to PublishAsync", zap.Error(err))
	}
	return nil
}

type courseStudentSyncInfo struct {
	courseStudent []*entities.CourseStudent
	studentIDs    []string
	courseIDs     []string
}
type syncCourseStudentMap map[npb.ActionKind]courseStudentSyncInfo

func constructMapSyncCourseStudent(req *npb.EventSyncStudentPackage) (syncCourseStudentMap, error) {
	result := syncCourseStudentMap{}
	upsertStudentIDs := []string{}
	upsertCourseStudents := []*entities.CourseStudent{}
	softDeleteStudentIDs := []string{}
	softDeleteCourseIDs := []string{}

	for _, request := range req.StudentPackages {
		courseStudents, err := getCourseStudentFromReq(request)
		if err != nil {
			return nil, err
		}

		switch request.ActionKind {
		case npb.ActionKind_ACTION_KIND_UPSERTED:
			upsertStudentIDs = append(upsertStudentIDs, request.StudentId)

			if len(courseStudents) == 0 {
				return nil, fmt.Errorf("constructMapSyncCourseStudent: Upsert Action Kind expect exist courseStudents")
			}

			upsertCourseStudents = append(upsertCourseStudents, courseStudents...)
		case npb.ActionKind_ACTION_KIND_DELETED:
			if len(courseStudents) == 0 {
				return nil, fmt.Errorf("constructMapSyncCourseStudent: Delete Action Kind expect exist courseStudents")
			}

			for _, courseStu := range courseStudents {
				softDeleteStudentIDs = append(softDeleteStudentIDs, courseStu.StudentID.String)
				softDeleteCourseIDs = append(softDeleteCourseIDs, courseStu.CourseID.String)
			}
		}
	}

	if len(upsertCourseStudents) != 0 && len(upsertStudentIDs) != 0 {
		result[npb.ActionKind_ACTION_KIND_UPSERTED] = courseStudentSyncInfo{
			studentIDs:    upsertStudentIDs,
			courseStudent: upsertCourseStudents,
		}
	}

	if len(softDeleteStudentIDs) != 0 && len(softDeleteCourseIDs) != 0 {
		result[npb.ActionKind_ACTION_KIND_DELETED] = courseStudentSyncInfo{
			studentIDs: softDeleteStudentIDs,
			courseIDs:  softDeleteCourseIDs,
		}
	}

	return result, nil
}

func ProcessSyncCourseStudent(ctx context.Context, db database.Ext, req *npb.EventSyncStudentPackage, repo interface {
	BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.CourseStudent) (map[repositories.CourseStudentKey]string, error)
	BulkUpsertV2(ctx context.Context, db database.QueryExecer, items []*entities.CourseStudent) error
	SoftDelete(ctx context.Context, db database.QueryExecer, studentIDs, courseIDs []string) error
	SoftDeleteByStudentID(ctx context.Context, db database.QueryExecer, studentID string) error
	SoftDeleteByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) error
	GetByCourseStudents(ctx context.Context, db database.QueryExecer, courseStudents entities.CourseStudents) (entities.CourseStudents, error)
},
) error {
	mapSyncCourseStudent, err := constructMapSyncCourseStudent(req)
	if err != nil {
		return err
	}

	if val, ok := mapSyncCourseStudent[npb.ActionKind_ACTION_KIND_UPSERTED]; ok {
		if err := database.ExecInTx(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
			err = repo.SoftDeleteByStudentIDs(ctx, tx, val.studentIDs)
			if err != nil {
				return fmt.Errorf("err SyncCourseStudent with Upsert action: %w", err)
			}
			err = repo.BulkUpsertV2(ctx, tx, val.courseStudent)
			if err != nil {
				return fmt.Errorf("err SyncCourseStudent with Upsert action: %w", err)
			}

			return nil
		}); err != nil {
			return err
		}
	}

	if val, ok := mapSyncCourseStudent[npb.ActionKind_ACTION_KIND_DELETED]; ok {
		if err := database.ExecInTx(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
			err = repo.SoftDelete(ctx, tx, val.studentIDs, val.courseIDs)
			if err != nil {
				return fmt.Errorf("err SyncCourseStudent with Deleted action: %w, studentIDs: %s, courseID: %s", err, val.studentIDs, val.courseIDs)
			}
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

// SyncCourseStudentV2 handle EventSyncStudentPackage event, upsert CourseStudent if ActionKind=UPSERTED and softDelete if ActionKind=DELETED.
// TODO: Subcribe when sql generate is ready
func (s *CourseStudentService) SyncCourseStudentV2(ctx context.Context, req *npb.EventSyncStudentPackage) error {
	if err := ProcessSyncCourseStudent(ctx, s.DB, req, s.CourseStudentRepo); err != nil {
		return err
	}
	// publish to course student evt
	event := &npb.EventCourseStudent{
		StudentIds: retrieveStudentIDsFromEventSyncStudentPackage(req),
	}
	data, err := proto.Marshal(event)
	if err != nil {
		s.Logger.Warn("upsertCourseStudent: error to marshal", zap.Error(err))
	}

	_, err = s.JSM.PublishAsyncContext(ctx, constants.SubjectCourseStudentEventNats, data)
	if err != nil {
		s.Logger.Warn("upsertCourseStudent: error to PublishAsync", zap.Error(err))
	}
	return nil
}

func (s *CourseStudentService) HandleStudentPackageEvent(ctx context.Context, req *npb.EventStudentPackage) error {
	s.Logger.Info("HandleStudentPackageEvent", zap.Any("req", req))

	courseStudents, courseStudentAccessPaths, err := convStudentPackageToCourseStudents(req)
	if err != nil {
		return fmt.Errorf("convStudentPackageToCourseStudents error: %w", err)
	}
	if !req.StudentPackage.IsActive {
		err = database.ExecInTxWithRetry(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
			if err := s.softDeleteCourseStudent(ctx, tx, courseStudents); err != nil {
				return fmt.Errorf("s.CourseStudentRepo.SoftDeleteByStudentID: %w", err)
			}
			if err := s.softDeleteStudentStudyPlanByStudent(ctx, tx, []string{req.StudentPackage.StudentId}); err != nil {
				return fmt.Errorf("s.softDeleteStudentStudyPlanByCourseStudent: %w", err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("database.ExecInTxWithRetry: %w", err)
		}

		return nil
	}

	err = database.ExecInTxWithRetry(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		courseStudentMap, err := s.upsertCourseStudent(ctx, tx, courseStudents)
		if err != nil {
			return fmt.Errorf("upsertCourseStudent: %w", err)
		}

		if len(courseStudentAccessPaths) != 0 {
			courseStudentIDs := make([]string, 0, len(courseStudentAccessPaths))
			for _, courseStudentAccessPath := range courseStudentAccessPaths {
				key := repositories.CourseStudentKey{
					CourseID:  courseStudentAccessPath.CourseID.String,
					StudentID: courseStudentAccessPath.StudentID.String,
				}

				if courseStudentID, ok := courseStudentMap[key]; ok {
					courseStudentAccessPath.CourseStudentID.Set(courseStudentID)
				}

				courseStudentIDs = append(courseStudentIDs, courseStudentAccessPath.CourseStudentID.String)
			}

			if err := s.CourseStudentAccessPathRepo.DeleteLatestCourseStudentAccessPathsByCourseStudentIDs(ctx, tx, database.TextArray(courseStudentIDs)); err != nil {
				return fmt.Errorf("s.CourseStudentAccessPathRepo.DeleteLatestCourseStudentAccessPathsByCourseStudentIDs: %w", err)
			}

			if err := s.CourseStudentAccessPathRepo.BulkUpsert(ctx, tx, courseStudentAccessPaths); err != nil {
				return fmt.Errorf("s.CourseStudentAccessPathRepo.BulkUpsert: %w", err)
			}
		}
		// upsert course study plans
		if err := s.upsertStudyPlanForStudent(ctx, req.StudentPackage.StudentId, req.StudentPackage.Package.CourseIds, tx); err != nil {
			return fmt.Errorf("upsertStudyPlanForStudent: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("database.ExecInTxWithRetry: %w", err)
	}

	// publish to course student evt
	event := &npb.EventCourseStudent{
		StudentIds: retrieveStudentIDsFromCourseStudents(courseStudents),
	}
	data, err := proto.Marshal(event)
	if err != nil {
		s.Logger.Warn("upsertCourseStudent: error to marshal", zap.Error(err))
	}

	_, err = s.JSM.PublishAsyncContext(ctx, constants.SubjectCourseStudentEventNats, data)
	if err != nil {
		s.Logger.Warn("upsertCourseStudent: error to PublishAsync", zap.Error(err))
	}

	return nil
}

func (s *CourseStudentService) HandleStudentPackageEventV2(ctx context.Context, req *npb.EventStudentPackageV2) error {
	s.Logger.Info("HandleStudentPackageEventV2", zap.Any("req", req))

	courseStudent, _, courseStudentAccessPath, err := convStudentPackageToCourseStudentsClassStudentV2(req)
	if err != nil {
		return fmt.Errorf("convStudentPackageToCourseStudentsV2 error: %w", err)
	}
	if !req.StudentPackage.IsActive {
		err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
			if err := s.softDeleteCourseStudent(ctx, tx, []*entities.CourseStudent{courseStudent}); err != nil {
				return fmt.Errorf("s.softDeleteCourseStudent: %w", err)
			}
			if err := s.softDeleteStudentStudyPlanByStudent(ctx, tx, []string{req.StudentPackage.StudentId}); err != nil {
				return fmt.Errorf("s.softDeleteStudentStudyPlanByCourseStudent: %w", err)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("database.ExecInTx: %w", err)
		}

		return nil
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		courseStudentMap, err := s.upsertCourseStudent(ctx, tx, []*entities.CourseStudent{courseStudent})
		if err != nil {
			return fmt.Errorf("upsertCourseStudent: %w", err)
		}

		if courseStudentAccessPath != nil {
			key := repositories.CourseStudentKey{
				CourseID:  courseStudentAccessPath.CourseID.String,
				StudentID: courseStudentAccessPath.StudentID.String,
			}

			if courseStudentID, ok := courseStudentMap[key]; ok {
				courseStudentAccessPath.CourseStudentID.Set(courseStudentID)
			}

			courseStudentID := courseStudentAccessPath.CourseStudentID.String

			if err := s.CourseStudentAccessPathRepo.DeleteLatestCourseStudentAccessPathsByCourseStudentIDs(ctx, tx, database.TextArray([]string{courseStudentID})); err != nil {
				return fmt.Errorf("s.CourseStudentAccessPathRepo.DeleteLatestCourseStudentAccessPathsByCourseStudentIDs: %w", err)
			}
		}

		if courseStudentAccessPath != nil {
			if err := s.CourseStudentAccessPathRepo.BulkUpsert(ctx, tx, []*entities.CourseStudentsAccessPath{courseStudentAccessPath}); err != nil {
				return fmt.Errorf("s.CourseStudentAccessPathRepo.BulkUpsert: %w", err)
			}
		}
		// upsert course study plans
		if err := s.upsertStudyPlanForStudent(ctx, req.StudentPackage.StudentId, []string{req.StudentPackage.Package.CourseId}, tx); err != nil {
			return fmt.Errorf("upsertStudyPlanForStudent: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("database.ExecInTx: %w", err)
	}

	// publish to course student evt
	event := &npb.EventCourseStudent{
		StudentIds: retrieveStudentIDsFromCourseStudents([]*entities.CourseStudent{courseStudent}),
	}
	data, err := proto.Marshal(event)
	if err != nil {
		s.Logger.Warn("HandleStudentPackageEventV2: error to marshal", zap.Error(err))
	}

	_, err = s.JSM.PublishAsyncContext(ctx, constants.SubjectCourseStudentEventNats, data)
	if err != nil {
		s.Logger.Warn("HandleStudentPackageEventV2: error to PublishAsync", zap.Error(err))
	}

	return nil
}

func convStudentPackageToCourseStudents(req *npb.EventStudentPackage) (_ []*entities.CourseStudent, courseStudentAccessPaths []*entities.CourseStudentsAccessPath, _ error) {
	if req == nil || req.GetStudentPackage() == nil || req.GetStudentPackage().GetPackage() == nil || req.GetStudentPackage().GetPackage().GetCourseIds() == nil {
		return nil, nil, fmt.Errorf("empty request")
	}
	courseStudents := make([]*entities.CourseStudent, 0, len(req.StudentPackage.Package.CourseIds))

	for _, courseID := range req.StudentPackage.Package.CourseIds {
		courseStudentID := idutil.ULIDNow()
		courseStudent := &entities.CourseStudent{}
		database.AllNullEntity(courseStudent)
		err := multierr.Combine(
			courseStudent.ID.Set(courseStudentID),
			courseStudent.StudentID.Set(req.StudentPackage.StudentId),
			courseStudent.CourseID.Set(courseID),
			courseStudent.CreatedAt.Set(timeutil.Now()),
			courseStudent.UpdatedAt.Set(timeutil.Now()),
			courseStudent.StartAt.Set(req.StudentPackage.Package.StartDate.AsTime()),
			courseStudent.EndAt.Set(req.StudentPackage.Package.EndDate.AsTime()),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("err set CourseStudent: %w", err)
		}
		courseStudents = append(courseStudents, courseStudent)

		for _, locationID := range req.StudentPackage.Package.LocationIds {
			courseStudentAccessPath := &entities.CourseStudentsAccessPath{}
			database.AllNullEntity(courseStudentAccessPath)
			err := multierr.Combine(
				courseStudentAccessPath.CourseStudentID.Set(courseStudentID),
				courseStudentAccessPath.StudentID.Set(req.StudentPackage.StudentId),
				courseStudentAccessPath.CourseID.Set(courseID),
				courseStudentAccessPath.LocationID.Set(locationID),
				courseStudentAccessPath.CreatedAt.Set(timeutil.Now()),
				courseStudentAccessPath.UpdatedAt.Set(timeutil.Now()),
			)
			if err != nil {
				return nil, nil, fmt.Errorf("err set CourseStudentAccessPath: %w", err)
			}
			courseStudentAccessPaths = append(courseStudentAccessPaths, courseStudentAccessPath)
		}
	}
	return courseStudents, courseStudentAccessPaths, nil
}

func convStudentPackageToCourseStudentsClassStudentV2(req *npb.EventStudentPackageV2) (*entities.CourseStudent, *entities.ClassStudent, *entities.CourseStudentsAccessPath, error) {
	if req == nil || req.GetStudentPackage() == nil || req.GetStudentPackage().GetPackage() == nil || len(req.GetStudentPackage().GetPackage().GetCourseId()) == 0 {
		return nil, nil, nil, fmt.Errorf("empty request")
	}

	courseStudent := &entities.CourseStudent{}
	database.AllNullEntity(courseStudent)

	err := multierr.Combine(
		courseStudent.ID.Set(idutil.ULIDNow()),
		courseStudent.StudentID.Set(req.StudentPackage.StudentId),
		courseStudent.CourseID.Set(req.StudentPackage.Package.CourseId),
		courseStudent.CreatedAt.Set(timeutil.Now()),
		courseStudent.UpdatedAt.Set(timeutil.Now()),
		courseStudent.StartAt.Set(req.StudentPackage.Package.StartDate.AsTime()),
		courseStudent.EndAt.Set(req.StudentPackage.Package.EndDate.AsTime()),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("err set CourseStudent: %w", err)
	}

	var classStudent *entities.ClassStudent

	if len(req.StudentPackage.Package.ClassId) != 0 {
		classStudent = &entities.ClassStudent{}
		database.AllNullEntity(classStudent)

		err = multierr.Combine(
			classStudent.StudentID.Set(req.StudentPackage.StudentId),
			classStudent.ClassID.Set(req.StudentPackage.Package.ClassId),
			classStudent.CreatedAt.Set(timeutil.Now()),
			classStudent.UpdatedAt.Set(timeutil.Now()),
		)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("err set ClassStudent: %w", err)
		}
	}

	var courseStudentAccessPath *entities.CourseStudentsAccessPath
	if len(req.StudentPackage.Package.LocationId) != 0 {
		courseStudentAccessPath = &entities.CourseStudentsAccessPath{}
		database.AllNullEntity(courseStudentAccessPath)
		err = multierr.Combine(
			courseStudentAccessPath.CourseStudentID.Set(courseStudent.ID),
			courseStudentAccessPath.StudentID.Set(req.StudentPackage.StudentId),
			courseStudentAccessPath.CourseID.Set(req.StudentPackage.Package.CourseId),
			courseStudentAccessPath.LocationID.Set(req.StudentPackage.Package.LocationId),
			courseStudentAccessPath.CreatedAt.Set(timeutil.Now()),
			courseStudentAccessPath.UpdatedAt.Set(timeutil.Now()),
		)
	}

	if err != nil {
		return nil, nil, nil, fmt.Errorf("err set CourseStudentAccessPath: %w", err)
	}
	return courseStudent, classStudent, courseStudentAccessPath, nil
}

func (s *CourseStudentPackageService) ProcessHandleStudentPackageEvent(ctx context.Context, req *npb.EventStudentPackageV2) (*entities.CourseStudent, error) {
	courseStudent, _, courseStudentAccessPath, err := convStudentPackageToCourseStudentsClassStudentV2(req)
	if err != nil {
		return nil, err
	}
	if !req.StudentPackage.IsActive {
		err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
			err := s.CourseStudentRepo.SoftDelete(ctx, tx, []string{courseStudent.StudentID.String}, []string{courseStudent.CourseID.String})
			// Will be remove after move all syllabus to use view study plan
			if err != nil {
				return fmt.Errorf("err s.CourseStudentRepo.SoftDelete: %w", err)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("database.ExecInTx: %w", err)
		}

		return courseStudent, nil
	} else if req.StudentPackage.IsActive {
		err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
			courseStudentMap, err := s.CourseStudentRepo.BulkUpsert(ctx, tx, []*entities.CourseStudent{courseStudent})
			if err != nil {
				return fmt.Errorf("err s.CourseStudentRepo.BulkUpsert: %w", err)
			}

			if courseStudentAccessPath != nil {
				key := repositories.CourseStudentKey{
					CourseID:  courseStudentAccessPath.CourseID.String,
					StudentID: courseStudentAccessPath.StudentID.String,
				}

				if courseStudentID, ok := courseStudentMap[key]; ok {
					courseStudentAccessPath.CourseStudentID.Set(courseStudentID)
				}

				courseStudentID := courseStudentAccessPath.CourseStudentID.String

				if err := s.CourseStudentAccessPathRepo.DeleteLatestCourseStudentAccessPathsByCourseStudentIDs(ctx, tx, database.TextArray([]string{courseStudentID})); err != nil {
					return fmt.Errorf("s.CourseStudentAccessPathRepo.DeleteLatestCourseStudentAccessPathsByCourseStudentIDs: %w", err)
				}
				if err := s.CourseStudentAccessPathRepo.BulkUpsert(ctx, tx, []*entities.CourseStudentsAccessPath{courseStudentAccessPath}); err != nil {
					return fmt.Errorf("s.CourseStudentAccessPathRepo.BulkUpsert: %w", err)
				}
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("database.ExecInTx: %w", err)
		}
	}
	return courseStudent, nil
}

func (s *CourseStudentService) HandleStudentPackageEventV3(ctx context.Context, req *npb.EventStudentPackageV2) error {
	s.Logger.Info("HandleStudentPackageEventV3", zap.Any("req", req))

	cspService := CourseStudentPackageService{
		DB:                          s.DB,
		CourseStudentRepo:           s.CourseStudentRepo,
		CourseStudentAccessPathRepo: s.CourseStudentAccessPathRepo,
	}
	courseStudent, err := cspService.ProcessHandleStudentPackageEvent(ctx, req)
	if err != nil {
		return err
	}

	// publish to course student evt
	event := &npb.EventCourseStudent{
		StudentIds: retrieveStudentIDsFromCourseStudents([]*entities.CourseStudent{courseStudent}),
	}
	data, err := proto.Marshal(event)
	if err != nil {
		s.Logger.Warn("HandleStudentPackageEventV3: error to marshal", zap.Error(err))
	}

	_, err = s.JSM.PublishAsyncContext(ctx, constants.SubjectCourseStudentEventNats, data)
	if err != nil {
		s.Logger.Warn("HandleStudentPackageEventV3: error to PublishAsync", zap.Error(err))
	}

	return nil
}
