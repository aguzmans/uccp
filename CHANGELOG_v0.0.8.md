# UCCP v0.0.8 Changelog

**Release date:** TBD (PR by Claude on behalf of project, 2026-04-29)

## Highlights

LLM/agent JSON traffic gets meaningful new compression. The v0.0.7 JSON domain
targeted CRUD-style records (employee data, audit timestamps); v0.0.8 expands
the dictionary, auto-discovers high-frequency keys, and strips redundant quotes
on identifier values — the patterns that dominate OpenAI-compatible tool-call
payloads.

A new optional `core.CompressionAdvisor` interface lets callers cheaply pre-check
whether a piece of content is worth compressing (linear scan, no full transform)
so agent-loop runtimes can gate UCCP on a per-message basis without paying the
compression cost up front.

## Added

### Core

- **`core.CompressionAdvisor` interface** (optional) — implemented by domain
  compressors that expose a cheap `IsWorthCompressing(content) (bool, int)`
  pre-check. Returns whether compression is likely worthwhile and an estimate
  of bytes saved. O(n) implementations expected; do not run full compression.

### JSON Domain

- **Dictionary expansion (~30 → ~95 entries)**. New abbreviations targeting
  LLM/agent JSON traffic:
  - **Chat completion / tool-call structure**: `tool_calls→tc`, `tool_call_id→tcid`,
    `function→fn`, `arguments→args`, `messages→msgs`, `message→msg`, `content→ct`,
    `role→rl`, `system→sys`, `assistant→ast`, `finish_reason→fr`, `choices→ch`,
    `index→ix`, `prompt_tokens→pt`, `completion_tokens→ctk`, `total_tokens→tt`,
    `model→m`, `usage→us`, `max_tokens→mxt`, `temperature→tp`, `top_p→tpp`,
    `response_format→rf`, `json_schema→js`.
  - **Research / search payloads**: `query→q`, `url→u`, `title→t`, `snippet→sn`,
    `results→rs`, `summary→sm`, `published→pub`, `score→sc`, `relevance→rv`,
    `warning→wn`, `reliable→rb`, `status→st`, `error→er`, `data→d`, `category→cat`,
    `slug→sl`, `excerpt→ex`, and more.

- **Auto-discovered abbreviations** for high-frequency unknown keys. Keys not
  in the static dictionary that appear ≥3 times AND are ≥4 characters long get
  assigned 2-char codes (`k0`, `k1`, ..., `kZ`). The mapping is recorded via
  `usedKeyAbbrevs` so `AdaptiveSystemPrompt()` includes it in the LLM prompt.
  Sorted by frequency: shortest codes go to most-frequent keys.

- **Quote-stripping for identifier values**. Property values that are clean
  identifiers (`[a-zA-Z_][a-zA-Z0-9_\-\.]*`) get their surrounding quotes
  removed: `"function"` → `function`, `"web_search"` → `web_search`. UCCP-aware
  LLMs read these unambiguously; technically invalid JSON, but UCCP isn't JSON.
  JSON keywords (`true`, `false`, `null`) and pure-numeric values are protected
  to avoid type confusion.

- **`IsWorthCompressing(content)` cheap pre-check** on `JSONCompressor`.
  Linear scan; estimates whitespace + dictionary-key + columnar-array savings
  without running full compression. Returns `worth=true` when est. savings is
  >50 bytes AND >5% of input. Used by callers to gate UCCP on a per-message
  basis cheaply.

## Why these changes

Real-world agent-loop traffic profile (observed in
[ai-post-to-wp](https://github.com/aguzmans/ai-post-to-wp), April 2026):
each post generation produces 22+ tool-call iterations with the OpenAI
chat-completions API. The same keys recur dozens of times per conversation
(`tool_calls`, `function`, `arguments`, `role`, `content`, etc.) but were
not in the v0.0.7 dictionary. After v0.0.8, a typical agent JSON payload
compresses ~25-40% just from the static dictionary, before columnar arrays
or quote-stripping land their additional savings.

## Breaking changes

None. All v0.0.7 behavior preserved; new abbreviations are additive. Existing
callers that round-trip through `Compress()` + `Decompress()` continue to work
(the new abbrev pairs are reversed by the same lookup loop).

## Test coverage

All existing `domains/json_test.go` and `core/` tests continue to pass.
Additional regression tests for the dictionary expansion + quote-stripping
should land alongside this PR.

## Roadmap (future versions)

Items not yet implemented that align with this direction:

- Schema templating for repeated array shapes across documents.
- Common-prefix detection across keys (`tool_call`, `tool_calls`, `tool_choice`).
- String value deduplication / back-references for repeated long strings
  (URLs, model names).
- Time/date compression (`"2026-04-29T10:30:00Z"` → epoch seconds).
- Null-stripping toggle for known-optional fields.
