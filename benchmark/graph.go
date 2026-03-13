package benchmark

import (
	"fmt"
	"math"
	"os"
	"strings"
)

// graphPoint holds per-category, per-scale data for chart rendering.
type graphPoint struct {
	pages       int
	rawPct      float64 // raw token compression %
	amortPct    float64 // amortized net savings %
	netPct      float64 // pessimistic net savings % (N=1)
	origTok     int
	savedTokRaw int
	savedTokNet int
}

// GenerateGraph produces a dual-panel SVG chart:
//   - Top panel: Raw token compression % (no system prompt overhead)
//   - Bottom panel: Net token savings % with system prompt amortized over N messages
func GenerateGraph(results []BenchmarkResult, outputPath string) error {
	if len(results) == 0 {
		return fmt.Errorf("no benchmark results to graph")
	}

	categories := []string{"HTML Documentation", "JSON API Responses", "Source Code"}
	categoryColors := map[string]string{
		"HTML Documentation": "#3b82f6",
		"JSON API Responses": "#f59e0b",
		"Source Code":        "#22c55e",
	}
	categoryLabels := map[string]string{
		"HTML Documentation": "HTML Pages",
		"JSON API Responses": "JSON Responses",
		"Source Code":        "Source Code",
	}

	data := make(map[string][]graphPoint)
	for _, r := range results {
		rawSaved := r.OriginalTokens - r.CompressedTokens
		data[r.Category] = append(data[r.Category], graphPoint{
			pages:       r.Pages,
			rawPct:      r.TokenRatio * 100,
			amortPct:    r.AmortizedNetRatio * 100,
			netPct:      r.NetTokenRatio * 100,
			origTok:     r.OriginalTokens,
			savedTokRaw: rawSaved,
			savedTokNet: r.AmortizedNetSavings,
		})
	}

	// Layout — two stacked charts
	const (
		svgWidth    = 900
		svgHeight   = 880
		leftMargin  = 80
		rightMargin = 30
		topMargin   = 70
		chartH      = 300
		chartGap    = 100
		botMargin   = 100
	)
	chartW := svgWidth - leftMargin - rightMargin

	// X-axis shared config
	xTicks := []int{1, 5, 10, 15, 20}
	xMin, xMax := 0.0, 22.0
	xRange := xMax - xMin

	toSvgX := func(pages int) int {
		return leftMargin + int(float64(chartW)*(float64(pages)-xMin)/xRange)
	}

	// --- Compute Y ranges for each panel ---

	// Top panel: raw compression %
	rawMin, rawMax := math.MaxFloat64, -math.MaxFloat64
	for _, pts := range data {
		for _, p := range pts {
			if p.rawPct < rawMin {
				rawMin = p.rawPct
			}
			if p.rawPct > rawMax {
				rawMax = p.rawPct
			}
		}
	}
	rawMin = math.Floor(rawMin/10) * 10
	if rawMin > 0 {
		rawMin = 0
	}
	rawMax = math.Ceil(rawMax/10) * 10
	if rawMax < 20 {
		rawMax = 20
	}
	rawRange := rawMax - rawMin

	// Bottom panel: amortized net %
	netMin, netMax := math.MaxFloat64, -math.MaxFloat64
	for _, pts := range data {
		for _, p := range pts {
			if p.amortPct < netMin {
				netMin = p.amortPct
			}
			if p.amortPct > netMax {
				netMax = p.amortPct
			}
		}
	}
	netMin = math.Floor(netMin/10) * 10
	if netMin > -10 {
		netMin = -10
	}
	netMax = math.Ceil(netMax/10) * 10
	if netMax < 10 {
		netMax = 10
	}
	netRange := netMax - netMin

	topChartY := topMargin
	botChartY := topMargin + chartH + chartGap

	toSvgYTop := func(pct float64) int {
		return topChartY + int(float64(chartH)*(rawMax-pct)/rawRange)
	}
	toSvgYBot := func(pct float64) int {
		return botChartY + int(float64(chartH)*(netMax-pct)/netRange)
	}

	// Colors
	const (
		colorBg      = "#ffffff"
		colorGrid    = "#e2e8f0"
		colorSubtext = "#64748b"
		colorZero    = "#cbd5e1"
	)

	var b strings.Builder

	// SVG header
	b.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, svgWidth, svgHeight, svgWidth, svgHeight))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="%s"/>`, svgWidth, svgHeight, colorBg))
	b.WriteString("\n")

	// Styles
	b.WriteString(`<style>
    .title { font-family: Arial, Helvetica, sans-serif; font-size: 18px; font-weight: bold; fill: #1e293b; }
    .chart-title { font-family: Arial, Helvetica, sans-serif; font-size: 14px; font-weight: bold; fill: #1e293b; }
    .subtitle { font-family: Arial, Helvetica, sans-serif; font-size: 11px; fill: #64748b; }
    .axis-label { font-family: Arial, Helvetica, sans-serif; font-size: 11px; fill: #64748b; }
    .axis-title { font-family: Arial, Helvetica, sans-serif; font-size: 12px; fill: #1e293b; font-weight: bold; }
    .legend-text { font-family: Arial, Helvetica, sans-serif; font-size: 11px; fill: #1e293b; }
    .note { font-family: Arial, Helvetica, sans-serif; font-size: 10px; fill: #64748b; font-style: italic; }
    .data-label { font-family: Arial, Helvetica, sans-serif; font-size: 10px; font-weight: bold; }
    .summary-title { font-family: Arial, Helvetica, sans-serif; font-size: 12px; font-weight: bold; fill: #1e293b; }
    .summary-text { font-family: Arial, Helvetica, sans-serif; font-size: 11px; fill: #1e293b; }
  </style>`)
	b.WriteString("\n")

	// Main title
	b.WriteString(fmt.Sprintf(`<text x="%d" y="25" class="title">UCCP Compression: Token Savings Benchmark</text>`, leftMargin))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`<text x="%d" y="42" class="subtitle">Measured with tiktoken cl100k_base on realistic generated test data</text>`, leftMargin))
	b.WriteString("\n")

	// Legend (top right)
	legendX := svgWidth - rightMargin - 140
	for i, cat := range categories {
		ly := 18 + i*18
		color := categoryColors[cat]
		label := categoryLabels[cat]
		b.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="3"/>`,
			legendX, ly, legendX+20, ly, color))
		b.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="3" fill="%s"/>`,
			legendX+10, ly, color))
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="legend-text" dominant-baseline="central">%s</text>`,
			legendX+26, ly, label))
		b.WriteString("\n")
	}

	// ===== TOP PANEL: Raw Token Compression =====
	b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="chart-title">Raw Token Compression %%</text>`,
		leftMargin, topChartY-8))
	b.WriteString("\n")

	drawChart(&b, topChartY, chartH, chartW, leftMargin, rawMin, rawMax, rawRange,
		colorGrid, colorSubtext, colorZero, xTicks, toSvgX, toSvgYTop,
		categories, categoryColors, data, "rawPct", "Number of pages / files", "Token savings (%)")

	// ===== BOTTOM PANEL: Net Token Savings (Amortized) =====
	b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="chart-title">Net Token Savings (system prompt amortized over %d messages)</text>`,
		leftMargin, botChartY-8, DefaultAmortizationDepth))
	b.WriteString("\n")

	drawChart(&b, botChartY, chartH, chartW, leftMargin, netMin, netMax, netRange,
		colorGrid, colorSubtext, colorZero, xTicks, toSvgX, toSvgYBot,
		categories, categoryColors, data, "amortPct", "Number of pages / files", "Net token savings (%)")

	// ===== SUMMARY BOX =====
	boxY := botChartY + chartH + 40
	boxX := leftMargin
	boxW := chartW
	boxH := 75

	b.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="6" fill="#f8fafc" stroke="%s" stroke-width="1"/>`,
		boxX, boxY, boxW, boxH, colorGrid))
	b.WriteString("\n")

	summaryY := boxY + 18
	b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="summary-title">At 20 pages/files:</text>`, boxX+12, summaryY))
	b.WriteString("\n")

	colW := boxW / 3
	colIdx := 0
	for _, cat := range categories {
		pts := data[cat]
		for _, p := range pts {
			if p.pages == 20 {
				cx := boxX + 12 + colIdx*colW
				b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="summary-text" fill="%s">%s: %.0f%% raw / %+.0f%% net (%s tok saved)</text>`,
					cx, summaryY+18, categoryColors[cat], categoryLabels[cat],
					p.rawPct, p.amortPct, formatInt(p.savedTokNet)))
				b.WriteString("\n")
				colIdx++
			}
		}
	}

	b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="note">Token counts: tiktoken cl100k_base · "Raw" = compression only · "Net" = after system prompt overhead amortized over %d messages per conversation</text>`,
		boxX+12, summaryY+40, DefaultAmortizationDepth))
	b.WriteString("\n")

	b.WriteString("</svg>\n")

	return os.WriteFile(outputPath, []byte(b.String()), 0644)
}

