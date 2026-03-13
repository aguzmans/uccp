package domains

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/aguzmans/uccp/core"
)

// JSONCompressor minifies JSON content for maximum token efficiency.
// For arrays of objects with repeated structure, it extracts the schema
// and emits a columnar format: keys once, then rows of values.
type JSONCompressor struct{}

func NewJSONCompressor() *JSONCompressor {
	return &JSONCompressor{}
}

// jsonKeyAbbrevs maps long common JSON keys to short forms.
var jsonKeyAbbrevs = map[string]string{
	"description":      "desc",
	"configuration":    "cfg",
	"environment":      "env",
	"application":      "app",
	"information":      "info",
	"repository":       "repo",
	"development":      "dev",
	"production":       "prod",
	"dependencies":     "deps",
	"created_at":       "cAt",
	"updated_at":       "uAt",
	"deleted_at":       "dAt",
	"first_name":       "fName",
	"last_name":        "lName",
	"email_address":    "email",
	"phone_number":     "phone",
	"is_active":        "active",
	"is_enabled":       "enabled",
	"is_deleted":       "deleted",
	"employee_id":      "eid",
	"manager_id":       "mgr",
	"hire_date":        "hDt",
	"last_review_date": "lrDt",
	"pay_period":       "pp",
	"currency":         "cur",
	"current_rating":   "cRat",
	"goals_completed":  "gDone",
	"goals_total":      "gTot",
	"peer_feedback_score": "pfScore",
	"certifications":   "certs",
}

// Compress minifies JSON. For arrays of objects it uses columnar format.
func (j *JSONCompressor) Compress(content string) (string, error) {
	content = strings.TrimSpace(content)

	// Try to parse as JSON
	var raw interface{}
	if err := json.Unmarshal([]byte(content), &raw); err != nil {
		// Not valid JSON — just strip whitespace and abbreviate
		return j.fallbackCompress(content), nil
	}

	// Check if top-level is an array of objects → columnar compression
	if arr, ok := raw.([]interface{}); ok && len(arr) >= 2 {
		if result, ok := j.compressArray(arr); ok {
			return result, nil
		}
	}

	// Single object or non-array: just compact + abbreviate keys
	return j.compactAndAbbreviate(content), nil
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

// compactAndAbbreviate does json.Compact + key abbreviation.
func (j *JSONCompressor) compactAndAbbreviate(content string) string {
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(content)); err == nil {
		content = buf.String()
	}
	for long, short := range jsonKeyAbbrevs {
		content = strings.ReplaceAll(content, `"`+long+`"`, `"`+short+`"`)
	}
	return content
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
		content = strings.ReplaceAll(content, `"`+long+`"`, `"`+short+`"`)
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
