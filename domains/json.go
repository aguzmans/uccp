package domains

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/aguzmans/uccp/core"
)

// JSONCompressor minifies JSON content for maximum token efficiency.
// For arrays of objects with repeated structure, it extracts the schema
// and emits a columnar format: keys once, then rows of values.
type JSONCompressor struct {
	usedKeyAbbrevs map[string]bool
}

func NewJSONCompressor() *JSONCompressor {
	return &JSONCompressor{
		usedKeyAbbrevs: make(map[string]bool),
	}
}

// jsonKeyAbbrevs maps long common JSON keys to short forms.
//
// v0.0.8 (2026-04-29): expanded for LLM/agent JSON traffic. The original list
// targeted CRUD-style records (employee data, audit timestamps). These cover
// the OpenAI-compatible tool-call format and common research-agent payloads
// where a single conversation can contain dozens of "tool_calls", "function",
// "arguments", "role", "content", "message" repetitions.
var jsonKeyAbbrevs = map[string]string{
	// Generic CRUD / record fields (original v0.0.7 set)
	"description":         "desc",
	"configuration":       "cfg",
	"environment":         "env",
	"application":         "app",
	"information":         "info",
	"repository":          "repo",
	"development":         "dev",
	"production":          "prod",
	"dependencies":        "deps",
	"created_at":          "cAt",
	"updated_at":          "uAt",
	"deleted_at":          "dAt",
	"first_name":          "fName",
	"last_name":           "lName",
	"email_address":       "email",
	"phone_number":        "phone",
	"is_active":           "active",
	"is_enabled":          "enabled",
	"is_deleted":          "deleted",
	"employee_id":         "eid",
	"manager_id":          "mgr",
	"hire_date":           "hDt",
	"last_review_date":    "lrDt",
	"pay_period":          "pp",
	"currency":            "cur",
	"current_rating":      "cRat",
	"goals_completed":     "gDone",
	"goals_total":         "gTot",
	"peer_feedback_score": "pfScore",
	"certifications":      "certs",

	// LLM / chat-completion / agent JSON (new in v0.0.8)
	"tool_calls":        "tc",
	"tool_call_id":      "tcid",
	"tool_choice":       "tch",
	"function":          "fn",
	"function_call":     "fnCall",
	"arguments":         "args",
	"messages":          "msgs",
	"message":           "msg",
	"content":           "ct",
	"role":              "rl",
	"system":            "sys",
	"assistant":         "ast",
	"finish_reason":     "fr",
	"choices":           "ch",
	"index":             "ix",
	"prompt_tokens":     "pt",
	"completion_tokens": "ctk",
	"total_tokens":      "tt",
	"cached_tokens":     "ctt",
	"model":             "m",
	"usage":             "us",
	"max_tokens":        "mxt",
	"max_results":       "mr",
	"max_chars":         "mc",
	"temperature":       "tp",
	"top_p":             "tpp",
	"presence_penalty":  "pp_p",
	"frequency_penalty": "fp",
	"stop":              "stp",
	"stream":            "str",
	"response_format":   "rf",
	"json_schema":       "js",

	// Research / search / web-fetch payloads (new in v0.0.8)
	"query":          "q",
	"url":            "u",
	"link":           "lk",
	"title":          "t",
	"snippet":        "sn",
	"results":        "rs",
	"search_results": "sr",
	"summary":        "sm",
	"author":         "au",
	"published":      "pub",
	"published_at":   "pAt",
	"date":           "dt",
	"score":          "sc",
	"rank":           "rk",
	"relevance":      "rv",
	"warning":        "wn",
	"reliable":       "rb",
	"status":         "st",
	"status_code":    "sct",
	"error":          "er",
	"success":        "ok",
	"data":           "d",
	"reason":         "rsn",
	"category":       "cat",
	"tags":           "tg",
	"slug":           "sl",
	"excerpt":        "ex",
	"complexity":     "cx",
}

// Compress minifies JSON. For arrays of objects it uses columnar format.
// Handles multiple JSON documents concatenated with whitespace.
func (j *JSONCompressor) Compress(content string) (string, error) {
	j.usedKeyAbbrevs = make(map[string]bool)
	content = strings.TrimSpace(content)

	// Try as a single JSON value first
	var raw interface{}
	if err := json.Unmarshal([]byte(content), &raw); err == nil {
		return j.compressValue(raw), nil
	}

	// Multiple JSON values concatenated — split and compress each
	chunks := splitJSONDocuments(content)
	if len(chunks) > 1 {
		var results []string
		for _, chunk := range chunks {
			var v interface{}
			if err := json.Unmarshal([]byte(chunk), &v); err == nil {
				results = append(results, j.compressValue(v))
			} else {
				results = append(results, j.fallbackCompress(chunk))
			}
		}
		return strings.Join(results, "\n"), nil
	}

	return j.fallbackCompress(content), nil
}

