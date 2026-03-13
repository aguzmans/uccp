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

// BenchmarkResult holds the results for a single scale benchmark
type BenchmarkResult struct {
	Name     string
	Domain   string
	Category string
	Pages    int

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

	// Net savings with system prompt amortized over DefaultAmortizationDepth messages
	AmortizedNetSavings int
	AmortizedNetRatio   float64

	// Performance
	CompressTime time.Duration
}

// DefaultAmortizationDepth is the number of messages over which the system
// prompt overhead is amortized when computing net savings.
const DefaultAmortizationDepth = 10

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
func countTokens(text string) int {
	enc, err := getTokenizer()
	if err != nil {
		log.Printf("WARNING: tiktoken unavailable, falling back to chars/4 estimate: %v", err)
		return len(text) / 4
	}
	tokens := enc.Encode(text, nil, nil)
	return len(tokens)
}

// compressorForDomain returns the appropriate compressor based on domain string.
// All compressors are wrapped with smart deduplication.
func compressorForDomain(domain string) core.Compressor {
	var inner core.Compressor
	switch domain {
	case "html":
		inner = domains.NewHTMLCompressor()
	case "json":
		inner = domains.NewJSONCompressor()
	default:
		inner = domains.NewCodeCompressor()
	}
	return core.NewDedupCompressor(inner, 40)
}

// RunBenchmarks iterates over all scale tests, reads each generated file,
// compresses with the appropriate domain compressor, and measures byte/token
// metrics using tiktoken for real token counts.
func RunBenchmarks(testDataDir string) ([]BenchmarkResult, error) {
	tests := ScaleBenchmarks()
	var results []BenchmarkResult

	for _, st := range tests {
		fullPath := filepath.Join(testDataDir, st.Path)

		content, err := os.ReadFile(fullPath)
		if err != nil {
			log.Printf("WARNING: skipping %s: %v", st.Name, err)
			continue
		}

		original := string(content)
		compressor := compressorForDomain(st.Domain)

		// Measure compression time
		start := time.Now()
		compressed, compErr := compressor.Compress(original)
		elapsed := time.Since(start)

		if compErr != nil {
			log.Printf("WARNING: compression failed for %s: %v", st.Name, compErr)
			continue
		}

		// Byte metrics
		origBytes := len(original)
		compBytes := len(compressed)
		var byteRatio float64
		if origBytes > 0 {
			byteRatio = 1.0 - float64(compBytes)/float64(origBytes)
		}

		// Token metrics using tiktoken
		origTokens := countTokens(original)
		compTokens := countTokens(compressed)
		var tokenRatio float64
		if origTokens > 0 {
			tokenRatio = 1.0 - float64(compTokens)/float64(origTokens)
		}

		// System prompt overhead
		sysPrompt := compressor.SystemPrompt()
		sysPromptTokens := countTokens(sysPrompt)

		// Net savings (full system prompt overhead)
		netSavings := origTokens - compTokens - sysPromptTokens
		var netRatio float64
		if origTokens > 0 {
			netRatio = float64(netSavings) / float64(origTokens)
		}

		// Amortized net savings (system prompt cost spread over N messages)
		amortizedOverhead := sysPromptTokens / DefaultAmortizationDepth
		amortizedNetSavings := origTokens - compTokens - amortizedOverhead
		var amortizedNetRatio float64
		if origTokens > 0 {
			amortizedNetRatio = float64(amortizedNetSavings) / float64(origTokens)
		}

		results = append(results, BenchmarkResult{
			Name:                st.Name,
			Domain:              st.Domain,
			Category:            st.Category,
			Pages:               st.Pages,
			OriginalBytes:       origBytes,
			CompressedBytes:     compBytes,
			ByteRatio:           byteRatio,
			OriginalTokens:      origTokens,
			CompressedTokens:    compTokens,
			TokenRatio:          tokenRatio,
			SystemPromptTokens:  sysPromptTokens,
			NetTokenSavings:     netSavings,
			NetTokenRatio:       netRatio,
			AmortizedNetSavings: amortizedNetSavings,
			AmortizedNetRatio:   amortizedNetRatio,
			CompressTime:        elapsed,
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

	fmt.Println(strings.Repeat("=", 130))
	fmt.Println("UCCP Compression Benchmark Results (tiktoken cl100k_base)")
	fmt.Println(strings.Repeat("=", 130))
	fmt.Printf("%-22s %5s %-18s %10s %10s %8s %10s %10s %8s %8s %10s\n",
		"Name", "Pages", "Category",
		"Orig(B)", "Comp(B)", "Byte%",
		"OrigTok", "CompTok", "Tok%",
		"NetTok", "Time",
	)
	fmt.Println(strings.Repeat("-", 130))

	var totalOrigBytes, totalCompBytes int
	var totalOrigTokens, totalCompTokens int
	var totalNetSavings int

	for _, r := range results {
		fmt.Printf("%-22s %5d %-18s %10d %10d %7.1f%% %10d %10d %7.1f%% %+8d %10s\n",
			truncate(r.Name, 22),
			r.Pages,
			truncate(r.Category, 18),
			r.OriginalBytes,
			r.CompressedBytes,
			r.ByteRatio*100,
			r.OriginalTokens,
			r.CompressedTokens,
			r.TokenRatio*100,
			r.NetTokenSavings,
			r.CompressTime.Round(time.Millisecond),
		)

		totalOrigBytes += r.OriginalBytes
		totalCompBytes += r.CompressedBytes
		totalOrigTokens += r.OriginalTokens
		totalCompTokens += r.CompressedTokens
		totalNetSavings += r.NetTokenSavings
	}

	fmt.Println(strings.Repeat("-", 130))

	var totalByteRatio, totalTokenRatio float64
	if totalOrigBytes > 0 {
		totalByteRatio = 1.0 - float64(totalCompBytes)/float64(totalOrigBytes)
	}
	if totalOrigTokens > 0 {
		totalTokenRatio = 1.0 - float64(totalCompTokens)/float64(totalOrigTokens)
	}

	fmt.Printf("%-22s %5s %-18s %10d %10d %7.1f%% %10d %10d %7.1f%% %+8d\n",
		"TOTAL", "", "",
		totalOrigBytes,
		totalCompBytes,
		totalByteRatio*100,
		totalOrigTokens,
		totalCompTokens,
		totalTokenRatio*100,
		totalNetSavings,
	)
	fmt.Println(strings.Repeat("=", 130))

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
