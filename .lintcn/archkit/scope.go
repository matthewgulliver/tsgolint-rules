package archkit

import (
	"path"
	"regexp"
	"strings"
)

func Includes(patterns []string, fileName string) bool {
	segments := split(fileName)
	for _, pattern := range patterns {
		if matches(split(pattern), segments) {
			return true
		}
	}
	return false
}

func split(p string) []string {
	normalised := strings.ReplaceAll(p, "\\", "/")
	parts := strings.Split(normalised, "/")
	kept := parts[:0]
	for _, part := range parts {
		if part != "" && part != "." {
			kept = append(kept, part)
		}
	}
	return kept
}

func matches(pattern []string, name []string) bool {
	if len(pattern) == 0 {
		return len(name) == 0
	}

	if pattern[0] == "**" {
		for consumed := 0; consumed <= len(name); consumed++ {
			if matches(pattern[1:], name[consumed:]) {
				return true
			}
		}
		return false
	}

	if len(name) == 0 {
		return false
	}
	ok, err := path.Match(pattern[0], name[0])
	if err != nil || !ok {
		return false
	}
	return matches(pattern[1:], name[1:])
}

func IsOutsideDependency(fileName string) bool {
	return IsPackageDependency(fileName) || IsStandardLibrary(fileName)
}

func IsPackageDependency(fileName string) bool {
	for _, segment := range split(fileName) {
		if segment == "node_modules" {
			return true
		}
	}
	return false
}

func IsStandardLibrary(fileName string) bool {
	base := path.Base(strings.ReplaceAll(fileName, "\\", "/"))
	return strings.HasPrefix(base, "lib.") && strings.HasSuffix(base, ".d.ts")
}

func Compile(patterns []string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		if expression, err := regexp.Compile(pattern); err == nil {
			compiled = append(compiled, expression)
		}
	}
	return compiled
}

func ContextOf(patterns []*regexp.Regexp, file string) (string, bool) {
	for _, expression := range patterns {
		group := captureIndex(expression)
		if group < 0 {
			continue
		}
		identity, identified := "", false
		for at := 0; at <= len(file); {
			match := expression.FindStringSubmatchIndex(file[at:])
			if match == nil {
				break
			}
			start, end := match[2*group], match[2*group+1]
			if start >= 0 && end > start {
				identity, identified = file[at+start:at+end], true
			}
			at += match[0] + 1
		}
		if identified {
			return identity, true
		}
	}
	return "", false
}

func captureIndex(expression *regexp.Regexp) int {
	for index, group := range expression.SubexpNames() {
		if group == "name" {
			return index
		}
	}
	return -1
}
