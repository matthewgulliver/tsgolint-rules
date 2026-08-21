package no_page_request_in_journey

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/tsgolint/internal/archrule"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/archtypes"
)

const ruleName = "no-page-request-in-journey"

var defaultFiles = []string{"**/e2e/**"}

func buildPageRequestMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "pageRequest",
		Description: "`page.request` uses Playwright's Node HTTP client, so this journey can stay green while the browser UI, cookie policy, and browser-generated request metadata are broken.",
		Help:        "Create prerequisite state through a fixture seam, then drive the journey through accessible page locators and observe its browser-caused request.",
	}
}

func isPlaywrightPage(ctx rule.RuleContext, expression *ast.Node) bool {
	for _, constituent := range archtypes.Constituents(ctx.TypeChecker.GetTypeAtLocation(expression)) {
		for _, file := range archtypes.DeclaringFiles(constituent) {
			normalized := strings.ReplaceAll(file, "\\", "/")
			if strings.Contains(normalized, "/node_modules/@playwright/") {
				return true
			}
		}
	}
	return false
}

var NoPageRequestInJourneyRule = archrule.Rule{Files: defaultFiles, Rule: rule.Rule{
	Name: ruleName,
	Run: func(ctx rule.RuleContext, _ any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				access := node.AsPropertyAccessExpression()
				if access.Expression == nil || access.Name() == nil || access.Name().Text() != "request" {
					return
				}
				if isPlaywrightPage(ctx, access.Expression) {
					ctx.ReportNode(node, buildPageRequestMessage())
				}
			},
		}
	},
}}
