// lintcn:name no-constructed-collaborators
// lintcn:severity error
// lintcn:description Disallow the inside of the hexagon from constructing a provider a package dependency declares

package no_constructed_collaborators

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/lintcn-rules/archkit"
)

var defaultFiles = []string{
	"**/hexagon/application/**",
	"**/hexagon/domain/**",
}

func buildProviderConstructionMessage(name string, source string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "providerConstruction",
		Description: "`new " + name + "` constructs a type declared by " + source +
			", so this file chooses its own collaborator and no test can substitute another.",
		Help: "Take the capability as a port parameter and let the composition root construct the provider.",
	}
}

var NoConstructedCollaboratorsRule = rule.Rule{
	Name: "no-constructed-collaborators",
	Run: archkit.Gated(defaultFiles, func(ctx rule.RuleContext, _ any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNewExpression: func(node *ast.Node) {
				constructed := node.AsNewExpression().Expression
				if constructed == nil {
					return
				}

				name := archkit.WrittenName(constructed)
				if name == "" {
					return
				}

				constructedType := ctx.TypeChecker.GetTypeAtLocation(constructed)

				if archkit.DeclaredByPackageDependency(archkit.DeclaringFiles(constructedType)) {
					ctx.ReportNode(node, buildProviderConstructionMessage(
						name, "a package dependency"))
					return
				}
			},
		}
	}),
}
