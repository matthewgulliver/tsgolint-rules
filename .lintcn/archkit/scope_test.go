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

func TestOrigin(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                                 string
		file                                 string
		standardLibrary, packageDep, outside bool
	}{
		{"the TypeScript standard library ships inside node_modules, and is not a dependency the repository chose",
			"/repo/node_modules/typescript/lib/lib.es5.d.ts", true, false, true},
		{"a pnpm store path reaches the same lib",
			"/repo/node_modules/.pnpm/typescript@5.9.2/node_modules/typescript/lib/lib.dom.d.ts", true, false, true},
		{"a real dependency is not the standard library",
			"/repo/node_modules/zod/index.d.ts", false, true, true},
		{"a scoped dependency is a dependency",
			"/repo/node_modules/@lumi/shared/index.d.ts", false, true, true},
		{"repository source is neither",
			"/repo/packages/gifting/hexagon/domain/src/occasion.ts", false, false, false},
		{"windows separators are normalised",
			`C:\repo\node_modules\typescript\lib\lib.es2015.d.ts`, true, false, true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := IsStandardLibrary(testCase.file); got != testCase.standardLibrary {
				t.Errorf("IsStandardLibrary(%q) = %v, want %v", testCase.file, got, testCase.standardLibrary)
			}
			if got := IsPackageDependency(testCase.file); got != testCase.packageDep {
				t.Errorf("IsPackageDependency(%q) = %v, want %v", testCase.file, got, testCase.packageDep)
			}
			if got := IsOutsideDependency(testCase.file); got != testCase.outside {
				t.Errorf("IsOutsideDependency(%q) = %v, want %v", testCase.file, got, testCase.outside)
			}
		})
	}
}

func TestDeclaredByPackageDependency(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		files []string
		want  bool
	}{
		{"a platform global that @types/node augments is not a collaborator a package supplies",
			[]string{"/lib.es5.d.ts", "/repo/node_modules/@types/node/globals.d.ts"}, false},
		{"the same holds when the standard library itself sits under node_modules",
			[]string{"/repo/node_modules/typescript/lib/lib.es5.d.ts", "/repo/node_modules/@types/node/globals.d.ts"}, false},
		{"a genuine third-party class is a collaborator",
			[]string{"/repo/node_modules/pg/lib/client.d.ts"}, true},
		{"a repository type is not",
			[]string{"/repo/packages/gifting/hexagon/domain/src/occasion.ts"}, false},
		{"a type the repository declares and a package augments is still the repository's",
			[]string{"/repo/packages/gifting/hexagon/domain/src/occasion.ts", "/lib.es5.d.ts"}, false},
		{"nothing declared is nothing to report",
			nil, false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := DeclaredByPackageDependency(testCase.files); got != testCase.want {
				t.Errorf("DeclaredByPackageDependency(%v) = %v, want %v", testCase.files, got, testCase.want)
			}
		})
	}
}
