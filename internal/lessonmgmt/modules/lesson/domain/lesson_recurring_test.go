package domain

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRecurringLesson_CommonFunc(t *testing.T) {
	t.Parallel()
	now := time.Now()
	rl := RecurringLesson{
		ID: "chain-1",
		Lessons: []*Lesson{
			{
				LessonID:  "l1",
				Persisted: true,
				Learners: LessonLearners{
					{
						LearnerID: "learner-1",
						CourseID:  "course-1",
					},
					{
						LearnerID: "learner-2",
						CourseID:  "course-1",
					},
					{
						LearnerID: "learner-3",
						CourseID:  "course-2",
					},
				},
			},
			{
				LessonID: "",
			},
			{
				LessonID: "",
			},
		},
	}
	t.Run("Save", func(t *testing.T) {
		rl.Save()
		for _, l := range rl.Lessons {
			if l.Persisted {
				require.WithinDuration(t, now, l.UpdatedAt, 100*time.Millisecond)
			} else {
				require.WithinDuration(t, now, l.CreatedAt, 100*time.Millisecond)
				require.WithinDuration(t, now, l.UpdatedAt, 100*time.Millisecond)
			}
		}
	})

	t.Run("GetBaseLesson", func(t *testing.T) {
		ls := rl.GetBaseLesson()
		require.Equal(t, "l1", ls.LessonID)
	})

	t.Run("GetIDs", func(t *testing.T) {
		for idx, l := range rl.Lessons {
			l.LessonID = fmt.Sprintf("l%d", idx+1)
		}
		ids := rl.GetIDs()
		require.Len(t, ids, len(rl.Lessons))
		require.Equal(t, []string{"l1", "l2", "l3"}, ids)
	})

	t.Run("GetLessonCourses", func(t *testing.T) {
		courseIds := rl.GetLessonCourses()
		require.Equal(t, []string{"course-1", "course-2"}, courseIds)
	})
}

func TestLesson_RecurrenceSet(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name             string
		startTime        time.Time
		endTime          time.Time
		untilDate        time.Time
		expectedRecurSet []RecurringSet
	}{
		{
			name:      "until date greater than last day within one day",
			startTime: time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
			endTime:   time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC),
			untilDate: time.Date(2022, 7, 12, 10, 0, 0, 0, time.UTC),
			expectedRecurSet: []RecurringSet{
				{
					StartTime: time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
					EndTime:   time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC),
				},
				{
					StartTime: time.Date(2022, 7, 4, 9, 0, 0, 0, time.UTC),
					EndTime:   time.Date(2022, 7, 4, 10, 0, 0, 0, time.UTC),
				},
				{
					StartTime: time.Date(2022, 7, 11, 9, 0, 0, 0, time.UTC),
					EndTime:   time.Date(2022, 7, 11, 10, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name:      "until date greater than last day within six day",
			startTime: time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
			endTime:   time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC),
			untilDate: time.Date(2022, 7, 10, 10, 0, 0, 0, time.UTC),
			expectedRecurSet: []RecurringSet{
				{
					StartTime: time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
					EndTime:   time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC),
				},
				{
					StartTime: time.Date(2022, 7, 4, 9, 0, 0, 0, time.UTC),
					EndTime:   time.Date(2022, 7, 4, 10, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name:      "until date equal last day",
			startTime: time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
			endTime:   time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC),
			untilDate: time.Date(2022, 7, 4, 12, 0, 0, 0, time.UTC),
			expectedRecurSet: []RecurringSet{
				{
					StartTime: time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
					EndTime:   time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC),
				},
				{
					StartTime: time.Date(2022, 7, 4, 9, 0, 0, 0, time.UTC),
					EndTime:   time.Date(2022, 7, 4, 10, 0, 0, 0, time.UTC),
				},
			},
		},
	}
	for _, tc := range testCases {
		rule, err := NewRecurrenceRule(Option{
			Freq:      WEEKLY,
			StartTime: tc.startTime,
			EndTime:   tc.endTime,
			UntilDate: tc.untilDate,
		})
		require.NoError(t, err)
		recurSet := rule.All()
		require.True(t, equalRecurSet(recurSet, tc.expectedRecurSet))
	}
}

func TestFollowingLessonID_GetNoLockedLessons(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name              string
		followingLessonID *FollowingLessonID
		lockedLessons     []string
		noLockedLesson    []string
		isBadCase         bool
	}{
		{
			name:              "get lesson no locked",
			followingLessonID: &FollowingLessonID{"lesson-1", "lesson-2", "lesson-3"},
			lockedLessons:     []string{"lesson-2"},
			noLockedLesson:    []string{"lesson-1", "lesson-3"},
		},
		{
			name:              "get lesson no locked with locked lesson does not include in followingLessonID",
			followingLessonID: &FollowingLessonID{"lesson-1", "lesson-2", "lesson-3"},
			lockedLessons:     []string{"lesson-2", "lesson-4"},
			noLockedLesson:    []string{"lesson-1", "lesson-3"},
		},
		{
			name:              "bad case",
			followingLessonID: &FollowingLessonID{"lesson-1", "lesson-2", "lesson-3"},
			lockedLessons:     []string{"lesson-2", "lesson-4"},
			noLockedLesson:    []string{"lesson-1"},
			isBadCase:         true,
		},
	}
	for _, tc := range testCases {
		result := tc.followingLessonID.GetNoLockedLessons(tc.lockedLessons)
		if tc.isBadCase {
			require.NotEqual(t, tc.noLockedLesson, result)
		} else {
			require.Equal(t, tc.noLockedLesson, result)
		}

	}
}

func equalRecurSet(got, expected []RecurringSet) bool {
	if len(got) != len(expected) {
		return false
	}
	for index, value := range got {
		if value.StartTime != expected[index].StartTime {
			return false
		}
		if value.EndTime != expected[index].EndTime {
			return false
		}
	}
	return true
}
