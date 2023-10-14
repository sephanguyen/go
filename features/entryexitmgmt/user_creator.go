package entryexitmgmt

import (
	"context"
	"errors"
	"fmt"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	bob_repo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	user_repo "github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	user_entities "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type userCreator struct {
	User     *bob_entities.User
	SchoolID int64
}

func (uc *userCreator) AsStudent(ctx context.Context, tx pgx.Tx, resourcePath string) error {
	studentRepo := bob_repo.StudentRepo{}

	_, err := studentRepo.Find(ctx, tx, uc.User.ID)
	if errors.Is(err, fmt.Errorf("row.Scan: %w", pgx.ErrNoRows)) {
		return err
	}

	e := &bob_entities.Student{}
	database.AllNullEntity(e)
	e.User = *uc.User
	err = multierr.Combine(
		e.ID.Set(uc.User.ID),
		e.SchoolID.Set(uc.SchoolID),
		e.StudentNote.Set("example-student-note"),
		e.ResourcePath.Set(resourcePath),
	)
	if err != nil {
		return err
	}
	err = studentRepo.CreateEn(ctx, tx, e)
	if err != nil {
		return err
	}

	return nil
}

func (uc *userCreator) AsTeacher(ctx context.Context, tx pgx.Tx, resourcePath string) error {
	teacherRepo := bob_repo.TeacherRepo{}

	_, err := teacherRepo.FindByID(ctx, tx, uc.User.ID)
	if err != pgx.ErrNoRows {
		return err
	}

	t := &bob_entities.Teacher{}
	database.AllNullEntity(t)
	t.ID = uc.User.ID
	err = t.SchoolIDs.Set([]int64{uc.SchoolID})
	if err != nil {
		return err
	}
	_ = t.ResourcePath.Set(resourcePath)

	err = teacherRepo.CreateMultiple(ctx, tx, []*bob_entities.Teacher{t})
	if err != nil {
		return err
	}

	return nil
}

func (uc *userCreator) AsSchoolAdmin(ctx context.Context, tx pgx.Tx, resourcePath string) error {
	schoolAdminRepo := bob_repo.SchoolAdminRepo{}

	_, err := schoolAdminRepo.Get(ctx, tx, uc.User.ID)
	if err != pgx.ErrNoRows {
		return err
	}

	schoolAdminAccount := &bob_entities.SchoolAdmin{}
	database.AllNullEntity(schoolAdminAccount)
	err = multierr.Combine(
		schoolAdminAccount.SchoolAdminID.Set(uc.User.ID.String),
		schoolAdminAccount.SchoolID.Set(uc.SchoolID),
		schoolAdminAccount.ResourcePath.Set(resourcePath),
	)
	if err != nil {
		return err
	}
	err = schoolAdminRepo.CreateMultiple(ctx, tx, []*bob_entities.SchoolAdmin{schoolAdminAccount})
	if err != nil {
		return err
	}
	return nil
}

func (uc *userCreator) AsParent(ctx context.Context, tx pgx.Tx, resourcePath string) error {
	parentRepo := user_repo.ParentRepo{}
	parentEnt := &user_entities.Parent{}
	database.AllNullEntity(parentEnt)
	err := multierr.Combine(
		parentEnt.ID.Set(uc.User.ID.String),
		parentEnt.SchoolID.Set(uc.SchoolID),
		parentEnt.ResourcePath.Set(resourcePath),
	)
	if err != nil {
		return err
	}
	err = parentRepo.CreateMultiple(ctx, tx, []*user_entities.Parent{parentEnt})
	if err != nil {
		return fmt.Errorf("cannot create parent: %w", err)
	}

	return nil
}
