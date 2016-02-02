// Copyright © 2016 See CONTRIBUTORS <ignasi.fosch@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package one2merge

import (
	one2merge "."

	"github.com/google/go-github/github"
	"reflect"
	"testing"
)

// token contains the GH token.
var token = "GITHUB_USERS_TOKEN"

// mockChangesService is a mock for github.PullRequestsService.
type mockChangesService struct {
	listPullRequests []github.PullRequest
}

// newMockChangesService creates a new ChangesService implementation.
func newMockChangesService(listPR []github.PullRequest) *mockChangesService {
	return &mockChangesService{
		listPullRequests: listPR,
	}
}

// mockChangesService is List implementation.
func (m *mockChangesService) List(owner string, repo string, opt *github.PullRequestListOptions) ([]github.PullRequest, *github.Response, error) {
	return m.listPullRequests, nil, nil
}

// mockChangesService is Get implementation.
func (m *mockChangesService) Get(owner string, repo string, number int) (*github.PullRequest, *github.Response, error) {
	return nil, nil, nil
}

// mockTicketsService is a mock for github.PullRequestsService.
type mockTicketsService struct {
	listIssueComments [][]github.IssueComment
}

// newMockTicketsService creates a new TicketsService implementation.
func newMockTicketsService(listIssueComments [][]github.IssueComment) *mockTicketsService {
	return &mockTicketsService{
		listIssueComments: listIssueComments,
	}
}

// mockTicketsService is List implementation.
func (m *mockTicketsService) ListComments(owner string, repo string, number int, opt *github.IssueListCommentsOptions) ([]github.IssueComment, *github.Response, error) {
	return nil, nil, nil
}

// mockRepositoriesService is mock for github.RepositoriesService.
type mockRepositoriesService struct {
	listCombinedStatuses []github.CombinedStatus
	current              int
}

// newMockRepositoriesService creates a new RepositoriesService implementation.
func newMockRepositoriesService(listCS []github.CombinedStatus) *mockRepositoriesService {
	return &mockRepositoriesService{
		listCombinedStatuses: listCS,
	}
}

// mockRepositoriesServices is GetCombinedStatus implmentation.
func (m *mockRepositoriesService) GetCombinedStatus(owner string, repo string, ref string, opt *github.ListOptions) (*github.CombinedStatus, *github.Response, error) {
	var combinedStatus github.CombinedStatus
	combinedStatus = m.listCombinedStatuses[m.current]
	return &combinedStatus, nil, nil
}

// Constructor for mockGHClient.
func newMockGHClient(listPR []github.PullRequest, listIssueComments [][]github.IssueComment, listCombinedStatuses []github.CombinedStatus) *one2merge.GHClient {
	client := &one2merge.GHClient{}
	client.Changes = newMockChangesService(listPR)
	client.Tickets = newMockTicketsService(listIssueComments)
	client.Repositories = newMockRepositoriesService(listCombinedStatuses)
	return client
}

func mockGetString(k string) string {
	if k == "authorization.token" {
		return token
	}
	return ""
}

func TestGetGHAuth(t *testing.T) {
	one2merge.GetString = mockGetString

	var result interface{}
	var errClient error
	result, errClient = one2merge.GetClient()

	if errClient != nil {
		t.Fatalf("GetClient returned error(%s) when everything was ok", errClient)
	}
	v, err := result.(one2merge.GHClient)
	if err {
		t.Fatalf("GetClient returned %s instead of github.Client", reflect.TypeOf(v))
	}
}

func TestCommentSuccessScore(t *testing.T) {

	testScore := func(comment string, expected int) {
		score := getCommentSuccessScore(comment)
		if expected != score {
			t.Fatalf("Bad score %v (expected %v) for comment %v", score, expected, comment)
		}
	}

	testScore("Don't do it", 0)
	testScore("Yes +1", 1)
	testScore(":+1", 1)
	testScore("-1", -1)
	testScore("Oops +1 :-1: +1", 0)
}

func newMockPullRequest(number int, title string, mergeable bool, SHA string) github.PullRequest {
	return github.PullRequest{
		Number:    &number,
		Title:     &title,
		Mergeable: &mergeable,
		Head: &github.PullRequestBranch{
			Label: nil,
			Ref:   nil,
			SHA:   &SHA,
			Repo:  nil,
			User:  nil,
		},
	}
}

