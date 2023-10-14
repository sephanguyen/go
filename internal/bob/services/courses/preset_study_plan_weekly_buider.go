package courses

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/services/courses/repo"
	"github.com/manabie-com/backend/internal/bob/services/topics"
	topicsRepo "github.com/manabie-com/backend/internal/bob/services/topics/repo"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type PresetStudyPlanWeeklyBuilder struct {
	db                        database.Ext
	courseRepo                repo.CourseRepo
	topicRepo                 topicsRepo.TopicRepo
	presetStudyPlanRepo       repo.PresetStudyPlanRepo
	presetStudyPlanWeeklyRepo repo.PresetStudyPlanWeeklyRepo
}

func NewPresetStudyPlanWeeklyBuilder(
	db database.Ext,
	courseRepo repo.CourseRepo,
	topicRepo topicsRepo.TopicRepo,
	presetStudyPlanRepo repo.PresetStudyPlanRepo,
	presetStudyPlanWeeklyRepo repo.PresetStudyPlanWeeklyRepo,
) *PresetStudyPlanWeeklyBuilder {
	return &PresetStudyPlanWeeklyBuilder{
		db:                        db,
		courseRepo:                courseRepo,
		topicRepo:                 topicRepo,
		presetStudyPlanRepo:       presetStudyPlanRepo,
		presetStudyPlanWeeklyRepo: presetStudyPlanWeeklyRepo,
	}
}

func (p *PresetStudyPlanWeeklyBuilder) initializePresetStudyPlanWeekly(presetStudyPlanID, lessonID, topicID string, startTime, endTime time.Time) (*entities.PresetStudyPlanWeekly, error) {
	e := &entities.PresetStudyPlanWeekly{}
	database.AllNullEntity(e)
	err := multierr.Combine(
		e.ID.Set(idutil.ULIDNow()),
		e.PresetStudyPlanID.Set(presetStudyPlanID),
		e.TopicID.Set(topicID),
		e.DeletedAt.Set(nil),
		e.StartDate.Set(startTime),
		e.EndDate.Set(endTime),
		e.LessonID.Set(lessonID),
		e.Week.Set(0),
	)
	if err != nil {
		return nil, fmt.Errorf("could not create preset study plan weekly: %w", err)
	}
	if err = e.IsValid(); err != nil {
		return nil, err
	}

	return e, nil
}

// CreatePresetStudyPlanWeekliesForLesson will create list preset_study_plan_weekly
// which hold data of a preset study plan (1-1 course) in a particular time range,
// topic id (which this case will hold relationship about "data" between a lesson and
// a course) and lesson id
func (p *PresetStudyPlanWeeklyBuilder) CreatePresetStudyPlanWeekliesForLesson(ctx context.Context, lesson *entities.Lesson) error {
	courseIDs := lesson.CourseIDs.CourseIDs
	if courseIDs.Status != pgtype.Present || len(courseIDs.Elements) == 0 {
		return nil
	}

	return p.createForLessonAndCourseIDs(ctx, *lesson, courseIDs)
}

