package benchmark

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// BenchmarkRun represents a single historical benchmark execution.
type BenchmarkRun struct {
	Timestamp string             `json:"timestamp"`
	GitCommit string             `json:"git_commit"`
	SVGFile   string             `json:"svg_file"`
	Summary   map[string]float64 `json:"summary"` // category -> raw compression % at 20 pages
}

// BenchmarkHistory tracks the last N benchmark runs.
type BenchmarkHistory struct {
	Runs []BenchmarkRun `json:"runs"`
}

const maxHistoryRuns = 5

// LoadHistory reads the history JSON file, returning empty history if missing.
func LoadHistory(path string) (*BenchmarkHistory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &BenchmarkHistory{}, nil
		}
		return nil, err
	}
	var h BenchmarkHistory
	if err := json.Unmarshal(data, &h); err != nil {
		return &BenchmarkHistory{}, nil
	}
	return &h, nil
}

// SaveHistory writes the history JSON file.
func SaveHistory(h *BenchmarkHistory, path string) error {
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// AddRun appends a run and trims to maxHistoryRuns, deleting SVGs from pruned runs.
func AddRun(h *BenchmarkHistory, run BenchmarkRun, historyDir string) {
	h.Runs = append(h.Runs, run)
	for len(h.Runs) > maxHistoryRuns {
		old := h.Runs[0]
		if old.SVGFile != "" {
			os.Remove(filepath.Join(historyDir, old.SVGFile))
		}
		h.Runs = h.Runs[1:]
	}
}

// ArchiveSVG copies the current SVG into the history directory with a timestamp.
func ArchiveSVG(currentSVG, historyDir string) (string, error) {
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return "", err
	}

	ts := time.Now().Format("2006-01-02T15-04-05")
	name := fmt.Sprintf("benchmark-%s.svg", ts)
	dst := filepath.Join(historyDir, name)

	src, err := os.Open(currentSVG)
	if err != nil {
		return "", err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		return "", err
	}
	return name, nil
}

// GitShortSHA returns the short git commit hash, or "unknown" if unavailable.
func GitShortSHA() string {
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

// BuildRunSummary extracts the raw compression % at 20 pages for each category.
func BuildRunSummary(results []BenchmarkResult) map[string]float64 {
	summary := make(map[string]float64)
	for _, r := range results {
		if r.Pages == 20 {
			summary[r.Category] = r.TokenRatio * 100
		}
	}
	return summary
}

// UpdateREADME replaces the ## Benchmarks section in the README with updated content.
func UpdateREADME(readmePath string, results []BenchmarkResult) error {
	data, err := os.ReadFile(readmePath)
	if err != nil {
		return err
	}
	content := string(data)

	// Find the ## Benchmarks section (everything up to the next ## heading)
	re := regexp.MustCompile(`(?s)(## Benchmarks\n.*?\n)(## \S)`)
	loc := re.FindStringIndex(content)
	if loc == nil {
		return fmt.Errorf("could not find ## Benchmarks section in README")
	}
	submatch := re.FindStringSubmatch(content)
	// loc[0]..loc[1] spans the full match; submatch[2] is the next heading start
	// We want to replace everything except the next heading
	match := []int{loc[0], loc[1] - len(submatch[2])}

	// Get summary stats for the description
	var htmlRaw, jsonRaw, codeRaw float64
	var htmlNet, jsonNet, codeNet float64
	for _, r := range results {
		if r.Pages == 20 {
			switch r.Category {
			case "HTML Documentation":
				htmlRaw = r.TokenRatio * 100
				htmlNet = r.AmortizedNetRatio * 100
			case "JSON API Responses":
				jsonRaw = r.TokenRatio * 100
				jsonNet = r.AmortizedNetRatio * 100
			case "Source Code":
				codeRaw = r.TokenRatio * 100
				codeNet = r.AmortizedNetRatio * 100
			}
		}
	}

	ts := time.Now().Format("2006-01-02")

	newSection := fmt.Sprintf(`## Benchmarks

Token savings measured with **tiktoken cl100k_base** on realistic generated test data (HTML pages, JSON API responses, source code).

The chart shows two views:
- **Top panel**: Raw token compression %% (compression alone, no overhead)
- **Bottom panel**: Net token savings with system prompt amortized over 10 messages (realistic conversation usage)

![UCCP Compression Benchmarks](docs/benchmark-results.svg)

**Last benchmarked: %s** — At 20 pages: HTML %.0f%% raw / %.0f%% net, JSON %.0f%% raw / %.0f%% net, Code %.0f%% raw / %.0f%% net

Historical benchmark results are saved in [` + "`" + `docs/benchmark-history/` + "`" + `](docs/benchmark-history/).

**Regenerate benchmarks locally:**
` + "```" + `bash
# With Go installed:
go run ./benchmark/cmd/

# Or with Docker (no Go required):
docker run --rm -v "$(pwd):/src" -w /src golang:1.21-alpine \
  sh -c "apk add --no-cache git >/dev/null 2>&1 && go run ./benchmark/cmd/"

# In containerized/CI environments (where volume mounts can't reach source):
docker build -f Dockerfile.bench -t uccp-bench .
docker run --name uccp-bench-run uccp-bench
docker cp uccp-bench-run:/src/docs/benchmark-results.svg ./docs/benchmark-results.svg
docker cp uccp-bench-run:/src/README.md ./README.md
docker rm uccp-bench-run
` + "```" + `

`, ts, htmlRaw, htmlNet, jsonRaw, jsonNet, codeRaw, codeNet)

	newContent := content[:match[0]] + newSection + content[match[1]:]
	return os.WriteFile(readmePath, []byte(newContent), 0644)
}
