package domains

import (
	"regexp"
	"strings"

	"github.com/aguzmans/uccp/core"
)

// Pre-compiled regexes for HTML parsing and compression
var (
	// removeNoiseBlocks: tag-specific regexes
	htmlNoiseRe = map[string]*regexp.Regexp{
		"script": regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`),
		"style":  regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`),
		"nav":    regexp.MustCompile(`(?is)<nav[^>]*>.*?</nav>`),
		"header": regexp.MustCompile(`(?is)<header[^>]*>.*?</header>`),
		"footer": regexp.MustCompile(`(?is)<footer[^>]*>.*?</footer>`),
	}

	// extractHeadings: tag-specific regexes
	htmlHeadingRe = map[string]*regexp.Regexp{
		"h1": regexp.MustCompile(`(?is)<h1[^>]*>(.*?)</h1>`),
		"h2": regexp.MustCompile(`(?is)<h2[^>]*>(.*?)</h2>`),
		"h3": regexp.MustCompile(`(?is)<h3[^>]*>(.*?)</h3>`),
		"h4": regexp.MustCompile(`(?is)<h4[^>]*>(.*?)</h4>`),
	}

	// extractCodeBlocks
	htmlCodeLangRe = regexp.MustCompile(`(?is)<pre[^>]*>\s*<code[^>]*class="[^"]*language-([^"\s]+)[^"]*"[^>]*>(.*?)</code>\s*</pre>`)
	htmlCodeRe     = regexp.MustCompile(`(?is)<pre[^>]*>\s*<code[^>]*>(.*?)</code>\s*</pre>`)

	// extractLists
	htmlListRe = regexp.MustCompile(`(?is)<li[^>]*>(.*?)</li>`)

	// extractTables
	htmlTableRowRe  = regexp.MustCompile(`(?is)<tr[^>]*>(.*?)</tr>`)
	htmlTableCellRe = regexp.MustCompile(`(?is)<t[hd][^>]*>(.*?)</t[hd]>`)

	// extractParagraphs
	htmlParagraphRe = regexp.MustCompile(`(?is)<p[^>]*>(.*?)</p>`)

	// extractLinks
	htmlLinkRe = regexp.MustCompile(`(?is)<a[^>]*href="([^"]+)"[^>]*>(.*?)</a>`)

	// htmlCompress
	htmlArticleRe    = regexp.MustCompile(`\b(the|a|an)\s`)
	htmlWhitespaceRe = regexp.MustCompile(`\s+`)

	// htmlCompressCode
	htmlBlankLineRe = regexp.MustCompile(`\n\s*\n`)

	// cleanText
	htmlTagsRe = regexp.MustCompile(`<[^>]*>`)
)

// HTMLCompressor compresses HTML and web content into UCCP format.
// Optimized for web scraping and content extraction use cases where
// HTML pages need to be fed to LLMs with maximum token efficiency.
type HTMLCompressor struct{}

// NewHTMLCompressor creates a new HTML domain compressor
func NewHTMLCompressor() *HTMLCompressor {
	return &HTMLCompressor{}
}

// Compress converts HTML content to UCCP pipe-delimited format.
// Parses headings, paragraphs, code blocks, lists, tables, and links
// into compact type:data records separated by |.
func (h *HTMLCompressor) Compress(content string) (string, error) {
	nodes := parseHTML(content)

	var records []string
	for _, node := range nodes {
		record := encodeNode(node)
		if record != "" {
			records = append(records, record)
		}
	}

	return strings.Join(records, "|"), nil
}

