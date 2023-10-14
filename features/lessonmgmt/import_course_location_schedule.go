package lessonmgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/domain"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

func (s *Suite) theCourseLocationScheduleRequestPayloadWith(ctx context.Context, _ string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	validRow1 := ",01GWK3QSBD4D0PKBXENQN1CPY6,01GWK39VWTEFF3DMH8MJCPARE0,1_2,1,,3"
	validRow2 := ",01GWK3QSBD4D0PKBXENQN1CPY6,01GWK39VWX66PNQGQQMEXT8MZ0,1_2,2,3,"
	validRow3 := ",01GWK5J5V1S3PDX2Z459X77YXR,01GWK5HJ1YAZEK3D9SPN08CT81,1_2,3,,"
	validRow4 := ",01GWK3NF2ZF2TKNGEN26YJQMPZ,01GWK39VVZ1FBCK4740RJQA0N8,1-3,4,,"

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	header := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s", domain.IDLabel, domain.CourseIDLabel, domain.LocationIDLabel, domain.AcademicWeekLabel, domain.ProductTypeScheduleLabel, domain.FrequencyLabel, domain.TotalNoLessonLabel)
	stepState.Request = &lpb.ImportCourseLocationScheduleRequest{
		Payload: []byte(fmt.Sprintf(`%s
		%s
		%s
		%s
		%s`, header, validRow1, validRow2, validRow3, validRow4)),
	}
	stepState.ValidCsvRows = []string{validRow1, validRow2, validRow3}
	stepState.ImportLessonPartnerInternalIDs = []string{"partner-internal-id-5-19"}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) importingCourseLocationSchedule(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()

	stepState.Response, stepState.ResponseErr = lpb.NewCourseLocationScheduleServiceClient(s.LessonMgmtConn).
		ImportCourseLocationSchedule(contextWithToken(s, ctx), stepState.Request.(*lpb.ImportCourseLocationScheduleRequest))
	return StepStateToContext(ctx, stepState), nil
}
