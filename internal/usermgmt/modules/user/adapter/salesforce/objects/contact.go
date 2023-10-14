package objects

import (
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type ContactRepo struct {
}

type ContactAttribute struct {
	ID         field.String `json:"Id"`
	FirstName  field.String `json:"FirstName"`
	LastName   field.String `json:"LastName"`
	Email      field.String `json:"Email"`
	HomePhone  field.String `json:"HomePhone"`
	Phone      field.String `json:"Phone"`
	OtherPhone field.String `json:"OtherPhone"`
	Department field.String `json:"Department"`
	BirthDate  field.String `json:"Birthdate"`
	AccountID  field.String `json:"AccountId"`
}

type Contact struct {
	ContactAttribute
	CreatedDate    field.Time `json:"CreatedDate"`
	SystemModStamp field.Time `json:"SystemModStamp"`
}

func (c *Contact) Attrs() []string {
	return []string{"Id", "FirstName", "LastName", "Email", "HomePhone", "Phone", "OtherPhone", "Department", "Birthdate", "AccountId"}
}

func (c *Contact) Object() string {
	return "Contact"
}

func (c *Contact) Objects() string {
	return "Contacts"
}
