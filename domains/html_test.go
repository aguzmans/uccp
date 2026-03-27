package domains

import (
	"strings"
	"testing"
)

func TestHTMLCompressor_Compress_Headings(t *testing.T) {
	h := NewHTMLCompressor()
	html := `<h1>Getting Started</h1><h2>Installation Guide</h2><h3>Requirements</h3>`

	result, err := h.Compress(html)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "# Getting Started") {
		t.Errorf("expected H1 markdown, got: %s", result)
	}
	if !strings.Contains(result, "## Installation Guide") {
		t.Errorf("expected H2 markdown, got: %s", result)
	}
	if !strings.Contains(result, "### Requirements") {
		t.Errorf("expected H3 markdown, got: %s", result)
	}
}

func TestHTMLCompressor_Compress_Paragraphs(t *testing.T) {
	h := NewHTMLCompressor()
	html := `<p>This is a paragraph about the database configuration for production environments.</p>`

	result, err := h.Compress(html)
	if err != nil {
		t.Fatal(err)
	}

	// Should compress "configuration" -> "cfg", "production" -> "prod", etc.
	if !strings.Contains(result, "cfg") {
		t.Errorf("expected 'configuration' to be abbreviated, got: %s", result)
	}
	if !strings.Contains(result, "prod") {
		t.Errorf("expected 'production' to be abbreviated, got: %s", result)
	}
	// Articles should be removed
	if strings.Contains(result, " the ") || strings.Contains(result, " a ") {
		t.Errorf("expected articles to be removed, got: %s", result)
	}
}

func TestHTMLCompressor_Compress_CodeBlocks(t *testing.T) {
	h := NewHTMLCompressor()
	html := `<pre><code class="language-python">def hello():
    return True</code></pre>`

	result, err := h.Compress(html)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "```python") {
		t.Errorf("expected python code block, got: %s", result)
	}
	if !strings.Contains(result, "def hello") {
		t.Errorf("expected code content preserved, got: %s", result)
	}
}

func TestHTMLCompressor_Compress_Lists(t *testing.T) {
	h := NewHTMLCompressor()
	html := `<ul><li>First item</li><li>Second item</li></ul>`

	result, err := h.Compress(html)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "- First item") {
		t.Error("expected markdown list item format")
	}
	if !strings.Contains(result, "- Second item") {
		t.Error("expected markdown list item format")
	}
}

func TestHTMLCompressor_Compress_Tables(t *testing.T) {
	h := NewHTMLCompressor()
	html := `<table><tr><th>Name</th><th>Value</th></tr><tr><td>CPU</td><td>95%</td></tr></table>`

	result, err := h.Compress(html)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "| Name | Value |") {
		t.Errorf("expected markdown table header, got: %s", result)
	}
	if !strings.Contains(result, "| CPU | 95% |") {
		t.Errorf("expected markdown table row, got: %s", result)
	}
}

func TestHTMLCompressor_Compress_Links(t *testing.T) {
	h := NewHTMLCompressor()
	html := `<a href="https://example.com/docs">Documentation</a>`

	result, err := h.Compress(html)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "[Documentation](https://example.com/docs)") {
		t.Errorf("expected markdown link format, got: %s", result)
	}
}

func TestHTMLCompressor_Compress_NoiseRemoval(t *testing.T) {
	h := NewHTMLCompressor()
	html := `<nav><a href="/">Home</a></nav><script>alert('x')</script><h2>Real Content</h2><style>.x{}</style><footer>Copyright</footer>`

	result, err := h.Compress(html)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "## Real Content") {
		t.Errorf("expected heading to survive noise removal, got: %s", result)
	}
	if strings.Contains(result, "alert") {
		t.Error("script content should be removed")
	}
	if strings.Contains(result, "Copyright") {
		t.Error("footer content should be removed")
	}
	if strings.Contains(result, "Home") {
		t.Error("nav content should be removed")
	}
}

func TestHTMLCompressor_Compress_Abbreviations(t *testing.T) {
	h := NewHTMLCompressor()
	html := `<p>The authentication implementation uses the database configuration for the production environment.</p>`

	result, err := h.Compress(html)
	if err != nil {
		t.Fatal(err)
	}

	// Articles should be removed
	if strings.Contains(result, " the ") || strings.Contains(result, " a ") {
		t.Errorf("articles should be removed, got: %s", result)
	}
	// Check abbreviations
	if !strings.Contains(result, "auth") {
		t.Errorf("expected 'authentication' abbreviated to 'auth', got: %s", result)
	}
	if !strings.Contains(result, "impl") {
		t.Errorf("expected 'implementation' abbreviated to 'impl', got: %s", result)
	}
	if !strings.Contains(result, "cfg") {
		t.Errorf("expected 'configuration' abbreviated to 'cfg', got: %s", result)
	}
	if !strings.Contains(result, "prod") {
		t.Errorf("expected 'production' abbreviated to 'prod', got: %s", result)
	}
}

