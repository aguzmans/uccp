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
