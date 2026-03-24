package core

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// DedupCompressor wraps any Compressor and deduplicates repeated sections.
// It only deduplicates when doing so actually saves space (smart threshold).
type DedupCompressor struct {
	inner       Compressor
	minFragment int // minimum chars for a block to be dedup-eligible
}

// NewDedupCompressor wraps inner with cross-format deduplication.
// minFragment is the minimum block size (in chars) to consider for dedup.
func NewDedupCompressor(inner Compressor, minFragment int) *DedupCompressor {
	if minFragment < 10 {
		minFragment = 10
	}
	return &DedupCompressor{inner: inner, minFragment: minFragment}
}

// Compress runs the inner compressor, then deduplicates repeated blocks.
func (d *DedupCompressor) Compress(content string) (string, error) {
	compressed, err := d.inner.Compress(content)
	if err != nil {
		return "", err
	}
	return d.deduplicate(compressed), nil
}

// Decompress expands dedup references, then runs inner decompressor.
func (d *DedupCompressor) Decompress(compressed string) (string, error) {
	expanded := d.expand(compressed)
	return d.inner.Decompress(expanded)
}

// SystemPrompt appends dedup explanation to the inner prompt.
func (d *DedupCompressor) SystemPrompt() string {
	base := d.inner.SystemPrompt()
	return base + "\nDEDUP: §N references point to the [§DICT] section at the top. Replace §N with the corresponding block."
}

func (d *DedupCompressor) EstimateTokens(content string) int {
	return d.inner.EstimateTokens(content)
}

// deduplicate splits content into blocks, finds duplicates, and replaces
// subsequent occurrences with §N references when it saves space.
func (d *DedupCompressor) deduplicate(content string) string {
	blocks := splitBlocks(content)
	if len(blocks) < 2 {
		return content
	}

	// Count occurrences of each block by hash
	type blockInfo struct {
		hash  string
		text  string
		count int
	}

	seen := make(map[string]*blockInfo) // hash -> info
	var order []string                   // preserve first-seen order

	for _, block := range blocks {
		trimmed := strings.TrimSpace(block)
		if len(trimmed) < d.minFragment {
			continue
		}
		h := hashBlock(trimmed)
		if info, ok := seen[h]; ok {
			info.count++
		} else {
			seen[h] = &blockInfo{hash: h, text: trimmed, count: 1}
			order = append(order, h)
		}
	}

	// Determine which blocks to deduplicate (must save space)
	type dedupEntry struct {
		id   int
		text string
		ref  string // §N
	}

	dedupMap := make(map[string]*dedupEntry) // hash -> entry
	nextID := 1

	for _, h := range order {
		info := seen[h]
		if info.count < 2 {
			continue
		}

		ref := fmt.Sprintf("§%d", nextID)
		// Dictionary entry cost: §N=<block>\n
		dictCost := len(ref) + 1 + len(info.text) + 1
		// Reference cost per occurrence: len(§N)
		refCost := len(ref)
		// Total cost with dedup: dictCost + count * refCost
		// Total cost without dedup: count * len(block)
		costWithout := info.count * len(info.text)
		costWith := dictCost + info.count*refCost

		if costWith < costWithout {
			dedupMap[h] = &dedupEntry{id: nextID, text: info.text, ref: ref}
			nextID++
		}
	}

	if len(dedupMap) == 0 {
		return content
	}

	// Build dictionary header
	var dict strings.Builder
	dict.WriteString("[§DICT]\n")
	// Write entries in order
	for _, h := range order {
		if entry, ok := dedupMap[h]; ok {
			dict.WriteString(fmt.Sprintf("%s=%s\n", entry.ref, entry.text))
		}
	}
	dict.WriteString("[/§DICT]\n")

	// Replace duplicate blocks with references
	// First occurrence stays, subsequent get replaced
	firstSeen := make(map[string]bool)
	var resultBlocks []string

	for _, block := range blocks {
		trimmed := strings.TrimSpace(block)
		h := hashBlock(trimmed)
		entry, isDedupable := dedupMap[h]

		if !isDedupable {
			resultBlocks = append(resultBlocks, block)
			continue
		}

		if !firstSeen[h] {
			// Keep first occurrence but also replace it with ref
			// (the content is in the dictionary)
			firstSeen[h] = true
			resultBlocks = append(resultBlocks, entry.ref)
		} else {
			resultBlocks = append(resultBlocks, entry.ref)
		}
	}

	return dict.String() + strings.Join(resultBlocks, "\n\n")
}

// expand reverses deduplication: parses the §DICT and replaces §N refs.
func (d *DedupCompressor) expand(content string) string {
	if !strings.HasPrefix(content, "[§DICT]") {
		return content
	}

	dictEnd := strings.Index(content, "[/§DICT]\n")
	if dictEnd < 0 {
		return content
	}

	dictSection := content[len("[§DICT]\n"):dictEnd]
	body := content[dictEnd+len("[/§DICT]\n"):]

	// Parse dictionary entries
	refs := make(map[string]string)
	for _, line := range strings.Split(dictSection, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		eqIdx := strings.Index(line, "=")
		if eqIdx < 0 {
			continue
		}
		key := line[:eqIdx]
		val := line[eqIdx+1:]
		refs[key] = val
	}

	// Replace references in body
	for key, val := range refs {
		body = strings.ReplaceAll(body, key, val)
	}

	return body
}

