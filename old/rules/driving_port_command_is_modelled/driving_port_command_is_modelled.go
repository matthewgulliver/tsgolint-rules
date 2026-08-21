package driving_port_command_is_modelled

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/archrule"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/archtypes"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

const ruleName = "driving-port-command-is-modelled"

var defaultFiles = []string{"**/ports/**"}

// The driving doc prefers a `For…` intention name. The rule does not require
// one — it uses it only to find the declarations it judges.
var defaultDrivingPortPatterns = []string{"^For"}

type Options struct {
	DrivingPortPatterns []string `json:"drivingPortPatterns,omitempty"`
}

func buildUnmodelledCommandMessage(port string, member string, parameter string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "unmodelledCommand",
		Description: "`" + port + "." + member + "` takes `" + parameter +
			"` as a bare primitive, so every actor may pass any string of the right shape and the port carries no vocabulary of its own.",
		Help: "Take a declared command type, or a branded domain value, so the type says which value this is.",
	}
}

// A primitive with nothing added. A branded `ContributorId` is an
// intersection, and a literal union is a modelled vocabulary; neither is bare,
// and no syntactic rule can tell them from `string`.
func isBarePrimitive(t *checker.Type) bool {
	return utils.IsTypeFlagSet(t, checker.TypeFlagsString|checker.TypeFlagsNumber|
		checker.TypeFlagsBoolean|checker.TypeFlagsAny|checker.TypeFlagsUnknown) &&
		!utils.IsTypeFlagSet(t, checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral|
			checker.TypeFlagsBooleanLiteral)
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

var DrivingPortCommandIsModelledRule = archrule.Rule{Files: defaultFiles, Rule: rule.Rule{
	Name: ruleName,
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[Options](options, ruleName)
		patterns := opts.DrivingPortPatterns
		if len(patterns) == 0 {
			patterns = defaultDrivingPortPatterns
		}
		drivingPort := compile(patterns)

		judge := func(node *ast.Node) {
			name := node.Name()
			if name == nil {
				return
			}
			port := name.Text()
			named := false
			for _, pattern := range drivingPort {
				if pattern.MatchString(port) {
					named = true
					break
				}
			}
			if !named {
				return
			}

			for _, member := range archtypes.Members(ctx.TypeChecker, archtypes.DeclaredType(ctx.TypeChecker, node)) {
				for _, signature := range archtypes.CallSignatures(ctx.TypeChecker, member.Type) {
					for _, parameter := range checker.Signature_parameters(signature) {
						if isBarePrimitive(checker.Checker_getTypeOfSymbol(ctx.TypeChecker, parameter)) {
							ctx.ReportNode(node, buildUnmodelledCommandMessage(port, member.Name, parameter.Name))
							return
						}
					}
				}
			}
		}

		return rule.RuleListeners{
			ast.KindInterfaceDeclaration: judge,
			ast.KindTypeAliasDeclaration: judge,
		}
	},
}}
