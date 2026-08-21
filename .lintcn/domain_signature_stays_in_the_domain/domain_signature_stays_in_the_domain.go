// lintcn:name domain-signature-stays-in-the-domain
// lintcn:severity error
// lintcn:description Disallow domain signatures from naming types the domain does not own

package domain_signature_stays_in_the_domain

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
	"github.com/typescript-eslint/tsgolint/lintcn-rules/archkit"
)

var defaultFiles = []string{"**/hexagon/domain/**"}

// The trees a domain signature may name. The shared kernel is here because
// `Money` is the domain's own vocabulary, published one level up.
var defaultOwnFiles = []string{"**/hexagon/domain/**", "**/shared-kernel/**"}

type Options struct {
	OwnFiles []string `json:"ownFiles,omitempty"`
}

func (o Options) ownPatterns() []string {
	if len(o.OwnFiles) == 0 {
		return defaultOwnFiles
	}
	return o.OwnFiles
}

func buildForeignParameterTypeMessage(parameter string, written string, source string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "foreignParameterType",
		Description: "Parameter `" + parameter + "` is declared as `" + written +
			"`, which is declared in " + source +
			", so the decision can only be made by a caller holding a type the model does not own.",
		Help: "Take the fact as a domain value or a shared-kernel value, and let the caller translate before it asks for the decision.",
	}
}

func buildForeignReturnTypeMessage(name string, written string, source string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "foreignReturnType",
		Description: "`" + name + "` returns `" + written + "`, which is declared in " + source +
			", so the model's answer is expressed in someone else's vocabulary.",
		Help: "Return a domain value or an explicit business refusal the model declares, and let the caller map it onward.",
	}
}

// where names the tree a foreign declaration came from, so the diagnostic says
// which boundary was crossed rather than only that one was.
func where(file string) string {
	if archkit.IsPackageDependency(file) {
		return "a package dependency"
	}
	return "`" + file + "`"
}

// reportable decides whether one declaration of a written type name puts that
// name outside the domain's own vocabulary.
//
// A type alias declared by a package dependency is deliberately excused, on the
// precedent `no-provider-type-in-signature` already records: `z.infer<typeof
// Schema>` is declared in `node_modules` and describes a shape this repository
// owns, and flagging it would make every schema-derived domain type a
// violation. A class or interface from a dependency is nominal and is not
// excused. Everything this repository itself declares outside the allowed trees
// is reportable whichever way it is spelled — an application port is a port
// whether it is written `interface` or `type`.
func reportable(declaration *ast.Node, own []string) bool {
	sourceFile := ast.GetSourceFileOfNode(declaration)
	if sourceFile == nil {
		return false
	}
	file := sourceFile.FileName()

	// The language itself is not foreign: `aggregate-root.md` holds
	// `readonly pledgedAt: Date`, and without this the rule flags every domain
	// type in the repository.
	if archkit.IsStandardLibrary(file) {
		return false
	}
	if archkit.Includes(own, file) {
		return false
	}
	if archkit.IsPackageDependency(file) {
		return declaration.Kind == ast.KindClassDeclaration ||
			declaration.Kind == ast.KindInterfaceDeclaration
	}
	return true
}

var DomainSignatureStaysInTheDomainRule = rule.Rule{
	// Inline literal: lintcn's discovery matches `Name: "..."` in the source to
	// bind the CLI name to this rule, so a const would silently desynchronize
	// from lintcn:name above.
	Name: "domain-signature-stays-in-the-domain",
	Run: archkit.Gated(defaultFiles, func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[Options](options, "domain-signature-stays-in-the-domain")
		own := opts.ownPatterns()

		// foreign returns the first declaring file that puts a written type
		// name outside the domain's vocabulary, and whether one was found.
		foreign := func(written *ast.Node) (string, bool) {
			symbol := ctx.TypeChecker.GetSymbolAtLocation(written)
			if symbol == nil {
				return "", false
			}
			symbol = checker.SkipAlias(symbol, ctx.TypeChecker)

			for _, declaration := range symbol.Declarations {
				if reportable(declaration, own) {
					return ast.GetSourceFileOfNode(declaration).FileName(), true
				}
			}
			return "", false
		}

		judge := func(signature *ast.Node, name *ast.Node) {
			for _, parameter := range signature.Parameters() {
				declared := parameter.AsParameterDeclaration()
				if declared.Type == nil || declared.Name() == nil {
					continue
				}
				for _, written := range archkit.TypeReferenceNames(declared.Type) {
					if file, found := foreign(written); found {
						ctx.ReportNode(parameter, buildForeignParameterTypeMessage(
							declared.Name().Text(), written.Text(), where(file)))
						break
					}
				}
			}

			returned := signature.Type()
			if returned == nil || name == nil {
				return
			}
			for _, written := range archkit.TypeReferenceNames(returned) {
				if file, found := foreign(written); found {
					ctx.ReportNode(name, buildForeignReturnTypeMessage(
						name.Text(), written.Text(), where(file)))
					return
				}
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				if !ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) {
					return
				}
				judge(node, node.Name())
			},
			ast.KindVariableDeclaration: func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()
				if declaration.Initializer == nil || !ast.IsFunctionLike(declaration.Initializer) {
					return
				}
				if !ast.HasSyntacticModifier(exported(node), ast.ModifierFlagsExport) {
					return
				}
				judge(declaration.Initializer, declaration.Name())
			},
		}
	}),
}

// A `const` carries its `export` on the statement, two nodes up.
func exported(node *ast.Node) *ast.Node {
	list := node.Parent
	if list == nil || list.Parent == nil {
		return node
	}
	return list.Parent
}
