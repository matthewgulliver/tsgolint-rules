// Package archtypes holds the checker questions this repository's rules ask
// more than once.
//
// All of them come down to the same move: resolve a written thing to a type,
// resolve that type to the file that declares it, and judge the file. That is
// the fact a syntactic rule cannot reach, and it is why these rules are in Go.
package archtypes

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

// DeclaringFiles returns the files declaring a type's symbol. An anonymous
// type — an object literal type, a union — has no symbol and no declaration,
// and returns nothing rather than a guess.
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

// DeclaringFilesOfSymbol returns the files declaring a symbol.
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

// Constituents flattens a union into its members, and returns a non-union
// unchanged. A check that skips this passes `string | undefined`, which is the
// recorded way one of these rules stops enforcing anything.
func Constituents(t *checker.Type) []*checker.Type {
	if t == nil {
		return nil
	}
	if utils.IsUnionType(t) {
		return t.Types()
	}
	return []*checker.Type{t}
}

// ElementTypes returns the element types of an array or tuple, and nothing for
// any other type. `readonly string[]` is a reference type rather than a
// string-flagged one, so a rule reading only type flags finds no string in it.
func ElementTypes(c *checker.Checker, t *checker.Type) []*checker.Type {
	if t == nil || !checker.Checker_isArrayOrTupleType(c, t) {
		return nil
	}
	return checker.Checker_getTypeArguments(c, t)
}

// Unwrapped resolves `Promise<T>` to `T`, and leaves a synchronous type alone.
// Every port in this repository is async, so a rule that judges a return type
// without this judges `Promise` instead.
func Unwrapped(c *checker.Checker, t *checker.Type) *checker.Type {
	if t == nil {
		return nil
	}
	if awaited := checker.Checker_getAwaitedType(c, t); awaited != nil {
		return awaited
	}
	return t
}

// IsVoidLike reports whether a type carries no answer — the shape of a write.
func IsVoidLike(t *checker.Type) bool {
	return utils.IsTypeFlagSet(t, checker.TypeFlagsVoid|checker.TypeFlagsUndefined)
}

// CallSignatures returns a type's call signatures.
func CallSignatures(c *checker.Checker, t *checker.Type) []*checker.Signature {
	if t == nil {
		return nil
	}
	return checker.Checker_getSignaturesOfType(c, t, checker.SignatureKindCall)
}

// IsCallable reports whether a type can be called — a behaviour, not data.
func IsCallable(c *checker.Checker, t *checker.Type) bool {
	return len(CallSignatures(c, t)) > 0
}

// DeclaredType resolves an interface or type-alias declaration to the type it
// declares.
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

// Members returns a type's properties paired with their own types, which is
// how both `{ save: (o) => void }` and `interface P { save(o): void }` are read
// as the same contract.
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

// Member is one property of a declared contract.
type Member struct {
	Name string
	Type *checker.Type
}

// TypeReferenceNames yields every type name written inside a type annotation,
// including the arguments of a generic and the members of a type literal.
//
// A name written `z.ZodType` or `storage.inner.Handle` is a qualified name, not
// an identifier, and yielding it whole took every caller down: `Node.Text` has
// no case for one, and a rule that asks for the text of a type it was handed
// panicked the whole run rather than reporting. The rightmost identifier is the
// type's own name and resolves to the same symbol, which is what both questions
// a caller asks — what is it called, and where is it declared — are about.
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

// declaredName unwraps a qualified name to the identifier it ends in.
func declaredName(name *ast.Node) *ast.Node {
	for name != nil && name.Kind == ast.KindQualifiedName {
		name = name.AsQualifiedName().Right
	}
	return name
}

// WrittenName is what a reference to a type or a class is called where it is
// written: the identifier itself, or the last name of a `namespace.Thing`
// reached through a qualified name or a property access. `Node.Text` has a case
// for the identifier only and panics on the other two, so a rule that names the
// thing it reports asks here instead. Anything else — a construction over a
// call or an element access — has no written name, and answers "".
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

// ReturnType returns a signature's return type.
func ReturnType(c *checker.Checker, signature *checker.Signature) *checker.Type {
	if signature == nil {
		return nil
	}
	return checker.Checker_getReturnTypeOfSignature(c, signature)
}

// DeclaredUnder reports whether an expression's symbol, aliases followed, is
// declared in a file whose path contains one of the fragments — how a rule
// tells a library's `vi` from a local object that happens to share the name.
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
