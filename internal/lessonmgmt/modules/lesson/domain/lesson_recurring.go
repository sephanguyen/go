package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
)

type RecurringLesson struct {
	ID      string
	Lessons []*Lesson
}

func (ls *RecurringLesson) Save() {
	now := time.Now()
	for _, l := range ls.Lessons {
		if !l.Persisted {
			l.preInsert()
			l.CreatedAt = now
		}
		l.UpdatedAt = now
	}
}

func (ls *RecurringLesson) IsValid(ctx context.Context, db database.Ext) error {
	if len(ls.Lessons) == 0 {
		return nil
	}
	// just check first lesson
	if err := ls.Lessons[0].IsValid(ctx, db); err != nil {
		return err
	}
	return nil
}

func (ls *RecurringLesson) GetBaseLesson() *Lesson {
	if len(ls.Lessons) == 0 {
		return nil
	}
	return ls.Lessons[0]
}

func (ls *RecurringLesson) GetIDs() []string {
	lessonIDs := []string{}
	for _, l := range ls.Lessons {
		lessonIDs = append(lessonIDs, l.LessonID)
	}
	return lessonIDs
}

func (ls *RecurringLesson) GetLessonCourses() (result []string) {
	if len(ls.Lessons) == 0 {
		return
	}
	baseLesson := ls.GetBaseLesson()
	courseMap := map[string]bool{}

	if baseLesson.TeachingMethod == LessonTeachingMethodGroup {
		if baseLesson.CourseID != "" {
			courseMap[baseLesson.CourseID] = true
			result = append(result, baseLesson.CourseID)
		}
	} else {
		for _, learner := range baseLesson.Learners {
			courseID := learner.CourseID
			if _, ok := courseMap[courseID]; !ok {
				result = append(result, courseID)
				courseMap[courseID] = true
			}
		}
	}
	return result
}

type FollowingLessonID []string

func (s *FollowingLessonID) IsEmpty() bool {
	return len(*s) == 0
}

func (s *FollowingLessonID) Add(str string) {
	*s = append(*s, str)
}

func (s *FollowingLessonID) Pop() (string, bool) {
	if s.IsEmpty() {
		return "", false
	}
	element := (*s)[0]
	*s = (*s)[1:]
	return element, true
}

func (s *FollowingLessonID) Remaining() []string {
	return (*s)
}

// in the following lessonIds, remove many lessonIds that lessons are locked from input
func (s *FollowingLessonID) GetNoLockedLessons(lockedLesson []string) (res []string) {
	if len(lockedLesson) == 0 {
		return (*s)
	}
	hash := make(map[string]bool)
	for _, item := range lockedLesson {
		hash[item] = true
	}
	for _, item := range *s {
		if !hash[item] {
			res = append(res, item)
		}
	}
	return
}

type StateChangedLesson struct {
	ChanTime     bool
	ChanLocation bool
}

func (s *StateChangedLesson) IsChanged() bool {
	return s.ChanTime || s.ChanLocation
}

type RecurrenceRule struct {
	Option Option
}

type Frequency int

const (
	ONCE Frequency = iota
	WEEKLY
)

var FrequencyName = map[Frequency]string{
	ONCE:   "once",
	WEEKLY: "weekly",
}

type Option struct {
	Freq         Frequency
	StartTime    time.Time
	EndTime      time.Time
	UntilDate    time.Time
	ExcludeDates map[string]string
}

func NewRecurrenceRule(opt Option) (RecurrenceRule, error) {
	r := RecurrenceRule{}
	now := time.Now()
	if opt.StartTime.IsZero() {
		opt.StartTime = now
	}
	if opt.EndTime.IsZero() {
		opt.StartTime = now
	}
	if opt.EndTime.IsZero() {
		opt.EndTime = time.Date(now.Year(), time.December, 31, 0, 0, 0, 0, time.UTC)
	}
	if opt.StartTime.Format(Ymd) > opt.UntilDate.Format(Ymd) {
		return r, fmt.Errorf("startTime could not be greater than utilDate")
	}
	if opt.EndTime.Format(Ymd) > opt.UntilDate.Format(Ymd) {
		return r, fmt.Errorf("endTime could not be greater than utilDate")
	}
	if opt.UntilDate.IsZero() {
		opt.UntilDate = opt.EndTime
	}
	r.Option = opt
	return r, nil
}

type RecurringSet struct {
	StartTime time.Time
	EndTime   time.Time
}

const Ymd = "2006-01-02"

func (r *RecurrenceRule) All() []RecurringSet {
	var years, months, days int
	switch r.Option.Freq {
	case WEEKLY:
		days = 7
		years = 0
		months = 0
	default:
		return []RecurringSet{}
	}
	events := make([]RecurringSet, 0)
	freqStart := r.Option.StartTime
	freqEnd := r.Option.EndTime

	for {
		weekDayStart := freqStart
		weekDateEnd := freqEnd
		isSkipDate := false
		for key, element := range r.Option.ExcludeDates {
			loc, _ := time.LoadLocation(element)
			if weekDayStart.In(loc).Format(Ymd) == key {
				isSkipDate = true
			}
		}

		if !isSkipDate {
			e := RecurringSet{
				StartTime: weekDayStart,
				EndTime:   weekDateEnd,
			}
			events = append(events, e)
		}
		freqStart = freqStart.AddDate(years, months, days)
		freqEnd = freqEnd.AddDate(years, months, days)
		if r.Option.UntilDate.Format(Ymd) < freqStart.Format(Ymd) {
			break
		}
	}
	return events
}

func (r *RecurrenceRule) ExceptFirst() []RecurringSet {
	set := r.All()
	if len(set) > 0 {
		return set[1:]
	}
	return []RecurringSet{}
}
