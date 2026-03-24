package domains

import (
	"regexp"
	"strings"

	"github.com/aguzmans/uccp/core"
)

// FinancialCompressor compresses plaintext financial content such as
// market summaries, predictions, and continuity context for financial blogs.
type FinancialCompressor struct{}

// NewFinancialCompressor creates a new financial domain compressor.
func NewFinancialCompressor() *FinancialCompressor {
	return &FinancialCompressor{}
}

// Compress converts financial text to UCCP format.
// Pipeline: term abbreviations → date compression → prediction symbols →
// price notation → phrase compression → whitespace normalization.
func (f *FinancialCompressor) Compress(content string) (string, error) {
	result := content

	// 1. Financial term abbreviations (case-insensitive, longest first)
	for _, pair := range finTerms {
		result = pair.re.ReplaceAllStringFunc(result, func(m string) string {
			return pair.replacement
		})
	}

	// 2. Date compression: "March 23, 2026" → "Mar23'26"
	result = finDateRe.ReplaceAllStringFunc(result, compressDate)

	// 3. Prediction status symbols
	for _, pair := range finPredictionTerms {
		result = pair.re.ReplaceAllString(result, pair.replacement)
	}

	// 4. Price notation: strip commas from numbers, remove "closed at" etc.
	result = finNumberCommaRe.ReplaceAllStringFunc(result, func(m string) string {
		return strings.ReplaceAll(m, ",", "")
	})
	for _, re := range finPriceVerbRe {
		result = re.ReplaceAllString(result, "")
	}

	// 5. Common phrase compression
	for _, pair := range finPhraseTerms {
		result = pair.re.ReplaceAllString(result, pair.replacement)
	}

	// 6. Whitespace normalization
	result = finMultiSpaceRe.ReplaceAllString(result, " ")
	result = finMultiNewlineRe.ReplaceAllString(result, "\n")
	// Trim each line
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	result = strings.Join(lines, "\n")
	result = strings.TrimSpace(result)

	return result, nil
}

// Decompress converts UCCP financial format back to readable content.
// Note: This is lossy — not all original phrasing is recovered.
func (f *FinancialCompressor) Decompress(compressed string) (string, error) {
	result := compressed

	// Reverse financial term abbreviations (main ones)
	for _, pair := range finDecompressTerms {
		result = pair.re.ReplaceAllString(result, pair.replacement)
	}

	// Reverse prediction symbols
	result = strings.ReplaceAll(result, "⏳", "pending")
	result = strings.ReplaceAll(result, "✓", "confirmed")
	result = strings.ReplaceAll(result, "✗", "invalidated")
	result = strings.ReplaceAll(result, "⌛", "expired")
	result = strings.ReplaceAll(result, "◐", "partial")

	// Reverse phrase compressions
	result = strings.ReplaceAll(result, "↑", "increased by")
	result = strings.ReplaceAll(result, "↓", "decreased by")
	result = strings.ReplaceAll(result, "→", "unchanged")

	// Reverse date compression: "Mar23'26" → "March 23, 2026"
	result = finCompressedDateRe.ReplaceAllStringFunc(result, decompressDate)

	// Reverse simple phrase replacements
	for _, pair := range finDecompressPhrases {
		result = pair.re.ReplaceAllString(result, pair.replacement)
	}

	return result, nil
}

// SystemPrompt returns the LLM prompt explaining UCCP financial compression format.
func (f *FinancialCompressor) SystemPrompt() string {
	return `UCCP (Ultra-Compact Content Protocol) - Financial Domain

You are reading content in UCCP format - an ultra-compact format for financial text designed to save tokens.

INDEX ABBREVIATIONS:
SPX=S&P 500, NDX=Nasdaq Composite, DJIA=Dow Jones Industrial Average
Fed=Federal Reserve, ECB=European Central Bank
WTI=West Texas Intermediate, oil=crude oil

METRIC ABBREVIATIONS:
YoY=year-over-year, QoQ=quarter-over-quarter, MoM=month-over-month
pp=percentage points, bp=basis points, YTD=year to date
EPS=earnings per share, P/E=price-to-earnings, MA=moving average, mktcap=market capitalization
ATH=all-time high, ATL=all-time low, 52wH=52-week high, 52wL=52-week low

UNIT ABBREVIATIONS:
B=billion, M=million, T=trillion, /bbl=per barrel, /oz=per ounce

DATE FORMAT:
Mar23'26=March 23, 2026 (MonDD'YY)

PREDICTION STATUS:
⏳=pending, ✓=confirmed, ✗=invalidated, ⌛=expired, ◐=partial

PHRASE SYMBOLS:
@=as of, vs=compared to, per=according to
↑=increased by, ↓=decreased by, →=unchanged, ~=approximately

COMPRESSION RULES:
- Commas stripped from numbers ($4,365.90 → $4365.90)
- "closed at"/"settled at"/"trading at" removed before prices
- Whitespace collapsed, lines trimmed
- Articles (the, a, an) NOT removed (financial text needs precision)

When reading UCCP financial format, decode abbreviations and symbols mentally.
The content is intentionally compact to save tokens while remaining readable.`
}

