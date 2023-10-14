package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLessonLearner_NewLessonLearner(t *testing.T) {
	t.Parallel()

	learnerId := "learner-1"
	courseId := "course-1"
	locationId := "location-1"
	attendStatus := StudentAttendStatusAttend
	attendNotice := NoticeEmpty
	attendReason := ReasonEmpty
	gradeId := "grade-1"
	courseName := "course_name"
	t.Run("init lesson learner", func(t *testing.T) {
		ll := NewLessonLearner(
			learnerId,
			courseId,
			locationId,
			string(attendStatus),
			string(attendNotice),
			string(attendReason),
			"",
		)
		ll.AddGrade(gradeId)
		ll.AddCourseName(courseName)
		ll.AddGrade("grade-1")
		ll.AddReallocate("lesson-2")
		require.Equal(t, gradeId, ll.Grade)
		require.Equal(t, courseName, ll.CourseName)
		require.Equal(t, learnerId, ll.LearnerID)
		require.Equal(t, courseId, ll.CourseID)
		require.Equal(t, locationId, ll.LocationID)
		require.Equal(t, attendStatus, ll.AttendStatus)
		require.Equal(t, &Reallocate{OriginalLessonID: "lesson-2"}, ll.Reallocate)
		attendanceInfoMap := map[string]*LessonLearner{
			"learner-1": {
				AttendStatus:     StudentAttendStatusInformedLate,
				AttendanceNotice: NoContact,
				AttendanceReason: FamilyReason,
				AttendanceNote:   "note",
			},
		}
		ll.EmptyAttendanceInfo()
		checkAttendanceInfo(t, ll, StudentAttendStatusEmpty, NoticeEmpty, ReasonEmpty, "")
		ll.SetAttendanceInfoIfPresentOnMap(attendanceInfoMap)
		checkAttendanceInfo(t, ll, StudentAttendStatusInformedLate, NoContact, FamilyReason, "note")
		ll.LearnerID = "learner-3"
		ll.SetAttendanceInfoIfPresentOnMap(attendanceInfoMap)
		checkAttendanceInfo(t, ll, StudentAttendStatusEmpty, NoticeEmpty, ReasonEmpty, "")
	})
}

func TestLessonLearner_IsReallocateStatus(t *testing.T) {
	t.Run("is reallocate student", func(t *testing.T) {
		learner := &LessonLearner{
			LearnerID:    "learner-1",
			CourseID:     "course-1",
			AttendStatus: StudentAttendStatusReallocate,
		}
		isReallocate := learner.IsReallocateStatus()
		require.True(t, isReallocate)
	})
}

func checkAttendanceInfo(t *testing.T, l *LessonLearner,
	status StudentAttendStatus, notice StudentAttendanceNotice,
	reason StudentAttendanceReason, note string) {
	require.Equal(t, status, l.AttendStatus)
	require.Equal(t, notice, l.AttendanceNotice)
	require.Equal(t, reason, l.AttendanceReason)
	require.Equal(t, note, l.AttendanceNote)
}

func TestLessonLearners_GetLearnerIDs(t *testing.T) {
	t.Run("list learners", func(t *testing.T) {
		learners := LessonLearners{
			{
				LearnerID:    "learner-1",
				AttendStatus: StudentAttendStatusReallocate,
			},
			{
				LearnerID:    "learner-2",
				AttendStatus: StudentAttendStatusAttend,
			},
			{
				LearnerID:    "learner-3",
				AttendStatus: StudentAttendStatusReallocate,
			},
		}
		learnerIds := learners.GetLearnerIDs()
		require.Equal(t, []string{"learner-1", "learner-2", "learner-3"}, learnerIds)
	})
}

func TestLessonLearners_GetCourseIDsOfLessonLearners(t *testing.T) {
	t.Run("list distinct courses", func(t *testing.T) {
		learners := LessonLearners{
			{
				LearnerID: "learner-1",
				CourseID:  "course-1",
			},
			{
				LearnerID: "learner-2",
				CourseID:  "course-2",
			},
			{
				LearnerID: "learner-3",
				CourseID:  "course-2",
			},
		}
		courseIds := learners.GetCourseIDsOfLessonLearners()
		require.ElementsMatch(t, []string{"course-1", "course-2"}, courseIds)
	})
}

