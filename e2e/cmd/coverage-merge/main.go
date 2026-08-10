package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

func main() {
	root := flag.String("root", "", "unique P0/P1 aggregate artifact root")
	output := flag.String("output", harness.UpgradeCoverageMatrixArtifactPath, "root-relative aggregate matrix path")
	sourceCommit := flag.String("source-commit", "", "full source commit shared by the P0/P1 gate")
	flag.Parse()
	if *root == "" {
		fmt.Fprintln(os.Stderr, "coverage merge requires -root")
		os.Exit(2)
	}
	if *sourceCommit == "" {
		fmt.Fprintln(os.Stderr, "coverage merge requires -source-commit")
		os.Exit(2)
	}
	matrix, err := harness.WriteMergedUpgradeCoverageMatrix(*root, *output, time.Now().UTC())
	if err != nil {
		if recordErr := harness.WriteP0P1ReleaseGateFailure(*root, "coverage-merge", err, time.Now().UTC()); recordErr != nil {
			fmt.Fprintf(os.Stderr, "record P0/P1 coverage merge failure: %v\n", recordErr)
		}
		fmt.Fprintf(os.Stderr, "merge P0/P1 upgrade coverage: %v\n", err)
		os.Exit(1)
	}
	gate, err := harness.WriteP0P1ReleaseGateManifest(*root, *sourceCommit, time.Now().UTC())
	if err != nil {
		if recordErr := harness.WriteP0P1ReleaseGateFailure(*root, "release-gate-validation", err, time.Now().UTC()); recordErr != nil {
			fmt.Fprintf(os.Stderr, "record P0/P1 release gate failure: %v\n", recordErr)
		}
		fmt.Fprintf(os.Stderr, "write P0/P1 release gate manifest: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf(
		"merged %d source matrices into %s with %d fully passed P0/P1 rows; verified %d live suites into %s\n",
		len(matrix.SourceMatrices), *output, len(matrix.Rows), len(gate.RequiredSuites), harness.P0P1ReleaseGateArtifactPath,
	)
}
