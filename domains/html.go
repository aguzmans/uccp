package domains

import (
	"regexp"
	"strings"

	"github.com/aguzmans/uccp/core"
)

// Pre-compiled regexes for HTML-to-markdown conversion.
var (
	// Noise blocks to remove entirely
	htmlNoiseRe = map[string]*regexp.Regexp{
		"script": regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`),
		"style":  regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`),
		"nav":    regexp.MustCompile(`(?is)<nav[^>]*>.*?</nav>`),
		"footer": regexp.MustCompile(`(?is)<footer[^>]*>.*?</footer>`),
		"svg":    regexp.MustCompile(`(?is)<svg[^>]*>.*?</svg>`),
		"noscript": regexp.MustCompile(`(?is)<noscript[^>]*>.*?</noscript>`),
	}

	// Structural replacements (order matters)
	htmlCodeBlockLangRe = regexp.MustCompile(`(?is)<pre[^>]*>\s*<code[^>]*class="[^"]*language-([^"\s]+)[^"]*"[^>]*>(.*?)</code>\s*</pre>`)
	htmlCodeBlockRe     = regexp.MustCompile(`(?is)<pre[^>]*>\s*<code[^>]*>(.*?)</code>\s*</pre>`)
	htmlPreRe           = regexp.MustCompile(`(?is)<pre[^>]*>(.*?)</pre>`)
	htmlInlineCodeRe    = regexp.MustCompile(`(?is)<code[^>]*>(.*?)</code>`)

	htmlH1Re = regexp.MustCompile(`(?is)<h1[^>]*>(.*?)</h1>`)
	htmlH2Re = regexp.MustCompile(`(?is)<h2[^>]*>(.*?)</h2>`)
	htmlH3Re = regexp.MustCompile(`(?is)<h3[^>]*>(.*?)</h3>`)
	htmlH4Re = regexp.MustCompile(`(?is)<h4[^>]*>(.*?)</h4>`)
	htmlH5Re = regexp.MustCompile(`(?is)<h5[^>]*>(.*?)</h5>`)
	htmlH6Re = regexp.MustCompile(`(?is)<h6[^>]*>(.*?)</h6>`)

	htmlLinkRe      = regexp.MustCompile(`(?is)<a[^>]*href="([^"]+)"[^>]*>(.*?)</a>`)
	htmlImgRe       = regexp.MustCompile(`(?is)<img[^>]*alt="([^"]*)"[^>]*/?>`)
	htmlListItemRe  = regexp.MustCompile(`(?is)<li[^>]*>(.*?)</li>`)
	htmlTableRowRe  = regexp.MustCompile(`(?is)<tr[^>]*>(.*?)</tr>`)
	htmlTableCellRe = regexp.MustCompile(`(?is)<t[hd][^>]*>(.*?)</t[hd]>`)
	htmlParagraphRe = regexp.MustCompile(`(?is)<p[^>]*>(.*?)</p>`)
	htmlDivNoteRe   = regexp.MustCompile(`(?is)<div[^>]*class="[^"]*note[^"]*"[^>]*>(.*?)</div>`)
	htmlStrongRe    = regexp.MustCompile(`(?is)<(?:strong|b)[^>]*>(.*?)</(?:strong|b)>`)
	htmlEmRe        = regexp.MustCompile(`(?is)<(?:em|i)[^>]*>(.*?)</(?:em|i)>`)
	htmlBrRe        = regexp.MustCompile(`(?is)<br\s*/?>`)

	// Cleanup
	htmlAllTagsRe       = regexp.MustCompile(`<[^>]+>`)
	htmlMultiBlankRe    = regexp.MustCompile(`\n{3,}`)
	htmlMultiSpaceRe    = regexp.MustCompile(`[ \t]+`)
	htmlTrailingSpaceRe = regexp.MustCompile(`(?m)[ \t]+$`)

	// Abbreviations
	htmlAbbrevMap = map[string]string{
		"Performance":    "Prf",
		"performance":    "prf",
		"Function":       "fn",
		"function":       "fn",
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
		"Authentication": "auth",
		"authentication": "auth",
		"Authorization":  "authz",
		"authorization":  "authz",
	}
	htmlArticleRe = regexp.MustCompile(`\b(the|a|an)\s`)
)