func TestLessonLearners_GetReallocateStudentRemoved(t *testing.T) {
	t.Parallel()
	learners := LessonLearners{
		{
			LearnerID: "learner-1",
		},
		{
			LearnerID: "learner-2",
		},
	}
	t.Run("found students are removed", func(t *testing.T) {
		learnerRemoved := &LessonLearner{
			LearnerID:    "learner-3",
			CourseID:     "course-1",
			AttendStatus: StudentAttendStatusAbsent,
		}
		oldLearner := append(learners, learnerRemoved)
		studentIds := learners.GetReallocateStudentRemoved(oldLearner)
		require.Equal(t, []string{learnerRemoved.LearnerID}, studentIds)
	})
	t.Run("found no students are removed", func(t *testing.T) {
		studentIds := learners.GetReallocateStudentRemoved(learners)
		require.Empty(t, studentIds)
	})
}

func TestLessonLearners_GetStudentNoPendingReallocate(t *testing.T) {
	t.Parallel()
	oldLearners := LessonLearners{
		{
			LearnerID:    "learner-1",
			AttendStatus: StudentAttendStatusReallocate,
		},
		{
			LearnerID:    "learner-2",
			AttendStatus: StudentAttendStatusAttend,
		},
		{
			LearnerID:    "learner-3",
			AttendStatus: StudentAttendStatusReallocate,
		},
		{
			LearnerID:    "learner-4",
			AttendStatus: StudentAttendStatusReallocate,
		},
	}
	t.Run("found students that no pending reallocation", func(t *testing.T) {
		newLearners := LessonLearners{
			{
				LearnerID:    "learner-1",
				AttendStatus: StudentAttendStatusAbsent,
			},
			{
				LearnerID:    "learner-2",
				AttendStatus: StudentAttendStatusAttend,
			},
			{
				LearnerID:    "learner-4",
				AttendStatus: StudentAttendStatusEmpty,
			},
		}
		studentIds := newLearners.GetStudentNoPendingReallocate(oldLearners)
		require.Equal(t, []string{"learner-1", "learner-4"}, studentIds)
	})

	t.Run("found no students that no pending reallocation", func(t *testing.T) {
		newLearners := LessonLearners{
			{
				LearnerID:    "learner-1",
				AttendStatus: StudentAttendStatusReallocate,
			},
			{
				LearnerID:    "learner-3",
				AttendStatus: StudentAttendStatusReallocate,
			},
		}
		studentIds := newLearners.GetStudentNoPendingReallocate(oldLearners)
		require.Empty(t, studentIds)
	})
}

func TestLessonLearners_GetStudentReallocatedDiffLocation(t *testing.T) {
	t.Parallel()

	t.Run("found no students that diff lesson location", func(t *testing.T) {
		newLearners := LessonLearners{
			{
				LearnerID:    "learner-1",
				AttendStatus: StudentAttendStatusReallocate,
				LocationID:   "center-id-1",
			},
			{
				LearnerID:    "learner-3",
				AttendStatus: StudentAttendStatusReallocate,
				LocationID:   "center-id-1",
			},
		}
		studentIds := newLearners.GetStudentReallocatedDiffLocation("center-id-1")
		require.Empty(t, studentIds)
	})
	t.Run("found students that reallocated to diff location", func(t *testing.T) {
		newLearners := LessonLearners{
			{
				LearnerID:    "learner-1",
				AttendStatus: StudentAttendStatusReallocate,
				LocationID:   "center-id-1",
				Reallocate: &Reallocate{
					OriginalLessonID: "ls-1",
				},
			},
			{
				LearnerID:    "learner-3",
				AttendStatus: StudentAttendStatusReallocate,
				LocationID:   "center-id-1",
			},
		}
		studentIds := newLearners.GetStudentReallocatedDiffLocation("center-id-2")
		require.Equal(t, []string{"learner-1"}, studentIds)
	})
}

func TestLessonLearners_GetStudentReallocate(t *testing.T) {
	t.Parallel()
	learners := LessonLearners{
		{
			LearnerID:    "learner-1",
			AttendStatus: StudentAttendStatusAttend,
		},
		{
			LearnerID:    "learner-2",
			AttendStatus: StudentAttendStatusAbsent,
		},
		{
			LearnerID:    "learner-2",
			AttendStatus: StudentAttendStatusReallocate,
		},
	}
	t.Run("have student who change attendance status to reallocate", func(t *testing.T) {
		attendanceStatusMap := map[string]StudentAttendStatus{
			"learner-1": StudentAttendStatusReallocate,
			"learner-2": StudentAttendStatusReallocate,
			"learner-3": StudentAttendStatusReallocate,
		}
		studentIds := learners.GetStudentReallocate(attendanceStatusMap)
		require.Equal(t, []string{"learner-1", "learner-2"}, studentIds)
	})
	t.Run("have no student who change attendance status to reallocate", func(t *testing.T) {
		attendanceStatusMap := map[string]StudentAttendStatus{
			"learner-1": StudentAttendStatusAbsent,
			"learner-2": StudentAttendStatusAttend,
			"learner-3": StudentAttendStatusReallocate,
		}
		studentIds := learners.GetStudentReallocate(attendanceStatusMap)
		require.Empty(t, studentIds)
	})
}

