package benchmark

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aguzmans/uccp/core"
	"github.com/aguzmans/uccp/domains"
	"github.com/pkoukk/tiktoken-go"
)

// BenchmarkResult holds the results for a single file benchmark
type BenchmarkResult struct {
	Name     string
	Domain   string
	Category string

	// Byte metrics
	OriginalBytes   int
	CompressedBytes int
	ByteRatio       float64 // 0.0-1.0, percentage saved

	// Token metrics (tiktoken cl100k_base encoding, used by GPT-4/Claude)
	OriginalTokens   int
	CompressedTokens int
	TokenRatio       float64 // percentage saved

	// System prompt overhead
	SystemPromptTokens int

	// Net token savings (accounting for system prompt)
	NetTokenSavings int
	NetTokenRatio   float64 // net percentage saved

	// Performance
	CompressTime time.Duration

	// Whether compression was applied (ShouldCompress decision)
	WasCompressed bool
}

// TestFile describes a test data file for benchmarking.
type TestFile struct {
	Name     string // display name
	Path     string // relative path under testDataDir
	Domain   string // "html" or "code"
	Category string // e.g. "documentation", "source", "config"
}

var (
	tiktokenOnce sync.Once
	tiktokenEnc  *tiktoken.Tiktoken
	tiktokenErr  error
)

func getTokenizer() (*tiktoken.Tiktoken, error) {
	tiktokenOnce.Do(func() {
		tiktokenEnc, tiktokenErr = tiktoken.GetEncoding("cl100k_base")
	})
	return tiktokenEnc, tiktokenErr
}

// countTokens uses tiktoken-go with cl100k_base encoding to count tokens.
// Falls back to a chars/4 estimate if tiktoken fails to initialize.
func countTokens(text string) int {
	enc, err := getTokenizer()
	if err != nil {
		log.Printf("WARNING: tiktoken unavailable, falling back to chars/4 estimate: %v", err)
		return len(text) / 4
	}
	tokens := enc.Encode(text, nil, nil)
	return len(tokens)
}

// compressorForFile returns the appropriate compressor based on file extension.
func compressorForFile(path string) core.Compressor {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".html":
		return domains.NewHTMLCompressor()
	case ".go", ".tsx", ".py", ".json":
		return domains.NewCodeCompressor()
	default:
		// Default to code compressor for unknown extensions
		return domains.NewCodeCompressor()
	}
}

// RunBenchmarks iterates over all test files from TestDataFiles(),
// reads each, compresses with the appropriate domain compressor,
// and measures byte/token metrics using tiktoken for real token counts.
func RunBenchmarks(testDataDir string) ([]BenchmarkResult, error) {
	files := TestDataFiles()
	var results []BenchmarkResult

	for _, tf := range files {
		fullPath := filepath.Join(testDataDir, tf.Path)

		content, err := os.ReadFile(fullPath)
		if err != nil {
			// Skip files that don't exist rather than failing the whole benchmark
			log.Printf("WARNING: skipping %s: %v", tf.Name, err)
			continue
		}

		original := string(content)
		compressor := compressorForFile(tf.Path)

		// Measure compression time
		start := time.Now()
		compressed, compErr := compressor.Compress(original)
		elapsed := time.Since(start)

		if compErr != nil {
			log.Printf("WARNING: compression failed for %s: %v", tf.Name, compErr)
			continue
		}

		// Byte metrics
		origBytes := len(original)
		compBytes := len(compressed)
		var byteRatio float64
		if origBytes > 0 {
			byteRatio = 1.0 - float64(compBytes)/float64(origBytes)
		}
		if byteRatio < 0 {
			byteRatio = 0
		}

		// Token metrics using tiktoken
		origTokens := countTokens(original)
		compTokens := countTokens(compressed)
		var tokenRatio float64
		if origTokens > 0 {
			tokenRatio = 1.0 - float64(compTokens)/float64(origTokens)
		}
		if tokenRatio < 0 {
			tokenRatio = 0
		}

		// System prompt overhead
		sysPrompt := compressor.SystemPrompt()
		sysPromptTokens := countTokens(sysPrompt)

		// Net savings
		netSavings := origTokens - compTokens - sysPromptTokens
		var netRatio float64
		if origTokens > 0 {
			netRatio = float64(netSavings) / float64(origTokens)
		}

		// Determine if ShouldCompress would have applied compression
		decision := core.ShouldCompress(compressor, original, core.DefaultThresholds)

		results = append(results, BenchmarkResult{
			Name:               tf.Name,
			Domain:             tf.Domain,
			Category:           tf.Category,
			OriginalBytes:      origBytes,
			CompressedBytes:    compBytes,
			ByteRatio:          byteRatio,
			OriginalTokens:     origTokens,
			CompressedTokens:   compTokens,
			TokenRatio:         tokenRatio,
			SystemPromptTokens: sysPromptTokens,
			NetTokenSavings:    netSavings,
			NetTokenRatio:      netRatio,
			CompressTime:       elapsed,
			WasCompressed:      decision.WasCompressed,
		})
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no benchmark files found in %s", testDataDir)
	}

	return results, nil
}

