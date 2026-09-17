// uccp is the command-line interface for the UCCP compression library.
//
// It reads content from a URL, a file, or stdin, runs it through the
// selected domain compressor, and writes the compressed (LLM-readable)
// output to stdout.
//
// Examples:
//
//	# Compress a webpage
//	uccp --url https://example.com/
//
//	# Compress a local HTML file
//	uccp --file page.html
//
//	# Pipe from stdin (auto-detects nothing; pass --domain explicitly if not HTML)
//	curl -s https://example.com/ | uccp
//
//	# Show compression stats on stderr
//	uccp --url https://example.com/ --stats > out.uccp
//
//	# Use a different domain
//	cat data.json | uccp --domain json
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/aguzmans/uccp/core"
	"github.com/aguzmans/uccp/domains"
)

// version is set at build time via -ldflags "-X main.version=..." by GoReleaser.
var version = "dev"

func main() {
	var (
		urlFlag     = flag.String("url", "", "fetch and compress the given URL")
		fileFlag    = flag.String("file", "", "read input from file path (use - for stdin)")
		domainFlag  = flag.String("domain", "html", "compression domain: html, code, json, financial")
		outputFlag  = flag.String("output", "", "write compressed output to file (default: stdout)")
		statsFlag   = flag.Bool("stats", false, "print compression stats to stderr")
		versionFlag = flag.Bool("version", false, "print version and exit")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "uccp %s — Ultra-Compact Content Protocol CLI\n\n", version)
		fmt.Fprintf(os.Stderr, "Compress HTML, JSON, or code into an LLM-readable format\n")
		fmt.Fprintf(os.Stderr, "that Claude and GPT read natively (no decompression step).\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  uccp --url URL       [--domain html|code|json|financial] [--stats]\n")
		fmt.Fprintf(os.Stderr, "  uccp --file PATH     [--domain ...] [--stats]\n")
		fmt.Fprintf(os.Stderr, "  cat file | uccp      [--domain ...] [--stats]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nMore info: https://github.com/aguzmans/uccp\n")
	}
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		return
	}

	input, source, err := readInput(*urlFlag, *fileFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "uccp: %v\n", err)
		os.Exit(1)
	}

	compressor, err := selectCompressor(*domainFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "uccp: %v\n", err)
		os.Exit(2)
	}

	compressed, err := compressor.Compress(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "uccp: compression failed: %v\n", err)
		os.Exit(3)
	}

	if err := writeOutput(*outputFlag, compressed); err != nil {
		fmt.Fprintf(os.Stderr, "uccp: %v\n", err)
		os.Exit(4)
	}

	if *statsFlag {
		printStats(os.Stderr, source, *domainFlag, input, compressed)
	}
}

func readInput(urlFlag, fileFlag string) (string, string, error) {
	switch {
	case urlFlag != "" && fileFlag != "":
		return "", "", fmt.Errorf("--url and --file are mutually exclusive")
	case urlFlag != "":
		body, err := fetchURL(urlFlag)
		if err != nil {
			return "", "", fmt.Errorf("fetch %s: %w", urlFlag, err)
		}
		return body, urlFlag, nil
	case fileFlag == "-" || fileFlag == "":
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", "", fmt.Errorf("read stdin: %w", err)
		}
		if len(data) == 0 {
			return "", "", fmt.Errorf("no input provided (pass --url, --file, or pipe via stdin; see --help)")
		}
		return string(data), "stdin", nil
	default:
		data, err := os.ReadFile(fileFlag)
		if err != nil {
			return "", "", fmt.Errorf("read %s: %w", fileFlag, err)
		}
		return string(data), fileFlag, nil
	}
}

func fetchURL(url string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "uccp-cli/"+version+" (+https://github.com/aguzmans/uccp)")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func selectCompressor(domain string) (core.Compressor, error) {
	switch domain {
	case "html":
		return domains.NewHTMLCompressor(), nil
	case "code":
		return domains.NewCodeCompressor(), nil
	case "json":
		return domains.NewJSONCompressor(), nil
	case "financial":
		return domains.NewFinancialCompressor(), nil
	default:
		return nil, fmt.Errorf("unknown domain %q (want: html, code, json, financial)", domain)
	}
}

func writeOutput(path, data string) error {
	if path == "" {
		_, err := io.WriteString(os.Stdout, data)
		return err
	}
	return os.WriteFile(path, []byte(data), 0644)
}

func printStats(w io.Writer, source, domain, original, compressed string) {
	ratio := core.CalculateCompressionRatio(original, compressed)
	tokens := core.EstimateTokenSavings(original, compressed)
	fmt.Fprintf(w, "\n--- uccp stats ---\n")
	fmt.Fprintf(w, "source:            %s\n", source)
	fmt.Fprintf(w, "domain:            %s\n", domain)
	fmt.Fprintf(w, "original bytes:    %d\n", len(original))
	fmt.Fprintf(w, "compressed bytes:  %d\n", len(compressed))
	fmt.Fprintf(w, "byte reduction:    %.1f%%\n", ratio*100)
	fmt.Fprintf(w, "tokens saved (~):  %d\n", tokens)
}