// compressValue compresses a parsed JSON value.
func (j *JSONCompressor) compressValue(raw interface{}) string {
	if arr, ok := raw.([]interface{}); ok && len(arr) >= 2 {
		if result, ok := j.compressArray(arr); ok {
			return result
		}
	}
	return j.compactAndAbbreviate(mustMarshal(raw))
}

// splitJSONDocuments splits a string containing multiple concatenated
// JSON documents (arrays or objects) into individual documents.
func splitJSONDocuments(content string) []string {
	var docs []string
	dec := json.NewDecoder(strings.NewReader(content))
	for dec.More() {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			break
		}
		docs = append(docs, string(raw))
	}
	if len(docs) == 0 {
		return []string{content}
	}
	return docs
}

// compressArray converts an array of objects with shared keys into
// a columnar format: COLS:k1,k2,k3\nROW:v1,v2,v3\nROW:v1,v2,v3
// This eliminates repeated key names across all objects.
func (j *JSONCompressor) compressArray(arr []interface{}) (string, bool) {
	// Verify all elements are objects
	var objects []map[string]interface{}
	for _, item := range arr {
		obj, ok := item.(map[string]interface{})
		if !ok {
			return "", false
		}
		objects = append(objects, obj)
	}

	// Collect all keys from all objects, in stable order
	keySet := make(map[string]bool)
	for _, obj := range objects {
		for k := range obj {
			keySet[k] = true
		}
	}
	keys := make([]string, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Abbreviate keys
	abbrKeys := make([]string, len(keys))
	for i, k := range keys {
		if short, ok := jsonKeyAbbrevs[k]; ok {
			abbrKeys[i] = short
			j.usedKeyAbbrevs[short+"="+k] = true
		} else {
			abbrKeys[i] = k
		}
	}

	// Build columnar output
	var b strings.Builder
	b.WriteString("COLS:")
	b.WriteString(strings.Join(abbrKeys, ","))
	b.WriteString("\n")

	for _, obj := range objects {
		vals := make([]string, len(keys))
		for i, k := range keys {
			v, exists := obj[k]
			if !exists {
				vals[i] = ""
				continue
			}
			vals[i] = j.flattenValue(v)
		}
		b.WriteString("ROW:")
		b.WriteString(strings.Join(vals, ","))
		b.WriteString("\n")
	}

	result := strings.TrimSpace(b.String())

	// Sanity check: only use columnar if it's actually smaller
	compact := j.compactAndAbbreviate(mustMarshal(arr))
	if len(result) >= len(compact) {
		return compact, true
	}
	return result, true
}

// flattenValue converts a JSON value to a compact string representation.
func (j *JSONCompressor) flattenValue(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return ""
	case bool:
		if val {
			return "1"
		}
		return "0"
	case float64:
		// Integers as clean ints
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%g", val)
	case string:
		// Escape commas in values since comma is our delimiter
		return strings.ReplaceAll(val, ",", "\\,")
	case []interface{}:
		// Flatten arrays: [a;b;c]
		parts := make([]string, len(val))
		for i, item := range val {
			parts[i] = j.flattenValue(item)
		}
		return "[" + strings.Join(parts, ";") + "]"
	case map[string]interface{}:
		// Flatten nested objects: {k=v;k=v}
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, len(keys))
		for i, k := range keys {
			short := k
			if s, ok := jsonKeyAbbrevs[k]; ok {
				short = s
				j.usedKeyAbbrevs[s+"="+k] = true
			}
			parts[i] = short + "=" + j.flattenValue(val[k])
		}
		return "{" + strings.Join(parts, ";") + "}"
	default:
		data, _ := json.Marshal(val)
		return string(data)
	}
}

func mustMarshal(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// compactAndAbbreviate does json.Compact + key abbreviation. v0.0.8 also
// auto-discovers high-frequency keys not in the static dictionary and assigns
// them 2-char codes, plus quote-strips simple identifier values.
func (j *JSONCompressor) compactAndAbbreviate(content string) string {
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(content)); err == nil {
		content = buf.String()
	}
	// Static dictionary first (deterministic abbrevs).
	for long, short := range jsonKeyAbbrevs {
		quoted := `"` + long + `"`
		if strings.Contains(content, quoted) {
			content = strings.ReplaceAll(content, quoted, `"`+short+`"`)
			j.usedKeyAbbrevs[short+"="+long] = true
		}
	}
	// Auto-discovered abbrevs for high-frequency keys not already abbreviated.
	content = j.autoAbbreviateFrequentKeys(content)
	// Quote-strip simple identifier values: "web_search" -> ws when the value
	// is a clean identifier, no whitespace/punctuation. Pure space saver, the
	// LLM reads identifier values the same with or without quotes.
	content = stripQuotesForIdentifierValues(content)
	return content
}