// Decompress converts UCCP format back to readable text.
// Note: This is lossy — HTML structure, articles, and some formatting are not recovered.
func (h *HTMLCompressor) Decompress(compressed string) (string, error) {
	records := strings.Split(compressed, "|")
	var lines []string

	for _, record := range records {
		parts := strings.SplitN(record, ":", 2)
		if len(parts) != 2 {
			continue
		}
		typeCode := parts[0]
		data := parts[1]

		switch typeCode {
		case "1":
			lines = append(lines, "# "+data)
		case "2":
			lines = append(lines, "## "+data)
		case "3":
			lines = append(lines, "### "+data)
		case "t":
			lines = append(lines, data)
		case "l":
			lines = append(lines, "- "+data)
		case "d":
			lines = append(lines, "| "+strings.ReplaceAll(data, ",", " | ")+" |")
		case "u":
			lines = append(lines, "[link]("+data+")")
		default:
			// Code blocks
			lines = append(lines, "```\n"+data+"\n```")
		}
	}

	return strings.Join(lines, "\n"), nil
}

// SystemPrompt returns the LLM prompt explaining UCCP HTML compression format
func (h *HTMLCompressor) SystemPrompt() string {
	return `
ULTRA-COMPACT CONTENT PROTOCOL (UCCP) - HTML Domain:
Format: type:data|type:data where | separates records
Types: 1=H1 2=H2 3=H3 t=text c=C p=Python j=Java g=Go s=JS q=SQL r=Rust b=Bash l=list d=table m=metric x=comparison k=key-value u=URL
Abbrev: Prf=Performance fn=function int=Integer PK=PrimaryKey qry=query db=database ms=millisecond bmk=benchmark impl=implementation cfg=config err=error rsp=response req=request auth=authentication app=application dev=development prod=production env=environment var=variable param=parameter repo=repository authz=authorization
Symbols: >=faster <=slower ~=approx ↑=increase ↓=decrease v=versus &=and !=fails 1°=primary 2°=secondary ret=return
`
}

// EstimateTokens estimates token count for HTML content
func (h *HTMLCompressor) EstimateTokens(content string) int {
	return core.EstimateTokenCount(content)
}

// --- HTML Node types and parser ---

// htmlNode represents a parsed HTML element
type htmlNode struct {
	Type     string     // h1, h2, h3, p, pre, code, li, tr, a
	Text     string     // Text content
	Language string     // For code blocks (python, go, c, etc.)
	Href     string     // For links
	Children []htmlNode // For nested elements (like table cells)
}

// parseHTML converts an HTML string into structured nodes.
// This is a lightweight regex-based parser optimized for article content.
func parseHTML(html string) []htmlNode {
	var nodes []htmlNode

	html = removeNoiseBlocks(html)

	nodes = append(nodes, extractHeadings(html, "h1")...)
	nodes = append(nodes, extractHeadings(html, "h2")...)
	nodes = append(nodes, extractHeadings(html, "h3")...)
	nodes = append(nodes, extractHeadings(html, "h4")...)
	nodes = append(nodes, extractCodeBlocks(html)...)
	nodes = append(nodes, extractLists(html)...)
	nodes = append(nodes, extractTables(html)...)
	nodes = append(nodes, extractParagraphs(html)...)
	nodes = append(nodes, extractLinks(html)...)

	return nodes
}

func removeNoiseBlocks(html string) string {
	for _, tag := range []string{"script", "style", "nav", "header", "footer"} {
		html = htmlNoiseRe[tag].ReplaceAllString(html, "")
	}
	return html
}

func extractHeadings(html, tag string) []htmlNode {
	var nodes []htmlNode
	re := htmlHeadingRe[tag]
	for _, match := range re.FindAllStringSubmatch(html, -1) {
		if len(match) > 1 {
			text := cleanText(match[1])
			if text != "" {
				nodes = append(nodes, htmlNode{Type: tag, Text: text})
			}
		}
	}
	return nodes
}