// For preset_study_plans_weekly:
//   - Create preset_study_plans_weekly for lesson for added courses
//   - Delete preset_study_plans_weekly for lesson for removed courses
//   - Also add/remove topic for each preset_study_plans_weekly
//   - Update name for all topics for relating preset_study_plans_weekly
func (p *PresetStudyPlanWeeklyBuilder) UpsertPresetStudyPlanWeekliesForLesson(ctx context.Context, oldLesson, newLesson entities.Lesson) error {
	removedCourseIDs, addedCourseIDs, unchangedCourseIDs := p.getRemovedAndAddedAndUnchangedCourse(oldLesson, newLesson)

	// For added courses, add new preset_study_plans_weekly similarly to create flow
	err := p.createForLessonAndCourseIDs(ctx, newLesson, addedCourseIDs)
	if err != nil {
		return fmt.Errorf("createForLessonAndCourseIDs: %s", err)
	}

	// For removed courses, delete preset_study_plans_weekly and topics belonging to them
	err = p.deleteForLessonAndCourseIDs(ctx, newLesson, removedCourseIDs)
	if err != nil {
		return fmt.Errorf("deleteForLessonAndCourseIDs: %s", err)
	}

	// Update name for all topics, both old and new
	if oldLesson.Name != newLesson.Name {
		err := p.topicRepo.UpdateNameByLessonID(ctx, p.db, oldLesson.LessonID, newLesson.Name)
		if err != nil {
			return fmt.Errorf("TopicRepo.UpdateNameByLessonID: %s", err)
		}
	}

	// Update start/end time for PSPWs.
	// Since for new PSPWs, their start/end time already follow new lesson start/end time,
	// we only need to do update for unchanged PSPWs.
	if (!newLesson.StartTime.Time.Equal(oldLesson.StartTime.Time) || !newLesson.EndTime.Time.Equal(oldLesson.EndTime.Time)) && len(unchangedCourseIDs.Elements) > 0 {
		err := p.presetStudyPlanWeeklyRepo.UpdateTimeByLessonAndCourses(
			ctx, p.db, newLesson.LessonID, unchangedCourseIDs, newLesson.StartTime, newLesson.EndTime,
		)
		if err != nil {
			return fmt.Errorf("presetStudyPlanWeeklyRepo.UpdateTimeByLessonAndCourses: %s", err)
		}
	}
	return nil
}

func (p *PresetStudyPlanWeeklyBuilder) getRemovedAndAddedAndUnchangedCourse(oldLesson, newLesson entities.Lesson) (removedIDs, addedIDs, unchangedIDs pgtype.TextArray) {
	oldCourseIDMap := database.MapFromTextArray(oldLesson.CourseIDs.CourseIDs)
	newCourseIDMap := database.MapFromTextArray(newLesson.CourseIDs.CourseIDs)

	removedIDs = pgtype.TextArray{Status: pgtype.Null}
	addedIDs = pgtype.TextArray{Status: pgtype.Null}
	unchangedIDs = pgtype.TextArray{Status: pgtype.Null}
	for cID := range newCourseIDMap {
		if _, ok := oldCourseIDMap[cID]; ok {
			unchangedIDs = database.AppendText(unchangedIDs, cID)
		} else {
			addedIDs = database.AppendText(addedIDs, cID)
		}
	}
	for cID := range oldCourseIDMap {
		if _, ok := newCourseIDMap[cID]; !ok {
			removedIDs = database.AppendText(removedIDs, cID)
		}
	}
	return
}

// createForLessonAndCourseIDs create preset study plans weekly for lesson and courseIDs.
func (p *PresetStudyPlanWeeklyBuilder) createForLessonAndCourseIDs(ctx context.Context, lesson entities.Lesson, courseIDs pgtype.TextArray) error {
	if len(courseIDs.Elements) == 0 {
		return nil
	}

	// TODO: courseRepo.FindByIDs is called multiple times for both CreateLiveLesson and UpdateLiveLesson.
	// Maybe we can refactor this.
	courses, err := p.courseRepo.FindByIDs(ctx, p.db, courseIDs)
	if err != nil {
		return fmt.Errorf("courseRepo.FindByIDs: %s", err)
	}
	if len(courses) != len(courseIDs.Elements) {
		return fmt.Errorf("expect %d course from database, found %d", len(courseIDs.Elements), len(courses))
	}
	for _, course := range courses {
		if course.PresetStudyPlanID.Status != pgtype.Present || len(course.PresetStudyPlanID.String) == 0 {
			return fmt.Errorf("course %s does not have preset study plan id", course.ID.String)
		}
	}

	// Create topics for each preset_study_plan_weekly first
	topicBuilder := topics.NewTopicBuilder(p.db, p.courseRepo, p.topicRepo)
	topicByCourseID, err := topicBuilder.CreateTopicByLiveLessonAndCourses(ctx, lesson, courses)
	if err != nil {
		return fmt.Errorf("topicBuilder.CreateTopicByLiveLessonAndCourses: %s", err)
	}
	if len(topicByCourseID) != len(courseIDs.Elements) {
		return fmt.Errorf("expect %d topic created, found %d", len(courseIDs.Elements), len(topicByCourseID))
	}

	// Initialize preset_study_plans_weekly and insert them to database
	pspws := make([]*entities.PresetStudyPlanWeekly, 0, len(topicByCourseID))
	for courseID, course := range courses {
		topic, ok := topicByCourseID[courseID]
		if !ok {
			return fmt.Errorf("topic for course ID \"%s\" was not created", courseID.String)
		}

		pspw, err := p.initializePresetStudyPlanWeekly(course.PresetStudyPlanID.String, lesson.LessonID.String, topic.ID.String, lesson.StartTime.Time, lesson.EndTime.Time)
		if err != nil {
			return fmt.Errorf("initializePresetStudyPlanWeekly: %s", err)
		}
		pspws = append(pspws, pspw)
	}
	err = p.presetStudyPlanWeeklyRepo.Create(ctx, p.db, pspws)
	if err != nil {
		return fmt.Errorf("presetStudyPlanWeeklyRepo.Create: %s", err)
	}

	return nil
}

