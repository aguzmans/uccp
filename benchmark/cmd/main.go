package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aguzmans/uccp/benchmark"
)

func main() {
	// Resolve paths relative to the repo root
	benchDir := findBenchmarkDir()
	testDataDir := filepath.Join(benchDir, "testdata")
	graphOutput := filepath.Join(benchDir, "..", "docs", "benchmark-results.svg")

	// Step 1: Generate test data
	fmt.Println("=== Generating test data ===")
	if err := benchmark.GenerateTestData(testDataDir); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR generating test data: %v\n", err)
		os.Exit(1)
	}
	fmt.Println()

	// Step 2: Run benchmarks
	fmt.Println("=== Running benchmarks (with tiktoken cl100k_base) ===")
	results, err := benchmark.RunBenchmarks(testDataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR running benchmarks: %v\n", err)
		os.Exit(1)
	}
	fmt.Println()

	// Step 3: Print results table
	benchmark.PrintResults(results)
	fmt.Println()

	// Step 4: Generate SVG graph
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
}

// findBenchmarkDir walks up from the executable/cwd to find the benchmark directory.
func findBenchmarkDir() string {
	// Try relative to cwd first
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

	// Fallback: use cwd/benchmark
	return filepath.Join(cwd, "benchmark")
}
