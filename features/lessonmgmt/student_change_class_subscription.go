package lessonmgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	class_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Suite) createClassesAndAssignToCourses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courseIDs := stepState.CourseIDs
	locationIDs := stepState.CenterIDs
	totalClass := 4
	format := "%s,%s,%s,%s,%s"
	r1 := fmt.Sprintf(format, courseIDs[0], locationIDs[0], "", "", idutil.ULIDNow())
	r2 := fmt.Sprintf(format, courseIDs[0], locationIDs[1], "", "", idutil.ULIDNow())
	r3 := fmt.Sprintf(format, courseIDs[1], locationIDs[0], "", "", idutil.ULIDNow())
	r4 := fmt.Sprintf(format, courseIDs[1], locationIDs[1], "", "", idutil.ULIDNow())

	request := fmt.Sprintf(`course_id,location_id,course_name,location_name,class_name
	%s
	%s
	%s
	%s`, r1, r2, r3, r4)
	stepState.Request = &mpb.ImportClassRequest{
		Payload: []byte(request),
	}

	req := stepState.Request.(*mpb.ImportClassRequest)
	_, err := mpb.NewClassServiceClient(s.MasterMgmtConn).ImportClass(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createClassesAndAssignToCourses: %w", err)
	}
	stmt :=
		`
		SELECT 
			class_id,
			course_id,
			location_id
		FROM class
		WHERE deleted_at is null
		AND  course_id = ANY($1) 
		AND  location_id = ANY($2) 
		`
	rows, err := s.BobDBTrace.DB.Query(
		ctx,
		stmt,
		courseIDs,
		locationIDs,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query classes failed")
	}
	defer rows.Close()
	classImported := make([]*class_domain.Class, 0, totalClass)

	for rows.Next() {
		c := &class_domain.Class{}
		err := rows.Scan(
			&c.ClassID,
			&c.CourseID,
			&c.LocationID,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan class failed")
		}

		classImported = append(classImported, c)
	}
	stepState.ImportedClass = classImported
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) createSomeLessonsToClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now().Round(time.Second)
	startTime := now.AddDate(0, 0, -10)
	for _, class := range stepState.ImportedClass {
		req := &lpb.CreateLessonRequest{
			StartTime:      timestamppb.New(startTime),
			EndTime:        timestamppb.New(startTime.Add(5 * time.Hour)),
			TeachingMedium: cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
			TeacherIds:     stepState.TeacherIDs,
			SavingOption: &lpb.CreateLessonRequest_SavingOption{
				Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE,
				Recurrence: &lpb.Recurrence{
					EndDate: timestamppb.New(now.AddDate(0, 4, 0)),
				},
			},
			SchedulingStatus: lpb.LessonStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
			TeachingMethod:   cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP,
			ClassId:          class.ClassID,
			CourseId:         class.CourseID,
			LocationId:       class.LocationID,
		}
		_, err := s.CommonSuite.UserCreateALessonWithRequestInLessonmgmt(ctx, req)
		if err != nil {
			return nil, errors.WithMessage(err, "s.CommonSuite.UserCreateALessonWithRequestInLessonmgmt")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) genStudentPackageProfileWithCourseID(class *class_domain.Class, startAt, endAt time.Time) *upb.UpsertStudentCoursePackageRequest_StudentPackageProfile {
	return &upb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
		Id: &upb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
			CourseId: class.CourseID,
		},
		StartTime: timestamppb.New(startAt),
		EndTime:   timestamppb.New(endAt),
		StudentPackageExtra: []*upb.StudentPackageExtra{
			{
				LocationId: class.LocationID,
				ClassId:    class.ClassID,
			},
		},
	}
}

