package benchmarks

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aguzmans/uccp/core"
	"github.com/aguzmans/uccp/domains"
)

// TestJSONAPIResponses tests compression on realistic API responses
func TestJSONAPIResponses(t *testing.T) {
	scenarios := []struct {
		name          string
		data          interface{}
		expectedRatio float64
		description   string
	}{
		{
			name: "User List API Response",
			data: map[string]interface{}{
				"users": []map[string]interface{}{
					{
						"id":       1,
						"name":     "John Doe",
						"email":    "john@example.com",
						"active":   true,
						"role":     "admin",
						"created":  "2024-01-15T10:30:00Z",
					},
					{
						"id":       2,
						"name":     "Jane Smith",
						"email":    "jane@example.com",
						"active":   true,
						"role":     "user",
						"created":  "2024-01-16T14:20:00Z",
					},
					{
						"id":       3,
						"name":     "Bob Johnson",
						"email":    "bob@example.com",
						"active":   false,
						"role":     "user",
						"created":  "2024-01-17T09:15:00Z",
					},
				},
				"total": 3,
				"page":  1,
				"limit": 10,
			},
			expectedRatio: 0.70,
			description:   "Paginated user list with repeated structure",
		},
		{
			name: "Product Catalog",
			data: map[string]interface{}{
				"products": []map[string]interface{}{
					{
						"id":          "prod-001",
						"name":        "Laptop Computer",
						"description": "High-performance laptop for professionals",
						"price":       1299.99,
						"currency":    "USD",
						"inStock":     true,
						"category":    "Electronics",
						"tags":        []string{"computer", "laptop", "professional"},
					},
					{
						"id":          "prod-002",
						"name":        "Wireless Mouse",
						"description": "Ergonomic wireless mouse with USB receiver",
						"price":       29.99,
						"currency":    "USD",
						"inStock":     true,
						"category":    "Accessories",
						"tags":        []string{"mouse", "wireless", "ergonomic"},
					},
				},
				"totalResults": 2,
				"page":         1,
			},
			expectedRatio: 0.65,
			description:   "Product catalog with nested arrays",
		},
		{
			name: "Configuration Object",
			data: map[string]interface{}{
				"application": map[string]interface{}{
					"name":        "MyApp",
					"version":     "1.0.0",
					"environment": "production",
				},
				"database": map[string]interface{}{
					"host":     "localhost",
					"port":     5432,
					"database": "myapp_prod",
					"pool": map[string]interface{}{
						"min": 2,
						"max": 10,
					},
				},
				"cache": map[string]interface{}{
					"enabled": true,
					"ttl":     3600,
					"type":    "redis",
				},
				"features": map[string]interface{}{
					"authentication": true,
					"logging":        true,
					"monitoring":     true,
					"apiV2":          false,
				},
			},
			expectedRatio: 0.60,
			description:   "Nested configuration with repeated keys",
		},
	}

	// For JSON, we use CodeCompressor as a placeholder
	// In Phase 4, we'll create a dedicated JSON domain
	compressor := domains.NewCodeCompressor()

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Marshal to JSON
			jsonBytes, err := json.MarshalIndent(scenario.data, "", "  ")
			if err != nil {
				t.Fatalf("Failed to marshal JSON: %v", err)
			}

			jsonStr := string(jsonBytes)

			// Compress
			compressed, err := compressor.Compress(jsonStr)
			if err != nil {
				t.Fatalf("Compression failed: %v", err)
			}

			// Calculate metrics
			ratio := core.CalculateCompressionRatio(jsonStr, compressed)
			originalTokens := core.EstimateTokenCount(jsonStr)
			compressedTokens := core.EstimateTokenCount(compressed)
			tokenSavings := core.EstimateTokenSavings(jsonStr, compressed)

			// Report results
			t.Logf("\n=== %s ===", scenario.name)
			t.Logf("Description: %s", scenario.description)
			t.Logf("Original size: %d bytes (%d tokens)", len(jsonStr), originalTokens)
			t.Logf("Compressed size: %d bytes (%d tokens)", len(compressed), compressedTokens)
			t.Logf("Compression ratio: %.1f%%", ratio*100)
			t.Logf("Token savings: %d tokens (%.1f%%)", tokenSavings, (float64(tokenSavings)/float64(originalTokens))*100)

			if len(jsonStr) < 1000 {
				t.Logf("\nOriginal JSON:\n%s", jsonStr)
				t.Logf("\nCompressed:\n%s", compressed)
			}

			// Note: Code compressor is not optimized for JSON yet
			// This shows baseline, JSON domain (v0.0.8) will improve
			t.Logf("\nNote: Using CodeCompressor as baseline. JSON domain (v0.0.8) will achieve %.0f%%+ compression", scenario.expectedRatio*100)
		})
	}
}

