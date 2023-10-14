package yasuo

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	bobRepo "github.com/manabie-com/backend/internal/bob/repositories"
	enigma_entites "github.com/manabie-com/backend/internal/enigma/entities"
	eureka_entities "github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) classWithCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.aClass(ctx)
	if err != nil {
		return ctx, err
	}

	token := stepState.AuthToken
	ctx, err = s.aSignedIn(ctx, "school admin")
	if err != nil {
		return ctx, err
	}

	ctx, err = s.aValidCourse(ctx)
	if err != nil {
		return ctx, err
	}

	stepState.AuthToken = token
	e := &entities.CourseClass{
		CourseID:  database.Text(stepState.CurrentCourseID),
		ClassID:   database.Int4(stepState.CurrentClassID),
		Status:    database.Text("COURSE_CLASS_STATUS_ACTIVE"),
		CreatedAt: database.Timestamptz(time.Now()),
		UpdatedAt: database.Timestamptz(time.Now()),
		DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
	}

	cmd, err := database.Insert(ctx, e, s.DBTrace.Exec)
	if err != nil {
		return ctx, err
	}
	if cmd.RowsAffected() != 1 {
		return ctx, fmt.Errorf("no rows affected")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) toStudentSyncMsg(ctx context.Context, status, actionKind string, schoolID, total int) (context.Context, []*npb.EventUserRegistration_Student, error) {
	if total == 0 {
		return ctx, []*npb.EventUserRegistration_Student{}, nil
	}

	stepState := StepStateFromContext(ctx)

	students := []*npb.EventUserRegistration_Student{}

	switch status {
	case "new":
		s.aRandomNumber()
		ctx, err := s.aTeacherAccountWithSchoolID(ctx, int32(schoolID))
		if err != nil {
			return ctx, nil, fmt.Errorf("aTeacherAccountWithSchoolID:%w", err)
		}
		ctx, err = s.classWithCourse(ctx)
		if err != nil {
			return ctx, nil, fmt.Errorf("classWithCourse:%w", err)
		}

		studentID := newID()
		for i := 0; i < total; i++ {
			students = append(students, &npb.EventUserRegistration_Student{
				ActionKind:  npb.ActionKind(npb.ActionKind_value[actionKind]),
				StudentId:   studentID,
				StudentDivs: []int64{1, 2},
				LastName:    fmt.Sprintf("student-%s", studentID),
				GivenName:   fmt.Sprintf("student-%s", studentID),
				Packages: []*npb.EventUserRegistration_Student_Package{
					{
						ClassId:   int64(stepState.CurrentClassID),
						StartDate: timestamppb.Now(),
						EndDate:   timestamppb.New(time.Now().Add(3 * time.Hour)),
					},
				},
			})
		}
	case "existed":
		stepState.Request = nil

		ctx, err1 := s.jprepSyncStudentsWithActionAndStudentsWithAction(StepStateToContext(ctx, stepState), strconv.Itoa(total), npb.ActionKind_ACTION_KIND_UPSERTED.String(), "0", "")
		ctx, err2 := s.theseStudentsMustBeStoreInOurSystem(StepStateToContext(ctx, stepState))
		err := multierr.Combine(err1, err2)

		if err != nil {
			return ctx, nil, fmt.Errorf("multierr.Combine: existed %v", err)
		}

		for _, s := range stepState.Request.([]*npb.EventUserRegistration_Student) {
			s.ActionKind = npb.ActionKind(npb.ActionKind_value[actionKind])
			students = append(students, s)
		}
	}

	return StepStateToContext(ctx, stepState), students, nil
}

