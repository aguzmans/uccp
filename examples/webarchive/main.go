// Fetches a snapshot of sesamedisk.com from the Wayback Machine and runs it
// through the UCCP HTML compressor, printing before/after sizes and the
// achieved compression ratio.
//
// Run:
//
//	go run ./examples/webarchive
//
// Override the URL:
//
//	go run ./examples/webarchive https://web.archive.org/web/2026*/https://example.com/
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aguzmans/uccp/core"
	"github.com/aguzmans/uccp/domains"
)

const defaultURL = "https://web.archive.org/web/20260916041506/https://sesamedisk.com/"

func main() {
	url := defaultURL
	if len(os.Args) > 1 {
		url = os.Args[1]
	}

	fmt.Printf("Fetching: %s\n", url)
	html, err := fetch(url)
	if err != nil {
		log.Fatalf("fetch failed: %v", err)
	}

	compressor := domains.NewHTMLCompressor()
	compressed, err := compressor.Compress(html)
	if err != nil {
		log.Fatalf("compress failed: %v", err)
	}

	originalSize := len(html)
	compressedSize := len(compressed)
	ratio := core.CalculateCompressionRatio(html, compressed)
	tokensSaved := core.EstimateTokenSavings(html, compressed)

	fmt.Println()
	fmt.Println("=== UCCP HTML Compression: sesamedisk.com (Wayback snapshot) ===")
	fmt.Printf("Original size:     %d bytes\n", originalSize)
	fmt.Printf("Compressed size:   %d bytes\n", compressedSize)
	fmt.Printf("Byte reduction:    %.1f%%\n", ratio*100)
	fmt.Printf("Estimated tokens saved: ~%d\n", tokensSaved)

	fmt.Println()
	fmt.Println("--- Compressed preview (first 800 chars) ---")
	preview := compressed
	if len(preview) > 800 {
		preview = preview[:800] + "..."
	}
	fmt.Println(preview)
}

func fetch(url string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "uccp-example/1.0 (+https://github.com/aguzmans/uccp)")

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