func TestLessonLearners_GetStudentUnReallocate(t *testing.T) {
	t.Parallel()
	learners := LessonLearners{
		{
			LearnerID:    "learner-1",
			AttendStatus: StudentAttendStatusReallocate,
		},
		{
			LearnerID:    "learner-2",
			AttendStatus: StudentAttendStatusReallocate,
		},
		{
			LearnerID:    "learner-3",
			AttendStatus: StudentAttendStatusReallocate,
		},
	}
	t.Run("have student who change attendance status to diff reallocate", func(t *testing.T) {
		attendanceStatusMap := map[string]StudentAttendStatus{
			"learner-1": StudentAttendStatusAbsent,
			"learner-2": StudentAttendStatusAttend,
			"learner-3": StudentAttendStatusReallocate,
		}
		studentIds := learners.GetStudentUnReallocate(attendanceStatusMap)
		require.Equal(t, []string{"learner-1", "learner-2"}, studentIds)
	})
	t.Run("have no student who change attendance status to diff reallocate", func(t *testing.T) {
		attendanceStatusMap := map[string]StudentAttendStatus{
			"learner-1": StudentAttendStatusReallocate,
			"learner-2": StudentAttendStatusReallocate,
			"learner-3": StudentAttendStatusReallocate,
		}
		studentIds := learners.GetStudentUnReallocate(attendanceStatusMap)
		require.Empty(t, studentIds)
	})
}

func TestLessonLearners_GroupByLearnerID(t *testing.T) {
	t.Parallel()
	learners := LessonLearners{
		{
			LearnerID:    "learner-1",
			AttendStatus: StudentAttendStatusAttend,
			CourseID:     "course-1",
		},
		{
			LearnerID:    "learner-2",
			AttendStatus: StudentAttendStatusReallocate,
			CourseID:     "course-1",
		},
	}
	t.Run("happy case", func(t *testing.T) {
		learnerMap := learners.GroupByLearnerID()
		require.Equal(t, map[string]*LessonLearner{
			"learner-1": {
				LearnerID:    "learner-1",
				AttendStatus: StudentAttendStatusAttend,
				CourseID:     "course-1",
			},
			"learner-2": {
				LearnerID:    "learner-2",
				AttendStatus: StudentAttendStatusReallocate,
				CourseID:     "course-1",
			},
		}, learnerMap)
	})
}

func TestLessonLearners_GetLearnerByID(t *testing.T) {
	t.Parallel()
	learners := LessonLearners{
		{
			LearnerID:    "learner-1",
			AttendStatus: StudentAttendStatusAttend,
			CourseID:     "course-1",
		},
		{
			LearnerID:    "learner-2",
			AttendStatus: StudentAttendStatusReallocate,
			CourseID:     "course-1",
		},
	}
	t.Run("happy case", func(t *testing.T) {
		learner := learners.GetLearnerByID("learner-1")
		require.Equal(t, &LessonLearner{
			LearnerID:    "learner-1",
			AttendStatus: StudentAttendStatusAttend,
			CourseID:     "course-1",
		}, learner)
	})
	t.Run("no found", func(t *testing.T) {
		learner := learners.GetLearnerByID("learner-3")
		require.Nil(t, learner)
	})
}

func TestLessonLearners_GetStudentRemoved(t *testing.T) {
	t.Parallel()
	newLearners := LessonLearners{
		{
			LearnerID: "learner-1",
		},
	}
	t.Run("have two student removed", func(t *testing.T) {
		oldLearners := LessonLearners{
			{
				LearnerID: "learner-1",
			},
			{
				LearnerID: "learner-2",
			},
			{
				LearnerID: "learner-3",
			},
		}
		studentIds := newLearners.GetStudentRemoved(oldLearners)
		require.Equal(t, []string{"learner-2", "learner-3"}, studentIds)
	})
	t.Run("have no student removed", func(t *testing.T) {
		oldLearners := LessonLearners{
			{
				LearnerID: "learner-1",
			},
		}
		studentIds := newLearners.GetStudentRemoved(oldLearners)
		require.Empty(t, studentIds)
	})
}