func extractCodeBlocks(html string) []htmlNode {
	var nodes []htmlNode

	// <pre><code class="language-X">...</code></pre>
	for _, match := range htmlCodeLangRe.FindAllStringSubmatch(html, -1) {
		if len(match) > 2 {
			nodes = append(nodes, htmlNode{
				Type:     "code",
				Text:     decodeHTMLEntities(match[2]),
				Language: match[1],
			})
		}
	}

	// <pre><code>...</code></pre> (no language)
	for _, match := range htmlCodeRe.FindAllStringSubmatch(html, -1) {
		if len(match) > 1 {
			code := decodeHTMLEntities(match[1])
			alreadyFound := false
			for _, existing := range nodes {
				if existing.Type == "code" && existing.Text == code {
					alreadyFound = true
					break
				}
			}
			if !alreadyFound {
				nodes = append(nodes, htmlNode{Type: "code", Text: code})
			}
		}
	}

	return nodes
}

func extractLists(html string) []htmlNode {
	var nodes []htmlNode
	re := htmlListRe
	for _, match := range re.FindAllStringSubmatch(html, -1) {
		if len(match) > 1 {
			text := cleanText(match[1])
			if text != "" {
				nodes = append(nodes, htmlNode{Type: "li", Text: text})
			}
		}
	}
	return nodes
}

func extractTables(html string) []htmlNode {
	var nodes []htmlNode
	reTr := htmlTableRowRe
	reCell := htmlTableCellRe

	for _, match := range reTr.FindAllStringSubmatch(html, -1) {
		if len(match) > 1 {
			var children []htmlNode
			for _, cellMatch := range reCell.FindAllStringSubmatch(match[1], -1) {
				if len(cellMatch) > 1 {
					children = append(children, htmlNode{
						Type: "td",
						Text: cleanText(cellMatch[1]),
					})
				}
			}
			if len(children) > 0 {
				nodes = append(nodes, htmlNode{Type: "tr", Children: children})
			}
		}
	}
	return nodes
}

func extractParagraphs(html string) []htmlNode {
	var nodes []htmlNode
	re := htmlParagraphRe
	for _, match := range re.FindAllStringSubmatch(html, -1) {
		if len(match) > 1 {
			text := cleanText(match[1])
			if text != "" && len(text) > 10 {
				nodes = append(nodes, htmlNode{Type: "p", Text: text})
			}
		}
	}
	return nodes
}

func extractLinks(html string) []htmlNode {
	var nodes []htmlNode
	re := htmlLinkRe
	for _, match := range re.FindAllStringSubmatch(html, -1) {
		if len(match) > 2 {
			href := match[1]
			text := cleanText(match[2])
			if href != "" && text != "" {
				nodes = append(nodes, htmlNode{Type: "a", Href: href, Text: text})
			}
		}
	}
	return nodes
}

// --- Encoding ---

func encodeNode(node htmlNode) string {
	switch node.Type {
	case "h1":
		return "1:" + htmlCompress(node.Text)
	case "h2":
		return "2:" + htmlCompress(node.Text)
	case "h3", "h4":
		return "3:" + htmlCompress(node.Text)
	case "p":
		text := htmlCompress(node.Text)
		if text != "" {
			return "t:" + text
		}
	case "pre", "code":
		lang := node.Language
		if lang == "" {
			lang = "c"
		}
		return htmlLangCode(lang) + ":" + htmlCompressCode(node.Text)
	case "li":
		return "l:" + htmlCompress(node.Text)
	case "tr":
		if len(node.Children) > 0 {
			var cells []string
			for _, cell := range node.Children {
				cells = append(cells, htmlCompress(cell.Text))
			}
			return "d:" + strings.Join(cells, ",")
		}
	case "a":
		if node.Href != "" && node.Text != "" {
			return "u:" + node.Href
		}
	}
	return ""
}

// --- HTML-domain compression ---

