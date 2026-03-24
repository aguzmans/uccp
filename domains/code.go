package domains

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/aguzmans/uccp/core"
)

// CodeCompressor compresses code, architecture, jobs, and technical content
// Optimized for agent-to-agent communication in software development contexts
type CodeCompressor struct {
	usedAbbreviations map[string]bool
	usedSymbols       map[string]bool
}

// NewCodeCompressor creates a new code domain compressor
func NewCodeCompressor() *CodeCompressor {
	return &CodeCompressor{
		usedAbbreviations: make(map[string]bool),
		usedSymbols:       make(map[string]bool),
	}
}

// Compress converts code/technical content to UCCP format.
// Pipeline: strip comments → collapse indentation → abbreviate terms.
func (c *CodeCompressor) Compress(content string) (string, error) {
	c.usedAbbreviations = make(map[string]bool)
	c.usedSymbols = make(map[string]bool)
	result := stripComments(content)
	result = collapseIndentation(result)
	result = c.compressTracked(result)
	return result, nil
}

// compressTracked applies abbreviations and symbols while tracking which ones
// were actually used.
func (c *CodeCompressor) compressTracked(text string) string {
	for old, new := range abbrevMap {
		if strings.Contains(text, old) {
			text = strings.ReplaceAll(text, old, new)
			c.usedAbbreviations[old+"="+new] = true
		}
	}
	for old, new := range symbolMap {
		if strings.Contains(text, old) {
			text = strings.ReplaceAll(text, old, new)
			c.usedSymbols[strings.TrimSpace(old)+"="+new] = true
		}
	}
	text = codeArticleRe.ReplaceAllString(text, "")
	text = codeWhitespaceRe.ReplaceAllString(text, " ")
	text = strings.ReplaceAll(text, "  ", " ")
	text = strings.ReplaceAll(text, " .", ".")
	text = strings.ReplaceAll(text, " ,", ",")
	return strings.TrimSpace(text)
}

// stripComments removes single-line (//, #) and block (/* */) comments.
func stripComments(code string) string {
	// Block comments first (may span lines)
	code = blockCommentRe.ReplaceAllString(code, "")
	// Full-line // comments
	code = lineCommentRe.ReplaceAllString(code, "")
	// Full-line # comments (Python, bash, etc.)
	code = hashCommentRe.ReplaceAllString(code, "")
	// Trailing inline comments (after code)
	code = inlineCommentRe.ReplaceAllString(code, "")
	return code
}

// collapseIndentation reduces indentation to single-space-per-level
// and removes excessive blank lines.
func collapseIndentation(code string) string {
	lines := strings.Split(code, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t\r")
		if trimmed == "" {
			// Keep at most one blank line (handled by blankLinesRe below)
			result = append(result, "")
			continue
		}
		// Count leading tabs/spaces and normalize to single space per level
		leading := 0
		for _, ch := range line {
			if ch == '\t' {
				leading++
			} else if ch == ' ' {
				leading++
			} else {
				break
			}
		}
		// Use 1 space per 2-4 original spaces (rough indent level)
		level := leading / 2
		if level > 0 {
			trimmed = strings.Repeat(" ", level) + strings.TrimSpace(trimmed)
		}
		result = append(result, trimmed)
	}
	code = strings.Join(result, "\n")
	// Collapse 3+ blank lines to 1
	code = blankLinesRe.ReplaceAllString(code, "\n\n")
	return strings.TrimSpace(code)
}

// Decompress converts UCCP format back to readable content
// Note: This is lossy - articles and some formatting are not recovered
func (c *CodeCompressor) Decompress(compressed string) (string, error) {
	result := decompress(compressed)
	return result, nil
}

