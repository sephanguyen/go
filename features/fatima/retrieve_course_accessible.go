package fatima

import (
	"context"
	"fmt"
	"strings"
	"time"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	bob_repo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/fatima/repositories"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"

	"github.com/jackc/pgx/v4"
	"github.com/lestrrat-go/jwx/jwt"
	"go.uber.org/multierr"
)

func (s *suite) returnsAllCourseAccessibleResponseOfThisUser(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	token, _ := jwt.ParseString(s.AuthToken)
	resp := stepState.Response.(*pb.RetrieveAccessibilityResponse)
	return StepStateToContext(ctx, stepState), s.returnAllCourseAccessibleWithUserID(ctx, resp.Courses, token.Subject())
}

func (s *suite) returnAllCourseAccessibleWithUserID(ctx context.Context, courses map[string]*pb.RetrieveAccessibilityResponse_CourseAccessibility, userID string) error {
	r := &repositories.StudentPackageRepo{}
	studentPackages, err := r.CurrentPackage(auth.InjectFakeJwtToken(ctx, fmt.Sprint(constants.ManabieSchool)), s.DB, database.Text(userID))
	if err != nil {
		return err
	}

	if len(studentPackages) == 0 && len(courses) > 0 {
		return fmt.Errorf("no package available, student must not access any courses")
	}

	for _, sp := range studentPackages {
		props, err := sp.GetProperties()
		if err != nil {
			return fmt.Errorf("err getProps: %w", err)
		}

		for _, courseID := range props.CanWatchVideo {
			c, ok := courses[courseID]
			if !ok {
				return fmt.Errorf("missing courseID '%s'", courseID)
			}

			if !c.CanWatchVideo {
				return fmt.Errorf("courseID '%s' must can watch video", courseID)
			}
		}

		for _, courseID := range props.CanViewStudyGuide {
			c, ok := courses[courseID]
			if !ok {
				return fmt.Errorf("missing courseID '%s'", courseID)
			}

			if !c.CanViewStudyGuide {
				return fmt.Errorf("courseID '%s' must can view study guides", courseID)
			}
		}

		for _, courseID := range props.CanDoQuiz {
			c, ok := courses[courseID]
			if !ok {
				return fmt.Errorf("missing courseID '%s'", courseID)
			}

			if !c.CanDoQuiz {
				return fmt.Errorf("courseID '%s' must can do quiz", courseID)
			}
		}
	}

	return nil
}

func (s *suite) SomePackageDataInDB() error {
	return s.somePackageDataInDb()
}
func (s *suite) somePackageDataInDb() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	packages := map[string]*entities.Package{}
	s.CourseIDs = []string{newID()}

	now := time.Now()

	courseRepo := bob_repo.CourseRepo{}
	coursesEntity := []*bob_entities.Course{}
	for _, courseId := range s.CourseIDs {
		courseEntityAllNullEmpty := &bob_entities.Course{}
		database.AllNullEntity(courseEntityAllNullEmpty)
		courseEntityAllNullEmpty.ID = database.Text(courseId)
		courseEntityAllNullEmpty.Name = database.Text(idutil.ULIDNow())
		courseEntityAllNullEmpty.SchoolID = database.Int4(constants.ManabieSchool)
		courseEntityAllNullEmpty.Grade = database.Int2(10)
		coursesEntity = append(coursesEntity, courseEntityAllNullEmpty)
	}
	err := courseRepo.Upsert(ctx, s.BobDB, coursesEntity)
	if err != nil {
		return fmt.Errorf("err upsert courses: %w", err)
	}

	studentID, err := s.insertStudentIntoBob(ctx)
	if err != nil {
		return fmt.Errorf("err create student: %w", err)
	}
	s.StudentID = studentID

	p1 := &entities.Package{}
	database.AllNullEntity(p1)
	err = multierr.Combine(
		p1.ID.Set("free_package"),
		p1.Country.Set(cpb.Country_COUNTRY_VN.String()),
		p1.Name.Set("Free"),
		p1.Descriptions.Set([]string{"Free quiz and study guides"}),
		p1.Price.Set(0),
		p1.DiscountedPrice.Set(0),
		p1.PrioritizeLevel.Set(1),
		p1.StartAt.Set(now),
		p1.EndAt.Set(now.Add(7*24*time.Hour)),
		p1.Duration.Set(nil),
		p1.Properties.Set(&entities.PackageProperties{
			CanWatchVideo:     s.CourseIDs,
			CanViewStudyGuide: s.CourseIDs,
			CanDoQuiz:         s.CourseIDs,
			LimitOnlineLesson: 0,
		}),
		p1.IsRecommended.Set(false),
		p1.IsActive.Set(true),
		p1.CreatedAt.Set(now),
		p1.UpdatedAt.Set(now),
	)
	if err != nil {
		return fmt.Errorf("err set: %w", err)
	}

	packages[p1.ID.String] = p1

	p2 := &entities.Package{}
	database.AllNullEntity(p2)
	err = multierr.Combine(
		p2.ID.Set("basic_trial_package"),
		p2.Country.Set(cpb.Country_COUNTRY_VN.String()),
		p2.Name.Set("Basic Trial"),
		p2.Descriptions.Set([]string{"Free quiz and study guides", "7 days access all basic features"}),
		p2.Price.Set(100000),
		p2.DiscountedPrice.Set(0),
		p2.PrioritizeLevel.Set(2),
		p1.StartAt.Set(nil),
		p1.EndAt.Set(nil),
		p1.Duration.Set(7),
		p2.Properties.Set(&entities.PackageProperties{
			CanWatchVideo:     []string{"course_1", "course_2", "course_3", "course_4", "course_5"},
			CanViewStudyGuide: []string{"course_1", "course_2", "course_3", "course_4"},
			CanDoQuiz:         []string{"course_1", "course_2", "course_3", "course_4"},
			LimitOnlineLesson: 0,
		}),
		p2.IsRecommended.Set(false),
		p2.IsActive.Set(true),
		p2.CreatedAt.Set(now),
		p2.UpdatedAt.Set(now),
	)
	if err != nil {
		return fmt.Errorf("err set: %w", err)
	}

	packages[p2.ID.String] = p2
	r := &repositories.PackageRepo{}

	for _, p := range packages {
		err = r.Upsert(ctx, s.DB, p)
		if err != nil {
			return fmt.Errorf("err insertPackage: %w", err)
		}
	}

	s.Packages = packages

	return nil
}

