package syllabus

type Config struct {
	CourseID     string `yaml:"course_id"`
	StudyPlanID  string `yaml:"study_plan_id"`
	ResourcePath string `yaml:"resource_path"`
	UserID       string `yaml:"user_id"`
}
