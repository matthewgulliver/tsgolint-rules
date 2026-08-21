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
	// type has no symbol; a synthesized declaration has no source file).
	// Gremlins cannot see that coverage cross-package, so the two surviving
	// NOT COVERED mutants in this file are recorded as covered there.
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
