// lintcn:name read-port-returns-an-answer
// lintcn:severity error
// lintcn:description Disallow read port methods that return nothing instead of the answer

package read_port_returns_an_answer

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
	"github.com/typescript-eslint/tsgolint/lintcn-rules/archkit"
)

var defaultFiles = []string{"**/ports/**"}

var defaultReadPortPatterns = []string{"Rows$", "Reader$", "View$", "Dashboard$"}

type Options struct {
	ReadPortPatterns []string `json:"readPortPatterns,omitempty"`
}

func buildAnswerlessReadMessage(port string, member string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "answerlessRead",
		Description: "`" + port + "." + member +
			"` returns nothing, so the only reason to call it is its effect — and this declaration is a read port, which a reviewer skims as a question.",
		Help: "Move the operation to the write port that owns the consistency boundary, or return the answer the caller asked for.",
	}
}

func compile(patterns []string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		if expression, err := regexp.Compile(pattern); err == nil {
			compiled = append(compiled, expression)
		}
	}
	return compiled
}

var ReadPortReturnsAnAnswerRule = rule.Rule{
	Name: "read-port-returns-an-answer",
	Run: archkit.Gated(defaultFiles, func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[Options](options, "read-port-returns-an-answer")
		patterns := opts.ReadPortPatterns
		if len(patterns) == 0 {
			patterns = defaultReadPortPatterns
		}
		readPort := compile(patterns)

		judge := func(node *ast.Node) {
			if !ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) {
				return
			}
			name := node.Name()
			if name == nil {
				return
			}
			port := name.Text()
			named := false
			for _, pattern := range readPort {
				if pattern.MatchString(port) {
					named = true
					break
				}
			}
			if !named {
				return
			}

			for _, member := range archkit.Members(ctx.TypeChecker, archkit.DeclaredType(ctx.TypeChecker, node)) {
				for _, signature := range archkit.CallSignatures(ctx.TypeChecker, member.Type) {
					returned := archkit.Unwrapped(ctx.TypeChecker,
						archkit.ReturnType(ctx.TypeChecker, signature))
					if archkit.IsVoidLike(returned) {
						ctx.ReportNode(node, buildAnswerlessReadMessage(port, member.Name))
						return
					}
				}
			}
		}

		return rule.RuleListeners{
			ast.KindInterfaceDeclaration: judge,
			ast.KindTypeAliasDeclaration: judge,
		}
	}),
}
