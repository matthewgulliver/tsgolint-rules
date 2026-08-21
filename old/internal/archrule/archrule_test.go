package archrule

import "testing"

func TestDocumentedAtMatchesTheJavaScriptHalf(t *testing.T) {
	t.Parallel()
	// `packages/oxlint/rule.ts` builds `${SOURCE}/${rule}.md` from the same base. The two
	// halves drifting would send readers of one to a page that does not exist.
	const want = "https://github.com/matthewgulliver/tsgolint-rules/blob/main/docs/rules/no-async-in-domain.md"
	if got := DocumentedAt("no-async-in-domain"); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
