
// @module annotgen/main
// @brief CLI entrypoint for annotgen
// @description
// This is the main entry point of the tool. It parses arguments,
// walks files or directories, analyzes them, and writes annotated output.

package main

import (
	"flag"
	"fmt"
	"os"

	"annotgen/core"
	"annotgen/fs"
)

func main() {
	var (
		path     string
		outDir   string
		dryRun   bool
		asAnnot  bool
	)

	flag.StringVar(&path, "path", "", "Path to .lua file or directory")
	flag.StringVar(&outDir, "out", "", "Write output to this directory (optional)")
	flag.BoolVar(&dryRun, "dry-run", false, "Do not write to disk, print output instead")
	flag.BoolVar(&asAnnot, "annot", false, "Write output as .annot.lua")
	flag.Parse()

	if path == "" {
		fmt.Fprintln(os.Stderr, "Error: --path is required")
		os.Exit(1)
	}

	luaFiles, err := fs.WalkLuaFiles(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while walking path: %v\n", err)
		os.Exit(1)
	}

	for _, file := range luaFiles {
		ann, err := core.AnalyzeFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to analyze %s: %v\n", file, err)
			continue
		}

		core.MergeWithFileContent(ann)

		result, err := core.WriteFile(ann, core.WriteOptions{
			DryRun:   dryRun,
			OutDir:   outDir,
			AnnotExt: asAnnot,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write %s: %v\n", file, err)
			continue
		}

		if dryRun {
			fmt.Printf("-- %s --\n", file)
			for _, line := range result {
				fmt.Println(line)
			}
		}
	}
}