func (s *suite) thisUserHasPackageIs(pkgs, rawStatuses string) error {
	time.Sleep(2 * time.Second)
	packages := strings.Split(pkgs, ",")
	statuses := strings.Split(rawStatuses, ",")

	if len(packages) != len(statuses) {
		return fmt.Errorf("args packages and statuses does not match, each elem should slit by comma")
	}

	t, _ := jwt.ParseString(s.AuthToken)
	return s.userPackage(t.Subject(), pkgs, rawStatuses)
}

func (s *suite) userPackage(userID, pkgs, rawStatuses string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	packages := strings.Split(pkgs, ",")
	statuses := strings.Split(rawStatuses, ",")

	if len(packages) != len(statuses) {
		return fmt.Errorf("args packages and statuses does not match, each elem should slit by comma")
	}

	r := &repositories.StudentPackageRepo{}
	now := time.Now().Add(-time.Minute)

	err := database.ExecInTxWithRetry(auth.InjectFakeJwtToken(ctx, fmt.Sprint(constants.ManabieSchool)), s.DB, func(ctx context.Context, tx pgx.Tx) error {
		for i, pkgID := range packages {
			p, ok := s.Packages[pkgID]
			if !ok {
				return fmt.Errorf("package %s does not existed", pkgID)
			}

			status := statuses[i]

			props, err := p.GetProperties()
			if err != nil {
				return fmt.Errorf("err GetProperties: %w", err)
			}

			startAt := p.StartAt.Time
			endAt := p.EndAt.Time
			if p.Duration.Int > 0 {
				startAt = now
				endAt = now.Add(time.Duration(p.Duration.Int) * time.Hour * 24)
			}

			if status == "expired" {
				// 30days ago
				startAt = startAt.Add(30 * 24 * time.Hour)
				endAt = endAt.Add(30 * 24 * time.Hour)
			}
			sp := &entities.StudentPackage{}
			database.AllNullEntity(sp)
			err = multierr.Combine(
				sp.ID.Set(idutil.ULIDNow()),
				sp.StudentID.Set(userID),
				sp.PackageID.Set(p.ID.String),
				sp.StartAt.Set(startAt),
				sp.EndAt.Set(endAt),
				sp.Properties.Set(&entities.StudentPackageProps{
					CanWatchVideo:     props.CanWatchVideo,
					CanViewStudyGuide: props.CanViewStudyGuide,
					CanDoQuiz:         props.CanDoQuiz,
					LimitOnlineLesson: props.LimitOnlineLesson,
				}),
				sp.IsActive.Set(true),
			)
			if err != nil {
				return fmt.Errorf("err set sp: %w", err)
			}

			err = r.Insert(ctx, tx, sp)
			if err != nil {
				return fmt.Errorf("err insert SP: %w", err)
			}
		}
		return nil
	})

	return err
}
func (s *suite) userRetrieveAccessibleCourse() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	s.Response, s.ResponseErr = pb.NewAccessibilityReadServiceClient(s.Conn).RetrieveAccessibility(contextWithToken(s, ctx), &pb.RetrieveAccessibilityRequest{})

	return nil
}
