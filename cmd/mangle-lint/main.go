// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Binary mangle-lint is a standalone linter for Mangle programs.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/mangle/lint"
)

var (
	format      = flag.String("format", "text", "output format: text or json")
	severity    = flag.String("severity", "info", "minimum severity to report: info, warning, or error")
	disable     = flag.String("disable", "", "comma-separated list of rule names to disable")
	enable      = flag.String("enable", "", "comma-separated list of rule names to enable (all others disabled)")
	listRules   = flag.Bool("list-rules", false, "list all available lint rules and exit")
	maxPremises = flag.Int("max-premises", 8, "threshold for overly-complex-rule check")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: mangle-lint [flags] <file.mg> [file.mg...]\n\n")
		fmt.Fprintf(os.Stderr, "A linter for the Mangle Datalog language.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExit codes:\n")
		fmt.Fprintf(os.Stderr, "  0  No findings (or only info)\n")
		fmt.Fprintf(os.Stderr, "  1  Warnings found\n")
		fmt.Fprintf(os.Stderr, "  2  Errors found\n")
	}
	flag.Parse()

	if *listRules {
		fmt.Println("Available lint rules:")
		fmt.Println()
		for _, r := range lint.AllRules() {
			fmt.Printf("  %-25s [%s]  %s\n", r.Name(), r.DefaultSeverity(), r.Description())
		}
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	// Expand glob patterns.
	var files []string
	for _, arg := range args {
		matches, err := filepath.Glob(arg)
		if err != nil || len(matches) == 0 {
			files = append(files, arg) // treat as literal path
		} else {
			files = append(files, matches...)
		}
	}

	config := lint.DefaultConfig()
	config.MaxPremises = *maxPremises
	config.MinSeverity = lint.ParseSeverity(*severity)

	if *disable != "" {
		for _, name := range strings.Split(*disable, ",") {
			config.DisabledRules[strings.TrimSpace(name)] = true
		}
	}
	if *enable != "" {
		// Disable all rules first, then enable only the specified ones.
		for _, r := range lint.AllRules() {
			config.DisabledRules[r.Name()] = true
		}
		for _, name := range strings.Split(*enable, ",") {
			delete(config.DisabledRules, strings.TrimSpace(name))
		}
	}

	linter := lint.NewLinter(config)
	var allResults []lint.LintResult
	hasParseError := false

	for _, path := range files {
		results, err := linter.LintFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			hasParseError = true
			continue
		}
		allResults = append(allResults, results...)
	}

	// Output.
	switch *format {
	case "json":
		if err := lint.FormatJSON(os.Stdout, allResults); err != nil {
			fmt.Fprintf(os.Stderr, "error writing JSON: %v\n", err)
			os.Exit(2)
		}
	default:
		lint.FormatText(os.Stdout, allResults)
	}

	// Exit code.
	if hasParseError {
		os.Exit(2)
	}
	maxSev := lint.SeverityInfo
	for _, r := range allResults {
		if r.Severity > maxSev {
			maxSev = r.Severity
		}
	}
	switch {
	case maxSev >= lint.SeverityError:
		os.Exit(2)
	case maxSev >= lint.SeverityWarning:
		os.Exit(1)
	default:
		os.Exit(0)
	}
}
