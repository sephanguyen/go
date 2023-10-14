package mock

import (
	"path/filepath"

	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/bob/services"
	"github.com/manabie-com/backend/internal/bob/services/uploads"
	"github.com/manabie-com/backend/internal/golibs/tools"

	"github.com/spf13/cobra"
)

func genBobRepo(cmd *cobra.Command, args []string) error {
	repos := map[string]interface{}{
		"learning_objective":               &repositories.LearningObjectiveRepo{},
		"preset_student_plan":              &repositories.PresetStudyPlanRepo{},
		"question":                         &repositories.QuestionRepo{},
		"quizsets":                         &repositories.QuestionSetRepo{},
		"student":                          &repositories.StudentRepo{},
		"student_event_log":                &repositories.StudentEventLogRepo{},
		"student_stat":                     &repositories.StudentStatRepo{},
		"topic":                            &repositories.TopicRepo{},
		"student_comment":                  &repositories.StudentCommentRepo{},
		"user":                             &repositories.UserRepo{},
		"activity_log":                     &repositories.ActivityLogRepo{},
		"student_topic_completeness":       &repositories.StudentTopicCompletenessRepo{},
		"course_service":                   &services.CourseService{},
		"student_orders":                   &repositories.StudentOrderRepo{},
		"config":                           &repositories.ConfigRepo{},
		"student_learning_time_daily":      &repositories.StudentLearningTimeDailyRepo{},
		"student_topic_overdue":            &repositories.StudentTopicOverdueRepo{},
		"class":                            &repositories.ClassRepo{},
		"class_member":                     &repositories.ClassMemberRepo{},
		"school_config":                    &repositories.SchoolConfigRepo{},
		"student_assignment":               &repositories.StudentAssignmentRepo{},
		"assignment":                       &repositories.AssignmentRepo{},
		"teacher":                          &repositories.TeacherRepo{},
		"school_admin":                     &repositories.SchoolAdminRepo{},
		"student_submission":               &repositories.StudentSubmissionRepo{},
		"student_submission_score":         &repositories.StudentSubmissionScoreRepo{},
		"lesson":                           &repositories.LessonRepo{},
		"course_access_path":               &repositories.CourseAccessPathRepo{},
		"lesson_report":                    &repositories.LessonReportRepo{},
		"lesson_report_detail":             &repositories.LessonReportDetailRepo{},
		"lessonGroup":                      &repositories.LessonGroupRepo{},
		"course_class":                     &repositories.CourseClassRepo{},
		"school":                           &repositories.SchoolRepo{},
		"course":                           &repositories.CourseRepo{},
		"chapter":                          &repositories.ChapterRepo{},
		"apple_user":                       &repositories.AppleUserRepo{},
		"book":                             &repositories.BookRepo{},
		"course_book":                      &repositories.CourseBookRepo{},
		"media":                            &repositories.MediaRepo{},
		"book_chapter":                     &repositories.BookChapterRepo{},
		"quiz":                             &repositories.QuizRepo{},
		"quizset":                          &repositories.QuizSetRepo{},
		"shuffledquizset":                  &repositories.ShuffledQuizSetRepo{},
		"conversion_task":                  &repositories.ConversionTaskRepo{},
		"lesson_member":                    &repositories.LessonMemberRepo{},
		"academic_year":                    &repositories.AcademicYearRepo{},
		"parent":                           &repositories.ParentRepo{},
		"student_parent":                   &repositories.StudentParentRepo{},
		"topics_learning_objectives":       &repositories.TopicsLearningObjectivesRepo{},
		"flashcard_progression":            &repositories.FlashcardProgressionRepo{},
		"partner_form_config":              &repositories.PartnerFormConfigRepo{},
		"lesson_report_approval_record":    &repositories.LessonReportApprovalRecordRepo{},
		"student_subscription":             &repositories.StudentSubscriptionRepo{},
		"student_subscription_access_path": &repositories.StudentSubscriptionAccessPathRepo{},
		"organization":                     &repositories.OrganizationRepo{},
		"virtual_classroom_log":            &repositories.VirtualClassroomLogRepo{},
		"student_enrollment_status":        &repositories.StudentEnrolledHistoryRepo{},
		"school_history":                   &repositories.SchoolHistoryRepo{},
		"tagged_user":                      &repositories.TaggedUserRepo{},
	}

	tools.MockRepository("mock_repositories", filepath.Join(args[0], "repositories"), "bob", repos)

	structs := map[string][]interface{}{
		"internal/bob/services/uploads": {
			&uploads.UploadReaderService{},
		},
	}

	if err := tools.GenMockStructs(structs); err != nil {
		return err
	}

	interfaces := map[string][]string{
		"internal/bob/services": {
			"EurekaStudentEventLogModifierSvc",
		},
	}
	return tools.GenMockInterfaces(interfaces)
}

func newGenBobCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bob [../../mock/bob]",
		Short: "Generate mock repositories for bob",
		Args:  cobra.ExactArgs(1),
		RunE:  genBobRepo,
	}
}
