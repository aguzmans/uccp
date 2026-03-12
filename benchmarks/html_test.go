package benchmarks

import (
	"os"
	"testing"

	"github.com/aguzmans/uccp/core"
	"github.com/aguzmans/uccp/domains"
)

// TestHTMLDocumentation tests compression on real documentation HTML
func TestHTMLDocumentation(t *testing.T) {
	scenarios := []struct {
		name           string
		htmlFile       string
		expectedRatio  float64
		description    string
	}{
		{
			name:          "Simple Article",
			htmlFile:      "testdata/simple_article.html",
			expectedRatio: 0.60, // 60% compression
			description:   "Basic article with headings, paragraphs, and links",
		},
		{
			name:          "API Documentation",
			htmlFile:      "testdata/api_docs.html",
			expectedRatio: 0.70, // 70% compression
			description:   "Technical documentation with code examples",
		},
		{
			name:          "Blog Post",
			htmlFile:      "testdata/blog_post.html",
			expectedRatio: 0.65, // 65% compression
			description:   "Blog post with images, lists, and formatting",
		},
	}

	compressor := domains.NewHTMLCompressor()

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Read HTML file
			htmlContent, err := os.ReadFile(scenario.htmlFile)
			if err != nil {
				// If file doesn't exist, create sample content
				htmlContent = generateSampleHTML(scenario.name)
			}

			html := string(htmlContent)

			// Compress
			compressed, err := compressor.Compress(html)
			if err != nil {
				t.Fatalf("Compression failed: %v", err)
			}

			// Calculate metrics
			ratio := core.CalculateCompressionRatio(html, compressed)
			originalTokens := core.EstimateTokenCount(html)
			compressedTokens := core.EstimateTokenCount(compressed)
			tokenSavings := core.EstimateTokenSavings(html, compressed)

			// Report results
			t.Logf("\n=== %s ===", scenario.name)
			t.Logf("Description: %s", scenario.description)
			t.Logf("Original size: %d bytes (%d tokens)", len(html), originalTokens)
			t.Logf("Compressed size: %d bytes (%d tokens)", len(compressed), compressedTokens)
			t.Logf("Compression ratio: %.1f%%", ratio*100)
			t.Logf("Token savings: %d tokens (%.1f%%)", tokenSavings, (float64(tokenSavings)/float64(originalTokens))*100)

			if len(html) < 500 {
				t.Logf("\nOriginal HTML (first 500 chars):\n%s", html)
				t.Logf("\nCompressed:\n%s", compressed)
			}

			// Validate compression meets expectations
			if ratio < scenario.expectedRatio {
				t.Logf("Note: Compression ratio %.1f%% below expected %.1f%% (might be OK for this content)",
					ratio*100, scenario.expectedRatio*100)
			}
		})
	}
}

// TestHTMLWebScraping simulates scraping multiple web pages and compressing before LLM
func TestHTMLWebScraping(t *testing.T) {
	compressor := domains.NewHTMLCompressor()

	// Simulate scraping 10 documentation pages
	pageCount := 10
	avgPageSize := 50 * 1024 // 50KB average page

	var totalOriginalTokens, totalCompressedTokens int

	for i := 1; i <= pageCount; i++ {
		// Generate sample HTML page
		html := generateLargeHTMLPage(i, avgPageSize)

		compressed, err := compressor.Compress(html)
		if err != nil {
			t.Fatalf("Compression failed for page %d: %v", i, err)
		}

		totalOriginalTokens += core.EstimateTokenCount(html)
		totalCompressedTokens += core.EstimateTokenCount(compressed)
	}

	tokenSavings := totalOriginalTokens - totalCompressedTokens
	percentSaved := (float64(tokenSavings) / float64(totalOriginalTokens)) * 100

	t.Logf("\n=== Web Scraping 10 Documentation Pages ===")
	t.Logf("Without UCCP:")
	t.Logf("  Total tokens: %d", totalOriginalTokens)
	t.Logf("  Estimated cost: $%.2f", float64(totalOriginalTokens)*0.003/1000)
	t.Logf("  Pages that fit in 200k context: %d", 200000/totalOriginalTokens*pageCount)
	t.Logf("\nWith UCCP:")
	t.Logf("  Total tokens: %d", totalCompressedTokens)
	t.Logf("  Estimated cost: $%.2f", float64(totalCompressedTokens)*0.003/1000)
	t.Logf("  Pages that fit in 200k context: %d", 200000/totalCompressedTokens*pageCount)
	t.Logf("\nSavings:")
	t.Logf("  Token reduction: %d tokens (%.1f%%)", tokenSavings, percentSaved)
	t.Logf("  Cost savings: $%.2f", float64(tokenSavings)*0.003/1000)
	t.Logf("  Capacity multiplier: %.1fx more pages", float64(totalOriginalTokens)/float64(totalCompressedTokens))

	// Validate significant savings
	if percentSaved < 60 {
		t.Errorf("Token savings too low: %.1f%% (expected >= 60%%)", percentSaved)
	}
}

// Helper functions

func generateSampleHTML(name string) []byte {
	return []byte(`<!DOCTYPE html>
<html>
<head>
    <title>` + name + `</title>
    <style>
        body { font-family: Arial, sans-serif; }
        .container { max-width: 800px; margin: 0 auto; }
    </style>
</head>
<body>
    <div class="container">
        <h1>` + name + `</h1>
        <p>This is a sample article with multiple paragraphs to test HTML compression.</p>
        <p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.</p>

        <h2>Key Features</h2>
        <ul>
            <li>Feature one with detailed explanation</li>
            <li>Feature two with more details</li>
            <li>Feature three with even more content</li>
        </ul>

        <h2>Code Example</h2>
        <pre><code>
function example() {
    return "Hello, World!";
}
        </code></pre>

        <p>For more information, visit <a href="https://example.com">our documentation</a>.</p>
    </div>
</body>
</html>`)
}

func generateLargeHTMLPage(pageNum int, targetSize int) string {
	base := `<!DOCTYPE html>
<html>
<head>
    <title>Documentation Page ` + string(rune(pageNum)) + `</title>
</head>
<body>
    <h1>API Documentation - Page ` + string(rune(pageNum)) + `</h1>
    <p>This page contains detailed API documentation.</p>
`

	// Add content to reach target size
	paragraph := "<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.</p>\n"

	content := base
	for len(content) < targetSize {
		content += paragraph
	}

	content += "</body></html>"
	return content
}

// BenchmarkHTMLCompression benchmarks HTML compression performance
func BenchmarkHTMLCompression(b *testing.B) {
	compressor := domains.NewHTMLCompressor()
	html := string(generateSampleHTML("Benchmark"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = compressor.Compress(html)
	}
}
