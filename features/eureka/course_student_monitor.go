package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/eureka/configurations"
	entities "github.com/manabie-com/backend/internal/eureka/entities/monitors"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	monitor_repo "github.com/manabie-com/backend/internal/eureka/repositories/monitors"
	services "github.com/manabie-com/backend/internal/eureka/services/monitoring"
	"github.com/manabie-com/backend/internal/golibs/alert"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) schoolAdminAddSomeStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.SchoolAdminToken
	locationID := idutil.ULIDNow()
	e := &bob_entities.Location{}
	database.AllNullEntity(e)
	if err := multierr.Combine(
		e.LocationID.Set(locationID),
		e.Name.Set(fmt.Sprintf("location-%s", locationID)),
		e.IsArchived.Set(false),
		e.CreatedAt.Set(time.Now()),
		e.UpdatedAt.Set(time.Now()),
	); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if _, err := database.Insert(ctx, e, s.BobDB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	n := rand.Intn(6) + 3
	for i := 0; i < n; i++ {
		unique := idutil.ULIDNow()
		resp, err := upb.NewUserModifierServiceClient(s.UsermgmtConn).CreateStudent(s.signedCtx(ctx), &upb.CreateStudentRequest{
			SchoolId: stepState.SchoolIDInt,
			StudentProfile: &upb.CreateStudentRequest_StudentProfile{
				Email:            formatEmail(unique, i),
				Password:         "abcdef",
				Name:             formatName(unique, i),
				EnrollmentStatus: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
				Grade:            11,
				CountryCode:      cpb.Country_COUNTRY_NONE,
				LocationIds:      []string{locationID},
			},
		})

		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create student: %w", err)
		}
		studentID := resp.StudentProfile.GetStudent().GetUserProfile().GetUserId()
		if _, err := upb.NewUserModifierServiceClient(s.UsermgmtConn).UpsertStudentCoursePackage(s.signedCtx(ctx), &upb.UpsertStudentCoursePackageRequest{
			StudentId: studentID,
			StudentPackageProfiles: []*upb.UpsertStudentCoursePackageRequest_StudentPackageProfile{{
				Id: &upb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
					CourseId: stepState.CourseID,
				},
				StartTime: timestamppb.New(time.Now().Add(time.Hour * -20)),
				EndTime:   timestamppb.New(time.Now().Add(time.Hour * 24 * 10)),
			}},
		}); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to update student course package: %w", err)
		}
		stepState.StudentIDs = append(stepState.StudentIDs, studentID)
	}
	return StepStateToContext(ctx, stepState), nil
}

func formatEmail(id string, i int) string {
	return "syllabus" + id + "_" + strconv.Itoa(i) + "@gmail.com"
}
func formatName(today string, i int) string {
	return "syllabus" + "Name" + today + "_" + strconv.Itoa(i)
}

func (s *suite) someStudentsStudyPlansNotCreated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	someStudentMissing := rand.Intn(2) + 1
	stepState.StudentIDsWithMissingStudyPlan = stepState.StudentIDs[:someStudentMissing]
	cmd := `UPDATE student_study_plans SET deleted_at = now() WHERE student_id = ANY($1::_TEXT) AND deleted_at IS NULL`
	_, err := s.DB.Exec(ctx, cmd, database.TextArray(stepState.StudentIDsWithMissingStudyPlan))

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to simulator missing study plan: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourMonitorSaveMissingStudentCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, items, err := s.getStudyPlanMonitor(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("getStudyPlanMonitor: %w", err)
	}
	mapItem := make(map[string]bool)
	for _, item := range items {
		mapItem[item.StudentID.String] = true
	}
	for _, id := range stepState.StudentIDsWithMissingStudyPlan {
		if _, ok := mapItem[id]; !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("system not save student with missing study plan correctly")
		}
	}
	studentIDsCorrectly := stepState.StudentIDs[len(stepState.StudentIDsWithMissingStudyPlan):]
	for _, id := range studentIDsCorrectly {
		if _, ok := mapItem[id]; ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("system save wrong: the student already have study plans")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getStudyPlanMonitor(ctx context.Context) (context.Context, []*entities.StudyPlanMonitor, error) {
	stepState := StepStateFromContext(ctx)
	cmd := `SELECT %s FROM %s WHERE student_id = ANY($1::_TEXT) AND type = $2::TEXT`
	var e entities.StudyPlanMonitor
	selectFields := database.GetFieldNames(&e)
	var items entities.StudyPlanMonitors
	err := database.Select(ctx, s.DB, fmt.Sprintf(cmd, strings.Join(selectFields, ","), e.TableName()), database.TextArray(stepState.StudentIDs), database.Text(entities.StudyPlanMonitorType_STUDENT_STUDY_PLAN)).ScanAll(&items)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}
	return StepStateToContext(ctx, stepState), items, nil
}

func (s *suite) runMonitorUpsertCourseStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedCtx(ctx)
	// TODO: consider add to config
	// check the resource path work or not-- like the config in local, in current Im not adding a lot config
	// for db RLS query
	ctx = auth.InjectFakeJwtToken(ctx, stepState.SchoolID)
	studentStudyPlanRepo := &repositories.StudentStudyPlanRepo{}
	courseStudentRepo := &repositories.CourseStudentRepo{}
	studyPlanRepo := &repositories.StudyPlanRepo{}
	studyPlanMonitorRepo := &monitor_repo.StudyPlanMonitorRepo{}
	c := &configurations.Config{
		SyllabusSlackWebHook: "https://hooks.slack.com/services/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",
		SchoolInformation: configurations.SchoolInfoConfig{
			SchoolID:   stepState.SchoolID,
			SchoolName: "Local shool",
		},
		Common: configs.CommonConfig{
			Environment: "local",
		},
	}
	httpClient := http.Client{Timeout: time.Duration(10) * time.Second}
	alertClient := &alert.SlackImpl{
		WebHookURL: "https://hooks.slack.com/services/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",
		HTTPClient: httpClient,
	}
	studyPlanMonitorService := &services.StudyPlanMonitorService{
		Cfg:                  c,
		Logger:               *zap.NewNop(),
		Alert:                alertClient,
		DB:                   s.DB,
		StudentStudyPlanRepo: studentStudyPlanRepo,
		CourseStudentRepo:    courseStudentRepo,
		StudyPlanRepo:        studyPlanRepo,
		StudyPlanMonitorRepo: studyPlanMonitorRepo,
	}
	err := studyPlanMonitorService.UpsertStudentCourse(ctx, 15)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to monitor upsert course student: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}
