package bob

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	master_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure/repo"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"go.uber.org/multierr"
)

func (s *suite) someMembersInSomeClasses(ctx context.Context) (context.Context, error) {
	ctx = common.ValidContext(ctx, constants.ManabieSchool, s.RootAccount[constants.ManabieSchool].UserID, s.RootAccount[constants.ManabieSchool].Token)
	ctx, err1 := s.generateSchool(ctx)
	ctx, err2 := s.generateTeacherWithCurrentSchoolID(ctx, 2)
	ctx, err3 := s.generateStudentsWithGivenNumber(ctx, rand.Intn(10)+10)
	ctx, err4 := s.aListOfLocationsInDB(ctx)
	ctx, err5 := s.CreateLiveCourse(ctx)
	ctx, err6 := s.generateTwoClasses(ctx)
	ctx, err7 := s.generateClassMembers(ctx)
	ctx, err8 := s.generateClassMembersV2(ctx)
	err := multierr.Combine(err1, err2, err3, err4, err5, err6, err7, err8)
	if err != nil {
		return ctx, fmt.Errorf("someMembersInSomeClasses: %w", err)
	}
	return ctx, nil
}
func (s *suite) generateSchool(ctx context.Context) (context.Context, error) {
	ctx, err1 := s.aSchoolNameCountryCityDistrict(ctx, "S2", pb.COUNTRY_VN.String(), "Hồ Chí Minh", "3")
	ctx, err2 := s.adminInsertsSchools(ctx)
	if err := multierr.Combine(err1, err2); err != nil {
		return ctx, fmt.Errorf("generateSchool: %w", err)
	}
	return ctx, nil
}
func (s *suite) generateTeacherWithCurrentSchoolID(ctx context.Context, n int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for i := 0; i < n; i++ {
		id := s.newID()
		stepState.TeacherIDs = append(stepState.TeacherIDs, id)
		if ctx, err := s.aValidTeacherProfileWithId(ctx, id, stepState.CurrentSchoolID); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable create teacher: %w", err)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) generateStudentsWithGivenNumber(ctx context.Context, n int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for i := 0; i < n; i++ {
		id := s.newID()
		stepState.StudentIds = append(stepState.StudentIds, id)
		ctx, err := s.aValidStudentInDB(ctx, id)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable create student: %w", err)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) generateTwoClasses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// Generate 2 class
	var classID int32
	rand.Seed(time.Now().UnixNano())
	classRepo := &repositories.ClassRepo{}
	masterClassRepo := &master_repo.ClassRepo{}

	for i := 0; i < 2; i++ {
		classID = rand.Int31()
		class, err := s.generateclass(classID, stepState.CurrentSchoolID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		err = classRepo.Create(ctx, s.DB, class)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		// insert new class table
		domainClass := &domain.Class{
			ClassID:    strconv.Itoa(int(classID)),
			Name:       "class name",
			CourseID:   stepState.courseIds[0],
			LocationID: stepState.LocationIDs[0],
		}
		err = masterClassRepo.Insert(ctx, s.DB, []*domain.Class{domainClass})
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.ClassIDs = append(stepState.ClassIDs, classID)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) generateClassMembers(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.ClassIDs) != 2 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected number of class, want: 2, actual: %d", len(stepState.ClassIDs))
	}
	classMemberRepo := &repositories.ClassMemberRepo{}

	classID := stepState.ClassIDs[0]
	var group = cpb.UserGroup_name[int32(cpb.UserGroup_USER_GROUP_STUDENT)]
	for idx, userID := range stepState.StudentIds {
		if idx > len(stepState.StudentIds)/2 {
			classID = stepState.ClassIDs[1]
		}
		classMember, err := s.generateAClassMember(userID, group, classID, false)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}

		err = classMemberRepo.Create(ctx, s.DB, classMember)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("classMemberRepo.Create: %w, class_id: %d", err, classMember.ClassID)
		}
	}
	group = cpb.UserGroup_name[int32(cpb.UserGroup_USER_GROUP_TEACHER)]
	for _, userID := range stepState.TeacherIDs {
		classMember, err := s.generateAClassMember(userID, group, classID, true)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}

		err = classMemberRepo.Create(ctx, s.DB, classMember)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("classMemberRepo.Create: %w, class_id: %d", err, classMember.ClassID)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateClassMembersV2(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.ClassIDs) != 2 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected number of class, want: 2, actual: %d", len(stepState.ClassIDs))
	}
	classMemberRepo := &master_repo.ClassMemberRepo{}

	classID := strconv.Itoa(int(stepState.ClassIDs[0]))
	for idx, userID := range stepState.StudentIds {
		if idx > len(stepState.StudentIds)/2 {
			classID = strconv.Itoa(int(stepState.ClassIDs[1]))
		}
		classMember := s.generateAClassMemberV2(userID, classID)

		err := classMemberRepo.UpsertClassMember(ctx, s.DB, classMember)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("classMemberRepo.UpsertClassMember: %w, class_id: %s", err, classMember.ClassID)
		}
	}

	for _, userID := range stepState.TeacherIDs {
		classMember := s.generateAClassMemberV2(userID, classID)
		err := classMemberRepo.UpsertClassMember(ctx, s.DB, classMember)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("classMemberRepo.Create: %w, class_id: %s", err, classMember.ClassID)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateAClassMember(userID, group string, classID int32, isOwner bool) (*entities.ClassMember, error) {
	classMember := new(entities.ClassMember)
	database.AllNullEntity(classMember)
	err := multierr.Combine(classMember.ID.Set(s.newID()), classMember.ClassID.Set(classID), classMember.UserID.Set(userID), classMember.UserGroup.Set(group), classMember.IsOwner.Set(isOwner), classMember.Status.Set(entities.ClassMemberStatusActive))
	if err != nil {
		return nil, fmt.Errorf("generateAClassMember.SetEntity: %w", err)
	}
	return classMember, nil
}
func (s *suite) generateAClassMemberV2(userID, classID string) *domain.ClassMember {
	return &domain.ClassMember{
		ClassMemberID: idutil.ULIDNow(),
		ClassID:       classID,
		UserID:        userID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		DeletedAt:     nil,
	}
}
func (s *suite) generateclass(id, currentSchoolID int32) (*entities.Class, error) {
	e := &entities.Class{}
	database.AllNullEntity(e)
	err := multierr.Combine(e.ID.Set(id),
		e.SchoolID.Set(currentSchoolID),
		e.Name.Set("name"+strconv.Itoa(int(id))), e.Grades.Set([]int32{11, 12}),
		e.PlanID.Set("School"), e.Country.Set("COUNTRY_VN"),
		e.Status.Set(entities.ClassStatusActive),
		e.Code.Set(s.newID()),
		e.Avatar.Set("AVT"))
	if err != nil {
		return nil, fmt.Errorf("generateclass: %w", err)
	}
	return e, nil
}
func (s *suite) theTeacherRetrievesClassMembers(ctx context.Context, args string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	limit := uint32(rand.Int31n(2) + 2)
	req := &bpb.RetrieveClassMembersRequest{
		Paging:    &cpb.Paging{Limit: limit},
		ClassIds:  convertInt32ArrayToStringArray(stepState.ClassIDs),
		UserGroup: cpb.UserGroup_USER_GROUP_NONE,
	}
	stepState.Request = req
	switch args {
	case "student":
		req.UserGroup = cpb.UserGroup_USER_GROUP_STUDENT
	case "teacher":
		req.UserGroup = cpb.UserGroup_USER_GROUP_TEACHER
	default:
		req.UserGroup = cpb.UserGroup_USER_GROUP_NONE
	}
	var res *bpb.RetrieveClassMembersResponse
	var err error
	for {
		res, err = bpb.NewClassReaderServiceClient(s.Conn).RetrieveClassMembers(contextWithToken(s, ctx), req)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("ClassReaderService.RetrieveClassMembers: %w", err)
		}
		if len(res.Members) == 0 {
			break
		}
		stepState.ClassMembers = append(stepState.ClassMembers, res.GetMembers()...)
		req.Paging = res.GetPaging()
	}

	return StepStateToContext(ctx, stepState), nil
}
func convertInt32ArrayToStringArray(ss []int32) []string {
	result := make([]string, 0, len(ss))
	for _, element := range ss {
		val := strconv.Itoa(int(element))
		result = append(result, (val))
	}
	return result
}
func (s *suite) ourSystemReturnsClassMembersCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	classMemberIDs := make([]string, 0, len(stepState.ClassMembers))
	for _, cm := range stepState.ClassMembers {
		classMemberIDs = append(classMemberIDs, cm.UserId)
	}
	users := entities.Users{}
	ent := &entities.User{}
	fields, _ := ent.FieldMap()

	cmd := `SELECT %s FROM users WHERE user_id = ANY($1::_TEXT) ORDER BY name`
	// retrieve users order by name
	if err := database.Select(ctx, s.DB, fmt.Sprintf(cmd, strings.Join(fields, ", ")), database.TextArray(classMemberIDs)).ScanAll(&users); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := stepState.Request.(*bpb.RetrieveClassMembersRequest)
	for i := 0; i < len(stepState.ClassMembers); i++ {
		if stepState.ClassMembers[i].UserId != users[i].ID.String {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong order (by name) of class members")
		}
	}
	switch req.UserGroup {
	case cpb.UserGroup_USER_GROUP_STUDENT:
		if len(stepState.ClassMembers) != len(stepState.StudentIds) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected length of class members: want: %d, actual: %d", len(stepState.StudentIds), len(stepState.ClassMembers))
		}
	case cpb.UserGroup_USER_GROUP_TEACHER:
		if len(stepState.ClassMembers) != len(stepState.TeacherIDs) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected length of class members: want: %d, actual: %d", len(stepState.TeacherIDs), len(stepState.ClassMembers))
		}
	default:
		if len(stepState.ClassMembers) != len(stepState.TeacherIDs)+len(stepState.StudentIds) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected length of class members: want: %d, actual: %d", len(stepState.TeacherIDs)+len(stepState.StudentIds), len(stepState.ClassMembers))
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
