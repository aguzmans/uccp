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

	if !strings.Contains(result, "1:Getting Started") {
		t.Errorf("expected H1 record, got: %s", result)
	}
	if !strings.Contains(result, "2:Installation Guide") {
		t.Errorf("expected H2 record, got: %s", result)
	}
	if !strings.Contains(result, "3:Requirements") {
		t.Errorf("expected H3 record, got: %s", result)
	}
}

func TestHTMLCompressor_Compress_Paragraphs(t *testing.T) {
	h := NewHTMLCompressor()
	html := `<p>This is a paragraph about the database configuration for production environments.</p>`

	result, err := h.Compress(html)
	if err != nil {
		t.Fatal(err)
	}

	// Should compress "database" -> "db", "configuration" -> "cfg", etc.
	if !strings.Contains(result, "t:") {
		t.Errorf("expected text record, got: %s", result)
	}
	if strings.Contains(result, "database") {
		t.Errorf("expected 'database' to be abbreviated, got: %s", result)
	}
	if strings.Contains(result, "configuration") {
		t.Errorf("expected 'configuration' to be abbreviated, got: %s", result)
	}
}

func TestHTMLCompressor_Compress_CodeBlocks(t *testing.T) {
	h := NewHTMLCompressor()
	html := `<pre><code class="language-python">def hello():
    return true</code></pre>`

	result, err := h.Compress(html)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(result, "p:") {
		t.Errorf("expected python code record (p:), got: %s", result)
	}
	if !strings.Contains(result, "ret 1") {
		t.Errorf("expected 'return true' -> 'ret 1', got: %s", result)
	}
}

func TestHTMLCompressor_Compress_Lists(t *testing.T) {
	h := NewHTMLCompressor()
	html := `<ul><li>First item</li><li>Second item</li></ul>`

	result, err := h.Compress(html)
	if err != nil {
		t.Fatal(err)
	}

	parts := strings.Split(result, "|")
	listCount := 0
	for _, p := range parts {
		if strings.HasPrefix(p, "l:") {
			listCount++
		}
	}
	if listCount != 2 {
		t.Errorf("expected 2 list records, got %d in: %s", listCount, result)
	}
}

func TestHTMLCompressor_Compress_Tables(t *testing.T) {
	h := NewHTMLCompressor()
	html := `<table><tr><th>Name</th><th>Value</th></tr><tr><td>CPU</td><td>95%</td></tr></table>`

	result, err := h.Compress(html)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "d:") {
		t.Errorf("expected table data record, got: %s", result)
	}
	if !strings.Contains(result, "CPU") {
		t.Errorf("expected table cell content, got: %s", result)
	}
}

func TestHTMLCompressor_Compress_Links(t *testing.T) {
	h := NewHTMLCompressor()
	html := `<a href="https://example.com/docs">Documentation</a>`

	result, err := h.Compress(html)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "u:https://example.com/docs") {
		t.Errorf("expected URL record, got: %s", result)
	}
}

func TestHTMLCompressor_Compress_NoiseRemoval(t *testing.T) {
	h := NewHTMLCompressor()
	html := `<nav><a href="/">Home</a></nav><script>alert('x')</script><h2>Real Content</h2><style>.x{}</style><footer>Copyright</footer>`

	result, err := h.Compress(html)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "2:Real Content") {
		t.Errorf("expected heading to survive noise removal, got: %s", result)
	}
	if strings.Contains(result, "alert") {
		t.Error("script content should be removed")
	}
	if strings.Contains(result, "Copyright") {
		t.Error("footer content should be removed")
	}
}

func TestHTMLCompressor_Compress_Abbreviations(t *testing.T) {
	h := NewHTMLCompressor()
	html := `<p>The authentication implementation uses the database configuration for the production environment.</p>`

	result, err := h.Compress(html)
	if err != nil {
		t.Fatal(err)
	}

	// "the" articles should be removed, terms abbreviated
	if strings.Contains(result, " the ") {
		t.Errorf("articles should be removed, got: %s", result)
	}
	if !strings.Contains(result, "auth") {
		t.Errorf("expected 'authentication' abbreviated to 'auth', got: %s", result)
	}
	if !strings.Contains(result, "impl") {
		t.Errorf("expected 'implementation' abbreviated to 'impl', got: %s", result)
	}
	if !strings.Contains(result, "db") {
		t.Errorf("expected 'database' abbreviated to 'db', got: %s", result)
	}
}

