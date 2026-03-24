package core

import (
	"math"
	"testing"
)

// --- CalculateCompressionRatio tests ---

func TestCalculateCompressionRatio_Normal(t *testing.T) {
	original := "Hello, this is a test message with some content."
	compressed := "Hll, ths s tst mssg."

	ratio := CalculateCompressionRatio(original, compressed)

	if ratio <= 0.0 || ratio >= 1.0 {
		t.Errorf("expected ratio between 0 and 1, got %f", ratio)
	}

	// compressed is shorter, so ratio should be positive
	expectedRatio := 1.0 - (float64(len(compressed)) / float64(len(original)))
	if math.Abs(ratio-expectedRatio) > 0.001 {
		t.Errorf("expected ratio ~%f, got %f", expectedRatio, ratio)
	}
}

func TestCalculateCompressionRatio_EmptyInput(t *testing.T) {
	ratio := CalculateCompressionRatio("", "")
	if ratio != 0.0 {
		t.Errorf("expected 0.0 for empty input, got %f", ratio)
	}
}

func TestCalculateCompressionRatio_CompressedLargerThanOriginal(t *testing.T) {
	original := "hi"
	compressed := "this is much longer than the original"

	ratio := CalculateCompressionRatio(original, compressed)

	// Should be clamped to 0.0
	if ratio != 0.0 {
		t.Errorf("expected 0.0 when compressed is larger, got %f", ratio)
	}
}

func TestCalculateCompressionRatio_IdenticalStrings(t *testing.T) {
	content := "identical content"
	ratio := CalculateCompressionRatio(content, content)

	if ratio != 0.0 {
		t.Errorf("expected 0.0 for identical strings, got %f", ratio)
	}
}

// --- EstimateTokenCount tests ---

func TestEstimateTokenCount_NormalText(t *testing.T) {
	content := "Hello world testing!"
	tokens := EstimateTokenCount(content)

	if tokens <= 0 {
		t.Errorf("expected positive token count, got %d", tokens)
	}
}

func TestEstimateTokenCountHeuristic_NormalText(t *testing.T) {
	// 20 chars -> ceil(20/4) = 5 tokens
	content := "Hello world testing!"
	tokens := EstimateTokenCountHeuristic(content)

	if tokens <= 0 {
		t.Errorf("expected positive token count, got %d", tokens)
	}

	expected := int(math.Ceil(float64(len(content)) / 4.0))
	if tokens != expected {
		t.Errorf("expected %d tokens, got %d", expected, tokens)
	}
}

func TestEstimateTokenCount_EmptyString(t *testing.T) {
	tokens := EstimateTokenCount("")
	if tokens != 0 {
		t.Errorf("expected 0 tokens for empty string, got %d", tokens)
	}
}

func TestEstimateTokenCount_WhitespaceOnly(t *testing.T) {
	tokens := EstimateTokenCount("   \t\n  ")
	if tokens != 0 {
		t.Errorf("expected 0 tokens for whitespace-only string, got %d", tokens)
	}
}

func TestEstimateTokenCountHeuristic_EmptyString(t *testing.T) {
	tokens := EstimateTokenCountHeuristic("")
	if tokens != 0 {
		t.Errorf("expected 0 tokens for empty string, got %d", tokens)
	}
}

func TestEstimateTokenCountHeuristic_WhitespaceOnly(t *testing.T) {
	tokens := EstimateTokenCountHeuristic("   \t\n  ")
	if tokens != 0 {
		t.Errorf("expected 0 tokens for whitespace-only string, got %d", tokens)
	}
}

// --- EstimateTokenSavings tests ---

func TestEstimateTokenSavings_PositiveSavings(t *testing.T) {
	original := "This is a long original message with lots of content inside."
	compressed := "Ths s shrt."

	savings := EstimateTokenSavings(original, compressed)

	if savings <= 0 {
		t.Errorf("expected positive savings, got %d", savings)
	}

	expectedSavings := EstimateTokenCount(original) - EstimateTokenCount(compressed)
	if savings != expectedSavings {
		t.Errorf("expected %d savings, got %d", expectedSavings, savings)
	}
}

func TestEstimateTokenSavings_NoSavings(t *testing.T) {
	original := "hi"
	compressed := "this is actually much longer than before"

	savings := EstimateTokenSavings(original, compressed)

	if savings != 0 {
		t.Errorf("expected 0 savings when compressed is bigger, got %d", savings)
	}
}

// --- CalculateCostSavings tests ---

func TestCalculateCostSavings_Basic(t *testing.T) {
	// 1,000,000 tokens/day * 30 days = 30,000,000 tokens/month
	// 30,000,000 / 1,000,000 * $3 = $90
	savings := CalculateCostSavings(1_000_000)

	expected := 90.0
	if math.Abs(savings-expected) > 0.001 {
		t.Errorf("expected $%.2f, got $%.2f", expected, savings)
	}
}

