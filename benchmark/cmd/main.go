package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aguzmans/uccp/benchmark"
)

func main() {
	// Resolve paths relative to the repo root
	benchDir := findBenchmarkDir()
	repoRoot := filepath.Join(benchDir, "..")
	testDataDir := filepath.Join(benchDir, "testdata")
	graphOutput := filepath.Join(repoRoot, "docs", "benchmark-results.svg")
	historyDir := filepath.Join(repoRoot, "docs", "benchmark-history")
	historyJSON := filepath.Join(repoRoot, "docs", "benchmark-history.json")
	readmePath := filepath.Join(repoRoot, "README.md")

	// Step 1: Generate test data
	fmt.Println("=== Generating test data ===")
	if err := benchmark.GenerateTestData(testDataDir); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR generating test data: %v\n", err)
		os.Exit(1)
	}
	fmt.Println()

	// Step 2: Run benchmarks
	fmt.Printf("=== Running benchmarks (tiktoken cl100k_base, amortization depth=%d) ===\n",
		benchmark.DefaultAmortizationDepth)
	results, err := benchmark.RunBenchmarks(testDataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR running benchmarks: %v\n", err)
		os.Exit(1)
	}
	fmt.Println()

	// Step 3: Print results table
	benchmark.PrintResults(results)
	fmt.Println()

	// Step 4: Generate SVG graph (dual-panel: raw + amortized net)
	fmt.Println("=== Generating benchmark graph ===")
	if err := os.MkdirAll(filepath.Dir(graphOutput), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR creating docs dir: %v\n", err)
		os.Exit(1)
	}
	if err := benchmark.GenerateGraph(results, graphOutput); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR generating graph: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Graph written to: %s\n", graphOutput)
	fmt.Println()

	// Step 5: Archive to history
	fmt.Println("=== Archiving benchmark results ===")
	history, err := benchmark.LoadHistory(historyJSON)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: could not load history: %v\n", err)
		history = &benchmark.BenchmarkHistory{}
	}

	svgFile, err := benchmark.ArchiveSVG(graphOutput, historyDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: could not archive SVG: %v\n", err)
	} else {
		run := benchmark.BenchmarkRun{
			Timestamp: time.Now().Format(time.RFC3339),
			GitCommit: benchmark.GitShortSHA(),
			SVGFile:   svgFile,
			Summary:   benchmark.BuildRunSummary(results),
		}
		benchmark.AddRun(history, run, historyDir)
		if err := benchmark.SaveHistory(history, historyJSON); err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: could not save history: %v\n", err)
		} else {
			fmt.Printf("Archived to: %s (%d runs in history)\n", svgFile, len(history.Runs))
		}
	}
	fmt.Println()

	// Step 6: Update README
	fmt.Println("=== Updating README.md ===")
	if err := benchmark.UpdateREADME(readmePath, results); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: could not update README: %v\n", err)
	} else {
		fmt.Println("README.md benchmarks section updated.")
	}
}

// findBenchmarkDir walks up from the executable/cwd to find the benchmark directory.
func findBenchmarkDir() string {
	candidates := []string{
		"benchmark",
		"../benchmark",
		filepath.Join(os.Getenv("GOPATH"), "src/github.com/aguzmans/uccp/benchmark"),
	}

	cwd, _ := os.Getwd()
	for _, c := range candidates {
		abs := c
		if !filepath.IsAbs(c) {
			abs = filepath.Join(cwd, c)
		}
		if info, err := os.Stat(abs); err == nil && info.IsDir() {
			return abs
		}
	}

	return filepath.Join(cwd, "benchmark")
}