// SystemPrompt returns the LLM prompt explaining UCCP code compression format
func (c *CodeCompressor) SystemPrompt() string {
	return `
UCCP (Ultra-Compact Content Protocol) - Code Domain

You are reading content in UCCP format - an ultra-compact pipe-delimited format designed to save tokens.

FORMAT:
type:data|type:data  (| separates records)

TYPE CODES:
Architecture: F=framework B=build L=language P=pattern D=directory E=entrypoint
Files: f=file i=import e=export u=usage t=type c=class m=method v=variable
Jobs: J=job S=step R=result E=error W=warning T=tests M=modified C=created
Status: ✓=success/completed ✗=failure/rejected ⏳=in-progress ?=unknown

ABBREVIATIONS:
Code: fn=function cls=class int=interface impl=implementation pkg=package mod=module
      var=variable const=constant ret=return param=parameter arg=argument prop=property
      cfg=config env=environment auth=authentication db=database api=API
Files: src=source dir=directory comp=component lib=library util=utility h=hooks p=pages
Jobs: exec=execution eval=evaluation dep=dependency req=requirement res=result err=error

SYMBOLS:
→ implements/creates/flows to
← uses/depends on/imports from
& and
| or (also record separator in context)
! not/error
~ approximately
@ at/in/located
# count/number
+ with/addition
- without/removal
✓ success/completed
✗ failure/rejected
∞ infinite

PATTERNS:
React: useState→st useEffect→ef useContext→cx useRef→rf useMemo→mm useCallback→cb
API: GET→G POST→P PUT→U DELETE→D PATCH→H
HTTP: 200→ok 201→cr 400→br 401→unauth 403→forbid 404→nf 500→se
Test: describe→d it→i expect→e toBe→= toEqual→== toContain→∋

COMPRESSION RULES:
- Articles (the, a, an) removed
- Whitespace collapsed
- Common terms abbreviated
- Symbols used instead of words
- Critical details preserved (file paths, IDs, test counts)

EXAMPLES:

1. Project Architecture:
   UCCP: F:R+TS|B:Vite|L:TS|P:api→api.get()←src/l/api.ts|P:state→st&cx
   Means: Framework=React with TypeScript, Build=Vite, Language=TypeScript,
          API pattern=use api.get() from src/lib/api.ts, State=useState and useContext

2. Job Summary:
   UCCP: J:job-021→✓|t:18m|M:src/comp/p/ActivityFeed.tsx|T:5✓0✗|R:impl comp+tests✓
   Means: Job job-021 completed successfully, took 18 minutes,
          modified src/components/pages/ActivityFeed.tsx,
          5 tests passed 0 failed, implemented component with passing tests

3. File Index:
   UCCP: f:src/l/api.ts→API client+auth→e:api obj+methods→u:←{api}←'@/l/api'
   Means: File src/lib/api.ts is an API client with authentication,
          exports api object with methods, usage: import { api } from '@/lib/api'

When reading UCCP format, decode the abbreviations and symbols mentally.
The content is intentionally compact to save tokens while remaining readable.
`
}

// AdaptiveSystemPrompt returns a system prompt containing only the
// abbreviations and symbols that were used in the last Compress() call.
func (c *CodeCompressor) AdaptiveSystemPrompt() string {
	// Collect unique abbreviation pairs (deduplicate case variants)
	seenAbbrevs := make(map[string]string) // short -> long (lowercase preferred)
	for entry := range c.usedAbbreviations {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}
		long, short := parts[0], parts[1]
		if _, ok := seenAbbrevs[short]; !ok || long[0] >= 'a' {
			seenAbbrevs[short] = long
		}
	}

	var abbrevParts []string
	for short, long := range seenAbbrevs {
		abbrevParts = append(abbrevParts, short+"="+long)
	}
	sort.Strings(abbrevParts)

	var symbolParts []string
	for entry := range c.usedSymbols {
		symbolParts = append(symbolParts, entry)
	}
	sort.Strings(symbolParts)

	var b strings.Builder
	b.WriteString("UCCP (Ultra-Compact Content Protocol) - Code Domain\n\n")
	b.WriteString("You are reading content in UCCP format - an ultra-compact pipe-delimited format designed to save tokens.\n\n")
	b.WriteString("COMPRESSION RULES:\n")
	b.WriteString("- Articles (the, a, an) removed\n")
	b.WriteString("- Whitespace collapsed\n")
	b.WriteString("- Common terms abbreviated\n")
	b.WriteString("- Symbols used instead of words\n")
	b.WriteString("- Critical details preserved (file paths, IDs, test counts)\n")

	if len(abbrevParts) > 0 {
		b.WriteString("\nABBREVIATIONS USED:\n")
		for _, a := range abbrevParts {
			b.WriteString(a + " ")
		}
		b.WriteString("\n")
	}

	if len(symbolParts) > 0 {
		b.WriteString("\nSYMBOLS USED:\n")
		for _, s := range symbolParts {
			b.WriteString(s + " ")
		}
		b.WriteString("\n")
	}

	return b.String()
}

// EstimateTokens estimates token count for code content
func (c *CodeCompressor) EstimateTokens(content string) int {
	return core.EstimateTokenCount(content)
}