// HTMLCompressor converts HTML to compact markdown-like format.
// Preserves semantic structure (headings, lists, code, tables, links)
// while stripping all tag overhead, attributes, styles, and noise.
type HTMLCompressor struct{}

func NewHTMLCompressor() *HTMLCompressor {
	return &HTMLCompressor{}
}

// Compress converts HTML to compact markdown, preserving document order
// and semantic tag types while eliminating all markup overhead.
func (h *HTMLCompressor) Compress(content string) (string, error) {
	md := htmlToMarkdown(content)
	md = applyAbbreviations(md)
	return md, nil
}

func (h *HTMLCompressor) Decompress(compressed string) (string, error) {
	// Already readable markdown — return as-is
	return compressed, nil
}

func (h *HTMLCompressor) SystemPrompt() string {
	return `Content compressed from HTML to compact markdown. Semantic structure preserved:
# H1, ## H2, ### H3 = headings. - = list items. ` + "```" + `lang = code blocks.
| col | col | = tables. [text](url) = links. **bold** *italic* = emphasis.
Abbreviations: Prf=Performance fn=function impl=implementation cfg=configuration param=parameter repo=repository app=application dev=development prod=production env=environment auth=authentication authz=authorization. Articles (the/a/an) removed.`
}

func (h *HTMLCompressor) EstimateTokens(content string) int {
	return core.EstimateTokenCount(content)
}

// htmlToMarkdown converts HTML to compact markdown in document order.
func htmlToMarkdown(html string) string {
	// 1. Remove noise blocks
	for _, tag := range []string{"script", "style", "nav", "footer", "svg", "noscript"} {
		html = htmlNoiseRe[tag].ReplaceAllString(html, "")
	}

	// 2. Remove <head> entirely
	headRe := regexp.MustCompile(`(?is)<head[^>]*>.*?</head>`)
	html = headRe.ReplaceAllString(html, "")

	// 3. Convert code blocks FIRST (before stripping other tags inside them)
	html = htmlCodeBlockLangRe.ReplaceAllStringFunc(html, func(match string) string {
		parts := htmlCodeBlockLangRe.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		lang := strings.ToLower(parts[1])
		code := decodeHTMLEntities(strings.TrimSpace(parts[2]))
		// Strip comments from code blocks
		code = stripCodeComments(code)
		return "\n```" + lang + "\n" + code + "\n```\n"
	})
	html = htmlCodeBlockRe.ReplaceAllStringFunc(html, func(match string) string {
		parts := htmlCodeBlockRe.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		code := decodeHTMLEntities(strings.TrimSpace(parts[1]))
		code = stripCodeComments(code)
		return "\n```\n" + code + "\n```\n"
	})
	html = htmlPreRe.ReplaceAllStringFunc(html, func(match string) string {
		parts := htmlPreRe.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		code := decodeHTMLEntities(strings.TrimSpace(parts[1]))
		return "\n```\n" + code + "\n```\n"
	})

	// 4. Convert structural elements to markdown
	html = htmlH1Re.ReplaceAllStringFunc(html, func(m string) string {
		return "\n# " + cleanInline(htmlH1Re.FindStringSubmatch(m)[1]) + "\n"
	})
	html = htmlH2Re.ReplaceAllStringFunc(html, func(m string) string {
		return "\n## " + cleanInline(htmlH2Re.FindStringSubmatch(m)[1]) + "\n"
	})
	html = htmlH3Re.ReplaceAllStringFunc(html, func(m string) string {
		return "\n### " + cleanInline(htmlH3Re.FindStringSubmatch(m)[1]) + "\n"
	})
	html = htmlH4Re.ReplaceAllStringFunc(html, func(m string) string {
		return "\n#### " + cleanInline(htmlH4Re.FindStringSubmatch(m)[1]) + "\n"
	})
	html = htmlH5Re.ReplaceAllStringFunc(html, func(m string) string {
		return "\n##### " + cleanInline(htmlH5Re.FindStringSubmatch(m)[1]) + "\n"
	})
	html = htmlH6Re.ReplaceAllStringFunc(html, func(m string) string {
		return "\n###### " + cleanInline(htmlH6Re.FindStringSubmatch(m)[1]) + "\n"
	})

	// Notes/callouts
	html = htmlDivNoteRe.ReplaceAllStringFunc(html, func(m string) string {
		parts := htmlDivNoteRe.FindStringSubmatch(m)
		if len(parts) < 2 {
			return m
		}
		return "\n> " + cleanInline(parts[1]) + "\n"
	})

	// Links → [text](url)
	html = htmlLinkRe.ReplaceAllStringFunc(html, func(m string) string {
		parts := htmlLinkRe.FindStringSubmatch(m)
		if len(parts) < 3 {
			return m
		}
		text := cleanInline(parts[2])
		href := parts[1]
		if text == "" {
			return href
		}
		// Skip anchor-only links
		if strings.HasPrefix(href, "#") {
			return text
		}
		return "[" + text + "](" + href + ")"
	})

	// Images → alt text only
	html = htmlImgRe.ReplaceAllString(html, "$1")

	// Inline formatting
	html = htmlStrongRe.ReplaceAllString(html, "**$1**")
	html = htmlEmRe.ReplaceAllString(html, "*$1*")
	html = htmlInlineCodeRe.ReplaceAllString(html, "`$1`")
	html = htmlBrRe.ReplaceAllString(html, "\n")

	// Tables → markdown tables
	html = convertTables(html)

	// List items → markdown
	html = htmlListItemRe.ReplaceAllStringFunc(html, func(m string) string {
		parts := htmlListItemRe.FindStringSubmatch(m)
		if len(parts) < 2 {
			return m
		}
		text := cleanInline(parts[1])
		if text == "" {
			return ""
		}
		return "\n- " + text
	})

	// Paragraphs → double newline separated text
	html = htmlParagraphRe.ReplaceAllStringFunc(html, func(m string) string {
		parts := htmlParagraphRe.FindStringSubmatch(m)
		if len(parts) < 2 {
			return m
		}
		text := cleanInline(parts[1])
		if text == "" {
			return ""
		}
		return "\n" + text + "\n"
	})

	// 5. Strip all remaining tags
	html = htmlAllTagsRe.ReplaceAllString(html, "")

	// 6. Decode remaining entities
	html = decodeHTMLEntities(html)

	// 7. Clean up whitespace
	html = htmlMultiSpaceRe.ReplaceAllString(html, " ")
	html = htmlTrailingSpaceRe.ReplaceAllString(html, "")
	html = htmlMultiBlankRe.ReplaceAllString(html, "\n\n")

	return strings.TrimSpace(html)
}

