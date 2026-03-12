# UCCP Benchmarks

Comprehensive benchmarks demonstrating UCCP compression performance across different use cases and content types.

---

## Quick Start

```bash
cd /Users/abel/Documents/Code-Experiments/uccp

# Run all benchmarks
go test ./benchmarks/... -v

# Run specific benchmark category
go test ./benchmarks/ -v -run TestAgentCommunication
go test ./benchmarks/ -v -run TestHTML
go test ./benchmarks/ -v -run TestJSON
go test ./benchmarks/ -v -run TestMarkdown

# Run performance benchmarks
go test ./benchmarks/ -bench=. -benchmem
```

---

## Benchmark Categories

### 1. Agent Communication (`agent_simulation_test.go`)

Simulates realistic agent-to-agent communication scenarios.

**Tests:**
- `TestAgentCommunicationScenario` - Individual message compression
- `TestManagerReadsMultipleJobs` - Batch compression (34 jobs)
- `BenchmarkAgentCommunication` - Performance measurement

**Use Cases:**
- Job result summaries
- Project architecture snapshots
- File index metadata
- Manager reading completed jobs

**Expected Results:**
- Job results: 70-85% compression
- Architecture: 60-75% compression
- Batch processing: 70%+ overall savings

**Example Output:**
```
=== Job Result Summary ===
Original size: 523 bytes (131 tokens)
Compressed size: 142 bytes (36 tokens)
Compression ratio: 72.8%
Token savings: 95 tokens (72.5%)

=== Manager Reading 34 Job Summaries ===
Without UCCP:
  Total tokens: 4,454
  Estimated cost: $0.0134

With UCCP:
  Total tokens: 1,224
  Estimated cost: $0.0037

Savings:
  Token reduction: 3,230 tokens (72.5%)
  Cost savings: $0.0097
```

---

### 2. HTML Compression (`html_test.go`)

Tests compression on HTML content from web scraping and documentation.

**Tests:**
- `TestHTMLDocumentation` - Various HTML document types
- `TestHTMLWebScraping` - Multi-page scraping scenario
- `BenchmarkHTMLCompression` - Performance measurement

**Use Cases:**
- Documentation pages
- Blog articles
- API documentation
- Web scraping results

**Expected Results:**
- Simple articles: 60%+ compression
- Technical docs: 70%+ compression
- Blog posts: 65%+ compression

**Example Output:**
```
=== Web Scraping 10 Documentation Pages ===
Without UCCP:
  Total tokens: 125,000
  Estimated cost: $0.38
  Pages that fit in 200k context: 16

With UCCP:
  Total tokens: 37,500
  Estimated cost: $0.11
  Pages that fit in 200k context: 53

Savings:
  Token reduction: 87,500 tokens (70.0%)
  Cost savings: $0.27
  Capacity multiplier: 3.3x more pages
```

---

### 3. JSON Compression (`json_test.go`)

Tests compression on JSON API responses and structured data.

**Tests:**
- `TestJSONAPIResponses` - Various API response formats
- `TestJSONArrayCompression` - Repeated array structures
- `TestAPIResponseChain` - Multiple API call sequence
- `BenchmarkJSONCompression` - Performance measurement

**Use Cases:**
- API responses
- User lists
- Product catalogs
- Configuration objects

**Expected Results (with CodeCompressor baseline):**
- User lists: 60-70% compression
- Product catalogs: 65-75% compression
- Config objects: 60-70% compression

**Note:** Dedicated JSON domain (v0.0.8) will achieve 75-85% compression by detecting repeated structure.

**Example Output:**
```
=== User List API Response ===
Description: Paginated user list with repeated structure
Original size: 456 bytes (114 tokens)
Compressed size: 187 bytes (47 tokens)
Compression ratio: 59.0%
Token savings: 67 tokens (58.8%)

Note: Using CodeCompressor as baseline. JSON domain (v0.0.8) will achieve 70%+ compression
```

---

### 4. Markdown Compression (`markdown_test.go`)

Tests compression on Markdown documentation and agent messages.

**Tests:**
- `TestMarkdownDocumentation` - Various Markdown formats
- `TestMarkdownAgentMessages` - Agent communication
- `BenchmarkMarkdownCompression` - Performance measurement

**Use Cases:**
- README files
- API documentation
- Planning summaries
- Agent status messages

**Expected Results (with CodeCompressor baseline):**
- READMEs: 60-65% compression
- API docs: 65-70% compression
- Agent messages: 55-65% compression

**Note:** Dedicated Markdown domain (v0.0.6) will achieve 70-80% compression with format-aware rules.

**Example Output:**
```
=== README Documentation ===
Description: Standard README with code blocks and lists
Original size: 892 bytes (223 tokens)
Compressed size: 356 bytes (89 tokens)
Compression ratio: 60.1%
Token savings: 134 tokens (60.1%)

Note: Using CodeCompressor as baseline. Markdown domain (v0.0.6) will achieve 65%+ compression
```

---

## Performance Benchmarks

Run performance benchmarks to measure compression speed:

```bash
go test ./benchmarks/ -bench=. -benchmem
```

