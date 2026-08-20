package use_case_throws_a_domain_error

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/tsgolint/internal/archrule"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/archscope"
	"github.com/typescript-eslint/tsgolint/internal/archtypes"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

const ruleName = "use-case-throws-a-domain-error"

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

var UseCaseThrowsADomainErrorRule = archrule.Rule{Files: defaultFiles, Rule: rule.Rule{
	Name: ruleName,
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[Options](options, ruleName)
		return rule.RuleListeners{
			ast.KindThrowStatement: func(node *ast.Node) {
				thrown := node.AsThrowStatement().Expression
				if thrown == nil {
					return
				}

				thrownType := ctx.TypeChecker.GetTypeAtLocation(thrown)
				files := archtypes.DeclaringFiles(thrownType)
				if len(files) == 0 {
					// A rethrown `unknown` has no declaration to judge, and
					// guessing is how a rule starts reporting on nothing.
					return
				}

				judgesStandardLibrary := archscope.Includes(
					opts.standardLibraryPatterns(), ctx.SourceFile.FileName())

				for _, file := range files {
					if archscope.IsPackageDependency(file) {
						ctx.ReportNode(node, buildForeignThrowMessage(
							ctx.TypeChecker.TypeToString(thrownType), "a package dependency"))
						return
					}
					if judgesStandardLibrary && archscope.IsStandardLibrary(file) {
						ctx.ReportNode(node, buildForeignThrowMessage(
							ctx.TypeChecker.TypeToString(thrownType), "the standard library"))
						return
					}
				}
			},
		}
	},
}}
