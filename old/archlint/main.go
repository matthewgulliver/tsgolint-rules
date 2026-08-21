// Command archlint runs this repository's type-aware rules over a TypeScript
// project and fails when one of them reports.
//
// Upstream's `cmd/tsgolint` cannot do that job. It registers every rule
// compiled into the binary — typescript-eslint's fifty-nine alongside ours —
// and returns 0 however many diagnostics it printed. Pointed at this
// repository it found 580 errors, none of them ours, and exited green.
//
// Output mirrors `oxlint`'s, so a reader moving between the two halves reads
// one format: findings on stdout, one per line, and the summary on stderr.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/typescript-eslint/tsgolint/internal/archrule"
	"github.com/typescript-eslint/tsgolint/internal/diagnostic"
	"github.com/typescript-eslint/tsgolint/internal/linter"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
)

const usage = `archlint — this repository's type-aware architecture rules

Usage:
    archlint [OPTIONS]

Options:
    --tsconfig PATH   Which tsconfig to use. Defaults to tsconfig.json.
    --config PATH     Which rule configuration to use. Defaults to
                      .archtypesrc.json beside the tsconfig, if there is one.
    --list-rules      List every rule this binary can run and where it came
                      from, then exit
    -h, --help        Show help

Exit codes:
    0   no rule reported
    1   at least one rule reported
    2   the run could not be completed
`

const (
	exitClean    = 0
	exitReported = 1
	exitFailed   = 2
)

// namespace prefixes a rule name in the output, the way oxlint's `arch(...)`
// names the JS half. The two halves are different binaries and a reader who
// sees a name needs to know which one to go and edit.
const namespace = "arch-types"

func main() { os.Exit(run()) }

func run() int {
	var (
		tsconfig   string
		configPath string
		listRules  bool
		help       bool
	)
	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	flag.StringVar(&tsconfig, "tsconfig", "", "which tsconfig to use")
	flag.StringVar(&configPath, "config", "", "which rule configuration to use")
	flag.BoolVar(&listRules, "list-rules", false, "list every rule this binary can run")
	flag.BoolVar(&help, "help", false, "show help")
	flag.BoolVar(&help, "h", false, "show help")
	flag.Parse()

	if help {
		flag.Usage()
		return exitClean
	}

	if listRules {
		out := bufio.NewWriter(os.Stdout)
		defer out.Flush()
		for _, r := range archRules {
			fmt.Fprintf(out, "%s\trepository\n", r.Name)
		}
		for _, r := range upstreamRules {
			fmt.Fprintf(out, "%s\tupstream\n", r.Name)
		}
		return exitClean
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "archlint: cannot read the working directory: %v\n", err)
		return exitFailed
	}

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))

	configFileName := tspath.ResolvePath(tspath.NormalizePath(workingDirectory), orDefault(tsconfig, "tsconfig.json"))
	if !fs.FileExists(configFileName) {
		fmt.Fprintf(os.Stderr, "archlint: no tsconfig at %s\n", configFileName)
		return exitFailed
	}

	projectDirectory := tspath.GetDirectoryPath(configFileName)

	byName := map[string]rule.Rule{}
	registered := registry{}
	documented := map[string]bool{}
	for _, r := range archRules {
		byName[r.Name] = r.Rule
		registered[r.Name] = registration{onByDefault: true, files: r.Files}
		documented[r.Name] = true
	}
	for _, r := range upstreamRules {
		byName[r.Name] = r
		registered[r.Name] = registration{}
	}

	rules, err := loadConfig(fs, projectDirectory, configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "archlint: %v\n", err)
		return exitFailed
	}
	if err := rules.validate(registered); err != nil {
		fmt.Fprintf(os.Stderr, "archlint: %v\n", err)
		return exitFailed
	}

	host := utils.CreateCompilerHost(projectDirectory, fs)

	program, _, err := utils.CreateProgram(false, fs, projectDirectory, configFileName, host, false)
	if err != nil || program == nil {
		fmt.Fprintf(os.Stderr, "archlint: cannot build a TypeScript program from %s: %v\n", configFileName, err)
		return exitFailed
	}

	files := projectFiles(program, projectDirectory)

	relativeTo := tspath.ComparePathsOptions{
		CurrentDirectory:          host.GetCurrentDirectory(),
		UseCaseSensitiveFileNames: host.FS().UseCaseSensitiveFileNames(),
	}

	reported := 0
	diagnostics := make(chan rule.RuleDiagnostic, 4096)

	var printing sync.WaitGroup
	printing.Go(func() {
		out := bufio.NewWriterSize(os.Stdout, 64*1024)
		defer out.Flush()
		for d := range diagnostics {
			reported++
			writeDiagnostic(out, d, relativeTo, documented)
		}
	})

	err = linter.RunLinterOnProgram(linter.RunLinterOnProgramOptions{
		LogLevel:             utils.GetLogLevel(),
		Program:              program,
		Files:                files,
		Workers:              runtime.GOMAXPROCS(0),
		GetRulesForFile: func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			running := rules.resolve(sourceFile.FileName(), registered)
			names := make([]string, 0, len(running))
			for name := range running {
				names = append(names, name)
			}
			// Sorted, so two runs over the same tree print the same lines in
			// the same order and a diff of the output means something.
			slices.Sort(names)

			configured := make([]linter.ConfiguredRule, 0, len(names))
			for _, name := range names {
				r, options := byName[name], running[name]
				configured = append(configured, linter.ConfiguredRule{
					Name: r.Name,
					Run:  func(ctx rule.RuleContext) rule.RuleListeners { return r.Run(ctx, options) },
				})
			}
			return configured
		},
		OnDiagnostic:         func(d rule.RuleDiagnostic) { diagnostics <- d },
		OnInternalDiagnostic: func(diagnostic.Internal) {},
	})
	close(diagnostics)
	printing.Wait()

	if err != nil {
		fmt.Fprintf(os.Stderr, "archlint: %v\n", err)
		return exitFailed
	}

	fmt.Fprintf(
		os.Stderr,
		"%s: %s in %s with %s\n",
		namespace,
		count(reported, "error"),
		count(len(files), "file"),
		count(len(rules.enabledRules(registered)), "rule"),
	)

	if reported > 0 {
		return exitReported
	}
	return exitClean
}

