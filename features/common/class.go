package common

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/try"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/gogo/protobuf/types"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/segmentio/ksuid"
)

func (s *suite) ACreateClassRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = &pb.CreateClassRequest{
		SchoolId:  0,
		ClassName: "",
		Grades:    []string{"G10", "G11"},
		Subjects:  []pb.Subject{pb.SUBJECT_BIOLOGY, pb.SUBJECT_CHEMISTRY, pb.SUBJECT_LITERATURE},
		OwnerId:   "",
		OwnerIds:  []string{},
	}
	return StepStateToContext(ctx, stepState), nil
}

//nolint
func (s *suite) ASchoolIdInCreateClassRequest(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	schoolID := int32(0)

	if arg1 == "valid" {
		if stepState.Schools != nil && len(stepState.Schools) != 0 {
			schoolID = stepState.Schools[0].ID.Int
		} else {
			schoolID = 1
		}
	} else {
		if n, err := strconv.Atoi(arg1); err == nil {
			schoolID = int32(n)
		}
	}
	stepState.CurrentSchoolID = schoolID
	stepState.Request.(*pb.CreateClassRequest).SchoolId = schoolID
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) AValidNameInCreateClassRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request.(*pb.CreateClassRequest).ClassName = "class-name" + ksuid.New().String()
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserCreateAClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()
	ctx, err := s.createClassUpsertedSubscribe(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createClassUpsertedSubscribe: %w", err)
	}
	stepState.Response, stepState.ResponseErr = pb.NewClassClient(s.BobConn).CreateClass(contextWithToken(s, ctx), stepState.Request.(*pb.CreateClassRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createClassUpsertedSubscribe(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}
	handlerClassUpsertedSubscription := func(ctx context.Context, data []byte) (bool, error) {
		r := &pb.EvtClassRoom{}
		err := r.Unmarshal(data)
		if err != nil {
			return false, err
		}
		switch r.Message.(type) {
		case *pb.EvtClassRoom_CreateClass_:
			stepState.FoundChanForJetStream <- r.Message
			return false, nil
		case *pb.EvtClassRoom_ActiveConversation_:
			stepState.FoundChanForJetStream <- r.Message
			return false, nil
		case *pb.EvtClassRoom_EditClass_:
			stepState.FoundChanForJetStream <- r.Message
			return false, nil
		case *pb.EvtClassRoom_JoinClass_:
			stepState.FoundChanForJetStream <- r.Message
			return false, nil
		case *pb.EvtClassRoom_LeaveClass_:
			stepState.FoundChanForJetStream <- r.Message
			return false, nil
		default:
			return true, errors.New("handlerClassUpsertedSubscription: wrong message")
		}
	}
	sub, err := s.JSM.Subscribe(constants.SubjectClassUpserted, opts, handlerClassUpsertedSubscription)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.JSM.Subscribe: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) BobMustCreateClassFromCreateClassRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.CreateClassRequest)
	resp := stepState.Response.(*pb.CreateClassResponse)

	class := &bob_entities.Class{}
	fields, values := class.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM classes WHERE class_id = $1", strings.Join(fields, ", "))

	err := s.BobDB.QueryRow(ctx, query, resp.ClassId).Scan(values...)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.DB.QueryRow: %w", err)
	}

	if class.Code.String == "" {
		return StepStateToContext(ctx, stepState), errors.New("bob does not create class code")
	}
	if req.ClassName != class.Name.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("class name not match expect %s, got %s", req.ClassName, class.Name.String)
	}
	if req.SchoolId != class.SchoolID.Int {
		return StepStateToContext(ctx, stepState), fmt.Errorf("school id not match expect %v, got %v", req.SchoolId, class.SchoolID.Int)
	}

	var respSubject []string
	class.Subjects.AssignTo(&respSubject)
	var reqSubject = make([]string, 0, len(req.Subjects))
	for _, v := range req.Subjects {
		reqSubject = append(reqSubject, v.String())
	}
	if !reflect.DeepEqual(respSubject, reqSubject) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("subjects not match expect %v, got %v", reqSubject, respSubject)
	}

	var respGrade []int
	class.Grades.AssignTo(&respGrade)
	var reqGrade = make([]int, 0, len(req.Grades))
	for _, v := range req.Grades {
		gradeInt, err := i18n.ConvertStringGradeToInt(pb.COUNTRY_VN, v)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("i18n.ConvertStringGradeToInt: %w", err)
		}
		reqGrade = append(reqGrade, gradeInt)
	}
	if !reflect.DeepEqual(respGrade, reqGrade) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("grades not match expect %v, got %v", respGrade, reqGrade)
	}

	stepState.CurrentClassCode = class.Code.String
	stepState.CurrentClassID = class.ID.Int

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ClassMustHasIs(ctx context.Context, key, value string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	conds := fmt.Sprintf("%s = '%s'", key, value)
	if value == "NULL" {
		conds = key + " IS NULL"
	}
	query := "SELECT COUNT(*) FROM classes WHERE class_id = $1 AND " + conds
	count := 0
	if err := try.Do(func(attempt int) (bool, error) {
		err := s.BobDB.QueryRow(ctx, query, stepState.CurrentClassID).Scan(&count)
		if err != nil {
			return false, err
		}

		if count == 0 {
			time.Sleep(350 * time.Millisecond)
			return attempt < 5, fmt.Errorf("not found class has %s is %s", key, value)
		}

		return false, nil
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if count == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not found class has %s is %s", key, value)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ClassMustHaveMemberIsAndIsOwnerAndStatus(ctx context.Context, total int, group, isOwner, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	count := 0
	err := s.BobDB.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM class_members WHERE class_id = $1 AND user_group = $2 AND is_owner = $3 AND status = $4",
		stepState.CurrentClassID, group, isOwner, status).Scan(&count)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if total != count {
		return StepStateToContext(ctx, stepState), fmt.Errorf("class member not match result expect %d, got %d", total, count)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) AJoinClassRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = &pb.JoinClassRequest{}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) AClassCodeInJoinClassRequest(ctx context.Context, arg string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if arg == "valid" {
		stepState.Request.(*pb.JoinClassRequest).ClassCode = stepState.CurrentClassCode
	}
	if arg == "wrong" {
		stepState.Request.(*pb.JoinClassRequest).ClassCode = "$1111111"
	}

	if stepState.CurrentStudentID == "" {
		t, _ := jwt.ParseString(stepState.AuthToken)
		stepState.CurrentStudentID = t.Subject()
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserJoinAClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()
	ctx, err := s.createClassUpsertedSubscribe(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createClassUpsertedSubscribe: %w", err)
	}
	stepState.Response, stepState.ResponseErr = pb.NewClassClient(s.BobConn).JoinClass(contextWithToken(s, ctx), stepState.Request.(*pb.JoinClassRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UpsertValidMediaList(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	n := rand.Intn(20) + 5
	mediaList := make([]*pb.Media, 0, n)
	for i := 0; i < n; i++ {
		var mediaType pb.MediaType
		switch {
		case int32(i)%int32(pb.MEDIA_TYPE_AUDIO) == 0:
			mediaType = pb.MEDIA_TYPE_AUDIO
		case int32(i)%int32(pb.MEDIA_TYPE_PDF) == 0:
			mediaType = pb.MEDIA_TYPE_PDF
		case int32(i)%int32(pb.MEDIA_TYPE_IMAGE) == 0:
			mediaType = pb.MEDIA_TYPE_IMAGE
		default:
			mediaType = pb.MEDIA_TYPE_VIDEO
		}
		mediaList = append(mediaList, s.GenerateMediaWithType(stepState.Random, mediaType))
	}
	req := &pb.UpsertMediaRequest{
		Media: mediaList,
	}
	stepState.RequestSentAt = time.Now()
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewClassClient(s.BobConn).UpsertMedia(contextWithToken(s, ctx), req)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	resp := stepState.Response.(*pb.UpsertMediaResponse)
	stepState.MediaIDs = resp.MediaIds
	for i, mediaID := range resp.MediaIds {
		mediaList[i].MediaId = mediaID
	}
	stepState.Medias = append(stepState.Medias, mediaList...)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) GenerateMediaWithType(randStr string, mediaType pb.MediaType) *pb.Media {
	return &pb.Media{
		MediaId:   "",
		Name:      fmt.Sprintf("random-name-%s", randStr),
		Resource:  s.newID(),
		CreatedAt: types.TimestampNow(),
		UpdatedAt: types.TimestampNow(),
		Comments: []*pb.Comment{
			{Comment: "Comment-1", Duration: types.DurationProto(10 * time.Second)},
			{Comment: "Comment-2", Duration: types.DurationProto(20 * time.Second)},
		},
		Type: mediaType,
	}
}
