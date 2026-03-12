package core

import (
	"fmt"
	"os"
	"path/filepath"
)

// CompressionThresholds define when to apply compression
type CompressionThresholds struct {
	// MinSize is the minimum content size (in bytes) to consider compression
	// Below this size, compression overhead isn't worth it
	MinSize int

	// MinRatio is the minimum compression ratio required to use compression
	// Example: 0.30 means compression must save at least 30% to be applied
	MinRatio float64

	// SystemPromptOverhead is the estimated token cost of the UCCP system prompt.
	// When set, ShouldCompress will factor this into the net savings calculation.
	// If 0, system prompt cost is ignored (backward-compatible default).
	SystemPromptOverhead int
}

// DefaultThresholds provide sensible defaults for most use cases
var DefaultThresholds = CompressionThresholds{
	MinSize:  200,  // Don't compress messages < 200 bytes
	MinRatio: 0.30, // Require 30% savings minimum
}

// AggressiveThresholds compress more aggressively
var AggressiveThresholds = CompressionThresholds{
	MinSize:  100,  // Compress smaller messages
	MinRatio: 0.20, // Accept 20% savings
}

// ConservativeThresholds only compress when very beneficial
var ConservativeThresholds = CompressionThresholds{
	MinSize:  500,  // Only compress large messages
	MinRatio: 0.50, // Require 50% savings
}

// ShouldCompress intelligently decides whether compression will save tokens
// Returns the compression result with metadata
func ShouldCompress(compressor Compressor, content string, thresholds CompressionThresholds) *CompressionResult {
	originalSize := len(content)

	result := &CompressionResult{
		Original:       content,
		Compressed:     content,
		WasCompressed:  false,
		Ratio:          0.0,
		OriginalSize:   originalSize,
		CompressedSize: originalSize,
	}

	// Check 1: Is content too small to benefit from compression?
	if originalSize < thresholds.MinSize {
		return result
	}

	// Check 2: Try compression
	compressed, err := compressor.Compress(content)
	if err != nil {
		// Compression failed, return original
		return result
	}

	compressedSize := len(compressed)
	ratio := CalculateCompressionRatio(content, compressed)

	// Check 3: Did compression achieve minimum savings?
	if ratio < thresholds.MinRatio {
		// Not enough savings, use original
		return result
	}

	// Check 4: Factor in system prompt overhead if configured
	if thresholds.SystemPromptOverhead > 0 {
		netSavings := NetTokenSavings(
			EstimateTokenCount(content),
			EstimateTokenCount(compressed),
			thresholds.SystemPromptOverhead,
		)
		if netSavings <= 0 {
			// System prompt overhead exceeds compression savings
			return result
		}
	}

	// Compression is beneficial!
	result.Compressed = compressed
	result.WasCompressed = true
	result.Ratio = ratio
	result.CompressedSize = compressedSize
	result.EstimatedTokenSavings = EstimateTokenSavings(content, compressed)
	result.NetTokenSavings = NetTokenSavings(
		EstimateTokenCount(content),
		EstimateTokenCount(compressed),
		thresholds.SystemPromptOverhead,
	)

	return result
}

// WriteMessage intelligently writes content, compressing only if beneficial
// This is the primary API for agent-to-agent communication
//
// Parameters:
//   - compressor: The compressor to use (code, HTML, etc.)
//   - content: The message content to write
//   - basePath: Base file path without extension
//   - thresholds: Compression decision thresholds
//
// Returns:
//   - writtenPath: Actual path written (.uccp or .txt)
//   - result: Compression result with metadata
//   - error: Any error during writing
//
// Example:
//
//	compressor := domains.NewCodeCompressor()
//	path, result, err := WriteMessage(compressor, jobSummary, "job-001", DefaultThresholds)
//	// Creates either job-001.uccp (compressed) or job-001.txt (plain)
func WriteMessage(compressor Compressor, content string, basePath string, thresholds CompressionThresholds) (string, *CompressionResult, error) {
	result := ShouldCompress(compressor, content, thresholds)

	var path string
	var data string

	if result.WasCompressed {
		path = basePath + ".uccp"
		data = result.Compressed
	} else {
		path = basePath + ".txt"
		data = result.Original
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", result, fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		return "", result, fmt.Errorf("failed to write file: %w", err)
	}

	return path, result, nil
}

// ReadMessage intelligently reads content, detecting compression automatically
// Tries .uccp extension first, then .txt, then no extension
//
// Parameters:
//   - basePath: Base file path without extension
//
// Returns:
//   - content: The message content (decompressed if needed)
//   - wasCompressed: Whether the file was in UCCP format
//   - systemPrompt: System prompt to include if compressed (empty if not)
//   - error: Any error during reading
//
// Example:
//
//	content, wasCompressed, prompt, err := ReadMessage("job-001")
//	if wasCompressed {
//	    // Include prompt when sending to LLM so it understands UCCP format
//	}
func ReadMessage(compressor Compressor, basePath string) (string, bool, string, error) {
	// Try compressed version first (.uccp)
	compressedPath := basePath + ".uccp"
	if data, err := os.ReadFile(compressedPath); err == nil {
		// File is compressed, return with system prompt
		return string(data), true, compressor.SystemPrompt(), nil
	}

	// Try plain text version (.txt)
	textPath := basePath + ".txt"
	if data, err := os.ReadFile(textPath); err == nil {
		return string(data), false, "", nil
	}

	// Try without extension (legacy support)
	if data, err := os.ReadFile(basePath); err == nil {
		return string(data), false, "", nil
	}

	return "", false, "", fmt.Errorf("message not found: %s", basePath)
}

// WriteMessageWithStats is like WriteMessage but also updates statistics
func WriteMessageWithStats(compressor Compressor, content string, basePath string, thresholds CompressionThresholds, stats *CompressionStats) (string, *CompressionResult, error) {
	path, result, err := WriteMessage(compressor, content, basePath, thresholds)

	if err == nil && stats != nil {
		UpdateStats(stats, result)
	}

	return path, result, err
}
