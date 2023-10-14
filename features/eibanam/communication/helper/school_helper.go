package helper

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/features/eibanam/communication/entity"
	bobEntities "github.com/manabie-com/backend/internal/bob/entities"
	bobRepositories "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"

	"github.com/jackc/pgtype"
)

func (h *CommunicationHelper) CreateNewSchool(accountType string) (*entity.School, error) {
	var (
		schoolID   int32
		locationID string
	)
	// schoolType := ""
	// switch accountType {
	// case AccountTypeSchoolAdmin:
	// 	schoolType = ""
	// case AccountTypeJPREPSchoolAdmin:
	// 	schoolType = SchoolTypeJPREP
	// 	schoolID = constants.JPREPSchool
	// 	locationID = constants.JPREPOrgLocation
	// default:
	// 	return nil, errors.New("unsupported account type")
	// }

	// schoolID, err := h.insertNewSchool(schoolType)
	if accountType != AccountTypeJPREPSchoolAdmin {
		newSchoolID, defaultLocationID, _, err := h.Suite.NewOrgWithOrgLocation(context.Background())
		if err != nil {
			return nil, err
		}
		schoolID = newSchoolID
		locationID = defaultLocationID
	} else {
		schoolID = constants.JPREPSchool
		locationID = constants.JPREPOrgLocation
	}

	return &entity.School{
		ID:              schoolID,
		DefaultLocation: locationID,
	}, nil
}

func (h *CommunicationHelper) insertNewSchool(schoolType string) (int32, error) {
	random := idutil.ULIDNow()
	sch := &bobEntities.School{
		Name:           database.Text(random),
		Country:        database.Text(constant.CountryVN),
		IsSystemSchool: database.Bool(false),
		CreatedAt:      database.Timestamptz(time.Now()),
		UpdatedAt:      database.Timestamptz(time.Now()),
		Point: pgtype.Point{
			P:      pgtype.Vec2{X: 0, Y: 0},
			Status: 2,
		},
	}

	if schoolType == SchoolTypeJPREP {
		sch.ID = database.Int4(constants.JPREPSchool)
	}

	city := &bobEntities.City{
		Name:         database.Text(random),
		Country:      database.Text(constant.CountryVN),
		CreatedAt:    database.Timestamptz(time.Now()),
		UpdatedAt:    database.Timestamptz(time.Now()),
		DisplayOrder: database.Int2(0),
	}

	district := &bobEntities.District{
		Name:    database.Text(random),
		Country: database.Text(constant.CountryVN),
		City:    city,
	}
	sch.City = city
	sch.District = district
	repo := &bobRepositories.SchoolRepo{}
	err := repo.Import(context.Background(), h.bobDBConn, []*bobEntities.School{sch})
	if err != nil {
		return 0, err
	}
	// fake org id = school id
	orgID := database.Text(strconv.Itoa(int(sch.ID.Int)))
	resourcePath := orgID
	schoolText := database.Text(strconv.Itoa(int(sch.ID.Int)))

	// multi-tenant needs this
	_, err = h.bobDBConn.Exec(context.Background(), `INSERT INTO organizations(
	organization_id, tenant_id, name, resource_path)
	VALUES ($1, $2, $3, $4)`, orgID, schoolText, sch.Name, resourcePath)
	if err != nil {
		return 0, err
	}

	//Init auth info
	stmt :=
		`
		INSERT INTO organization_auths
			(organization_id, auth_project_id, auth_tenant_id)
		VALUES
			($1, 'fake_aud', ''),
			($2, 'dev-manabie-online', ''),
			($2, 'dev-manabie-online', 'integration-test-1-909wx')

		ON CONFLICT 
			DO NOTHING
		;
		`
	_, err = h.bobDBConn.Exec(context.Background(), stmt, sch.ID.Int, sch.ID.Int)
	if err != nil {
		return 0, fmt.Errorf("cannot init auth info: %v", err)
	}

	return sch.ID.Int, nil
}
