// lintcn:name no-provider-type-in-signature
// lintcn:severity error
// lintcn:description Disallow signatures in the hexagon from naming types a package dependency or the transport owns

package no_provider_type_in_signature

import (
	"regexp"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
	"github.com/typescript-eslint/tsgolint/lintcn-rules/archkit"
)

var defaultFiles = []string{
	"**/hexagon/application/**",
	"**/hexagon/domain/**",
}

type Options struct {
	PrincipalParameterPatterns []string `json:"principalParameterPatterns,omitempty"`
	TransportTypeNamePatterns  []string `json:"transportTypeNamePatterns,omitempty"`
}

var defaultPrincipalParameterPatterns = []string{"^principal$", "^actor$", "^currentUser$"}

var defaultTransportTypeNamePatterns = []string{"^(?:Request|Response|Headers|FormData)$"}

func (o Options) principalPatterns() []string {
	if len(o.PrincipalParameterPatterns) == 0 {
		return defaultPrincipalParameterPatterns
	}
	return o.PrincipalParameterPatterns
}

func (o Options) transportPatterns() []string {
	if len(o.TransportTypeNamePatterns) == 0 {
		return defaultTransportTypeNamePatterns
	}
	return o.TransportTypeNamePatterns
}

func buildProviderParameterTypeMessage(subject string, name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "providerParameterType",
		Description: subject + " is declared as `" + name +
			"`, a class or interface owned by a package dependency, so the inside of the hexagon can only be called by code holding that vendor's object.",
		Help: "Declare the parameter as a port or a domain type this repository owns, and let the adapter translate the vendor's object into it.",
	}
}

func buildProviderReturnTypeMessage(subject string, name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "providerReturnType",
		Description: subject + " is declared as `" + name +
			"`, a class or interface owned by a package dependency, so every consumer of the inside now holds the vendor's object and follows the vendor when it changes.",
		Help: "Return a domain type or port result this repository owns, and translate at the adapter.",
	}
}

func buildTransportTypeMessage(subject string, name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "transportTypeInSignature",
		Description: subject + " is declared as `" + name +
			"`, the platform's transport type, so the inside of the hexagon now depends on one transport and cannot be driven by a queue, a CLI or a test without it.",
		Help: "Take a typed application command and return a domain result, and let the driving adapter read the request.",
	}
}

func buildUnbrandedPrincipalMessage(parameter string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "unbrandedPrincipal",
		Description: "Parameter `" + parameter +
			"` carries identity but its type has no brand, so anything holding the ids can assemble one and the inside cannot tell an authenticated caller from an invented one.",
		Help: "Type it as the principal branded with a non-exported `unique symbol`, and mint it only in the module that authenticates.",
	}
}

func isBranded(c *checker.Checker, t *checker.Type) bool {
	if t == nil {
		return false
	}
	for _, property := range checker.Checker_getPropertiesOfType(c, t) {
		for _, declaration := range property.Declarations {
			name := declaration.Name()
			if name == nil || name.Kind != ast.KindComputedPropertyName {
				continue
			}
			key := name.AsComputedPropertyName().Expression
			if !utils.IsTypeFlagSet(c.GetTypeAtLocation(key), checker.TypeFlagsUniqueESSymbol) {
				continue
			}
			for _, file := range archkit.DeclaringFilesOfSymbol(c.GetSymbolAtLocation(key)) {
				if !archkit.IsStandardLibrary(file) && !archkit.IsPackageDependency(file) {
					return true
				}
			}
		}
	}
	return false
}

func reportable(declaration *ast.Node) bool {
	switch declaration.Kind {
	case ast.KindClassDeclaration, ast.KindInterfaceDeclaration:
		sourceFile := ast.GetSourceFileOfNode(declaration)
		return sourceFile != nil && archkit.IsPackageDependency(sourceFile.FileName())
	default:
		return false
	}
}

