package core

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// testCompressor implements Compressor for testing purposes.
// It "compresses" by removing vowels.
type testCompressor struct{}

func (t *testCompressor) Compress(content string) (string, error) {
	result := strings.Map(func(r rune) rune {
		if strings.ContainsRune("aeiouAEIOU", r) {
			return -1
		}
		return r
	}, content)
	return result, nil
}

func (t *testCompressor) Decompress(compressed string) (string, error) {
	return compressed, nil
}

func (t *testCompressor) SystemPrompt() string {
	return "test prompt"
}

func (t *testCompressor) EstimateTokens(content string) int {
	return EstimateTokenCount(content)
}

// errorCompressor always returns an error on Compress.
type errorCompressor struct {
	testCompressor
}

func (e *errorCompressor) Compress(content string) (string, error) {
	return "", fmt.Errorf("compression failed")
}

// --- ShouldCompress tests ---

func TestShouldCompress_ContentTooSmall(t *testing.T) {
	comp := &testCompressor{}
	content := "short"

	result := ShouldCompress(comp, content, DefaultThresholds)

	if result.WasCompressed {
		t.Error("expected WasCompressed=false for small content")
	}
	if result.Compressed != content {
		t.Error("expected Compressed to equal original for small content")
	}
}

func TestShouldCompress_CompressesWell(t *testing.T) {
	comp := &testCompressor{}
	// Generate content with lots of vowels so removing them saves > 30%
	content := strings.Repeat("aeiou testing content here ", 20)

	result := ShouldCompress(comp, content, DefaultThresholds)

	if !result.WasCompressed {
		t.Errorf("expected WasCompressed=true, ratio was %f", result.Ratio)
	}
	if result.Ratio < DefaultThresholds.MinRatio {
		t.Errorf("expected ratio >= %f, got %f", DefaultThresholds.MinRatio, result.Ratio)
	}
	if result.CompressedSize >= result.OriginalSize {
		t.Errorf("expected CompressedSize < OriginalSize, got %d >= %d", result.CompressedSize, result.OriginalSize)
	}
	// With tiktoken, vowel removal may not reduce token count (gibberish
	// fragments can tokenize into more pieces). Just verify the field is set
	// consistently with EstimateTokenSavings.
	expectedSavings := EstimateTokenSavings(content, result.Compressed)
	if result.EstimatedTokenSavings != expectedSavings {
		t.Errorf("expected EstimatedTokenSavings=%d, got %d", expectedSavings, result.EstimatedTokenSavings)
	}
}

func TestShouldCompress_DoesNotMeetMinRatio(t *testing.T) {
	comp := &testCompressor{}
	// Content with very few vowels so removal barely saves anything
	content := strings.Repeat("bcdfghjklmnpqrstvwxyz ", 15)

	result := ShouldCompress(comp, content, DefaultThresholds)

	if result.WasCompressed {
		t.Errorf("expected WasCompressed=false when ratio is too low, ratio was %f", result.Ratio)
	}
}

func TestShouldCompress_CompressionError(t *testing.T) {
	comp := &errorCompressor{}
	content := strings.Repeat("This is some content to compress. ", 10)

	result := ShouldCompress(comp, content, DefaultThresholds)

	if result.WasCompressed {
		t.Error("expected WasCompressed=false when compressor returns error")
	}
	if result.Compressed != content {
		t.Error("expected Compressed to equal original when compressor errors")
	}
}

// --- WriteMessage / ReadMessage tests ---

func tmpBasePath(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "uccp-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir + "/message"
}

func TestWriteMessage_Compressed(t *testing.T) {
	comp := &testCompressor{}
	basePath := tmpBasePath(t)
	// Content with lots of vowels to ensure compression
	content := strings.Repeat("aeiou testing content here ", 20)

	path, result, err := WriteMessage(comp, content, basePath, DefaultThresholds)

	if err != nil {
		t.Fatalf("WriteMessage error: %v", err)
	}
	if !result.WasCompressed {
		t.Fatal("expected compressed result")
	}
	if !strings.HasSuffix(path, ".uccp") {
		t.Errorf("expected .uccp extension, got %s", path)
	}

	// Verify file exists and has compressed content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(data) != result.Compressed {
		t.Error("file content does not match compressed result")
	}
}

