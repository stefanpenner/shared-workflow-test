package adapters

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v66/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stefanpenner-cs/reusable-workflows/internal/shadow/core"
)

func testClient(t *testing.T, h http.Handler) *github.Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	cl := github.NewClient(nil)
	u, _ := url.Parse(srv.URL + "/")
	cl.BaseURL = u
	return cl
}

func json200(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func TestDispatchReceiverExtractsRunID(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/o/runner/actions/workflows/receiver.yaml/dispatches", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		json200(w, map[string]any{"workflow_run_id": 1234567890})
	})
	id, err := dispatchReceiver(testClient(t, mux), "o/runner", core.ShadowContext{WorkflowsPR: 7})
	require.NoError(t, err)
	assert.Equal(t, 1234567890, id)
}

func TestFindPrURL(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/o/runner/pulls", func(w http.ResponseWriter, _ *http.Request) {
		json200(w, []map[string]any{{"html_url": "https://github.com/o/runner/pull/5", "number": 5}})
	})
	u, err := findPrURL(testClient(t, mux), "o/runner", "shadow/x")
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/o/runner/pull/5", u)
}

func TestFindPrURLNone(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/o/runner/pulls", func(w http.ResponseWriter, _ *http.Request) { json200(w, []any{}) })
	u, err := findPrURL(testClient(t, mux), "o/runner", "shadow/x")
	require.NoError(t, err)
	assert.Equal(t, "", u)
}

func TestEnsurePRCreatesWhenNone(t *testing.T) {
	created := false
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/o/runner/pulls", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			json200(w, []any{})
			return
		}
		created = true
		json200(w, map[string]any{"html_url": "https://github.com/o/runner/pull/9"})
	})
	u, err := ensurePR(testClient(t, mux), "o/runner", "shadow/x", "main", "t", "b")
	require.NoError(t, err)
	assert.True(t, created)
	assert.Equal(t, "https://github.com/o/runner/pull/9", u)
}

func TestEnsurePRReturnsExisting(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/o/runner/pulls", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method) // must NOT create
		json200(w, []map[string]any{{"html_url": "https://github.com/o/runner/pull/3"}})
	})
	u, err := ensurePR(testClient(t, mux), "o/runner", "shadow/x", "main", "t", "b")
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/o/runner/pull/3", u)
}

func TestAwaitRunSuccessAndFailure(t *testing.T) {
	ok := http.NewServeMux()
	ok.HandleFunc("/repos/o/runner/actions/runs/42", func(w http.ResponseWriter, _ *http.Request) {
		json200(w, map[string]any{"status": "completed", "conclusion": "success"})
	})
	assert.NoError(t, awaitRun(testClient(t, ok), "o/runner", 42, "run", 3, 0))

	bad := http.NewServeMux()
	bad.HandleFunc("/repos/o/runner/actions/runs/42", func(w http.ResponseWriter, _ *http.Request) {
		json200(w, map[string]any{"status": "completed", "conclusion": "failure"})
	})
	assert.ErrorContains(t, awaitRun(testClient(t, bad), "o/runner", 42, "run", 3, 0), "failure")

	pending := http.NewServeMux()
	pending.HandleFunc("/repos/o/runner/actions/runs/42", func(w http.ResponseWriter, _ *http.Request) {
		json200(w, map[string]any{"status": "in_progress", "conclusion": nil})
	})
	assert.ErrorContains(t, awaitRun(testClient(t, pending), "o/runner", 42, "run", 2, 0), "timed out")
}

func TestWatchCommitRun(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/o/runner/actions/runs", func(w http.ResponseWriter, _ *http.Request) {
		json200(w, map[string]any{"total_count": 1, "workflow_runs": []map[string]any{{"id": 42}}})
	})
	mux.HandleFunc("/repos/o/runner/actions/runs/42", func(w http.ResponseWriter, _ *http.Request) {
		json200(w, map[string]any{"status": "completed", "conclusion": "success"})
	})
	assert.NoError(t, watchCommitRun(testClient(t, mux), "o/runner", "deadbeef", 3, 0, 3, 0))
}

func TestClosePRAndDeleteBranch(t *testing.T) {
	closed, deleted := false, false
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/o/runner/pulls", func(w http.ResponseWriter, _ *http.Request) {
		json200(w, []map[string]any{{"number": 5, "html_url": "u"}})
	})
	mux.HandleFunc("/repos/o/runner/pulls/5", func(w http.ResponseWriter, _ *http.Request) {
		closed = true
		json200(w, map[string]any{"number": 5})
	})
	mux.HandleFunc("/repos/o/runner/git/refs/heads/shadow/x", func(w http.ResponseWriter, _ *http.Request) {
		deleted = true
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, closePRAndDeleteBranch(testClient(t, mux), "o/runner", "shadow/x"))
	assert.True(t, closed)
	assert.True(t, deleted)
}
