// lintcn:name domain-function-returns-an-answer
// lintcn:severity error
// lintcn:description Disallow exported domain functions with no written return type whose inferred return is void or undefined on every path

package domain_function_returns_an_answer

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/lintcn-rules/archkit"
)

var defaultFiles = []string{"**/hexagon/domain/**"}

func buildAnswerlessDomainFunctionMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "answerlessDomainFunction",
		Description: "`" + name +
			"` returns nothing, so its caller learns neither the next state nor why the transition was refused, and the effect it must have had lives somewhere the model cannot show.",
		Help: "Return the new aggregate or an explicit business refusal; leave the effect to the caller.",
	}
}

func finalReturn(c *checker.Checker, t *checker.Type) *checker.Type {
	for range 8 {
		signatures := archkit.CallSignatures(c, t)
		if len(signatures) == 0 {
			return t
		}
		t = archkit.ReturnType(c, signatures[len(signatures)-1])
	}
	return t
}

func answerless(t *checker.Type) bool {
	if t == nil {
		return false
	}
	for _, constituent := range archkit.Constituents(t) {
		if !archkit.IsVoidLike(constituent) {
			return false
		}
	}
	return true
}

var DomainFunctionReturnsAnAnswerRule = rule.Rule{
	Name: "domain-function-returns-an-answer",
	Run: archkit.Gated(defaultFiles, func(ctx rule.RuleContext, _ any) rule.RuleListeners {
		judge := func(node *ast.Node, function *ast.Node, name *ast.Node) {
			if name == nil || !ast.HasSyntacticModifier(exported(node), ast.ModifierFlagsExport) {
				return
			}
			if function.Type() != nil {
				return
			}
			if answerless(finalReturn(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(node))) {
				ctx.ReportNode(name, buildAnswerlessDomainFunctionMessage(name.Text()))
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: func(node *ast.Node) { judge(node, node, node.Name()) },
			ast.KindVariableDeclaration: func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()
				if declaration.Initializer == nil || !ast.IsFunctionLike(declaration.Initializer) {
					return
				}
				judge(node, declaration.Initializer, declaration.Name())
			},
		}
	}),
}

func exported(node *ast.Node) *ast.Node {
	if node.Kind != ast.KindVariableDeclaration {
		return node
	}
	list := node.Parent
	if list == nil || list.Parent == nil {
		return node
	}
	return list.Parent
}