// convertTables converts HTML tables to markdown tables.
func convertTables(html string) string {
	return htmlTableRowRe.ReplaceAllStringFunc(html, func(rowMatch string) string {
		parts := htmlTableRowRe.FindStringSubmatch(rowMatch)
		if len(parts) < 2 {
			return rowMatch
		}
		cells := htmlTableCellRe.FindAllStringSubmatch(parts[1], -1)
		if len(cells) == 0 {
			return ""
		}
		var cellTexts []string
		for _, cell := range cells {
			if len(cell) > 1 {
				cellTexts = append(cellTexts, cleanInline(cell[1]))
			}
		}
		return "\n| " + strings.Join(cellTexts, " | ") + " |"
	})
}

// cleanInline strips HTML tags from inline content and normalizes whitespace.
func cleanInline(s string) string {
	s = htmlAllTagsRe.ReplaceAllString(s, " ")
	s = decodeHTMLEntities(s)
	s = htmlMultiSpaceRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// stripCodeComments removes comments from code blocks for compactness.
var (
	codeBlockCommentRe = regexp.MustCompile(`(?s)/\*.*?\*/`)
	codeLineCommentRe  = regexp.MustCompile(`(?m)^\s*//[^\n]*\n?`)
	codeBlankLinesRe   = regexp.MustCompile(`\n{3,}`)
)

func stripCodeComments(code string) string {
	code = codeBlockCommentRe.ReplaceAllString(code, "")
	code = codeLineCommentRe.ReplaceAllString(code, "")
	code = codeBlankLinesRe.ReplaceAllString(code, "\n")
	return strings.TrimSpace(code)
}

// applyAbbreviations shortens common terms and removes articles.
func applyAbbreviations(text string) string {
	for long, short := range htmlAbbrevMap {
		text = strings.ReplaceAll(text, long, short)
	}
	text = htmlArticleRe.ReplaceAllString(text, "")
	// Clean up double spaces from article removal
	text = htmlMultiSpaceRe.ReplaceAllString(text, " ")
	return text
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
