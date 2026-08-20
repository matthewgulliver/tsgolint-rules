package no_constructed_collaborators

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/tsgolint/internal/archrule"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/archscope"
	"github.com/typescript-eslint/tsgolint/internal/archtypes"
)

const ruleName = "no-constructed-collaborators"

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

var NoConstructedCollaboratorsRule = archrule.Rule{Files: defaultFiles, Rule: rule.Rule{
	Name: ruleName,
	Run: func(ctx rule.RuleContext, _ any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNewExpression: func(node *ast.Node) {
				constructed := node.AsNewExpression().Expression
				if constructed == nil {
					return
				}

				// A construction over anything but a written name — a call, an
				// element access — cannot be reported by name, and naming it
				// is half of what this message says.
				name := archtypes.WrittenName(constructed)
				if name == "" {
					return
				}

				// The constructed value's own type, not the instance's: a
				// class's declaration is what says who wrote it.
				constructedType := ctx.TypeChecker.GetTypeAtLocation(constructed)

				for _, file := range archtypes.DeclaringFiles(constructedType) {
					// Only an installed package. `new Date()`, `new Map()` and
					// `new URL()` are declared in `lib.*.d.ts` and are values,
					// not collaborators — banning them is the recorded reason
					// the syntactic version of this rule was dropped.
					if archscope.IsPackageDependency(file) {
						ctx.ReportNode(node, buildProviderConstructionMessage(
							name, "a package dependency"))
						return
					}
				}
			},
		}
	},
}}
