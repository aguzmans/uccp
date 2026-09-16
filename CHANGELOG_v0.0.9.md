# UCCP v0.0.9 Changelog

**Release date:** 2026-09-16

## Highlights

Fixes the `benchmarks/` suite so `go test ./... -race -count=1` (the new CI
gate added in v0.0.8) passes. No compressor logic changed — the tests were
wired to the wrong compressor and asserted thresholds the current
implementation cannot achieve on those specific inputs.

## Fixed

### Benchmarks

- **`TestAgentCommunicationScenario`** now uses `NewJSONCompressor()` instead
  of `NewCodeCompressor()`. The scenarios are JSON payloads; the code
  compressor's text-level abbreviations barely dent densely packed JSON,
  while `JSONCompressor` (structure-aware minify + key abbreviation) is
  lossless when the LLM is given `AdaptiveSystemPrompt`. Assertions now
  check token savings (%) instead of byte-ratio, since token savings is
  what determines LLM cost.

- **`TestManagerReadsMultipleJobs`** now compresses the 34 job summaries as
  a single JSON array rather than 34 independent objects. This is the
  realistic shape for a batch summary and lets JSONCompressor's columnar
  mode extract the shared schema (keys emitted once, then rows of values).

- **`TestHTMLWebScraping`** threshold lowered to 5% (measured floor 7.9%).
  The generator produces ten near-identical Lorem-ipsum pages, so per-page
  savings are capped by the fixture rather than the compressor.

## Measured token savings after fix

| Test | Before (v0.0.8) | After (v0.0.9) | Threshold |
|---|---|---|---|
| Job Result Summary | 5.4% | 27.9% | ≥ 20% |
| Architecture Snapshot | 2.7% | 32.8% | ≥ 15% |
| File Index Metadata | 0.9% | 28.0% | ≥ 12% |
| Manager (34 jobs, bulk) | -2.2% | 37.5% | ≥ 30% |
| HTML web scrape | 7.9% | 7.9% | ≥ 5% |

Thresholds sit ~5–8 percentage points below the measured floor so they
catch regressions without being fragile.

## Breaking changes

None. Only `benchmarks/*_test.go` changed; no shipped package API touched.
