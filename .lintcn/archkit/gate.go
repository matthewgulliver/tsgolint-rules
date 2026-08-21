package archkit

import "github.com/typescript-eslint/tsgolint/internal/rule"

func Gated(files []string, run func(ctx rule.RuleContext, options any) rule.RuleListeners) func(ctx rule.RuleContext, options any) rule.RuleListeners {
	return func(ctx rule.RuleContext, options any) rule.RuleListeners {
		if ctx.SourceFile == nil || !Includes(files, ctx.SourceFile.FileName()) {
			return nil
		}
		return run(ctx, options)
	}
}