func TestCalculateCostSavings_Zero(t *testing.T) {
	savings := CalculateCostSavings(0)
	if savings != 0.0 {
		t.Errorf("expected $0.00, got $%.2f", savings)
	}
}

// --- UpdateStats tests ---

func TestUpdateStats_SuccessfulCompression(t *testing.T) {
	stats := &CompressionStats{}
	result := &CompressionResult{
		WasCompressed:         true,
		OriginalSize:          1000,
		CompressedSize:        300,
		Ratio:                 0.70,
		EstimatedTokenSavings: 175,
	}

	UpdateStats(stats, result)

	if stats.TotalCompressions != 1 {
		t.Errorf("expected TotalCompressions=1, got %d", stats.TotalCompressions)
	}
	if stats.SuccessfulCompressions != 1 {
		t.Errorf("expected SuccessfulCompressions=1, got %d", stats.SuccessfulCompressions)
	}
	if stats.SkippedCompressions != 0 {
		t.Errorf("expected SkippedCompressions=0, got %d", stats.SkippedCompressions)
	}
	if stats.TotalBytesSaved != 700 {
		t.Errorf("expected TotalBytesSaved=700, got %d", stats.TotalBytesSaved)
	}
	if stats.TotalTokensSaved != 175 {
		t.Errorf("expected TotalTokensSaved=175, got %d", stats.TotalTokensSaved)
	}
	if stats.BestRatio != 0.70 {
		t.Errorf("expected BestRatio=0.70, got %f", stats.BestRatio)
	}
	if stats.WorstRatio != 0.70 {
		t.Errorf("expected WorstRatio=0.70, got %f", stats.WorstRatio)
	}
}

func TestUpdateStats_SkippedCompression(t *testing.T) {
	stats := &CompressionStats{}
	result := &CompressionResult{
		WasCompressed:  false,
		OriginalSize:   50,
		CompressedSize: 50,
		Ratio:          0.0,
	}

	UpdateStats(stats, result)

	if stats.TotalCompressions != 1 {
		t.Errorf("expected TotalCompressions=1, got %d", stats.TotalCompressions)
	}
	if stats.SuccessfulCompressions != 0 {
		t.Errorf("expected SuccessfulCompressions=0, got %d", stats.SuccessfulCompressions)
	}
	if stats.SkippedCompressions != 1 {
		t.Errorf("expected SkippedCompressions=1, got %d", stats.SkippedCompressions)
	}
	if stats.TotalBytesSaved != 0 {
		t.Errorf("expected TotalBytesSaved=0, got %d", stats.TotalBytesSaved)
	}
}

func TestUpdateStats_MultipleUpdates(t *testing.T) {
	stats := &CompressionStats{}

	// First successful compression: ratio 0.80
	UpdateStats(stats, &CompressionResult{
		WasCompressed:         true,
		OriginalSize:          1000,
		CompressedSize:        200,
		Ratio:                 0.80,
		EstimatedTokenSavings: 200,
	})

	// Second successful compression: ratio 0.50
	UpdateStats(stats, &CompressionResult{
		WasCompressed:         true,
		OriginalSize:          500,
		CompressedSize:        250,
		Ratio:                 0.50,
		EstimatedTokenSavings: 63,
	})

	// One skipped
	UpdateStats(stats, &CompressionResult{
		WasCompressed:  false,
		OriginalSize:   30,
		CompressedSize: 30,
	})

	if stats.TotalCompressions != 3 {
		t.Errorf("expected TotalCompressions=3, got %d", stats.TotalCompressions)
	}
	if stats.SuccessfulCompressions != 2 {
		t.Errorf("expected SuccessfulCompressions=2, got %d", stats.SuccessfulCompressions)
	}
	if stats.SkippedCompressions != 1 {
		t.Errorf("expected SkippedCompressions=1, got %d", stats.SkippedCompressions)
	}
	if stats.BestRatio != 0.80 {
		t.Errorf("expected BestRatio=0.80, got %f", stats.BestRatio)
	}
	if stats.WorstRatio != 0.50 {
		t.Errorf("expected WorstRatio=0.50, got %f", stats.WorstRatio)
	}

	// Average should be (0.80 + 0.50) / 2 = 0.65
	expectedAvg := 0.65
	if math.Abs(stats.AverageRatio-expectedAvg) > 0.001 {
		t.Errorf("expected AverageRatio ~%f, got %f", expectedAvg, stats.AverageRatio)
	}

	expectedBytesSaved := int64(800 + 250)
	if stats.TotalBytesSaved != expectedBytesSaved {
		t.Errorf("expected TotalBytesSaved=%d, got %d", expectedBytesSaved, stats.TotalBytesSaved)
	}

	expectedTokensSaved := int64(200 + 63)
	if stats.TotalTokensSaved != expectedTokensSaved {
		t.Errorf("expected TotalTokensSaved=%d, got %d", expectedTokensSaved, stats.TotalTokensSaved)
	}
}
