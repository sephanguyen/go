package fatima

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/cmd/server/fatima"
	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/fatima/repositories"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"go.uber.org/multierr"
)

func (s *suite) aNumberOfExistingStudentPackages(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	num := idutil.ULIDNow()
	courseIDs := []string{
		fmt.Sprintf("existing-course-1-%s", num),
		fmt.Sprintf("existing-course-2-%s", num),
		fmt.Sprintf("existing-course-3-%s", num),
	}

	locationIDs := []string{constants.ManabieOrgLocation, constants.JPREPOrgLocation}

	for i := range courseIDs {
		sp, err := generateStudentPackageByCourseIDsAndLocationIDs(idutil.ULIDNow(), courseIDs[i:], nil)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("err generateStudentPackageByCourseIDsAndLocationIDs: %w", err)
		}
		stepState.StudentPackages = append(stepState.StudentPackages, sp)
		for j := range locationIDs {
			sp, err := generateStudentPackageByCourseIDsAndLocationIDs(idutil.ULIDNow(), courseIDs[i:], locationIDs[j:])
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("err generateStudentPackageByCourseIDsAndLocationIDs: %w", err)
			}
			stepState.StudentPackages = append(stepState.StudentPackages, sp)
		}
	}

	spRepo := &repositories.StudentPackageRepo{}

	// remember to set auth.InjectFakeJwtToken to EXISTING ctx
	ctx = auth.InjectFakeJwtToken(ctx, fmt.Sprint(constants.ManabieSchool))

	err := spRepo.BulkInsert(ctx, s.DB, stepState.StudentPackages)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("err spRepo.BulkInsert: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func generateStudentPackageByCourseIDsAndLocationIDs(studentID string, courseIDs, locationIDs []string) (*entities.StudentPackage, error) {
	var sp entities.StudentPackage
	database.AllNullEntity(&sp)
	now := timeutil.Now()
	startAt := now
	endAt := now.Add(time.Duration(90) * time.Hour * 24)
	err := multierr.Combine(
		sp.ID.Set(idutil.ULIDNow()),
		sp.StudentID.Set(studentID),
		sp.PackageID.Set("free_package"),
		sp.StartAt.Set(startAt),
		sp.EndAt.Set(endAt),
		sp.Properties.Set(&entities.StudentPackageProps{
			CanWatchVideo:     courseIDs,
			CanViewStudyGuide: courseIDs,
			CanDoQuiz:         courseIDs,
			LimitOnlineLesson: 0,
		}),
		sp.CreatedAt.Set(now),
		sp.UpdatedAt.Set(now),
		sp.IsActive.Set(true),
		sp.LocationIDs.Set(locationIDs),
	)
	return &sp, err
}

func (s *suite) systemRunJobToMigrateStudentPackagesToStudentPackageAccessPath(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	fatima.RunMigrateStudentPackagesToStudentPackageAccessPath(ctx, &fatimaConfig, &bobConfig)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentPackagesAndStudentPackageAccessPathAreCorrespondent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, sp := range stepState.StudentPackages {
		err := s.validateStudentPackageAccessPath(ctx, sp)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("err s.validateStudentPackageAccessPath: %w, student_package_id: %s", err, sp.ID.String)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
