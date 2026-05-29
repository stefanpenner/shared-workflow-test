package core

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func baseSummary() ShadowSummaryInput {
	return ShadowSummaryInput{
		ConsumerRepo:  "o/consumer",
		ConsumerRef:   "main",
		WorkflowsRepo: "o/workflows",
		WorkflowsRef:  "abc1234567",
		WorkflowsPR:   7,
		RunURL:        "https://example.com/run",
		PRURL:         "https://example.com/pr",
	}
}

func TestRenderShadowSummaryPassing(t *testing.T) {
	in := baseSummary()
	in.Result = ShadowPassed
	md := RenderShadowSummary(in)
	assert.Contains(t, md, "## ✅ Shadow test passed")
	assert.Contains(t, md, "| --- | --- |")
	assert.Contains(t, md, "| Result | ✅ passed |")
	assert.Contains(t, md, "| Consumer | [`o/consumer`](https://github.com/o/consumer) `@main` |")
	assert.Contains(t, md, "[PR #7](https://github.com/o/workflows/pull/7)")
	assert.Contains(t, md, "[`abc1234`](https://github.com/o/workflows/commit/abc1234567)")
	assert.Contains(t, md, "| Shadow PR | [consumer CI](https://example.com/pr) |")
}

func TestRenderShadowSummaryFailing(t *testing.T) {
	in := baseSummary()
	in.Result = ShadowFailed
	md := RenderShadowSummary(in)
	assert.Contains(t, md, "## ❌ Shadow test failed")
	assert.Contains(t, md, "| Result | ❌ failed |")
}

func TestRenderShadowSummaryOmitsShadowPRRow(t *testing.T) {
	in := baseSummary()
	in.Result = ShadowPassed
	in.PRURL = ""
	md := RenderShadowSummary(in)
	assert.NotContains(t, md, "Shadow PR")
	assert.Contains(t, md, "Runner run")
}

func TestRenderShadowLog(t *testing.T) {
	in := baseSummary()
	in.Result = ShadowPassed
	text := strings.Join(RenderShadowLog(in), "\n")
	assert.Contains(t, text, "✅ Shadow test passed: o/consumer@main")
	assert.Contains(t, text, "runner run: https://example.com/run")
	assert.Contains(t, text, "shadow PR:  https://example.com/pr")
	assert.NotContains(t, text, "](") // no markdown links
}

func TestRenderShadowLogOmitsPRLine(t *testing.T) {
	in := baseSummary()
	in.Result = ShadowFailed
	in.PRURL = ""
	text := strings.Join(RenderShadowLog(in), "\n")
	assert.Contains(t, text, "❌ Shadow test failed")
	assert.NotContains(t, text, "shadow PR")
}

func TestURLBuilders(t *testing.T) {
	assert.Equal(t, "https://github.com/o/w/pull/7", WorkflowsPrURL("o/w", 7))
	assert.Equal(t, "https://github.com/o/w/commit/abc", CommitURL("o/w", "abc"))
}

func TestRenderShadowList(t *testing.T) {
	md := RenderShadowList([]Consumer{{Repo: "o/consumer-a", Ref: "main"}, {Repo: "o/consumer-b", Ref: "dev"}}, 2, "o/runner")
	assert.Contains(t, md, "## 🛰️ Shadow tests")
	assert.Contains(t, md, "[`o/consumer-a`](https://github.com/o/consumer-a) `@main`")
	assert.Contains(t, md, "`@dev`")
	assert.Contains(t, md, "github.com/o/runner/pulls?q=")
	assert.Contains(t, md, "pr-2-o-consumer-a")
}

func TestRenderShadowListEmpty(t *testing.T) {
	md := RenderShadowList(nil, 1, "o/runner")
	assert.Contains(t, md, "## 🛰️ Shadow tests")
	assert.NotContains(t, md, "github.com/o/runner/pulls")
}
