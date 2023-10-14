package entity

import (
	"time"

	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"
)

type School struct {
	ID              int32
	Admins          []*Admin
	DefaultLocation string
}
type Teacher struct {
	User     *User
	SchoolID int64
}

type Admin struct {
	ID        string
	Name      string
	Email     string
	Token     string
	Password  string
	SchoolIds []int64
	UserGroup string
}

func (a *Admin) ToUser() *User {
	return &User{
		ID:       a.ID,
		Group:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
		Name:     a.Name,
		Email:    a.Email,
		Token:    a.Token,
		Password: a.Password,
	}
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

func (u *User) FromParentPB(par *upb.Parent) {
	u.ID = par.UserProfile.UserId
	u.Group = cpb.UserGroup_USER_GROUP_PARENT.String()
	u.Name = par.UserProfile.Name
	u.Email = par.UserProfile.Email
	u.Phone = par.UserProfile.PhoneNumber
}

func (u *User) FromStudentPB(stu *upb.Student) {
	u.ID = stu.UserProfile.UserId
	u.Group = cpb.UserGroup_USER_GROUP_STUDENT.String()
	u.Name = stu.UserProfile.Name
	u.Email = stu.UserProfile.Email
	u.Phone = stu.UserProfile.PhoneNumber
}

type Student struct {
	User
	Courses  []*Course
	Grade    *Grade
	Parents  []*User
	SchoolID int32
}

type Grade struct {
	ID int32
}

type Course struct {
	ID        string
	Name      string
	GradeID   int32
	GradeName string
	SchoolID  int32
}

type GradeFilter struct {
	Type   cpb.NotificationTargetGroupSelect
	Grades []int32
}

type CourseFilter struct {
	Type    cpb.NotificationTargetGroupSelect
	Courses []string
}

type LocationFilter struct {
	Type      cpb.NotificationTargetGroupSelect
	Locations []string
}

type ClassFilter struct {
	Type    cpb.NotificationTargetGroupSelect
	Classes []string
}

type Notification struct {
	ID                  string
	SchoolID            int32
	Title               string
	Content             string
	HTMLContent         string
	ScheduledAt         time.Time
	FilterByGrade       GradeFilter
	FilterByCourse      CourseFilter
	FilterByLocation    LocationFilter
	FilterByClass       ClassFilter
	IndividualReceivers []string
	MediaIds            []string
	ReceiverGroup       []cpb.UserGroup
	Status              cpb.NotificationStatus
	Data                map[string]interface{}
}

type NotificationCommonError struct {
	UpsertError  error
	DiscardError error
	SentError    error
}

type State struct {
	School           *School
	SystemAdmin      *Admin
	Students         []*Student
	Notify           *Notification
	NotifyErr        NotificationCommonError
	Client           *Client
	NatsNotification *ypb.NatsCreateNotificationRequest
}

type Client struct {
	ClientID string
}

type StateKey struct{}
