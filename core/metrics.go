package core

import (
	"math"
	"strings"
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

// EstimateTokenCount estimates the number of tokens in content
// Uses a simple heuristic: ~4 characters per token for English text
// This is approximate - actual tokenization varies by model
func EstimateTokenCount(content string) int {
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
