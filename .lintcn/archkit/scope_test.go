package archkit

import "testing"

func TestIncludes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		patterns []string
		file     string
		want     bool
	}{
		{"a doublestar spans any number of segments",
			[]string{"**/hexagon/application/**"},
			"/repo/packages/gifting/hexagon/application/src/pledge-to-occasion.ts", true},
		{"a doublestar spans no segments at all",
			[]string{"**/ports/**"},
			"/ports/pledge-persistence.ts", true},
		{"a neighbouring tree is not judged",
			[]string{"**/hexagon/application/**"},
			"/repo/packages/gifting/hexagon/domain/src/occasion.ts", false},
		{"a segment glob stays inside its segment",
			[]string{"**/*.test.ts"},
			"/repo/src/pledge.test.ts", true},
		{"a segment glob does not cross a separator",
			[]string{"**/application/*.ts"},
			"/repo/application/ports/pledge.ts", false},
		{"any pattern matching is enough",
			[]string{"**/domain/**", "**/application/**"},
			"/repo/application/x.ts", true},
		{"no pattern matches nothing",
			nil, "/repo/application/x.ts", false},
		{"windows separators are normalised",
			[]string{"**/ports/**"},
			`C:\repo\ports\pledge.ts`, true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := Includes(testCase.patterns, testCase.file); got != testCase.want {
				t.Errorf("Includes(%v, %q) = %v, want %v",
					testCase.patterns, testCase.file, got, testCase.want)
			}
		})
	}
}

func TestOutsideDependency(t *testing.T) {
	t.Parallel()

	cases := []struct {
		file string
		want bool
	}{
		{"/repo/node_modules/pg/index.d.ts", true},
		{"/repo/node_modules/.pnpm/pg@8/node_modules/pg/index.d.ts", true},
		{"/usr/lib/typescript/lib.dom.d.ts", true},
		{"/usr/lib/typescript/lib.es2022.d.ts", true},
		{"/repo/packages/gifting/hexagon/domain/src/occasion.ts", false},
		{"/repo/src/node_modules_report.ts", false},
	}

	for _, testCase := range cases {
		if got := IsOutsideDependency(testCase.file); got != testCase.want {
			t.Errorf("IsOutsideDependency(%q) = %v, want %v", testCase.file, got, testCase.want)
		}
	}
}

func TestContextOf(t *testing.T) {
	t.Parallel()

	packagesRoot := []string{`(^|/)packages/(?<name>[^/]+)(/|$)`}

	cases := []struct {
		name       string
		patterns   []string
		file       string
		want       string
		identified bool
	}{
		{"a context is the segment the pattern captures",
			packagesRoot,
			"/repo/packages/gifting/src/occasions/occasion.ts", "gifting", true},
		{"the innermost match wins, so a checkout under its own `packages/` still sees the context",
			packagesRoot,
			"/home/me/packages/monorepo/packages/gifting/src/occasions/occasion.ts", "gifting", true},
		{"a path no pattern matches identifies no context",
			packagesRoot,
			"/repo/apps/web/src/page.ts", "", false},
		{"a pattern capturing no `name` identifies no context",
			[]string{`(^|/)packages/[^/]+`},
			"/repo/packages/gifting/src/occasions/occasion.ts", "", false},
		{"the first pattern that identifies the file wins",
			[]string{`(^|/)contexts/(?<name>[^/]+)`, `(^|/)packages/(?<name>[^/]+)`},
			"/repo/packages/gifting/src/occasions/occasion.ts", "gifting", true},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, identified := ContextOf(Compile(test.patterns), test.file)
			if got != test.want || identified != test.identified {
				t.Errorf("ContextOf(%q) = %q, %v; want %q, %v", test.file, got, identified, test.want, test.identified)
			}
		})
	}
}
