package bob

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/try"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/jackc/pgtype"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
)

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

func (s *suite) aCreateClassRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = &pb.CreateClassRequest{
		SchoolId:  stepState.CurrentSchoolID,
		ClassName: "",
		Grades:    []string{"G10", "G11"},
		Subjects:  []pb.Subject{pb.SUBJECT_BIOLOGY, pb.SUBJECT_CHEMISTRY, pb.SUBJECT_LITERATURE},
		OwnerId:   "",
		OwnerIds:  []string{},
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) ACreateClassRequest(ctx context.Context) (context.Context, error) {
	return s.aCreateClassRequest(ctx)
}
func (s *suite) createClassRequestHasGradeIsAndSubjectIs(ctx context.Context, grade, subject string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request.(*pb.CreateClassRequest).Grades = []string{grade}
	stepState.Request.(*pb.CreateClassRequest).Subjects = []pb.Subject{pb.Subject(pb.Subject_value[subject])}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userCreateAClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = pb.NewClassClient(s.Conn).CreateClass(contextWithToken(s, ctx), stepState.Request.(*pb.CreateClassRequest))
	stepState.RequestSentAt = time.Now()
	ctx, err := s.createClassUpsertedSubscribe(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createClassUpsertedSubscribe: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) UserCreateAClass(ctx context.Context) (context.Context, error) {
	return s.userCreateAClass(ctx)
}

func (s *suite) ASignedInTeacher(ctx context.Context) (context.Context, error) {
	return s.aSignedInTeacher(ctx)
}
func (s *suite) bobMustCreateClassFromCreateClassRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.CreateClassRequest)
	resp := stepState.Response.(*pb.CreateClassResponse)

	class := &entities.Class{}
	fields, values := class.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM classes WHERE class_id = $1", strings.Join(fields, ", "))

	err := s.DB.QueryRow(ctx, query, resp.ClassId).Scan(values...)
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

	var respSubject, reqSubject []string
	class.Subjects.AssignTo(&respSubject)
	for _, v := range req.Subjects {
		reqSubject = append(reqSubject, v.String())
	}
	if !reflect.DeepEqual(respSubject, reqSubject) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("subjects not match expect %v, got %v", reqSubject, respSubject)
	}

	var respGrade, reqGrade []int
	class.Grades.AssignTo(&respGrade)
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
func (s *suite) BobMustCreateClassFromCreateClassRequest(ctx context.Context) (context.Context, error) {
	return s.bobMustCreateClassFromCreateClassRequest(ctx)
}
func (s *suite) classMustHaveMemberIsAndIsOwnerAndStatus(ctx context.Context, total int, group, isOwner, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	count := 0
	err := s.DB.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM class_members WHERE class_id = $1 AND user_group = $2 AND is_owner = $3 AND status = $4",
		stepState.CurrentClassID, group, isOwner, status).Scan(&count)

	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	if total != count {
		return StepStateToContext(ctx, stepState), errors.Errorf("class member not match result expect %d, got %d", total, count)
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) ClassMustHaveMemberIsAndIsOwnerAndStatus(ctx context.Context, total int, group, isOwner, status string) (context.Context, error) {
	return s.classMustHaveMemberIsAndIsOwnerAndStatus(ctx, total, group, isOwner, status)
}
func (s *suite) aSchoolIdInCreateClassRequest(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	schoolID := int32(0)

	if arg1 == "valid" {
		schoolID = constants.ManabieSchool
	} else {
		if n, err := strconv.Atoi(arg1); err == nil {
			schoolID = int32(n)
		}
	}
	stepState.CurrentSchoolID = schoolID
	stepState.Request.(*pb.CreateClassRequest).SchoolId = schoolID
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) ASchoolIdInCreateClassRequest(ctx context.Context, arg1 string) (context.Context, error) {
	return s.aSchoolIdInCreateClassRequest(ctx, arg1)
}
func (s *suite) aValidNameInCreateClassRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request.(*pb.CreateClassRequest).ClassName = "class-name" + ksuid.New().String()
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) AValidNameInCreateClassRequest(ctx context.Context) (context.Context, error) {
	return s.aValidNameInCreateClassRequest(ctx)
}
func (s *suite) aOwnerIdWithSchoolIdIsInCreateClassRequest(ctx context.Context, number int, role string, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if role == "" {
		return StepStateToContext(ctx, stepState), nil
	}

	authToken := stepState.AuthToken
	ownerIDs := stepState.Request.(*pb.CreateClassRequest).OwnerIds

	for number > 0 {
		ctx, err := s.aSignedInWithSchool(ctx, role, schoolID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}

		t, _ := jwt.ParseString(stepState.AuthToken)
		ownerIDs = append(ownerIDs, t.Subject())

		number--
	}

	stepState.Request.(*pb.CreateClassRequest).OwnerIds = ownerIDs
	stepState.AuthToken = authToken
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) AOwnerIdWithSchoolIdIsInCreateClassRequest(ctx context.Context, number int, role string, schoolID int) (context.Context, error) {
	return s.aOwnerIdWithSchoolIdIsInCreateClassRequest(ctx, number, role, schoolID)
}
func (s *suite) defaultConfigForHasIs(ctx context.Context, group, key, value string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	_, err := s.DB.Exec(ctx, `INSERT INTO configs VALUES
	($1, $2, 'COUNTRY_VN', $3, now(), now())
	ON CONFLICT  ON CONSTRAINT config_pk 
	DO UPDATE SET config_value = $3;`, key, group, value)

	return StepStateToContext(ctx, stepState), err
}
func (s *suite) DefaultConfigForHasIs(ctx context.Context, group, key, value string) (context.Context, error) {
	return s.defaultConfigForHasIs(ctx, group, key, value)
}
func (s *suite) classMustHasIs(ctx context.Context, key, value string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	conds := fmt.Sprintf("%s = '%s'", key, value)
	if value == "NULL" {
		conds = key + " IS NULL"
	}
	query := "SELECT COUNT(*) FROM classes WHERE class_id = $1 AND " + conds
	count := 0
	if err := try.Do(func(attempt int) (bool, error) {
		err := s.DB.QueryRow(ctx, query, stepState.CurrentClassID).Scan(&count)
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
func (s *suite) ClassMustHasIs(ctx context.Context, key, value string) (context.Context, error) {
	return s.classMustHasIs(ctx, key, value)
}
func (s *suite) thisSchoolHasConfigIsIsIs(ctx context.Context, planField, planValue, expiredAtField, expiredAtValue, durationField string, durationValue int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var expiredAt interface{} = expiredAtValue
	if expiredAtValue == "NULL" {
		var pgExpiredAtValue pgtype.Timestamptz
		_ = pgExpiredAtValue.Set(nil)
		expiredAt = &pgExpiredAtValue
	}
	_, err := s.DB.Exec(ctx, `INSERT INTO school_configs VALUES
	($1, $2, 'COUNTRY_VN', $3, $4, now(), now())
	ON CONFLICT  ON CONSTRAINT school_configs_pk 
	DO UPDATE SET plan_id = $2, plan_expired_at = $3, plan_duration = $4;`, stepState.CurrentSchoolID, planValue, expiredAt, durationValue)

	return StepStateToContext(ctx, stepState), err
}
func (s *suite) ThisSchoolHasConfigIsIsIs(ctx context.Context, planField, planValue, expiredAtField, expiredAtValue, durationField string, durationValue int) (context.Context, error) {
	return s.thisSchoolHasConfigIsIsIs(ctx, planField, planValue, expiredAtField, expiredAtValue, durationField, durationValue)
}
