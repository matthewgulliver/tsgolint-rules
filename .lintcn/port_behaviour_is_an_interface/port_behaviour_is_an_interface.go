// lintcn:name port-behaviour-is-an-interface
// lintcn:severity error
// lintcn:description Disallow declaring a port's behaviour as a type alias instead of an interface

package port_behaviour_is_an_interface

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
	"github.com/typescript-eslint/tsgolint/lintcn-rules/archkit"
)

var defaultFiles = []string{"**/ports/**"}

type Options struct {
	ExportedOnly *bool `json:"exportedOnly,omitempty"`
}

func buildBareFunctionTypeAliasMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "bareFunctionTypeAlias",
		Description: "`" + name + "` resolves to a bare function type, so the role has no name and an adapter holds a callable rather than a collaborator.",
		Help:        "Declare an interface whose method is the operation, so the actor's second operation is an added method rather than a second injected parameter.",
	}
}

func buildBehaviourContractAsTypeAliasMessage(name string, member string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "behaviourContractAsTypeAlias",
		Description: "`" + name + "." + member + "` resolves to a callable, so `" + name +
			"` is a behaviour contract an adapter implements, declared with the keyword this repository reserves for data shapes.",
		Help: "Declare `" + name + "` as an `interface`; keep `type` for the pure data shapes beside it.",
	}
}

// readableAsWritten reports whether `port-contract-is-an-interface` already
// judges this declaration. That rule reads a written `TSFunctionType` on an
// exported alias, so a callable spelled out in place is its diagnostic, not
// ours — both firing made one mistake produce two sentences saying the same
// thing. Everything it cannot see is still this rule's: an alias, an
// intersection, and any declaration it never opens.
func readableAsWritten(node *ast.Node, written *ast.Node) bool {
	if !ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) {
		return false
	}
	if written.Kind == ast.KindFunctionType {
		return true
	}
	if written.Kind != ast.KindTypeLiteral {
		return false
	}
	for _, member := range written.AsTypeLiteralNode().Members.Nodes {
		if member.Kind != ast.KindPropertySignature {
			continue
		}
		if annotation := member.AsPropertySignatureDeclaration().Type; annotation != nil && annotation.Kind == ast.KindFunctionType {
			return true
		}
	}
	return false
}

func isCallable(c *checker.Checker, t *checker.Type) bool {
	return len(checker.Checker_getSignaturesOfType(c, t, checker.SignatureKindCall)) > 0
}

var PortBehaviourIsAnInterfaceRule = rule.Rule{
	// Inline literal: lintcn's discovery matches `Name: "..."` in the source to
	// bind the CLI name to this rule, so a const would silently desynchronize
	// from lintcn:name above.
	Name: "port-behaviour-is-an-interface",
	Run: archkit.Gated(defaultFiles, func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[Options](options, "port-behaviour-is-an-interface")
		exportedOnly := opts.ExportedOnly == nil || *opts.ExportedOnly

		return rule.RuleListeners{
			ast.KindTypeAliasDeclaration: func(node *ast.Node) {
				declaration := node.AsTypeAliasDeclaration()
				name := declaration.Name()
				if name == nil || declaration.Type == nil {
					return
				}
				if exportedOnly && !ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) {
					return
				}
				if readableAsWritten(node, declaration.Type) {
					return
				}

				alias := name.Text()
				declared := checker.Checker_getTypeFromTypeNode(ctx.TypeChecker, declaration.Type)

				if isCallable(ctx.TypeChecker, declared) {
					ctx.ReportNode(node, buildBareFunctionTypeAliasMessage(alias))
					return
				}

				// Only a declared object or intersection carries members of its own.
				// A primitive's apparent type carries `String.prototype`, whose
				// methods are callable and none of this repository's business.
				if !utils.IsTypeFlagSet(declared, checker.TypeFlagsObject|checker.TypeFlagsIntersection) {
					return
				}

				// One report per alias: the first callable member is enough to
				// settle which keyword the declaration should have used.
				for _, member := range checker.Checker_getPropertiesOfType(ctx.TypeChecker, declared) {
					if isCallable(ctx.TypeChecker, checker.Checker_getTypeOfSymbol(ctx.TypeChecker, member)) {
						ctx.ReportNode(node, buildBehaviourContractAsTypeAliasMessage(alias, member.Name))
						return
					}
				}
			},
		}
	}),
}
