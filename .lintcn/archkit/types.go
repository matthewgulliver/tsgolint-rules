// Checker questions the lintcn rules in this repository share.
//
// All of them come down to the same move: resolve a written thing to a type,
// resolve that type to the file that declares it, and judge the file. That is
// the fact a syntactic rule cannot reach, and it is why these rules are in Go.
//
// Helpers are added when the first rule needs them, ported from the old
// fork's `internal/archtypes`, not kept speculatively.

package archkit

import (
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
	// The nil arms of this check and the sourceFile check below are exercised
	// by the rule packages' rule_tester cases (an anonymous object-literal
	// type has no symbol; a synthesized declaration has no source file), and
	// so are the checker-dependent arms of the helpers below (a non-array
	// type, an unawaited one). The TypeReferenceNames and declaredName walks
	// run on every rule_tester signature. Gremlins cannot see that coverage
	// cross-package, so the nine surviving NOT COVERED mutants in this file
	// are recorded as covered there; DeclaredType's two nil arms are
	// defensive and documented at the function.
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

// CallSignatures returns a type's call signatures.
func CallSignatures(c *checker.Checker, t *checker.Type) []*checker.Signature {
	if t == nil {
		return nil
	}
	return checker.Checker_getSignaturesOfType(c, t, checker.SignatureKindCall)
}

// IsCallable reports whether a type is a function — behaviour, which is not
// state and not worth following for its mutability.
func IsCallable(c *checker.Checker, t *checker.Type) bool {
	return len(CallSignatures(c, t)) > 0
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

// IsVoidLike reports whether a type carries no answer — the shape of a write.
func IsVoidLike(t *checker.Type) bool {
	return utils.IsTypeFlagSet(t, checker.TypeFlagsVoid|checker.TypeFlagsUndefined)
}

// DeclaredType resolves an interface or type-alias declaration to the type it
// declares. The nil arms are defensive: the consumers call it on
// interface/type-alias declarations whose names always resolve.
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

// TypeReferenceNames returns the written type names a type annotation
// mentions, qualified names reduced to their rightmost name. This is how a
// signature's vocabulary is read syntactically — before resolution decides
// whether each name is foreign.
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

// declaredName reduces a type name, which may be qualified like
// `some.Namespace.Type`, to the rightmost identifier the resolver knows.
func declaredName(name *ast.Node) *ast.Node {
	for name != nil && name.Kind == ast.KindQualifiedName {
		name = name.AsQualifiedName().Right
	}
	return name
}

// ReturnType returns a signature's return type.
func ReturnType(c *checker.Checker, signature *checker.Signature) *checker.Type {
	if signature == nil {
		return nil
	}
	return checker.Checker_getReturnTypeOfSignature(c, signature)
}