// drawChart renders a single chart panel (grid, axes, lines, labels).
func drawChart(b *strings.Builder, chartY, chartH, chartW, leftMargin int,
	yMin, yMax, yRange float64,
	colorGrid, colorSubtext, colorZero string,
	xTicks []int,
	toSvgX func(int) int,
	toSvgY func(float64) int,
	categories []string,
	categoryColors map[string]string,
	data map[string][]graphPoint,
	metric string, // "rawPct" or "amortPct"
	xTitle, yTitle string,
) {
	// Y-axis grid lines and labels
	yStep := 10.0
	if yRange > 80 {
		yStep = 20
	}
	for y := yMin; y <= yMax; y += yStep {
		sy := toSvgY(y)
		strokeColor := colorGrid
		strokeWidth := "1"
		if y == 0 {
			strokeColor = colorZero
			strokeWidth = "2"
		}
		b.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="%s"/>`,
			leftMargin, sy, leftMargin+chartW, sy, strokeColor, strokeWidth))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="axis-label" text-anchor="end" dominant-baseline="central">%.0f%%</text>`,
			leftMargin-8, sy, y))
		b.WriteString("\n")
	}

	// X-axis ticks
	for _, x := range xTicks {
		sx := toSvgX(x)
		b.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="1"/>`,
			sx, chartY, sx, chartY+chartH, colorGrid))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="axis-label" text-anchor="middle">%d</text>`,
			sx, chartY+chartH+18, x))
		b.WriteString("\n")
	}

	// Axis titles
	b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="axis-title" text-anchor="middle">%s</text>`,
		leftMargin+chartW/2, chartY+chartH+38, xTitle))
	b.WriteString("\n")
	midY := chartY + chartH/2
	b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="axis-title" text-anchor="middle" transform="rotate(-90 %d %d)">%s</text>`,
		20, midY, 20, midY, yTitle))
	b.WriteString("\n")

	// Chart border
	b.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="2"/>`,
		leftMargin, chartY, leftMargin, chartY+chartH, colorSubtext))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="2"/>`,
		leftMargin, chartY+chartH, leftMargin+chartW, chartY+chartH, colorSubtext))
	b.WriteString("\n")

	// Plot lines and data points
	for _, cat := range categories {
		pts, ok := data[cat]
		if !ok || len(pts) == 0 {
			continue
		}
		color := categoryColors[cat]

		var polyPoints []string
		for _, p := range pts {
			val := p.rawPct
			if metric == "amortPct" {
				val = p.amortPct
			}
			sx := toSvgX(p.pages)
			sy := toSvgY(val)
			polyPoints = append(polyPoints, fmt.Sprintf("%d,%d", sx, sy))
		}

		b.WriteString(fmt.Sprintf(`<polyline points="%s" fill="none" stroke="%s" stroke-width="2.5" stroke-linejoin="round"/>`,
			strings.Join(polyPoints, " "), color))
		b.WriteString("\n")

		for _, p := range pts {
			val := p.rawPct
			if metric == "amortPct" {
				val = p.amortPct
			}
			sx := toSvgX(p.pages)
			sy := toSvgY(val)

			b.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="4" fill="%s" stroke="white" stroke-width="1.5"/>`,
				sx, sy, color))
			b.WriteString("\n")

			// Label first and last points
			if p.pages == 20 || p.pages == 1 {
				labelY := sy - 10
				anchor := "middle"
				if p.pages == 20 {
					anchor = "end"
				}
				if p.pages == 1 {
					anchor = "start"
				}
				sign := ""
				if val >= 0 {
					sign = "+"
				}
				b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="data-label" fill="%s" text-anchor="%s">%s%.1f%%</text>`,
					sx, labelY, color, anchor, sign, val))
				b.WriteString("\n")
			}
		}
	}
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
