package usermgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"google.golang.org/grpc"
)

type Student struct {
	UserID string
}

type GrpClient interface {
	ImportStudent(ctx context.Context, in *upb.ImportStudentRequest, opts ...grpc.CallOption) (*upb.ImportStudentResponse, error)
}

func getOneLocation(ctx context.Context, db database.Ext, schoolID string) (string, error) {
	var ret string
	err := db.QueryRow(ctx, "select partner_internal_id from locations where resource_path=$1", database.Text(schoolID)).Scan(&ret)
	return ret, err
}

// CreateStudents ctx must already has grpc token
func CreateStudents(ctx context.Context, num int, schoolID string, bobDB database.Ext, userSvc GrpClient) ([]Student, error) {
	location, err := getOneLocation(ctx, bobDB, schoolID)
	if err != nil {
		return nil, err
	}
	randomID := idutil.ULIDNow()
	payload := "first_name,last_name,email,enrollment_status,grade,phone_number,birthday,gender,location\n"
	emails := []string{}
	for i := 0; i < num; i++ {
		email := fmt.Sprintf("student-%s-%d@example.com", randomID, i)
		payload += fmt.Sprintf("%s,%s,%s,1,0,,,,%s\n", email, email, email, location)
		emails = append(emails, email)
	}
	res, err := userSvc.ImportStudent(ctx, &upb.ImportStudentRequest{Payload: []byte(payload)})

	if err != nil {
		return nil, err
	}
	if len(res.Errors) > 0 {
		return nil, fmt.Errorf("%v", res.Errors)
	}
	rows, err := bobDB.Query(ctx, "select user_id from users where email=any($1)", emails)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	students := []Student{}
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			panic(err)
		}
		students = append(students, Student{UserID: id})
	}

	return students, nil
}
