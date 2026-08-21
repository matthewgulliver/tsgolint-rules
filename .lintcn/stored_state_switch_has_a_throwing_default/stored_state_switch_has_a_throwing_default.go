// lintcn:name stored-state-switch-has-a-throwing-default
// lintcn:severity error
// lintcn:description Require a throwing `default` on driven-adapter switches deciding on a property of a persistence row type — the row type is a claim about the database, not a fact about it

package stored_state_switch_has_a_throwing_default

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
	"github.com/typescript-eslint/tsgolint/lintcn-rules/archkit"
)

var defaultFiles = []string{"**/adapters/driven/**"}

// Where a persistence row type is declared. The row is the adapter's own
// vocabulary — `no-row-type-in-domain` is the rule that keeps it there — so a
// discriminant reaching one is a discriminant read out of the database.
var defaultRowTypeFiles = []string{"**/adapters/driven/**"}

type Options struct {
	RowTypeFiles []string `json:"rowTypeFiles,omitempty"`
}

func (o Options) rowTypePatterns() []string {
	if len(o.RowTypeFiles) == 0 {
		return defaultRowTypeFiles
	}
	return o.RowTypeFiles
}

func buildStoredStateSwitchWithoutDefaultMessage(discriminant string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "storedStateSwitchWithoutDefault",
		Description: "This `switch` decides on `" + discriminant +
			"`, which is read from a stored row, and has no `default`. Exhaustiveness over the row type proves nothing about the database, which may already hold a value the type denies.",
		Help: "Add a `default` that binds the discriminant to `never` and throws, so an unrecognised stored value fails loudly instead of returning `undefined` into the model.",
	}
}

func buildStoredStateDefaultDoesNotThrowMessage(discriminant string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "storedStateDefaultDoesNotThrow",
		Description: "This `switch` decides on `" + discriminant +
			"`, which is read from a stored row, and its `default` does not throw, so a row holding an unrecognised value is turned into a value the model accepts.",
		Help: "Throw from the `default`. A stored state the type denies is a corrupt row, not an outcome the model should be asked to represent.",
	}
}

// discriminantOfARow reports whether a switch is deciding on a property of a
// type a persistence adapter declares, and names the property.
//
// This is the whole reason the rule is type-aware. `switch (row.state)` and
// `switch (command.state)` are the same syntax; only the resolved type of the
// thing being read says which of them came out of the database.
func discriminantOfARow(ctx rule.RuleContext, expression *ast.Node, rowTypes []string) (string, bool) {
	if expression == nil || expression.Kind != ast.KindPropertyAccessExpression {
		return "", false
	}
	access := expression.AsPropertyAccessExpression()
	if access.Expression == nil || access.Name() == nil {
		return "", false
	}

	held := ctx.TypeChecker.GetTypeAtLocation(access.Expression)
	for _, constituent := range archkit.Constituents(held) {
		for _, file := range archkit.DeclaringFiles(constituent) {
			if archkit.Includes(rowTypes, file) {
				return access.Expression.Text() + "." + access.Name().Text(), true
			}
		}
	}
	return "", false
}

// throws reports whether a clause's statements reach a `throw`, including one
// inside the block a `never`-binding default is usually written as.
func throws(clause *ast.Node) bool {
	found := false
	var walk func(node *ast.Node) bool
	walk = func(node *ast.Node) bool {
		if found {
			return true
		}
		if node.Kind == ast.KindThrowStatement {
			found = true
			return true
		}
		// A nested function's `throw` is not this clause's.
		if ast.IsFunctionLike(node) {
			return false
		}
		node.ForEachChild(walk)
		return false
	}
	walk(clause)
	return found
}

var StoredStateSwitchHasAThrowingDefaultRule = rule.Rule{
	// Inline literal: lintcn's discovery matches `Name: "..."` in the source to
	// bind the CLI name to this rule, so a const would silently desynchronize
	// from lintcn:name above.
	Name: "stored-state-switch-has-a-throwing-default",
	Run: archkit.Gated(defaultFiles, func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[Options](options, "stored-state-switch-has-a-throwing-default")
		rowTypes := opts.rowTypePatterns()

		return rule.RuleListeners{
			ast.KindSwitchStatement: func(node *ast.Node) {
				statement := node.AsSwitchStatement()
				discriminant, stored := discriminantOfARow(ctx, statement.Expression, rowTypes)
				if !stored {
					return
				}

				for _, clause := range statement.CaseBlock.AsCaseBlock().Clauses.Nodes {
					if clause.Kind != ast.KindDefaultClause {
						continue
					}
					if !throws(clause) {
						ctx.ReportNode(clause, buildStoredStateDefaultDoesNotThrowMessage(discriminant))
					}
					return
				}

				ctx.ReportNode(statement.Expression, buildStoredStateSwitchWithoutDefaultMessage(discriminant))
			},
		}
	}),
}
