package nocomments

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var allow = regexp.MustCompile(`^//\s*lintcn:`)

var Analyzer = &analysis.Analyzer{
	Name: "nocomments",
	Doc:  "reports every comment that is not a compiler or lintcn directive",
	Run:  run,
}

func init() {
	Analyzer.Flags.Func("allow", "regexp for comments to permit", func(s string) error {
		re, err := regexp.Compile(s)
		if err == nil {
			allow = re
		}
		return err
	})
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		// file.Comments is the only complete list: go/ast.Walk has an explicit
		// `case *File` that skips it, so ast.Inspect never visits a free-floating
		// or trailing comment.
		src, _ := pass.ReadFile(pass.Fset.Position(file.Pos()).Filename)
		for _, group := range file.Comments {
			for _, c := range group.List {
				if isDirective(c) || allow.MatchString(c.Text) {
					continue
				}
				start, end := wholeLine(pass.Fset, src, c)
				pass.Report(analysis.Diagnostic{
					Pos:     c.Pos(),
					End:     c.End(),
					Message: "comment is not permitted; record it in docs/rule-comments.md",
					SuggestedFixes: []analysis.SuggestedFix{{
						Message:   "remove comment",
						TextEdits: []analysis.TextEdit{{Pos: start, End: end}},
					}},
				})
			}
		}
	}
	return nil, nil
}

// wholeLine widens the deletion to the whole line when the comment is the only
// thing on it, so -fix does not leave a blank line where the comment was.
func wholeLine(fset *token.FileSet, src []byte, c *ast.Comment) (token.Pos, token.Pos) {
	pos, endPos := fset.Position(c.Pos()), fset.Position(c.End())
	f := fset.File(c.Pos())
	if src == nil || endPos.Line >= f.LineCount() {
		return c.Pos(), c.End()
	}
	lineStart := int(f.LineStart(pos.Line)) - f.Base()
	nextLine := int(f.LineStart(endPos.Line+1)) - f.Base()
	before, after := src[lineStart:pos.Offset], src[endPos.Offset:nextLine]
	if strings.TrimSpace(string(before)) != "" || strings.TrimSpace(string(after)) != "" {
		return c.Pos(), c.End()
	}
	return f.LineStart(pos.Line), f.LineStart(endPos.Line + 1)
}

// Copied from go/ast, which exports IsGenerated but not isDirective. Without it
// //go:build, //go:generate and //nolint would all be reported.
func isDirective(c *ast.Comment) bool {
	if !strings.HasPrefix(c.Text, "//") {
		return false
	}
	s := c.Text[2:]
	if strings.HasPrefix(s, "line ") || strings.HasPrefix(s, "extern ") || strings.HasPrefix(s, "export ") {
		return true
	}
	colon := strings.Index(s, ":")
	if colon <= 0 || colon+1 >= len(s) {
		return false
	}
	for i := 0; i <= colon+1; i++ {
		if i == colon {
			continue
		}
		if b := s[i]; !('a' <= b && b <= 'z' || '0' <= b && b <= '9') {
			return false
		}
	}
	return true
}