**Expected Performance:**
```
BenchmarkAgentCommunication-8         50000    25000 ns/op    5000 B/op    50 allocs/op
BenchmarkHTMLCompression-8            20000    45000 ns/op   10000 B/op   100 allocs/op
BenchmarkJSONCompression-8            40000    30000 ns/op    7000 B/op    70 allocs/op
BenchmarkMarkdownCompression-8        45000    28000 ns/op    6000 B/op    60 allocs/op
```

**Interpretation:**
- 50,000 ops/sec for agent messages
- 20,000 ops/sec for HTML
- Fast enough for real-time compression
- Memory usage: 5-10KB per operation

---

## Token Cost Analysis

### Agent System (34 Jobs/Day)

**Scenario:** Manager reads 34 job summaries daily

| Metric | Without UCCP | With UCCP | Savings |
|--------|--------------|-----------|---------|
| Tokens/day | 680,000 | 277,100 | 402,900 (59%) |
| Cost/day | $2.04 | $0.83 | $1.21 (59%) |
| Cost/month | $61.20 | $24.90 | $36.30 (59%) |
| Cost/year | $734.40 | $299.16 | $435.24 (59%) |

### Web Scraping (10 Pages/Session)

**Scenario:** Feed 10 documentation pages to LLM

| Metric | Without UCCP | With UCCP | Savings |
|--------|--------------|-----------|---------|
| Tokens/session | 125,000 | 37,500 | 87,500 (70%) |
| Cost/session | $0.38 | $0.11 | $0.27 (70%) |
| Pages in 200k context | 16 | 53 | 3.3x capacity |

### API Response Processing (100 Requests/Day)

**Scenario:** Process API responses before sending to LLM

| Metric | Without UCCP | With UCCP | Savings |
|--------|--------------|-----------|---------|
| Tokens/day | 50,000 | 20,000 | 30,000 (60%) |
| Cost/day | $0.15 | $0.06 | $0.09 (60%) |
| Cost/month | $4.50 | $1.80 | $2.70 (60%) |

---

## Adding New Benchmarks

### 1. Create Test File

```go
package benchmarks

import (
    "testing"
    "github.com/aguzmans/uccp/core"
    "github.com/aguzmans/uccp/domains"
)

func TestMyScenario(t *testing.T) {
    compressor := domains.NewCodeCompressor()

    original := "your test content here"
    compressed, err := compressor.Compress(original)
    if err != nil {
        t.Fatalf("Compression failed: %v", err)
    }

    ratio := core.CalculateCompressionRatio(original, compressed)
    tokenSavings := core.EstimateTokenSavings(original, compressed)

    t.Logf("Compression: %.1f%%", ratio*100)
    t.Logf("Token savings: %d", tokenSavings)
}
```

### 2. Add Benchmark

```go
func BenchmarkMyScenario(b *testing.B) {
    compressor := domains.NewCodeCompressor()
    content := "your test content"

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = compressor.Compress(content)
    }
}
```

### 3. Run Tests

```bash
go test ./benchmarks/ -v -run TestMyScenario
go test ./benchmarks/ -bench=BenchmarkMyScenario
```

---

## Interpreting Results

### Compression Ratio

- **70-99%**: Excellent (agent messages, repeated data)
- **60-70%**: Good (HTML, JSON with structure)
- **50-60%**: Moderate (complex content)
- **<50%**: Poor (already compact or random data)

### Token Savings

Calculate cost savings:
```
Daily savings = (Original tokens - Compressed tokens) × API rate
Monthly savings = Daily savings × 30
Annual savings = Daily savings × 365
```

**Claude Sonnet 4.5 rates:**
- Input: $3.00 per 1M tokens
- Output: $15.00 per 1M tokens

### Performance

- **>10,000 ops/sec**: Excellent (real-time compression)
- **1,000-10,000 ops/sec**: Good (batch processing)
- **<1,000 ops/sec**: Acceptable (large files)

---

## Validation Criteria

Tests validate that UCCP:

1. ✅ **Achieves target compression** (≥50% for most content)
2. ✅ **Reduces token count** (measured via tiktoken estimates)
3. ✅ **Fast performance** (≥1,000 compressions/sec)
4. ✅ **LLM-readable** (compressed format is valid UCCP)
5. ✅ **Smart decisions** (only compresses when beneficial)

---

## Future Domains

### v0.0.6: Markdown Domain
- Heading compression (# → H1:)
- List optimization
- Code fence handling
- Target: 70-80% compression

### v0.0.8: JSON Domain
- Schema detection
- Array structure recognition
- Key abbreviation
- Target: 75-85% compression

### v0.0.9: Multi-Language Code
- Python patterns
- Java conventions
- Rust idioms
- Target: 70-80% compression

---

## Contributing

Add benchmarks for:
- New use cases
- Different content types
- Edge cases
- Performance regressions

**Guidelines:**
1. Use realistic data (not "hello world")
2. Include multiple scenarios per test
3. Report both compression and token metrics
4. Add performance benchmarks
5. Document expected results

---

**Run benchmarks before releases to validate improvements!**
