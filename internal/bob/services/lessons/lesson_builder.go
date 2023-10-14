package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/services/courses"
	coursesRepo "github.com/manabie-com/backend/internal/bob/services/courses/repo"
	media_builder "github.com/manabie-com/backend/internal/bob/services/media"
	mediaRepo "github.com/manabie-com/backend/internal/bob/services/media/repo"
	topicsRepo "github.com/manabie-com/backend/internal/bob/services/topics/repo"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

// LessonBuilder implement builder for Lesson
type LessonBuilder struct {
	Lesson   *entities.Lesson
	SchoolID int32
	Country  string
	LessonRepo
	LessonGroupRepo
	coursesRepo.CourseRepo
	coursesRepo.PresetStudyPlanRepo
	coursesRepo.PresetStudyPlanWeeklyRepo
	topicsRepo.TopicRepo
	mediaRepo.MediaRepo
	UserRepo
	SchoolAdminRepo
	TeacherRepo
	StudentRepo
}

func NewLessonBuilder(
	lessonRp LessonRepo,
	lessonGroupRp LessonGroupRepo,
	courseRp coursesRepo.CourseRepo,
	presetStudyPlanRp coursesRepo.PresetStudyPlanRepo,
	presetStudyPlanWeeklyRp coursesRepo.PresetStudyPlanWeeklyRepo,
	topicRp topicsRepo.TopicRepo,
	mediaRp mediaRepo.MediaRepo,
	userRp UserRepo,
	schoolAdminRp SchoolAdminRepo,
	teacherRp TeacherRepo,
	studentRp StudentRepo,
) *LessonBuilder {
	return &LessonBuilder{
		LessonRepo:                lessonRp,
		LessonGroupRepo:           lessonGroupRp,
		CourseRepo:                courseRp,
		PresetStudyPlanRepo:       presetStudyPlanRp,
		PresetStudyPlanWeeklyRepo: presetStudyPlanWeeklyRp,
		TopicRepo:                 topicRp,
		MediaRepo:                 mediaRp,
		UserRepo:                  userRp,
		SchoolAdminRepo:           schoolAdminRp,
		TeacherRepo:               teacherRp,
		StudentRepo:               studentRp,
	}
}

func (l *LessonBuilder) insertLesson(ctx context.Context, db database.Ext) error {
	res, err := l.LessonRepo.Create(ctx, db, l.Lesson)
	if err != nil {
		return err
	}

	err = l.LessonRepo.UpsertLessonCourses(ctx, db, l.Lesson.LessonID, l.Lesson.CourseIDs.CourseIDs)
	if err != nil {
		return err
	}

	err = l.LessonRepo.UpsertLessonTeachers(ctx, db, l.Lesson.LessonID, l.Lesson.TeacherIDs.TeacherIDs)
	if err != nil {
		return err
	}

	err = l.LessonRepo.UpsertLessonMembers(ctx, db, l.Lesson.LessonID, l.Lesson.LearnerIDs.LearnerIDs)
	if err != nil {
		return err
	}
	l.Lesson = res

	return nil
}

// HACK: createLessonGroup will create a lesson group which will hold
// relationship first course and all medias
func (l *LessonBuilder) createLessonGroup(ctx context.Context, db database.Ext, medias entities.Medias) (*entities.LessonGroup, error) {
	// check existing media ids
	existingMediaIDs := medias.GetUniqueIDs()
	err := l.checkExistingMedias(ctx, db, existingMediaIDs)
	if err != nil {
		return nil, fmt.Errorf("checkExistingMedias: %s", err)
	}

	// create new medias
	newMediaIDs, err := l.createNewMedias(ctx, db, medias)
	if err != nil {
		return nil, fmt.Errorf("createNewMedias: %s", err)
	}
	existingMediaIDs = database.AppendTextArray(existingMediaIDs, newMediaIDs)

	lgModifier := NewLessonGroupModifier(db, l.LessonGroupRepo)
	lessonGr, err := lgModifier.CreateWithMedias(ctx, l.Lesson.CourseID, existingMediaIDs)
	if err != nil {
		return nil, err
	}

	return lessonGr, nil
}