// htmlAbbrevMap contains abbreviation replacements for HTML content compression.
var htmlAbbrevMap = map[string]string{
	"Performance":    "Prf",
	"performance":    "prf",
	"Function":       "fn",
	"function":       "fn",
	"Integer":        "int",
	"PRIMARY KEY":    "PK",
	"Query":          "qry",
	"query":          "qry",
	"Database":       "db",
	"database":       "db",
	"millisecond":    "ms",
	"milliseconds":   "ms",
	"Benchmark":      "bmk",
	"Implementation": "impl",
	"implementation": "impl",
	"Configuration":  "cfg",
	"configuration":  "cfg",
	"Parameter":      "param",
	"parameter":      "param",
	"Repository":     "repo",
	"repository":     "repo",
	"Application":    "app",
	"application":    "app",
	"Development":    "dev",
	"development":    "dev",
	"Production":     "prod",
	"production":     "prod",
	"Environment":    "env",
	"environment":    "env",
	"Variable":       "var",
	"variable":       "var",
	"Authentication": "auth",
	"authentication": "auth",
	"Authorization":  "authz",
	"authorization":  "authz",
	"Response":       "rsp",
	"response":       "rsp",
	"Request":        "req",
	"request":        "req",
	"Error":          "err",
	"error":          "err",
}

// htmlSymbolMap contains symbol replacements for HTML content compression.
var htmlSymbolMap = map[string]string{
	" faster than ":   ">",
	" slower than ":   "<",
	" equal to ":      "=",
	" approximately ": "~",
	" versus ":        "v",
	" and ":           "&",
	" fails to ":      "!",
	" failed to ":     "!",
	" Primary ":       "1° ",
	" primary ":       "1° ",
	" Secondary ":     "2° ",
	" secondary ":     "2° ",
	" increase ":      "↑ ",
	" decrease ":      "↓ ",
	" decreases ":     "↓ ",
	" increases ":     "↑ ",
}

func htmlCompress(text string) string {
	for old, repl := range htmlAbbrevMap {
		text = strings.ReplaceAll(text, old, repl)
	}
	for old, repl := range htmlSymbolMap {
		text = strings.ReplaceAll(text, old, repl)
	}

	// Remove articles
	text = htmlArticleRe.ReplaceAllString(text, "")

	// Clean up punctuation and whitespace
	text = strings.ReplaceAll(text, "  ", " ")
	text = strings.ReplaceAll(text, " .", ".")
	text = strings.ReplaceAll(text, " ,", ",")
	text = htmlWhitespaceRe.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

func htmlCompressCode(code string) string {
	code = htmlBlankLineRe.ReplaceAllString(code, "\n")
	code = strings.TrimSpace(code)
	code = strings.ReplaceAll(code, "return true", "ret 1")
	code = strings.ReplaceAll(code, "return false", "ret 0")
	code = strings.ReplaceAll(code, "function ", "fn ")
	return code
}

func htmlLangCode(lang string) string {
	lang = strings.ToLower(lang)
	switch {
	case strings.Contains(lang, "python") || lang == "py":
		return "p"
	case strings.Contains(lang, "javascript") || lang == "js" || strings.Contains(lang, "typescript") || lang == "ts":
		return "s"
	case strings.Contains(lang, "java"):
		return "j"
	case strings.Contains(lang, "go") || lang == "golang":
		return "g"
	case strings.Contains(lang, "sql"):
		return "q"
	case strings.Contains(lang, "rust") || lang == "rs":
		return "r"
	case strings.Contains(lang, "bash") || lang == "sh" || lang == "shell":
		return "b"
	case strings.Contains(lang, "c++") || lang == "cpp":
		return "c"
	case lang == "c":
		return "c"
	default:
		return "c"
	}
}

// --- Text helpers ---

func cleanText(s string) string {
	s = htmlTagsRe.ReplaceAllString(s, " ")
	s = decodeHTMLEntities(s)
	s = htmlWhitespaceRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func decodeHTMLEntities(s string) string {
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", `"`)
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&#x27;", "'")
	s = strings.ReplaceAll(s, "&#x2F;", "/")
	s = strings.ReplaceAll(s, "&#8217;", "'")
	s = strings.ReplaceAll(s, "&#038;", "&")
	return s
}
