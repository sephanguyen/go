package entities

import (
	"time"

	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
)

type Organization struct {
	ID                  int32
	Staffs              []*Staff
	DefaultLocation     *Location
	DescendantLocations []*Location
}

type Staff struct {
	ID                 string
	Name               string
	Email              string
	Token              string
	OrganizationIDs    []int32
	GrandtedRoles      []string
	GrantedLocationIDs []string

	//Deprecated: Old logic
	UserGroup string
}

type User struct {
	ID          string
	Group       string
	Name        string
	Email       string
	Password    string
	Phone       string
	Token       string
	DeviceToken string
}

type Student struct {
	User
	Courses        []*Course
	GradeMaster    *GradeMaster
	Parents        []*User
	OrganizationID int32
	Packages       []*StudentPackage
}

type GradeMaster struct {
	ID                string
	Name              string
	PartnerInternalID string
}

type Course struct {
	ID             string
	Name           string
	OrganizationID int32
	LocationIDs    []string
	Classes        []*Class
}

type GradeFilter struct {
	Type   cpb.NotificationTargetGroupSelect
	Grades []int32
}

type CourseFilter struct {
	Type    cpb.NotificationTargetGroupSelect
	Courses []string
}

type Notification struct {
	ID                  string
	OrganizationID      int32
	Title               string
	Content             string
	HTMLContent         string
	ScheduledAt         time.Time
	FilterByGrade       GradeFilter
	FilterByCourse      CourseFilter
	IndividualReceivers []string
	MediaIds            []string
	ReceiverGroup       []cpb.UserGroup
	Status              cpb.NotificationStatus
	Data                map[string]interface{}
}

type Client struct {
	ClientID string
}

type Class struct {
	ID             string
	Name           string
	CourseID       string
	OrganizationID string
	LocationID     string
}

type StudentPackage struct {
	ID         string
	CourseID   string
	ClassID    string
	LocationID string
}

type Tag struct {
	ID         string
	Name       string
	IsArchived bool
}

type Location struct {
	ID               string
	Name             string
	AccessPath       string
	ParentLocationID string
	TypeLocation     string
	TypeLocationID   string
}

type SchoolLevel struct {
	ID   string
	Name string
}

type School struct {
	ID    string
	Name  string
	Level *SchoolLevel
}