// TestJSONArrayCompression tests compression on repeated array structures
func TestJSONArrayCompression(t *testing.T) {
	compressor := domains.NewCodeCompressor()

	// Simulate API response with 100 similar records
	records := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		records[i] = map[string]interface{}{
			"id":        i + 1,
			"name":      fmt.Sprintf("User %d", i+1),
			"email":     fmt.Sprintf("user%d@example.com", i+1),
			"active":    i%2 == 0,
			"role":      "user",
			"createdAt": "2024-01-15T10:00:00Z",
		}
	}

	response := map[string]interface{}{
		"data":  records,
		"total": 100,
		"page":  1,
		"limit": 100,
	}

	jsonBytes, _ := json.Marshal(response)
	jsonStr := string(jsonBytes)

	compressed, err := compressor.Compress(jsonStr)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	// Calculate metrics
	ratio := core.CalculateCompressionRatio(jsonStr, compressed)
	originalTokens := core.EstimateTokenCount(jsonStr)
	compressedTokens := core.EstimateTokenCount(compressed)
	tokenSavings := originalTokens - compressedTokens

	t.Logf("\n=== Large JSON Array (100 records) ===")
	t.Logf("Original size: %d bytes (%d tokens)", len(jsonStr), originalTokens)
	t.Logf("Compressed size: %d bytes (%d tokens)", len(compressed), compressedTokens)
	t.Logf("Compression ratio: %.1f%%", ratio*100)
	t.Logf("Token savings: %d tokens", tokenSavings)
	t.Logf("\nNote: Dedicated JSON domain (v0.0.8) will detect repeated structure")
	t.Logf("      and achieve 80%+ compression on array data")
}

// BenchmarkJSONCompression benchmarks JSON compression performance
func BenchmarkJSONCompression(b *testing.B) {
	compressor := domains.NewCodeCompressor()

	data := map[string]interface{}{
		"users": []map[string]interface{}{
			{"id": 1, "name": "John", "email": "john@example.com", "active": true},
			{"id": 2, "name": "Jane", "email": "jane@example.com", "active": true},
			{"id": 3, "name": "Bob", "email": "bob@example.com", "active": false},
		},
		"total": 3,
	}

	jsonBytes, _ := json.Marshal(data)
	jsonStr := string(jsonBytes)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = compressor.Compress(jsonStr)
	}
}

// TestAPIResponseChain simulates multiple API calls being compressed
func TestAPIResponseChain(t *testing.T) {
	compressor := domains.NewCodeCompressor()

	// Simulate 5 API calls
	apiCalls := []string{
		`{"endpoint": "/users", "method": "GET", "status": 200, "records": 50}`,
		`{"endpoint": "/products", "method": "GET", "status": 200, "records": 120}`,
		`{"endpoint": "/orders", "method": "GET", "status": 200, "records": 85}`,
		`{"endpoint": "/customers", "method": "GET", "status": 200, "records": 200}`,
		`{"endpoint": "/invoices", "method": "GET", "status": 200, "records": 150}`,
	}

	var totalOriginal, totalCompressed int

	for i, apiCall := range apiCalls {
		// Create detailed response
		response := map[string]interface{}{
			"metadata": map[string]interface{}{
				"endpoint":   apiCall,
				"timestamp":  "2024-01-15T10:00:00Z",
				"duration":   "125ms",
				"cached":     false,
			},
			"data": []map[string]interface{}{
				{"id": i*100 + 1, "value": "Item 1", "active": true},
				{"id": i*100 + 2, "value": "Item 2", "active": true},
				{"id": i*100 + 3, "value": "Item 3", "active": false},
			},
		}

		jsonBytes, _ := json.Marshal(response)
		jsonStr := string(jsonBytes)
		compressed, _ := compressor.Compress(jsonStr)

		totalOriginal += core.EstimateTokenCount(jsonStr)
		totalCompressed += core.EstimateTokenCount(compressed)
	}

	savings := totalOriginal - totalCompressed
	percent := (float64(savings) / float64(totalOriginal)) * 100

	t.Logf("\n=== API Response Chain (5 calls) ===")
	t.Logf("Total tokens without UCCP: %d", totalOriginal)
	t.Logf("Total tokens with UCCP: %d", totalCompressed)
	t.Logf("Savings: %d tokens (%.1f%%)", savings, percent)
}