// CompressProjectSnapshot compresses architecture snapshot to UCCP format
func (c *CodeCompressor) CompressProjectSnapshot(data map[string]interface{}) (string, error) {
	var records []string

	// Framework info
	if arch, ok := data["architecture"].(map[string]interface{}); ok {
		if fw, ok := arch["framework"].(string); ok {
			records = append(records, "F:"+compress(fw))
		}
		if build, ok := arch["build_tool"].(string); ok {
			records = append(records, "B:"+compress(build))
		}
		if lang, ok := arch["language"].(string); ok {
			records = append(records, "L:"+compress(lang))
		}
	}

	// Patterns
	if patterns, ok := data["patterns"].(map[string]interface{}); ok {
		for key, val := range patterns {
			if strVal, ok := val.(string); ok {
				records = append(records, "P:"+compress(key)+"→"+compress(strVal))
			}
		}
	}

	// Key directories
	if dirs, ok := data["key_directories"].(map[string]interface{}); ok {
		for key, val := range dirs {
			if strVal, ok := val.(string); ok {
				records = append(records, "D:"+compress(key)+"@"+strVal)
			}
		}
	}

	return strings.Join(records, "|"), nil
}

// CompressFileIndex compresses file metadata to UCCP format
func (c *CodeCompressor) CompressFileIndex(files map[string]interface{}) (string, error) {
	var records []string

	for filepath, metadata := range files {
		if meta, ok := metadata.(map[string]interface{}); ok {
			// File record: f:path→purpose
			purpose := ""
			if p, ok := meta["purpose"].(string); ok {
				purpose = compress(p)
			}

			// Exports: e:export1,export2
			exports := ""
			if exp, ok := meta["exports"].([]interface{}); ok {
				var items []string
				for _, e := range exp {
					if str, ok := e.(string); ok {
						items = append(items, compress(str))
					}
				}
				exports = strings.Join(items, ",")
			}

			// Usage pattern: u:import{X}from'Y';X.method()
			usage := ""
			if u, ok := meta["usage_pattern"].(string); ok {
				usage = compressCode(u)
			}

			// Combine into compact record
			record := "f:" + compressPath(filepath) + "→" + purpose
			if exports != "" {
				record += "→e:" + exports
			}
			if usage != "" {
				record += "→u:" + usage
			}

			records = append(records, record)
		}
	}

	return strings.Join(records, "|"), nil
}

// CompressJobDescription compresses job metadata to UCCP format
func (c *CodeCompressor) CompressJobDescription(job map[string]interface{}) (string, error) {
	var records []string

	// Job header: J:id→title
	id := ""
	if jobID, ok := job["id"].(string); ok {
		id = jobID
	}
	title := ""
	if t, ok := job["title"].(string); ok {
		title = compress(t)
	}
	records = append(records, "J:"+id+"→"+title)

	// Steps: S:step1|S:step2
	if desc, ok := job["description"].(string); ok {
		steps := parseSteps(desc)
		for _, step := range steps {
			records = append(records, "S:"+compress(step))
		}
	}

	// Dependencies: D:dep1←dep2←dep3
	if deps, ok := job["dependencies"].([]interface{}); ok {
		var depIDs []string
		for _, dep := range deps {
			if str, ok := dep.(string); ok {
				depIDs = append(depIDs, str)
			}
		}
		if len(depIDs) > 0 {
			records = append(records, "D:"+strings.Join(depIDs, "←"))
		}
	}

	// Files needed: F:file1,file2,file3
	if files, ok := job["files_needed"].([]interface{}); ok {
		var paths []string
		for _, f := range files {
			if str, ok := f.(string); ok {
				paths = append(paths, compressPath(str))
			}
		}
		if len(paths) > 0 {
			records = append(records, "F:"+strings.Join(paths, ","))
		}
	}

	return strings.Join(records, "|"), nil
}

// CompressJobResult compresses job execution result to UCCP format
func (c *CodeCompressor) CompressJobResult(result map[string]interface{}) (string, error) {
	var parts []string

	// Job ID and status
	if id, ok := result["job_id"].(string); ok {
		status := "?"
		if s, ok := result["status"].(string); ok {
			if s == "completed" {
				status = "✓"
			} else if s == "failed" {
				status = "✗"
			} else if s == "in_progress" {
				status = "⏳"
			}
		}
		parts = append(parts, fmt.Sprintf("J:%s→%s", id, status))
	}

	// Worker ID
	if worker, ok := result["worker_id"].(string); ok {
		parts = append(parts, "w:"+worker)
	}

	// Execution time
	if execTime, ok := result["execution_time"].(string); ok {
		// Remove spaces: "18m 32s" → "18m32s"
		execTime = strings.ReplaceAll(execTime, " ", "")
		parts = append(parts, "t:"+execTime)
	}

	// Files modified
	if files, ok := result["files_modified"].([]interface{}); ok {
		for _, f := range files {
			if str, ok := f.(string); ok {
				parts = append(parts, "M:"+compressPath(str))
			}
		}
	}

	// Files created
	if files, ok := result["files_created"].([]interface{}); ok {
		for _, f := range files {
			if str, ok := f.(string); ok {
				parts = append(parts, "C:"+compressPath(str))
			}
		}
	}

	// Tests
	testsRun := 0
	testsPassed := 0
	testsFailed := 0
	if tr, ok := result["tests_run"].(int); ok {
		testsRun = tr
	}
	if tp, ok := result["tests_passed"].(int); ok {
		testsPassed = tp
	}
	if tf, ok := result["tests_failed"].(int); ok {
		testsFailed = tf
	}
	if testsRun > 0 {
		parts = append(parts, fmt.Sprintf("T:%d✓%d✗", testsPassed, testsFailed))
	}

	// Result summary
	if r, ok := result["result"].(string); ok {
		parts = append(parts, "R:"+compress(r))
	}

	return strings.Join(parts, "|"), nil
}

