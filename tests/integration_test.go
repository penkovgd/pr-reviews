package integration_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/penkovgd/closer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const baseURL = "http://localhost:8080"

type Team struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type User struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type PullRequest struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
	CreatedAt         string   `json:"createdAt,omitempty"`
	MergedAt          string   `json:"mergedAt,omitempty"`
}

type PullRequestShort struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func makeRequest(t *testing.T, method, path string, body any) (*http.Response, []byte) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		require.NoError(t, err)
	}

	req, err := http.NewRequest(method, baseURL+path, bytes.NewBuffer(reqBody))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)

	defer closer.CloseOrPanic(nil, resp.Body)
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, respBody
}

func uniqueID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func createTeam(t *testing.T, team Team) Team {
	t.Helper()
	resp, body := makeRequest(t, "POST", "/team/add", team)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create team body: %s", string(body))

	var createResp struct {
		Team Team `json:"team"`
	}
	require.NoError(t, json.Unmarshal(body, &createResp))
	return createResp.Team
}

func getTeam(t *testing.T, teamName string) Team {
	t.Helper()
	resp, body := makeRequest(t, "GET", "/team/get?team_name="+teamName, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "get team body: %s", string(body))

	var team Team
	require.NoError(t, json.Unmarshal(body, &team))
	return team
}

func setUserActive(t *testing.T, userID string, active bool) User {
	t.Helper()
	req := map[string]any{"user_id": userID, "is_active": active}
	resp, body := makeRequest(t, "POST", "/users/setIsActive", req)
	require.Equal(t, http.StatusOK, resp.StatusCode, "setIsActive body: %s", string(body))

	var r struct {
		User User `json:"user"`
	}
	require.NoError(t, json.Unmarshal(body, &r))
	return r.User
}

func createPR(t *testing.T, prID, name, author string, expectStatus int) PullRequest {
	t.Helper()
	req := map[string]string{"pull_request_id": prID, "pull_request_name": name, "author_id": author}
	resp, body := makeRequest(t, "POST", "/pullRequest/create", req)
	require.Equal(t, expectStatus, resp.StatusCode, "createPR body: %s", string(body))

	if expectStatus == http.StatusCreated {
		var r struct {
			PR PullRequest `json:"pr"`
		}
		require.NoError(t, json.Unmarshal(body, &r))
		return r.PR
	}
	t.Fatalf("unexpected createPR status: %d", resp.StatusCode)
	return PullRequest{}
}

func mergePR(t *testing.T, prID string, expectStatus int) PullRequest {
	t.Helper()
	req := map[string]string{"pull_request_id": prID}
	resp, body := makeRequest(t, "POST", "/pullRequest/merge", req)
	require.Equal(t, expectStatus, resp.StatusCode, "mergePR body: %s", string(body))

	var r struct {
		PR PullRequest `json:"pr"`
	}
	require.NoError(t, json.Unmarshal(body, &r))
	return r.PR
}

func reassignPR(t *testing.T, prID, oldUserID string, expectStatus int) (PullRequest, string) {
	t.Helper()
	req := map[string]string{"pull_request_id": prID, "old_reviewer_id": oldUserID}
	resp, body := makeRequest(t, "POST", "/pullRequest/reassign", req)
	require.Equal(t, expectStatus, resp.StatusCode, "reassign body: %s", string(body))

	if expectStatus == http.StatusOK {
		var r struct {
			PR         PullRequest `json:"pr"`
			ReplacedBy string      `json:"replaced_by"`
		}
		require.NoError(t, json.Unmarshal(body, &r))
		return r.PR, r.ReplacedBy
	}

	var errResp ErrorResponse
	_ = json.Unmarshal(body, &errResp)
	return PullRequest{}, errResp.Error.Code
}

