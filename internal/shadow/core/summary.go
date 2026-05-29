package core

import (
	"fmt"
	"net/url"
	"strings"
)

// ShadowResult is the outcome of one consumer's shadow test.
type ShadowResult string

const (
	ShadowPassed ShadowResult = "passed"
	ShadowFailed ShadowResult = "failed"
)

// ShadowSummaryInput is everything needed to render one consumer's shadow result.
type ShadowSummaryInput struct {
	ConsumerRepo  string
	ConsumerRef   string
	WorkflowsRepo string
	WorkflowsRef  string
	WorkflowsPR   int
	Result        ShadowResult
	RunURL        string // the runner (receiver) run
	PRURL         string // the shadow PR (consumer CI); "" when none exists
}

// WorkflowsPrURL builds the PR URL for a repo + PR number.
func WorkflowsPrURL(repo string, pr int) string {
	return fmt.Sprintf("https://github.com/%s/pull/%d", repo, pr)
}

// CommitURL builds the commit URL for a repo + SHA.
func CommitURL(repo, sha string) string {
	return fmt.Sprintf("https://github.com/%s/commit/%s", repo, sha)
}

func repoLink(repo string) string  { return "[`" + repo + "`](https://github.com/" + repo + ")" }
func link(label, href string) string { return "[" + label + "](" + href + ")" }

func short(ref string) string {
	if len(ref) > 7 {
		return ref[:7]
	}
	return ref
}

// RenderShadowSummary renders the result as a markdown table for the job-summary page.
func RenderShadowSummary(in ShadowSummaryInput) string {
	passed := in.Result == ShadowPassed
	icon, word := "❌", "failed"
	if passed {
		icon, word = "✅", "passed"
	}
	rows := [][2]string{
		{"Result", fmt.Sprintf("%s %s", icon, word)},
		{"Consumer", fmt.Sprintf("%s `@%s`", repoLink(in.ConsumerRepo), in.ConsumerRef)},
		{"Draft", fmt.Sprintf("%s · %s · %s",
			repoLink(in.WorkflowsRepo),
			link(fmt.Sprintf("PR #%d", in.WorkflowsPR), WorkflowsPrURL(in.WorkflowsRepo, in.WorkflowsPR)),
			link("`"+short(in.WorkflowsRef)+"`", CommitURL(in.WorkflowsRepo, in.WorkflowsRef)))},
		{"Runner run", link("logs", in.RunURL)},
	}
	if in.PRURL != "" {
		rows = append(rows, [2]string{"Shadow PR", link("consumer CI", in.PRURL)})
	}
	lines := []string{
		fmt.Sprintf("## %s Shadow test %s", icon, word),
		"",
		"| | |",
		"| --- | --- |",
	}
	for _, r := range rows {
		lines = append(lines, fmt.Sprintf("| %s | %s |", r[0], r[1]))
	}
	lines = append(lines, "")
	return strings.Join(lines, "\n")
}

// RenderShadowLog renders the result as plain-text log lines (GitHub logs don't render markdown).
func RenderShadowLog(in ShadowSummaryInput) []string {
	icon := "❌"
	if in.Result == ShadowPassed {
		icon = "✅"
	}
	lines := []string{
		fmt.Sprintf("%s Shadow test %s: %s@%s", icon, in.Result, in.ConsumerRepo, in.ConsumerRef),
		fmt.Sprintf("   vs %s PR #%d (%s)", in.WorkflowsRepo, in.WorkflowsPR, short(in.WorkflowsRef)),
		fmt.Sprintf("   runner run: %s", in.RunURL),
	}
	if in.PRURL != "" {
		lines = append(lines, fmt.Sprintf("   shadow PR:  %s", in.PRURL))
	}
	return lines
}

// RenderShadowList renders the up-front index of all shadow tests: one row per consumer with a link
// to its shadow PR in the runner (from the deterministic branch name — knowable before dispatch).
func RenderShadowList(consumers []Consumer, workflowsPR int, runnerRepo string) string {
	lines := []string{
		"## 🛰️ Shadow tests",
		"",
		"_Pass/fail is on the `Shadow: …` checks above. Each shadow PR runs the consumer’s CI._",
		"",
		"| Consumer | Runner |",
		"| --- | --- |",
	}
	for _, c := range consumers {
		branch := ShadowBranchName(workflowsPR, c.Repo)
		shadowPR := "https://github.com/" + runnerRepo + "/pulls?q=" + url.QueryEscape("is:pr head:"+branch)
		lines = append(lines, fmt.Sprintf("| %s `@%s` | %s |", repoLink(c.Repo), c.Ref, link("shadow PR + run", shadowPR)))
	}
	lines = append(lines, "")
	return strings.Join(lines, "\n")
}
