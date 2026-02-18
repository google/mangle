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

// Package lint provides a standalone linter for Mangle programs. It reuses the
// existing analysis infrastructure and adds additional style and quality checks.
package lint

import (
	"bufio"
	"fmt"
	"os"

	"github.com/google/mangle/analysis"
	"github.com/google/mangle/ast"
	"github.com/google/mangle/parse"
)

// Severity levels for lint findings.
type Severity int

const (
	// SeverityInfo is for informational findings that may not indicate a problem.
	SeverityInfo Severity = iota
	// SeverityWarning is for findings that likely indicate a problem.
	SeverityWarning
	// SeverityError is for findings that definitely indicate a problem.
	SeverityError
)

// MarshalJSON encodes severity as a JSON string.
func (s Severity) MarshalJSON() ([]byte, error) {
	return []byte(`"` + s.String() + `"`), nil
}

// String returns the human-readable name of a severity level.
func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	default:
		return "unknown"
	}
}

// ParseSeverity parses a severity string. Returns SeverityInfo if unrecognized.
func ParseSeverity(s string) Severity {
	switch s {
	case "warning":
		return SeverityWarning
	case "error":
		return SeverityError
	default:
		return SeverityInfo
	}
}

// LintResult represents a single finding from a lint check.
type LintResult struct {
	// RuleName is the machine-readable name of the lint rule.
	RuleName string `json:"rule"`
	// Severity of the finding.
	Severity Severity `json:"severity"`
	// File is the source file path.
	File string `json:"file,omitempty"`
	// Message is a human-readable description of the finding.
	Message string `json:"message"`
	// Predicate is the predicate involved, if applicable.
	Predicate string `json:"predicate,omitempty"`
}

// LintConfig holds the toggleable configuration for all lint rules.
type LintConfig struct {
	// MaxPremises is the threshold for the overly-complex-rule check.
	MaxPremises int
	// DisabledRules is a set of rule names to skip.
	DisabledRules map[string]bool
	// MinSeverity: findings below this severity are suppressed from output.
	MinSeverity Severity
}

// DefaultConfig returns a LintConfig with sensible defaults.
func DefaultConfig() LintConfig {
	return LintConfig{
		MaxPremises:   8,
		DisabledRules: map[string]bool{},
		MinSeverity:   SeverityInfo,
	}
}

// LintInput bundles everything a lint check needs.
type LintInput struct {
	// File is the source file path.
	File string
	// Unit is the parsed source unit.
	Unit parse.SourceUnit
	// ProgramInfo is the result of analysis.Analyze.
	ProgramInfo *analysis.ProgramInfo
	// Strata from analysis.Stratify (nil if stratification failed).
	Strata []analysis.Nodeset
	// PredToStratum maps each predicate to its stratum number (nil if stratification failed).
	PredToStratum map[ast.PredicateSym]int
}

// Linter orchestrates parsing, analysis, and lint checks.
type Linter struct {
	config LintConfig
	rules  []LintRule
}

// NewLinter creates a Linter with the given config and all registered rules.
func NewLinter(config LintConfig) *Linter {
	return &Linter{
		config: config,
		rules:  AllRules(),
	}
}

// LintFile parses and lints a single .mg file. Returns results and any hard
// errors (parse failure, etc.). Analysis errors are surfaced as lint findings
// rather than hard errors when possible.
func (l *Linter) LintFile(path string) ([]LintResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	unit, err := parse.Unit(bufio.NewReaderSize(f, 4096))
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return l.LintUnit(path, unit)
}

// LintUnit lints a pre-parsed source unit.
func (l *Linter) LintUnit(file string, unit parse.SourceUnit) ([]LintResult, error) {
	// Run existing analysis.
	programInfo, err := analysis.AnalyzeOneUnit(unit, nil)
	if err != nil {
		// Return analysis error as a lint finding so we still report something useful.
		return []LintResult{{
			RuleName: "analysis",
			Severity: SeverityError,
			File:     file,
			Message:  fmt.Sprintf("analysis error: %v", err),
		}}, nil
	}

	var results []LintResult

	// Surface temporal warnings from existing analysis.
	for _, w := range programInfo.Warnings {
		sev := convertTemporalSeverity(w.Severity)
		if sev < l.config.MinSeverity {
			continue
		}
		results = append(results, LintResult{
			RuleName:  "temporal-recursion",
			Severity:  sev,
			File:      file,
			Message:   w.Message,
			Predicate: w.Predicate.Symbol,
		})
	}

	// Run stratification check.
	program := analysis.Program{
		EdbPredicates: programInfo.EdbPredicates,
		IdbPredicates: programInfo.IdbPredicates,
		Rules:         programInfo.Rules,
	}
	strata, predToStratum, stratErr := analysis.Stratify(program)
	if stratErr != nil {
		results = append(results, LintResult{
			RuleName: "stratification",
			Severity: SeverityError,
			File:     file,
			Message:  stratErr.Error(),
		})
	}

	// Build input for lint rules.
	input := &LintInput{
		File:          file,
		Unit:          unit,
		ProgramInfo:   programInfo,
		Strata:        strata,
		PredToStratum: predToStratum,
	}

	// Run each enabled lint rule.
	for _, rule := range l.rules {
		if l.config.DisabledRules[rule.Name()] {
			continue
		}
		findings := rule.Check(input, l.config)
		for _, f := range findings {
			if f.Severity >= l.config.MinSeverity {
				f.File = file
				results = append(results, f)
			}
		}
	}

	return results, nil
}

func convertTemporalSeverity(s analysis.WarningSeverity) Severity {
	switch s {
	case analysis.SeverityWarning:
		return SeverityWarning
	case analysis.SeverityCritical:
		return SeverityError
	default:
		return SeverityInfo
	}
}