// EstimateTokens estimates token count for financial content.
func (f *FinancialCompressor) EstimateTokens(content string) int {
	return core.EstimateTokenCount(content)
}

// --- regex and replacement definitions ---

type finReplacePair struct {
	re          *regexp.Regexp
	replacement string
}

// Financial term abbreviations, ordered longest-match-first to avoid partial replacements.
var finTerms = []finReplacePair{
	{regexp.MustCompile(`(?i)Dow Jones Industrial Average`), "DJIA"},
	{regexp.MustCompile(`(?i)Nasdaq Composite`), "NDX"},
	{regexp.MustCompile(`(?i)West Texas Intermediate`), "WTI"},
	{regexp.MustCompile(`(?i)European Central Bank`), "ECB"},
	{regexp.MustCompile(`(?i)market capitalization`), "mktcap"},
	{regexp.MustCompile(`(?i)quarter-over-quarter`), "QoQ"},
	{regexp.MustCompile(`(?i)earnings per share`), "EPS"},
	{regexp.MustCompile(`(?i)price-to-earnings`), "P/E"},
	{regexp.MustCompile(`(?i)year-over-year`), "YoY"},
	{regexp.MustCompile(`(?i)month-over-month`), "MoM"},
	{regexp.MustCompile(`(?i)percentage points`), "pp"},
	{regexp.MustCompile(`(?i)Federal Reserve`), "Fed"},
	{regexp.MustCompile(`(?i)year[\s-]to[\s-]date`), "YTD"},
	{regexp.MustCompile(`(?i)moving average`), "MA"},
	{regexp.MustCompile(`(?i)trading session`), "session"},
	{regexp.MustCompile(`(?i)basis points`), "bp"},
	{regexp.MustCompile(`(?i)all-time high`), "ATH"},
	{regexp.MustCompile(`(?i)all-time low`), "ATL"},
	{regexp.MustCompile(`(?i)52-week high`), "52wH"},
	{regexp.MustCompile(`(?i)52-week low`), "52wL"},
	{regexp.MustCompile(`(?i)Dow Jones`), "DJIA"},
	{regexp.MustCompile(`(?i)crude oil`), "oil"},
	{regexp.MustCompile(`(?i)per barrel`), "/bbl"},
	{regexp.MustCompile(`(?i)per ounce`), "/oz"},
	{regexp.MustCompile(`(?i)S&P 500`), "SPX"},
	{regexp.MustCompile(`(?i)billion`), "B"},
	{regexp.MustCompile(`(?i)million`), "M"},
	{regexp.MustCompile(`(?i)trillion`), "T"},
}

// Date pattern: "March 23, 2026"
var finDateRe = regexp.MustCompile(`(?i)(January|February|March|April|May|June|July|August|September|October|November|December)\s+(\d{1,2}),\s*(\d{4})`)

var finMonthAbbrev = map[string]string{
	"january": "Jan", "february": "Feb", "march": "Mar", "april": "Apr",
	"may": "May", "june": "Jun", "july": "Jul", "august": "Aug",
	"september": "Sep", "october": "Oct", "november": "Nov", "december": "Dec",
}

var finMonthFull = map[string]string{
	"Jan": "January", "Feb": "February", "Mar": "March", "Apr": "April",
	"May": "May", "Jun": "June", "Jul": "July", "Aug": "August",
	"Sep": "September", "Oct": "October", "Nov": "November", "Dec": "December",
}

func compressDate(match string) string {
	parts := finDateRe.FindStringSubmatch(match)
	if len(parts) < 4 {
		return match
	}
	month := strings.ToLower(parts[1])
	day := parts[2]
	year := parts[3]
	abbrev, ok := finMonthAbbrev[month]
	if !ok {
		return match
	}
	// Last two digits of year
	shortYear := year
	if len(year) == 4 {
		shortYear = year[2:]
	}
	return abbrev + day + "'" + shortYear
}

