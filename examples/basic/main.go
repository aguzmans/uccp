package main

import (
	"fmt"
	"log"

	"github.com/aguzmans/uccp/core"
	"github.com/aguzmans/uccp/domains"
)

func main() {
	fmt.Println("UCCP - Ultra-Compact Content Protocol Examples\n")

	// Example 1: Basic compression
	fmt.Println("=== Example 1: Basic Compression ===")
	basicCompressionExample()

	// Example 2: Smart compression decision
	fmt.Println("\n=== Example 2: Smart Compression Decision ===")
	smartCompressionExample()

	// Example 3: Write and read messages
	fmt.Println("\n=== Example 3: Write and Read Messages ===")
	writeReadExample()

	// Example 4: Project snapshot compression
	fmt.Println("\n=== Example 4: Project Snapshot Compression ===")
	projectSnapshotExample()

	// Example 5: Job result compression
	fmt.Println("\n=== Example 5: Job Result Compression ===")
	jobResultExample()

	// Example 6: Compression statistics
	fmt.Println("\n=== Example 6: Compression Statistics ===")
	statsExample()
}

func basicCompressionExample() {
	compressor := domains.NewCodeCompressor()

	original := "Use the function to implement authentication for the application"
	compressed, _ := compressor.Compress(original)

	fmt.Printf("Original:   %s (%d bytes)\n", original, len(original))
	fmt.Printf("Compressed: %s (%d bytes)\n", compressed, len(compressed))

	ratio := core.CalculateCompressionRatio(original, compressed)
	fmt.Printf("Compression ratio: %.1f%%\n", ratio*100)
}

func smartCompressionExample() {
	compressor := domains.NewCodeCompressor()

	examples := []string{
		"Hello",                                                       // Too small
		"This is a simple test message",                              // Too small
		"This is a longer message that describes the implementation", // Should compress
		"Successfully implemented the ActivityFeed component with infinite scroll pagination using the useInfiniteScroll hook from src/lib/hooks/useInfiniteScroll.ts. All tests passing.", // Should compress
	}

	for i, content := range examples {
		result := core.ShouldCompress(compressor, content, core.DefaultThresholds)

		fmt.Printf("\n%d. Content: %s\n", i+1, content)
		fmt.Printf("   Size: %d bytes\n", result.OriginalSize)
		fmt.Printf("   Compressed: %v\n", result.WasCompressed)

		if result.WasCompressed {
			fmt.Printf("   Ratio: %.1f%%\n", result.Ratio*100)
			fmt.Printf("   Tokens saved: ~%d\n", result.EstimatedTokenSavings)
			fmt.Printf("   Result: %s\n", result.Compressed)
		} else {
			fmt.Printf("   Reason: ")
			if result.OriginalSize < core.DefaultThresholds.MinSize {
				fmt.Printf("Too small (< %d bytes)\n", core.DefaultThresholds.MinSize)
			} else {
				fmt.Printf("Savings too low (< %.0f%%)\n", core.DefaultThresholds.MinRatio*100)
			}
		}
	}
}

func writeReadExample() {
	compressor := domains.NewCodeCompressor()

	// Write a message
	jobSummary := "Job job-021-activity-feed completed successfully in 18 minutes 32 seconds. Modified file: src/components/pages/ActivityFeed.tsx. Created test: src/components/pages/__tests__/ActivityFeed.test.tsx. Tests: 5 passed, 0 failed. Result: Successfully implemented ActivityFeed component with infinite scroll. All tests passing."

	path, result, err := core.WriteMessage(compressor, jobSummary, "/tmp/job-021", core.DefaultThresholds)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Written to: %s\n", path)
	fmt.Printf("Compressed: %v\n", result.WasCompressed)
	fmt.Printf("Original size: %d bytes\n", result.OriginalSize)
	fmt.Printf("Final size: %d bytes\n", result.CompressedSize)

	if result.WasCompressed {
		fmt.Printf("Compression ratio: %.1f%%\n", result.Ratio*100)
		fmt.Printf("Tokens saved: ~%d\n", result.EstimatedTokenSavings)
	}

	// Read it back
	content, wasCompressed, prompt, err := core.ReadMessage(compressor, "/tmp/job-021")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nRead back from file:\n")
	fmt.Printf("Was compressed: %v\n", wasCompressed)
	if wasCompressed {
		fmt.Printf("System prompt length: %d chars\n", len(prompt))
		fmt.Printf("Content: %s\n", content)
	}
}