func (l *LessonBuilder) createNewMedias(ctx context.Context, db database.Ext, medias entities.Medias) (pgtype.TextArray, error) {
	newMedias := medias.GetUncreatedMedias()
	mediaBuilder := media_builder.NewMediaBuilder(db, l.MediaRepo)
	newMedias, err := mediaBuilder.Upsert(ctx, newMedias)
	if err != nil {
		return pgtype.TextArray{}, err
	}

	return newMedias.GetUniqueIDs(), nil
}

// checkExistingMedias will return error if media's ids not exist in db
func (l *LessonBuilder) checkExistingMedias(ctx context.Context, db database.Ext, existingMediaIDs pgtype.TextArray) error {
	mediaBuilder := media_builder.NewMediaBuilder(db, l.MediaRepo)
	err := mediaBuilder.CheckMediaIDs(ctx, existingMediaIDs)
	if err != nil {
		return err
	}

	return nil
}

func (l *LessonBuilder) updateCourseAvailableRanges(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) error {
	// find start time and end time in list lesson of every course
	avlRange, err := l.LessonRepo.FindEarliestAndLatestTimeLessonByCourses(ctx, db, courseIDs)
	if err != nil {
		return fmt.Errorf("LessonRepo.FindEarlierAndLatestTimeLessonByCourses: %s %s", err, database.FromTextArray(courseIDs))
	}

	// update courses with available range
	crsBuilder := courses.NewCourseBuilder(db, l.CourseRepo)
	err = crsBuilder.UpdateCourseAvailableRanges(ctx, avlRange)
	if err != nil {
		return fmt.Errorf("crsBuilder.UpdateCourseAvailableRanges: %s", err)
	}

	return nil
}

func (l *LessonBuilder) preCreate() error {
	// generate lesson id
	if err := l.Lesson.LessonID.Set(idutil.ULIDNow()); err != nil {
		return err
	}

	if err := l.Lesson.Normalize(); err != nil {
		return err
	}

	return nil
}

func (l *LessonBuilder) isValid(ctx context.Context, db database.Ext, schoolID *int32) error {
	if err := l.Lesson.IsValid(); err != nil {
		return err
	}

	// get all teachers
	teachers, err := l.TeacherRepo.Retrieve(ctx, db, l.Lesson.TeacherIDs.TeacherIDs)
	if err != nil {
		return fmt.Errorf("TeacherRepo.Retrieve: %s", err)
	}

	if schoolID == nil {
		// default get first school id of first teacher in lesson
		schoolID = &teachers[0].SchoolIDs.Elements[0].Int
	}

	// check school ID and country code of all teachers
	for _, teacher := range teachers {
		if !teacher.IsInSchool(*schoolID) {
			return fmt.Errorf("teacher %s is not belong to school %d", teacher.ID.String, *schoolID)
		}
	}

	// get all courses
	// check school ID and country code of all course
	cs, err := l.CourseRepo.FindByIDs(ctx, db, l.Lesson.CourseIDs.CourseIDs)
	if err != nil {
		return fmt.Errorf("CourseRepo.FindByIDs: %s", err)
	}

	for _, c := range cs {
		if c.SchoolID.Int != *schoolID {
			return fmt.Errorf("course %s is not belong to school %d", c.ID.String, *schoolID)
		}
	}

	// get all students
	learners, err := l.StudentRepo.Retrieve(ctx, db, l.Lesson.LearnerIDs.LearnerIDs)
	if err != nil {
		return fmt.Errorf("StudentRepo.Retrieve: %s", err)
	}

	// check school ID and country code of all learners
	for _, learner := range learners {
		if learner.School.ID.Int != *schoolID {
			return fmt.Errorf("student %s is not belong to school %d", learner.Student.ID.String, *schoolID)
		}
	}

	return nil
}

type Option func(context.Context, database.Ext, *LessonBuilder) error

func WithMedia(medias entities.Medias) Option {
	return func(ctx context.Context, db database.Ext, l *LessonBuilder) error {
		lessonGr, err := l.createLessonGroup(ctx, db, medias)
		if err != nil {
			return fmt.Errorf("createLessonGroup: %s", err)
		}
		l.Lesson.LessonGroupID = lessonGr.LessonGroupID

		return nil
	}
}

