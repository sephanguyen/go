package helpers

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/features/communication/common/entities"
	"github.com/manabie-com/backend/features/eibanam/communication/util"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb_ms "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/pkg/errors"
)

func (helper *CommunicationHelper) CreateCourses(admin *entities.Staff, schoolID int32, numOfCourses int, locationIDs []string) ([]*entities.Course, error) {
	ctx, cancel := util.ContextWithTokenAndTimeOut(context.Background(), admin.Token)
	defer cancel()

	var courses []*entities.Course

	for i := 0; i < numOfCourses; i++ {
		id := rand.Int31()

		course := &entities.Course{
			ID:   idutil.ULIDNow(),
			Name: fmt.Sprintf("course-%d", id),
		}

		locIDs := []string{}
		if len(locationIDs) > 0 {
			locIDs = locationIDs
			course.LocationIDs = locationIDs
		}

		req := &pb_ms.UpsertCoursesRequest{
			Courses: []*pb_ms.UpsertCoursesRequest_Course{
				{
					Id:          course.ID,
					Name:        fmt.Sprintf("course-%d", id),
					SchoolId:    schoolID,
					LocationIds: locIDs,
				},
			},
		}

		res, err := pb_ms.NewMasterDataCourseServiceClient(helper.MasterMgmtGRPCConn).UpsertCourses(ctx, req)

		if err != nil {
			return nil, err
		}
		if !res.Successful {
			return nil, errors.New("failed to create course")
		}
		courses = append(courses, course)
	}
	return courses, nil
}

func (helper *CommunicationHelper) CreateCoursesWithClass(admin *entities.Staff, schoolID int32, numOfCourses, numOfClasses int, locationIDs []string) ([]*entities.Course, []*entities.Class, error) {
	courses, err := helper.CreateCourses(admin, schoolID, numOfCourses, locationIDs)
	if err != nil {
		return nil, nil, err
	}

	classList := make([]*entities.Class, 0)
	for _, course := range courses {
		classes, err := helper.CreateClass(admin, schoolID, course.ID, locationIDs[0], numOfClasses)
		if err != nil {
			return nil, nil, err
		}

		course.Classes = classes
		classList = append(classList, classes...)
	}
	return courses, classList, nil
}
