package syllabus

import (
	"github.com/manabie-com/backend/features/syllabus/allocate_marker"
	"github.com/manabie-com/backend/features/syllabus/course_statistical"
	"github.com/manabie-com/backend/features/syllabus/individual_study_plan"
	"github.com/manabie-com/backend/features/syllabus/learning_history_data_sync"
	"github.com/manabie-com/backend/features/syllabus/nat_sync"
	student_event_logs "github.com/manabie-com/backend/features/syllabus/student_event_logs"
	"github.com/manabie-com/backend/features/syllabus/student_progress"
	"github.com/manabie-com/backend/features/syllabus/study_plan"
	"github.com/manabie-com/backend/features/syllabus/study_plan_item"
	"github.com/manabie-com/backend/features/syllabus/utils"
)

func initStudyPlanStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{}
	topicStatisticalStep := course_statistical.InitStep((*course_statistical.Suite)(utils.NewEntitySuite(s.StepState.TopicStatisticalState, s.Connections, s.ZapLogger, s.Cfg, s.AuthHelper)))
	individualStudyPlanStep := individual_study_plan.InitStep(((*individual_study_plan.Suite)(utils.NewEntitySuite(s.StepState.IndividualStudyPlanStepState, s.Connections, s.ZapLogger, s.Cfg, s.AuthHelper))))
	masterStudyPlanStep := study_plan.InitStep((*study_plan.Suite)(utils.NewEntitySuite(s.StepState.StudyPlanStepState, s.Connections, s.ZapLogger, s.Cfg, s.AuthHelper)))
	studyPlanItemStep := study_plan_item.InitStep((*study_plan_item.Suite)(utils.NewEntitySuite(s.StepState.StudyPlanItemStepState, s.Connections, s.ZapLogger, s.Cfg, s.AuthHelper)))
	studentEventLogsStep := student_event_logs.InitStep((*student_event_logs.Suite)(utils.NewEntitySuite(s.StepState.StudentEventLogsStepState, s.Connections, s.ZapLogger, s.Cfg, s.AuthHelper)))
	studentProgressStep := student_progress.InitStep((*student_progress.Suite)(utils.NewEntitySuite(s.StepState.StudentProgressStepState, s.Connections, s.ZapLogger, s.Cfg, s.AuthHelper)))
	jprepSyncStep := nat_sync.InitStep((*nat_sync.Suite)(utils.NewEntitySuite(s.StepState.JprepSync, s.Connections, s.ZapLogger, s.Cfg, s.AuthHelper)))
	allocateMarkerStep := allocate_marker.InitStep((*allocate_marker.Suite)(utils.NewEntitySuite(s.StepState.AllocateMarkerStepState, s.Connections, s.ZapLogger, s.Cfg, s.AuthHelper)))
	learningHistoryDataSyncStep := learning_history_data_sync.InitStep((*learning_history_data_sync.Suite)(utils.NewEntitySuite(s.StepState.LearningHistoryDataSyncStepState, s.Connections, s.ZapLogger, s.Cfg, s.AuthHelper)))

	utils.AppendSteps(steps, individualStudyPlanStep)
	utils.AppendSteps(steps, masterStudyPlanStep)
	utils.AppendSteps(steps, studyPlanItemStep)
	utils.AppendSteps(steps, studentEventLogsStep)
	utils.AppendSteps(steps, topicStatisticalStep)
	utils.AppendSteps(steps, studentProgressStep)
	utils.AppendSteps(steps, jprepSyncStep)
	utils.AppendSteps(steps, allocateMarkerStep)
	utils.AppendSteps(steps, learningHistoryDataSyncStep)

	return steps
}
