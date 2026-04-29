package core

// Domain represents the content type being compressed
type Domain int

const (
	// DomainAuto automatically detects the best domain
	DomainAuto Domain = iota
	// DomainCode for code, architecture, jobs, technical content
	DomainCode
	// DomainHTML for HTML, web content
	DomainHTML
	// DomainJSON for JSON data
	DomainJSON
	// DomainText for plain text, markdown
	DomainText
	// DomainFinancial for financial content (market summaries, predictions)
	DomainFinancial
)

// Compressor interface that all domain-specific compressors implement
type Compressor interface {
	// Compress converts content to UCCP format
	Compress(content string) (string, error)

	// Decompress converts UCCP format back to readable content
	Decompress(compressed string) (string, error)

	// SystemPrompt returns the LLM prompt explaining this compression format
	SystemPrompt() string

	// EstimateTokens estimates token count for given content
	EstimateTokens(content string) int
}

// AdaptivePrompter is an optional interface for compressors that can generate
// prompts based on which abbreviations were actually used.
type AdaptivePrompter interface {
	// AdaptiveSystemPrompt returns a system prompt containing only the
	// abbreviations and symbols that were used in the last Compress() call.
	AdaptiveSystemPrompt() string
}

// CompressionAdvisor is an optional interface for compressors that can quickly
// pre-check whether compressing a piece of content is likely to be worthwhile,
// without actually running the full compression. This is the low-cost screen
// callers can use to skip UCCP for content that is too small or too random
// to benefit. Compressors that implement this should aim for O(n) or better
// (linear scan), avoiding full parse/transform.
//
// Added in v0.0.8 (2026-04-29) for callers that need to gate compression on
// many candidate strings cheaply (e.g., agent-loop runtimes deciding whether
// to UCCP each tool result before sending it back to the model).
type CompressionAdvisor interface {
	// IsWorthCompressing returns whether the content is large enough and
	// structured enough that UCCP compression would meaningfully reduce
	// it. The estimate is approximate bytes saved (post-compression size
	// vs original); negative means compression would grow the content.
	//
	// Implementations should be cheap (linear scan, no full compression).
	// Used as a fast pre-check: callers can skip ShouldCompress() entirely
	// when this returns false.
	IsWorthCompressing(content string) (worth bool, estBytesSaved int)
}

// CompressionResult contains compression outcome and metadata
type CompressionResult struct {
	// Original content
	Original string

	// Compressed content (may be same as Original if compression didn't help)
	Compressed string

	// WasCompressed indicates if compression was applied
	WasCompressed bool

	// Ratio is the compression ratio achieved (0.0 to 1.0)
	// 0.7 means 70% compression (30% of original size)
	Ratio float64

	// OriginalSize in bytes
	OriginalSize int

	// CompressedSize in bytes
	CompressedSize int

	// EstimatedTokenSavings is approximate tokens saved
	EstimatedTokenSavings int

	// NetTokenSavings accounts for system prompt overhead (may be negative)
	NetTokenSavings int

	// Domain used for compression
	Domain Domain
}

// CompressionStats tracks aggregate compression statistics
type CompressionStats struct {
	// TotalCompressions attempted
	TotalCompressions int

	// SuccessfulCompressions where ratio met threshold
	SuccessfulCompressions int

	// SkippedCompressions where content was too small or ratio too low
	SkippedCompressions int

	// TotalBytesSaved across all compressions
	TotalBytesSaved int64

	// TotalTokensSaved across all compressions
	TotalTokensSaved int64

	// AverageRatio across successful compressions
	AverageRatio float64

	// BestRatio achieved
	BestRatio float64

	// WorstRatio among successful compressions
	WorstRatio float64
}
