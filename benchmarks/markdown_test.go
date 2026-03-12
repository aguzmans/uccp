package benchmarks

import (
	"testing"

	"github.com/aguzmans/uccp/core"
	"github.com/aguzmans/uccp/domains"
)

// TestMarkdownDocumentation tests compression on various Markdown formats
func TestMarkdownDocumentation(t *testing.T) {
	scenarios := []struct {
		name          string
		markdown      string
		expectedRatio float64
		description   string
	}{
		{
			name: "README Documentation",
			markdown: `# MyProject

## Overview

This is a comprehensive project for building web applications.

## Features

- **Authentication**: JWT-based authentication with refresh tokens
- **Database**: PostgreSQL with migrations and seeding
- **API**: RESTful API with OpenAPI documentation
- **Testing**: Unit tests with Jest, integration tests with Supertest

## Installation

\`\`\`bash
npm install
npm run migrate
npm run seed
npm start
\`\`\`

## Configuration

Create a \`.env\` file with the following variables:

\`\`\`env
DATABASE_URL=postgresql://localhost:5432/myapp
JWT_SECRET=your-secret-key
PORT=3000
\`\`\`

## API Endpoints

### Authentication

- \`POST /auth/login\` - Login with email and password
- \`POST /auth/register\` - Register new user
- \`POST /auth/refresh\` - Refresh access token

### Users

- \`GET /users\` - List all users (admin only)
- \`GET /users/:id\` - Get user by ID
- \`PUT /users/:id\` - Update user
- \`DELETE /users/:id\` - Delete user (admin only)

## License

MIT License - see LICENSE file for details.
`,
			expectedRatio: 0.65,
			description:   "Standard README with code blocks and lists",
		},
		{
			name: "API Documentation",
			markdown: `# API Reference

## User Management

### List Users

\`\`\`http
GET /api/users
\`\`\`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| page | integer | No | Page number (default: 1) |
| limit | integer | No | Results per page (default: 20) |
| sort | string | No | Sort field (default: created_at) |

**Response:**

\`\`\`json
{
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com"
    }
  ],
  "total": 100,
  "page": 1
}
\`\`\`

### Create User

\`\`\`http
POST /api/users
\`\`\`

**Request Body:**

\`\`\`json
{
  "name": "Jane Smith",
  "email": "jane@example.com",
  "password": "securepassword"
}
\`\`\`

**Response:**

\`\`\`json
{
  "id": 2,
  "name": "Jane Smith",
  "email": "jane@example.com",
  "created_at": "2024-01-15T10:00:00Z"
}
\`\`\`
`,
			expectedRatio: 0.70,
			description:   "API documentation with tables and code examples",
		},
		{
			name: "Agent Communication",
			markdown: `# Planning Session Summary

## Goals

Create a REST API for e-commerce platform

## Architecture Decisions

1. **Framework**: Express.js with TypeScript
2. **Database**: PostgreSQL with Sequelize ORM
3. **Authentication**: JWT tokens with bcrypt password hashing
4. **Testing**: Jest for unit tests, Supertest for integration

## Tasks Breakdown

### Phase 1: Setup (Priority 1-2)

- [ ] Initialize Node.js project with TypeScript
- [ ] Configure database connection
- [ ] Set up authentication middleware
- [ ] Create user model and migrations

### Phase 2: Core Features (Priority 3-5)

- [ ] Implement product catalog API
- [ ] Add shopping cart functionality
- [ ] Create order management system
- [ ] Implement payment integration

### Phase 3: Polish (Priority 6-8)

- [ ] Add API documentation
- [ ] Write integration tests
- [ ] Set up logging and monitoring
- [ ] Create deployment scripts

## Patterns to Follow

- Use async/await for all database operations
- Validate input using Joi schemas
- Handle errors with centralized error middleware
- Return consistent API response format

## Next Steps

Planning agent will create job files for each task. Workers will execute them in priority order.
`,
			expectedRatio: 0.60,
			description:   "Planning session with tasks and decisions",
		},
	}

	// For Markdown, we use CodeCompressor as baseline
	// In Phase 2 (v0.0.6), we'll create dedicated Markdown domain
	compressor := domains.NewCodeCompressor()

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Compress
			compressed, err := compressor.Compress(scenario.markdown)
			if err != nil {
				t.Fatalf("Compression failed: %v", err)
			}

			// Calculate metrics
			ratio := core.CalculateCompressionRatio(scenario.markdown, compressed)
			originalTokens := core.EstimateTokenCount(scenario.markdown)
			compressedTokens := core.EstimateTokenCount(compressed)
			tokenSavings := core.EstimateTokenSavings(scenario.markdown, compressed)

			// Report results
			t.Logf("\n=== %s ===", scenario.name)
			t.Logf("Description: %s", scenario.description)
			t.Logf("Original size: %d bytes (%d tokens)", len(scenario.markdown), originalTokens)
			t.Logf("Compressed size: %d bytes (%d tokens)", len(compressed), compressedTokens)
			t.Logf("Compression ratio: %.1f%%", ratio*100)
			t.Logf("Token savings: %d tokens (%.1f%%)", tokenSavings, (float64(tokenSavings)/float64(originalTokens))*100)

			if len(scenario.markdown) < 1000 {
				t.Logf("\nOriginal Markdown:\n%s", scenario.markdown[:min(len(scenario.markdown), 500)])
				t.Logf("\nCompressed:\n%s", compressed)
			}

			// Note: Code compressor provides baseline
			t.Logf("\nNote: Using CodeCompressor as baseline. Markdown domain (v0.0.6) will achieve %.0f%%+ compression", scenario.expectedRatio*100)
		})
	}
}

