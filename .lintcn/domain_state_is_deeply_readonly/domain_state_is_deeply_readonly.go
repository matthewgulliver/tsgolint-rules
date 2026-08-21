// lintcn:name domain-state-is-deeply-readonly
// lintcn:severity error
// lintcn:description Disallow domain state whose members are writable deeper than the root guards

package domain_state_is_deeply_readonly

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
	"github.com/typescript-eslint/tsgolint/lintcn-rules/archkit"
)

var defaultFiles = []string{"**/hexagon/domain/**", "**/shared-kernel/**"}

var defaultMutableCollectionTypeNames = []string{"Array", "Map", "Set"}

var traversedContainerNames = map[string]bool{
	"Array": true, "ReadonlyArray": true,
	"Map": true, "ReadonlyMap": true,
	"Set": true, "ReadonlySet": true,
}

const maxDepth = 6

type Options struct {
	MutableCollectionTypeNames []string `json:"mutableCollectionTypeNames,omitempty"`
}

func (o Options) collectionNames() []string {
	if len(o.MutableCollectionTypeNames) == 0 {
		return defaultMutableCollectionTypeNames
	}
	return o.MutableCollectionTypeNames
}

func buildMutableMemberResolvedMessage(path string, through string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "mutableMemberResolved",
		Description: "`" + path + "` is writable once `" + through +
			"` is followed, so the state the root guards can be altered through a member the root never sees changing.",
		Help: "Mark the member `readonly` where it is declared, and return the next state from the transition instead.",
	}
}

func buildMutableCollectionResolvedMessage(path string, collection string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "mutableCollectionResolved",
		Description: "`" + path + "` resolves to a mutable `" + collection +
			"`, so a `readonly` binding still hands the caller something it can add to and remove from.",
		Help: "Replace the collection with `ReadonlyArray`, `ReadonlyMap` or `ReadonlySet` where the type is declared.",
	}
}

func owner(node *ast.Node) *ast.Node {
	for current := node; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindClassDeclaration:
			return current
		}
	}
	return nil
}

func isExported(declaration *ast.Node) bool {
	return declaration != nil && ast.HasSyntacticModifier(declaration, ast.ModifierFlagsExport)
}

func repositoryOwned(file string) bool {
	return !archkit.IsOutsideDependency(file)
}

func writtenCollection(node *ast.Node, names map[string]bool) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindArrayType:
		return true
	case ast.KindTypeOperator:
		if node.AsTypeOperatorNode().Operator == ast.KindReadonlyKeyword {
			return false
		}
		return writtenCollection(node.AsTypeOperatorNode().Type, names)
	case ast.KindParenthesizedType:
		return writtenCollection(node.AsParenthesizedTypeNode().Type, names)
	case ast.KindUnionType:
		for _, member := range node.AsUnionTypeNode().Types.Nodes {
			if writtenCollection(member, names) {
				return true
			}
		}
		return false
	case ast.KindTypeReference:
		name := node.AsTypeReferenceNode().TypeName
		return name != nil && name.Kind == ast.KindIdentifier && names[name.Text()]
	}
	return false
}

func annotationOf(declaration *ast.Node) *ast.Node {
	if declaration == nil {
		return nil
	}
	return declaration.Type()
}

func symbolName(t *checker.Type) string {
	symbol := checker.Type_symbol(t)
	if symbol == nil {
		return ""
	}
	return symbol.Name
}

var DomainStateIsDeeplyReadonlyRule = rule.Rule{
	Name: "domain-state-is-deeply-readonly",
	Run: archkit.Gated(defaultFiles, func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[Options](options, "domain-state-is-deeply-readonly")
		c := ctx.TypeChecker
		mutable := make(map[string]bool, len(opts.collectionNames()))
		for _, name := range opts.collectionNames() {
			mutable[name] = true
		}

		mutableCollectionIn := func(t *checker.Type) string {
			for _, constituent := range archkit.Constituents(t) {
				if name := symbolName(constituent); mutable[name] {
					return name
				}
			}
			return ""
		}

		judge := func(declaration *ast.Node, name *ast.Node) {
			declared := archkit.DeclaredType(c, declaration)
			if declared == nil {
				return
			}
			visited := map[*checker.Type]bool{}
			reported := map[string]bool{}

			var walk func(t *checker.Type, path string, through string, depth int)
			walk = func(t *checker.Type, path string, through string, depth int) {
				if t == nil || depth > maxDepth || visited[t] {
					return
				}
				visited[t] = true

				for _, constituent := range archkit.Constituents(t) {
					if traversedContainerNames[symbolName(constituent)] {
						for _, argument := range checker.Checker_getTypeArguments(c, constituent) {
							walk(argument, path+"[]", through, depth+1)
						}
						continue
					}
					if !utils.IsObjectType(constituent) && !utils.IsIntersectionType(constituent) {
						continue
					}
					for _, property := range checker.Checker_getPropertiesOfType(c, constituent) {
						if len(property.Declarations) == 0 {
							continue
						}
						propertyDeclaration := property.Declarations[0]
						file := ast.GetSourceFileOfNode(propertyDeclaration)
						if file == nil || !repositoryOwned(file.FileName()) {
							continue
						}
						propertyType := checker.Checker_getTypeOfSymbol(c, property)
						if archkit.IsCallable(c, propertyType) {
							continue
						}

						memberPath := path + "." + property.Name
						container := owner(propertyDeclaration)
						inside := container == declaration
						if !inside && isExported(container) {
							continue
						}

						if inside {
							if !writtenCollection(annotationOf(propertyDeclaration), mutable) {
								if collection := mutableCollectionIn(propertyType); collection != "" && !reported[memberPath] {
									reported[memberPath] = true
									ctx.ReportNode(name, buildMutableCollectionResolvedMessage(memberPath, collection))
								}
							}
						} else {
							if !checker.Checker_isReadonlySymbol(c, property) && !reported[memberPath] {
								reported[memberPath] = true
								ctx.ReportNode(name, buildMutableMemberResolvedMessage(memberPath, through))
							}
							if collection := mutableCollectionIn(propertyType); collection != "" && !reported[memberPath] {
								reported[memberPath] = true
								ctx.ReportNode(name, buildMutableCollectionResolvedMessage(memberPath, collection))
							}
						}

						nextThrough := through
						if names := archkit.TypeReferenceNames(annotationOf(propertyDeclaration)); len(names) > 0 {
							nextThrough = names[0].Text()
						}
						walk(propertyType, memberPath, nextThrough, depth+1)
					}
				}
			}

			walk(declared, name.Text(), name.Text(), 0)
		}

		return rule.RuleListeners{
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				if isExported(node) && node.Name() != nil {
					judge(node, node.Name())
				}
			},
			ast.KindTypeAliasDeclaration: func(node *ast.Node) {
				if isExported(node) && node.Name() != nil {
					judge(node, node.Name())
				}
			},
		}
	}),
}
