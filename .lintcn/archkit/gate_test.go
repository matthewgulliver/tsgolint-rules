package archkit

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/typescript-eslint/tsgolint/internal/rule"
)

func sourceFile(t *testing.T, fileName string) *ast.SourceFile {
	t.Helper()
	return parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: fileName}, "", core.ScriptKindTS)
}

func TestGated(t *testing.T) {
	t.Parallel()

	var ran bool
	gated := Gated([]string{"**/e2e/**"}, func(rule.RuleContext, any) rule.RuleListeners {
		ran = true
		return rule.RuleListeners{}
	})

	if listeners := gated(rule.RuleContext{}, nil); listeners != nil {
		t.Errorf("out-of-scope file got listeners %v, want nil", listeners)
	}
	if ran {
		t.Error("out-of-scope file ran the rule")
	}

	if listeners := gated(rule.RuleContext{SourceFile: sourceFile(t, "/repo/apps/web/e2e/spec.ts")}, nil); listeners == nil {
		t.Error("in-scope file got no listeners")
	}
	if !ran {
		t.Error("in-scope file did not run the rule")
	}
}
