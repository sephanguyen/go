package objects

import (
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/salesforce"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type DomainStudentRepo struct {
}

type Student struct {
	entity.EmptyUser
	AccountAttribute
	Contacts salesforce.QueryResponse[ContactAttribute] `json:"Contacts"`
}

func (student *Student) UserID() field.String {
	return student.ID
}
func (student *Student) CurrentGrade() field.Int16 {
	return field.NewInt16(1)
}
func (student *Student) EnrollmentStatus() field.String {
	return field.NewNullString()
}
func (student *Student) StudentNote() field.String {
	return student.Description
}
func (student *Student) GradeID() field.String {
	return field.NewNullString()
}
func (student *Student) SchoolID() field.Int32 {
	return field.NewInt32(1)
}
func (student *Student) ContactPreference() field.String {
	return field.NewNullString()
}
func (student *Student) OrganizationID() field.String {
	return field.NewNullString()
}
func (student *Student) ExternalStudentID() field.String {
	return field.NewNullString()
}

func (repo *DomainStudentRepo) Get(client salesforce.SFClient, limit int, offset int) ([]entity.DomainStudent, error) {
	c := &Contact{}
	a := &Account{}
	query := fmt.Sprintf(
		"SELECT %s, (SELECT %s FROM %s) FROM %s LIMIT %d OFFSET %d",
		strings.Join(a.Attrs(), ","),
		strings.Join(c.Attrs(), ","),
		c.Objects(),
		a.Object(),
		limit,
		offset,
	)
	var results salesforce.QueryResponse[Student]
	err := client.Query(query, &results)

	if err != nil {
		return nil, err
	}
	students := []entity.DomainStudent{}

	for i := range results.Records {
		students = append(students, &results.Records[i])
	}
	return students, nil
}

func (repo *DomainStudentRepo) GetByID(client salesforce.SFClient, studentID string) (entity.DomainStudent, error) {
	c := &Contact{}
	a := &Account{}
	query := fmt.Sprintf(
		"SELECT %s, (SELECT %s FROM %s) FROM %s WHERE Id = '%s'",
		strings.Join(a.Attrs(), ","),
		strings.Join(c.Attrs(), ","),
		c.Objects(),
		a.Object(),
		studentID,
	)
	var results salesforce.QueryResponse[Student]
	err := client.Query(query, &results)

	if err != nil {
		return nil, err
	}
	if len(results.Records) > 0 {
		return &results.Records[0], nil
	}
	return &Student{}, nil
}

func (repo *DomainStudentRepo) Create(client salesforce.SFClient, student entity.DomainStudent) error {
	account := &Account{
		AccountAttribute: AccountAttribute{
			Name:        student.FullName(),
			Description: student.StudentNote(),
		},
	}
	accResp, err := client.Post(account.Object(), account)
	if err != nil {
		return err
	}
	contact := &Contact{
		ContactAttribute: ContactAttribute{
			FirstName: student.FirstName(),
			LastName:  student.LastName(),
			Email:     student.Email(),
			AccountID: field.NewString(accResp.ID),
		},
	}
	_, err = client.Post(contact.Object(), contact)
	if err != nil {
		return err
	}
	return nil
}
