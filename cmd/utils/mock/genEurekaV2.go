package mock

import (
	assessment_learnosity "github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/repository/learnosity"
	assessment_postgres "github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/repository/postgres"
	assessment_usecase "github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/usecase"
	book_postgres "github.com/manabie-com/backend/internal/eureka/v2/modules/book/repository/postgres"
	book_usecase "github.com/manabie-com/backend/internal/eureka/v2/modules/book/usecase"
	course_external "github.com/manabie-com/backend/internal/eureka/v2/modules/course/repository/external"
	course_postgres "github.com/manabie-com/backend/internal/eureka/v2/modules/course/repository/postgres"
	course_usecase "github.com/manabie-com/backend/internal/eureka/v2/modules/course/usecase"
	item_bank_learnosity "github.com/manabie-com/backend/internal/eureka/v2/modules/item_bank/repository/learnosity"
	item_bank_usecase "github.com/manabie-com/backend/internal/eureka/v2/modules/item_bank/usecase"
	study_plan_postgres "github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/repository/postgres"
	study_plan_usecase "github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/usecase"
	"github.com/manabie-com/backend/internal/golibs/tools"

	"github.com/spf13/cobra"
	"go.uber.org/multierr"
)

func genEurekaV2(_ *cobra.Command, _ []string) error {
	bookRepos := map[string][]interface{}{
		"internal/eureka/v2/modules/book/repository/postgres": {
			&book_postgres.BookRepo{},
			&book_postgres.LearningMaterialRepo{},
		},
	}
	bookUseCases := map[string][]interface{}{
		"internal/eureka/v2/modules/book/usecase/repo": {
			&book_usecase.BookUsecase{},
			&book_usecase.LearningMaterialUsecase{},
		},
	}
	assessmentRepos := map[string][]interface{}{
		"internal/eureka/v2/modules/assessment/repository/postgres": {
			&assessment_postgres.AssessmentSessionRepo{},
			&assessment_postgres.AssessmentRepo{},
			&assessment_postgres.SubmissionRepo{},
			&assessment_postgres.FeedbackSessionRepo{},
			&assessment_postgres.StudentEventLogRepo{},
		},
		"internal/eureka/v2/modules/assessment/repository/learnosity": {
			&assessment_learnosity.SessionRepo{},
		},
	}
	assessmentUseCases := map[string][]interface{}{
		"internal/eureka/v2/modules/assessment/usecase": {
			&assessment_usecase.AssessmentUsecaseImpl{},
		},
	}
	itemBankRepos := map[string][]interface{}{
		"internal/eureka/v2/modules/item_bank/repository/learnosity": {
			&item_bank_learnosity.ItemBankRepo{},
		},
	}
	itemBankUseCases := map[string][]interface{}{
		"internal/eureka/v2/modules/item_bank/usecase": {
			&item_bank_usecase.ActivityUsecase{},
		},
	}
	studyPlanRepo := map[string][]interface{}{
		"internal/eureka/v2/modules/study_plan/repository/postgres": {
			&study_plan_postgres.StudyPlanItemRepo{},
			&study_plan_postgres.LmListRepo{},
		},
	}
	studyPlanItemUseCases := map[string][]interface{}{
		"internal/eureka/v2/modules/study_plan/usecase": {
			&study_plan_usecase.StudyPlanItemUseCase{},
		},
	}
	studyPlanRepos := map[string][]interface{}{
		"internal/eureka/v2/modules/study_plan/repository/postgres": {
			&study_plan_postgres.StudyPlanRepo{},
		},
	}
	studyPlanUseCases := map[string][]interface{}{
		"internal/eureka/v2/modules/study_plan/usecase": {
			&study_plan_usecase.StudyPlanUsecaseImpl{},
		},
	}
	err := multierr.Combine(
		tools.GenMockStructs(bookRepos),
		tools.GenMockStructs(bookUseCases),
		tools.GenMockStructs(assessmentRepos),
		tools.GenMockStructs(assessmentUseCases),
		tools.GenMockStructs(itemBankRepos),
		tools.GenMockStructs(itemBankUseCases),
		tools.GenMockStructs(studyPlanRepo),
		tools.GenMockStructs(studyPlanItemUseCases),
		tools.GenMockStructs(studyPlanRepos),
		tools.GenMockStructs(studyPlanUseCases),
	)

	courseRepos := map[string][]interface{}{
		"internal/eureka/v2/modules/course/repository/postgres": {
			&course_postgres.CourseRepo{},
			&course_postgres.CourseBookRepo{},
		},
		"internal/eureka/v2/modules/course/repository/external": {
			&course_external.CerebryRepo{},
			&course_external.CourseRepo{},
		},
	}
	courseUseCases := map[string][]interface{}{
		"internal/eureka/v2/modules/course/usecase": {
			&course_usecase.CourseUsecase{},
			&course_usecase.CerebryUsecase{},
		},
	}

	tools.RemoveImport("github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain")
	tools.AddImport("github.com/manabie-com/backend/internal/eureka/v2/modules/course/domain")
	return multierr.Combine(
		err,
		tools.GenMockStructs(courseRepos),
		tools.GenMockStructs(courseUseCases),
	)
}

func newGenEurekaV2Cmd() *cobra.Command {
	return &cobra.Command{
		Use:   "eureka_v2 [../../mock/eureka/v2]",
		Short: "generate eureka v2 repository type",
		Args:  cobra.NoArgs,
		RunE:  genEurekaV2,
	}
}