func TestTeamCreate_Get(t *testing.T) {
	teamName := uniqueID("team")
	team := Team{
		TeamName: teamName,
		Members: []TeamMember{
			{UserID: "u1", Username: "A", IsActive: true},
			{UserID: "u2", Username: "B", IsActive: true},
		},
	}
	created := createTeam(t, team)
	assert.Equal(t, teamName, created.TeamName)
	assert.Len(t, created.Members, 2)

	got := getTeam(t, teamName)
	assert.Equal(t, teamName, got.TeamName)
	assert.Len(t, got.Members, 2)
}

func TestUserSetIsActive(t *testing.T) {
	teamName := uniqueID("team")
	userID := "user-activate"
	team := Team{
		TeamName: teamName,
		Members: []TeamMember{
			{UserID: userID, Username: "C", IsActive: true},
		},
	}
	_ = createTeam(t, team)

	u := setUserActive(t, userID, false)
	assert.False(t, u.IsActive)
	assert.Equal(t, userID, u.UserID)

	u = setUserActive(t, userID, true)
	assert.True(t, u.IsActive)
}

func TestPRCreate_AssignsUpToTwoActiveReviewers(t *testing.T) {
	teamName := uniqueID("team")
	author := "author-a"
	rev1 := "r1"
	rev2 := "r2"

	team := Team{
		TeamName: teamName,
		Members: []TeamMember{
			{UserID: author, Username: "Author", IsActive: true},
			{UserID: rev1, Username: "R1", IsActive: true},
			{UserID: rev2, Username: "R2", IsActive: true},
		},
	}
	_ = createTeam(t, team)

	prID := uniqueID("pr")
	pr := createPR(t, prID, "pr test", author, http.StatusCreated)

	assert.Equal(t, "OPEN", pr.Status)
	assert.LessOrEqual(t, len(pr.AssignedReviewers), 2)
	for _, rid := range pr.AssignedReviewers {
		assert.NotEqual(t, author, rid)
		found := false
		for _, m := range team.Members {
			if m.UserID == rid {
				found = true
				assert.True(t, m.IsActive)
			}
		}
		assert.True(t, found, "assigned reviewer must be from the author's team")
	}
}

func TestPRCreate_NoCandidates(t *testing.T) {
	teamName := uniqueID("team")
	author := "author-b"
	team := Team{
		TeamName: teamName,
		Members: []TeamMember{
			{UserID: author, Username: "AuthorB", IsActive: true},
			{UserID: "x1", Username: "X1", IsActive: false},
			{UserID: "x2", Username: "X2", IsActive: false},
		},
	}
	_ = createTeam(t, team)

	prID := uniqueID("pr")
	pr := createPR(t, prID, "no candidates", author, http.StatusCreated)

	assert.Equal(t, 0, len(pr.AssignedReviewers), "expected no assigned reviewers when no active candidates")
}

func TestPRMerge_Idempotent(t *testing.T) {
	teamName := uniqueID("team")
	author := "author-c"
	rev1 := "r10"
	team := Team{
		TeamName: teamName,
		Members: []TeamMember{
			{UserID: author, Username: "AuthorC", IsActive: true},
			{UserID: rev1, Username: "R10", IsActive: true},
		},
	}
	_ = createTeam(t, team)

	prID := uniqueID("pr")
	_ = createPR(t, prID, "merge test", author, http.StatusCreated)

	pr := mergePR(t, prID, http.StatusOK)
	assert.Equal(t, "MERGED", pr.Status)
	assert.NotZero(t, pr.MergedAt)

	pr2 := mergePR(t, prID, http.StatusOK)
	assert.Equal(t, "MERGED", pr2.Status)
	assert.Equal(t, pr.PullRequestID, pr2.PullRequestID)
}