func (s *suite) jprepSyncStudentsWithActionAndStudentsWithAction(ctx context.Context, numberOfNewStudent, newStudentAction, numberOfExistedStudent, existedStudentAction string) (context.Context, error) {
	total, err := strconv.Atoi(numberOfNewStudent)
	if err != nil {
		return ctx, err
	}

	stepState := StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()
	stepState.CurrentSchoolID = constants.JPREPSchool

	ctx, err = s.aSignedIn(ctx, "school admin")
	if err != nil {
		return ctx, err
	}

	ctx, newStudents, err := s.toStudentSyncMsg(ctx, "new", newStudentAction, constants.JPREPSchool, total)
	if err != nil {
		return ctx, err
	}

	total, err = strconv.Atoi(numberOfExistedStudent)
	if err != nil {
		return ctx, err
	}

	ctx, existedStudents, err := s.toStudentSyncMsg(ctx, "existed", existedStudentAction, constants.JPREPSchool, total)
	if err != nil {
		return ctx, err
	}

	students := append(newStudents, existedStudents...)
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
		return StepStateToContext(ctx, stepState), fmt.Errorf("publish: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

//nolint:gocyclo
func (s *suite) theseStudentsMustBeStoreInOurSystem(ctx context.Context) (context.Context, error) {
	time.Sleep(10 * time.Second)
	stepState := StepStateFromContext(ctx)

	var err error
	for _, student := range stepState.Request.([]*npb.EventUserRegistration_Student) {
		var resp *pb.GetStudentProfileResponse
		if student.ActionKind != npb.ActionKind_ACTION_KIND_DELETED {
			stepState.AuthToken, err = s.generateExchangeToken(student.StudentId, pb.USER_GROUP_STUDENT.String())
			if err != nil {
				return ctx, err
			}
			ctx = StepStateToContext(ctx, stepState)
			resp, err = pb.NewStudentClient(s.BobConn).GetStudentProfile(contextWithToken(s, ctx), &pb.GetStudentProfileRequest{})
			if err != nil {
				return ctx, err
			}
		} else {
			continue
		}

		if len(resp.Datas) != 1 {
			return ctx, fmt.Errorf("not found student")
		}

		studentResp := resp.Datas[0]

		if studentResp.Profile.Id != student.StudentId {
			return ctx, fmt.Errorf("studentID does not match")
		}

		if studentResp.Profile.Name != student.GivenName+" "+student.LastName {
			return ctx, fmt.Errorf("name does not match")
		}

		if studentResp.Profile.Country != pb.COUNTRY_JP {
			return ctx, fmt.Errorf("country does not match")
		}

		if len(student.StudentDivs) != len(studentResp.Profile.Divs) {
			return ctx, fmt.Errorf("divs does not match")
		}

		if studentResp.Profile.School.Id != constants.JPREPSchool {
			return ctx, fmt.Errorf("school does not match")
		}

		for i := range student.StudentDivs {
			if student.StudentDivs[i] != studentResp.Profile.Divs[i] {
				return ctx, fmt.Errorf("student divs does not match")
			}
		}
		classRepo := &bobRepo.ClassRepo{}
		classes, err := classRepo.FindJoined(ctx, s.DBTrace, database.Text(student.StudentId))
		if err != nil {
			return ctx, err
		}

		if len(classes) != len(student.Packages) {
			fmt.Println("len classes: ", len(classes), "len packages: ", len(student.Packages))
			return ctx, fmt.Errorf("total classes does not match")
		}

		for _, c := range classes {
			found := false
			for _, p := range student.Packages {
				if p.ClassId == int64(c.ID.Int) {
					found = true
				}
			}

			if !found {
				return ctx, fmt.Errorf("classID not joined %d", c.ID.Int)
			}
		}

		classIDs := make([]int32, 0)
		for _, class := range classes {
			classIDs = append(classIDs, class.ID.Int)
		}
		courseClassRepo := &bobRepo.CourseClassRepo{}
		mapClassIDToCourseID, err := courseClassRepo.Find(ctx, s.DBTrace, database.Int4Array(classIDs))
		if err != nil {
			return ctx, err
		}

		expectedCourseStudents := make([]*eureka_entities.CourseStudent, 0)

		courseIDs := make([]string, 0)
		for _, p := range student.Packages {
			for _, courseID := range mapClassIDToCourseID[database.Int4(int32(p.ClassId))].Elements {
				courseIDs = append(courseIDs, courseID.String)
				expectedCourseStudents = append(expectedCourseStudents, &eureka_entities.CourseStudent{
					CourseID:  courseID,
					StudentID: database.Text(student.StudentId),
					StartAt:   database.TimestamptzFromPb(p.StartDate),
					EndAt:     database.TimestamptzFromPb(p.EndDate),
				})
			}
		}

		checkCourseStudentFunc := func() error {
			currentCourseStudent, err := s.getCourseStudentFromDB(ctx, student.StudentId, courseIDs)
			if err != nil {
				return err
			}

			if len(expectedCourseStudents) != len(currentCourseStudent) {
				return fmt.Errorf("epxect number of course student is %d but got %d", len(expectedCourseStudents), len(currentCourseStudent))
			}

			sort.Slice(expectedCourseStudents, func(i, j int) bool {
				return expectedCourseStudents[i].CourseID.String < expectedCourseStudents[j].CourseID.String
			})
			sort.Slice(currentCourseStudent, func(i, j int) bool {
				return currentCourseStudent[i].CourseID.String < currentCourseStudent[j].CourseID.String
			})
			for i := 0; i < len(expectedCourseStudents); i++ {
				e := expectedCourseStudents[i]
				c := currentCourseStudent[i]
				if e.CourseID.String != c.CourseID.String {
					return fmt.Errorf("expect course id %s but got %s", e.CourseID.String, c.CourseID.String)
				}
				if e.StudentID.String != c.StudentID.String {
					return fmt.Errorf("expect student id %s but got %s", e.StudentID.String, c.StudentID.String)
				}

				if !e.StartAt.Time.Round(time.Second).Equal(c.StartAt.Time.Round(time.Second)) {
					return fmt.Errorf("expect course student start at %v but got %v", e.StartAt.Time, c.StartAt.Time)
				}
				if !e.EndAt.Time.Round(time.Second).Equal(c.EndAt.Time.Round(time.Second)) {
					return fmt.Errorf("expect course student start at %v but got %v", e.StartAt.Time, c.StartAt.Time)
				}
			}
			return nil
		}

		err = try.Do(func(attempt int) (bool, error) {
			err := checkCourseStudentFunc()
			if err != nil {
				time.Sleep(500 * time.Millisecond)
				return attempt < 10, nil
			}
			return true, err
		})

		if err != nil {
			return ctx, err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getCourseStudentFromDB(ctx context.Context, studentID string, courseIDs []string) ([]*eureka_entities.CourseStudent, error) {
	e := &eureka_entities.CourseStudent{}
	query := fmt.Sprintf(`SELECT %s FROM course_students WHERE student_id = $1 AND course_id = ANY($2)`, strings.Join(database.GetFieldNames(e), ","))

	result := make([]*eureka_entities.CourseStudent, 0)
	rows, err := s.EurekaDB.Query(ctx, query, database.Text(studentID), database.TextArray(courseIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		e := &eureka_entities.CourseStudent{}
		err := rows.Scan(database.GetScanFields(e, database.GetFieldNames(e))...)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
