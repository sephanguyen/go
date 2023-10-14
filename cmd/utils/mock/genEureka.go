package mock

import (
	"path/filepath"

	"github.com/manabie-com/backend/internal/eureka/repositories"
	items_bank_repo "github.com/manabie-com/backend/internal/eureka/repositories/items_bank"
	lhds_repo "github.com/manabie-com/backend/internal/eureka/repositories/learning_history_data_sync"
	monitor_repo "github.com/manabie-com/backend/internal/eureka/repositories/monitors"
	"github.com/manabie-com/backend/internal/eureka/services"
	"github.com/manabie-com/backend/internal/golibs/tools"

	"github.com/spf13/cobra"
)

func genEurekaRepo(cmd *cobra.Command, args []string) error {
	repos := map[string]interface{}{
		"study_plan_repo":                 &repositories.StudyPlanRepo{},
		"assignment_repo":                 &repositories.AssignmentRepo{},
		"assignment_study_plan_item":      &repositories.AssignmentStudyPlanItemRepo{},
		"lo_study_plan_item":              &repositories.LoStudyPlanItemRepo{},
		"course_study_plan":               &repositories.CourseStudyPlanRepo{},
		"class_study_plan":                &repositories.ClassStudyPlanRepo{},
		"student_study_plan":              &repositories.StudentStudyPlanRepo{},
		"student":                         &repositories.StudentRepo{},
		"student_latest_submission":       &repositories.StudentLatestSubmissionRepo{},
		"student_submission":              &repositories.StudentSubmissionRepo{},
		"student_submission_grade":        &repositories.StudentSubmissionGradeRepo{},
		"class_student":                   &repositories.ClassStudentRepo{},
		"course_class":                    &repositories.CourseClassRepo{},
		"course_student":                  &repositories.CourseStudentRepo{},
		"assign_study_plan_task_modifier": &repositories.AssignStudyPlanTaskRepo{},
		"topics_assignments":              &repositories.TopicsAssignmentsRepo{},
		"book":                            &repositories.BookRepo{},
		"chapter":                         &repositories.ChapterRepo{},
		"book_chapter":                    &repositories.BookChapterRepo{},
		"course_book":                     &repositories.CourseBookRepo{},
		"topic_repo":                      &repositories.TopicRepo{},
		"flashcard_progression":           &repositories.FlashcardProgressionRepo{},
		"learning_objective":              &repositories.LearningObjectiveRepo{},
		"quiz":                            &repositories.QuizRepo{},
		"student_learning_time_daily":     &repositories.StudentLearningTimeDailyRepo{},
		"students_learning_objectives_completeness": &repositories.StudentsLearningObjectivesCompletenessRepo{},
		"topics_learning_objectives":                &repositories.TopicsLearningObjectivesRepo{},
		"student_event_log":                         &repositories.StudentEventLogRepo{},
		"course_student_access_path":                &repositories.CourseStudentAccessPathRepo{},
		"general_assignment":                        &repositories.GeneralAssignmentRepo{},
		"flashcard":                                 &repositories.FlashcardRepo{},
		"learning_material":                         &repositories.LearningMaterialRepo{},
		"learning_objective_v2":                     &repositories.LearningObjectiveRepoV2{},
		"exam_lo":                                   &repositories.ExamLORepo{},
		"exam_lo_submission":                        &repositories.ExamLOSubmissionRepo{},
		"exam_lo_submission_answer":                 &repositories.ExamLOSubmissionAnswerRepo{},
		"exam_lo_submission_score":                  &repositories.ExamLOSubmissionScoreRepo{},
		"user":                                      &repositories.UserRepo{},
		"master_study_plan":                         &repositories.MasterStudyPlanRepo{},
		"individual_study_plan":                     &repositories.IndividualStudyPlan{},
		"task_assignment":                           &repositories.TaskAssignmentRepo{},
		"shuffled_quiz_set":                         &repositories.ShuffledQuizSetRepo{},
		"question_group":                            &repositories.QuestionGroupRepo{},
		"question_tag_type":                         &repositories.QuestionTagTypeRepo{},
		"statistics":                                &repositories.StatisticsRepo{},
		"question_tag":                              &repositories.QuestionTagRepo{},
		"speeches":                                  &repositories.SpeechesRepository{},
		"allocate_marker":                           &repositories.AllocateMarkerRepo{},
		"lo_submission_answer":                      &repositories.LOSubmissionAnswerRepo{},
		"flash_card_submission_answer":              &repositories.FlashCardSubmissionAnswerRepo{},
		"lo_progression":                            &repositories.LOProgressionRepo{},
		"lo_progression_answer":                     &repositories.LOProgressionAnswerRepo{},
		"import_study_plan_task":                    &repositories.ImportStudyPlanTaskRepo{},
		"assessment_session":                        &repositories.AssessmentSessionRepo{},
		"content_bank_repo":                         &repositories.ContentBankMediaRepo{},
		"assessment":                                &repositories.AssessmentRepo{},
		// TODO: https://manabie.slack.com/archives/C01SYKX1BPE/p1653015354300239
	}
	services := map[string]interface{}{
		"assignment_modifier_service": &services.AssignmentModifierService{},
		"course_reader_service":       &services.CourseReaderService{},
	}
	tools.MockRepository("mock_repositories", filepath.Join(args[0], "repositories"), "eureka", repos)
	tools.MockRepository("mock_services", filepath.Join(args[0], "services"), "eureka", services)

	// This specific entity's import paths are different so we need to handle it separately
	tools.RemoveImport("github.com/manabie-com/backend/internal/eureka/entities")
	tools.AddImport("github.com/manabie-com/backend/internal/bob/entities")
	if err := tools.GenMockStructs(map[string][]interface{}{
		"internal/eureka/repositories": {
			// &repositories.TopicRepo{},
		},
	}); err != nil {
		return err
	}
	tools.ResetImports()

	tools.RemoveImport("github.com/manabie-com/backend/internal/eureka/entities")
	tools.RemoveImport("github.com/manabie-com/backend/internal/eureka/repositories")
	tools.AddImport("github.com/manabie-com/backend/internal/eureka/entities/monitors")
	tools.AddImport("github.com/manabie-com/backend/internal/eureka/repositories/monitors")
	if err := tools.GenMockStructs(map[string][]interface{}{
		"internal/eureka/repositories/monitors": {
			&monitor_repo.StudyPlanMonitorRepo{},
		},
	}); err != nil {
		return err
	}
	tools.ResetImports()

	tools.RemoveImport("github.com/manabie-com/backend/internal/eureka/entities")
	tools.RemoveImport("github.com/manabie-com/backend/internal/eureka/repositories")
	tools.AddImport("github.com/manabie-com/backend/internal/eureka/entities/learning_history_data_sync")
	tools.AddImport("github.com/manabie-com/backend/internal/eureka/repositories/learning_history_data_sync")
	if err := tools.GenMockStructs(map[string][]interface{}{
		"internal/eureka/repositories/learning_history_data_sync": {
			&lhds_repo.LearningHistoryDataSyncRepo{},
		},
	}); err != nil {
		return err
	}

	tools.ResetImports()

	tools.RemoveImport("github.com/manabie-com/backend/internal/eureka/entities")
	tools.RemoveImport("github.com/manabie-com/backend/internal/eureka/repositories")
	tools.AddImport("github.com/manabie-com/backend/internal/eureka/entities/items_bank")
	tools.AddImport("github.com/manabie-com/backend/internal/eureka/repositories/items_bank")
	if err := tools.GenMockStructs(map[string][]interface{}{
		"internal/eureka/repositories/items_bank": {
			&items_bank_repo.ItemsBankRepo{},
		},
	}); err != nil {
		return err
	}

	tools.ResetImports()

	tools.AddImportWithPkgAlias("github.com/manabie-com/backend/pkg/manabuf/syllabus/v1", "sspb")
	if err := tools.GenMockStructs(map[string][]interface{}{
		"internal/eureka/repositories": {
			&repositories.StudyPlanItemRepo{},
		},
	}); err != nil {
		return err
	}
	tools.ResetImports()

	interfaces := map[string][]string{
		"internal/eureka/services": {
			"BobStudentReaderServiceClient",
			"BobCourseClientServiceClient",
			"YasuoUploadModifierServiceClient",
			"YasuoUploadReaderServiceClient",
		},
		"internal/eureka/services/learning_history_data_sync": {
			"YasuoUploadModifierService",
		},
	}
	return tools.GenMockInterfaces(interfaces)
}

func newGenEurekaCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "eureka [../../mock/eureka]",
		Short: "generate eureka repository type",
		Args:  cobra.ExactArgs(1),
		RunE:  genEurekaRepo,
	}
}
