package nocomments

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestReportsEveryCommentButDirectives(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "a")
}
