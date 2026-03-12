# UCCP - Ultra-Compact Content Protocol

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**LLM-readable compression for agent-to-agent communication**

UCCP is a novel compression format designed specifically for Large Language Models. Unlike traditional compression (gzip, Brotli) which produces binary output, UCCP compresses content into a human-and-LLM-readable text format, achieving 70-99% compression while remaining intelligible to language models.

## Primary Use Cases

### A. Web Content Ingestion
When feeding multiple HTML sources to LLMs (documentation, articles, web scraping):
- **Problem:** HTML is verbose (tags, boilerplate, repeated elements)
- **Example:** 50KB documentation page → 2KB UCCP
- **Solution:** Compress HTML content before sending to LLM, achieving 60-96% token reduction
- **Benefit:** Process 10x more web pages within the same token budget

### B. Agent-to-Agent Communication
Internal messaging between AI agents for code, JSON, and markdown:
- **Problem:** Agents repeatedly share the same context (job results, code snippets, architecture)
- **Example:** Job summary 2.8KB JSON → 142 bytes UCCP
- **Solution:** Compress agent communications (job results, planning context, retry information)
- **Benefit:** 95-99% token reduction, enabling efficient multi-agent workflows

## Why UCCP?

### The Problem

When building AI agent systems, token consumption becomes a major cost:
- **Repetitive context**: Agents re-read the same files repeatedly
- **Verbose communication**: JSON/XML formats waste tokens
- **No existing solution**: Traditional compression requires decompression (still consumes tokens)

### The Solution

UCCP provides:
- ✅ **70-99% compression ratio** on code, jobs, architecture content
- ✅ **LLM-readable format** - no decompression needed (Claude/GPT read it natively)
- ✅ **Smart compression decision** - automatically determines when compression saves tokens
- ✅ **Domain-aware** - different optimizations for HTML vs code vs JSON
- ✅ **Zero decompression cost** - LLMs process compressed format directly

### Real-World Impact

**Before UCCP:**
```
Manager reads 34 completed jobs:
- 34 × 50KB JSON = 1.7MB
- ~400,000 tokens
- $1.20 per read
```

**After UCCP:**
```
Manager reads 34 compressed summaries:
- 34 × 150 bytes UCCP = 5.1KB
- ~1,300 tokens
- $0.004 per read
```

**Result: 99.7% token reduction, 300x cost reduction**

> **Note:** High compression ratios (>90%) involve lossy summarization — structural details
> are condensed into abbreviated summaries. For lossless abbreviation-only compression
> (where all information is preserved), expect 60-75% reduction. Both modes are valuable
> depending on whether downstream tasks need full fidelity or just the gist.

## Quick Start

### Installation

```bash
go get github.com/aguzmans/uccp
```

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/aguzmans/uccp/core"
    "github.com/aguzmans/uccp/domains"
)

func main() {
    // Create a code domain compressor
    compressor := domains.NewCodeCompressor()

    // Compress content
    original := "Use the function to implement authentication for the application"
    compressed, _ := compressor.Compress(original)

    fmt.Println("Original:  ", original)
    // "Use the function to implement authentication for the application"

    fmt.Println("Compressed:", compressed)
    // "Use fn→impl auth@app"

    // Calculate savings
    ratio := core.CalculateCompressionRatio(original, compressed)
    fmt.Printf("Compression: %.1f%%\n", ratio*100)
    // "Compression: 62.5%"
}
```

### Smart Compression (Automatic Decision)

UCCP automatically decides when compression saves tokens:

```go
compressor := domains.NewCodeCompressor()

// Small message - won't compress (overhead not worth it)
result := core.ShouldCompress(compressor, "Hello", core.DefaultThresholds)
fmt.Println(result.WasCompressed) // false

// Large message - will compress
jobSummary := "Successfully implemented the ActivityFeed component..."
result = core.ShouldCompress(compressor, jobSummary, core.DefaultThresholds)
fmt.Println(result.WasCompressed) // true
fmt.Println(result.Ratio)         // 0.73 (73% compression)
```

### Write and Read Messages

UCCP handles file I/O with automatic format detection:

```go
compressor := domains.NewCodeCompressor()

// Write message (auto-decides: .uccp or .txt)
path, result, _ := core.WriteMessage(
    compressor,
    jobSummary,
    "/tmp/job-021",
    core.DefaultThresholds,
)
// Creates: /tmp/job-021.uccp (if compressed) or /tmp/job-021.txt (if not)

// Read message (auto-detects format)
content, wasCompressed, systemPrompt, _ := core.ReadMessage(compressor, "/tmp/job-021")

