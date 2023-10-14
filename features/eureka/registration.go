package eureka

import (
	"context"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

func (s *suite) aSchoolNameCountryCityDistrict(ctx context.Context, arg1, arg2, arg3, arg4 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	city := &entities_bob.City{
		Name:    database.Text(arg3),
		Country: database.Text(arg2),
	}
	district := &entities_bob.District{
		Name:    database.Text(arg4),
		Country: database.Text(arg2),
		City:    city,
	}
	school := &entities_bob.School{
		Name:           database.Text(arg1 + stepState.Random),
		Country:        database.Text(arg2),
		City:           city,
		District:       district,
		IsSystemSchool: pgtype.Bool{Bool: true, Status: pgtype.Present},
		Point:          pgtype.Point{Status: pgtype.Null},
	}
	stepState.Schools = append(stepState.Schools, school)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminInsertsSchools(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	r := &repositories.SchoolRepo{}
	if err := r.Import(ctx, s.BobDB, stepState.Schools); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentSchoolID = stepState.Schools[len(stepState.Schools)-1].ID.Int
	return StepStateToContext(ctx, stepState), nil
}
