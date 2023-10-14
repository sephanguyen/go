package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	ent "github.com/manabie-com/backend/internal/eureka/entities"
	entities "github.com/manabie-com/backend/internal/eureka/entities/monitors"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	monitor_repo "github.com/manabie-com/backend/internal/eureka/repositories/monitors"
	services "github.com/manabie-com/backend/internal/eureka/services/monitoring"
	"github.com/manabie-com/backend/internal/golibs/alert"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"go.uber.org/zap"
)

func (s *suite) someStudyPlanItemsNotCreated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	someStudentLo := rand.Intn(len(stepState.LoIDs)) + 1
	stepState.LoItemMissing = stepState.LoIDs[:someStudentLo]
	stepState.AssignmentItemMissing = stepState.AssignmentIDs
	cmd := `UPDATE study_plan_items SET deleted_at = now(), content_structure_flatten = NULL
	 WHERE (content_structure->>'lo_id' = ANY($1::TEXT[]) OR content_structure ->> 'assignment_id' = ANY($2::TEXT[])) AND deleted_at IS NULL`
	_, err := s.DB.Exec(ctx, cmd, database.TextArray(stepState.LoItemMissing), database.TextArray(stepState.AssignmentItemMissing))

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to simulator missing study plan item: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourMonitorAutoUpsertMissingLearningItemCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studyPlanItem := &ent.StudyPlanItem{}
	fieldNames := database.GetFieldNames(studyPlanItem)
	query := fmt.Sprintf(`select %s from study_plan_items spi where 
	spi.study_plan_id = any(select spm.payload ->> 'study_plan_id' from study_plan_monitors spm) 
	and (
			(content_structure ->> 'lo_id'= any(select spm.payload ->>'lo_id' from study_plan_monitors spm))
			or 
			(content_structure ->> 'assignment_id'= any(select spm.payload ->>'assignment_id' from study_plan_monitors spm))
		) 
	and spi.deleted_at is null order by created_at desc limit $1
	`, strings.Join(fieldNames, ", "))

	var studyPlanItems ent.StudyPlanItems
	err := database.Select(
		ctx,
		db,
		query,
		len(stepState.StudyPlanItemMonitors),
	).ScanAll(&studyPlanItems)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(studyPlanItems) != len(stepState.StudyPlanItemMonitors) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("auto upsert missing learning item not correct expected %v got %v items",
			len(stepState.StudyPlanItemMonitors),
			len(studyPlanItems),
		)
	}

	type StudyPlanLearningItem struct {
		StudyPlanID        string
		LearningMaterialID string
	}
	mapStudyPlanMonitors := map[StudyPlanLearningItem]*entities.StudyPlanMonitor{}

	for _, item := range stepState.StudyPlanItemMonitors {
		payload := &entities.StudyPlanMonitorPayload{}
		item.Payload.AssignTo(&payload)
		learningMaterialId := payload.LoID.String
		if payload.AssignmentID.String != "" {
			learningMaterialId = payload.AssignmentID.String
		}

		if _, ok := mapStudyPlanMonitors[StudyPlanLearningItem{
			StudyPlanID:        payload.StudyPlanID.String,
			LearningMaterialID: learningMaterialId,
		}]; !ok {
			mapStudyPlanMonitors[StudyPlanLearningItem{
				StudyPlanID:        payload.StudyPlanID.String,
				LearningMaterialID: learningMaterialId,
			}] = item
		}
	}

	for _, item := range studyPlanItems {
		contentStructure := &ent.ContentStructure{}
		payload := &entities.StudyPlanMonitorPayload{}
		item.ContentStructure.AssignTo(&contentStructure)
		learningMaterialId := contentStructure.LoID
		if contentStructure.AssignmentID != "" {
			learningMaterialId = contentStructure.AssignmentID
		}

		spm, ok := mapStudyPlanMonitors[StudyPlanLearningItem{
			StudyPlanID:        item.StudyPlanID.String,
			LearningMaterialID: learningMaterialId,
		}]
		if !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("auto upsert missing learning item not correct: missing study plan item")
		}

		if (spm.StudentID.String != "" && item.CopyStudyPlanItemID.String == "") ||
			(spm.StudentID.String == "" && item.CopyStudyPlanItemID.String != "") {
			return StepStateToContext(ctx, stepState), fmt.Errorf("auto upsert missing learning item not correct: missing master or individual study plan item")
		}

		spm.Payload.AssignTo(&payload)
		if spm.CourseID.String != contentStructure.CourseID ||
			payload.BookID.String != contentStructure.BookID ||
			payload.ChapterID.String != contentStructure.ChapterID ||
			payload.TopicID.String != contentStructure.TopicID {
			return StepStateToContext(ctx, stepState), fmt.Errorf("auto upsert missing learning item not correct: invalid content structure")
		}
	}

	ctx, items, err := s.getStudyPlanItemMonitor(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("getStudyPlanMonitor: %w", err)
	}

	for _, item := range items {
		if item.AutoUpsertedAt.Status == pgtype.Null {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not mark auto upserted at in study plan monitor")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourMonitorSaveMissingLearningItemCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, items, err := s.getStudyPlanItemMonitor(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("getStudyPlanMonitor: %w", err)
	}
	stepState.StudyPlanItemMonitors = items
	ctx, err = s.getStudyPlanIDsByCourse(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("getStudyPlanIDsByCourse: %w", err)
	}
	if len(items) != (len(stepState.StudentIDs)+1)*(len(stepState.LoItemMissing)+len(stepState.AssignmentItemMissing)) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("system not store correctly amount missing data")
	}
	mapMissingItemID := make(map[string]bool)
	for _, spID := range stepState.StudyPlanIDs {
		mapMissingItemID[spID] = true
	}
	for _, item := range items {
		tempItempContent := entities.StudyPlanMonitorPayload{}
		item.Payload.AssignTo(&tempItempContent)
		if _, ok := mapMissingItemID[tempItempContent.StudyPlanID.String]; !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("store wrong missing item")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getStudyPlanItemMonitor(ctx context.Context) (context.Context, []*entities.StudyPlanMonitor, error) {
	stepState := StepStateFromContext(ctx)
	cmd := `SELECT %s FROM %s WHERE type = $2::TEXT AND (payload->>'lo_id' = ANY($1::TEXT[]) OR payload->>'assignment_id' = ANY($3::TEXT[]))`
	var e entities.StudyPlanMonitor
	selectFields := database.GetFieldNames(&e)
	var items entities.StudyPlanMonitors

	if err := database.Select(ctx, s.DB, fmt.Sprintf(cmd, strings.Join(selectFields, ","), e.TableName()),
		database.TextArray(stepState.LoItemMissing),
		database.Text(entities.StudyPlanMonitorType_STUDY_PLAN_ITEM),
		database.TextArray(stepState.AssignmentItemMissing)).ScanAll(&items); err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}

	return StepStateToContext(ctx, stepState), items, nil
}

func (s *suite) getStudyPlanIDsByCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	cmd := `SELECT %s FROM %s WHERE course_id = $1::TEXT AND deleted_at IS NULL`
	var e ent.StudyPlan
	selectFields := database.GetFieldNames(&e)
	var items ent.StudyPlans
	err := database.Select(ctx, s.DB, fmt.Sprintf(cmd, strings.Join(selectFields, ","), e.TableName()), database.Text(stepState.CourseID)).ScanAll(&items)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	for _, sp := range items {
		stepState.StudyPlanIDs = append(stepState.StudyPlanIDs, sp.ID.String)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) runMonitorUpsertLearningItem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedCtx(ctx)

	// TODO: consider add to config
	// check the resource path work or not-- like the config in local, in current Im not adding a lot config
	// for db RLS query
	ctx = auth.InjectFakeJwtToken(ctx, stepState.SchoolID)
	studentStudyPlanRepo := &repositories.StudentStudyPlanRepo{}
	courseStudentRepo := &repositories.CourseStudentRepo{}
	studyPlanRepo := &repositories.StudyPlanRepo{}
	studyPlanMonitorRepo := &monitor_repo.StudyPlanMonitorRepo{}
	assignmentRepo := &repositories.AssignmentRepo{}
	studyPlanItemRepo := &repositories.StudyPlanItemRepo{}
	learningObjectiveRepo := &repositories.LearningObjectiveRepo{}
	loStudyPlanItemRepo := &repositories.LoStudyPlanItemRepo{}
	assignmentStudyPlanItemRepo := &repositories.AssignmentStudyPlanItemRepo{}
	c := &configurations.Config{
		SyllabusSlackWebHook: "https://hooks.slack.com/services/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",
		SchoolInformation: configurations.SchoolInfoConfig{
			SchoolID:   stepState.SchoolID,
			SchoolName: "Local shool",
		},
		Common: configs.CommonConfig{
			Environment: "local",
		},
	}

	httpClient := http.Client{Timeout: time.Duration(10) * time.Second}
	alertClient := &alert.SlackImpl{
		WebHookURL: "https://hooks.slack.com/services/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",
		HTTPClient: httpClient,
	}

	studyPlanMonitorService := &services.StudyPlanMonitorService{
		DB:                          s.DB,
		Cfg:                         c,
		Logger:                      *zap.NewNop(),
		Alert:                       alertClient,
		StudentStudyPlanRepo:        studentStudyPlanRepo,
		CourseStudentRepo:           courseStudentRepo,
		StudyPlanRepo:               studyPlanRepo,
		StudyPlanMonitorRepo:        studyPlanMonitorRepo,
		AssignmentRepo:              assignmentRepo,
		StudyPlanItemRepo:           studyPlanItemRepo,
		LearningObjectiveRepo:       learningObjectiveRepo,
		LoStudyPlanItemRepo:         loStudyPlanItemRepo,
		AssignmentStudyPlanItemRepo: assignmentStudyPlanItemRepo,
	}
	err := studyPlanMonitorService.UpsertLearningItems(ctx, 1, stepState.SchoolID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to monitor upsert course student: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}