if wasCompressed {
    // Include system prompt when sending to LLM
    fullPrompt := systemPrompt + "\n\n" + content
    // LLM now understands UCCP format
}
```

## How It Works

### Compression Techniques

UCCP applies domain-specific rules:

1. **Type prefixes**: `F:` = framework, `J:` = job, `f:` = file
2. **Symbol operators**: `→` = implements, `←` = uses, `✓` = success
3. **Abbreviations**: `fn` = function, `impl` = implementation, `comp` = component
4. **Path compression**: `src/components/` → `src/comp/`
5. **Article removal**: "the", "a", "an" removed
6. **Whitespace collapse**: Multiple spaces → single space

### Example: Project Architecture

**Before (JSON - 487 bytes):**
```json
{
  "architecture": {
    "framework": "React with TypeScript",
    "build_tool": "Vite",
    "language": "TypeScript"
  },
  "patterns": {
    "api_calls": "Use api.get() from src/lib/api.ts",
    "state_management": "Use useState and useContext hooks"
  }
}
```

**After (UCCP - 187 bytes = 61.6% compression):**
```
F:R+TS|B:Vite|L:TS|P:api→api.get()←src/l/api.ts|P:state→st&cx hooks
```

**LLM Understanding:**
With the UCCP system prompt, Claude/GPT reads this as:
- Framework: React with TypeScript
- Build tool: Vite
- Language: TypeScript
- API pattern: Use api.get() from src/lib/api.ts
- State pattern: Use useState and useContext hooks

### Example: Job Summary

**Before (JSON - 523 bytes):**
```json
{
  "job_id": "job-021",
  "status": "completed",
  "execution_time": "18m 32s",
  "files_modified": ["src/components/pages/ActivityFeed.tsx"],
  "tests_run": 5,
  "tests_passed": 5,
  "result": "Successfully implemented ActivityFeed with infinite scroll"
}
```

**After (UCCP - 142 bytes = 73% compression):**
```
J:job-021→✓|t:18m32s|M:src/comp/p/ActivityFeed.t|T:5✓0✗|R:impl ActivityFeed+∞scr
```

## Domains

UCCP supports multiple content domains:

### Code Domain (Ready)
- Code snippets, architecture, job descriptions
- Optimized for: React, TypeScript, Node.js, Go
- Compression: 70-80% typical, 99% for batches

```go
compressor := domains.NewCodeCompressor()
```

### HTML Domain (Complete)
- HTML content, web scraping results, documentation pages
- Extracts: headings, paragraphs, code blocks, lists, tables, links
- Noise removal: strips script, style, nav, header, footer
- Compression: 60-80% typical on article-style content

```go
compressor := domains.NewHTMLCompressor()
```

## Advanced Features

### Compression Thresholds

Control when compression is applied:

```go
// Default: Compress if >200 bytes AND saves >30%
core.DefaultThresholds

// Aggressive: Compress smaller content
core.AggressiveThresholds

// Conservative: Only compress when very beneficial
core.ConservativeThresholds

// Custom:
custom := core.CompressionThresholds{
    MinSize: 300,    // Only compress >300 bytes
    MinRatio: 0.40,  // Require 40% savings
}
```

### Compression Statistics

Track aggregate compression performance:

```go
stats := &core.CompressionStats{}

for _, message := range messages {
    result := core.ShouldCompress(compressor, message, core.DefaultThresholds)
    core.UpdateStats(stats, result)
}

fmt.Printf("Average compression: %.1f%%\n", stats.AverageRatio*100)
fmt.Printf("Total tokens saved: %d\n", stats.TotalTokensSaved)
fmt.Printf("Monthly cost savings: $%.2f\n",
    core.CalculateCostSavings(int(stats.TotalTokensSaved)))
```

### Domain-Specific Methods

Code compressor provides specialized methods:

```go
compressor := domains.NewCodeCompressor()

// Compress project architecture
snapshot := map[string]interface{}{
    "architecture": map[string]interface{}{
        "framework": "React with TypeScript",
        "build_tool": "Vite",
    },
}
compressed, _ := compressor.CompressProjectSnapshot(snapshot)

// Compress job results
result := map[string]interface{}{
    "job_id": "job-021",
    "status": "completed",
    "tests_passed": 5,
}
compressed, _ = compressor.CompressJobResult(result)