// Compressed date pattern: "Mar23'26"
var finCompressedDateRe = regexp.MustCompile(`(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)(\d{1,2})'(\d{2})`)

func decompressDate(match string) string {
	parts := finCompressedDateRe.FindStringSubmatch(match)
	if len(parts) < 4 {
		return match
	}
	abbrev := parts[1]
	day := parts[2]
	shortYear := parts[3]
	full, ok := finMonthFull[abbrev]
	if !ok {
		return match
	}
	return full + " " + day + ", 20" + shortYear
}

// Prediction status terms (matched in context-free way for simplicity)
var finPredictionTerms = []finReplacePair{
	{regexp.MustCompile(`\bpending\b`), "⏳"},
	{regexp.MustCompile(`\bconfirmed\b`), "✓"},
	{regexp.MustCompile(`\binvalidated\b`), "✗"},
	{regexp.MustCompile(`\bexpired\b`), "⌛"},
	{regexp.MustCompile(`\bpartial\b`), "◐"},
}

// Number comma stripping: match numbers with commas like 4,365.90
var finNumberCommaRe = regexp.MustCompile(`\d{1,3}(?:,\d{3})+(?:\.\d+)?`)

// Price verb phrases to remove
var finPriceVerbRe = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bclosed at\s+`),
	regexp.MustCompile(`(?i)\bsettled at\s+`),
	regexp.MustCompile(`(?i)\btrading at\s+`),
}

// Common phrase compressions
var finPhraseTerms = []finReplacePair{
	{regexp.MustCompile(`(?i)\bcompared to\b`), "vs"},
	{regexp.MustCompile(`(?i)\baccording to\b`), "per"},
	{regexp.MustCompile(`(?i)\bincreased by\b`), "↑"},
	{regexp.MustCompile(`(?i)\bdecreased by\b`), "↓"},
	{regexp.MustCompile(`(?i)\bunchanged\b`), "→"},
	{regexp.MustCompile(`(?i)\bapproximately\b`), "~"},
	{regexp.MustCompile(`(?i)\bas of\b`), "@"},
}

// Whitespace normalization
var (
	finMultiSpaceRe   = regexp.MustCompile(`[ \t]+`)
	finMultiNewlineRe = regexp.MustCompile(`\n{3,}`)
)

// Decompression: reverse term abbreviations
var finDecompressTerms = []finReplacePair{
	{regexp.MustCompile(`\bSPX\b`), "S&P 500"},
	{regexp.MustCompile(`\bNDX\b`), "Nasdaq Composite"},
	{regexp.MustCompile(`\bDJIA\b`), "Dow Jones Industrial Average"},
	{regexp.MustCompile(`\bYoY\b`), "year-over-year"},
	{regexp.MustCompile(`\bQoQ\b`), "quarter-over-quarter"},
	{regexp.MustCompile(`\bMoM\b`), "month-over-month"},
	{regexp.MustCompile(`\bYTD\b`), "year to date"},
	{regexp.MustCompile(`\bATH\b`), "all-time high"},
	{regexp.MustCompile(`\bATL\b`), "all-time low"},
	{regexp.MustCompile(`\b52wH\b`), "52-week high"},
	{regexp.MustCompile(`\b52wL\b`), "52-week low"},
	{regexp.MustCompile(`\bFed\b`), "Federal Reserve"},
	{regexp.MustCompile(`\bECB\b`), "European Central Bank"},
	{regexp.MustCompile(`\bWTI\b`), "West Texas Intermediate"},
	{regexp.MustCompile(`\bmktcap\b`), "market capitalization"},
	{regexp.MustCompile(`\bEPS\b`), "earnings per share"},
	{regexp.MustCompile(`\bP/E\b`), "price-to-earnings"},
	{regexp.MustCompile(`\bMA\b`), "moving average"},
	{regexp.MustCompile(`\bpp\b`), "percentage points"},
	{regexp.MustCompile(`\bbp\b`), "basis points"},
	{regexp.MustCompile(`/bbl\b`), "per barrel"},
	{regexp.MustCompile(`/oz\b`), "per ounce"},
}

// Decompression: reverse phrase replacements
var finDecompressPhrases = []finReplacePair{
	{regexp.MustCompile(`\bvs\b`), "compared to"},
	{regexp.MustCompile(`\bper\b`), "according to"},
	{regexp.MustCompile(`@`), "as of"},
}
