// lintcn:name published-contract-publishes-no-mutable-value
// lintcn:severity error
// lintcn:description Disallow a published contract from exposing a value any context can mutate

package published_contract_publishes_no_mutable_value

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
	"github.com/typescript-eslint/tsgolint/lintcn-rules/archkit"
)

var defaultFiles = []string{"**/public.ts", "**/public.tsx"}

var defaultMutatingMemberNames = []string{
	"set", "add", "delete", "clear",
	"push", "pop", "shift", "unshift", "splice", "sort", "reverse",
	"fill", "copyWithin",
}

type Options struct {
	MutatingMemberNames []string `json:"mutatingMemberNames,omitempty"`
}

func (o Options) mutatingMembers() []string {
	if len(o.MutatingMemberNames) == 0 {
		return defaultMutatingMemberNames
	}
	return o.MutatingMemberNames
}

func buildMutableMemberInContractMessage(name string, member string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "mutableMemberInContract",
		Description: "`" + name + "` is published with a writable `" + member +
			"`, so any context holding the contract can change what every other context reads.",
		Help: "Publish a value whose members are `readonly` — freezing it, or declaring the contract type with `readonly` on every member — and communicate change as an event.",
	}
}

func buildMutableContainerInContractMessage(name string, member string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "mutableContainerInContract",
		Description: "`" + name + "` is published as a standard-library container exposing `" + member +
			"`, so a consumer can write into state every other context reads.",
		Help: "Publish the readonly view of the container — `ReadonlyMap`, `ReadonlySet`, `ReadonlyArray` — or a function that answers a question instead of handing the collection over.",
	}
}

func languageDeclared(symbol *ast.Symbol) bool {
	files := archkit.DeclaringFilesOfSymbol(symbol)
	if len(files) == 0 {
		return false
	}
	for _, file := range files {
		if !archkit.IsStandardLibrary(file) {
			return false
		}
	}
	return true
}

var PublishedContractPublishesNoMutableValueRule = rule.Rule{
	Name: "published-contract-publishes-no-mutable-value",
	Run: archkit.Gated(defaultFiles, func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[Options](options, "published-contract-publishes-no-mutable-value")
		mutating := make(map[string]bool, len(opts.mutatingMembers()))
		for _, name := range opts.mutatingMembers() {
			mutating[name] = true
		}

		return rule.RuleListeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()
				name := declaration.Name()
				if name == nil {
					return
				}
				if !ast.HasSyntacticModifier(exported(node), ast.ModifierFlagsExport) {
					return
				}

				published := ctx.TypeChecker.GetTypeAtLocation(node)
				if published == nil {
					return
				}
				if archkit.IsCallable(ctx.TypeChecker, published) {
					return
				}

				properties := checker.Checker_getPropertiesOfType(ctx.TypeChecker, published)

				for _, property := range properties {
					if !mutating[property.Name] || !languageDeclared(property) {
						continue
					}
					if !archkit.IsCallable(ctx.TypeChecker, checker.Checker_getTypeOfSymbol(ctx.TypeChecker, property)) {
						continue
					}
					ctx.ReportNode(name, buildMutableContainerInContractMessage(
						name.Text(), property.Name))
					return
				}

				for _, property := range properties {
					if archkit.IsCallable(ctx.TypeChecker, checker.Checker_getTypeOfSymbol(ctx.TypeChecker, property)) {
						continue
					}
					if !checker.Checker_isReadonlySymbol(ctx.TypeChecker, property) {
						ctx.ReportNode(name, buildMutableMemberInContractMessage(
							name.Text(), property.Name))
						return
					}
				}
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
