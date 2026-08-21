// lintcn:name context-model-does-not-cross-the-boundary
// lintcn:severity error
// lintcn:description Disallow an exported signature from naming a type declared inside another context's internals

package context_model_does_not_cross_the_boundary

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
	"github.com/typescript-eslint/tsgolint/lintcn-rules/archkit"
)

var defaultFiles = []string{"**/packages/*/**"}

var defaultContextRootPatterns = []string{`(^|/)packages/(?<name>[^/]+)(/|$)`}

var defaultPublicContractPathPatterns = []string{`/public\.tsx?$`, `/public/`}

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

var ContextModelDoesNotCrossTheBoundaryRule = rule.Rule{
	Name: "context-model-does-not-cross-the-boundary",
	Run: archkit.Gated(defaultFiles, func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[Options](options, "context-model-does-not-cross-the-boundary")
		roots := archkit.Compile(or(opts.ContextRootPatterns, defaultContextRootPatterns))
		published := archkit.Compile(or(opts.PublicContractPathPatterns, defaultPublicContractPathPatterns))
		shared := or(opts.SharedFiles, defaultSharedFiles)

		here, identified := archkit.ContextOf(roots, ctx.SourceFile.FileName())
		if !identified {
			return rule.RuleListeners{}
		}

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

				if archkit.IsStandardLibrary(file) || archkit.IsPackageDependency(file) {
					continue
				}
				if archkit.Includes(shared, file) {
					continue
				}
				if matchesAny(published, file) {
					continue
				}
				there, identified := archkit.ContextOf(roots, file)
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
			for _, written := range archkit.TypeReferenceNames(annotation) {
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
	}),
}

func exported(node *ast.Node) *ast.Node {
	list := node.Parent
	if list == nil || list.Parent == nil {
		return node
	}
	return list.Parent
}