func TestHTMLCompressor_Compress_HTMLEntities(t *testing.T) {
	h := NewHTMLCompressor()
	html := `<p>Use &amp; for the ampersand. Quotes: &quot;hello&quot; and &lt;tag&gt;</p>`

	result, err := h.Compress(html)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(result, "&amp;") || strings.Contains(result, "&quot;") {
		t.Errorf("HTML entities should be decoded, got: %s", result)
	}
	if !strings.Contains(result, "&") || !strings.Contains(result, "\"") {
		t.Errorf("entities should be decoded to actual characters, got: %s", result)
	}
}

func TestHTMLCompressor_Compress_Empty(t *testing.T) {
	h := NewHTMLCompressor()
	result, err := h.Compress("")
	if err != nil {
		t.Fatal(err)
	}
	if result != "" {
		t.Errorf("expected empty result for empty input, got: %s", result)
	}
}

func TestHTMLCompressor_Decompress(t *testing.T) {
	h := NewHTMLCompressor()
	markdown := "# Title\n## Section\nSome text\n- Item one"

	result, err := h.Decompress(markdown)
	if err != nil {
		t.Fatal(err)
	}

	// Decompress just returns the markdown as-is (it's already readable)
	if result != markdown {
		t.Errorf("Decompress should return markdown as-is, got: %s", result)
	}
}

func TestHTMLCompressor_SystemPrompt(t *testing.T) {
	h := NewHTMLCompressor()
	prompt := h.SystemPrompt()

	if !strings.Contains(prompt, "markdown") {
		t.Error("SystemPrompt should mention markdown")
	}
	if !strings.Contains(prompt, "HTML") {
		t.Error("SystemPrompt should mention HTML")
	}
}

func TestHTMLCompressor_EstimateTokens(t *testing.T) {
	h := NewHTMLCompressor()
	tokens := h.EstimateTokens("Hello world, this is a test")
	if tokens <= 0 {
		t.Error("expected positive token estimate")
	}
}

func TestHTMLCompressor_RealWorldPage(t *testing.T) {
	h := NewHTMLCompressor()
	html := `
	<html>
	<head><title>Docker Networking</title></head>
	<body>
	<nav><a href="/">Home</a><a href="/docs">Docs</a></nav>
	<h1>Docker Networking Guide</h1>
	<h2>Overview</h2>
	<p>Docker provides several network drivers for container communication. The default bridge network is suitable for development environments.</p>
	<h2>Configuration</h2>
	<p>The database configuration requires authentication parameters in the production environment.</p>
	<pre><code class="language-bash">docker network create mynet</code></pre>
	<ul>
		<li>Bridge: default driver</li>
		<li>Host: removes network isolation</li>
	</ul>
	<table>
		<tr><th>Driver</th><th>Scope</th></tr>
		<tr><td>bridge</td><td>local</td></tr>
	</table>
	<footer>Copyright 2026</footer>
	</body>
	</html>`

	result, err := h.Compress(html)
	if err != nil {
		t.Fatal(err)
	}

	// Verify compression reduces size
	if len(result) >= len(html) {
		t.Errorf("compressed (%d bytes) should be smaller than original (%d bytes)", len(result), len(html))
	}

	// Verify key content is preserved
	if !strings.Contains(result, "# Docker Networking Guide") {
		t.Error("H1 should be preserved as markdown")
	}
	if !strings.Contains(result, "## Overview") {
		t.Error("H2 should be preserved as markdown")
	}
	if !strings.Contains(result, "```bash") && !strings.Contains(result, "docker network create") {
		t.Error("bash code block should be preserved")
	}

	// Verify abbreviations applied
	if !strings.Contains(result, "cfg") {
		t.Error("'configuration' should be abbreviated to 'cfg'")
	}
	if !strings.Contains(result, "auth") {
		t.Error("'authentication' should be abbreviated to 'auth'")
	}
	if !strings.Contains(result, "dev") {
		t.Error("'development' should be abbreviated to 'dev'")
	}

	// Nav and footer should be stripped
	if strings.Contains(result, "Home") || strings.Contains(result, "Docs") {
		t.Error("nav content should be removed")
	}
	if strings.Contains(result, "Copyright") {
		t.Error("footer content should be removed")
	}
}

