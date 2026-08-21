// lintcn:name no-outside-declaration-in-the-hexagon
// lintcn:severity error
// lintcn:description Disallow the inside of the hexagon from importing symbols declared in adapters, apps or composition

package no_outside_declaration_in_the_hexagon

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
	"github.com/typescript-eslint/tsgolint/lintcn-rules/archkit"
)

var defaultFiles = []string{"**/hexagon/application/**", "**/hexagon/domain/**"}

var defaultOutsideFiles = []string{"**/adapters/**", "**/apps/*/src/**", "**/composition/**"}

type Options struct {
	OutsideFiles []string `json:"outsideFiles,omitempty"`
}

func (o Options) outsidePatterns() []string {
	if len(o.OutsideFiles) == 0 {
		return defaultOutsideFiles
	}
	return o.OutsideFiles
}

func buildOutsideDeclarationImportedMessage(name string, file string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "outsideDeclarationImported",
		Description: "`" + name + "` is declared in `" + file +
			"`, outside the hexagon, so the inside now depends on an adapter, a host or the composition root and can no longer be driven without it — however the import was spelled.",
		Help: "Depend on a port declared inside the hexagon and let composition hand the adapter in.",
	}
}

func buildOutsideModuleImportedMessage(file string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "outsideModuleImported",
		Description: "This module resolves to `" + file +
			"`, outside the hexagon, so loading the inside now loads an adapter, a host or the composition root — however the specifier was spelled.",
		Help: "Import the port or domain module the inside owns, and let composition load the adapter.",
	}
}

var NoOutsideDeclarationInTheHexagonRule = rule.Rule{
	Name: "no-outside-declaration-in-the-hexagon",
	Run: archkit.Gated(defaultFiles, func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[Options](options, "no-outside-declaration-in-the-hexagon")
		outside := opts.outsidePatterns()

		outsideFile := func(symbol *ast.Symbol) (string, bool) {
			if symbol == nil {
				return "", false
			}
			symbol = checker.SkipAlias(symbol, ctx.TypeChecker)
			for _, file := range archkit.DeclaringFilesOfSymbol(symbol) {
				if archkit.IsPackageDependency(file) {
					continue
				}
				if archkit.Includes(outside, file) {
					return file, true
				}
			}
			return "", false
		}

		judgeBinding := func(node *ast.Node) {
			name := node.Name()
			if name == nil {
				return
			}
			if file, found := outsideFile(ctx.TypeChecker.GetSymbolAtLocation(name)); found {
				ctx.ReportNode(node, buildOutsideDeclarationImportedMessage(name.Text(), file))
			}
		}

		judgeModule := func(node *ast.Node, specifier *ast.Node) {
			if specifier == nil {
				return
			}
			if file, found := outsideFile(ctx.TypeChecker.GetSymbolAtLocation(specifier)); found {
				ctx.ReportNode(node, buildOutsideModuleImportedMessage(file))
			}
		}

		return rule.RuleListeners{
			ast.KindImportSpecifier: judgeBinding,
			ast.KindNamespaceImport: judgeBinding,
			ast.KindImportClause:    judgeBinding,
			ast.KindImportDeclaration: func(node *ast.Node) {
				declaration := node.AsImportDeclaration()
				if declaration.ImportClause != nil {
					return
				}
				judgeModule(node, declaration.ModuleSpecifier)
			},
			ast.KindExportDeclaration: func(node *ast.Node) {
				declaration := node.AsExportDeclaration()
				if declaration.ModuleSpecifier == nil {
					return
				}
				judgeModule(node, declaration.ModuleSpecifier)
			},
		}
	}),
}