// TestMarkdownAgentMessages tests compression of agent-to-agent messages
func TestMarkdownAgentMessages(t *testing.T) {
	compressor := domains.NewCodeCompressor()

	messages := []string{
		`# Task Complete

Successfully implemented the authentication endpoints.

## What was done:
- Created POST /auth/login endpoint
- Created POST /auth/register endpoint
- Added JWT token generation
- Implemented password hashing with bcrypt

## Tests:
- All 12 tests passing
- Coverage: 95%

## Files modified:
- src/routes/auth.ts
- src/controllers/auth.ts
- src/middleware/auth.ts
- tests/auth.test.ts
`,
		`# Planning Summary

Breaking down the e-commerce API into 15 jobs.

## Architecture:
- Monorepo with packages: api, database, shared
- TypeScript throughout
- PostgreSQL for data persistence
- Redis for caching

## Job Priority:
1. Database setup (priority 1)
2. Authentication (priority 2)
3. Product catalog (priority 3)
4. Shopping cart (priority 4)
5. Orders & payments (priority 5)
`,
		`# Error Report

Job failed during database migration.

## Error:
\`\`\`
Error: Connection refused to localhost:5432
\`\`\`

## Cause:
PostgreSQL service not running

## Fix:
Start PostgreSQL: \`docker-compose up -d postgres\`

## Retry:
Job can be retried after fix
`,
	}

	var totalOriginal, totalCompressed int

	for i, msg := range messages {
		compressed, _ := compressor.Compress(msg)

		original := core.EstimateTokenCount(msg)
		comp := core.EstimateTokenCount(compressed)

		totalOriginal += original
		totalCompressed += comp

		t.Logf("Message %d: %d → %d tokens (%.1f%% compression)",
			i+1, original, comp, core.CalculateCompressionRatio(msg, compressed)*100)
	}

	savings := totalOriginal - totalCompressed
	percent := (float64(savings) / float64(totalOriginal)) * 100

	t.Logf("\n=== Agent Messages Summary ===")
	t.Logf("Total tokens without UCCP: %d", totalOriginal)
	t.Logf("Total tokens with UCCP: %d", totalCompressed)
	t.Logf("Savings: %d tokens (%.1f%%)", savings, percent)
}

// BenchmarkMarkdownCompression benchmarks Markdown compression
func BenchmarkMarkdownCompression(b *testing.B) {
	compressor := domains.NewCodeCompressor()

	markdown := `# API Documentation

## Overview

This is the API documentation.

## Endpoints

- \`GET /users\` - List users
- \`POST /users\` - Create user
- \`GET /users/:id\` - Get user by ID
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = compressor.Compress(markdown)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
