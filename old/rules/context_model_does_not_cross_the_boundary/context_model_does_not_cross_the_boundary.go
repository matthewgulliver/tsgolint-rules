package context_model_does_not_cross_the_boundary

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/archrule"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/archscope"
	"github.com/typescript-eslint/tsgolint/internal/archtypes"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

const ruleName = "context-model-does-not-cross-the-boundary"

var defaultFiles = []string{"**/packages/*/**"}

// A context's identity, captured by a group named `name` — the same decision
// `no-cross-context-internal-import` records: one context spelled two ways must
// capture the same text, or an intra-context reference reads as a crossing.
//
// Spelled without the lookahead its JavaScript sibling uses. Go's regexp is
// RE2 and has no lookaround; only the capture is read, so a trailing group
// costs nothing.
var defaultContextRootPatterns = []string{`(^|/)packages/(?<name>[^/]+)(/|$)`}

// A file another context may legitimately be reached through.
// A barrel is not a published contract: it is what a context ends up publishing
// by accident. dependency-cruiser forbids importing an index file at all, so it
// cannot also be the surface a second context is allowed to depend on.
var defaultPublicContractPathPatterns = []string{`/public\.tsx?$`, `/public/`}

// Trees that belong to no context and are shared on purpose.
var defaultSharedFiles = []string{"**/shared-kernel/**"}

type Options struct {
	ContextRootPatterns        []string `json:"contextRootPatterns,omitempty"`
	PublicContractPathPatterns []string `json:"publicContractPathPatterns,omitempty"`
	SharedFiles                []string `json:"sharedFiles,omitempty"`
}

func or(given []string, fallback []string) []string {
	if len(given) == 0 {
		return fallback
	}
	return given
}

func buildCrossContextModelTypeMessage(written string, context string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "crossContextModelType",
		Description: "`" + written + "` is declared inside the `" + context +
			"` context's own internals, and this context's exported signature names it, so the two models are now one and a change to theirs is a change to yours.",
		Help: "Translate it at an Anti-Corruption Layer into a type this context declares, or take it from that context's published contract, which is the surface it agreed to keep stable.",
	}
}

func matchesAny(patterns []*regexp.Regexp, file string) bool {
	for _, expression := range patterns {
		if expression.MatchString(file) {
			return true
		}
	}
	return false
}

var ContextModelDoesNotCrossTheBoundaryRule = archrule.Rule{Files: defaultFiles, Rule: rule.Rule{
	Name: ruleName,
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[Options](options, ruleName)
		roots := archscope.Compile(or(opts.ContextRootPatterns, defaultContextRootPatterns))
		published := archscope.Compile(or(opts.PublicContractPathPatterns, defaultPublicContractPathPatterns))
		shared := or(opts.SharedFiles, defaultSharedFiles)

		// A file no pattern identifies belongs to no context, and this rule has
		// nothing to compare. Returning early is the shipped precedent, and the
		// pages record that it makes a misconfigured pattern a green run.
		here, identified := archscope.ContextOf(roots, ctx.SourceFile.FileName())
		if !identified {
			return rule.RuleListeners{}
		}

		// crossing reports the first declaring file that puts a written type
		// name inside another context's internals.
		crossing := func(written *ast.Node) (string, bool) {
			symbol := ctx.TypeChecker.GetSymbolAtLocation(written)
			if symbol == nil {
				return "", false
			}
			symbol = checker.SkipAlias(symbol, ctx.TypeChecker)

			for _, declaration := range symbol.Declarations {
				sourceFile := ast.GetSourceFileOfNode(declaration)
				if sourceFile == nil {
					continue
				}
				file := sourceFile.FileName()

				if archscope.IsStandardLibrary(file) || archscope.IsPackageDependency(file) {
					continue
				}
				if archscope.Includes(shared, file) {
					continue
				}
				// Declared *in* another context's published contract is the
				// surface that context offered. Merely re-exported through one
				// is not: the declaration still lives in its internals, which
				// is the leak §2.3 names.
				if matchesAny(published, file) {
					continue
				}
				there, identified := archscope.ContextOf(roots, file)
				if !identified || there == here {
					continue
				}
				return there, true
			}
			return "", false
		}

		judge := func(annotation *ast.Node, at *ast.Node) {
			if annotation == nil || at == nil {
				return
			}
			for _, written := range archtypes.TypeReferenceNames(annotation) {
				if context, found := crossing(written); found {
					ctx.ReportNode(at, buildCrossContextModelTypeMessage(written.Text(), context))
					return
				}
			}
		}

		judgeSignature := func(signature *ast.Node, name *ast.Node) {
			for _, parameter := range signature.Parameters() {
				declared := parameter.AsParameterDeclaration()
				judge(declared.Type, parameter)
			}
			judge(signature.Type(), name)
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				if !ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) {
					return
				}
				judgeSignature(node, node.Name())
			},
			ast.KindVariableDeclaration: func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()
				if declaration.Initializer == nil || !ast.IsFunctionLike(declaration.Initializer) {
					return
				}
				if !ast.HasSyntacticModifier(exported(node), ast.ModifierFlagsExport) {
					return
				}
				judgeSignature(declaration.Initializer, declaration.Name())
			},
			// A published `interface` or `type` carries the model just as a
			// function signature does, and is how a context most often leaks.
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				if !ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) {
					return
				}
				judge(node, node.Name())
			},
			ast.KindTypeAliasDeclaration: func(node *ast.Node) {
				if !ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) {
					return
				}
				judge(node, node.Name())
			},
		}
	},
}}

// A `const` carries its `export` on the statement, two nodes up.
func exported(node *ast.Node) *ast.Node {
	list := node.Parent
	if list == nil || list.Parent == nil {
		return node
	}
	return list.Parent
}
