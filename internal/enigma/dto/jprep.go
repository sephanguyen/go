package dto

type Action string

const (
	ActionKindUpserted Action = "upserted"
	ActionKindDeleted  Action = "deleted"

	CourseIDJuku  = 1
	CourseIDKid   = 2
	CourseIDAPlus = 3
)

// MasterRegistration
type (
	Course struct {
		ActionKind Action `json:"action_kind"`
		// JPREP courseID
		CourseID int `json:"m_course_name_id"`
		// Manabie courseName
		CourseName string `json:"course_name"`
		// to identify Kids course.
		// We will need to reference m_course_student_div to find Kids course type.
		CourseStudentDivID int `json:"m_course_student_div_id"`
	}

	Class struct {
		ActionKind Action `json:"action_kind"`
		// manabie className
		ClassName string `json:"m_class_name"`
		// manabie classID
		ClassID int `json:"m_course_id"`
		// courseID
		CourseID int `json:"m_course_name_id"`
		// startDate of package
		StartDate string `json:"startdate"`
		// endDate of package
		EndDate        string `json:"enddate"`
		AcademicYearID int    `json:"m_academic_year_id"`
	}

	Lesson struct {
		ActionKind    Action `json:"action_kind"`
		LessonID      int    `json:"m_lesson_id"`
		LessonType    string `json:"lesson_type"`
		CourseID      int    `json:"m_course_name_id"`
		StartDatetime int    `json:"start_datetime"`
		EndDatetime   int    `json:"end_datetime"`
		ClassName     string `json:"class_name"`
		Week          string `json:"week"`
	}

	AcademicYear struct {
		ActionKind     Action `json:"action_kind"`
		AcademicYearID int    `json:"m_academic_year_id"`
		Name           string `json:"year_name"`
		StartYearDate  int64  `json:"start_year_date"`
		EndYearDate    int64  `json:"end_year_date"`
	}

	MasterRegistrationRequest struct {
		Timestamp int `json:"timestamp"`
		Payload   struct {
			Courses       []Course       `json:"m_course_name"`
			Classes       []Class        `json:"m_regular_course"`
			Lessons       []Lesson       `json:"m_lesson"`
			AcademicYears []AcademicYear `json:"m_academic_year"`
		} `json:"payload"`
	}

	MasterRegistrationResponse struct{}
)

// UserRegistration
type (
	Student struct {
		ActionKind  Action `json:"action_kind"`
		StudentID   string `json:"student_id"`
		StudentDivs []struct {
			MStudentDivID int `json:"m_student_div_id"`
		} `json:"student_divs"`
		LastName       string `json:"last_name"`
		GivenName      string `json:"given_name"`
		Regularcourses []struct {
			// class id
			ClassID int `json:"m_course_id"`
			// The start date of a package, which the student purchased to have the access into the class ID (m_course_id)
			Startdate string `json:"startdate"`
			// The end date of a package, which the student purchased to have the access into the class ID (m_course_id)
			Enddate string `json:"enddate"`
		} `json:"regularcourses"`
	}

	Staff struct {
		ActionKind Action `json:"action_kind"`
		StaffID    string `json:"staff_id"`
		Name       string `json:"name"`
	}

	UserRegistrationRequest struct {
		Timestamp int `json:"timestamp"`
		Payload   struct {
			Students []Student `json:"m_student"`
			Staffs   []Staff   `json:"m_staff"`
		} `json:"payload"`
	}
)

// LiveLesson
type (
	StudentLesson struct {
		ActionKind Action `json:"action_kind"`
		StudentID  string `json:"student_id"`
		LessonIDs  []int  `json:"m_lesson_ids"`
	}

	SyncUserCourseRequest struct {
		Timestamp int `json:"timestamp"`
		Payload   struct {
			StudentLessons []StudentLesson `json:"student_lesson"`
		} `json:"payload"`
	}

	SyncUserCourseResponse struct{}
)

type (
	PartnerLogRequest struct {
		Timestamp int `json:"timestamp"`
		Payload   struct {
			Signature string `json:"signature"`
		}
	}
	PartnerLogResponse struct {
		PartnerSyncDataLogSplitID string `json:"partner_sync_data_log_split_id"`
		Status                    string `json:"status"`
		UpdatedAt                 int64  `json:"updated_at"`
	}

	PartnerLogRequestByDate struct {
		Timestamp int `json:"timestamp"`
		Payload   struct {
			FromDate string `json:"from_date"`
			ToDate   string `json:"to_date"`
		}
	}
)