func projectSnapshotExample() {
	compressor := domains.NewCodeCompressor()

	snapshot := map[string]interface{}{
		"architecture": map[string]interface{}{
			"framework":  "React with TypeScript",
			"build_tool": "Vite",
			"language":   "TypeScript",
		},
		"patterns": map[string]interface{}{
			"api_calls":        "Use api.get() from src/lib/api.ts",
			"state_management": "Use useState and useContext hooks",
			"routing":          "React Router in src/App.tsx",
		},
	}

	compressed, _ := compressor.CompressProjectSnapshot(snapshot)

	fmt.Printf("Compressed snapshot:\n%s\n\n", compressed)

	// Calculate savings
	import "encoding/json"
	original, _ := json.Marshal(snapshot)
	ratio := core.CalculateCompressionRatio(string(original), compressed)

	fmt.Printf("Original size: %d bytes\n", len(original))
	fmt.Printf("Compressed size: %d bytes\n", len(compressed))
	fmt.Printf("Compression ratio: %.1f%%\n", ratio*100)
	fmt.Printf("Token savings: ~%d tokens\n", core.EstimateTokenSavings(string(original), compressed))
}

func jobResultExample() {
	compressor := domains.NewCodeCompressor()

	result := map[string]interface{}{
		"job_id":         "job-021-activity-feed",
		"status":         "completed",
		"worker_id":      "worker-csa-abc123",
		"execution_time": "18m 32s",
		"files_modified": []interface{}{"src/components/pages/ActivityFeed.tsx"},
		"files_created":  []interface{}{"src/components/pages/__tests__/ActivityFeed.test.tsx"},
		"tests_run":      5,
		"tests_passed":   5,
		"tests_failed":   0,
		"result":         "Successfully implemented ActivityFeed component with infinite scroll. All tests passing.",
	}

	compressed, _ := compressor.CompressJobResult(result)

	fmt.Printf("Compressed job result:\n%s\n\n", compressed)

	// Calculate savings
	import "encoding/json"
	original, _ := json.Marshal(result)
	ratio := core.CalculateCompressionRatio(string(original), compressed)

	fmt.Printf("Original size: %d bytes\n", len(original))
	fmt.Printf("Compressed size: %d bytes\n", len(compressed))
	fmt.Printf("Compression ratio: %.1f%%\n", ratio*100)
}

func statsExample() {
	compressor := domains.NewCodeCompressor()
	stats := &core.CompressionStats{}

	// Simulate compressing multiple messages
	messages := []string{
		"Successfully implemented the feature",
		"Error: Authentication failed with status code 401",
		"Updated configuration to use the new API endpoint from src/lib/api.ts",
		"Test suite completed: 15 tests passed, 2 tests failed, 1 test skipped",
		"Deployed application to production environment with zero downtime",
	}

	fmt.Println("Compressing multiple messages...\n")

	for _, msg := range messages {
		result := core.ShouldCompress(compressor, msg, core.DefaultThresholds)
		core.UpdateStats(stats, result)
	}

	fmt.Printf("Total compressions attempted: %d\n", stats.TotalCompressions)
	fmt.Printf("Successful compressions: %d\n", stats.SuccessfulCompressions)
	fmt.Printf("Skipped compressions: %d\n", stats.SkippedCompressions)
	fmt.Printf("Average compression ratio: %.1f%%\n", stats.AverageRatio*100)
	fmt.Printf("Best compression ratio: %.1f%%\n", stats.BestRatio*100)
	fmt.Printf("Worst compression ratio: %.1f%%\n", stats.WorstRatio*100)
	fmt.Printf("Total bytes saved: %d\n", stats.TotalBytesSaved)
	fmt.Printf("Total tokens saved: %d\n", stats.TotalTokensSaved)

	// Calculate monthly cost savings
	tokensPerDay := int(stats.TotalTokensSaved) * 100 // Assume 100x daily usage
	monthlySavings := core.CalculateCostSavings(tokensPerDay)
	fmt.Printf("\nProjected monthly savings (100x scale): $%.2f\n", monthlySavings)
}
