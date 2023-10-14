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
	// GET token manually and put it here
	token := ``
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
id,name,question_tag_type_id
id-1,name-1,question_tag_type_id-1
id-2,name-2,question_tag_type_id-2
...
*/
func importQuestionTagTypes(ctx context.Context, conn *grpc.ClientConn) error {
	content, err := os.ReadFile("examples/syllabus/import_question_tag/data.csv")
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("please add rows to data.csv file in this folder")
	}
	fmt.Println("-------------------IMPORTING!-------------------")
	_, err = sspb.NewQuestionTagClient(conn).ImportQuestionTag(ctx, &sspb.ImportQuestionTagRequest{
		Payload: content,
	})
	if err != nil {
		return fmt.Errorf("ImportQuestionTag: %v", err)
	}
	fmt.Println("-------------------SUCCESS!-------------------")
	return nil
}