// keyOccurrenceRe matches "key": at the start of an object property.
var keyOccurrenceRe = regexp.MustCompile(`"([a-zA-Z_][a-zA-Z0-9_]{2,})":`)

// autoAbbreviateFrequentKeys scans the content for object keys that appear
// 3+ times AND are not already in jsonKeyAbbrevs. It assigns 2-character codes
// (k0, k1, ..., k9, kA, kB, ...) and rewrites them. The mapping is recorded
// in usedKeyAbbrevs so AdaptiveSystemPrompt can include it.
//
// Only fires for keys at least 3 chars long and appearing 3+ times — the
// abbreviation only saves bytes when the key is long enough and used enough
// to amortize the header cost.
func (j *JSONCompressor) autoAbbreviateFrequentKeys(content string) string {
	matches := keyOccurrenceRe.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return content
	}
	// Count occurrences per key.
	counts := make(map[string]int)
	for _, m := range matches {
		key := m[1]
		// Skip keys already abbreviated by the static dictionary or this run.
		if _, hit := jsonKeyAbbrevs[key]; hit {
			continue
		}
		counts[key]++
	}
	// Only abbreviate keys that appear 3+ times AND are at least 4 chars long
	// (shorter ones don't save enough vs the 2-char code).
	type cand struct {
		key   string
		count int
	}
	var cands []cand
	for k, c := range counts {
		if c >= 3 && len(k) >= 4 {
			cands = append(cands, cand{k, c})
		}
	}
	if len(cands) == 0 {
		return content
	}
	// Sort by frequency desc to assign shortest codes to most-frequent keys.
	sort.Slice(cands, func(i, j int) bool { return cands[i].count > cands[j].count })

	codeAlphabet := []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	for i, c := range cands {
		if i >= len(codeAlphabet) {
			break
		}
		short := "k" + string(codeAlphabet[i])
		quoted := `"` + c.key + `"`
		content = strings.ReplaceAll(content, quoted, `"`+short+`"`)
		j.usedKeyAbbrevs[short+"="+c.key] = true
	}
	return content
}

// identifierValueRe matches `"identifier"` style string values: clean identifiers
// (alpha + digits + underscore + dash + dot) that are property values, not keys.
// We require the preceding char to be `:` or `[` (value position) and the
// following char to be `,`, `}`, `]`, or end-of-string.
var identifierValueRe = regexp.MustCompile(`([:\[])"([a-zA-Z_][a-zA-Z0-9_\-\.]{0,32})"([,\}\]]|$)`)

