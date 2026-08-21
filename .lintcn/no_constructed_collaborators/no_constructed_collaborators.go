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
	// Inline literal: lintcn's discovery matches `Name: "..."` in the source to
	// bind the CLI name to this rule, so a const would silently desynchronize
	// from lintcn:name above.
	Name: "no-constructed-collaborators",
	Run: archkit.Gated(defaultFiles, func(ctx rule.RuleContext, _ any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNewExpression: func(node *ast.Node) {
				constructed := node.AsNewExpression().Expression
				if constructed == nil {
					return
				}

				// A construction over anything but a written name — a call, an
				// element access — cannot be reported by name, and naming it
				// is half of what this message says.
				name := archkit.WrittenName(constructed)
				if name == "" {
					return
				}

				// The constructed value's own type, not the instance's: a
				// class's declaration is what says who wrote it.
				constructedType := ctx.TypeChecker.GetTypeAtLocation(constructed)

				for _, file := range archkit.DeclaringFiles(constructedType) {
					// Only an installed package. `new Date()`, `new Map()` and
					// `new URL()` are declared in `lib.*.d.ts` and are values,
					// not collaborators — banning them is the recorded reason
					// the syntactic version of this rule was dropped.
					if archkit.IsPackageDependency(file) {
						ctx.ReportNode(node, buildProviderConstructionMessage(
							name, "a package dependency"))
						return
					}
				}
			},
		}
	}),
}
