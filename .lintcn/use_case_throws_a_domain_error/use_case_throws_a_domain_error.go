// lintcn:name use-case-throws-a-domain-error
// lintcn:severity error
// lintcn:description Disallow throwing error types this repository does not declare inside the hexagon — a package dependency anywhere in it, a standard-library Error in the application tree

package use_case_throws_a_domain_error

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
	"github.com/typescript-eslint/tsgolint/lintcn-rules/archkit"
)

var defaultFiles = []string{
	"**/hexagon/application/**",
	"**/hexagon/domain/**",
}

// A bare `Error` is the hexagonal skill's own way to reject an impossible
// state at construction — its worked example throws one from `createMoney`
// while returning `exceeds-budget` as a result value in the same function. An
// invariant violation is not a business outcome, so the standard-library arm
// judges the application tree only; the package-dependency arm judges both.
var defaultStandardLibraryFiles = []string{
	"**/hexagon/application/**",
}

type Options struct {
	StandardLibraryFiles []string `json:"standardLibraryFiles,omitempty"`
}

func (o Options) standardLibraryPatterns() []string {
	if len(o.StandardLibraryFiles) == 0 {
		return defaultStandardLibraryFiles
	}
	return o.StandardLibraryFiles
}

func buildForeignThrowMessage(name string, source string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "foreignThrow",
		Description: "This throws `" + name + "`, declared by " + source +
			", so the failure carries no business meaning and every surface must invent copy for it.",
		Help: "Return the expected outcome as a member of the result union, or throw an error type the domain declares.",
	}
}

var UseCaseThrowsADomainErrorRule = rule.Rule{
	// Inline literal: lintcn's discovery matches `Name: "..."` in the source to
	// bind the CLI name to this rule, so a const would silently desynchronize
	// from lintcn:name above.
	Name: "use-case-throws-a-domain-error",
	Run: archkit.Gated(defaultFiles, func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[Options](options, "use-case-throws-a-domain-error")
		return rule.RuleListeners{
			ast.KindThrowStatement: func(node *ast.Node) {
				thrown := node.AsThrowStatement().Expression
				if thrown == nil {
					return
				}

				thrownType := ctx.TypeChecker.GetTypeAtLocation(thrown)
				files := archkit.DeclaringFiles(thrownType)
				if len(files) == 0 {
					// A rethrown `unknown` has no declaration to judge, and
					// guessing is how a rule starts reporting on nothing.
					return
				}

				judgesStandardLibrary := archkit.Includes(
					opts.standardLibraryPatterns(), ctx.SourceFile.FileName())

				for _, file := range files {
					if archkit.IsPackageDependency(file) {
						ctx.ReportNode(node, buildForeignThrowMessage(
							ctx.TypeChecker.TypeToString(thrownType), "a package dependency"))
						return
					}
					if judgesStandardLibrary && archkit.IsStandardLibrary(file) {
						ctx.ReportNode(node, buildForeignThrowMessage(
							ctx.TypeChecker.TypeToString(thrownType), "the standard library"))
						return
					}
				}
			},
		}
	}),
}