// stripQuotesForIdentifierValues removes quotes around clean-identifier string
// values. JSON parsers reject this (it's invalid JSON), but UCCP-aware LLMs
// understand the value the same way. Saves 2 bytes per occurrence and helps
// when those values repeat (e.g., role="assistant", type="function").
//
// Carefully avoids: keys (the regex requires : or [ before), strings with
// spaces or special chars, and JSON keywords (true/false/null) by checking
// after the strip that the resulting bareword isn't a JSON literal.
func stripQuotesForIdentifierValues(content string) string {
	return identifierValueRe.ReplaceAllStringFunc(content, func(m string) string {
		sub := identifierValueRe.FindStringSubmatch(m)
		if len(sub) != 4 {
			return m
		}
		prefix, value, suffix := sub[1], sub[2], sub[3]
		// Don't strip JSON literals — parsers would interpret them as bools/null.
		switch value {
		case "true", "false", "null":
			return m
		}
		// Don't strip pure-numeric values — they'd parse as numbers and lose type.
		if isAllDigits(value) {
			return m
		}
		return prefix + value + suffix
	})
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// fallbackCompress handles non-JSON content.
func (j *JSONCompressor) fallbackCompress(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	content = strings.Join(result, "\n")
	for long, short := range jsonKeyAbbrevs {
		quoted := `"` + long + `"`
		if strings.Contains(content, quoted) {
			content = strings.ReplaceAll(content, quoted, `"`+short+`"`)
			j.usedKeyAbbrevs[short+"="+long] = true
		}
	}
	return content
}

func (j *JSONCompressor) Decompress(compressed string) (string, error) {
	// Handle columnar format
	if strings.HasPrefix(compressed, "COLS:") {
		return j.decompressColumnar(compressed)
	}
	// Reverse key abbreviations
	for long, short := range jsonKeyAbbrevs {
		compressed = strings.ReplaceAll(compressed, `"`+short+`"`, `"`+long+`"`)
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(compressed), "", "  "); err == nil {
		return buf.String(), nil
	}
	return compressed, nil
}

func (j *JSONCompressor) decompressColumnar(compressed string) (string, error) {
	// Reverse abbreviations in the header
	reverseAbbrevs := make(map[string]string)
	for long, short := range jsonKeyAbbrevs {
		reverseAbbrevs[short] = long
	}

	lines := strings.Split(compressed, "\n")
	if len(lines) < 2 {
		return compressed, nil
	}

	// Parse COLS header
	colLine := strings.TrimPrefix(lines[0], "COLS:")
	cols := strings.Split(colLine, ",")
	for i, c := range cols {
		if full, ok := reverseAbbrevs[c]; ok {
			cols[i] = full
		}
	}

	// Parse ROW lines
	var objects []map[string]string
	for _, line := range lines[1:] {
		if !strings.HasPrefix(line, "ROW:") {
			continue
		}
		valStr := strings.TrimPrefix(line, "ROW:")
		vals := strings.Split(valStr, ",")
		obj := make(map[string]string)
		for i, col := range cols {
			if i < len(vals) {
				obj[col] = vals[i]
			}
		}
		objects = append(objects, obj)
	}

	data, _ := json.MarshalIndent(objects, "", "  ")
	return string(data), nil
}

// AdaptiveSystemPrompt returns a system prompt containing only the key
// abbreviations that were used in the last Compress() call.
func (j *JSONCompressor) AdaptiveSystemPrompt() string {
	base := `UCCP JSON: Minified JSON. Arrays use columnar format:
COLS:key1,key2,key3
ROW:val1,val2,val3
ROW:val1,val2,val3
Nested objects: {k=v;k=v} Arrays: [a;b;c] Booleans: 1/0`

	if len(j.usedKeyAbbrevs) == 0 {
		return base
	}

	var parts []string
	for entry := range j.usedKeyAbbrevs {
		parts = append(parts, entry)
	}
	sort.Strings(parts)

	return base + "\nKey abbreviations: " + strings.Join(parts, " ")
}

func (j *JSONCompressor) SystemPrompt() string {
	return `UCCP JSON: Minified JSON. Arrays use columnar format:
COLS:key1,key2,key3
ROW:val1,val2,val3
ROW:val1,val2,val3
Nested objects: {k=v;k=v} Arrays: [a;b;c] Booleans: 1/0
Key abbreviations: desc=description cfg=configuration env=environment app=application info=information repo=repository dev=development prod=production deps=dependencies cAt=created_at uAt=updated_at eid=employee_id mgr=manager_id fName=first_name lName=last_name`
}

func (j *JSONCompressor) EstimateTokens(content string) int {
	return core.EstimateTokenCount(content)
}

// IsWorthCompressing implements core.CompressionAdvisor. Cheap heuristic that
// avoids running full Compress() — counts how many bytes the obvious savings
// (whitespace + dictionary key abbreviations) would yield, plus a bonus when
// the content looks like an array of similarly-shaped objects (eligible for
// columnar compression).
//
// Returns:
//   - worth = true when estimated savings > 50 bytes AND > 5% of input.
//   - estBytesSaved = approximate byte reduction.
//
// Designed to be O(n) — single linear scan plus a fixed-size dictionary loop.
// Use this before ShouldCompress() to skip cheap rejects without running the
// full compression pipeline.
func (j *JSONCompressor) IsWorthCompressing(content string) (worth bool, estBytesSaved int) {
	n := len(content)
	if n < 200 {
		return false, 0
	}

	// 1. Whitespace savings (json.Compact removes pretty-print whitespace).
	whitespace := 0
	for i := 0; i < n; i++ {
		c := content[i]
		if c == ' ' || c == '\n' || c == '\t' || c == '\r' {
			whitespace++
		}
	}

	// 2. Dictionary-key savings (each match saves len(key) - len(short) - 0
	//    because both are quoted; subtract bytes saved per match).
	dictSavings := 0
	for long, short := range jsonKeyAbbrevs {
		needle := `"` + long + `":`
		hits := strings.Count(content, needle)
		if hits == 0 {
			continue
		}
		// Each match replaces "long" with "short", net delta per match is len(long)-len(short).
		dictSavings += hits * (len(long) - len(short))
	}

	// 3. Columnar-array bonus: if content has a top-level array with multiple
	//    object elements that share a structure, columnar format wins big.
	//    Cheap detection: starts with `[{` and contains a key occurring 3+ times.
	columnarBonus := 0
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "[{") {
		matches := keyOccurrenceRe.FindAllStringSubmatch(content, -1)
		counts := make(map[string]int)
		for _, m := range matches {
			counts[m[1]]++
		}
		repeats := 0
		for _, c := range counts {
			if c >= 3 {
				repeats++
			}
		}
		// Each repeated key, after columnar, drops to one COLS line.
		// Approximate savings: 5 bytes per repeated key occurrence beyond the first.
		columnarBonus = repeats * 20
	}

	estBytesSaved = whitespace + dictSavings + columnarBonus
	worth = estBytesSaved > 50 && estBytesSaved*100 > n*5
	return worth, estBytesSaved
}
