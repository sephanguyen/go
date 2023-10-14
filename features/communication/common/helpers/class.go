package helpers

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"

	"github.com/manabie-com/backend/features/communication/common/entities"
	"github.com/manabie-com/backend/features/eibanam/communication/util"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
)

func (helper *CommunicationHelper) CreateClass(admin *entities.Staff, schoolID int32, courseID, locationID string, numOfClasses int) ([]*entities.Class, error) {
	ctx, cancel := util.ContextWithTokenAndTimeOut(context.Background(), admin.Token)
	defer cancel()

	resourcePath := strconv.Itoa(int(schoolID))

	ctxWithResourcePath := golibs.ResourcePathToCtx(ctx, resourcePath)

	var classes []*entities.Class

	for i := 0; i < numOfClasses; i++ {
		classID := idutil.ULIDNow()
		class := &entities.Class{
			ID:             classID,
			Name:           fmt.Sprintf("class-%s", classID),
			OrganizationID: fmt.Sprint(schoolID),
			CourseID:       courseID,
			LocationID:     locationID,
		}

		if schoolID == constants.JPREPSchool {
			// generate random number as class id
			intClassID, err := rand.Int(rand.Reader, big.NewInt(100000))
			if err != nil {
				return nil, fmt.Errorf("failed random class id: %v", err)
			}
			// convert to string
			strClassID := strconv.Itoa(int(intClassID.Int64()))
			class.ID = strClassID
			class.Name = fmt.Sprintf("class-%s", strClassID)
		}

		//TODO: refactor to use official API
		query := `
			INSERT INTO class (class_id, name, course_id, school_id, location_id, created_at, updated_at, deleted_at, resource_path)
			VALUES ($1, $2, $3, $4, $5, now(), now(), NULL, $6)
		`
		if _, err := helper.BobDBConn.Exec(ctxWithResourcePath, query, class.ID, class.Name, class.CourseID, class.OrganizationID, class.LocationID, resourcePath); err != nil {
			return nil, err
		}
		classes = append(classes, class)
	}

	return classes, nil
}