func TestPRReassign_Success(t *testing.T) {
	teamName := uniqueID("team")
	author := "author-d"
	old := "oldrev"
	rev := "rev"
	candidate := "candidate"
	team := Team{
		TeamName: teamName,
		Members: []TeamMember{
			{UserID: author, Username: "AuthorD", IsActive: true},
			{UserID: old, Username: "OldR", IsActive: true},
			{UserID: rev, Username: "Rev", IsActive: true},
			{UserID: candidate, Username: "Cand", IsActive: true},
		},
	}
	_ = createTeam(t, team)

	prID := uniqueID("pr")
	pr := createPR(t, prID, "reassign test", author, http.StatusCreated)
	var toReplace string
	if len(pr.AssignedReviewers) == 0 {
		t.Fatalf("expected at least one assigned reviewer to test reassign; assigned: %v", pr.AssignedReviewers)
	}
	toReplace = pr.AssignedReviewers[0]

	prAfter, replacedBy := reassignPR(t, prID, toReplace, http.StatusOK)
	assert.NotEqual(t, toReplace, replacedBy)
	found := false
	for _, r := range prAfter.AssignedReviewers {
		if r == replacedBy {
			found = true
		}
	}
	assert.True(t, found, "replaced_by should be present in PR assigned reviewers")
}

func TestPRReassign_FailOnMerged(t *testing.T) {
	teamName := uniqueID("team")
	author := "author"
	old := "oldrev"
	rev := "rev"
	candidate := "cand"
	team := Team{
		TeamName: teamName,
		Members: []TeamMember{
			{UserID: author, Username: "Author", IsActive: true},
			{UserID: old, Username: "Old", IsActive: true},
			{UserID: rev, Username: "Rev", IsActive: true},
			{UserID: candidate, Username: "Cand", IsActive: true},
		},
	}
	_ = createTeam(t, team)

	prID := uniqueID("pr")
	pr := createPR(t, prID, "reassign fail merged", author, http.StatusCreated)
	if len(pr.AssignedReviewers) == 0 {
		t.Fatalf("expected assigned reviewers to test reassign fail")
	}
	toReplace := pr.AssignedReviewers[0]

	_ = mergePR(t, prID, http.StatusOK)

	_, errCode := reassignPR(t, prID, toReplace, http.StatusConflict)
	assert.Equal(t, "PR_MERGED", errCode)
}

func TestPRReassign_FailNoCandidate(t *testing.T) {
	teamName := uniqueID("team")
	author := "author"
	old := "oldrev"
	rev := "rev"
	team := Team{
		TeamName: teamName,
		Members: []TeamMember{
			{UserID: author, Username: "Author", IsActive: true},
			{UserID: old, Username: "Old", IsActive: true},
			{UserID: rev, Username: "Rev", IsActive: true},
		},
	}
	_ = createTeam(t, team)

	prID := uniqueID("pr")
	pr := createPR(t, prID, "reassign no candidates", author, http.StatusCreated)
	if len(pr.AssignedReviewers) == 0 {
		t.Fatalf("expected assigned reviewers to test reassign fail")
	}
	toReplace := pr.AssignedReviewers[0]

	_, errCode := reassignPR(t, prID, toReplace, http.StatusConflict)
	assert.Equal(t, "NO_CANDIDATE", errCode)
}

func TestUsersGetReview(t *testing.T) {
	teamName := uniqueID("team")
	author := "author-f"
	reviewer := "rev-f"
	team := Team{
		TeamName: teamName,
		Members: []TeamMember{
			{UserID: author, Username: "AuthorF", IsActive: true},
			{UserID: reviewer, Username: "RevF", IsActive: true},
		},
	}
	_ = createTeam(t, team)

	prID := uniqueID("pr")
	_ = createPR(t, prID, "get review test", author, http.StatusCreated)

	resp, body := makeRequest(t, "GET", "/users/getReview?user_id="+reviewer, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "getReview body: %s", string(body))

	var respStruct struct {
		UserID       string             `json:"user_id"`
		PullRequests []PullRequestShort `json:"pull_requests"`
	}
	require.NoError(t, json.Unmarshal(body, &respStruct))
	assert.Equal(t, reviewer, respStruct.UserID)
}