func (l *LessonBuilder) create(ctx context.Context, db database.Ext, opts ...Option) error {
	for _, opt := range opts {
		if err := opt(ctx, db, l); err != nil {
			return err
		}
	}

	// create preset study plans for courses
	pspBldr := courses.NewPresetStudyPlanBuilder(db, l.PresetStudyPlanRepo, l.CourseRepo)
	err := pspBldr.CreatePresetStudyPlansByCourseIDs(ctx, l.Lesson.CourseIDs.CourseIDs)
	if err != nil {
		return fmt.Errorf("pspBldr.CreatePresetStudyPlansByCourseIDs: %s", err)
	}

	// insert lesson
	err = l.insertLesson(ctx, db)
	if err != nil {
		return fmt.Errorf("insertLesson: %s", err)
	}

	// create preset study plan weeklies
	pSPWBuilder := courses.NewPresetStudyPlanWeeklyBuilder(db, l.CourseRepo, l.TopicRepo, l.PresetStudyPlanRepo, l.PresetStudyPlanWeeklyRepo)
	err = pSPWBuilder.CreatePresetStudyPlanWeekliesForLesson(ctx, l.Lesson)
	if err != nil {
		return fmt.Errorf("pSPWBuilder.CreatePresetStudyPlanWeekliesForLesson: %s", err)
	}

	err = l.updateCourseAvailableRanges(ctx, db, l.Lesson.CourseIDs.CourseIDs)
	if err != nil {
		return fmt.Errorf("updateCourseAvailableRanges: %s", err)
	}

	return nil
}

// Create will create a lesson with list medias
// Steps to create:
//   - Check existing medias and create new medias.
//   - Hack: create a lesson group to hold list medias by first course
//     of lesson.
//   - Create preset study plans for every courses of lesson if they have
//     not preset study plan.
//   - Create preset study plan weeklies for every above study plans and
//     lesson.
//   - Update course's start date and end date by min(start time) and
//     max(end time) of all lessons.
//
// Always using pgx.Tx to executive query in db.
func (l *LessonBuilder) Create(ctx context.Context, db database.Ext, lesson *entities.Lesson, schoolID *int32, opts ...Option) (err error) {
	l.Lesson = lesson
	if err = l.preCreate(); err != nil {
		return err
	}

	if err = l.isValid(ctx, db, schoolID); err != nil {
		return err
	}

	switch db.(type) {
	case pgx.Tx:
		err = l.create(ctx, db, opts...)
	default:
		err = database.ExecInTx(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
			if err := l.create(ctx, tx, opts...); err != nil {
				return err
			}

			return nil
		})
	}
	if err != nil {
		return err
	}

	return nil
}

// UpdateWithMedias updates the lesson with medias.
func (l *LessonBuilder) UpdateWithMedias(ctx context.Context, db database.Ext, lesson *entities.Lesson, medias entities.Medias) error {
	l.Lesson = lesson
	if err := l.Lesson.Normalize(); err != nil {
		return fmt.Errorf("Lesson.Normalize: %s", err)
	}

	// Query current lesson in database
	currentLesson, err := l.getCurrentLesson(ctx, db)
	if err != nil {
		return fmt.Errorf("getCurrentLesson: %s", err)
	}

	// get current school id and country code
	course, err := l.CourseRepo.FindByID(ctx, db, currentLesson.CourseIDs.CourseIDs.Elements[0])
	if err != nil {
		return fmt.Errorf("CourseRepo.FindByID: %s", err)
	}

	if err = l.isValid(ctx, db, &course.SchoolID.Int); err != nil {
		return fmt.Errorf("Lesson.IsValid: %s", err)
	}

	switch db := db.(type) {
	case pgx.Tx:
		return l.update(ctx, db, currentLesson, medias)
	default:
		return database.ExecInTx(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
			return l.update(ctx, tx, currentLesson, medias)
		})
	}
}

