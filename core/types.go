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
