package domain_probe_returns_void

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/tsgolint/internal/archrule"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/archtypes"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

const ruleName = "domain-probe-returns-void"

var defaultFiles = []string{"**/hexagon/application/**/ports/**"}
var defaultProbeNamePatterns = []string{"(?:Instrumentation|Probe)$"}

type Options struct {
	ProbeNamePatterns []string `json:"probeNamePatterns,omitempty"`
}

func buildProbeMemberReturnsValueMessage(port string, member string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "probeMemberReturnsValue",
		Description: "`" + port + "." + member +
			"` returns an answer. A Domain Probe announces a fact fire-and-forget, so its method must return `void`.",
		Help: "Declare the method as returning `void`; a probe announces and does not wait, so an awaited acknowledgement is a different collaboration.",
	}
}

func compile(patterns []string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		if expression, err := regexp.Compile(pattern); err == nil {
			compiled = append(compiled, expression)
		}
	}
	return compiled
}

var DomainProbeReturnsVoidRule = archrule.Rule{Files: defaultFiles, Rule: rule.Rule{
	Name: ruleName,
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[Options](options, ruleName)
		patterns := opts.ProbeNamePatterns
		if len(patterns) == 0 {
			patterns = defaultProbeNamePatterns
		}
		probeNames := compile(patterns)

		judge := func(node *ast.Node) {
			if !ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) {
				return
			}
			name := node.Name()
			if name == nil {
				return
			}
			port := name.Text()
			for _, pattern := range probeNames {
				if pattern.MatchString(port) {
					for _, member := range archtypes.Members(ctx.TypeChecker, archtypes.DeclaredType(ctx.TypeChecker, node)) {
						for _, signature := range archtypes.CallSignatures(ctx.TypeChecker, member.Type) {
							if !archtypes.IsVoidLike(archtypes.ReturnType(ctx.TypeChecker, signature)) {
								ctx.ReportNode(node, buildProbeMemberReturnsValueMessage(port, member.Name))
								return
							}
						}
					}
					return
				}
			}
		}

		return rule.RuleListeners{
			ast.KindInterfaceDeclaration: judge,
			ast.KindTypeAliasDeclaration: judge,
		}
	},
}}