func notOurs(declaration *ast.Node) bool {
	sourceFile := ast.GetSourceFileOfNode(declaration)
	if sourceFile == nil {
		return false
	}
	file := sourceFile.FileName()
	return archkit.IsStandardLibrary(file) || archkit.IsPackageDependency(file)
}

func compileAll(patterns []string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		if expression, err := regexp.Compile(pattern); err == nil {
			compiled = append(compiled, expression)
		}
	}
	return compiled
}

func matchesAny(patterns []*regexp.Regexp, name string) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(name) {
			return true
		}
	}
	return false
}

type finding struct {
	name      string
	transport bool
}

var NoProviderTypeInSignatureRule = rule.Rule{
	Name: "no-provider-type-in-signature",
	Run: archkit.Gated(defaultFiles, func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[Options](options, "no-provider-type-in-signature")
		principals := compileAll(opts.principalPatterns())
		transports := compileAll(opts.transportPatterns())

		judge := func(annotation *ast.Node) (finding, bool) {
			for _, written := range archkit.TypeReferenceNames(annotation) {
				symbol := ctx.TypeChecker.GetSymbolAtLocation(written)
				if symbol == nil {
					continue
				}
				symbol = checker.SkipAlias(symbol, ctx.TypeChecker)

				if matchesAny(transports, written.Text()) {
					if slices.ContainsFunc(symbol.Declarations, notOurs) {
						return finding{name: symbol.Name, transport: true}, true
					}
				}
				if slices.ContainsFunc(symbol.Declarations, reportable) {
					return finding{name: symbol.Name}, true
				}
			}
			return finding{}, false
		}

		reportAnnotation := func(node *ast.Node, annotation *ast.Node, subject string) {
			if annotation == nil {
				return
			}
			found, ok := judge(annotation)
			if !ok {
				return
			}
			if found.transport {
				ctx.ReportNode(node, buildTransportTypeMessage(subject, found.name))
				return
			}
			ctx.ReportNode(node, buildProviderReturnTypeMessage(subject, found.name))
		}

		returnOf := func(node *ast.Node) {
			subject := "The return type"
			if name := node.Name(); name != nil && name.Kind == ast.KindIdentifier {
				subject = "The return type of `" + name.Text() + "`"
			}
			reportAnnotation(node.Type(), node.Type(), subject)
		}

		return rule.RuleListeners{
			ast.KindParameter: func(node *ast.Node) {
				parameter := node.AsParameterDeclaration()
				if parameter.Type == nil || parameter.Name() == nil {
					return
				}
				name := ""
				subject := "A destructured parameter"
				if given := parameter.Name(); given.Kind == ast.KindIdentifier {
					name = given.Text()
					subject = "Parameter `" + name + "`"
				}

				if found, ok := judge(parameter.Type); ok {
					if found.transport {
						ctx.ReportNode(node, buildTransportTypeMessage(subject, found.name))
					} else {
						ctx.ReportNode(node, buildProviderParameterTypeMessage(subject, found.name))
					}
					return
				}

				if name == "" || !matchesAny(principals, name) {
					return
				}
				declared := checker.Checker_getTypeFromTypeNode(ctx.TypeChecker, parameter.Type)
				if !isBranded(ctx.TypeChecker, declared) {
					ctx.ReportNode(node, buildUnbrandedPrincipalMessage(name))
				}
			},
			ast.KindFunctionDeclaration: returnOf,
			ast.KindFunctionExpression:  returnOf,
			ast.KindArrowFunction:       returnOf,
			ast.KindMethodDeclaration:   returnOf,
			ast.KindMethodSignature:     returnOf,
			ast.KindFunctionType:        returnOf,
			ast.KindPropertySignature: func(node *ast.Node) {
				if node.Type() != nil && node.Type().Kind == ast.KindFunctionType {
					return
				}
				subject := "The member"
				if name := node.Name(); name != nil && name.Kind == ast.KindIdentifier {
					subject = "Member `" + name.Text() + "`"
				}
				reportAnnotation(node, node.Type(), subject)
			},
		}
	}),
}
