package nocomments

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestReportsEveryCommentButDirectives(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), Analyzer, "a")
}
