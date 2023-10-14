package eureka

import (
	"fmt"

	bpb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func generateValidQuestion(id string) bpb.Question {
	return bpb.Question{Id: id, MasterQuestionId: "", Country: bpb.COUNTRY_VN, Question: fmt.Sprintf("valid question %v", id), Answers: []string{fmt.Sprintf("valid anwser %d", 0), fmt.Sprintf("valid anwser %d", 1), fmt.Sprintf("valid anser %d", 2), fmt.Sprintf("valid anser %d", 3)}, Explanation: "Correct explanation for question", DifficultyLevel: 2, UpdatedAt: nil, CreatedAt: nil, QuestionsTagLo: []string{"", "", ""}, ExplanationWrongAnswer: []string{"explanation for answer 2", "explanation for answer 3", "explanation for answer 4"}}
}
