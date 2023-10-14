package usermgmt

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/nats"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/encoding/protojson"
)

const countSampleStudents = 10

func (s *suite) generateImportStudentEventRecords(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	importUserEvents := make([]*entity.ImportUserEvent, 0)

	_, err := s.aSignedInSchoolAdminWithSchoolID(ctx, entity.UserGroupSchoolAdmin, constants.ManabieSchool)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.aSignedInSchoolAdmin err: %v", err)
	}

	for i := 0; i < countSampleStudents; i++ {
		studentID := idutil.ULIDNow()
		_, err := s.aValidStudentInDB(ctx, studentID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.aValidStudentInDB err: %v", err)
		}
		createStudentEvent := &pb.EvtUser{
			Message: &pb.EvtUser_CreateStudent_{
				CreateStudent: &pb.EvtUser_CreateStudent{
					StudentId:   studentID,
					SchoolId:    fmt.Sprint(constants.ManabieSchool),
					LocationIds: []string{fmt.Sprint(constants.ManabieSchool)},
				},
			},
		}
		importUserEvent := &entity.ImportUserEvent{}
		database.AllNullEntity(importUserEvent)

		payload, err := protojson.Marshal(createStudentEvent)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("protojson.Marshal err: %v", err)
		}

		err = multierr.Combine(
			importUserEvent.ImporterID.Set(stepState.CurrentUserID),
			importUserEvent.Status.Set(cpb.ImportUserEventStatus_IMPORT_USER_EVENT_STATUS_WAITING.String()),
			importUserEvent.UserID.Set(studentID),
			importUserEvent.Payload.Set(payload),
			importUserEvent.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("multierr.Combine err: %v", err)
		}

		importUserEvents = append(importUserEvents, importUserEvent)
	}

	ctx = auth.InjectFakeJwtToken(ctx, fmt.Sprint(constants.ManabieSchool))
	repo := &repository.ImportUserEventRepo{}
	importUserEvents, err = repo.Upsert(ctx, s.BobDBTrace, importUserEvents)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("repo.Upsert err: %v", err)
	}
	stepState.ImportUserEvents = importUserEvents

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateImportParentEventRecords(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	schoolID := int64(stepState.CurrentSchoolID)
	if schoolID == 0 {
		schoolID = constants.ManabieSchool
	}

	importUserEvents := make([]*entity.ImportUserEvent, 0)

	_, err := s.aSignedInSchoolAdminWithSchoolID(ctx, entity.UserGroupSchoolAdmin, constants.ManabieSchool)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.aSignedInSchoolAdmin err: %v", err)
	}

	for i := 0; i < countSampleStudents; i++ {
		parentID := idutil.ULIDNow()
		studentID := idutil.ULIDNow()
		num := rand.Int()
		_, err := s.aValidStudentInDB(ctx, studentID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.aValidStudentInDB err: %v", err)
		}
		_, err = aValidParentInDB(auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID)), s.BobDBTrace, parentID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.aValidParentInDB err: %v", err)
		}
		createStudentEvent := &pb.EvtUser{
			Message: &pb.EvtUser_CreateParent_{
				CreateParent: &pb.EvtUser_CreateParent{
					StudentId:   studentID,
					StudentName: fmt.Sprintf("valid-user-%d", num),
					ParentId:    parentID,
					SchoolId:    fmt.Sprint(constants.ManabieSchool),
				},
			},
		}
		importUserEvent := &entity.ImportUserEvent{}
		database.AllNullEntity(importUserEvent)

		payload, err := protojson.Marshal(createStudentEvent)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("protojson.Marshal err: %v", err)
		}

		err = multierr.Combine(
			importUserEvent.ImporterID.Set(stepState.CurrentUserID),
			importUserEvent.Status.Set(cpb.ImportUserEventStatus_IMPORT_USER_EVENT_STATUS_WAITING.String()),
			importUserEvent.UserID.Set(parentID),
			importUserEvent.Payload.Set(payload),
			importUserEvent.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("multierr.Combine err: %v", err)
		}

		importUserEvents = append(importUserEvents, importUserEvent)
	}

	ctx = auth.InjectFakeJwtToken(ctx, fmt.Sprint(constants.ManabieSchool))
	repo := &repository.ImportUserEventRepo{}
	importUserEvents, err = repo.Upsert(auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID)), s.BobDBTrace, importUserEvents)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("repo.Upsert err: %v", err)
	}
	stepState.ImportUserEvents = importUserEvents

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) systemRunTaskToPublishImportUserEvents(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	schoolID := int64(stepState.CurrentSchoolID)
	if schoolID == 0 {
		schoolID = constants.ManabieSchool
	}

	importUserEventIDs := stepState.ImportUserEvents.IDs()

	err := s.TaskQueue.Add(nats.PublishImportUserEventsTask(auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID)), s.BobDBTrace, s.JSM, &nats.PublishImportUserEventsTaskOptions{
		ImportUserEventIDs: importUserEventIDs,
		ResourcePath:       golibs.ResourcePathFromCtx(ctx),
	}))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("task.TaskQueue.Add err: %v", err)
	}
	time.Sleep(5 * time.Second)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) downstreamServicesConsumeTheEvents(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) statusOfImportUserEventsGetUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	schoolID := int64(stepState.CurrentSchoolID)
	if schoolID == 0 {
		schoolID = constants.ManabieSchool
	}

	importUserEventIDs := stepState.ImportUserEvents.IDs()

	repo := &repository.ImportUserEventRepo{}
	importUserEvents, err := repo.GetByIDs(auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID)), s.BobDBTrace, database.Int8Array(importUserEventIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("repo.Find err: %v", err)
	}

	for _, importUserEvent := range importUserEvents {
		if importUserEvent.Status.String != cpb.ImportUserEventStatus_IMPORT_USER_EVENT_STATUS_FINISHED.String() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("publish import_user_events id: %v failed: expected %v, but actual %v", importUserEvent.ID.Int, cpb.ImportUserEventStatus_IMPORT_USER_EVENT_STATUS_FINISHED.String(), importUserEvent.Status.String)
		}

		if importUserEvent.SequenceNumber.Status == pgtype.Null {
			return StepStateToContext(ctx, stepState), fmt.Errorf("publish import_user_events id: %v sequence_number is not nil", importUserEvent.ID.Int)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