func (s *Suite) studentJoinSomeClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studentID := stepState.StudentIds[0]
	classes := stepState.ImportedClass
	studentPackageProfiles := make([]*upb.UpsertStudentCoursePackageRequest_StudentPackageProfile, 0, 10)

	joinedClass := make([]string, 0, len(classes))

	mapCourses := make(map[string]bool)

	for _, class := range classes {
		if _, ok := mapCourses[class.CourseID]; !ok {
			startAt := time.Now()
			endAt := startAt.AddDate(0, 2, 0)
			studentPackageProfiles = append(studentPackageProfiles, s.genStudentPackageProfileWithCourseID(class, startAt, endAt))
			mapCourses[class.CourseID] = true
			joinedClass = append(joinedClass, class.ClassID)
		}
	}
	stepState.JoinedClass = joinedClass

	req := &upb.UpsertStudentCoursePackageRequest{
		StudentId:              studentID,
		StudentPackageProfiles: studentPackageProfiles,
	}

	res, err := upb.NewUserModifierServiceClient(s.UserMgmtConn).UpsertStudentCoursePackage(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.WithMessage(err, "Upsert Student Course Package Fail class failed")
	}
	stepState.Response = res
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkStudentJoinClass(ctx context.Context) (context.Context, error) {
	// time sleep for lesson member sync before check
	time.Sleep(3 * time.Second)
	stepState := StepStateFromContext(ctx)
	studentID := stepState.StudentIds[0]
	stmt :=
		`
		SELECT 
			l.lesson_id,
			lm.user_id,
			l.course_id
		from
			lessons l
		JOIN lesson_members lm on l.lesson_id = lm.lesson_id
		WHERE lm.deleted_at is null
		AND  lm.user_id = $1 
		AND	 l.class_id = ANY($2)
		`
	rows, err := s.BobDBTrace.DB.Query(
		ctx,
		stmt,
		studentID,
		stepState.JoinedClass,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query lesson_members failed")
	}
	defer rows.Close()
	lessonMembers := make([]*domain.LessonMember, 0)

	for rows.Next() {
		lessonMember := &domain.LessonMember{}
		err := rows.Scan(
			&lessonMember.LessonID,
			&lessonMember.StudentID,
			&lessonMember.CourseID,
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), errors.WithMessage(err, "rows.Scan class failed")
		}

		lessonMembers = append(lessonMembers, lessonMember)
	}

	if len(lessonMembers) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lesson member")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) studentChangeOtherClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studentID := stepState.StudentIds[0]
	classes := stepState.ImportedClass

	studentPackage := stepState.Response.(*upb.UpsertStudentCoursePackageResponse)
	studentPackageProfiles := make([]*upb.UpsertStudentCoursePackageRequest_StudentPackageProfile, 0, 10)
	startAt := time.Now()
	endAt := startAt.AddDate(0, 2, 0)
	totalClass := len(classes)
	joinedClass := make([]string, 0, totalClass)
	leavedClass := make([]string, 0, totalClass)

	for _, studentPackage := range studentPackage.StudentPackageProfiles {
		for _, packageExtra := range studentPackage.StudentPackageExtra {
			for _, class := range classes {
				if class.CourseID == studentPackage.CourseId && class.ClassID != packageExtra.ClassId {
					studentPackageProfiles = append(studentPackageProfiles, &upb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
						Id: &upb.UpsertStudentCoursePackageRequest_StudentPackageProfile_StudentPackageId{
							StudentPackageId: studentPackage.StudentCoursePackageId,
						},
						StartTime: timestamppb.New(startAt),
						EndTime:   timestamppb.New(endAt),
						StudentPackageExtra: []*upb.StudentPackageExtra{
							{
								LocationId: class.LocationID,
								ClassId:    class.ClassID,
							},
						},
					})
					joinedClass = append(joinedClass, class.ClassID)
					leavedClass = append(leavedClass, packageExtra.ClassId)
					break
				}
			}
		}
	}
	stepState.JoinedClass = joinedClass
	stepState.LeavedClass = leavedClass
	req := &upb.UpsertStudentCoursePackageRequest{
		StudentId:              studentID,
		StudentPackageProfiles: studentPackageProfiles,
	}

	res, err := upb.NewUserModifierServiceClient(s.UserMgmtConn).UpsertStudentCoursePackage(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	if err != nil {
		return ctx, errors.WithMessage(err, "Upsert Student Course Package Fail class failed")
	}
	stepState.Response = res
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) studentChangeDurationClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studentID := stepState.StudentIds[0]

	studentPackage := stepState.Response.(*upb.UpsertStudentCoursePackageResponse)
	studentPackageProfiles := make([]*upb.UpsertStudentCoursePackageRequest_StudentPackageProfile, 0, 10)
	startAt := time.Now()
	endAt := startAt.AddDate(0, 1, 0)

	for _, studentPackage := range studentPackage.StudentPackageProfiles {
		for _, packageExtra := range studentPackage.StudentPackageExtra {
			studentPackageProfiles = append(studentPackageProfiles, &upb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
				Id: &upb.UpsertStudentCoursePackageRequest_StudentPackageProfile_StudentPackageId{
					StudentPackageId: studentPackage.StudentCoursePackageId,
				},
				StartTime: timestamppb.New(startAt),
				EndTime:   timestamppb.New(endAt),
				StudentPackageExtra: []*upb.StudentPackageExtra{
					{
						LocationId: packageExtra.LocationId,
						ClassId:    packageExtra.ClassId,
					},
				},
			})
		}
	}
	req := &upb.UpsertStudentCoursePackageRequest{
		StudentId:              studentID,
		StudentPackageProfiles: studentPackageProfiles,
	}

	_, err := upb.NewUserModifierServiceClient(s.UserMgmtConn).UpsertStudentCoursePackage(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	if err != nil {
		return ctx, errors.WithMessage(err, "Upsert Student Course Package Fail class failed")
	}
	stepState.EndDate = endAt
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkStudentLeaveClass(ctx context.Context) (context.Context, error) {
	// time sleep for lesson member sync before check
	time.Sleep(3 * time.Second)
	stepState := StepStateFromContext(ctx)
	studentID := stepState.StudentIds[0]
	endDate := stepState.EndDate
	stmt :=
		`
		SELECT 
			l.lesson_id,
			lm.user_id,
			l.course_id
		from
			lessons l
		JOIN lesson_members lm on l.lesson_id = lm.lesson_id
		WHERE lm.deleted_at is NOT null
		AND  lm.user_id = $1 
		AND	 l.class_id = ANY($2)
		AND l.start_time::date > $3::date
		`
	rows, err := s.BobDBTrace.DB.Query(
		ctx,
		stmt,
		studentID,
		stepState.JoinedClass,
		endDate,
	)
	if err != nil {
		return ctx, errors.Wrap(err, "query lesson_members failed")
	}
	defer rows.Close()
	lessonMembers := make([]*domain.LessonMember, 0)

	for rows.Next() {
		lessonMember := &domain.LessonMember{}
		err := rows.Scan(
			&lessonMember.LessonID,
			&lessonMember.StudentID,
			&lessonMember.CourseID,
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), errors.WithMessage(err, "rows.Scan class failed")
		}

		lessonMembers = append(lessonMembers, lessonMember)
	}

	if len(lessonMembers) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student not leave lesson member")
	}

	return StepStateToContext(ctx, stepState), nil
}
