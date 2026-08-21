// lintcn:name domain-type-is-declared-once
// lintcn:severity error
// lintcn:description Disallow declaring the same domain type twice in one context

package domain_type_is_declared_once

import (
	"regexp"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
	"github.com/typescript-eslint/tsgolint/lintcn-rules/archkit"
)

var defaultFiles = []string{"**/hexagon/domain/**"}

var defaultContextRootPatterns = []string{`(^|/)packages/(?<name>[^/]+)(/|$)`}

type Options struct {
	Files               []string `json:"files,omitempty"`
	ContextRootPatterns []string `json:"contextRootPatterns,omitempty"`
}

func (o Options) judged() []string {
	if len(o.Files) == 0 {
		return defaultFiles
	}
	return o.Files
}

func (o Options) contextRootPatterns() []string {
	if len(o.ContextRootPatterns) == 0 {
		return defaultContextRootPatterns
	}
	return o.ContextRootPatterns
}

func buildRedeclaredDomainTypeMessage(name string, elsewhere string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "redeclaredDomainType",
		Description: "`" + name + "` is also declared by `" + elsewhere +
			"` in this context, so there is no canonical `" + name +
			"` and two files can disagree about what the model is while both compile.",
		Help: "Declare it once in the file that owns the concept and import it everywhere else, or give the second one the name of the different concept it actually describes.",
	}
}

type index struct {
	byName map[string][]string
}

var indexes sync.Map

func indexOf(program *compiler.Program, judged []string) *index {
	if cached, ok := indexes.Load(program); ok {
		if idx, ok := cached.(*index); ok {
			return idx
		}
	}
	built := &index{byName: make(map[string][]string, 256)}
	for _, sourceFile := range program.SourceFiles() {
		file := sourceFile.FileName()
		if !archkit.Includes(judged, file) {
			continue
		}
		for _, name := range exportedTypeNames(sourceFile) {
			built.byName[name] = append(built.byName[name], file)
		}
	}
	actual, _ := indexes.LoadOrStore(program, built)
	if idx, ok := actual.(*index); ok {
		return idx
	}
	return built
}

func exportedTypeNames(sourceFile *ast.SourceFile) []string {
	if sourceFile.Statements == nil {
		return nil
	}
	names := make([]string, 0, 8)
	for _, statement := range sourceFile.Statements.Nodes {
		if statement.Kind != ast.KindInterfaceDeclaration &&
			statement.Kind != ast.KindTypeAliasDeclaration {
			continue
		}
		if !ast.HasSyntacticModifier(statement, ast.ModifierFlagsExport) {
			continue
		}
		if name := statement.Name(); name != nil {
			names = append(names, name.Text())
		}
	}
	return names
}

func contextOf(roots []*regexp.Regexp, file string) string {
	name, _ := archkit.ContextOf(roots, file)
	return name
}

var DomainTypeIsDeclaredOnceRule = rule.Rule{
	Name: "domain-type-is-declared-once",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[Options](options, "domain-type-is-declared-once")
		judged := opts.judged()
		if ctx.SourceFile == nil || !archkit.Includes(judged, ctx.SourceFile.FileName()) {
			return nil
		}
		here := ctx.SourceFile.FileName()

		roots := archkit.Compile(opts.contextRootPatterns())
		thisContext := contextOf(roots, here)
		declared := indexOf(ctx.Program, judged)

		judge := func(node *ast.Node) {
			if node.Parent != ctx.SourceFile.AsNode() {
				return
			}
			if !ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) {
				return
			}
			name := node.Name()
			if name == nil {
				return
			}

			for _, file := range declared.byName[name.Text()] {
				if file == here || contextOf(roots, file) != thisContext {
					continue
				}
				ctx.ReportNode(name, buildRedeclaredDomainTypeMessage(name.Text(), file))
				return
			}
		}

		return rule.RuleListeners{
			ast.KindInterfaceDeclaration: judge,
			ast.KindTypeAliasDeclaration: judge,
		}
	},
}
