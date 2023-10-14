package helpers

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/communication/common/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

// Create 5 schools with the same school level
func (helper *CommunicationHelper) CreateSchoolsForOrganization(ctx context.Context, orgID string) ([]*entities.School, error) {
	ctx2 := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: orgID,
		},
	})

	now := time.Now()

	// create school level
	schoolLevel, err := helper.createASchoolLevel(ctx2, orgID)
	if err != nil {
		return nil, err
	}

	// create school level grade
	err = helper.createSchoolLevelGrade(ctx2, schoolLevel.ID, orgID)
	if err != nil {
		return nil, err
	}

	// create schools
	schools := []*entities.School{}
	for i := 0; i < 5; i++ {
		school := &entities.School{
			ID:    idutil.ULIDNow(),
			Name:  idutil.ULIDNow() + "-school",
			Level: schoolLevel,
		}
		query := `
			INSERT INTO school_info
			(school_id, school_name, is_archived, created_at, updated_at, school_level_id, school_partner_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7);
		`
		_, err := helper.BobDBConn.Exec(ctx2, query,
			database.Text(school.ID),
			database.Text(school.Name),
			database.Bool(false),
			database.Timestamptz(now),
			database.Timestamptz(now),
			database.Text(school.Level.ID),
			database.Text(school.ID), // for now it's same as school.ID
		)
		if err != nil {
			return nil, fmt.Errorf("failed CreateSchoolsForOrganization: %v", err)
		}
		schools = append(schools, school)
	}

	return schools, nil
}

func (helper *CommunicationHelper) createASchoolLevel(ctx context.Context, orgID string) (*entities.SchoolLevel, error) {
	schoolLevelID := idutil.ULIDNow()
	schoolLevelName := idutil.ULIDNow() + "-school-level"
	// nolint
	sequence := rand.Intn(999999999)
	isArchived := false
	createdAt := time.Now()
	updatedAt := time.Now()
	query := `
		INSERT INTO school_level
		(school_level_id, school_level_name, sequence, is_archived, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6);
	`
	_, err := helper.BobDBConn.Exec(ctx, query,
		database.Text(schoolLevelID),
		database.Text(schoolLevelName),
		sequence,
		database.Bool(isArchived),
		database.Timestamptz(createdAt),
		database.Timestamptz(updatedAt),
	)
	if err != nil {
		return nil, fmt.Errorf("failed createSchoolLevel: %v", err)
	}

	return &entities.SchoolLevel{
		ID:   schoolLevelID,
		Name: schoolLevelName,
	}, nil
}

// a student only count as studying at a school if have the school level grade reference
func (helper *CommunicationHelper) createSchoolLevelGrade(ctx context.Context, schoolLevelID, orgID string) error {
	query := `
		INSERT INTO school_level_grade
		(school_level_id, grade_id, created_at, updated_at, resource_path)
		(
			SELECT $1, g.grade_id, now(), now(), g.resource_path
			FROM grade g
			where g.resource_path = $2
		);
	`

	_, err := helper.BobDBConn.Exec(ctx, query,
		database.Text(schoolLevelID),
		database.Text(orgID),
	)
	if err != nil {
		return fmt.Errorf("failed createSchoolLevelGrade: %v", err)
	}
	return nil
}