// DecompressToJSON converts UCCP back to readable JSON
func (c *CodeCompressor) DecompressToJSON(uccp string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	records := strings.Split(uccp, "|")

	for _, record := range records {
		parts := strings.SplitN(record, ":", 2)
		if len(parts) != 2 {
			continue
		}

		typeCode := parts[0]
		data := parts[1]

		switch typeCode {
		case "F":
			result["framework"] = decompress(data)
		case "B":
			result["build_tool"] = decompress(data)
		case "L":
			result["language"] = decompress(data)
		case "P":
			if result["patterns"] == nil {
				result["patterns"] = make(map[string]string)
			}
			kv := strings.Split(data, "→")
			if len(kv) == 2 {
				result["patterns"].(map[string]string)[decompress(kv[0])] = decompress(kv[1])
			}
		}
	}

	return result, nil
}

// Abbreviation map for code context
var abbrevMap = map[string]string{
	// Code terms
	"function":       "fn",
	"Function":       "fn",
	"class":          "cls",
	"Class":          "cls",
	"interface":      "int",
	"Interface":      "int",
	"implementation": "impl",
	"Implementation": "impl",
	"package":        "pkg",
	"Package":        "pkg",
	"module":         "mod",
	"Module":         "mod",
	"variable":       "var",
	"Variable":       "var",
	"constant":       "const",
	"Constant":       "const",
	"parameter":      "param",
	"Parameter":      "param",
	"argument":       "arg",
	"Argument":       "arg",
	"property":       "prop",
	"Property":       "prop",
	"method":         "meth",
	"Method":         "meth",
	"constructor":    "ctor",
	"Constructor":    "ctor",

	// Architecture
	"framework":      "fw",
	"Framework":      "fw",
	"architecture":   "arch",
	"Architecture":   "arch",
	"configuration":  "cfg",
	"Configuration":  "cfg",
	"environment":    "env",
	"Environment":    "env",
	"development":    "dev",
	"Development":    "dev",
	"production":     "prod",
	"Production":     "prod",
	"testing":        "test",
	"Testing":        "test",

	// Files
	"directory":      "dir",
	"Directory":      "dir",
	"component":      "comp",
	"Component":      "comp",
	"components":     "comp",
	"Components":     "comp",
	"library":        "lib",
	"Library":        "lib",
	"utility":        "util",
	"Utility":        "util",
	"utilities":      "utils",
	"Utilities":      "utils",
	"documentation":  "doc",
	"Documentation":  "doc",
	"specification":  "spec",
	"Specification":  "spec",

	// Common services
	"authentication": "auth",
	"Authentication": "auth",
	"authorization":  "authz",
	"Authorization":  "authz",
	"database":       "db",
	"Database":       "db",
	"application":    "app",
	"Application":    "app",
	"repository":     "repo",
	"Repository":     "repo",

	// Actions
	"implement": "impl",
	"Implement": "impl",
	"create":    "cr",
	"Create":    "cr",
	"update":    "upd",
	"Update":    "upd",
	"delete":    "del",
	"Delete":    "del",
	"execute":   "exec",
	"Execute":   "exec",
	"evaluate":  "eval",
	"Evaluate":  "eval",

	// React specific
	"useState":    "st",
	"useEffect":   "ef",
	"useContext":  "cx",
	"useRef":      "rf",
	"useMemo":     "mm",
	"useCallback": "cb",
	"useReducer":  "rd",

	// Common words
	"TypeScript":      "TS",
	"JavaScript":      "JS",
	"Python":          "Py",
	"React":           "R",
	"infinite scroll": "∞scr",
	"pagination":      "pag",
}