// deleteForLessonAndCourseIDs deletes preset study plans weekly for lesson and courseIDs.
func (p *PresetStudyPlanWeeklyBuilder) deleteForLessonAndCourseIDs(ctx context.Context, lesson entities.Lesson, courseIDs pgtype.TextArray) error {
	if len(courseIDs.Elements) == 0 {
		return nil
	}

	// Get IDs of preset_study_plans of all the courses being deleted
	pspIDs, err := p.courseRepo.GetPresetStudyPlanIDsByCourseIDs(ctx, p.db, courseIDs)
	if err != nil {
		return fmt.Errorf("courseRepo.GetPresetStudyPlanIDsByCourseIDs: %s", err)
	}

	// Get IDs for all preset_study_plans_weekly related to this lesson and the courses
	pspwIDs, err := p.presetStudyPlanWeeklyRepo.GetIDsByLessonIDAndPresetStudyPlanIDs(ctx, p.db, lesson.LessonID, database.TextArray(pspIDs))
	if err != nil {
		return fmt.Errorf("presetStudyPlanWeeklyRepo.GetIDsByLessonIDAndPresetStudyPlanIDs: %s", err)
	}

	// Delete all for both topics and preset_study_plans_weekly
	err = p.topicRepo.SoftDeleteByPresetStudyPlanWeeklyIDs(ctx, p.db, database.TextArray(pspwIDs))
	if err != nil {
		return fmt.Errorf("topicRepo.SoftDeleteByPresetStudyPlanWeeklyIDs: %s", err)
	}
	err = p.presetStudyPlanWeeklyRepo.SoftDelete(ctx, p.db, database.TextArray(pspwIDs))
	if err != nil {
		return fmt.Errorf("presetStudyPlanWeeklyRepo.SoftDelete: %s", err)
	}

	return nil
}

func (p *PresetStudyPlanWeeklyBuilder) Delete(ctx context.Context, items []*entities.PresetStudyPlanWeekly) error {
	// get list topic's ids
	pSPWIDs := make([]string, 0, len(items))
	topicIDs := make([]string, 0, len(items))
	for _, item := range items {
		topicIDs = append(topicIDs, item.TopicID.String)
		pSPWIDs = append(pSPWIDs, item.ID.String)
	}

	err := p.topicRepo.SoftDeleteV2(ctx, p.db, database.TextArray(topicIDs))
	if err != nil {
		return fmt.Errorf("TopicRepo.SoftDeleteV2: %w", err)
	}

	err = p.presetStudyPlanWeeklyRepo.SoftDelete(ctx, p.db, database.TextArray(pSPWIDs))
	if err != nil {
		return fmt.Errorf("PresetStudyPlanWeeklyRepo.SoftDelete: %w", err)
	}

	return nil
}