func TestHTMLCompressor_Compress_Symbols(t *testing.T) {
	h := NewHTMLCompressor()
	html := `<p>Redis is faster than MySQL and Memcached fails to connect.</p>`

	result, err := h.Compress(html)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, ">") {
		t.Errorf("expected 'faster than' -> '>', got: %s", result)
	}
	if !strings.Contains(result, "&") {
		t.Errorf("expected ' and ' -> '&', got: %s", result)
	}
	if !strings.Contains(result, "!") {
		t.Errorf("expected 'fails to' -> '!', got: %s", result)
	}
}

func TestHTMLCompressor_Compress_HTMLEntities(t *testing.T) {
	h := NewHTMLCompressor()
	html := `<p>Use &amp; for the ampersand. Quotes: &quot;hello&quot; and &lt;tag&gt;</p>`

	result, err := h.Compress(html)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(result, "&amp;") {
		t.Errorf("HTML entities should be decoded, got: %s", result)
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
	uccp := "1:Title|2:Section|t:Some text|l:Item one|d:Name,Value"

	result, err := h.Decompress(uccp)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "# Title") {
		t.Errorf("expected H1 decompression, got: %s", result)
	}
	if !strings.Contains(result, "## Section") {
		t.Errorf("expected H2 decompression, got: %s", result)
	}
	if !strings.Contains(result, "Some text") {
		t.Errorf("expected text decompression, got: %s", result)
	}
	if !strings.Contains(result, "- Item one") {
		t.Errorf("expected list decompression, got: %s", result)
	}
}

func TestHTMLCompressor_SystemPrompt(t *testing.T) {
	h := NewHTMLCompressor()
	prompt := h.SystemPrompt()

	if !strings.Contains(prompt, "UCCP") {
		t.Error("SystemPrompt should mention UCCP")
	}
	if !strings.Contains(prompt, "HTML") {
		t.Error("SystemPrompt should mention HTML domain")
	}
}

func TestHTMLCompressor_EstimateTokens(t *testing.T) {
	h := NewHTMLCompressor()
	tokens := h.EstimateTokens("Hello world, this is a test")
	if tokens <= 0 {
		t.Error("expected positive token estimate")
	}
}

func TestHTMLCompressor_LangCodes(t *testing.T) {
	tests := []struct {
		html     string
		wantCode string
	}{
		{`<pre><code class="language-python">x=1</code></pre>`, "p:"},
		{`<pre><code class="language-go">x:=1</code></pre>`, "g:"},
		{`<pre><code class="language-javascript">let x=1</code></pre>`, "s:"},
		{`<pre><code class="language-bash">echo hi</code></pre>`, "b:"},
		{`<pre><code class="language-rust">let x=1;</code></pre>`, "r:"},
		{`<pre><code class="language-sql">SELECT 1</code></pre>`, "q:"},
	}

	h := NewHTMLCompressor()
	for _, tt := range tests {
		result, err := h.Compress(tt.html)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(result, tt.wantCode) {
			t.Errorf("Compress(%q) = %q, want prefix %q", tt.html, result, tt.wantCode)
		}
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

	// Verify compression actually reduces size
	if len(result) >= len(html) {
		t.Errorf("compressed (%d bytes) should be smaller than original (%d bytes)", len(result), len(html))
	}

	// Verify key content is preserved
	if !strings.Contains(result, "1:Docker Networking Guide") {
		t.Error("H1 should be preserved")
	}
	if !strings.Contains(result, "2:Overview") {
		t.Error("H2 should be preserved")
	}
	if !strings.Contains(result, "b:docker network create mynet") {
		t.Error("bash code block should be preserved with correct lang code")
	}
	// Nav and footer should be stripped
	if strings.Contains(result, "Home") {
		t.Error("nav content should be removed")
	}
	if strings.Contains(result, "Copyright") {
		t.Error("footer content should be removed")
	}
}