// Pre-compiled regexes for compress and compressCode
var (
	codeArticleRe      = regexp.MustCompile(`\b(the|a|an)\s`)
	codeWhitespaceRe   = regexp.MustCompile(`\s+`)
	codeStepsRe        = regexp.MustCompile(`(?m)^\s*\d+\.\s*(.+)$`)
	lineCommentRe      = regexp.MustCompile(`(?m)^(\s*)//[^\n]*$`)
	inlineCommentRe    = regexp.MustCompile(`\s*//[^\n]*`)
	blockCommentRe     = regexp.MustCompile(`(?s)/\*.*?\*/`)
	hashCommentRe      = regexp.MustCompile(`(?m)^(\s*)#[^\n]*$`)
	blankLinesRe       = regexp.MustCompile(`\n{3,}`)
	leadingWhitespaceRe = regexp.MustCompile(`(?m)^([ \t]+)`)
)

// Symbol replacements
var symbolMap = map[string]string{
	" implements ":    "→",
	" uses ":          "←",
	" depends on ":    "←",
	" creates ":       "→",
	" and ":           "&",
	" or ":            "|",
	" not ":           "!",
	" approximately ": "~",
	" at ":            "@",
	" in ":            "@",
	" with ":          "+",
	" without ":       "-",
	" optional ":      "?",
	" required ":      "*",
	" from ":          "←",
	" to ":            "→",
}

func compress(text string) string {
	// Apply abbreviations
	for old, new := range abbrevMap {
		text = strings.ReplaceAll(text, old, new)
	}

	// Apply symbols
	for old, new := range symbolMap {
		text = strings.ReplaceAll(text, old, new)
	}

	// Remove articles
	text = codeArticleRe.ReplaceAllString(text, "")

	// Collapse whitespace
	text = codeWhitespaceRe.ReplaceAllString(text, " ")

	// Remove extra punctuation
	text = strings.ReplaceAll(text, "  ", " ")
	text = strings.ReplaceAll(text, " .", ".")
	text = strings.ReplaceAll(text, " ,", ",")

	return strings.TrimSpace(text)
}

func compressCode(code string) string {
	// Compress common code patterns
	replacements := map[string]string{
		"import ":     "←",
		"from ":       "←",
		"export ":     "→",
		"function ":   "fn ",
		"const ":      "c ",
		"let ":        "l ",
		"var ":        "v ",
		"return ":     "ret ",
		"async ":      "asy ",
		"await ":      "awt ",
		"interface ":  "int ",
		"class ":      "cls ",
		"extends ":    ":",
		"implements ": "→",
	}

	for old, new := range replacements {
		code = strings.ReplaceAll(code, old, new)
	}

	// Remove extra whitespace
	code = codeWhitespaceRe.ReplaceAllString(code, " ")

	return strings.TrimSpace(code)
}

func compressPath(path string) string {
	// Compress common path patterns
	replacements := map[string]string{
		"src/components/": "src/comp/",
		"src/lib/":        "src/l/",
		"src/utils/":      "src/u/",
		"src/pages/":      "src/p/",
		"src/hooks/":      "src/h/",
		"mobile-frontend/": "mf/",
		"backend/":        "be/",
		"frontend/":       "fe/",
		"/index.tsx":      "/i.tsx",
		"/index.ts":       "/i.ts",
		"/index.js":       "/i.js",
		".tsx":            ".t",
		".jsx":            ".j",
		".test.":          ".T.",
		".spec.":          ".S.",
	}

	for old, new := range replacements {
		path = strings.ReplaceAll(path, old, new)
	}

	return path
}

func parseSteps(description string) []string {
	// Extract numbered steps from description
	re := codeStepsRe
	matches := re.FindAllStringSubmatch(description, -1)

	var steps []string
	for _, match := range matches {
		if len(match) > 1 {
			steps = append(steps, match[1])
		}
	}

	// If no numbered steps, split by newlines
	if len(steps) == 0 {
		lines := strings.Split(description, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				steps = append(steps, line)
			}
		}
	}

	return steps
}

func decompress(text string) string {
	// Reverse abbreviations
	reverseMap := make(map[string]string)
	for old, new := range abbrevMap {
		reverseMap[new] = old
	}

	for old, new := range reverseMap {
		text = strings.ReplaceAll(text, old, new)
	}

	// Reverse symbols
	reverseSymbols := map[string]string{
		"→": " implements ",
		"←": " uses ",
		"&": " and ",
		"|": " or ",
		"!": " not ",
		"~": " approximately ",
		"@": " at ",
	}

	for old, new := range reverseSymbols {
		text = strings.ReplaceAll(text, old, new)
	}

	return text
}