// splitBlocks divides content into logical blocks separated by double newlines.
func splitBlocks(content string) []string {
	// Split on double newlines (paragraphs, sections)
	raw := strings.Split(content, "\n\n")
	var blocks []string
	for _, b := range raw {
		b = strings.TrimSpace(b)
		if b != "" {
			blocks = append(blocks, b)
		}
	}
	return blocks
}

// hashBlock returns a short hex hash for dedup comparison.
func hashBlock(block string) string {
	h := sha256.Sum256([]byte(block))
	return fmt.Sprintf("%x", h[:8])
}

// BatchDedupCompressor wraps any compressor and maintains a shared dictionary
// across multiple Compress() calls. Use for compressing related documents that
// share common phrases (e.g., financial posts about the same tickers).
type BatchDedupCompressor struct {
	inner       Compressor
	minFragment int
	dictionary  map[string]string // hash → §N reference
	fragments   map[string]string // §N → original text
	nextRef     int
}

// NewBatchDedupCompressor creates a batch deduplicator that shares a dictionary
// across multiple compress calls. minFragment is the minimum block size (in chars)
// to consider for dedup.
func NewBatchDedupCompressor(inner Compressor, minFragment int) *BatchDedupCompressor {
	if minFragment < 10 {
		minFragment = 10
	}
	return &BatchDedupCompressor{
		inner:       inner,
		minFragment: minFragment,
		dictionary:  make(map[string]string),
		fragments:   make(map[string]string),
		nextRef:     1,
	}
}

// Compress delegates to the inner compressor (single-document mode).
func (b *BatchDedupCompressor) Compress(content string) (string, error) {
	return b.inner.Compress(content)
}

// Decompress delegates to the inner compressor.
func (b *BatchDedupCompressor) Decompress(compressed string) (string, error) {
	return b.inner.Decompress(compressed)
}

// SystemPrompt appends batch dedup explanation to the inner prompt.
func (b *BatchDedupCompressor) SystemPrompt() string {
	base := b.inner.SystemPrompt()
	return base + "\nBATCH DEDUP: §N references point to the shared [§DICT] section. Replace §N with the corresponding block."
}

// EstimateTokens delegates to the inner compressor.
func (b *BatchDedupCompressor) EstimateTokens(content string) int {
	return b.inner.EstimateTokens(content)
}

// CompressBatch compresses multiple content strings, sharing a dictionary across all.
// Returns compressed strings in same order and the shared dictionary header.
func (b *BatchDedupCompressor) CompressBatch(contents []string) ([]string, string, error) {
	// Step 1: Run inner.Compress() on each content string
	compressed := make([]string, len(contents))
	for i, content := range contents {
		c, err := b.inner.Compress(content)
		if err != nil {
			return nil, "", fmt.Errorf("compressing item %d: %w", i, err)
		}
		compressed[i] = c
	}

	// Step 2: Split all results into blocks and build frequency map across ALL results
	type blockOccurrence struct {
		docIndex   int
		blockIndex int
	}

	allDocBlocks := make([][]string, len(compressed))
	freq := make(map[string]int)           // hash → count across all docs
	blockText := make(map[string]string)    // hash → original text

	for i, c := range compressed {
		blocks := splitBlocks(c)
		allDocBlocks[i] = blocks
		for _, block := range blocks {
			trimmed := strings.TrimSpace(block)
			if len(trimmed) < b.minFragment {
				continue
			}
			h := hashBlock(trimmed)
			freq[h]++
			if _, ok := blockText[h]; !ok {
				blockText[h] = trimmed
			}
		}
	}

	// Step 3: For blocks appearing 2+ times, assign §N references
	// Reuse existing dictionary entries or create new ones
	dedupRefs := make(map[string]string) // hash → §N

	for h, count := range freq {
		if count < 2 {
			continue
		}
		// Check if already in the shared dictionary
		if ref, ok := b.dictionary[h]; ok {
			dedupRefs[h] = ref
			continue
		}
		// Cost check: only dedup if it saves space
		text := blockText[h]
		ref := fmt.Sprintf("§%d", b.nextRef)
		dictCost := len(ref) + 1 + len(text) + 1 // §N=<block>\n
		refCost := len(ref)
		costWithout := count * len(text)
		costWith := dictCost + count*refCost

		if costWith < costWithout {
			dedupRefs[h] = ref
			b.dictionary[h] = ref
			b.fragments[ref] = text
			b.nextRef++
		}
	}

	// Step 4: Replace occurrences with §N references
	results := make([]string, len(compressed))
	for i, blocks := range allDocBlocks {
		var resultBlocks []string
		for _, block := range blocks {
			trimmed := strings.TrimSpace(block)
			h := hashBlock(trimmed)
			if ref, ok := dedupRefs[h]; ok {
				resultBlocks = append(resultBlocks, ref)
			} else {
				resultBlocks = append(resultBlocks, block)
			}
		}
		results[i] = strings.Join(resultBlocks, "\n\n")
	}

	// Step 5: Build the shared dictionary header
	var dict strings.Builder
	if len(b.fragments) > 0 {
		dict.WriteString("[§DICT]\n")
		// Write entries in reference order (§1, §2, ...)
		for i := 1; i < b.nextRef; i++ {
			ref := fmt.Sprintf("§%d", i)
			if text, ok := b.fragments[ref]; ok {
				dict.WriteString(fmt.Sprintf("%s=%s\n", ref, text))
			}
		}
		dict.WriteString("[/§DICT]\n")
	}

	return results, dict.String(), nil
}

// Reset clears the shared dictionary for a new batch.
func (b *BatchDedupCompressor) Reset() {
	b.dictionary = make(map[string]string)
	b.fragments = make(map[string]string)
	b.nextRef = 1
}