func TestWriteMessage_Uncompressed(t *testing.T) {
	comp := &testCompressor{}
	basePath := tmpBasePath(t)
	content := "tiny"

	path, result, err := WriteMessage(comp, content, basePath, DefaultThresholds)

	if err != nil {
		t.Fatalf("WriteMessage error: %v", err)
	}
	if result.WasCompressed {
		t.Fatal("expected uncompressed result for small content")
	}
	if !strings.HasSuffix(path, ".txt") {
		t.Errorf("expected .txt extension, got %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(data) != content {
		t.Error("file content does not match original content")
	}
}

func TestReadMessage_CompressedFile(t *testing.T) {
	comp := &testCompressor{}
	basePath := tmpBasePath(t)
	content := strings.Repeat("aeiou testing content here ", 20)

	_, _, err := WriteMessage(comp, content, basePath, DefaultThresholds)
	if err != nil {
		t.Fatalf("WriteMessage error: %v", err)
	}

	readContent, wasCompressed, prompt, err := ReadMessage(comp, basePath)
	if err != nil {
		t.Fatalf("ReadMessage error: %v", err)
	}
	if !wasCompressed {
		t.Error("expected wasCompressed=true")
	}
	if prompt != "test prompt" {
		t.Errorf("expected system prompt 'test prompt', got %q", prompt)
	}
	if readContent == "" {
		t.Error("expected non-empty content")
	}
}

func TestReadMessage_UncompressedFile(t *testing.T) {
	comp := &testCompressor{}
	basePath := tmpBasePath(t)
	content := "small"

	_, _, err := WriteMessage(comp, content, basePath, DefaultThresholds)
	if err != nil {
		t.Fatalf("WriteMessage error: %v", err)
	}

	readContent, wasCompressed, prompt, err := ReadMessage(comp, basePath)
	if err != nil {
		t.Fatalf("ReadMessage error: %v", err)
	}
	if wasCompressed {
		t.Error("expected wasCompressed=false")
	}
	if prompt != "" {
		t.Errorf("expected empty prompt, got %q", prompt)
	}
	if readContent != content {
		t.Errorf("expected %q, got %q", content, readContent)
	}
}

func TestReadMessage_FileNotFound(t *testing.T) {
	comp := &testCompressor{}

	_, _, _, err := ReadMessage(comp, "/tmp/uccp-test-nonexistent-path-xyz")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
	if !strings.Contains(err.Error(), "message not found") {
		t.Errorf("expected 'message not found' error, got: %v", err)
	}
}

// --- WriteMessageWithStats tests ---

func TestWriteMessageWithStats_UpdatesStats(t *testing.T) {
	comp := &testCompressor{}
	basePath := tmpBasePath(t)
	content := strings.Repeat("aeiou testing content here ", 20)
	stats := &CompressionStats{}

	path, result, err := WriteMessageWithStats(comp, content, basePath, DefaultThresholds, stats)

	if err != nil {
		t.Fatalf("WriteMessageWithStats error: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty path")
	}
	if !result.WasCompressed {
		t.Fatal("expected compressed result for stats test")
	}
	if stats.TotalCompressions != 1 {
		t.Errorf("expected TotalCompressions=1, got %d", stats.TotalCompressions)
	}
	if stats.SuccessfulCompressions != 1 {
		t.Errorf("expected SuccessfulCompressions=1, got %d", stats.SuccessfulCompressions)
	}
	if stats.TotalBytesSaved <= 0 {
		t.Errorf("expected positive TotalBytesSaved, got %d", stats.TotalBytesSaved)
	}
	if stats.BestRatio <= 0 {
		t.Errorf("expected positive BestRatio, got %f", stats.BestRatio)
	}
}
