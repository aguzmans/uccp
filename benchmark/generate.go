package benchmark

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

// GenerateTestData creates realistic test content at various page counts.
// Each "page" is a self-contained chunk (~30-60KB) that simulates real content
// an LLM would receive for review.
func GenerateTestData(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create testdata dir: %w", err)
	}

	type genEntry struct {
		name string
		gen  func(n int) string
	}

	entries := []genEntry{
		{"html_pages", genHTMLPages},
		{"json_responses", genJSONResponses},
		{"code_files", genCodeFiles},
	}

	// Generate at each scale: 1, 5, 10, 15, 20 pages
	scales := []int{1, 5, 10, 15, 20}

	for _, e := range entries {
		for _, n := range scales {
			fname := fmt.Sprintf("%s_%02d.txt", e.name, n)
			path := filepath.Join(dir, fname)
			if _, err := os.Stat(path); err == nil {
				continue
			}
			content := e.gen(n)
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return fmt.Errorf("write %s: %w", fname, err)
			}
			fmt.Printf("  Generated %s (%s, %d pages)\n", fname, humanSize(len(content)), n)
		}
	}
	return nil
}

func humanSize(b int) string {
	if b >= 1024*1024 {
		return fmt.Sprintf("%.1fMB", float64(b)/(1024*1024))
	}
	return fmt.Sprintf("%.1fKB", float64(b)/1024)
}

// ScaleBenchmarks returns the benchmark configurations: domain × page count.
func ScaleBenchmarks() []ScaleTest {
	scales := []int{1, 5, 10, 15, 20}
	var tests []ScaleTest

	for _, n := range scales {
		tests = append(tests, ScaleTest{
			Name:     fmt.Sprintf("HTML %d pages", n),
			Path:     fmt.Sprintf("html_pages_%02d.txt", n),
			Domain:   "html",
			Pages:    n,
			Category: "HTML Documentation",
		})
	}
	for _, n := range scales {
		tests = append(tests, ScaleTest{
			Name:     fmt.Sprintf("JSON %d responses", n),
			Path:     fmt.Sprintf("json_responses_%02d.txt", n),
			Domain:   "code",
			Pages:    n,
			Category: "JSON API Responses",
		})
	}
	for _, n := range scales {
		tests = append(tests, ScaleTest{
			Name:     fmt.Sprintf("Code %d files", n),
			Path:     fmt.Sprintf("code_files_%02d.txt", n),
			Domain:   "code",
			Pages:    n,
			Category: "Source Code",
		})
	}
	return tests
}

// ScaleTest describes a benchmark at a given scale.
type ScaleTest struct {
	Name     string
	Path     string
	Domain   string
	Pages    int
	Category string
}

// ---------------------------------------------------------------------------
// HTML page generator — realistic documentation pages ~40-50KB each
// ---------------------------------------------------------------------------

func genHTMLPages(n int) string {
	var pages []string
	for i := 0; i < n; i++ {
		pages = append(pages, genSingleHTMLPage(i))
	}
	return strings.Join(pages, "\n\n")
}

