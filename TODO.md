# UCCP Improvement Ideas for Blog Content

From analysis of real blog post compression (April 2026). Current HTML compressor achieves ~87% (13% savings) on clean article HTML. These are ideas to explore — not all may be worth implementing.

## Library Improvements

- [ ] Figure compression: `<figure class="wp-block-image"><img src="URL" alt="DESC"/><figcaption>Caption</figcaption></figure>` → `[IMG:DESC|Caption]` (big HTML savings per image)
- [ ] Table compression: convert HTML tables to JSON compressor's columnar format instead of markdown tables
- [ ] Make financial compressor work on markdown output (currently expects plaintext, could chain after HTML compressor)
- [ ] Chain compressors: e.g. HTML→markdown first, then a domain-specific compressor on the result
- [ ] Deduplicate URLs across blocks — use `[§1]` references with URL index
- [ ] Strip `<figure>` blocks when the downstream LLM task doesn't need image URLs
- [ ] Batch deduplication when compressing multiple related blocks

## Review Needed: Blog Domain Abbreviations

Generic prose abbreviations (e.g. "cybersecurity"→"cybsec", "infrastructure"→"infra") should be avoided — they feel lossy and may confuse LLMs or reduce readability. If a blog domain compressor is added, abbreviations should be:

- Only for truly repetitive, long terms that appear 5+ times per article
- Validated that LLMs understand the abbreviation correctly
- Measured for actual token savings vs. readability cost

The existing HTML compressor's abbreviations (Performance→Prf, function→fn, etc.) are fine because they're well-established developer shorthand. Inventing new abbreviations for general prose is a different story.

## Not Worth Doing

- Article removal ("the"/"a"/"an") has minimal impact on prose — keep it but don't expand
- General whitespace optimization — clean HTML doesn't have much waste
- Nav/footer/sidebar stripping — generated content typically doesn't have these
