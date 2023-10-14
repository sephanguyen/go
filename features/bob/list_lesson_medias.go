package bob

import (
	"context"
	"fmt"
	"sort"

	"github.com/lestrrat-go/jwx/jwt"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpbV1 "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"go.uber.org/multierr"
)

func (s *suite) aListOfMediaWhichAttachedToALesson(ctx context.Context) (context.Context, error) {
	ctx, err1 := s.signedAsAccount(ctx, "student")
	ctx, err2 := s.userUpsertValidMediaList(ctx)
	ctx, err3 := s.bobMustRecordAllMediaList(ctx)

	err := multierr.Combine(err1, err2, err3)

	if err != nil {
		return ctx, err

	}

	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*bpb.UpsertMediaResponse)
	stepState.MediaIDs = resp.MediaIds

	// attach media to lesson

	ctx, err1 = s.aTeacherAndAClassWithSomeStudents(ctx)
	ctx, err2 = s.aListOfCoursesAreExistedInDBOf(ctx, "above teacher")
	ctx, err3 = s.aStudentWithValidLesson(ctx)
	ctx, err4 := s.upsertMediaToLessonGroup(ctx)
	err = multierr.Combine(err1, err2, err3, err4)

	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aTeacherAndAClassWithSomeStudents(ctx context.Context) (context.Context, error) {
	ctx, err1 := s.aRandomNumber(ctx)
	ctx, err2 := s.aSignedInAdmin(ctx)
	ctx, err3 := s.aSchoolNameCountryCityDistrict(ctx, "S1", pb.COUNTRY_VN.String(), "Hồ Chí Minh", "2")
	ctx, err4 := s.aSchoolNameCountryCityDistrict(ctx, "S2", pb.COUNTRY_VN.String(), "Hồ Chí Minh", "3")
	ctx, err5 := s.adminInsertsSchools(ctx)
	if err := multierr.Combine(err1, err2, err3, err4, err5); err != nil {
		return ctx, err
	}

	ctx, err1 = s.aSignedInTeacher(ctx)
	ctx, err2 = s.aCreateClassRequest(ctx)
	ctx, err3 = s.aSchoolIdInCreateClassRequest(ctx, "valid")
	ctx, err4 = s.aValidNameInCreateClassRequest(ctx)
	ctx, err5 = s.thisSchoolHasConfigIsIsIs(ctx,
		"plan_id", "School",
		"plan_expired_at", "2055-06-30 23:59:59",
		"plan_duration", 0)
	ctx, err6 := s.userCreateAClass(ctx)
	ctx, err7 := s.bobMustCreateClassFromCreateClassRequest(ctx)
	if err := multierr.Combine(err1, err2, err3, err4, err5, err6, err7); err != nil {
		return ctx, err
	}

	ctx, err1 = s.studentJoinTeacherCurrentClass(ctx)
	ctx, err2 = s.studentJoinTeacherCurrentClass(ctx)
	ctx, err3 = s.studentJoinTeacherCurrentClass(ctx)
	ctx, err4 = s.studentJoinTeacherCurrentClass(ctx)
	ctx, err5 = s.studentJoinTeacherCurrentClass(ctx)
	if err := multierr.Combine(err1, err2, err3, err4, err5); err != nil {
		return ctx, err
	}

	return ctx, nil
}
func (s *suite) studentJoinTeacherCurrentClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	teacherToken := stepState.AuthToken

	req := &pb.JoinClassRequest{}
	req.ClassCode = stepState.CurrentClassCode
	s.aSignedInStudent(ctx)
	t, _ := jwt.ParseString(stepState.AuthToken)
	stepState.CurrentStudentID = t.Subject()

	stepState.Response, stepState.ResponseErr = pb.NewClassClient(s.Conn).JoinClass(contextWithToken(s, ctx), req)

	stepState.AuthToken = teacherToken
	stepState.StudentInCurrentClass = append(stepState.StudentInCurrentClass, stepState.CurrentStudentID)

	return StepStateToContext(ctx, stepState), stepState.ResponseErr
}

func (s *suite) ATeacherAndAClassWithSomeStudents(ctx context.Context) (context.Context, error) {
	return s.aTeacherAndAClassWithSomeStudents(ctx)
}
func (s *suite) upsertMediaToLessonGroup(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	_, err := s.DB.Exec(ctx, `INSERT INTO lesson_groups (lesson_group_id, course_id, media_ids, updated_at, created_at) 
	VALUES($1, $2, $3, now(), now())
	ON CONFLICT ON CONSTRAINT pk__lesson_groups 
	DO UPDATE SET media_ids = excluded.media_ids WHERE lesson_groups.media_ids IS NULL;`, stepState.CurrentLessonGroupID, stepState.CurrentCourseID, stepState.MediaIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) theListOfMediaMatchWithResponseMedias(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.MediaItems) != len(stepState.MediaIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect %v media item but got %v", len(stepState.MediaIDs), len(stepState.MediaItems))
	}

	mediaRepo := repositories.MediaRepo{}
	expectMedia, err := mediaRepo.RetrieveByIDs(ctx, s.DB, database.TextArray(stepState.MediaIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	sort.SliceStable(expectMedia, func(i, j int) bool {
		return expectMedia[i].MediaID.String > expectMedia[j].MediaID.String
	})
	for i, item := range stepState.MediaItems {
		if item.MediaId != expectMedia[i].MediaID.String {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect %v media id but got %v", stepState.MediaIDs[i], item.MediaId)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userGetLessonMedias(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp, err := bpbV1.NewCourseReaderServiceClient(s.Conn).ListLessonMedias(s.signedCtx(ctx), &bpbV1.ListLessonMediasRequest{
		LessonId: stepState.CurrentLessonID,
		Paging: &cpb.Paging{
			Limit: 1,
		},
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	stepState.MediaItems = append(stepState.MediaItems, resp.Items...)

	nextPage := resp.NextPage
	for len(resp.Items) != 0 {
		resp, err = bpbV1.NewCourseReaderServiceClient(s.Conn).ListLessonMedias(s.signedCtx(ctx), &bpbV1.ListLessonMediasRequest{
			LessonId: stepState.CurrentLessonID,
			Paging:   nextPage,
		})
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}
		nextPage = resp.NextPage
		stepState.MediaItems = append(stepState.MediaItems, resp.Items...)
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userGetLessonMediasAndReturnsStatusCode(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = bpbV1.NewCourseReaderServiceClient(s.Conn).ListLessonMedias(s.signedCtx(ctx), &bpbV1.ListLessonMediasRequest{
		LessonId: stepState.CurrentLessonID,
		Paging: &cpb.Paging{
			Limit: 1,
		},
	})
	s.returnsStatusCode(ctx, "PermissionDenied")
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userGetMediaOfNonExistedLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = bpbV1.NewCourseReaderServiceClient(s.Conn).ListLessonMedias(s.signedCtx(ctx), &bpbV1.ListLessonMediasRequest{
		LessonId: s.newID(),
		Paging: &cpb.Paging{
			Limit: 1,
		},
	})
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) emptyMediaResult(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*bpbV1.ListLessonMediasResponse)
	if len(resp.Items) != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected empty result but got %v", resp.Items)
	}
	return StepStateToContext(ctx, stepState), nil
}
