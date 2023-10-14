package yasuo

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) jprepSyncAcademicYearToOurSystem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &npb.EventMasterRegistration{RawPayload: []byte("{}"), Signature: idutil.ULIDNow(), AcademicYears: stepState.Request.([]*npb.EventMasterRegistration_AcademicYear)}
	data, _ := proto.Marshal(req)
	_, err := s.JSM.PublishContext(ctx, constants.SubjectSyncMasterRegistration, data)
	if err != nil {
		return ctx, fmt.Errorf("Publish: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) someAcademicYearMessage(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	academicYears := []*npb.EventMasterRegistration_AcademicYear{}

	for i := 0; i < rand.Intn(10)+3; i++ {
		academicYears = append(academicYears, &npb.EventMasterRegistration_AcademicYear{
			ActionKind:     npb.ActionKind_ACTION_KIND_UPSERTED,
			AcademicYearId: idutil.ULIDNow(),
			Name:           "Year " + idutil.ULIDNow(),
			StartYearDate:  timestamppb.Now(),
			EndYearDate: &timestamppb.Timestamp{
				Seconds: time.Now().Unix() + 200*24*60*60,
			},
		})
	}

	stepState.Request = academicYears

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theseAcademicYearsMustBeStoreInOurSystem(ctx context.Context) (context.Context, error) {
	time.Sleep(time.Second)

	stepState := StepStateFromContext(ctx)

	academicYears := stepState.Request.([]*npb.EventMasterRegistration_AcademicYear)
	academicYearRepo := &repositories.AcademicYearRepo{}
	for _, a := range academicYears {
		academicYear, err := academicYearRepo.Get(ctx, s.DBTrace, database.Text(a.AcademicYearId))
		if err != nil {
			return ctx, err
		}

		if academicYear.Name.String != a.Name {
			return ctx, fmt.Errorf("academicYear name does not match, expected %s, got %s", a.Name, academicYear.Name.String)
		}

		if academicYear.StartYearDate.Time.Unix() != a.StartYearDate.Seconds {
			return ctx, fmt.Errorf("academicYear startYearDate does not match, expected %s, got %s", a.StartYearDate.AsTime(), academicYear.StartYearDate.Time)
		}

		if academicYear.EndYearDate.Time.Unix() != a.EndYearDate.Seconds {
			return ctx, fmt.Errorf("academicYear endYearDate does not match, expected %s, got %s", a.EndYearDate.AsTime(), academicYear.EndYearDate.Time)
		}

		if academicYear.Status.String != entities.AcademicYearStatusActive {
			return ctx, fmt.Errorf("status must be active")
		}

		if academicYear.SchoolID.Int != constants.JPREPSchool {
			return ctx, fmt.Errorf("schoolID must be jprep school %v", constants.JPREPSchool)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
