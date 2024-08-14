// Copyright 2022 Google LLC
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

// Binary mg is a shell for the interactive interpreter.
package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"flag"
	
	log "github.com/golang/glog"
	"github.com/google/mangle/interpreter"
)

var (
	load  = flag.String("load", "", "comma-separated list of libraries to load")
	exec  = flag.String("exec", "", "if non-empty, runs single query and exits with code 0 if result is non-empty")
	root  = flag.String("root", "", "all ::load commands are relative to this directory.")
	out   = flag.String("out", "", "if non-empty, output to file.")
	stats = flag.String("stats", "", "comma-separated list of predicates to show stats for, or 'all'")
)

func main() {
	flag.Parse()
	writer := os.Stdout
	if *out != "" {
		f, err := os.Create(*out)
		if err != nil {
			log.Exit(err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Exit(err)
			}
		}()
		writer = f
	}
	i := interpreter.New(writer, *root, strings.Split(*stats, ","))
	if len(*load) > 0 {
		if err := i.Load(*load); err != nil {
			log.Exitf("error loading src %s: %v", *load, err)
		}
	}

	if *exec != "" {
		query, err := i.ParseQuery(*exec)
		if err != nil {
			log.Exitf("error parsing query %q: %v", *exec, err)
		}
		res, err := i.Query(query)
		if err != nil {
			log.Exitf("error evaluating query %v: %v", query, err)
		}
		var results []string
		for _, fact := range res {
			results = append(results, fact.String())
		}
		fmt.Fprintf(writer, "%s\n", strings.Join(results, "\n"))

		if len(res) != 0 {
			fmt.Fprintln(writer, "#PASS")
			os.Exit(0)
		}
		fmt.Fprintln(writer, "#FAIL")
		os.Exit(1)
	} else if err := i.Loop(); err != io.EOF {
		log.Exit(err)
	}
	os.Exit(0)
}
