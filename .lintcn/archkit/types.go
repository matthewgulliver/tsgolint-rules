package archkit

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func DeclaringFiles(t *checker.Type) []string {
	if t == nil {
		return nil
	}
	symbol := checker.Type_symbol(t)
	if symbol == nil {
		return nil
	}
	return DeclaringFilesOfSymbol(symbol)
}

func DeclaringFilesOfSymbol(symbol *ast.Symbol) []string {
	if symbol == nil {
		return nil
	}
	files := make([]string, 0, len(symbol.Declarations))
	for _, declaration := range symbol.Declarations {
		sourceFile := ast.GetSourceFileOfNode(declaration)
		if sourceFile != nil {
			files = append(files, sourceFile.FileName())
		}
	}
	return files
}

func DeclaredUnder(c *checker.Checker, expression *ast.Node, fragments []string) bool {
	if expression == nil {
		return false
	}
	symbol := c.GetSymbolAtLocation(expression)
	if symbol == nil {
		return false
	}
	symbol = checker.SkipAlias(symbol, c)
	for _, file := range DeclaringFilesOfSymbol(symbol) {
		normalized := strings.ReplaceAll(file, "\\", "/")
		for _, fragment := range fragments {
			if strings.Contains(normalized, fragment) {
				return true
			}
		}
	}
	return false
}

func Constituents(t *checker.Type) []*checker.Type {
	if t == nil {
		return nil
	}
	if utils.IsUnionType(t) {
		return t.Types()
	}
	return []*checker.Type{t}
}

func ElementTypes(c *checker.Checker, t *checker.Type) []*checker.Type {
	if t == nil || !checker.Checker_isArrayOrTupleType(c, t) {
		return nil
	}
	return checker.Checker_getTypeArguments(c, t)
}

func Unwrapped(c *checker.Checker, t *checker.Type) *checker.Type {
	if t == nil {
		return nil
	}
	if awaited := checker.Checker_getAwaitedType(c, t); awaited != nil {
		return awaited
	}
	return t
}

func CallSignatures(c *checker.Checker, t *checker.Type) []*checker.Signature {
	if t == nil {
		return nil
	}
	return checker.Checker_getSignaturesOfType(c, t, checker.SignatureKindCall)
}

func IsCallable(c *checker.Checker, t *checker.Type) bool {
	return len(CallSignatures(c, t)) > 0
}

func Members(c *checker.Checker, t *checker.Type) []Member {
	if t == nil {
		return nil
	}
	properties := checker.Checker_getPropertiesOfType(c, t)
	members := make([]Member, 0, len(properties))
	for _, property := range properties {
		members = append(members, Member{
			Name: property.Name,
			Type: checker.Checker_getTypeOfSymbol(c, property),
		})
	}
	return members
}

type Member struct {
	Name string
	Type *checker.Type
}

func IsVoidLike(t *checker.Type) bool {
	return utils.IsTypeFlagSet(t, checker.TypeFlagsVoid|checker.TypeFlagsUndefined)
}

func DeclaredType(c *checker.Checker, declaration *ast.Node) *checker.Type {
	name := declaration.Name()
	if name == nil {
		return nil
	}
	symbol := c.GetSymbolAtLocation(name)
	if symbol == nil {
		return nil
	}
	return checker.Checker_getDeclaredTypeOfSymbol(c, symbol)
}

func TypeReferenceNames(annotation *ast.Node) []*ast.Node {
	if annotation == nil {
		return nil
	}
	names := make([]*ast.Node, 0, 4)
	var walk func(node *ast.Node) bool
	walk = func(node *ast.Node) bool {
		if node.Kind == ast.KindTypeReference {
			if name := declaredName(node.AsTypeReferenceNode().TypeName); name != nil {
				names = append(names, name)
			}
		}
		node.ForEachChild(walk)
		return false
	}
	walk(annotation)
	return names
}

func declaredName(name *ast.Node) *ast.Node {
	for name != nil && name.Kind == ast.KindQualifiedName {
		name = name.AsQualifiedName().Right
	}
	return name
}

func WrittenName(reference *ast.Node) string {
	if reference == nil {
		return ""
	}
	switch reference.Kind {
	case ast.KindIdentifier:
		return reference.Text()
	case ast.KindQualifiedName:
		return WrittenName(reference.AsQualifiedName().Right)
	case ast.KindPropertyAccessExpression:
		return WrittenName(reference.AsPropertyAccessExpression().Name())
	default:
		return ""
	}
}

func ReturnType(c *checker.Checker, signature *checker.Signature) *checker.Type {
	if signature == nil {
		return nil
	}
	return checker.Checker_getReturnTypeOfSignature(c, signature)
}
