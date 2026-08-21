// lintcn:name no-double-library-in-use-case-test
// lintcn:severity warn
// lintcn:description Disallow test-double library scaffolding in use-case tests

package no_double_library_in_use_case

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
	"github.com/typescript-eslint/tsgolint/lintcn-rules/archkit"
)

var defaultFiles = []string{
	"**/hexagon/application/**/*.test.ts",
	"**/hexagon/application/**/*.test.tsx",
}

var defaultDoubleMethodNames = []string{"fn", "mock", "spyOn"}
var defaultDoubleLibraryPathFragments = []string{
	"/node_modules/vitest/",
	"/node_modules/jest/",
	"/node_modules/@types/jest/",
}

type Options struct {
	DoubleMethodNames          []string `json:"doubleMethodNames,omitempty"`
	DoubleLibraryPathFragments []string `json:"doubleLibraryPathFragments,omitempty"`
}

func (o Options) doubleMethodNames() []string {
	if len(o.DoubleMethodNames) == 0 {
		return defaultDoubleMethodNames
	}
	return o.DoubleMethodNames
}

func (o Options) doubleLibraryPathFragments() []string {
	if len(o.DoubleLibraryPathFragments) == 0 {
		return defaultDoubleLibraryPathFragments
	}
	return o.DoubleLibraryPathFragments
}

func buildDoubleLibraryMessage(method string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "doubleLibrary",
		Description: "`" + method + "` comes from a test-double library, so this use-case test doubles a driven port with library scaffolding rather than a hand-written fake, stub or spy that implements the port's real interface — a contract change no longer breaks at compile time.",
		Help:        "Write the double as `createFake…`, `createStub…` or `createSpy…` in the doubles module, implementing the port interface, and assert on its state.",
	}
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

var NoDoubleLibraryInUseCaseTestRule = rule.Rule{
	// Inline literal: lintcn's discovery matches `Name: "..."` in the source to
	// bind the CLI name to this rule, so a const would silently desynchronize
	// from lintcn:name above.
	Name: "no-double-library-in-use-case-test",
	Run: archkit.Gated(defaultFiles, func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[Options](options, "no-double-library-in-use-case-test")
		methods := opts.doubleMethodNames()
		fragments := opts.doubleLibraryPathFragments()

		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				access := node.AsPropertyAccessExpression()
				if access.Expression == nil || access.Name() == nil || !contains(methods, access.Name().Text()) {
					return
				}
				if archkit.DeclaredUnder(ctx.TypeChecker, access.Expression, fragments) {
					ctx.ReportNode(node, buildDoubleLibraryMessage(access.Name().Text()))
				}
			},
		}
	}),
}
