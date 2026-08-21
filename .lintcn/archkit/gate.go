// Package archkit is the shared helper package for the lintcn rules in this
// repository. It replaces the old fork's internal packages `archscope`,
// `archtypes`, and `archrule`: `rule.Rule` has no Files field in lintcn's
// registration, so each rule gates its own listeners with Includes instead.
package archkit

import "github.com/typescript-eslint/tsgolint/internal/rule"

// Gated wraps a rule's Run so the rule only analyzes files its tree names.
//
// lintcn registers plain rule.Rule values with no per-rule file scope, so this
// is where the old `archrule.Rule{Files: ...}` scoping lives now: a rule whose
// file matches none of the patterns gets no listeners at all.
func Gated(files []string, run func(ctx rule.RuleContext, options any) rule.RuleListeners) func(ctx rule.RuleContext, options any) rule.RuleListeners {
	return func(ctx rule.RuleContext, options any) rule.RuleListeners {
		if !Includes(files, ctx.SourceFile.FileName()) {
			return nil
		}
		return run(ctx, options)
	}
}