// projectFiles is every file the tsconfig includes that the project actually
// owns. A dependency's declarations are in the program because the checker
// needs them, and are nobody's to judge.
func projectFiles(program *compiler.Program, projectDirectory string) []*ast.SourceFile {
	prefix := string(tspath.ToPath("", projectDirectory, true).EnsureTrailingDirectorySeparator())
	files := make([]*ast.SourceFile, 0, len(program.SourceFiles()))
	for _, file := range program.SourceFiles() {
		path := string(file.Path())
		if strings.Contains(path, "/node_modules/") {
			continue
		}
		if _, inside := strings.CutPrefix(path, prefix); inside {
			files = append(files, file)
		}
	}
	// Longest first: the linter's workers finish sooner when the slowest file
	// starts first.
	slices.SortFunc(files, func(a, b *ast.SourceFile) int { return len(b.Text()) - len(a.Text()) })
	return files
}

// writeDiagnostic prints one finding in oxlint's format:
//
//	path/to/file.ts:12:3: error arch-types(rule-name): What is wrong. What to do.
func writeDiagnostic(out *bufio.Writer, d rule.RuleDiagnostic, relativeTo tspath.ComparePathsOptions, documented map[string]bool) {
	line, column := scanner.GetECMALineAndUTF16CharacterOfPosition(d.SourceFile, d.Range.Pos())

	out.WriteString(tspath.ConvertToRelativePath(d.SourceFile.FileName(), relativeTo))
	out.WriteByte(':')
	out.WriteString(strconv.Itoa(line + 1))
	out.WriteByte(':')
	out.WriteString(strconv.Itoa(int(column) + 1))
	out.WriteString(": error ")
	out.WriteString(namespace)
	out.WriteByte('(')
	out.WriteString(d.RuleName)
	out.WriteString("): ")
	out.WriteString(oneLine(d.Message.Description))
	if d.Message.Help != "" {
		out.WriteByte(' ')
		out.WriteString(oneLine(d.Message.Help))
	}
	// Upstream's rules are documented by upstream, and this repository has no
	// page to send anyone to.
	if documented[d.RuleName] {
		out.WriteByte(' ')
		out.WriteString(archrule.DocumentedAt(d.RuleName))
	}
	out.WriteByte('\n')
}

// oneLine keeps a diagnostic to a single line, so the output stays greppable
// and `wc -l` counts findings.
func oneLine(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

func count(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return strconv.Itoa(n) + " " + noun + "s"
}

// loadConfig reads the configuration the run was pointed at, or the one beside
// the tsconfig, or none at all. Only an explicitly named file that is missing
// is an error: a repository that has not configured anything gets every rule on
// the defaults its own source declares.
func loadConfig(fs vfs.FS, projectDirectory string, configPath string) (config, error) {
	named := configPath != ""
	path := tspath.ResolvePath(projectDirectory, orDefault(configPath, configFileBasename))

	source, ok := fs.ReadFile(path)
	if !ok {
		if named {
			return config{}, fmt.Errorf("no configuration at %s", path)
		}
		return empty(), nil
	}

	parsed, err := parseConfig([]byte(source))
	if err != nil {
		return config{}, fmt.Errorf("%s: %w", path, err)
	}
	return parsed, nil
}

func orDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
