package read_port_returns_an_answer

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/tsgolint/internal/archrule"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/archtypes"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

const ruleName = "read-port-returns-an-answer"

var defaultFiles = []string{"**/ports/**"}

// The same vocabulary `arch/read-port-writes-nothing` uses to decide which
// declarations are read ports. That rule judges member names; this one judges
// what the member gives back.
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

var ReadPortReturnsAnAnswerRule = archrule.Rule{Files: defaultFiles, Rule: rule.Rule{
	Name: ruleName,
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[Options](options, ruleName)
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

			for _, member := range archtypes.Members(ctx.TypeChecker, archtypes.DeclaredType(ctx.TypeChecker, node)) {
				for _, signature := range archtypes.CallSignatures(ctx.TypeChecker, member.Type) {
					// `Promise<void>` is the shape this rule exists for: the
					// annotation reads as a value and resolves to nothing.
					returned := archtypes.Unwrapped(ctx.TypeChecker,
						archtypes.ReturnType(ctx.TypeChecker, signature))
					if archtypes.IsVoidLike(returned) {
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
	},
}}
