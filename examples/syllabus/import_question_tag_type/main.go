package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/manabie-com/backend/examples/syllabus"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"google.golang.org/grpc"
)

const (
	// Run below command if test with local environment
	// kubectl port-forward $(kubectl get pods -l app.kubernetes.io/name=eureka -n backend -o=name) 15090:15090 5550:5550 -n backend
	EUREKA_LOCAL_API_URL = "127.0.0.1:5550"
	STAG_API_URL         = "web-api.staging-green.manabie.io:443"
	UAT_API_URL          = "api.uat.manabie.io:443"
	PROD_TOKYO_API_URL   = "https://web-api.prod.tokyo.manabie.io:31400"
)

func main() {
	// GET token manually :)
	token := `eyJhbGciOiJSUzI1NiIsImtpZCI6ImE1OGFmM2EyZTBkMzdiMDk2NDViNjZmOTRmYWMzM2I5ODZkMmJjMDkifQ.eyJpc3MiOiJtYW5hYmllIiwic3ViIjoiMDFHR0NGODJHUzJOQkowV0pUTldTUDZCU0ciLCJhdWQiOiJtYW5hYmllLWxvY2FsIiwiZXhwIjoxNjY2ODY5NTY3LCJpYXQiOjE2NjY4NjU5NjIsImp0aSI6IjAxR0dDRjg3V0c2OTU2RFJZTU5FMDFOVlJYIiwiaHR0cHM6Ly9oYXN1cmEuaW8vand0L2NsYWltcyI6eyJ4LWhhc3VyYS1hbGxvd2VkLXJvbGVzIjpbIlVTRVJfR1JPVVBfU0NIT09MX0FETUlOIl0sIngtaGFzdXJhLWRlZmF1bHQtcm9sZSI6IlVTRVJfR1JPVVBfU0NIT09MX0FETUlOIiwieC1oYXN1cmEtdXNlci1pZCI6IjAxR0dDRjgyR1MyTkJKMFdKVE5XU1A2QlNHIiwieC1oYXN1cmEtc2Nob29sLWlkcyI6InstMjE0NzQ4MzY0OH0iLCJ4LWhhc3VyYS11c2VyLWdyb3VwIjoiVVNFUl9HUk9VUF9TQ0hPT0xfQURNSU4iLCJ4LWhhc3VyYS1yZXNvdXJjZS1wYXRoIjoiLTIxNDc0ODM2NDgifSwibWFuYWJpZSI6eyJhbGxvd2VkX3JvbGVzIjpbIlVTRVJfR1JPVVBfU0NIT09MX0FETUlOIl0sImRlZmF1bHRfcm9sZSI6IlVTRVJfR1JPVVBfU0NIT09MX0FETUlOIiwidXNlcl9pZCI6IjAxR0dDRjgyR1MyTkJKMFdKVE5XU1A2QlNHIiwic2Nob29sX2lkcyI6WyItMjE0NzQ4MzY0OCJdLCJ1c2VyX2dyb3VwIjoiVVNFUl9HUk9VUF9TQ0hPT0xfQURNSU4iLCJyZXNvdXJjZV9wYXRoIjoiLTIxNDc0ODM2NDgifSwicmVzb3VyY2VfcGF0aCI6Ii0yMTQ3NDgzNjQ4IiwidXNlcl9ncm91cCI6IlVTRVJfR1JPVVBfU0NIT09MX0FETUlOIn0.h6WdYMMNQLdd0nVds5i5kAQG75b_67HZ0xvUWkW4_okaI0eLE5le149sTbvofD10VulVeSzd_uklp7yzjPmNSQ58Nl_DIkva2VxCD-E-sS-_-z0UBRIVpRKgHGtq-498345cJIRR4FGYQpNvpuPb5KBZStGGTRKZDYEYKES1Pvook7xxvJUgIiTjjf-o0XJwySIpOuW3LjsDqKH87s1ogM7cN6KeWRv0pUR4HdR9Mej_ZtvZq3kfMBTEKCBAQhrgsLHAe6KRGHLRjceN6qAoyfYanBkaeLrpomJfbqTCkO7EGWRzaIspObxGiXJk9Ps5QxFvDo90U_O2mOebMP5o_g`
	//set "second arg" to "true" if run with local environment
	conn := syllabus.Connect(EUREKA_LOCAL_API_URL, false)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*100)
	defer cancel()

	ctx = syllabus.AuthorizedContext(ctx, token)
	err := importQuestionTagTypes(ctx, conn)
	if err != nil {
		log.Fatal(err.Error())
	}
}

/*format "data.csv" file similar to
id,name
id-1,name-1
id-2,name-2
...
*/
func importQuestionTagTypes(ctx context.Context, conn *grpc.ClientConn) error {
	content, err := os.ReadFile("examples/syllabus/import_question_tag_type/data.csv")
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("please add rows to data.csv file in this folder")
	}
	fmt.Println("-------------------IMPORTING!-------------------")
	_, err = sspb.NewQuestionTagTypeClient(conn).ImportQuestionTagTypes(ctx, &sspb.ImportQuestionTagTypesRequest{
		Payload: content,
	})
	if err != nil {
		return fmt.Errorf("ImportQuestionTagTypes: %v", err)
	}
	fmt.Println("-------------------SUCCESS!-------------------")
	return nil
}
