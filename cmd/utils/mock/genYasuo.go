package mock

import (
	"path/filepath"

	"github.com/manabie-com/backend/internal/golibs/tools"

	"github.com/spf13/cobra"
)

func genYasuoRepo(cmd *cobra.Command, args []string) error {
	repos := map[string]interface{}{
		// "course":                   &repo_bob.CourseRepo{},
		// "chapter":                  &repo_bob.ChapterRepo{},
		// "activity_log":             &repositories.ActivityLogRepo{},
		// "lesson":                   &repositories.LessonRepo{},
		// "teacher":                  &repositories.TeacherRepo{},
		// "topic":                    &repositories.TopicRepo{},
		// "preset_study_plan_weekly": &repositories.PresetStudyPlanWeeklyRepo{},
		// "book":                     &repo_bob.BookRepo{},
		// "course_book":              &repo_bob.CourseBookRepo{},
		// "book_chapter":             &repo_bob.BookChapterRepo{},
		// "course_class":             &repositories.CourseClassRepo{},
		// "speeches":                 &repositories.SpeechesRepository{},
	}

	tools.MockRepository("mock_repositories", filepath.Join(args[0], "repositories"), "yasuo", repos)
	structs := map[string][]interface{}{
		"internal/yasuo/repositories": {},
	}
	if err := tools.GenMockStructs(structs); err != nil {
		return err
	}

	interfaces := map[string][]string{
		"internal/yasuo/services": {
			"BobMediaModifierServiceClient",
			"Uploader",
		},
	}
	return tools.GenMockInterfaces(interfaces)
}

func newGenYasuoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "yasuo [../../mock/yasuo]",
		Short: "Generate mock repositories for yasuo",
		Args:  cobra.ExactArgs(1),
		RunE:  genYasuoRepo,
	}
}