// PrintResults prints a formatted table of benchmark results to stdout.
func PrintResults(results []BenchmarkResult) {
	if len(results) == 0 {
		fmt.Println("No benchmark results to display.")
		return
	}

	// Header
	fmt.Println(strings.Repeat("=", 120))
	fmt.Println("UCCP Compression Benchmark Results")
	fmt.Println(strings.Repeat("=", 120))
	fmt.Printf("%-20s %-8s %-10s %10s %10s %8s %10s %10s %8s %8s %10s\n",
		"Name", "Domain", "Category",
		"Orig(B)", "Comp(B)", "Byte%",
		"OrigTok", "CompTok", "Tok%",
		"NetTok", "Time",
	)
	fmt.Println(strings.Repeat("-", 120))

	// Rows
	var totalOrigBytes, totalCompBytes int
	var totalOrigTokens, totalCompTokens int
	var totalNetSavings int

	for _, r := range results {
		compressed := "  "
		if r.WasCompressed {
			compressed = "Y "
		}

		fmt.Printf("%-20s %-8s %-10s %10d %10d %7.1f%% %10d %10d %7.1f%% %+8d %10s %s\n",
			truncate(r.Name, 20),
			r.Domain,
			r.Category,
			r.OriginalBytes,
			r.CompressedBytes,
			r.ByteRatio*100,
			r.OriginalTokens,
			r.CompressedTokens,
			r.TokenRatio*100,
			r.NetTokenSavings,
			r.CompressTime.Round(time.Microsecond),
			compressed,
		)

		totalOrigBytes += r.OriginalBytes
		totalCompBytes += r.CompressedBytes
		totalOrigTokens += r.OriginalTokens
		totalCompTokens += r.CompressedTokens
		totalNetSavings += r.NetTokenSavings
	}

	// Summary
	fmt.Println(strings.Repeat("-", 120))

	var totalByteRatio, totalTokenRatio float64
	if totalOrigBytes > 0 {
		totalByteRatio = 1.0 - float64(totalCompBytes)/float64(totalOrigBytes)
	}
	if totalOrigTokens > 0 {
		totalTokenRatio = 1.0 - float64(totalCompTokens)/float64(totalOrigTokens)
	}

	fmt.Printf("%-20s %-8s %-10s %10d %10d %7.1f%% %10d %10d %7.1f%% %+8d\n",
		"TOTAL", "", "",
		totalOrigBytes,
		totalCompBytes,
		totalByteRatio*100,
		totalOrigTokens,
		totalCompTokens,
		totalTokenRatio*100,
		totalNetSavings,
	)
	fmt.Println(strings.Repeat("=", 120))

	// Token cost estimate
	fmt.Printf("\nEstimated monthly savings at 1000 calls/day: $%.2f\n",
		core.CalculateCostSavings(totalNetSavings*1000))
}

// truncate shortens a string to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