func genSingleHTMLPage(idx int) string {
	topics := []struct{ title, pkg string }{
		{"HTTP Server Configuration", "net/http"},
		{"Database Connection Pooling", "database/sql"},
		{"JSON Encoding and Decoding", "encoding/json"},
		{"Concurrent Task Execution", "sync"},
		{"File System Operations", "os"},
		{"Template Rendering Engine", "html/template"},
		{"Cryptographic Operations", "crypto"},
		{"Network Socket Programming", "net"},
		{"Regular Expression Matching", "regexp"},
		{"Command Line Argument Parsing", "flag"},
		{"Logging and Diagnostics", "log/slog"},
		{"Time and Duration Handling", "time"},
		{"Context Propagation Patterns", "context"},
		{"IO Reader Writer Interfaces", "io"},
		{"String Manipulation Utilities", "strings"},
		{"Error Handling Strategies", "errors"},
		{"Testing Framework Features", "testing"},
		{"Reflection and Type Inspection", "reflect"},
		{"Memory Allocation Profiling", "runtime"},
		{"Embedded File Systems", "embed"},
	}

	t := topics[idx%len(topics)]
	var b strings.Builder

	b.WriteString(fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>%s - Go Standard Library Documentation</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; max-width: 960px; margin: 0 auto; padding: 2rem; color: #333; }
    nav { background: #f5f5f5; padding: 1rem; margin-bottom: 2rem; border-radius: 4px; }
    nav a { margin-right: 1rem; color: #0366d6; text-decoration: none; }
    h1 { border-bottom: 2px solid #eee; padding-bottom: .5rem; }
    h2 { color: #24292e; margin-top: 2rem; }
    pre { background: #f6f8fa; padding: 1rem; border-radius: 4px; overflow-x: auto; font-size: 14px; }
    code { font-family: 'SF Mono', Consolas, monospace; }
    table { border-collapse: collapse; width: 100%%; margin: 1rem 0; }
    th, td { border: 1px solid #ddd; padding: 8px 12px; text-align: left; }
    th { background: #f6f8fa; font-weight: 600; }
    .sidebar { float: right; width: 250px; background: #f9f9f9; padding: 1rem; margin-left: 1rem; border: 1px solid #eee; border-radius: 4px; }
    .note { background: #fffbdd; border: 1px solid #e8d44d; padding: 1rem; border-radius: 4px; margin: 1rem 0; }
    footer { margin-top: 3rem; padding: 1rem 0; border-top: 1px solid #eee; color: #666; font-size: 0.9rem; }
    ul { line-height: 1.8; }
  </style>
</head>
<body>
  <nav>
    <a href="/">Home</a> <a href="/pkg">Packages</a> <a href="/doc">Documentation</a> <a href="/blog">Blog</a> <a href="/play">Playground</a>
  </nav>
  <div class="sidebar">
    <h3>Package Index</h3>
    <ul>
      <li><a href="#overview">Overview</a></li>
      <li><a href="#types">Types</a></li>
      <li><a href="#functions">Functions</a></li>
      <li><a href="#examples">Examples</a></li>
      <li><a href="#best-practices">Best Practices</a></li>
      <li><a href="#errors">Error Handling</a></li>
      <li><a href="#performance">Performance</a></li>
    </ul>
  </div>
  <h1>Package %s</h1>
  <p><code>import "%s"</code></p>
`, t.title, t.pkg, t.pkg))

	// Overview
	b.WriteString(fmt.Sprintf(`
  <h2 id="overview">Overview</h2>
  <p>Package %s provides comprehensive functionality for %s in Go applications. It is designed for production use with proper error handling, context support, and concurrent safety.</p>
  <p>This package is part of the Go standard library and is maintained by the Go team. It has been battle-tested in production systems handling millions of requests per day at companies like Google, Cloudflare, and Uber.</p>
  <p>The primary design goals are simplicity, correctness, and performance. All public APIs are safe for concurrent use unless explicitly documented otherwise.</p>
  <div class="note">
    <strong>Note:</strong> Starting with Go 1.22, this package includes enhanced support for generics and improved error wrapping. See the migration guide for details on upgrading from earlier versions.
  </div>
`, t.pkg, strings.ToLower(t.title)))

	// Generate 8 type definitions
	typeNames := []string{"Server", "Client", "Handler", "Config", "Pool", "Manager", "Worker", "Result"}
	b.WriteString(`  <h2 id="types">Types</h2>`)
	for _, tn := range typeNames {
		b.WriteString(fmt.Sprintf(`
  <h3>type %s</h3>
  <pre><code class="language-go">type %s struct {
    // Address specifies the TCP address for the %s to operate on.
    Address string

    // Timeout defines the maximum duration for operations.
    // A zero value means no timeout.
    Timeout time.Duration

    // MaxRetries specifies the maximum number of retry attempts.
    MaxRetries int

    // Logger is the structured logger for diagnostics.
    Logger *slog.Logger

    // ErrorHandler is called when an unrecoverable error occurs.
    ErrorHandler func(error)
}</code></pre>
  <p>%s represents the primary operational unit for %s. The zero value is not usable; use New%s to create a properly initialized instance with sensible defaults for production environments.</p>
  <p>A %s is safe for concurrent use by multiple goroutines once initialized. Do not copy a %s after first use, as this may lead to race conditions on internal state.</p>
`, tn, tn, tn, tn, strings.ToLower(t.title), tn, tn, tn))
	}

	// Generate 6 function definitions with examples
	funcNames := []string{"New", "Start", "Stop", "Configure", "Execute", "Validate"}
	b.WriteString(`  <h2 id="functions">Functions</h2>`)
	for _, fn := range funcNames {
		b.WriteString(fmt.Sprintf(`
  <h3>func %s</h3>
  <pre><code class="language-go">func %s(ctx context.Context, opts ...Option) (*Result, error)</code></pre>
  <p>%s initializes and returns a new operational instance configured with the given options. It validates all provided options and returns an error if any configuration is invalid.</p>
  <p>The context is used for cancellation and deadline propagation. If the context is cancelled before the operation completes, %s returns the context's error immediately.</p>
  <h4>Example</h4>
  <pre><code class="language-go">ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := %s.%s(ctx,
    WithTimeout(15*time.Second),
    WithMaxRetries(3),
    WithLogger(slog.Default()),
)
if err != nil {
    log.Fatalf("%%s failed: %%v", "%s", err)
}
defer result.Close()

log.Printf("%%s completed successfully: %%+v", "%s", result.Stats())</code></pre>
`, fn, fn, fn, fn, t.pkg, fn, fn, fn))
	}

	// Best practices section with substantial content
	b.WriteString(fmt.Sprintf(`
  <h2 id="best-practices">Best Practices</h2>
  <p>When using package %s in production, follow these guidelines to ensure reliability, performance, and maintainability:</p>
  <h3>1. Always Use Context</h3>
  <p>Pass context.Context through all function calls to enable cancellation, deadline propagation, and request-scoped values. Never use context.Background() in request handlers; use the request's context instead.</p>
  <pre><code class="language-go">func handleRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    result, err := service.Execute(ctx, params)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            http.Error(w, "request timeout", http.StatusGatewayTimeout)
            return
        }
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(result)
}</code></pre>

  <h3>2. Handle Errors Properly</h3>
  <p>Use error wrapping with %%w to preserve the error chain. Check for specific error types using errors.Is() and errors.As() rather than string comparison.</p>
  <pre><code class="language-go">result, err := service.Execute(ctx, params)
if err != nil {
    var validationErr *ValidationError
    if errors.As(err, &validationErr) {
        return fmt.Errorf("validation failed for field %%s: %%w", validationErr.Field, err)
    }
    if errors.Is(err, ErrNotFound) {
        return nil, ErrNotFound // propagate without wrapping
    }
    return nil, fmt.Errorf("execute service: %%w", err)
}</code></pre>

  <h3>3. Configure Timeouts</h3>
  <p>Always set appropriate timeouts to prevent resource exhaustion. Use context deadlines for cascading timeouts across service boundaries.</p>
  <table>
    <tr><th>Operation</th><th>Recommended Timeout</th><th>Notes</th></tr>
    <tr><td>HTTP request</td><td>15-30 seconds</td><td>Include DNS, TLS, and response body</td></tr>
    <tr><td>Database query</td><td>5-15 seconds</td><td>Depends on query complexity</td></tr>
    <tr><td>Cache lookup</td><td>1-3 seconds</td><td>Fast failure, fallback to DB</td></tr>
    <tr><td>External API</td><td>10-30 seconds</td><td>With circuit breaker</td></tr>
    <tr><td>File I/O</td><td>5-10 seconds</td><td>For network-mounted filesystems</td></tr>
    <tr><td>gRPC call</td><td>5-15 seconds</td><td>With retry and backoff</td></tr>
  </table>

  <h3>4. Use Connection Pooling</h3>
  <pre><code class="language-go">pool := &Pool{
    MaxOpen:     50,             // Maximum open connections
    MaxIdle:     10,             // Maximum idle connections
    MaxLifetime: 30 * time.Minute, // Connection max lifetime
    IdleTimeout: 5 * time.Minute,  // Idle connection timeout
}

// Monitor pool statistics
go func() {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        stats := pool.Stats()
        slog.Info("pool stats",
            "open", stats.OpenConnections,
            "idle", stats.Idle,
            "in_use", stats.InUse,
            "wait_count", stats.WaitCount,
            "wait_duration", stats.WaitDuration,
        )
    }
}()</code></pre>
`, t.pkg))

	// Performance section with benchmarks
	b.WriteString(`
  <h2 id="performance">Performance Considerations</h2>
  <p>Benchmark results on an Apple M2 Pro with Go 1.22:</p>
  <table>
    <tr><th>Operation</th><th>Time</th><th>Allocations</th><th>Memory</th></tr>
    <tr><td>Create instance</td><td>125 ns/op</td><td>2 allocs/op</td><td>256 B/op</td></tr>
    <tr><td>Execute (small)</td><td>890 ns/op</td><td>5 allocs/op</td><td>1,024 B/op</td></tr>
    <tr><td>Execute (medium)</td><td>3.2 µs/op</td><td>12 allocs/op</td><td>4,096 B/op</td></tr>
    <tr><td>Execute (large)</td><td>15.4 µs/op</td><td>28 allocs/op</td><td>16,384 B/op</td></tr>
    <tr><td>Concurrent (8 goroutines)</td><td>2.1 µs/op</td><td>8 allocs/op</td><td>2,048 B/op</td></tr>
    <tr><td>With caching</td><td>45 ns/op</td><td>0 allocs/op</td><td>0 B/op</td></tr>
  </table>
  <p>Key optimization tips:</p>
  <ul>
    <li>Reuse instances instead of creating new ones per request</li>
    <li>Enable connection pooling for network operations</li>
    <li>Use sync.Pool for frequently allocated temporary objects</li>
    <li>Prefer io.Reader/Writer interfaces over []byte for large data</li>
    <li>Profile with pprof before optimizing — measure, don't guess</li>
    <li>Use buffered channels for producer-consumer patterns</li>
  </ul>
`)

	// Error handling section
	b.WriteString(`
  <h2 id="errors">Error Types</h2>
  <table>
    <tr><th>Error</th><th>Description</th><th>Retry?</th></tr>
    <tr><td>ErrNotFound</td><td>The requested resource does not exist</td><td>No</td></tr>
    <tr><td>ErrTimeout</td><td>The operation exceeded the configured timeout</td><td>Yes, with backoff</td></tr>
    <tr><td>ErrPermission</td><td>Insufficient permissions for the operation</td><td>No</td></tr>
    <tr><td>ErrConflict</td><td>A conflicting operation is in progress</td><td>Yes, after delay</td></tr>
    <tr><td>ErrInvalid</td><td>The provided input failed validation</td><td>No, fix input</td></tr>
    <tr><td>ErrUnavailable</td><td>The service is temporarily unavailable</td><td>Yes, with circuit breaker</td></tr>
    <tr><td>ErrInternal</td><td>An unexpected internal error occurred</td><td>Yes, with logging</td></tr>
  </table>
`)

	b.WriteString(`
  <footer>
    <p>Generated by godoc. Copyright 2024 The Go Authors. All rights reserved.</p>
    <p>Use of this source code is governed by a BSD-style license found in the LICENSE file.</p>
  </footer>
</body>
</html>`)

	return b.String()
}

// ---------------------------------------------------------------------------
// JSON response generator — realistic API responses ~30-40KB each
// ---------------------------------------------------------------------------

func genJSONResponses(n int) string {
	var pages []string
	for i := 0; i < n; i++ {
		pages = append(pages, genSingleJSONResponse(i))
	}
	return strings.Join(pages, "\n\n")
}

func genSingleJSONResponse(idx int) string {
	var b strings.Builder
	// Each "response" is a paginated list of 100 records
	b.WriteString("[\n")

	departments := []string{"Engineering", "Product", "Design", "Marketing", "Sales", "Operations", "Finance", "Legal", "HR", "Support"}
	skills := []string{"Go", "Python", "TypeScript", "React", "PostgreSQL", "Redis", "Docker", "Kubernetes", "AWS", "gRPC", "GraphQL", "Terraform"}

	offset := idx * 100
	for i := 0; i < 100; i++ {
		if i > 0 {
			b.WriteString(",\n")
		}
		uid := offset + i + 1
		dept := departments[uid%len(departments)]
		s1 := skills[(uid*3)%len(skills)]
		s2 := skills[(uid*3+1)%len(skills)]
		s3 := skills[(uid*3+2)%len(skills)]

		b.WriteString(fmt.Sprintf(`  {
    "id": %d,
    "employee_id": "EMP-%06d",
    "first_name": "Employee",
    "last_name": "Number%d",
    "email": "employee%d@company.com",
    "department": "%s",
    "title": "Senior %s Engineer",
    "manager_id": %d,
    "active": %t,
    "hire_date": "2023-%02d-%02dT09:00:00Z",
    "last_review_date": "2025-%02d-%02dT14:00:00Z",
    "salary": {
      "amount": %d,
      "currency": "USD",
      "pay_period": "annual"
    },
    "address": {
      "street": "%d Technology Drive",
      "suite": "Suite %d",
      "city": "%s",
      "state": "%s",
      "zip": "%05d",
      "country": "US"
    },
    "skills": ["%s", "%s", "%s"],
    "certifications": [
      {"name": "AWS Solutions Architect", "issued": "2024-01-15", "expires": "2027-01-15"},
      {"name": "Kubernetes Administrator", "issued": "2024-06-01", "expires": "2027-06-01"}
    ],
    "performance": {
      "current_rating": %d,
      "goals_completed": %d,
      "goals_total": %d,
      "peer_feedback_score": %.1f
    },
    "metadata": {
      "created_at": "2023-%02d-%02dT10:00:00Z",
      "updated_at": "2026-03-01T08:30:00Z",
      "version": %d,
      "source": "hr-system"
    }
  }`, uid, uid, uid, uid,
			dept, dept, (uid/10)+1,
			uid%5 != 0,
			(uid%12)+1, (uid%28)+1,
			(uid%12)+1, (uid%28)+1,
			85000+uid*500,
			uid*100+1, uid%500+100,
			[]string{"San Francisco", "New York", "Austin", "Seattle", "Chicago", "Denver", "Boston", "Portland"}[uid%8],
			[]string{"CA", "NY", "TX", "WA", "IL", "CO", "MA", "OR"}[uid%8],
			10000+uid,
			s1, s2, s3,
			3+(uid%3), 5+(uid%6), 8+(uid%4),
			3.5+float64(uid%15)*0.1,
			(uid%12)+1, (uid%28)+1, uid%10+1,
		))
	}
	b.WriteString("\n]")
	return b.String()
}

// ---------------------------------------------------------------------------
// Code file generator — realistic source files ~15-25KB each
// ---------------------------------------------------------------------------

func genCodeFiles(n int) string {
	var files []string
	for i := 0; i < n; i++ {
		files = append(files, genSingleCodeFile(i))
	}
	return strings.Join(files, "\n\n// " + strings.Repeat("=", 60) + "\n\n")
}

func genSingleCodeFile(idx int) string {
	// Alternate between Go, TypeScript, and Python
	switch idx % 3 {
	case 0:
		return genGoFile(idx)
	case 1:
		return genTSXFile(idx)
	default:
		return genPyFile(idx)
	}
}

func genGoFile(idx int) string {
	services := []string{"user", "product", "order", "payment", "notification", "analytics", "search", "auth"}
	svc := services[idx%len(services)]
	Svc := strings.Title(svc)

	return fmt.Sprintf(`package %s

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// %s represents the domain model for %s management.
type %s struct {
	ID          int64     `+"`"+`json:"id"`+"`"+`
	Name        string    `+"`"+`json:"name"`+"`"+`
	Description string    `+"`"+`json:"description"`+"`"+`
	Status      string    `+"`"+`json:"status"`+"`"+`
	Priority    int       `+"`"+`json:"priority"`+"`"+`
	Tags        []string  `+"`"+`json:"tags"`+"`"+`
	Metadata    Metadata  `+"`"+`json:"metadata"`+"`"+`
	CreatedAt   time.Time `+"`"+`json:"created_at"`+"`"+`
	UpdatedAt   time.Time `+"`"+`json:"updated_at"`+"`"+`
}

type Metadata struct {
	CreatedBy string `+"`"+`json:"created_by"`+"`"+`
	UpdatedBy string `+"`"+`json:"updated_by"`+"`"+`
	Version   int    `+"`"+`json:"version"`+"`"+`
	Source    string `+"`"+`json:"source"`+"`"+`
}

type Create%sRequest struct {
	Name        string   `+"`"+`json:"name" validate:"required,min=1,max=200"`+"`"+`
	Description string   `+"`"+`json:"description" validate:"max=2000"`+"`"+`
	Priority    int      `+"`"+`json:"priority" validate:"min=1,max=5"`+"`"+`
	Tags        []string `+"`"+`json:"tags" validate:"max=10"`+"`"+`
}

type Update%sRequest struct {
	Name        *string  `+"`"+`json:"name,omitempty" validate:"omitempty,min=1,max=200"`+"`"+`
	Description *string  `+"`"+`json:"description,omitempty" validate:"omitempty,max=2000"`+"`"+`
	Status      *string  `+"`"+`json:"status,omitempty" validate:"omitempty,oneof=active inactive archived"`+"`"+`
	Priority    *int     `+"`"+`json:"priority,omitempty" validate:"omitempty,min=1,max=5"`+"`"+`
	Tags        []string `+"`"+`json:"tags,omitempty" validate:"omitempty,max=10"`+"`"+`
}

type ListOptions struct {
	Page     int    `+"`"+`json:"page"`+"`"+`
	PageSize int    `+"`"+`json:"page_size"`+"`"+`
	SortBy   string `+"`"+`json:"sort_by"`+"`"+`
	Order    string `+"`"+`json:"order"`+"`"+`
	Search   string `+"`"+`json:"search"`+"`"+`
	Status   string `+"`"+`json:"status"`+"`"+`
}

// %sRepository defines the data access interface for %s operations.
type %sRepository interface {
	FindByID(ctx context.Context, id int64) (*%s, error)
	List(ctx context.Context, opts ListOptions) ([]*%s, int, error)
	Create(ctx context.Context, req Create%sRequest) (*%s, error)
	Update(ctx context.Context, id int64, req Update%sRequest) (*%s, error)
	Delete(ctx context.Context, id int64) error
}

type postgres%sRepo struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewPostgres%sRepository(db *sql.DB, logger *slog.Logger) %sRepository {
	return &postgres%sRepo{db: db, logger: logger}
}

func (r *postgres%sRepo) FindByID(ctx context.Context, id int64) (*%s, error) {
	var item %s
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, description, status, priority, created_at, updated_at FROM %ss WHERE id = $1",
		id,
	).Scan(&item.ID, &item.Name, &item.Description, &item.Status, &item.Priority, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%s not found: %%d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("query %s: %%w", err)
	}
	return &item, nil
}

func (r *postgres%sRepo) List(ctx context.Context, opts ListOptions) ([]*%s, int, error) {
	if opts.PageSize <= 0 || opts.PageSize > 100 {
		opts.PageSize = 20
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}
	offset := (opts.Page - 1) * opts.PageSize

	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM %ss WHERE ($1 = '' OR status = $1)", opts.Status).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count %ss: %%w", err)
	}

	rows, err := r.db.QueryContext(ctx,
		"SELECT id, name, description, status, priority, created_at, updated_at FROM %ss WHERE ($1 = '' OR status = $1) ORDER BY created_at DESC LIMIT $2 OFFSET $3",
		opts.Status, opts.PageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list %ss: %%w", err)
	}
	defer rows.Close()

	var items []*%s
	for rows.Next() {
		var item %s
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Status, &item.Priority, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan %s: %%w", err)
		}
		items = append(items, &item)
	}
	return items, total, rows.Err()
}

// %sService provides business logic for %s operations.
type %sService struct {
	repo   %sRepository
	logger *slog.Logger
}

func New%sService(repo %sRepository, logger *slog.Logger) *%sService {
	return &%sService{repo: repo, logger: logger}
}

// %sHandler provides HTTP handlers for %s operations.
type %sHandler struct {
	service *%sService
	logger  *slog.Logger
}

func New%sHandler(service *%sService, logger *slog.Logger) *%sHandler {
	return &%sHandler{service: service, logger: logger}
}

func (h *%sHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/%ss", h.List)
	mux.HandleFunc("GET /api/%ss/{id}", h.Get)
	mux.HandleFunc("POST /api/%ss", h.Create)
	mux.HandleFunc("PUT /api/%ss/{id}", h.Update)
	mux.HandleFunc("DELETE /api/%ss/{id}", h.Delete)
}

func (h *%sHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	items, total, err := h.service.repo.List(r.Context(), ListOptions{
		Page: page, PageSize: pageSize,
		Search: r.URL.Query().Get("search"),
		Status: r.URL.Query().Get("status"),
	})
	if err != nil {
		h.logger.Error("list %ss failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"data": items, "total": total, "page": page, "page_size": pageSize})
}

func (h *%sHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ID"})
		return
	}
	item, err := h.service.repo.FindByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "%s not found"})
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *%sHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req Create%sRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	item, err := h.service.repo.Create(r.Context(), req)
	if err != nil {
		h.logger.Error("create %s failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *%sHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var req Update%sRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	item, err := h.service.repo.Update(r.Context(), id, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *%sHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := h.service.repo.Delete(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
`, svc, Svc, svc, Svc,
		Svc, Svc,
		Svc, svc, Svc, Svc, Svc, Svc, Svc, Svc, Svc,
		Svc, Svc, Svc, Svc,
		Svc, Svc, Svc, svc, svc, svc,
		Svc, Svc, svc, svc, svc, svc, Svc, Svc, svc,
		Svc, svc, Svc, Svc,
		Svc, Svc, Svc, Svc,
		Svc, svc, Svc, Svc,
		Svc, Svc, Svc,
		Svc, svc, svc, svc, svc, svc,
		Svc, svc, Svc, svc,
		Svc, Svc, svc, Svc, Svc, Svc, Svc,
	)
}

func genTSXFile(idx int) string {
	components := []string{"Dashboard", "UserList", "ProductGrid", "OrderHistory", "Analytics", "Settings", "Notifications", "Search"}
	comp := components[idx%len(components)]

	return fmt.Sprintf(`import React, { useState, useEffect, useCallback, useMemo, useRef, useContext } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { api } from '@/lib/api';
import { AuthContext } from '@/contexts/AuthContext';
import { useDebounce } from '@/hooks/useDebounce';
import { useInfiniteScroll } from '@/hooks/useInfiniteScroll';
import { Button, Input, Card, Spinner, Badge, Modal, Select, DatePicker } from '@/components/ui';
import { formatDate, formatCurrency } from '@/lib/utils';
import type { PaginatedResponse, ApiError } from '@/types';

interface %sItem {
  id: string;
  name: string;
  description: string;
  status: 'active' | 'inactive' | 'archived';
  priority: number;
  tags: string[];
  createdAt: string;
  updatedAt: string;
  metadata: {
    createdBy: string;
    version: number;
  };
}

interface %sFilters {
  search: string;
  status: string;
  priority: string;
  sortBy: string;
  sortOrder: 'asc' | 'desc';
  dateRange: [Date | null, Date | null];
}

interface %sStats {
  total: number;
  active: number;
  inactive: number;
  archived: number;
  avgPriority: number;
  lastUpdated: string;
}

function use%sData(filters: %sFilters) {
  const [items, setItems] = useState<%sItem[]>([]);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<ApiError | null>(null);
  const [stats, setStats] = useState<%sStats | null>(null);
  const debouncedSearch = useDebounce(filters.search, 300);

  const fetchItems = useCallback(async (pageNum: number, reset: boolean) => {
    setLoading(true);
    setError(null);
    try {
      const [itemsRes, statsRes] = await Promise.all([
        api.get<PaginatedResponse<%sItem>>('/api/items', {
          params: {
            page: pageNum,
            limit: 25,
            search: debouncedSearch,
            status: filters.status || undefined,
            priority: filters.priority || undefined,
            sort_by: filters.sortBy,
            sort_order: filters.sortOrder,
          },
        }),
        pageNum === 1 ? api.get<%sStats>('/api/items/stats') : null,
      ]);

      if (reset) {
        setItems(itemsRes.data.items);
      } else {
        setItems(prev => [...prev, ...itemsRes.data.items]);
      }
      setHasMore(itemsRes.data.hasMore);
      if (statsRes) setStats(statsRes.data);
    } catch (err) {
      setError(err as ApiError);
    } finally {
      setLoading(false);
    }
  }, [debouncedSearch, filters.status, filters.priority, filters.sortBy, filters.sortOrder]);

  useEffect(() => {
    setPage(1);
    fetchItems(1, true);
  }, [fetchItems]);

  const loadMore = useCallback(() => {
    if (!loading && hasMore) {
      const next = page + 1;
      setPage(next);
      fetchItems(next, false);
    }
  }, [loading, hasMore, page, fetchItems]);

  return { items, loading, error, hasMore, loadMore, stats };
}

function ItemCard({ item, onEdit, onDelete }: {
  item: %sItem;
  onEdit: (item: %sItem) => void;
  onDelete: (item: %sItem) => void;
}) {
  const priorityColors = ['', 'bg-gray-100', 'bg-blue-100', 'bg-yellow-100', 'bg-orange-100', 'bg-red-100'];

  return (
    <Card className="p-4 hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <h3 className="font-semibold text-lg">{item.name}</h3>
          <p className="text-gray-600 text-sm mt-1 line-clamp-2">{item.description}</p>
        </div>
        <Badge variant={item.status === 'active' ? 'success' : item.status === 'inactive' ? 'warning' : 'secondary'}>
          {item.status}
        </Badge>
      </div>
      <div className="flex items-center gap-2 mt-3">
        <span className={"px-2 py-1 rounded text-xs " + priorityColors[item.priority]}>
          P{item.priority}
        </span>
        {item.tags.slice(0, 3).map(tag => (
          <Badge key={tag} variant="outline" size="sm">{tag}</Badge>
        ))}
        {item.tags.length > 3 && (
          <span className="text-xs text-gray-400">+{item.tags.length - 3} more</span>
        )}
      </div>
      <div className="flex items-center justify-between mt-4 pt-3 border-t">
        <span className="text-xs text-gray-400">Updated {formatDate(item.updatedAt)}</span>
        <div className="flex gap-1">
          <Button size="sm" variant="ghost" onClick={() => onEdit(item)}>Edit</Button>
          <Button size="sm" variant="ghost" className="text-red-500" onClick={() => onDelete(item)}>Delete</Button>
        </div>
      </div>
    </Card>
  );
}

export default function %s() {
  const { user } = useContext(AuthContext);
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const sentinelRef = useRef<HTMLDivElement>(null);
  const [editItem, setEditItem] = useState<%sItem | null>(null);
  const [deleteItem, setDeleteItem] = useState<%sItem | null>(null);

  const [filters, setFilters] = useState<%sFilters>({
    search: searchParams.get('q') || '',
    status: searchParams.get('status') || '',
    priority: searchParams.get('priority') || '',
    sortBy: searchParams.get('sort') || 'updatedAt',
    sortOrder: (searchParams.get('order') as 'asc' | 'desc') || 'desc',
    dateRange: [null, null],
  });

  const { items, loading, error, hasMore, loadMore, stats } = use%sData(filters);
  useInfiniteScroll(sentinelRef, loadMore, { enabled: hasMore && !loading });

  useEffect(() => {
    const params = new URLSearchParams();
    if (filters.search) params.set('q', filters.search);
    if (filters.status) params.set('status', filters.status);
    if (filters.priority) params.set('priority', filters.priority);
    if (filters.sortBy !== 'updatedAt') params.set('sort', filters.sortBy);
    if (filters.sortOrder !== 'desc') params.set('order', filters.sortOrder);
    setSearchParams(params, { replace: true });
  }, [filters, setSearchParams]);

  const handleDelete = useCallback(async () => {
    if (!deleteItem) return;
    try {
      await api.delete('/api/items/' + deleteItem.id);
      setDeleteItem(null);
      window.location.reload();
    } catch (err) {
      console.error('Delete failed:', err);
    }
  }, [deleteItem]);

  if (!user) { navigate('/login'); return null; }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto px-4 py-6">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-2xl font-bold">%s</h1>
          <Button onClick={() => setEditItem({} as %sItem)}>Create New</Button>
        </div>

        {stats && (
          <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6">
            <Card className="p-3 text-center"><div className="text-2xl font-bold">{stats.total}</div><div className="text-xs text-gray-500">Total</div></Card>
            <Card className="p-3 text-center"><div className="text-2xl font-bold text-green-600">{stats.active}</div><div className="text-xs text-gray-500">Active</div></Card>
            <Card className="p-3 text-center"><div className="text-2xl font-bold text-yellow-600">{stats.inactive}</div><div className="text-xs text-gray-500">Inactive</div></Card>
            <Card className="p-3 text-center"><div className="text-2xl font-bold text-gray-400">{stats.archived}</div><div className="text-xs text-gray-500">Archived</div></Card>
            <Card className="p-3 text-center"><div className="text-2xl font-bold">{stats.avgPriority.toFixed(1)}</div><div className="text-xs text-gray-500">Avg Priority</div></Card>
          </div>
        )}

        <div className="flex flex-wrap gap-3 mb-6 p-4 bg-white rounded-lg shadow-sm">
          <Input placeholder="Search..." value={filters.search} onChange={e => setFilters(f => ({...f, search: e.target.value}))} className="w-64" />
          <Select value={filters.status} onChange={e => setFilters(f => ({...f, status: e.target.value}))}>
            <option value="">All Status</option>
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
            <option value="archived">Archived</option>
          </Select>
          <Select value={filters.sortBy} onChange={e => setFilters(f => ({...f, sortBy: e.target.value}))}>
            <option value="updatedAt">Last Updated</option>
            <option value="name">Name</option>
            <option value="priority">Priority</option>
            <option value="createdAt">Created</option>
          </Select>
        </div>

        {error && <div className="text-red-600 p-4 bg-red-50 rounded mb-4">Error: {error.message}</div>}

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {items.map(item => (
            <ItemCard key={item.id} item={item} onEdit={setEditItem} onDelete={setDeleteItem} />
          ))}
        </div>

        {loading && <div className="flex justify-center py-8"><Spinner size="lg" /></div>}
        <div ref={sentinelRef} />
      </div>

      {deleteItem && (
        <Modal onClose={() => setDeleteItem(null)} title="Confirm Delete">
          <p>Delete <strong>{deleteItem.name}</strong>? This cannot be undone.</p>
          <div className="flex justify-end gap-2 mt-4">
            <Button variant="ghost" onClick={() => setDeleteItem(null)}>Cancel</Button>
            <Button variant="danger" onClick={handleDelete}>Delete</Button>
          </div>
        </Modal>
      )}
    </div>
  );
}
`, comp, comp, comp, comp, comp, comp, comp, comp, comp,
		comp, comp, comp,
		comp, comp, comp, comp, comp,
		comp, comp, comp,
	)
}

func genPyFile(idx int) string {
	services := []string{"users", "products", "orders", "payments", "notifications", "reports", "search", "inventory"}
	svc := services[idx%len(services)]

	return fmt.Sprintf(`"""
%s Service - FastAPI Application
Handles CRUD operations, authentication, and business logic for %s.
"""
from datetime import datetime, timedelta
from typing import Optional, List
from enum import Enum

from fastapi import FastAPI, HTTPException, Depends, Query, status
from fastapi.middleware.cors import CORSMiddleware
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from pydantic import BaseModel, Field, EmailStr
from sqlalchemy import create_engine, Column, Integer, String, Boolean, DateTime, Float, ForeignKey, Text
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, Session, relationship
import jwt
import logging

DATABASE_URL = "postgresql://user:password@localhost:5432/%s_db"
JWT_SECRET = "secret-key-change-in-production"
JWT_ALGORITHM = "HS256"

logger = logging.getLogger(__name__)
engine = create_engine(DATABASE_URL, pool_size=20, max_overflow=10)
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
Base = declarative_base()
security = HTTPBearer()

app = FastAPI(title="%s Service API", version="2.0.0")
app.add_middleware(CORSMiddleware, allow_origins=["*"], allow_methods=["*"], allow_headers=["*"])


class StatusEnum(str, Enum):
    ACTIVE = "active"
    INACTIVE = "inactive"
    ARCHIVED = "archived"
    PENDING = "pending"


class PriorityEnum(int, Enum):
    LOW = 1
    MEDIUM = 2
    HIGH = 3
    CRITICAL = 4
    URGENT = 5


class %sDB(Base):
    __tablename__ = "%s"
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(200), nullable=False, index=True)
    description = Column(Text)
    status = Column(String(20), default="active", index=True)
    priority = Column(Integer, default=3)
    tags = Column(String(500))
    owner_id = Column(Integer, ForeignKey("users.id"))
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class UserDB(Base):
    __tablename__ = "users"
    id = Column(Integer, primary_key=True, index=True)
    email = Column(String(255), unique=True, index=True)
    name = Column(String(100))
    role = Column(String(20), default="user")
    active = Column(Boolean, default=True)


class CreateRequest(BaseModel):
    name: str = Field(..., min_length=1, max_length=200)
    description: Optional[str] = Field(None, max_length=2000)
    priority: PriorityEnum = PriorityEnum.MEDIUM
    tags: List[str] = Field(default_factory=list, max_items=10)


class UpdateRequest(BaseModel):
    name: Optional[str] = Field(None, min_length=1, max_length=200)
    description: Optional[str] = Field(None, max_length=2000)
    status: Optional[StatusEnum] = None
    priority: Optional[PriorityEnum] = None
    tags: Optional[List[str]] = Field(None, max_items=10)


class ItemResponse(BaseModel):
    id: int
    name: str
    description: Optional[str]
    status: str
    priority: int
    tags: List[str]
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class PaginatedResponse(BaseModel):
    items: List[ItemResponse]
    total: int
    page: int
    limit: int
    has_more: bool


class StatsResponse(BaseModel):
    total: int
    active: int
    inactive: int
    archived: int
    avg_priority: float
    last_updated: Optional[datetime]


def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


def get_current_user(credentials: HTTPAuthorizationCredentials = Depends(security), db: Session = Depends(get_db)):
    try:
        payload = jwt.decode(credentials.credentials, JWT_SECRET, algorithms=[JWT_ALGORITHM])
        user = db.query(UserDB).filter(UserDB.id == payload["sub"]).first()
        if not user or not user.active:
            raise HTTPException(status_code=401, detail="Invalid or inactive user")
        return user
    except jwt.InvalidTokenError:
        raise HTTPException(status_code=401, detail="Invalid token")


def require_admin(user: UserDB = Depends(get_current_user)):
    if user.role != "admin":
        raise HTTPException(status_code=403, detail="Admin access required")
    return user


@app.get("/%s", response_model=PaginatedResponse)
def list_items(
    page: int = Query(1, ge=1),
    limit: int = Query(25, ge=1, le=100),
    search: Optional[str] = None,
    status: Optional[StatusEnum] = None,
    priority: Optional[PriorityEnum] = None,
    sort_by: str = Query("updated_at", regex="^(name|created_at|updated_at|priority|status)$"),
    sort_order: str = Query("desc", regex="^(asc|desc)$"),
    user: UserDB = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    query = db.query(%sDB)
    if search:
        query = query.filter(%sDB.name.ilike(f"%%{search}%%") | %sDB.description.ilike(f"%%{search}%%"))
    if status:
        query = query.filter(%sDB.status == status.value)
    if priority:
        query = query.filter(%sDB.priority == priority.value)

    total = query.count()
    offset = (page - 1) * limit

    sort_col = getattr(%sDB, sort_by)
    if sort_order == "desc":
        sort_col = sort_col.desc()
    items = query.order_by(sort_col).offset(offset).limit(limit).all()

    for item in items:
        item.tags = item.tags.split(",") if item.tags else []

    return PaginatedResponse(items=items, total=total, page=page, limit=limit, has_more=(offset + limit) < total)


@app.get("/%s/stats", response_model=StatsResponse)
def get_stats(user: UserDB = Depends(get_current_user), db: Session = Depends(get_db)):
    from sqlalchemy import func
    total = db.query(%sDB).count()
    active = db.query(%sDB).filter(%sDB.status == "active").count()
    inactive = db.query(%sDB).filter(%sDB.status == "inactive").count()
    archived = db.query(%sDB).filter(%sDB.status == "archived").count()
    avg_p = db.query(func.avg(%sDB.priority)).scalar() or 0
    last = db.query(func.max(%sDB.updated_at)).scalar()
    return StatsResponse(total=total, active=active, inactive=inactive, archived=archived, avg_priority=float(avg_p), last_updated=last)


@app.get("/%s/{item_id}", response_model=ItemResponse)
def get_item(item_id: int, user: UserDB = Depends(get_current_user), db: Session = Depends(get_db)):
    item = db.query(%sDB).filter(%sDB.id == item_id).first()
    if not item:
        raise HTTPException(status_code=404, detail="Item not found")
    item.tags = item.tags.split(",") if item.tags else []
    return item


@app.post("/%s", response_model=ItemResponse, status_code=201)
def create_item(req: CreateRequest, user: UserDB = Depends(get_current_user), db: Session = Depends(get_db)):
    item = %sDB(name=req.name, description=req.description, priority=req.priority.value, tags=",".join(req.tags), owner_id=user.id)
    db.add(item)
    db.commit()
    db.refresh(item)
    item.tags = req.tags
    return item


@app.put("/%s/{item_id}", response_model=ItemResponse)
def update_item(item_id: int, req: UpdateRequest, user: UserDB = Depends(get_current_user), db: Session = Depends(get_db)):
    item = db.query(%sDB).filter(%sDB.id == item_id).first()
    if not item:
        raise HTTPException(status_code=404, detail="Item not found")
    for field, value in req.dict(exclude_unset=True).items():
        if field == "tags":
            setattr(item, field, ",".join(value))
        elif isinstance(value, Enum):
            setattr(item, field, value.value)
        else:
            setattr(item, field, value)
    db.commit()
    db.refresh(item)
    item.tags = item.tags.split(",") if item.tags else []
    return item


@app.delete("/%s/{item_id}", status_code=204)
def delete_item(item_id: int, admin: UserDB = Depends(require_admin), db: Session = Depends(get_db)):
    item = db.query(%sDB).filter(%sDB.id == item_id).first()
    if not item:
        raise HTTPException(status_code=404, detail="Item not found")
    db.delete(item)
    db.commit()


@app.get("/health")
def health(db: Session = Depends(get_db)):
    try:
        db.execute("SELECT 1")
        return {"status": "healthy", "service": "%s", "timestamp": datetime.utcnow().isoformat()}
    except Exception as e:
        raise HTTPException(status_code=503, detail=str(e))
`, strings.Title(svc), svc, svc, strings.Title(svc),
		strings.Title(svc), svc,
		svc, strings.Title(svc), strings.Title(svc), strings.Title(svc), strings.Title(svc), strings.Title(svc),
		strings.Title(svc),
		svc, strings.Title(svc), strings.Title(svc), strings.Title(svc), strings.Title(svc), strings.Title(svc),
		strings.Title(svc), strings.Title(svc),
		svc, strings.Title(svc), strings.Title(svc),
		svc, strings.Title(svc),
		svc, strings.Title(svc), strings.Title(svc),
		svc, strings.Title(svc), strings.Title(svc),
		svc,
	)
}

// suppress unused warning
var _ = rand.Intn
