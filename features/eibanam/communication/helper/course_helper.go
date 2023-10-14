package helper

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/features/eibanam/communication/entity"
	"github.com/manabie-com/backend/features/eibanam/communication/util"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	bobPb "github.com/manabie-com/backend/pkg/genproto/bob"
	yasuoPb "github.com/manabie-com/backend/pkg/genproto/yasuo"

	"github.com/pkg/errors"
)

func (h *CommunicationHelper) CreateCourses(admin *entity.Admin, schoolId, gradeId int32, numOfCourses int) ([]*entity.Course, error) {
	ctx, cancel := util.ContextWithTokenAndTimeOut(context.Background(), admin.Token)
	defer cancel()

	var courses []*entity.Course

	for i := 0; i < numOfCourses; i++ {
		id := rand.Int31()

		course := &entity.Course{
			ID:      idutil.ULIDNow(),
			Name:    fmt.Sprintf("course-%d", id),
			GradeID: gradeId,
		}

		country := bobPb.COUNTRY_VN
		grade, _ := i18n.ConvertIntGradeToString(country, int(gradeId))

		course.GradeName = grade

		req := &yasuoPb.UpsertCoursesRequest{
			Courses: []*yasuoPb.UpsertCoursesRequest_Course{
				{
					Id:       course.ID,
					Name:     fmt.Sprintf("course-%d", id),
					Country:  country,
					Subject:  bobPb.SUBJECT_BIOLOGY,
					SchoolId: schoolId,
					Grade:    course.GradeName,
				},
			},
		}

		res, err := yasuoPb.NewCourseServiceClient(h.yasuoGRPCConn).UpsertCourses(ctx, req)
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
