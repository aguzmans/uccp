package benchmark

import (
	"fmt"
	"math"
	"os"
	"strings"
)

// GenerateGraph produces an SVG line chart showing how compression scales
// with content size. X axis = number of pages, Y axis = net token savings %.
// Each domain (HTML, JSON, Code) gets its own line.
func GenerateGraph(results []BenchmarkResult, outputPath string) error {
	if len(results) == 0 {
		return fmt.Errorf("no benchmark results to graph")
	}

	// Group results by category
	type point struct {
		pages    int
		netPct   float64
		tokPct   float64
		origTok  int
		savedTok int
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

	data := make(map[string][]point)
	for _, r := range results {
		data[r.Category] = append(data[r.Category], point{
			pages:    r.Pages,
			netPct:   r.NetTokenRatio * 100,
			tokPct:   r.TokenRatio * 100,
			origTok:  r.OriginalTokens,
			savedTok: r.NetTokenSavings,
		})
	}

	// Layout
	const (
		svgWidth    = 900
		svgHeight   = 540
		leftMargin  = 80
		rightMargin = 30
		topMargin   = 80
		botMargin   = 120
	)
	chartW := svgWidth - leftMargin - rightMargin
	chartH := svgHeight - topMargin - botMargin

	// Y-axis range: find min and max net savings %
	yMin := math.MaxFloat64
	yMax := -math.MaxFloat64
	for _, pts := range data {
		for _, p := range pts {
			if p.netPct < yMin {
				yMin = p.netPct
			}
			if p.netPct > yMax {
				yMax = p.netPct
			}
		}
	}

	// Round axis bounds for nice ticks
	yMin = math.Floor(yMin/10) * 10
	if yMin > -10 {
		yMin = -10
	}
	yMax = math.Ceil(yMax/10) * 10
	if yMax < 10 {
		yMax = 10
	}
	yRange := yMax - yMin

	// X-axis: page counts
	xTicks := []int{1, 5, 10, 15, 20}
	xMin := 0.0
	xMax := 22.0
	xRange := xMax - xMin

	toSvgX := func(pages int) int {
		return leftMargin + int(float64(chartW)*(float64(pages)-xMin)/xRange)
	}
	toSvgY := func(pct float64) int {
		return topMargin + int(float64(chartH)*(yMax-pct)/yRange)
	}

	// Colors
	const (
		colorBg      = "#ffffff"
		colorText    = "#1e293b"
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
	b.WriteString(`<style>`)
	b.WriteString(fmt.Sprintf(`
    .title { font-family: Arial, Helvetica, sans-serif; font-size: 18px; font-weight: bold; fill: %s; }
    .subtitle { font-family: Arial, Helvetica, sans-serif; font-size: 12px; fill: %s; }
    .axis-label { font-family: Arial, Helvetica, sans-serif; font-size: 11px; fill: %s; }
    .axis-title { font-family: Arial, Helvetica, sans-serif; font-size: 12px; fill: %s; font-weight: bold; }
    .legend-text { font-family: Arial, Helvetica, sans-serif; font-size: 11px; fill: %s; }
    .note { font-family: Arial, Helvetica, sans-serif; font-size: 10px; fill: %s; font-style: italic; }
    .data-label { font-family: Arial, Helvetica, sans-serif; font-size: 10px; font-weight: bold; }
  `, colorText, colorSubtext, colorSubtext, colorText, colorText, colorSubtext))
	b.WriteString(`</style>`)
	b.WriteString("\n")

	// Title
	b.WriteString(fmt.Sprintf(`<text x="%d" y="30" class="title">UCCP Compression: Token Savings at Scale</text>`, leftMargin))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`<text x="%d" y="48" class="subtitle">Net token savings (%%) vs content size — measured with tiktoken cl100k_base (includes system prompt overhead)</text>`, leftMargin))
	b.WriteString("\n")

	// Legend
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

	// X-axis ticks and labels
	for _, x := range xTicks {
		sx := toSvgX(x)
		b.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="1"/>`,
			sx, topMargin, sx, topMargin+chartH, colorGrid))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="axis-label" text-anchor="middle">%d</text>`,
			sx, topMargin+chartH+18, x))
		b.WriteString("\n")
	}

	// Axis titles
	b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="axis-title" text-anchor="middle">Number of pages / files</text>`,
		leftMargin+chartW/2, topMargin+chartH+38))
	b.WriteString("\n")
	// Rotated Y axis title
	b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="axis-title" text-anchor="middle" transform="rotate(-90 %d %d)">Net token savings (%%)</text>`,
		20, topMargin+chartH/2, 20, topMargin+chartH/2))
	b.WriteString("\n")

	// Chart border (axes)
	b.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="2"/>`,
		leftMargin, topMargin, leftMargin, topMargin+chartH, colorSubtext))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="2"/>`,
		leftMargin, topMargin+chartH, leftMargin+chartW, topMargin+chartH, colorSubtext))
	b.WriteString("\n")

	// Plot lines and data points for each category
	for _, cat := range categories {
		pts, ok := data[cat]
		if !ok || len(pts) == 0 {
			continue
		}
		color := categoryColors[cat]

		// Build polyline points
		var polyPoints []string
		for _, p := range pts {
			sx := toSvgX(p.pages)
			sy := toSvgY(p.netPct)
			polyPoints = append(polyPoints, fmt.Sprintf("%d,%d", sx, sy))
		}

		// Line
		b.WriteString(fmt.Sprintf(`<polyline points="%s" fill="none" stroke="%s" stroke-width="2.5" stroke-linejoin="round"/>`,
			strings.Join(polyPoints, " "), color))
		b.WriteString("\n")

		// Data points and labels
		for _, p := range pts {
			sx := toSvgX(p.pages)
			sy := toSvgY(p.netPct)

			// Circle marker
			b.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="4" fill="%s" stroke="white" stroke-width="1.5"/>`,
				sx, sy, color))
			b.WriteString("\n")

			// Label on last point and first point
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
				if p.netPct >= 0 {
					sign = "+"
				}
				b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="data-label" fill="%s" text-anchor="%s">%s%.1f%%</text>`,
					sx, labelY, color, anchor, sign, p.netPct))
				b.WriteString("\n")
			}
		}
	}

	// Summary box at bottom
	boxY := svgHeight - botMargin + 55
	boxX := leftMargin
	boxW := chartW
	boxH := 70

	b.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="6" fill="#f8fafc" stroke="%s" stroke-width="1"/>`,
		boxX, boxY, boxW, boxH, colorGrid))
	b.WriteString("\n")

	// Summary: at 20 pages — three columns
	summaryY := boxY + 18
	b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="axis-title">At 20 pages/files:</text>`, boxX+12, summaryY))
	b.WriteString("\n")

	colW := boxW / 3
	colIdx := 0
	for _, cat := range categories {
		pts := data[cat]
		for _, p := range pts {
			if p.pages == 20 {
				sign := ""
				if p.netPct >= 0 {
					sign = "+"
				}
				cx := boxX + 12 + colIdx*colW
				b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="axis-label" fill="%s">%s: %s%.1f%% net (%s tokens)</text>`,
					cx, summaryY+18, categoryColors[cat], categoryLabels[cat], sign, p.netPct, formatInt(p.savedTok)))
				b.WriteString("\n")
				colIdx++
			}
		}
	}

	b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="note">Token counts measured with tiktoken cl100k_base · net savings include one-time system prompt overhead per domain</text>`,
		boxX+12, summaryY+38))
	b.WriteString("\n")

	b.WriteString("</svg>\n")

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
