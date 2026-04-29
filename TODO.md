# UCCP Improvement Ideas for Blog Content

From analysis of real blog post compression (April 2026). Current HTML compressor achieves ~87% (13% savings) on clean article HTML. These are ideas to explore — not all may be worth implementing.

## Priority: How We Call It (ai-post-to-wp side)

- [ ] Chain compressors: HTML→markdown first, then financial compressor on financial posts (currently only one runs)
- [ ] Strip `<figure>` blocks from LLM context — AI doesn't need image URLs when harmonizing/rewriting text
- [ ] Deduplicate URLs across blocks (many blocks link to the same sources) — use `[§1]` references with URL index
- [ ] Use `BatchDedupCompressor` when compressing multiple blocks from the same workspace

## Priority: Library Improvements

- [ ] Figure compression: `<figure class="wp-block-image"><img src="URL" alt="DESC"/><figcaption>Caption</figcaption></figure>` → `[IMG:DESC|Caption]` (big HTML savings per image)
- [ ] Table compression: convert HTML tables to JSON compressor's columnar format instead of markdown tables
- [ ] Make financial compressor work on markdown output (currently expects plaintext, could chain after HTML compressor)

## Review Needed: Blog Domain Abbreviations

**NOTE:** The user does NOT want generic blog-term abbreviations like "cybersecurity"→"cybsec", "infrastructure"→"infra", etc. These feel lossy and may confuse LLMs or reduce readability. If a blog domain compressor is created, abbreviations should be:
- Only for truly repetitive, long terms that appear 5+ times per article
- Validated that LLMs understand the abbreviation correctly
- Measured for actual token savings vs. readability cost

The existing HTML compressor's abbreviations (Performance→Prf, function→fn, etc.) are fine because they're well-established developer shorthand. Inventing new abbreviations for general prose is a different story.

## Not Worth Doing

- Article removal ("the"/"a"/"an") has minimal impact on prose — keep it but don't expand
- General whitespace optimization — clean HTML doesn't have much waste
- Nav/footer/sidebar stripping — our generated content doesn't have these