// NEW TESTS FOR WHITESPACE FIX

func TestHTMLCompressor_EmptyFormattingRemoval(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string // strings that should be present
		avoid []string // strings that should NOT be present
	}{
		{
			name:  "removes empty bold markers",
			input: `<p>Text</p><strong></strong><p>More</p>`,
			want:  []string{"Text", "More"},
			avoid: []string{"**\n\n**", "**\n**"},
		},
		{
			name:  "removes navigation icons",
			input: `<nav><span>*</span><span>/</span></nav><p>Content</p>`,
			want:  []string{"Content"},
			avoid: []string{"*\n", "/\n", "*\n\n/"},
		},
		{
			name:  "removes standalone pipes and slashes",
			input: `<div>Before</div><span>|</span><div>After</div>`,
			want:  []string{"Before", "After"},
			avoid: []string{"|\n", "Before\n|\n"},
		},
		{
			name:  "preserves list markers",
			input: `<ul><li>Item one</li><li>Item two</li></ul>`,
			want:  []string{"- Item one", "- Item two"},
			avoid: []string{},
		},
		{
			name:  "preserves legitimate single asterisk in content",
			input: `<p>Price: $5*</p><p>*Taxes not included</p>`,
			want:  []string{"$5*", "*Taxes not included"},
			avoid: []string{},
		},
		{
			name:  "removes Britannica navigation noise",
			input: `<nav>**</nav><div>*</div><span>/</span><p>Real content</p>`,
			want:  []string{"Real content"},
			avoid: []string{"\n**\n", "\n*\n\n", "\n/\n"},
		},
		{
			name: "collapses multiple blanks after cleanup",
			input: `<p>Para 1</p>
				<nav>*</nav>
				<nav>**</nav>
				<nav>/</nav>
				<p>Para 2</p>`,
			want:  []string{"Para 1", "Para 2"},
			avoid: []string{"\n\n\n", "**", "*\n", "/\n"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewHTMLCompressor()
			got, err := c.Compress(tt.input)
			if err != nil {
				t.Fatalf("Compress() error = %v", err)
			}

			// Check for required strings
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("result should contain %q, got: %s", want, got)
				}
			}

			// Check that unwanted strings are absent
			for _, avoid := range tt.avoid {
				if strings.Contains(got, avoid) {
					t.Errorf("result should NOT contain %q, got: %s", avoid, got)
				}
			}
		})
	}
}

func TestHTMLCompressor_RealWorldBritannica(t *testing.T) {
	h := NewHTMLCompressor()
	// Simulated Britannica HTML with navigation noise
	html := `<html>
	<head><title>2026 Iran War</title></head>
	<body>
	<nav>
		<strong></strong>
		<span>*</span>
		<span>/</span>
		<a href="/search">Search Britannica</a>
		<span>*</span>
		<span>*</span>
	</nav>
	<header>
		<div>**</div>
		<div>/</div>
	</header>
	<h1>2026 Iran War</h1>
	<h2>Overview</h2>
	<p>The conflict began in early 2026 with military tensions.</p>
	<footer>Copyright Encyclopedia Britannica</footer>
	</body>
	</html>`

	result, err := h.Compress(html)
	if err != nil {
		t.Fatal(err)
	}

	// Should have clean content
	if !strings.Contains(result, "# 2026 Iran War") {
		t.Error("H1 should be preserved")
	}
	if !strings.Contains(result, "## Overview") {
		t.Error("H2 should be preserved")
	}
	if !strings.Contains(result, "conflict began in early 2026") {
		t.Error("paragraph content should be preserved")
	}

	// Should NOT have garbage navigation markers
	if strings.Contains(result, "**\n") {
		t.Error("empty bold markers should be removed")
	}
	if strings.Contains(result, "*\n") && !strings.Contains(result, "- ") {
		t.Error("standalone asterisks should be removed (except list markers)")
	}
	if strings.Contains(result, "/\n") {
		t.Error("standalone slashes should be removed")
	}

	// Count consecutive blank lines (should be max 2: \n\n)
	if strings.Contains(result, "\n\n\n") {
		t.Error("should not have 3+ consecutive blank lines")
	}

	// Footer should be removed
	if strings.Contains(result, "Copyright") || strings.Contains(result, "Encyclopedia") {
		t.Error("footer content should be removed")
	}
}
