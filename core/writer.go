package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"annotgen/types"
)

type WriteOptions struct {
	DryRun   bool
	OutDir   string
	AnnotExt bool
}

func WriteFile(fa *types.FileAnnotations, opts WriteOptions) ([]string, error) {
	var final []string

	if opts.AnnotExt {
		final = BuildStandaloneHeader(fa)
	} else if fa.Updated != nil {
		final = fa.Updated
	} else {
		final = fa.Existing
	}

	if opts.DryRun {
		return final, nil
	}

	var outPath string
	if opts.OutDir != "" {
		base := filepath.Base(fa.Path)
		outPath = filepath.Join(opts.OutDir, base)
		if opts.AnnotExt {
			outPath = strings.TrimSuffix(outPath, ".lua") + ".annot.lua"
		}
	} else if opts.AnnotExt {
		outPath = strings.TrimSuffix(fa.Path, ".lua") + ".annot.lua"
	} else {
		outPath = fa.Path
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", outPath, err)
	}
	defer outFile.Close()

	for _, line := range final {
		_, _ = outFile.WriteString(line + "\n")
	}

	return final, nil
}
