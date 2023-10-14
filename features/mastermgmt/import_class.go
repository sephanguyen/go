package mastermgmt

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"google.golang.org/protobuf/proto"
)

func (s *suite) validClassesPayload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courseIDs := stepState.CourseIDs
	locationIDs := stepState.CenterIDs
	randLocation := locationIDs[rand.Intn(len(locationIDs))]
	randCourse := func() string {
		return courseIDs[rand.Intn(len(courseIDs))]
	}
	format := "%s,%s,%s,%s,%s"
	r1 := fmt.Sprintf(format, randCourse(), randLocation, "", "", idutil.ULIDNow())
	r2 := fmt.Sprintf(format, randCourse(), randLocation, "", "", idutil.ULIDNow())
	r3 := fmt.Sprintf(format, randCourse(), randLocation, "", "", idutil.ULIDNow())

	request := fmt.Sprintf(`course_id,location_id,course_name,location_name,class_name
	%s
	%s
	%s`, r1, r2, r3)
	stepState.Request = &mpb.ImportClassRequest{
		Payload: []byte(request),
	}
	stepState.ValidCsvRows = []string{r1, r2, r3}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validAndInvalidlassesPayload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courseIDs := stepState.CourseIDs
	locationIDs := stepState.CenterIDs
	randLocation := locationIDs[rand.Intn(len(locationIDs))]
	randCourse := func() string {
		return courseIDs[rand.Intn(len(courseIDs))]
	}
	format := "%s,%s,%s,%s,%s"
	r1 := fmt.Sprintf(format, randCourse(), randLocation, "", "", idutil.ULIDNow())
	r2 := fmt.Sprintf(format, randCourse(), randLocation, "", "", idutil.ULIDNow())
	r3 := fmt.Sprintf(format, randCourse(), randLocation, "", "", idutil.ULIDNow())
	r4 := fmt.Sprintf(format, randCourse(), "", "", "", idutil.ULIDNow())
	r5 := fmt.Sprintf(format, "invalid-course", randLocation, "", "", idutil.ULIDNow())
	r6 := fmt.Sprintf(format, "invalid-course", "invalid-location", "", "", idutil.ULIDNow())
	r7 := fmt.Sprintf(format, randCourse(), randLocation, "", "", string([]byte{0xff, 0xfe, 0xfd}))

	request := fmt.Sprintf(`course_id,location_id,course_name,location_name,class_name
	%s
	%s
	%s
	%s
	%s
	%s
	%s`, r1, r2, r3, r4, r5, r6, r7)
	stepState.Request = &mpb.ImportClassRequest{
		Payload: []byte(request),
	}
	stepState.ValidCsvRows = []string{r1, r2, r3}
	stepState.InvalidCsvRows = []string{r4, r5, r6, r7}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*mpb.ImportClassRequest)
	ctx, err := s.subscribeEventClass(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.subscribe: %w", err)
	}
	stepState.Response, stepState.ResponseErr = mpb.NewClassServiceClient(s.MasterMgmtConn).ImportClass(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) subscribeEventClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}
	classEvtHandler := func(ctx context.Context, data []byte) (bool, error) {
		classEvent := &mpb.EvtClass{}
		if err := proto.Unmarshal(data, classEvent); err != nil {
			return false, err
		}
		switch msg := classEvent.Message.(type) {
		case *mpb.EvtClass_CreateClass_:
			stepState.FoundChanForJetStream <- msg
			return false, nil
		case *mpb.EvtClass_UpdateClass_:
			stepState.FoundChanForJetStream <- msg
			return false, nil
		case *mpb.EvtClass_DeleteClass_:
			stepState.FoundChanForJetStream <- msg
			return false, nil
		default:
			return true, fmt.Errorf("wrong message type")
		}
	}
	subs, err := s.JSM.Subscribe(constants.SubjectMasterMgmtClassUpserted, opts, classEvtHandler)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("SubscribeUpsertedClass:s.JSM.Subscribe: %w", err)
	}
	stepState.Subs = append(stepState.Subs, subs.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createdClassProperly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	validRows := s.StepState.ValidCsvRows
	classNames := make([]string, 0, len(validRows))
	expectedClass := make([]string, 0, len(validRows))
	for _, row := range validRows {
		rowS := strings.Split(row, ",")
		courseId := rowS[0]
		locationId := rowS[1]
		name := rowS[4]
		classNames = append(classNames, name)
		expectedClass = append(expectedClass, fmt.Sprintf("%s,%s,%s",
			name, courseId, locationId))
	}
	var (
		name       string
		courseID   string
		locationID string
	)
	query := "SELECT name,course_id,location_id FROM class WHERE name = ANY($1) AND deleted_at IS NULL"
	rows, err := s.BobDBTrace.Query(ctx, query, classNames)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	defer rows.Close()
	gotClass := []string{}
	for rows.Next() {
		if err := rows.Scan(&name, &courseID, &locationID); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
		}
		class := fmt.Sprintf("%s,%s,%s", name, courseID, locationID)
		gotClass = append(gotClass, class)
	}
	if equal := stringutil.SliceElementsMatch(expectedClass, gotClass); !equal {
		return StepStateToContext(ctx, stepState), fmt.Errorf("classes are not created properly")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnErrorOfInvalidClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*mpb.ImportClassResponse)
	rowInValid := stepState.InvalidCsvRows
	expectedErrorLen := len(rowInValid)
	gotErrorLen := len(resp.GetErrors())
	if expectedErrorLen == 0 && gotErrorLen == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	if expectedErrorLen != gotErrorLen {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected = %d,got = %d", expectedErrorLen, gotErrorLen)
	}
	req := stepState.Request.(*mpb.ImportClassRequest)
	reqSplit := strings.Split(string(req.Payload), "\n")
	gotRow := []string{}
	for _, error := range resp.GetErrors() {
		row := strings.TrimSpace(reqSplit[error.RowNumber-1])
		gotRow = append(gotRow, row)
	}
	if equal := stringutil.SliceElementsMatch(gotRow, rowInValid); !equal {
		return StepStateToContext(ctx, stepState), fmt.Errorf("invalid class not correct")
	}
	return StepStateToContext(ctx, stepState), nil
}
