package core

import (
	"math"
	"strings"
	"sync"

	"github.com/pkoukk/tiktoken-go"
)

// CalculateCompressionRatio returns the compression ratio achieved
// Returns a value between 0.0 and 1.0
// Example: 0.7 means 70% compression (compressed to 30% of original size)
func CalculateCompressionRatio(original, compressed string) float64 {
	origSize := len(original)
	compSize := len(compressed)

	if origSize == 0 {
		return 0.0
	}

	ratio := 1.0 - (float64(compSize) / float64(origSize))

	// Clamp between 0 and 1
	if ratio < 0 {
		return 0.0
	}
	if ratio > 1 {
		return 1.0
	}

	return ratio
}

var (
	tiktokenEncoder *tiktoken.Tiktoken
	tiktokenOnce    sync.Once
)

func initTiktoken() {
	enc, err := tiktoken.GetEncoding("cl100k_base")
	if err == nil {
		tiktokenEncoder = enc
	}
}

// EstimateTokenCount estimates the number of tokens in content using tiktoken
// cl100k_base encoding for accurate counts. Falls back to ~4 chars/token
// heuristic if tiktoken fails to initialize.
func EstimateTokenCount(content string) int {
	tiktokenOnce.Do(initTiktoken)
	if tiktokenEncoder != nil {
		content = strings.TrimSpace(content)
		if len(content) == 0 {
			return 0
		}
		return len(tiktokenEncoder.Encode(content, nil, nil))
	}
	// Fallback to heuristic
	return EstimateTokenCountHeuristic(content)
}

// EstimateTokenCountHeuristic estimates tokens using a simple ~4 characters
// per token heuristic. Useful for deterministic testing or when tiktoken
// is unavailable.
func EstimateTokenCountHeuristic(content string) int {
	const charsPerToken = 4.0

	// Remove extra whitespace
	content = strings.TrimSpace(content)

	// Count characters
	charCount := len(content)

	// Estimate tokens
	tokens := int(math.Ceil(float64(charCount) / charsPerToken))

	return tokens
}

// EstimateTokenSavings calculates approximate tokens saved by compression
func EstimateTokenSavings(original, compressed string) int {
	originalTokens := EstimateTokenCount(original)
	compressedTokens := EstimateTokenCount(compressed)

	savings := originalTokens - compressedTokens

	if savings < 0 {
		return 0
	}

	return savings
}

// NetTokenSavings calculates actual token savings after accounting for
// the system prompt overhead that must be included when sending UCCP to an LLM.
// Returns negative if compression actually costs more tokens than it saves.
func NetTokenSavings(originalTokens, compressedTokens, systemPromptTokens int) int {
	return originalTokens - compressedTokens - systemPromptTokens
}

// CalculateCostSavings estimates monthly cost savings
// Assumes Claude API pricing: $3 per million input tokens
func CalculateCostSavings(tokensSavedPerDay int) float64 {
	const costPerMillionTokens = 3.0
	const daysPerMonth = 30

	tokensPerMonth := float64(tokensSavedPerDay * daysPerMonth)
	costSavings := (tokensPerMonth / 1_000_000.0) * costPerMillionTokens

	return costSavings
}

// UpdateStats updates compression statistics with a new result
func UpdateStats(stats *CompressionStats, result *CompressionResult) {
	stats.TotalCompressions++

	if result.WasCompressed {
		stats.SuccessfulCompressions++
		stats.TotalBytesSaved += int64(result.OriginalSize - result.CompressedSize)
		stats.TotalTokensSaved += int64(result.EstimatedTokenSavings)

		// Update average ratio
		n := float64(stats.SuccessfulCompressions)
		stats.AverageRatio = ((stats.AverageRatio * (n - 1.0)) + result.Ratio) / n

		// Update best/worst ratios
		if result.Ratio > stats.BestRatio {
			stats.BestRatio = result.Ratio
		}
		if stats.WorstRatio == 0 || result.Ratio < stats.WorstRatio {
			stats.WorstRatio = result.Ratio
		}
	} else {
		stats.SkippedCompressions++
	}
}