func TestGetPullRequestsInfo(t *testing.T) {
	//TODO: https://github.com/gophergala2016/reviewer/issues/22
	var emptyListPR []github.PullRequest
	emptyListPR = make([]github.PullRequest, 0)
	var emptyListIC [][]github.IssueComment
	emptyListIC = make([][]github.IssueComment, 0)
	var emptyListCS []github.CombinedStatus
	emptyListCS = make([]github.CombinedStatus, 0)
	client := newMockGHClient(emptyListPR, emptyListIC, emptyListCS)

	var result []one2merge.PullRequestInfo
	var err error
	result, err = one2merge.GetPullRequestInfos(client, "user", "repo", []string{})

	if err != nil {
		t.Fatalf("Something went wrong when getting PR information")
	}
	if len(result) != 0 {
		t.Fatal("Got a populated list of PRInfos")
	}

	onePR := make([]github.PullRequest, 1)
	onePR[0] = newMockPullRequest(10, "Initial PR", false, "")
	client = newMockGHClient(onePR, emptyListIC, emptyListCS)

	result, err = one2merge.GetPullRequestInfos(client, "user", "repo", []string{})

	if err != nil {
		t.Fatalf("Something went wrong when getting PR information")
	}
	if len(result) != 1 {
		t.Fatal("Got a incorrect quantity of PRInfos:", len(result))
	}

	twoPR := make([]github.PullRequest, 2)
	twoPR[0] = newMockPullRequest(10, "Initial PR", true, "")
	twoPR[1] = newMockPullRequest(11, "Not so initial PR", false, "")
	client = newMockGHClient(twoPR, emptyListIC, emptyListCS)

	result, err = one2merge.GetPullRequestInfos(client, "user", "repo", []string{})

	if err != nil {
		t.Fatalf("Something went wrong when getting PR information")
	}
	if len(result) != 2 {
		t.Fatal("Got a incorrect quantity of PRInfos:", len(result))
	}
}

func TestIsMergeable(t *testing.T) {
	id := 1
	title := "Initial PR"
	mergeable := true
	pr := newMockPullRequest(id, title, mergeable, "")

	if !one2merge.IsMergeable(&pr) {
		t.Fatalf("PR #%d, %s should be mergeable", id, title)
	}
}

func TestPassedTests(t *testing.T) {
	var emptyListPR []github.PullRequest
	emptyListPR = make([]github.PullRequest, 0)
	var emptyListIC [][]github.IssueComment
	emptyListIC = make([][]github.IssueComment, 0)
	var oneListCS []github.CombinedStatus
	oneListCS = make([]github.CombinedStatus, 1)
	successState := "success"
	successCombinedStatus := github.CombinedStatus{
		State:         &successState,
		Name:          nil,
		SHA:           nil,
		TotalCount:    nil,
		Statuses:      nil,
		CommitURL:     nil,
		RepositoryURL: nil,
	}
	oneListCS[0] = successCombinedStatus
	client := newMockGHClient(emptyListPR, emptyListIC, oneListCS)
	id := 1
	title := "Initial PR"
	mergeable := true
	pr := newMockPullRequest(id, title, mergeable, "mysha")
	owner := "owner"
	repo := "repo"

	result, err := one2merge.PassedTests(client, &pr, owner, repo)
	if !result {
		t.Fatalf("PR #%v, %v should passed (%v)", id, title, err)
	}

	failedState := "failed"
	failedCombinedStatus := github.CombinedStatus{
		State:         &failedState,
		Name:          nil,
		SHA:           nil,
		TotalCount:    nil,
		Statuses:      nil,
		CommitURL:     nil,
		RepositoryURL: nil,
	}
	oneListCS[0] = failedCombinedStatus
	pr = newMockPullRequest(2, title, mergeable, "mysha2")

	result, err = one2merge.PassedTests(client, &pr, owner, repo)
	if result {
		t.Fatalf("PR #%v, %v should not passed (%v)", id, title, err)
	}
}
