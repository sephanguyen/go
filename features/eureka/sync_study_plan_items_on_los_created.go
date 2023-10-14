package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"go.uber.org/multierr"
)

func (s *suite) studyPlanItemsHaveCreatedCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	spie := &entities.StudyPlanItem{}
	fields, _ := spie.FieldMap()

	stmt := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE content_structure ->> 'lo_id' = ANY($1)
	`, strings.Join(fields, ","), spie.TableName())
	var studyPlanItems []*entities.StudyPlanItem
	if err := try.Do(func(attempt int) (retry bool, err error) {
		var items []*entities.StudyPlanItem
		rows, err := s.DB.Query(ctx, stmt, stepState.LoIDs)
		if err != nil {
			return false, err
		}
		defer rows.Close()
		for rows.Next() {
			var e entities.StudyPlanItem
			if err := rows.Scan(database.GetScanFields(&e, fields)...); err != nil {
				return false, err
			}
			items = append(items, &e)
		}
		if len(items) != len(stepState.LOs)*len(stepState.StudyPlanIDs) {
			time.Sleep(2 * time.Second)
			return attempt < 5, fmt.Errorf("study plan items created is wrong, expect %v but got %v", len(stepState.LOs)*len(stepState.StudyPlanIDs), len(items))
		}
		studyPlanItems = items
		return false, nil
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	getKey := func(content *entities.ContentStructure) string {
		return strings.Join([]string{
			content.CourseID,
			content.BookID,
			content.ChapterID,
			content.TopicID,
			content.LoID,
		}, "|")
	}

	m := make(map[string]bool)

	for _, item := range studyPlanItems {
		var content *entities.ContentStructure
		item.ContentStructure.AssignTo(&content)
		m[getKey(content)] = true
	}

	for _, lo := range stepState.LOs {
		if ok := m[getKey(lo)]; !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("missing study plan item of content: %s", getKey(lo))
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCreatesSomeLosInBook(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = contextWithToken(s, ctx)

	if ctx, err := s.schoolAdminCreateAtopicAndAChapter(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, los := s.prepareLOV1(ctx, stepState.TopicID, 4, cpb.LearningObjectiveType_LEARNING_OBJECTIVE_TYPE_FLASH_CARD)
	if ctx, err := s.createLOV1(contextWithValidVersion(ctx), los); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create los: %w", err)
	}
	loIDs := make([]string, 0, len(los))
	loContents := make([]*entities.ContentStructure, 0, len(los))
	courseIDs := stepState.BookCourseMap[stepState.BookID]
	for _, lo := range los {
		loIDs = append(loIDs, lo.Info.Id)
	}
	for _, courseID := range courseIDs {
		for _, lo := range los {
			loContents = append(loContents, &entities.ContentStructure{
				CourseID:  courseID,
				BookID:    stepState.BookID,
				ChapterID: stepState.ChapterID,
				TopicID:   stepState.TopicID,
				LoID:      lo.Info.Id,
			})
		}
	}
	stepState.LoIDs = loIDs
	stepState.LOs = loContents
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCreatesSomeLosInBooks(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = contextWithToken(s, ctx)
	n := 4
	loIDs := make([]string, 0, len(stepState.BookIDs)*n)
	loContents := make([]*entities.ContentStructure, 0, len(stepState.BookIDs)*n)
	for _, bookID := range stepState.BookIDs {
		stepState.BookID = bookID
		if ctx, err := s.schoolAdminCreateAtopicAndAChapter(ctx); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		ctx, los := s.prepareLOV1(ctx, stepState.TopicID, n, cpb.LearningObjectiveType_LEARNING_OBJECTIVE_TYPE_FLASH_CARD)
		if ctx, err := s.createLOV1(contextWithValidVersion(ctx), los); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create los: %w", err)
		}
		courseIDs := stepState.BookCourseMap[stepState.BookID]
		for _, lo := range los {
			loIDs = append(loIDs, lo.Info.Id)
		}
		for _, courseID := range courseIDs {
			for _, lo := range los {
				loContents = append(loContents, &entities.ContentStructure{
					CourseID:  courseID,
					BookID:    stepState.BookID,
					ChapterID: stepState.ChapterID,
					TopicID:   stepState.TopicID,
					LoID:      lo.Info.Id,
				})
			}
		}
	}
	stepState.LoIDs = loIDs
	stepState.LOs = loContents
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) hasCreatedSomeStudyplansExactMatchWithSomeBooksContentForStudent(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = contextWithToken(s, ctx)

	n := rand.Intn(1) + 2
	studyPlanIDs := make([]string, 0, n)
	m := make(map[string][]string)
	// create study plans with different book and different course
	for i := 1; i <= n; i++ {
		if ctx, err := s.createBook(ctx); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create book: %w", err)
		}
		if ctx, err := s.createACourse(ctx); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create course: %w", err)
		}
		if ctx, err := s.userAddCourseToStudent(ctx); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if ctx, err := s.userCreateStudyPlan(ctx); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if _, ok := m[stepState.BookID]; !ok {
			m[stepState.BookID] = []string{}
		}
		m[stepState.BookID] = append(m[stepState.BookID], stepState.CourseID)
		studyPlanIDs = append(studyPlanIDs, stepState.StudyPlanID)
	}
	stepState.BookCourseMap = m
	stepState.StudyPlanIDs = studyPlanIDs
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) hasCreatedSomeStudyplansExactMatchWithTheBookContentForStudent(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = contextWithToken(s, ctx)
	n := rand.Intn(1) + 2
	studyPlanIDs := make([]string, 0, n)
	m := make(map[string][]string)
	// create study plans with same book and different course
	for i := 1; i <= n; i++ {
		if ctx, err := s.createACourse(ctx); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create course: %w", err)
		}
		if ctx, err := s.userAddCourseToStudent(ctx); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if ctx, err := s.userCreateStudyPlan(ctx); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if _, ok := m[stepState.BookID]; !ok {
			m[stepState.BookID] = []string{}
		}
		m[stepState.BookID] = append(m[stepState.BookID], stepState.CourseID)
		studyPlanIDs = append(studyPlanIDs, stepState.StudyPlanID)
	}
	stepState.BookCourseMap = m
	stepState.StudyPlanIDs = studyPlanIDs
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userAddBookToCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courseBook := entities.CoursesBooks{}
	now := time.Now()
	if err := multierr.Combine(
		courseBook.BookID.Set(stepState.BookID),
		courseBook.CourseID.Set(stepState.CourseID),
		courseBook.CreatedAt.Set(now),
		courseBook.UpdatedAt.Set(now),
		courseBook.DeletedAt.Set(nil),
	); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to set value : %w", err)
	}

	if _, err := database.Insert(ctx, &courseBook, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to add book to course: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareLOV1(ctx context.Context, topicID string, numberOfLOs int, loType cpb.LearningObjectiveType) (context.Context, []*cpb.LearningObjective) {
	stepState := StepStateFromContext(ctx)
	pbLOs := make([]*cpb.LearningObjective, 0, numberOfLOs)
	for i := 0; i < numberOfLOs; i++ {
		stepState.LoID = idutil.ULIDNow()
		stepState.LoIDs = append(stepState.LoIDs, stepState.LoID)
		pbLOs = append(pbLOs, &cpb.LearningObjective{
			Info: &cpb.ContentBasicInfo{
				Id:           stepState.LoID,
				Name:         fmt.Sprintf("lo-%s-name+%s", loType.String(), stepState.LoID),
				Country:      cpb.Country_COUNTRY_VN,
				Grade:        stepState.Grade,
				Subject:      cpb.Subject_SUBJECT_BIOLOGY,
				DisplayOrder: int32(i + 1),

				SchoolId: stepState.SchoolIDInt,
			},
			Type:    loType,
			TopicId: topicID,
		})
	}
	return StepStateToContext(ctx, stepState), pbLOs
}

func (s *suite) createLOV1(ctx context.Context, pbLOs []*cpb.LearningObjective) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp, err := epb.NewLearningObjectiveModifierServiceClient(s.Conn).UpsertLOs(ctx, &epb.UpsertLOsRequest{
		LearningObjectives: pbLOs,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create learning objective: %w", err)
	}
	if resp.GetLoIds() == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable create LO: empty")
	}
	return StepStateToContext(ctx, stepState), nil
}
