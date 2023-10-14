package eureka

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	bobproto "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	master_mgmt_proto "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/segmentio/ksuid"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/proto"
)

func (s *suite) avalideventMasterRegistrationUpsert(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, event := s.aEvtMasterRegistrationUpsert(ctx)
	stepState.Event = event
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) avalideventMasterRegistrationDelete(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.avalideventMasterRegistrationUpsert(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	data, err := proto.Marshal(stepState.Event.(*npb.EventMasterRegistration))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	_, err = s.JSM.PublishContext(context.Background(), constants.SubjectSyncMasterRegistration, data)
	if ctx, err := s.eurekaMustCreateCourseClass(ctx); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot create courseClass for next step: %w", err)
	}

	for index := 0; index < len(stepState.Event.(*npb.EventMasterRegistration).Classes); index++ {
		stepState.Event.(*npb.EventMasterRegistration).Classes[index].ActionKind = npb.ActionKind_ACTION_KIND_DELETED
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aEvtMasterRegistrationUpsert(ctx context.Context) (context.Context, *npb.EventMasterRegistration) {
	stepState := StepStateFromContext(ctx)
	event := &npb.EventMasterRegistration{
		Classes: []*npb.EventMasterRegistration_Class{
			{
				ClassId:    1,
				CourseId:   ksuid.New().String(),
				ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
			},
			{
				ClassId:    2,
				CourseId:   ksuid.New().String(),
				ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
			},
			{
				ClassId:    3,
				CourseId:   ksuid.New().String(),
				ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
			},
		},
	}
	return StepStateToContext(ctx, stepState), event
}

func (s *suite) sendEventToNatsJS(ctx context.Context, eventType string, topic string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var data []byte
	var err error
	switch eventType {
	case "JprefMasterRegistration":
		data, err = proto.Marshal(stepState.Event.(*npb.EventMasterRegistration))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "SyncStudentPackageEvent":
		data, err = proto.Marshal(stepState.Event.(*npb.EventSyncStudentPackage))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "ClassEvent":
		data, err = stepState.Event.(*bobproto.EvtClassRoom).Marshal()
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "MasterMgmtClassEvent":
		data, err = proto.Marshal(stepState.Event.(*master_mgmt_proto.EvtClass))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	stepState.SchoolID = strconv.Itoa(constants.ManabieSchool)
	cctx := context.Background()
	cctx = s.setFakeClaimToContext(cctx, stepState.SchoolID, cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String())

	_, err = s.JSM.PublishContext(cctx, topic, data)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.jsm.PublishContext: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getCourseClass(ctx context.Context) (context.Context, []*entities.CourseClass, error) {
	stepState := StepStateFromContext(ctx)
	courseClasses := make([]*entities.CourseClass, 0, len(stepState.Event.(*npb.EventMasterRegistration).Classes))

	for _, item := range stepState.Event.(*npb.EventMasterRegistration).Classes {
		courseClass := &entities.CourseClass{}
		database.AllNullEntity(courseClass)
		err := multierr.Combine(
			courseClass.ID.Set(idutil.ULIDNow()),
			courseClass.CourseID.Set(item.CourseId),
			courseClass.ClassID.Set(strconv.Itoa(int(item.ClassId))),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), nil, fmt.Errorf("err set CourseClass: %w", err)
		}
		courseClasses = append(courseClasses, courseClass)
	}
	return StepStateToContext(ctx, stepState), courseClasses, nil
}

func (s *suite) eurekaMustCreateCourseClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, courseClass, err := s.getCourseClass(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	keys := make([]string, 0, len(courseClass))
	for _, item := range courseClass {
		keys = append(keys, "('"+item.CourseID.String+"', '"+item.ClassID.String+"')")
	}

	count := 0
	query := ("SELECT count(*) FROM course_classes WHERE (course_id, class_id) IN (" + strings.Join(keys, ",") + ")")
	if err := try.Do(func(attempt int) (retry bool, err error) {
		time.Sleep(250 * time.Millisecond)

		err = s.DB.QueryRow(ctx, query).Scan(&count)
		if err != nil {
			return true, err
		}

		if count != len(courseClass) {
			return true, fmt.Errorf("Eureka does not create course class correctly")
		}
		return attempt < 5, err
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eurekaMustUpdateCourseClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, courseClass, err := s.getCourseClass(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	keys := make([]string, 0, len(courseClass))
	for _, item := range courseClass {
		keys = append(keys, "('"+item.CourseID.String+"', '"+item.ClassID.String+"')")
	}

	count := 0
	query := ("SELECT count(*) FROM course_classes WHERE (course_id, class_id) IN (" + strings.Join(keys, ",") + ") AND deleted_at IS NOT NULL")
	if err := try.Do(func(attempt int) (retry bool, err error) {
		time.Sleep(2 * time.Second)

		err = s.DB.QueryRow(ctx, query).Scan(&count)
		if err != nil {
			return true, nil
		}
		if count == len(courseClass) {
			return false, nil
		}

		return attempt < 10, fmt.Errorf("Eureka does not update course class correctly")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
