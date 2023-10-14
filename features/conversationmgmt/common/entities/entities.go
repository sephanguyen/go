package entities

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
	FirstName   string
	LastName    string
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

type Conversation struct {
	ID             string
	Name           string
	MemberIDs      []string
	OptionalConfig []byte
}
