// Copyright 2026 Google LLC
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

// Command scinfo prints header information about a simplecolumn file.
//
// Usage:
//
//	scinfo [file]
//
// Compression is detected by file extension: .gz uses gzip, .zst or .zstd uses
// zstd, anything else is read as raw simplecolumn data.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"codeberg.org/TauCeti/mangle-go/factstore"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: scinfo <file>\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	path := flag.Arg(0)
	if err := run(path, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "scinfo: %v\n", err)
		os.Exit(1)
	}
}

func run(path string, out *os.File) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var (
		store       *factstore.SimpleColumnStore
		compression string
	)
	switch {
	case strings.HasSuffix(path, ".gz"):
		compression = "gzip"
		store, err = factstore.NewSimpleColumnStoreFromGzipBytes(data)
	case strings.HasSuffix(path, ".zst"), strings.HasSuffix(path, ".zstd"):
		compression = "zstd"
		store, err = factstore.NewSimpleColumnStoreFromZstdBytes(data)
	default:
		compression = "none"
		store, err = factstore.NewSimpleColumnStoreFromBytes(data)
	}
	if err != nil {
		return err
	}

	preds := store.ListPredicates()
	fmt.Fprintf(out, "file:        %s\n", path)
	fmt.Fprintf(out, "size:        %d bytes\n", len(data))
	fmt.Fprintf(out, "compression: %s\n", compression)
	fmt.Fprintf(out, "predicates:  %d\n", len(preds))
	fmt.Fprintf(out, "facts:       %d\n", store.EstimateFactCount())
	if len(preds) == 0 {
		return nil
	}
	fmt.Fprintln(out)
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PREDICATE\tARITY\tFACTS")
	for _, p := range preds {
		fmt.Fprintf(tw, "%s\t%d\t%d\n", p.Symbol, p.Arity, store.FactCount(p))
	}
	return tw.Flush()
}
