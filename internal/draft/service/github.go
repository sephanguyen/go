package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/manabie-com/backend/internal/draft/configurations"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v41/github"
	"github.com/jackc/pgtype"
	"go.uber.org/zap"
)

type Event string

const (
	PullRequest  Event = "pull_request"
	Push         Event = "push"
	IssueComment Event = "issue_comment"
)

var Events = []Event{
	PullRequest,
	Push,
	IssueComment,
}

func (g *GithubCollectDataController) GithubEventFuncMap() map[string]func(context.Context, EventPayload) error {
	consumers := map[string]func(context.Context, EventPayload) error{
		string(PullRequest):  g.consumePullRequestEvent,
		string(IssueComment): g.consumeIssueCommentEvent,
	}
	return consumers
}

type GithubClient interface {
	CreateCommitStatus(ctx context.Context, owner, repo, commitSHA string, status *github.RepoStatus) error
	GetLastCommit(ctx context.Context, owner, repo string, prnumer int) (*github.RepositoryCommit, error)
}

type GithubCollectDataController struct {
	CFG *configurations.Config
	DB  database.Ext
	// CTX             context.Context
	Logger          *zap.SugaredLogger
	GithubClient    GithubClient
	MergeStatusRepo interface {
		GetRepoMergeStatus(ctx context.Context, db database.QueryExecer, org, repo string) (isBlocked bool, err error)
		SetRepoMergeStatus(ctx context.Context, db database.QueryExecer, org, repo string, isBlocked bool) error
	}

	GithubEventRepo interface {
		AddEventRawData(ctx context.Context, db database.QueryExecer, eventName string, data pgtype.JSONB) error
	}
}

func RegisterGithubCollectDataController(r *gin.Engine, g *GithubCollectDataController) {
	// to be removed
	r.POST("/api/v1/github/payload", g.GithubWebHookFuncs())

	newgr := r.Group("/draft-http/v1")

	// webhook from github
	newgr.POST("/github/payload", g.GithubWebHookFuncs())

	// non-webhook
	newgr.POST("/status", g.HandleStatus)
}

func (g *GithubCollectDataController) HandleStatus(c *gin.Context) {
	ctx := c.Request.Context()
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("io.ReadAll: %s", err)})
		return
	}
	g.Logger.Debugf("HandleStatus.payload: %s", string(payload))

	var ap SetStatusPayload
	if err := json.Unmarshal(payload, &ap); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("json.Unmarshal: %s", err)})
		return
	}
	// borrow this secret header from github
	if !verifyKey(c.GetHeader("X-Hub-Signature-256"), g.CFG.GitHubWebhookSecret) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err = g.SetAllPrStatus(ctx, ap)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("g.SetAllPrStatus: %s", err)})
	}
}

func verifyKey(cKey, sKey string) bool {
	return cKey == sKey
}
func VerifySignature(payload []byte, signature string, secret []byte) bool {
	key := hmac.New(sha256.New, secret)
	key.Write(payload)
	computedSignature := "sha256=" + hex.EncodeToString(key.Sum(nil))

	return computedSignature == signature
}

func (g *GithubCollectDataController) AddAllEventData(ctx context.Context, eventName string, payload []byte) error {
	var data interface{}
	err := json.Unmarshal(payload, &data)
	if err != nil {
		return err
	}

	err = g.GithubEventRepo.AddEventRawData(ctx, g.DB, eventName, database.JSONB(data))
	if err != nil {
		return err
	}
	return nil
}

func (g *GithubCollectDataController) GithubWebHookFuncs() func(c *gin.Context) {
	eventFuncMap := g.GithubEventFuncMap()
	return func(c *gin.Context) {
		payload, _ := io.ReadAll(c.Request.Body)
		if !VerifySignature(payload, c.GetHeader("X-Hub-Signature-256"), []byte(g.CFG.GitHubWebhookSecret)) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		event := c.GetHeader("X-GitHub-Event")

		err := g.AddAllEventData(c.Request.Context(), event, payload)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("g.AddAllEventData: %s", err)})
			return
		}

		for _, e := range Events {
			if string(e) == event {
				var p EventPayload
				if err := json.Unmarshal(payload, &p); err != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("json.Unmarshal: %s", err)})
					return
				}
				eventFunc, ok := eventFuncMap[string(e)]
				if !ok {
					continue
				}
				if err := eventFunc(c.Request.Context(), p); err != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("eventFunc: %s", err)})
					return
				}
				c.AbortWithStatus(http.StatusNoContent)
				g.Logger.Debugf("consumed event: %s", event)
				return
			}
		}
	}
}

func getStatusAndDescription(isBlocked bool) (string, string) {
	status := "success"
	description := "Merge allowed"
	if isBlocked {
		status = "failure"
		description = "Merge blocked"
	}
	return status, description
}

func (g *GithubCollectDataController) consumeIssueCommentEvent(ctx context.Context, payload EventPayload) error {
	owner := payload.Repository.Owner.Login
	repoName := payload.Repository.Name

	if payload.Comment.Body != "/draft/merge-check" {
		return nil
	}
	commit, err := g.GithubClient.GetLastCommit(ctx, owner, repoName, payload.Issue.Number)
	if err != nil {
		return fmt.Errorf("ghClient.GetLastCommit %w", err)
	}
	return g.doMergeCheck(ctx, owner, repoName, *commit.SHA)
}

func (g *GithubCollectDataController) doMergeCheck(ctx context.Context, owner, repoName string, sha string) error {
	isBlocked, err := g.MergeStatusRepo.GetRepoMergeStatus(ctx, g.DB, owner, repoName)
	if err != nil {
		return fmt.Errorf("GetMergeStatus %w", err)
	}
	mergeStatus, description := getStatusAndDescription(isBlocked)

	err = g.dispatchPrMergeStatus(ctx, owner, repoName, sha, mergeStatus, description)
	if err != nil {
		return fmt.Errorf("consumeUpdateStatus %w", err)
	}

	return nil
}

func (g *GithubCollectDataController) consumePullRequestEvent(ctx context.Context, payload EventPayload) error {
	owner := payload.Repository.Owner.Login
	repoName := payload.Repository.Name

	if payload.Action != "labeled" || payload.Label.Name != "Ok to test" {
		return nil
	}

	return g.doMergeCheck(ctx, owner, repoName, payload.PullRequest.Head.Sha)
}

func (g *GithubCollectDataController) dispatchPrMergeStatus(ctx context.Context, owner, repo, sha, status string, description string) error {
	repoStatus := &github.RepoStatus{}
	repoStatus.State = &status
	ContextStatus := "merge-check"
	repoStatus.Context = &ContextStatus
	repoStatus.Description = &description
	err := g.GithubClient.CreateCommitStatus(ctx,
		owner, repo, sha,
		repoStatus)
	if err != nil {
		return fmt.Errorf("ghClient.Repositories.CreateCommitStatus %w", err)
	}
	return nil
}

func (g *GithubCollectDataController) SetAllPrStatus(ctx context.Context, ap SetStatusPayload) error {
	err := g.MergeStatusRepo.SetRepoMergeStatus(ctx, g.DB, ap.Owner, ap.Repo, ap.BlockMerge)
	if err != nil {
		return fmt.Errorf("SetMergeStatus %w", err)
	}
	return nil
}
