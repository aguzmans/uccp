package benchmark

import (
	"fmt"
	"os"
	"strings"
)

// BenchmarkResult holds the outcome of a single benchmark run.
type BenchmarkResult struct {
	FileName           string
	Category           string
	OriginalTokens     int
	CompressedTokens   int
	NetTokens          int // compressed tokens + system prompt overhead
	ByteCompressionPct float64
	TokenCompressionPct float64
	NetSavingsPct      float64
}

// GenerateGraph produces an SVG bar chart from benchmark results and writes it
// to outputPath. The chart is a grouped horizontal bar chart comparing original,
// compressed, and net token counts for each test file.
func GenerateGraph(results []BenchmarkResult, outputPath string) error {
	if len(results) == 0 {
		return fmt.Errorf("no benchmark results to graph")
	}

	// ── Layout constants ──────────────────────────────────────────────
	const (
		svgWidth     = 900
		leftMargin   = 220 // space for labels
		rightMargin  = 80
		topMargin    = 90  // space for title + subtitle
		bottomMargin = 130 // space for summary box
		groupGap     = 20  // gap between groups
		barHeight    = 22
		barGap       = 4 // gap between bars within a group
		barsPerGroup = 3
	)

	chartWidth := svgWidth - leftMargin - rightMargin
	groupHeight := barsPerGroup*barHeight + (barsPerGroup-1)*barGap
	totalGroupHeight := groupHeight + groupGap
	chartHeight := len(results)*totalGroupHeight - groupGap
	svgHeight := topMargin + chartHeight + bottomMargin

	// ── Find the max token count for scaling ──────────────────────────
	maxTokens := 0
	for _, r := range results {
		if r.OriginalTokens > maxTokens {
			maxTokens = r.OriginalTokens
		}
	}
	if maxTokens == 0 {
		maxTokens = 1 // avoid division by zero
	}

	// ── Compute summary statistics ────────────────────────────────────
	var sumBytePct, sumTokenPct, sumNetPct float64
	for _, r := range results {
		sumBytePct += r.ByteCompressionPct
		sumTokenPct += r.TokenCompressionPct
		sumNetPct += r.NetSavingsPct
	}
	n := float64(len(results))
	avgBytePct := sumBytePct / n
	avgTokenPct := sumTokenPct / n
	avgNetPct := sumNetPct / n

	// ── Colors ────────────────────────────────────────────────────────
	const (
		colorOriginal   = "#94a3b8"
		colorCompressed = "#3b82f6"
		colorNet        = "#22c55e"
		colorBg         = "#ffffff"
		colorText       = "#1e293b"
		colorGrid       = "#e2e8f0"
		colorSubtext    = "#64748b"
	)

	// ── Build SVG ─────────────────────────────────────────────────────
	var b strings.Builder

	// Header
	b.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, svgWidth, svgHeight, svgWidth, svgHeight))
	b.WriteString("\n")

	// Background
	b.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="%s"/>`, svgWidth, svgHeight, colorBg))
	b.WriteString("\n")

	// Styles
	b.WriteString(`<style>`)
	b.WriteString(fmt.Sprintf(`
    .title { font-family: Arial, Helvetica, sans-serif; font-size: 20px; font-weight: bold; fill: %s; }
    .subtitle { font-family: Arial, Helvetica, sans-serif; font-size: 13px; fill: %s; }
    .label { font-family: Arial, Helvetica, sans-serif; font-size: 12px; fill: %s; }
    .label-cat { font-family: Arial, Helvetica, sans-serif; font-size: 10px; fill: %s; }
    .value { font-family: Arial, Helvetica, sans-serif; font-size: 11px; fill: %s; }
    .legend-text { font-family: Arial, Helvetica, sans-serif; font-size: 11px; fill: %s; }
    .summary-title { font-family: Arial, Helvetica, sans-serif; font-size: 13px; font-weight: bold; fill: %s; }
    .summary-text { font-family: Arial, Helvetica, sans-serif; font-size: 12px; fill: %s; }
    .summary-note { font-family: Arial, Helvetica, sans-serif; font-size: 10px; fill: %s; font-style: italic; }
  `, colorText, colorSubtext, colorText, colorSubtext, colorText, colorText, colorText, colorText, colorSubtext))
	b.WriteString(`</style>`)
	b.WriteString("\n")

	// Title and subtitle
	b.WriteString(fmt.Sprintf(`<text x="%d" y="35" class="title">UCCP Compression Benchmarks</text>`, leftMargin))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`<text x="%d" y="55" class="subtitle">Token savings measured with cl100k_base tokenizer</text>`, leftMargin))
	b.WriteString("\n")

	// Legend (right-aligned, stacked vertically)
	legendX := svgWidth - rightMargin - 230
	legendY := 20
	legends := []struct {
		color string
		label string
	}{
		{colorOriginal, "Original tokens"},
		{colorCompressed, "Compressed tokens"},
		{colorNet, "Net tokens (with prompt overhead)"},
	}
	for i, lg := range legends {
		ly := legendY + i*16
		b.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="12" height="12" rx="2" fill="%s"/>`, legendX, ly, lg.color))
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="legend-text">%s</text>`, legendX+16, ly+10, lg.label))
		b.WriteString("\n")
	}

	// ── Grid lines ────────────────────────────────────────────────────
	gridSteps := 5
	for i := 0; i <= gridSteps; i++ {
		x := leftMargin + int(float64(chartWidth)*float64(i)/float64(gridSteps))
		y1 := topMargin
		y2 := topMargin + chartHeight
		b.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="1"/>`, x, y1, x, y2, colorGrid))
		b.WriteString("\n")

		// Grid label
		tokenVal := int(float64(maxTokens) * float64(i) / float64(gridSteps))
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="label-cat" text-anchor="middle">%s</text>`, x, y1-5, formatInt(tokenVal)))
		b.WriteString("\n")
	}

	// ── Bars ──────────────────────────────────────────────────────────
	for i, r := range results {
		groupY := topMargin + i*totalGroupHeight

		// File name label
		labelY := groupY + groupHeight/2
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="label" text-anchor="end" dominant-baseline="central">%s</text>`, leftMargin-10, labelY-7, escapeXML(r.FileName)))
		b.WriteString("\n")

		// Category label (below file name)
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="label-cat" text-anchor="end" dominant-baseline="central">%s</text>`, leftMargin-10, labelY+8, escapeXML(r.Category)))
		b.WriteString("\n")

		// Bar data: value, color, pct label
		bars := []struct {
			value int
			color string
			pct   string
		}{
			{r.OriginalTokens, colorOriginal, ""},
			{r.CompressedTokens, colorCompressed, fmt.Sprintf("-%.0f%%", r.TokenCompressionPct)},
			{r.NetTokens, colorNet, fmt.Sprintf("-%.0f%% net", r.NetSavingsPct)},
		}

		for j, bar := range bars {
			barY := groupY + j*(barHeight+barGap)
			barW := int(float64(chartWidth) * float64(bar.value) / float64(maxTokens))
			if barW < 1 && bar.value > 0 {
				barW = 1
			}

			// Bar rectangle
			b.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="3" fill="%s"/>`,
				leftMargin, barY, barW, barHeight, bar.color))
			b.WriteString("\n")

			// Value label at end of bar
			valLabel := formatInt(bar.value)
			if bar.pct != "" {
				valLabel = fmt.Sprintf("%s (%s)", valLabel, bar.pct)
			}
			textX := leftMargin + barW + 6
			b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="value" dominant-baseline="central">%s</text>`,
				textX, barY+barHeight/2, valLabel))
			b.WriteString("\n")
		}
	}

	// ── Summary box ───────────────────────────────────────────────────
	summaryY := topMargin + chartHeight + 25
	boxX := leftMargin
	boxW := chartWidth
	boxH := 90

	// Box background
	b.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="6" fill="#f8fafc" stroke="%s" stroke-width="1"/>`,
		boxX, summaryY, boxW, boxH, colorGrid))
	b.WriteString("\n")

	// Summary title
	b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="summary-title">Summary (averages across %d files)</text>`,
		boxX+15, summaryY+22, len(results)))
	b.WriteString("\n")

	// Summary stats in columns
	col1X := boxX + 15
	col2X := boxX + boxW/3
	col3X := boxX + 2*boxW/3
	statsY := summaryY + 45

	b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="summary-text">Byte compression: %.1f%%</text>`,
		col1X, statsY, avgBytePct))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="summary-text">Token compression: %.1f%%</text>`,
		col2X, statsY, avgTokenPct))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="summary-text">Net token savings: %.1f%%</text>`,
		col3X, statsY, avgNetPct))
	b.WriteString("\n")

	// Note
	b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="summary-note">Measured with tiktoken cl100k_base</text>`,
		col1X, statsY+25))
	b.WriteString("\n")

	// Close SVG
	b.WriteString("</svg>\n")

	// Write to file
	return os.WriteFile(outputPath, []byte(b.String()), 0644)
}

// formatInt formats an integer with comma separators (e.g., 1,234).
func formatInt(n int) string {
	if n < 0 {
		return "-" + formatInt(-n)
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result strings.Builder
	remainder := len(s) % 3
	if remainder > 0 {
		result.WriteString(s[:remainder])
	}
	for i := remainder; i < len(s); i += 3 {
		if result.Len() > 0 {
			result.WriteByte(',')
		}
		result.WriteString(s[i : i+3])
	}
	return result.String()
}

// escapeXML escapes special characters for safe embedding in SVG/XML text.
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
