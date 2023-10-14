package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/draft/configurations"
	"github.com/manabie-com/backend/internal/draft/entities"
	"github.com/manabie-com/backend/internal/golibs/configs"
	mock_repositories "github.com/manabie-com/backend/mock/draft/repositories"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v41/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

var testConfig = configurations.Config{
	Github: configs.GithubConfig{},
}

func baseWebhookPayload() EventPayload {
	ret := EventPayload{
		Ref:    "refs/heads/main",
		After:  "065e5b2e533963b0bd1de32e943f982020e23f17",
		Before: "d67353690a220d82b7096deb5b2800c8222e610b",
	}
	ret.PullRequest.Number = 1
	ret.Repository.Name = "backend"
	ret.Repository.Owner.Login = "manabie-com"
	ret.PullRequest.Base.Ref = "develop"
	ret.Issue.Number = 1
	return ret
}

var baseGithubDataReturn = entities.GithubPRData{
	ID:                  1,
	Number:              1,
	BranchName:          "test",
	Create:              time.Now(),
	Close:               time.Now(),
	NumOfComments:       0,
	TotalToFirstComment: 0,
	TotalTimeConsuming:  0,
	IsMerged:            false,
}

var nilData *entities.GithubPRData

func TestAddAllEventData(t *testing.T) {
	t.Run("successfully add raw data of github events to db", func(t *testing.T) {
		var testEvent = "push"
		testPayLoadByte, _ := json.Marshal(baseWebhookPayload())

		githubEvent := &mock_repositories.MockGithubEvent{}
		githubEvent.On("AddEventRawData", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		ctr := &GithubCollectDataController{
			GithubEventRepo: githubEvent,
		}

		err := ctr.AddAllEventData(context.Background(), testEvent, testPayLoadByte)
		assert.Nil(t, err)
	})
}

type githubTest struct {
	cl         *mock_repositories.MockGithubClient
	eventRepo  *mock_repositories.MockGithubEvent
	statusRepo *mock_repositories.MockGithubMergeStatusRepo
	prRepo     *mock_repositories.MockGithubPrRepo
	controller *GithubCollectDataController
}

func newGithubTest() *githubTest {
	githubEvent := &mock_repositories.MockGithubEvent{}
	mergeStatusRepo := &mock_repositories.MockGithubMergeStatusRepo{}
	prRepo := &mock_repositories.MockGithubPrRepo{}
	cl := &mock_repositories.MockGithubClient{}
	ctr := &GithubCollectDataController{
		Logger:          zap.NewNop().Sugar(),
		GithubEventRepo: githubEvent,
		CFG:             &testConfig,
		MergeStatusRepo: mergeStatusRepo,
		GithubClient:    cl,
	}
	return &githubTest{
		eventRepo:  githubEvent,
		statusRepo: mergeStatusRepo,
		prRepo:     prRepo,
		cl:         cl,
		controller: ctr,
	}
}
func TestConsumeIssueCommentEvent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	t.Run("open", func(t *testing.T) {
		t.Run("success dispatch merge-check", func(t *testing.T) {
			testPayload := baseWebhookPayload()
			testPayload.Comment.Body = "/draft/merge-check"
			suite := newGithubTest()
			commit := &github.RepositoryCommit{SHA: &testPayload.PullRequest.Head.Sha}
			suite.cl.On("GetLastCommit", mock.Anything, "manabie-com", "backend", 1).Once().Return(commit, nil)
			suite.statusRepo.On("GetRepoMergeStatus", mock.Anything, mock.Anything, "manabie-com", "backend").Once().Return(false, nil)
			suite.cl.On("CreateCommitStatus", mock.Anything, "manabie-com", "backend", "commitID", mock.Anything).Once().Return(nil)
			err := suite.controller.consumePullRequestEvent(ctx, testPayload)
			assert.Nil(t, err)
		})
	})
}

func TestConsumePullRequestEvent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	t.Run("open", func(t *testing.T) {
		t.Run("success insert pr allowed merge", func(t *testing.T) {
			testPayload := baseWebhookPayload()
			testPayload.Action = "labeled"
			testPayload.Label.Name = "Ok to Test"
			testPayload.PullRequest.Head.Ref = "feature/a"
			testPayload.PullRequest.Head.Sha = "commitID"
			suite := newGithubTest()
			suite.statusRepo.On("GetRepoMergeStatus", mock.Anything, mock.Anything, "manabie-com", "backend").Once().Return(false, nil)
			suite.cl.On("CreateCommitStatus", mock.Anything, "manabie-com", "backend", "commitID", mock.Anything).Once().Return(nil)
			err := suite.controller.consumePullRequestEvent(ctx, testPayload)
			assert.Nil(t, err)
		})

		t.Run("success insert pr block merge", func(t *testing.T) {
			testPayload := baseWebhookPayload()
			testPayload.Action = "labeled"
			testPayload.PullRequest.Head.Ref = "feature/a"
			testPayload.PullRequest.Head.Sha = "commitID"
			suite := newGithubTest()
			suite.statusRepo.On("GetRepoMergeStatus", mock.Anything, mock.Anything, "manabie-com", "backend").Once().Return(true, nil)
			suite.cl.On("CreateCommitStatus", mock.Anything, "manabie-com", "backend", "commitID", mock.Anything).Once().Return(nil)
			err := suite.controller.consumePullRequestEvent(ctx, testPayload)
			assert.Nil(t, err)
		})
	})
}

func Test_SetAllPRStatus(t *testing.T) {
	t.Parallel()

	suite := newGithubTest()
	pl := SetStatusPayload{
		Repo:       "backend",
		Owner:      "manabie-com",
		BlockMerge: false,
	}
	suite.controller.CFG.GitHubWebhookSecret = "secret"
	rawReq, err := json.Marshal(pl)
	assert.NoError(t, err)

	suite.statusRepo.On("SetRepoMergeStatus", mock.Anything, mock.Anything, "manabie-com", "backend", false).Once().Return(nil)

	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)
	c.Request, err = http.NewRequest("POST", "", bytes.NewReader(rawReq))
	c.Request.Header.Add("X-Hub-Signature-256", "secret")
	assert.NoError(t, err)
	suite.controller.HandleStatus(c)
	assert.Equal(t, 200, r.Code)
}
