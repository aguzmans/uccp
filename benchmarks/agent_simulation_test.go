package benchmarks

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aguzmans/uccp/core"
	"github.com/aguzmans/uccp/domains"
)

// TestAgentCommunicationScenario simulates realistic agent-to-agent communication
func TestAgentCommunicationScenario(t *testing.T) {
	scenarios := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "Job Result Summary",
			content: `{
  "job_id": "job-021",
  "status": "completed",
  "worker_id": "worker-abc123",
  "execution_time": "18m 32s",
  "files_modified": [
    "src/components/pages/ActivityFeed.tsx",
    "src/components/common/InfiniteScroll.tsx"
  ],
  "tests_run": 5,
  "tests_passed": 5,
  "result": "Successfully implemented ActivityFeed component with infinite scroll functionality. Added pagination support and loading states."
}`,
			expected: "UCCP compression achieves 70-85% reduction",
		},
		{
			name: "Project Architecture Snapshot",
			content: `{
  "architecture": {
    "framework": "React with TypeScript",
    "build_tool": "Vite",
    "language": "TypeScript"
  },
  "patterns": {
    "api_calls": "Use api.get() from src/lib/api.ts",
    "state_management": "Use useState and useContext hooks",
    "routing": "React Router v6 with lazy loading"
  },
  "conventions": {
    "imports": "Absolute imports with @ alias",
    "naming": "PascalCase for components, camelCase for files",
    "styling": "TailwindCSS utility classes"
  }
}`,
			expected: "UCCP compression achieves 60-75% reduction",
		},
		{
			name: "File Index Metadata",
			content: `{
  "src/lib/api.ts": {
    "purpose": "API client with authentication and error handling",
    "exports": ["api object"],
    "dependencies": ["axios", "auth"]
  },
  "src/components/ActivityFeed.tsx": {
    "purpose": "Main activity feed component with infinite scroll",
    "exports": ["ActivityFeed component"],
    "dependencies": ["InfiniteScroll", "ActivityCard"]
  },
  "src/hooks/useAuth.ts": {
    "purpose": "Authentication state management hook",
    "exports": ["useAuth hook"],
    "dependencies": ["react", "auth-context"]
  }
}`,
			expected: "UCCP compression achieves 65-80% reduction",
		},
	}

	compressor := domains.NewCodeCompressor()

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Compress
			compressed, err := compressor.Compress(scenario.content)
			if err != nil {
				t.Fatalf("Compression failed: %v", err)
			}

			// Calculate metrics
			ratio := core.CalculateCompressionRatio(scenario.content, compressed)
			originalTokens := core.EstimateTokenCount(scenario.content)
			compressedTokens := core.EstimateTokenCount(compressed)
			tokenSavings := core.EstimateTokenSavings(scenario.content, compressed)

			// Report results
			t.Logf("\n=== %s ===", scenario.name)
			t.Logf("Original size: %d bytes (%d tokens)", len(scenario.content), originalTokens)
			t.Logf("Compressed size: %d bytes (%d tokens)", len(compressed), compressedTokens)
			t.Logf("Compression ratio: %.1f%%", ratio*100)
			t.Logf("Token savings: %d tokens (%.1f%%)", tokenSavings, (float64(tokenSavings)/float64(originalTokens))*100)
			t.Logf("\nOriginal:\n%s", scenario.content)
			t.Logf("\nCompressed:\n%s", compressed)

			// Validate compression is beneficial
			if ratio < 0.50 {
				t.Errorf("Compression ratio too low: %.1f%% (expected >= 50%%)", ratio*100)
			}
		})
	}
}

// TestManagerReadsMultipleJobs simulates manager reading 34 completed job summaries
func TestManagerReadsMultipleJobs(t *testing.T) {
	compressor := domains.NewCodeCompressor()

	// Simulate 34 job results
	jobCount := 34
	var totalOriginalTokens, totalCompressedTokens int

	for i := 1; i <= jobCount; i++ {
		jobResult := map[string]interface{}{
			"job_id":         fmt.Sprintf("job-%03d", i),
			"status":         "completed",
			"worker_id":      "worker-abc",
			"execution_time": "15m",
			"files_modified": []string{"src/component.tsx"},
			"tests_passed":   5,
			"result":         "Successfully implemented the feature",
		}

		resultJSON, _ := json.Marshal(jobResult)
		compressed, _ := compressor.Compress(string(resultJSON))

		totalOriginalTokens += core.EstimateTokenCount(string(resultJSON))
		totalCompressedTokens += core.EstimateTokenCount(compressed)
	}

	tokenSavings := totalOriginalTokens - totalCompressedTokens
	percentSaved := (float64(tokenSavings) / float64(totalOriginalTokens)) * 100

	t.Logf("\n=== Manager Reading 34 Job Summaries ===")
	t.Logf("Without UCCP:")
	t.Logf("  Total tokens: %d", totalOriginalTokens)
	t.Logf("  Estimated cost: $%.4f", float64(totalOriginalTokens)*0.003/1000)
	t.Logf("\nWith UCCP:")
	t.Logf("  Total tokens: %d", totalCompressedTokens)
	t.Logf("  Estimated cost: $%.4f", float64(totalCompressedTokens)*0.003/1000)
	t.Logf("\nSavings:")
	t.Logf("  Token reduction: %d tokens (%.1f%%)", tokenSavings, percentSaved)
	t.Logf("  Cost savings: $%.4f", float64(tokenSavings)*0.003/1000)

	// Validate significant savings
	if percentSaved < 70 {
		t.Errorf("Token savings too low: %.1f%% (expected >= 70%%)", percentSaved)
	}
}

// BenchmarkAgentCommunication benchmarks compression performance
func BenchmarkAgentCommunication(b *testing.B) {
	compressor := domains.NewCodeCompressor()

	jobResult := `{
  "job_id": "job-021",
  "status": "completed",
  "worker_id": "worker-abc123",
  "execution_time": "18m 32s",
  "files_modified": ["src/components/pages/ActivityFeed.tsx"],
  "tests_run": 5,
  "tests_passed": 5,
  "result": "Successfully implemented ActivityFeed with infinite scroll"
}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = compressor.Compress(jobResult)
	}
}

// BenchmarkSmartCompressionDecision benchmarks the decision logic
func BenchmarkSmartCompressionDecision(b *testing.B) {
	compressor := domains.NewCodeCompressor()
	content := "Successfully implemented the authentication feature with JWT tokens"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = core.ShouldCompress(compressor, content, core.DefaultThresholds)
	}
}