func (l *LessonBuilder) update(ctx context.Context, tx pgx.Tx, currentLesson *entities.Lesson, medias entities.Medias) error { //nolint: interfacer
	// Update the media in lesson_groups
	l.Lesson.LessonGroupID = currentLesson.LessonGroupID
	if currentLesson.CourseID.String == l.Lesson.CourseID.String {
		err := l.updateMedias(ctx, tx, currentLesson.LessonGroupID, l.Lesson.CourseID, medias)
		if err != nil {
			return fmt.Errorf("updateMedias: %s", err)
		}
	} else {
		lessonGr, err := l.createLessonGroup(ctx, tx, medias)
		if err != nil {
			return fmt.Errorf("createLessonGroup: %s", err)
		}
		l.Lesson.LessonGroupID = lessonGr.LessonGroupID
	}

	// Add a preset_study_plan for each course that doesn't have it
	// The current preset_study_plans are kept unchanged
	pspBuilder := courses.NewPresetStudyPlanBuilder(tx, l.PresetStudyPlanRepo, l.CourseRepo)
	err := pspBuilder.CreatePresetStudyPlansByCourseIDs(ctx, l.Lesson.CourseIDs.CourseIDs)
	if err != nil {
		return fmt.Errorf("pspBuilder.CreatePresetStudyPlanByCourseIDs: %s", err)
	}

	// Update preset_study_plans_weekly and their topics
	pspwBuilder := courses.NewPresetStudyPlanWeeklyBuilder(tx, l.CourseRepo, l.TopicRepo, l.PresetStudyPlanRepo, l.PresetStudyPlanWeeklyRepo)
	err = pspwBuilder.UpsertPresetStudyPlanWeekliesForLesson(ctx, *currentLesson, *l.Lesson)
	if err != nil {
		return fmt.Errorf("pspwBuilder.UpsertPresetStudyPlanWeekliesForLesson: %s", err)
	}

	// Update the lesson itself
	err = l.LessonRepo.Update(ctx, tx, l.Lesson)
	if err != nil {
		return fmt.Errorf("LessonRepo.Update: %s", err)
	}
	err = l.LessonRepo.UpsertLessonCourses(ctx, tx, l.Lesson.LessonID, l.Lesson.CourseIDs.CourseIDs)
	if err != nil {
		return fmt.Errorf("LessonRepo.UpsertLessonCourses: %s", err)
	}
	err = l.LessonRepo.UpsertLessonTeachers(ctx, tx, l.Lesson.LessonID, l.Lesson.TeacherIDs.TeacherIDs)
	if err != nil {
		return fmt.Errorf("LessonRepo.UpsertLessonTeachers: %s", err)
	}
	err = l.LessonRepo.UpsertLessonMembers(ctx, tx, l.Lesson.LessonID, l.Lesson.LearnerIDs.LearnerIDs)
	if err != nil {
		return fmt.Errorf("LessonRepo.UpsertLessonMembers: %s", err)
	}

	// Update start/end time for all affected courses
	allCourseIDs := database.FromTextArray(l.Lesson.CourseIDs.CourseIDs)
	allCourseIDs = append(allCourseIDs, database.FromTextArray(currentLesson.CourseIDs.CourseIDs)...)
	allCourseIDs = golibs.Uniq(allCourseIDs)
	err = l.CourseRepo.UpdateStartAndEndDate(ctx, tx, database.TextArray(allCourseIDs))
	if err != nil {
		return fmt.Errorf("CourseRepo.UpdateStartAndEndDate: %s", err)
	}
	return nil
}

// getCurrentLesson queries the current lesson from database.
// It also populates the CourseIDs field for lesson.
func (l *LessonBuilder) getCurrentLesson(ctx context.Context, db database.Ext) (*entities.Lesson, error) {
	currentLesson, err := l.LessonRepo.FindByID(ctx, db, l.Lesson.LessonID)
	if err != nil {
		return nil, fmt.Errorf("LessonRepo.FindByID: %s", err)
	}

	currentCourseIDs, err := l.LessonRepo.GetCourseIDsOfLesson(ctx, db, l.Lesson.LessonID)
	if err != nil {
		return nil, fmt.Errorf("LessonRepo.GetCourseIDsByLesson: %s", err)
	}
	currentLesson.CourseIDs.CourseIDs = currentCourseIDs
	return currentLesson, nil
}

func (l *LessonBuilder) updateMedias(ctx context.Context, db database.Ext, lessonGroupID, updatedCourseID pgtype.Text, updatedMedias entities.Medias) error {
	// For existing media IDs, they must be in database
	existingMediaIDs := updatedMedias.GetUniqueIDs()
	err := l.checkExistingMedias(ctx, db, existingMediaIDs)
	if err != nil {
		return fmt.Errorf("checkExistingMedias: %s", err)
	}

	// For uncreated media IDs (brightcove videos), create them in db
	// TODO: remove this since frontend only sends media with IDs
	newMediaIDs, err := l.createNewMedias(ctx, db, updatedMedias)
	if err != nil {
		return fmt.Errorf("createNewMedias: %s", err)
	}
	allMediaIDs := database.AppendTextArray(existingMediaIDs, newMediaIDs)
	err = l.updateLessonGroup(ctx, db, lessonGroupID, updatedCourseID, allMediaIDs)
	if err != nil {
		return fmt.Errorf("updateLessonGroup: %s", err)
	}
	return nil
}

