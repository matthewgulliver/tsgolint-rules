// lintcn:name use-case-result-is-discriminated
// lintcn:severity error
// lintcn:description Disallow a use-case result a caller cannot narrow on — a union of object shapes with no shared literal discriminant, or a failure member whose type admits any string

package use_case_result_is_discriminated

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
	"github.com/typescript-eslint/tsgolint/lintcn-rules/archkit"
)

var defaultFiles = []string{"**/hexagon/application/**", "**/hexagon/domain/**"}

type Options struct {
	FailureReasonMemberPatterns []string `json:"failureReasonMemberPatterns,omitempty"`
}

var defaultFailureReasonMemberPatterns = []string{"^error$", "^violations$", "^faults$", "^problems$"}

func (o Options) failureReasonPatterns() []string {
	if len(o.FailureReasonMemberPatterns) == 0 {
		return defaultFailureReasonMemberPatterns
	}
	return o.FailureReasonMemberPatterns
}

func buildUndiscriminatedResultMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "undiscriminatedResult",
		Description: "`" + name + "` returns a union of object shapes with no property that is a literal in every member, so a caller must guess which outcome it holds by probing for fields.",
		Help:        "Give every member the same discriminant property with a distinct literal value, so callers narrow on it and a missed outcome is a type error.",
	}
}

func buildGenericFailureReasonMessage(name string, member string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "genericFailureReason",
		Description: "`" + name + "` returns a `" + member + "` that admits any string, so the caller cannot switch on the outcome and the compiler cannot prove every failure is handled.",
		Help:        "Type it as a union of the specific reason literals this operation can produce.",
	}
}

func admitsAnyString(c *checker.Checker, t *checker.Type) bool {
	return admitsAnyStringWithin(c, t, 0)
}

func admitsAnyStringWithin(c *checker.Checker, t *checker.Type, depth int) bool {
	if depth > 4 {
		return false
	}
	for _, constituent := range archkit.Constituents(t) {
		if utils.IsTypeFlagSet(constituent, checker.TypeFlagsString) {
			return true
		}
		for _, element := range archkit.ElementTypes(c, constituent) {
			if admitsAnyStringWithin(c, element, depth+1) {
				return true
			}
		}
	}
	return false
}

func genericFailureReasons(c *checker.Checker, returned *checker.Type, patterns []*regexp.Regexp) []string {
	seen := map[string]bool{}
	names := make([]string, 0, 1)
	for _, shape := range archkit.Constituents(returned) {
		if !utils.IsTypeFlagSet(shape, checker.TypeFlagsObject) {
			continue
		}
		for _, member := range archkit.Members(c, shape) {
			if seen[member.Name] || !matchesAny(patterns, member.Name) {
				continue
			}
			if admitsAnyString(c, member.Type) {
				seen[member.Name] = true
				names = append(names, member.Name)
			}
		}
	}
	return names
}

func matchesAny(patterns []*regexp.Regexp, name string) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(name) {
			return true
		}
	}
	return false
}

func finalReturn(c *checker.Checker, t *checker.Type) *checker.Type {
	for range 8 {
		signatures := archkit.CallSignatures(c, t)
		if len(signatures) == 0 {
			return archkit.Unwrapped(c, t)
		}
		t = archkit.Unwrapped(c, archkit.ReturnType(c, signatures[len(signatures)-1]))
	}
	return t
}

func isLiteral(t *checker.Type) bool {
	return utils.IsTypeFlagSet(t, checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral|
		checker.TypeFlagsBooleanLiteral|checker.TypeFlagsUniqueESSymbol)
}

func outcomeShapes(t *checker.Type) []*checker.Type {
	shapes := make([]*checker.Type, 0, 4)
	for _, constituent := range archkit.Constituents(t) {
		if utils.IsTypeFlagSet(constituent, checker.TypeFlagsNullable) {
			continue
		}
		if !utils.IsTypeFlagSet(constituent, checker.TypeFlagsObject) {
			return nil
		}
		shapes = append(shapes, constituent)
	}
	return shapes
}

func discriminated(c *checker.Checker, shapes []*checker.Type) bool {
	for _, candidate := range archkit.Members(c, shapes[0]) {
		if !isLiteral(candidate.Type) {
			continue
		}
		shared := true
		for _, shape := range shapes[1:] {
			property := checker.Checker_getPropertyOfType(c, shape, candidate.Name)
			if property == nil || !isLiteral(checker.Checker_getTypeOfSymbol(c, property)) {
				shared = false
				break
			}
		}
		if shared {
			return true
		}
	}
	return false
}

var UseCaseResultIsDiscriminatedRule = rule.Rule{
	Name: "use-case-result-is-discriminated",
	Run: archkit.Gated(defaultFiles, func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[Options](options, "use-case-result-is-discriminated")
		reasons := archkit.Compile(opts.failureReasonPatterns())

		judge := func(node *ast.Node, name *ast.Node) {
			if name == nil || !ast.HasSyntacticModifier(exported(node), ast.ModifierFlagsExport) {
				return
			}

			returned := finalReturn(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(node))
			for _, member := range genericFailureReasons(ctx.TypeChecker, returned, reasons) {
				ctx.ReportNode(name, buildGenericFailureReasonMessage(name.Text(), member))
			}
			if !utils.IsUnionType(returned) {
				return
			}

			shapes := outcomeShapes(returned)
			if len(shapes) < 2 {
				return
			}
			if !discriminated(ctx.TypeChecker, shapes) {
				ctx.ReportNode(name, buildUndiscriminatedResultMessage(name.Text()))
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: func(node *ast.Node) { judge(node, node.Name()) },
			ast.KindVariableDeclaration: func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()
				if declaration.Initializer == nil || !ast.IsFunctionLike(declaration.Initializer) {
					return
				}
				judge(node, declaration.Name())
			},
		}
	}),
}

func exported(node *ast.Node) *ast.Node {
	if node.Kind != ast.KindVariableDeclaration {
		return node
	}
	list := node.Parent
	if list == nil || list.Parent == nil {
		return node
	}
	return list.Parent
}
