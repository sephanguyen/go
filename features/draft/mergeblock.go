package draft

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/manabie-com/backend/internal/draft/repositories"
	"github.com/manabie-com/backend/internal/draft/service"
)

type mergeBlockSuite struct {
	prNumber int
	repo     string
	owner    string
}

func isStringBlock(str string) bool {
	return str == "blocked" || str == "block"
}

func (s *suite) aRepoOfOwner(ctx context.Context, repo, owner string) (context.Context, error) {
	s.owner = owner
	s.repo = repo
	_, err := s.DB.Exec(ctx, `INSERT INTO github_repo_state(org,repo,is_blocked) 
	VALUES ($1,$2,'false') ON CONFLICT ON CONSTRAINT github_repo_pkey DO NOTHING`, owner, repo)
	return ctx, err
}

// read check status of pr
func (s *suite) repoHasMergeStatusIs(ctx context.Context, blockstatus string) (context.Context, error) {
	repo := &repositories.GithubMergeStatusRepo{}
	isblocked, err := repo.GetRepoMergeStatus(ctx, s.DB, s.owner, s.repo)
	if err != nil {
		return ctx, err
	}
	if isblocked != isStringBlock(blockstatus) {
		return ctx, fmt.Errorf("want repo status %v, has %v", isStringBlock(blockstatus), isblocked)
	}
	return ctx, nil
}

var draftAddr = "http://draft.backend.svc.cluster.local:6080"

// call set status
func (s *suite) workflowIsCalledToRepo(ctx context.Context, blockstatus string) (context.Context, error) {
	payload := service.SetStatusPayload{
		Repo:       s.repo,
		Owner:      s.owner,
		BlockMerge: isStringBlock(blockstatus),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return ctx, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/draft-http/v1/status", draftAddr), bytes.NewReader(raw))
	if err != nil {
		return ctx, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", s.Cfg.GitHubWebhookSecret)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return ctx, err
	}
	if res.StatusCode > 299 {
		return ctx, fmt.Errorf("POST /status returns code %d", res.StatusCode)
	}
	return ctx, nil
}

// func computeSignature(payload []byte, secret []byte) string {
// 	key := hmac.New(sha256.New, []byte(secret))
// 	key.Write([]byte(string(payload)))
// 	return "sha256=" + hex.EncodeToString(key.Sum(nil))
// }