func (l *LessonBuilder) updateLessonGroup(ctx context.Context, db database.QueryExecer, lessonGroupID, updatedCourseID pgtype.Text, updatedMediaIDs pgtype.TextArray) error {
	lg := &entities.LessonGroup{}
	database.AllNullEntity(lg)
	lg.LessonGroupID = lessonGroupID
	lg.CourseID = updatedCourseID
	lg.MediaIDs = updatedMediaIDs
	err := l.LessonGroupRepo.UpdateMedias(ctx, db, lg)
	if err != nil {
		return fmt.Errorf("LessonGroupRepo.UpdateMedias: %s", err)
	}
	return nil
}

func (l *LessonBuilder) delete(ctx context.Context, db database.Ext, lessonID string) error {
	lesson, err := l.LessonRepo.FindByID(ctx, db, database.Text(lessonID))
	if err != nil {
		return fmt.Errorf("LessonRepo.FindByID: %w", err)
	}

	if !lesson.Deletable() {
		return fmt.Errorf("only can delete not started live lesson")
	}

	// find all preset study plan weeklies of lesson
	presetStudyPlanWeeklies, err := l.PresetStudyPlanWeeklyRepo.FindByLessonIDs(
		ctx,
		db,
		database.TextArray([]string{lesson.LessonID.String}),
		false,
	)
	if err != nil {
		return fmt.Errorf("PresetStudyPlanWeeklyRepo.FindByLessonIDs: %w", err)
	}
	if len(presetStudyPlanWeeklies) == 0 {
		return fmt.Errorf("could not find preset study plan weekly")
	}

	pSPWs := make([]*entities.PresetStudyPlanWeekly, 0, len(presetStudyPlanWeeklies))
	for i := range presetStudyPlanWeeklies {
		pSPWs = append(pSPWs, presetStudyPlanWeeklies[i])
	}

	// delete preset study plan weeklies of lesson
	pSPWBuilder := courses.NewPresetStudyPlanWeeklyBuilder(db, l.CourseRepo, l.TopicRepo, l.PresetStudyPlanRepo, l.PresetStudyPlanWeeklyRepo)
	err = pSPWBuilder.Delete(ctx, pSPWs)
	if err != nil {
		return fmt.Errorf("could not delete preset study plan weeklies: %v", err)
	}

	// save courses of lesson
	courses, err := l.CourseRepo.FindByLessonID(ctx, db, lesson.LessonID)
	if err != nil {
		return fmt.Errorf("CourseRepo.FindByLessonID: %w", err)
	}

	// delete lesson
	err = l.LessonRepo.DeleteLessonCourses(ctx, db, lesson.LessonID)
	if err != nil {
		return fmt.Errorf("LessonRepo.DeleteLessonCourses: %w", err)
	}

	err = l.LessonRepo.DeleteLessonTeachers(ctx, db, lesson.LessonID)
	if err != nil {
		return fmt.Errorf("LessonRepo.DeleteLessonTeachers: %w", err)
	}

	err = l.LessonRepo.DeleteLessonMembers(ctx, db, lesson.LessonID)
	if err != nil {
		return fmt.Errorf("LessonRepo.DeleteLessonMembers: %w", err)
	}

	err = l.LessonRepo.Delete(ctx, db, database.TextArray([]string{lessonID}))
	if err != nil {
		return fmt.Errorf("LessonRepo.Delete: %s %w", lessonID, err)
	}

	// update courses
	courseIDs := make([]string, 0, len(courses))
	for _, course := range courses {
		courseIDs = append(courseIDs, course.ID.String)
	}
	err = l.updateCourseAvailableRanges(ctx, db, database.TextArray(courseIDs))
	if err != nil {
		return err
	}

	return nil
}

func (l *LessonBuilder) Delete(ctx context.Context, db database.Ext, lessonID string) (err error) {
	switch db.(type) {
	case pgx.Tx:
		err = l.delete(ctx, db, lessonID)
	default:
		err = database.ExecInTx(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
			if err := l.delete(ctx, tx, lessonID); err != nil {
				return err
			}

			return nil
		})
	}
	if err != nil {
		return err
	}

	return nil
}
