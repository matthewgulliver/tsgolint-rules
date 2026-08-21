package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"tsgolint-rules/tools/nocomments"
)

func main() { singlechecker.Main(nocomments.Analyzer) }
