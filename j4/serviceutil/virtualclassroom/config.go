package virtualclassroom

type Config struct {
	LessonInfo LessonInfoConfig `yaml:"lesson_info"`
	SchoolID   string           `yaml:"school_id"`
	AdminID    string           `yaml:"admin_id"`
}

type LessonInfoConfig struct {
	CourseID   string `yaml:"course_id"`
	LocationID string `yaml:"location_id"`
}
