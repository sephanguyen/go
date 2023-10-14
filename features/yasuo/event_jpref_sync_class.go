package yasuo

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	enigma_entites "github.com/manabie-com/backend/internal/enigma/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	master_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure/repo"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/proto"
)

func (s *suite) AvalidAcademicYear() error {
	e := &entities_bob.AcademicYear{}
	err := multierr.Combine(
		e.ID.Set("2021"),
		e.SchoolID.Set(constants.JPREPSchool),
		e.Name.Set("2021"),
		e.StartYearDate.Set(time.Now()),
		e.EndYearDate.Set(time.Now().Add(200*24*time.Hour)),
		e.Status.Set(entities_bob.AcademicYearStatusActive),
	)
	if err != nil {
		return err
	}

	aRepo := &repositories.AcademicYearRepo{}
	err = aRepo.Create(context.Background(), s.DBTrace, e)
	if err != nil {
		return err
	}

	s.AcademicID = e.ID.String

	return nil
}

func (s *suite) JPREPSyncClassWithAction(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	now := time.Now()
	stepState.CurrentClassID = int32(rand.Intn(999999999))
	classes := []*npb.EventMasterRegistration_Class{{
		ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
		ClassName:  "class name " + idutil.ULIDNow(),
		ClassId:    uint64(stepState.CurrentClassID),
		CourseId:   stepState.CurrentCourseID,
		StartDate: &timestamp.Timestamp{
			Seconds: now.Unix(),
		},
		EndDate: &timestamp.Timestamp{
			Seconds: now.Unix(),
		},
		AcademicYearId: s.AcademicID,
	}}

	stepState.RequestSentAt = now
	stepState.Request = classes
	req := &npb.EventMasterRegistration{RawPayload: []byte("{}"), Signature: idutil.ULIDNow(), Classes: classes}
	data, _ := proto.Marshal(req)
	_, err := s.JSM.PublishContext(ctx, constants.SubjectSyncMasterRegistration, data)
	if err != nil {
		return ctx, fmt.Errorf("Publish: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) toClassSyncMsg(ctx context.Context, class, actionKind string, total int) (context.Context, []*npb.EventMasterRegistration_Class, error) {
	stepState := StepStateFromContext(ctx)

	classes := []*npb.EventMasterRegistration_Class{}

	e := &entities_bob.AcademicYear{}
	err := multierr.Combine(
		e.ID.Set("2021"),
		e.SchoolID.Set(constants.JPREPSchool),
		e.Name.Set("2021"),
		e.StartYearDate.Set(time.Now()),
		e.EndYearDate.Set(time.Now().Add(200*24*time.Hour)),
		e.Status.Set(entities_bob.AcademicYearStatusActive),
	)
	if err != nil {
		return ctx, nil, err
	}

	aRepo := &repositories.AcademicYearRepo{}
	err = aRepo.Create(ctx, s.DBTrace, e)
	if err != nil {
		return ctx, nil, err
	}

	ctx, err = s.aSignedIn(ctx, "school admin")
	if err != nil {
		return ctx, nil, err
	}

	ctx, err = s.aValidCourse(ctx)
	if err != nil {
		return ctx, nil, err
	}

	switch class {
	case "new class":
		now := time.Now()
		for i := 0; i < total; i++ {
			classes = append(classes, &npb.EventMasterRegistration_Class{
				ActionKind: npb.ActionKind(npb.ActionKind_value[actionKind]),
				ClassName:  "class name " + idutil.ULIDNow(),
				ClassId:    uint64(rand.Intn(999999999)),
				CourseId:   stepState.CurrentCourseID,
				StartDate: &timestamp.Timestamp{
					Seconds: now.Unix(),
				},
				EndDate: &timestamp.Timestamp{
					Seconds: now.Unix(),
				},
				AcademicYearId: e.ID.String,
			})
		}
	case "existed class":
		classRepo := &repositories.ClassRepo{}
		for i := 0; i < total; i++ {
			class := &entities_bob.Class{}
			now := time.Now()
			database.AllNullEntity(class)
			err = multierr.Combine(
				class.ID.Set(uint64(rand.Intn(999999999))),
				class.Name.Set("class name "+idutil.ULIDNow()),
				class.SchoolID.Set(constants.JPREPSchool),
				class.Avatar.Set("avatar"),
				class.Status.Set(entities_bob.ClassStatusActive),
				class.Country.Set("COUNTRY_JP"),
			)
			if err != nil {
				return ctx, nil, fmt.Errorf("err set class: %w", err)
			}

			err = classRepo.Create(ctx, s.DBTrace, class)
			if err != nil {
				return ctx, nil, fmt.Errorf("err Insert: %w", err)
			}

			classes = append(classes, &npb.EventMasterRegistration_Class{
				ActionKind:     npb.ActionKind(npb.ActionKind_value[actionKind]),
				ClassName:      "update class name " + idutil.ULIDNow(),
				ClassId:        uint64(class.ID.Int),
				CourseId:       stepState.CurrentCourseID,
				StartDate:      &timestamp.Timestamp{Seconds: now.Unix()},
				EndDate:        &timestamp.Timestamp{Seconds: now.Unix()},
				AcademicYearId: e.ID.String,
			})
		}
	}

	return StepStateToContext(ctx, stepState), classes, nil
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
		case *pb.EvtClassRoom_JoinClass_:
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

func (s *suite) jprepSyncClassWithActionAndClassWithAction(ctx context.Context, numberOfNewClass, newClassAction, numberOfExistedClass, existedClassAction string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aSignedIn(ctx, "school admin")
	if err != nil {
		return nil, err
	}

	_, err = s.createClassUpsertedSubscribe(s.signedCtx(ctx))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createClassUpsertedSubscribe: %w", err)
	}

	total, err := strconv.Atoi(numberOfNewClass)
	if err != nil {
		return ctx, err
	}
	stepState.RequestSentAt = time.Now()

	_, newClasses, err := s.toClassSyncMsg(s.signedCtx(ctx), "new class", newClassAction, total)
	if err != nil {
		return ctx, err
	}

	total, err = strconv.Atoi(numberOfExistedClass)
	if err != nil {
		return ctx, err
	}

	_, existedClasses, err := s.toClassSyncMsg(s.signedCtx(ctx), "existed class", existedClassAction, total)
	if err != nil {
		return ctx, err
	}

	classes := append(newClasses, existedClasses...)
	stepState.Request = classes
	signature := idutil.ULIDNow()
	_, err = s.createPartnerSyncDataLog(s.signedCtx(ctx), signature, 0)
	if err != nil {
		return ctx, fmt.Errorf("create partner sync data log error: %w", err)
	}
	_, err = s.createLogSyncDataSplit(s.signedCtx(ctx), string(enigma_entites.KindClass))
	if err != nil {
		return ctx, fmt.Errorf("create partner sync data log split error: %w", err)
	}

	req := &npb.EventMasterRegistration{
		RawPayload: []byte("{}"),
		Signature:  signature,
		Classes:    classes,
		LogId:      stepState.PartnerSyncDataLogSplitId,
	}
	data, _ := proto.Marshal(req)
	_, err = s.JSM.PublishContext(s.signedCtx(ctx), constants.SubjectSyncMasterRegistration, data)
	if err != nil {
		return ctx, fmt.Errorf("Publish: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) TheseClassesMustBeStoreInOutSystem(ctx context.Context) (context.Context, error) {
	return s.theseClassesMustBeStoreInOutSystem(ctx)
}

func (s *suite) theseNewClassesMustBeStoreInOutSystem(ctx context.Context) (context.Context, error) {
	time.Sleep(time.Second)

	stepState := StepStateFromContext(ctx)
	classRepo := &master_repo.ClassRepo{}

	classes := stepState.Request.([]*npb.EventMasterRegistration_Class)
	for _, c := range classes {
		class, err := classRepo.GetByID(s.signedCtx(ctx), s.DBTrace, strconv.Itoa(int(c.ClassId)))
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return ctx, fmt.Errorf("err find class: %w", err)
		}

		if c.ActionKind == npb.ActionKind_ACTION_KIND_UPSERTED {
			if class == nil {
				return ctx, fmt.Errorf("class does not existed")
			}

			if class.Name != c.ClassName {
				return ctx, fmt.Errorf("class name does not match, expected: %s, got: %s", c.ClassName, class.Name)
			}

			if class.CourseID != c.CourseId {
				return ctx, fmt.Errorf("course id does not match, expected: %s, got: %s", c.CourseId, class.CourseID)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theseClassesMustBeStoreInOutSystem(ctx context.Context) (context.Context, error) {
	time.Sleep(time.Second)

	stepState := StepStateFromContext(ctx)

	classRepo := &repositories.ClassRepo{}
	courseClassRepo := &repositories.CourseClassRepo{}

	classes := stepState.Request.([]*npb.EventMasterRegistration_Class)
	for _, c := range classes {
		class, err := classRepo.FindByID(s.signedCtx(ctx), s.DBTrace, database.Int4(int32(c.ClassId)))
		if err != nil {
			return ctx, fmt.Errorf("err find class: %w", err)
		}

		stepState.CurrentClassID = class.ID.Int

		switch c.ActionKind {
		case npb.ActionKind_ACTION_KIND_UPSERTED:
			if class == nil {
				return ctx, fmt.Errorf("class does not existed")
			}

			if class.Name.String != c.ClassName {
				return ctx, fmt.Errorf("class name does not match, expected: %s, got: %s", c.ClassName, class.Name.String)
			}

			if class.Country.String != "COUNTRY_JP" {
				return ctx, fmt.Errorf("class country does not match, expected: %s, got: %s", "COUNTRY_JP", class.Country.String)
			}

			// check course_class
			mapByClass, err := courseClassRepo.Find(s.signedCtx(ctx), s.DBTrace, database.Int4Array([]int32{int32(c.ClassId)}))
			if err != nil {
				return ctx, fmt.Errorf("err find Class Course")
			}

			courses, ok := mapByClass[class.ID]
			if !ok {
				return ctx, fmt.Errorf("not found any courses")
			}

			found := false
			for _, course := range courses.Elements {
				if course.String == c.CourseId {
					found = true
				}
			}

			if !found {
				return ctx, fmt.Errorf("course id not found")
			}

			if !strings.Contains(c.ClassName, "update") {
				ctx, err1 := s.bobMustPushMsgSubjectToNats(s.signedCtx(ctx), "CreateClass", constants.SubjectClassUpserted)
				ctx, err2 := s.bobMustPushMsgSubjectToNats(s.signedCtx(ctx), "ActiveConversation", constants.SubjectClassUpserted)

				err = multierr.Append(err1, err2)
				if err != nil {
					return ctx, err
				}
			}

			count := 0
			query := `SELECT COUNT(*) FROM courses_academic_years WHERE course_id = $1 AND academic_year_id = $2`
			s.DBTrace.QueryRow(ctx, query, c.CourseId, c.AcademicYearId).Scan(&count)

			if count == 0 {
				return ctx, fmt.Errorf("academicYearId does not match, expected %v", c.AcademicYearId)
			}

		case npb.ActionKind_ACTION_KIND_DELETED:
			// check course_class deleted
			if class.Status.String == entities_bob.ClassStatusActive {
				return ctx, fmt.Errorf("class does not deleted, still active")
			}

			// check course_class
			mapByClass, err := courseClassRepo.Find(ctx, s.DBTrace, database.Int4Array([]int32{int32(c.ClassId)}))
			if err != nil {
				return ctx, fmt.Errorf("err find Class Course")
			}

			if len(mapByClass) != 0 {
				return ctx, fmt.Errorf("course class does not deleted")
			}

			ctx, err = s.bobMustPushMsgSubjectToNats(ctx, "InActiveConversation", constants.SubjectClassUpserted)
			if err != nil {
				return ctx, err
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) toClassMemberSyncMsg(ctx context.Context, status, actionKind string, total int) (context.Context, []*npb.EventUserRegistration_Student, error) {
	if total == 0 {
		return ctx, []*npb.EventUserRegistration_Student{}, nil
	}

	stepState := StepStateFromContext(ctx)

	students := []*npb.EventUserRegistration_Student{}

	classRepo := &repositories.ClassRepo{}
	masterClassRepo := &master_repo.ClassRepo{}

	class := &entities_bob.Class{}
	database.AllNullEntity(class)
	err := multierr.Combine(
		class.ID.Set(uint64(rand.Intn(999999999))),
		class.Name.Set("class name "+idutil.ULIDNow()),
		class.SchoolID.Set(constants.JPREPSchool),
		class.Avatar.Set("avatar"),
		class.Status.Set(entities_bob.ClassStatusActive),
		class.Country.Set("COUNTRY_JP"),
		class.Code.Set(entities_bob.GenerateClassCode(8)),
		class.PlanID.Set("School"),
		class.PlanDuration.Set("30"),
	)
	masterClasses := make([]*domain.Class, 0, 1)
	if err != nil {
		return ctx, nil, fmt.Errorf("err set class: %w", err)
	}

	err = classRepo.Create(ctx, s.DBTrace, class)
	if err != nil {
		return ctx, nil, fmt.Errorf("err Insert: %w", err)
	}
	masterClasses = append(masterClasses, &domain.Class{
		ClassID:    strconv.Itoa(int(class.ID.Int)),
		Name:       "class name " + idutil.ULIDNow(),
		LocationID: constants.JPREPOrgLocation,
		CourseID:   stepState.CourseIds[0],
	})
	err = masterClassRepo.UpsertClasses(ctx, s.DBTrace, masterClasses)
	if err != nil {
		return ctx, nil, fmt.Errorf("err Insert new class: %w", err)
	}
	switch status {
	case "new class member":
		for i := 0; i < total; i++ {
			ctx, err := s.aSignedIn(ctx, "student")
			if err != nil {
				return ctx, nil, err
			}

			students = append(students, &npb.EventUserRegistration_Student{
				ActionKind: npb.ActionKind(npb.ActionKind_value[actionKind]),
				StudentId:  stepState.CurrentUserID,
				Packages: []*npb.EventUserRegistration_Student_Package{
					{
						ClassId: int64(class.ID.Int),
					},
				},
			})
		}
	case "existed class member":
		stepState.Request = nil
		ctx, err1 := s.jprepSyncClassMembersWithActionAndClassMembersWithAction(ctx, strconv.Itoa(total), npb.ActionKind_ACTION_KIND_UPSERTED.String(), "0", "")
		ctx, err2 := s.theseClassMembersMustBeStoreInOutSystem(ctx)
		err := multierr.Combine(err1, err2)

		if err != nil {
			return ctx, nil, err
		}

		for _, s := range stepState.Request.([]*npb.EventUserRegistration_Student) {
			s.ActionKind = npb.ActionKind(npb.ActionKind_value[actionKind])
			students = append(students, s)
		}
	}

	return StepStateToContext(ctx, stepState), students, nil
}

func (s *suite) jprepSyncClassMembersWithActionAndClassMembersWithAction(ctx context.Context, numberOfNewClassMembers, newClassMemberAction, numberOfExistedClassMembers, existedClassMemberAction string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentSchoolID = constants.JPREPSchool

	ctx, err := s.aSignedIn(ctx, "school admin")
	if err != nil {
		return ctx, err
	}

	ctx, err = s.createClassUpsertedSubscribe(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createClassUpsertedSubscribe: %w", err)
	}

	total, err := strconv.Atoi(numberOfNewClassMembers)
	if err != nil {
		return ctx, err
	}
	stepState.RequestSentAt = time.Now()

	ctx, newClassesMembers, err := s.toClassMemberSyncMsg(ctx, "new class member", newClassMemberAction, total)
	if err != nil {
		return ctx, err
	}

	total, err = strconv.Atoi(numberOfExistedClassMembers)
	if err != nil {
		return ctx, err
	}

	ctx, existedClassMembers, err := s.toClassMemberSyncMsg(ctx, "existed class member", existedClassMemberAction, total)
	if err != nil {
		return ctx, err
	}

	students := append(newClassesMembers, existedClassMembers...)
	stepState.Request = students

	signature := idutil.ULIDNow()
	ctx, err = s.createPartnerSyncDataLog(ctx, signature, 0)
	if err != nil {
		return ctx, fmt.Errorf("create partner sync data log error: %w", err)
	}
	ctx, err = s.createLogSyncDataSplit(ctx, string(enigma_entites.KindStudent))
	if err != nil {
		return ctx, fmt.Errorf("create partner sync data log split error: %w", err)
	}

	req := &npb.EventUserRegistration{
		RawPayload: []byte("{}"),
		Signature:  signature,
		Students:   students,
		LogId:      stepState.PartnerSyncDataLogSplitId,
	}
	data, _ := proto.Marshal(req)
	_, err = s.JSM.PublishContext(ctx, constants.SubjectUserRegistrationNatsJS, data)
	if err != nil {
		return ctx, fmt.Errorf("Publish: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) TheseClassMembersMustBeStoreInOutSystem(ctx context.Context) (context.Context, error) {
	return s.theseClassMembersMustBeStoreInOutSystem(ctx)
}

func (s *suite) theseClassMembersMustBeStoreInOutSystem(ctx context.Context) (context.Context, error) {
	time.Sleep(time.Second)

	stepState := StepStateFromContext(ctx)

	classMemberRepo := &repositories.ClassMemberRepo{}

	students := stepState.Request.([]*npb.EventUserRegistration_Student)
	stepState.CurrentClassIDs = []int32{}
	for _, student := range students {
		status := entities_bob.ClassMemberStatusActive
		if student.ActionKind == npb.ActionKind_ACTION_KIND_DELETED {
			status = entities_bob.ClassMemberStatusInactive
		}

		for _, p := range student.Packages {
			member, err := classMemberRepo.Get(ctx, s.DBTrace,
				database.Int4(int32(p.ClassId)),
				database.Text(student.StudentId),
				database.Text(status))
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return ctx, fmt.Errorf("err find class: %w", err)
			}

			if member == nil {
				return ctx, fmt.Errorf("not found member in class")
			}

			stepState.CurrentClassIDs = append(stepState.CurrentClassIDs, int32(p.ClassId))
		}
	}
	ctx, err := s.bobMustPushMsgSubjectToNats(ctx, "JoinClass", constants.SubjectClassUpserted)
	if err != nil {
		return ctx, fmt.Errorf("pushing message failed: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theseClassMembersNewMustBeStoreInOutSystem(ctx context.Context) (context.Context, error) {
	time.Sleep(time.Second)

	stepState := StepStateFromContext(ctx)

	classMemberRepo := &master_repo.ClassMemberRepo{}

	students := stepState.Request.([]*npb.EventUserRegistration_Student)
	for _, student := range students {
		for _, p := range student.Packages {
			member, err := classMemberRepo.GetByClassIDAndUserID(ctx, s.DBTrace,
				strconv.Itoa(int(p.ClassId)),
				student.StudentId)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return ctx, fmt.Errorf("err find class member: %w", err)
			}
			if student.ActionKind == npb.ActionKind_ACTION_KIND_DELETED {
				if member != nil {
					return ctx, fmt.Errorf("expect new class member deleted, actually not")
				}
			} else {
				if member == nil {
					return ctx, fmt.Errorf("not found member in new class")
				}
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) JprepSyncClassMembersWithAction(ctx context.Context, actionKind string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()
	students := []*npb.EventUserRegistration_Student{{
		ActionKind: npb.ActionKind(npb.ActionKind_value[actionKind]),
		StudentId:  stepState.CurrentUserID,
		Packages: []*npb.EventUserRegistration_Student_Package{
			{
				ClassId: int64(stepState.CurrentClassID),
			},
		},
	}}

	stepState.Request = students
	req := &npb.EventUserRegistration{RawPayload: []byte("{}"), Signature: idutil.ULIDNow(), Students: students}
	data, _ := proto.Marshal(req)
	_, err := s.JSM.PublishContext(ctx, constants.SubjectUserRegistrationNatsJS, data)
	if err != nil {
		return ctx, fmt.Errorf("Publish: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) AClass(ctx context.Context) (context.Context, error) {
	return s.aClass(ctx)
}

func (s *suite) aClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentClassID = int32(rand.Intn(999999999))

	classRepo := &repositories.ClassRepo{}

	class := &entities_bob.Class{}
	database.AllNullEntity(class)

	err := multierr.Combine(
		class.ID.Set(stepState.CurrentClassID),
		class.SchoolID.Set(stepState.CurrentSchoolID),
		class.Name.Set(idutil.ULIDNow()),
		class.Subjects.Set([]string{pb.SUBJECT_BIOLOGY.String(), pb.SUBJECT_CHEMISTRY.String()}),
		class.Status.Set(entities_bob.ClassStatusActive),
		class.Avatar.Set("avatar"),
		class.PlanID.Set("School"),
		class.Country.Set("COUNTRY_VN"),
		class.Code.Set(entities_bob.GenerateClassCode(7)),
	)
	if err != nil {
		return ctx, fmt.Errorf("class combine: %w", err)
	}
	err = classRepo.Create(ctx, s.DBTrace, class)
	if err != nil {
		return ctx, fmt.Errorf("classRepo.Create: %w", err)
	}
	classMemberRepo := &repositories.ClassMemberRepo{}
	classMember, err := s.generateAClassMember(stepState.CurrentTeacherID, entities_bob.UserGroupTeacher, stepState.CurrentClassID, true)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = classMemberRepo.Create(ctx, s.DB, classMember)
	if err != nil {
		return ctx, fmt.Errorf("classMemberRepo.Create: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateAClassMember(userID, group string, classID int32, isOwner bool) (*entities_bob.ClassMember, error) {
	classMember := new(entities_bob.ClassMember)
	database.AllNullEntity(classMember)
	err := multierr.Combine(classMember.ID.Set(idutil.ULIDNow()), classMember.ClassID.Set(classID), classMember.UserID.Set(userID), classMember.UserGroup.Set(group), classMember.IsOwner.Set(isOwner), classMember.Status.Set(entities_bob.ClassMemberStatusActive))
	if err != nil {
		return nil, fmt.Errorf("generateAClassMember.SetEntity: %w", err)
	}
	return classMember, nil
}