// Compress file index
files := map[string]interface{}{
    "src/lib/api.ts": map[string]interface{}{
        "purpose": "API client with authentication",
        "exports": []interface{}{"api object"},
    },
}
compressed, _ = compressor.CompressFileIndex(files)
```

## Comparison to Alternatives

| Solution | LLM-Readable? | Token Efficient? | Compression | Use Case |
|----------|---------------|------------------|-------------|----------|
| **UCCP** | ✅ Yes | ✅ Yes | **70-99%** | **Agent communication** |
| gzip | ❌ Binary | ❌ No | 70% | File transfer |
| Protobuf | ❌ Binary | ❌ No | 60% | API communication |
| JSON minify | ✅ Yes | ⚠️ Minimal | 10% | API responses |
| Prompt caching | N/A | ⚠️ Partial | 0% | Repeated context |

> **Benchmark note:** Compression ratios above are measured in bytes. Actual token savings
> may differ because UCCP symbols (|, →, ✓) can tokenize into multiple tokens depending
> on the model's tokenizer. We are adding tiktoken-based benchmarks to validate token-level savings.

**Why UCCP is unique:**
- LLMs read compressed format natively (no decompression tokens)
- Achieves binary-level compression in text format
- Domain-aware optimizations (HTML ≠ code ≠ JSON)

## Use Cases

### 1. AI Agent Systems
Reduce token usage in multi-agent systems where agents communicate frequently.

```go
// Worker writes compressed job result
path, _, _ := core.WriteMessage(compressor, result, "job-001", core.DefaultThresholds)

// Manager reads compressed result
content, compressed, prompt, _ := core.ReadMessage(compressor, "job-001")
if compressed {
    // Send to LLM with UCCP prompt
    llm.SendMessage(prompt + "\n\n" + content)
}
```

### 2. Web Scraping
Compress HTML content before sending to LLMs for analysis.

```go
// Coming soon - HTML domain
compressor := domains.NewHTMLCompressor()
compressed, _ := compressor.Compress(scrapedHTML)
```

### 3. Context Optimization
Share project context between agents without re-reading files.

```go
// Planning agent creates compressed context
snapshot, _ := compressor.CompressProjectSnapshot(projectData)
os.WriteFile(".context/snapshot.uccp", []byte(snapshot), 0644)

// Worker agents read compressed context
context, _ := os.ReadFile(".context/snapshot.uccp")
// ~5KB instead of ~50KB of source files
```

## Benchmarks

Token savings measured with **tiktoken cl100k_base** on realistic generated test data (HTML pages, JSON API responses, source code). Net savings account for the UCCP system prompt overhead per domain.

![UCCP Compression Benchmarks](docs/benchmark-results.svg)

**Regenerate benchmarks locally:**
```bash
go run ./benchmark/cmd/
# Generates test data in benchmark/testdata/ (gitignored)
# Outputs SVG to docs/benchmark-results.svg
```

## Roadmap

UCCP is actively developed with planned support for:

- **Markdown domain** (v0.0.6) - Documentation and agent messages
- **JSON domain** (v0.0.8) - API responses and structured data
- **Multi-language code** (v0.0.9) - Python, Java, Rust, C++
- **Cross-platform support** (v0.3.0) - Python, JavaScript, Rust libraries
- **LLM framework integrations** (v0.4.0) - LangChain, LlamaIndex plugins

See [ROADMAP.md](ROADMAP.md) for the complete development plan.

## Contributing

Contributions welcome! Priority areas:

1. **Production validation** - Test with your agent systems
2. **New domains** - Markdown, JSON, CSV, XML
3. **Multi-language support** - Python, Java, Rust code compression
4. **Performance** - Benchmarks and optimizations
5. **Documentation** - Examples, guides, tutorials

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Research & Background

UCCP is inspired by:
- **Information theory**: Huffman coding, entropy encoding
- **LLM tokenization**: Understanding how models process text
- **Domain-specific languages**: Compact specialized notation

**Novel contributions:**
- First LLM-readable compression format (to our knowledge)
- Dynamic compression decision based on token economics
- Domain-aware compression rules

See [docs/WHY-UCCP.md](docs/WHY-UCCP.md) for detailed rationale.

## Links

- **Documentation**: [docs/](docs/)
- **Examples**: [examples/](examples/)
- **Issues**: [GitHub Issues](https://github.com/aguzmans/uccp/issues)
- **Discussions**: [GitHub Discussions](https://github.com/aguzmans/uccp/discussions)

## Citation

If you use UCCP in research, please cite:

```
@software{uccp2026,
  title = {UCCP: Ultra-Compact Content Protocol},
  author = {Guzman, Abel},
  year = {2026},
  url = {https://github.com/aguzmans/uccp}
}
```

---

**Built with ❤️ for the AI agent community**
