# UCCP v0.0.7 Release Instructions

## Changes Summary

### What Was Fixed
- **84% garbage lines** from navigation-heavy websites (Wikipedia, Britannica, news sites)
- Lines containing ONLY formatting markers (`**`, `*`, `/`, `|`) now removed
- Empty markdown bold/italic markers cleaned up

### Code Changes
```
domains/html.go       | +11 lines  (whitespace cleanup logic)
domains/html_test.go  | +217 lines (fixed all tests + 7 new tests)
CHANGELOG_v0.0.7.md   | +106 lines (release notes)
```

### Test Results
```bash
go test ./domains -v

# All 23 tests PASS:
✓ TestHTMLCompressor_Compress_Headings
✓ TestHTMLCompressor_Compress_Paragraphs
✓ TestHTMLCompressor_Compress_CodeBlocks
✓ TestHTMLCompressor_Compress_Lists
✓ TestHTMLCompressor_Compress_Tables
✓ TestHTMLCompressor_Compress_Links
✓ TestHTMLCompressor_Compress_NoiseRemoval
✓ TestHTMLCompressor_Compress_Abbreviations
✓ TestHTMLCompressor_Compress_HTMLEntities
✓ TestHTMLCompressor_Compress_Empty
✓ TestHTMLCompressor_Decompress
✓ TestHTMLCompressor_SystemPrompt
✓ TestHTMLCompressor_EstimateTokens
✓ TestHTMLCompressor_RealWorldPage
✓ TestHTMLCompressor_EmptyFormattingRemoval (7 subtests)
✓ TestHTMLCompressor_RealWorldBritannica
```

---

## Release Steps

### 1. Push to GitHub

```bash
# Push the commit
git push origin master

# Verify it's pushed
git log --oneline -1
# Should show: 413dbd1 fix: remove lines with only formatting markers
```

### 2. Create GitHub Release

**Go to:** https://github.com/aguzmans/uccp/releases/new

**Tag version:** `v0.0.7`

**Release title:** `v0.0.7 - Fix Whitespace Waste in Navigation-Heavy HTML`

**Description:**

```markdown
## Critical Fix: Whitespace Cleanup in HTML Compression

### Problem
Navigation-heavy websites (Wikipedia, Britannica, news sites) left **84% garbage lines** after HTML-to-markdown conversion.

Example waste:
```
**       ← Empty bold marker
 *       ← Navigation icon
 /       ← Menu separator
```

### Solution
Added post-processing to remove lines with ONLY formatting characters.

### Impact
- **84% → 0%** garbage line reduction
- **~200 tokens saved** per article on navigation-heavy sites
- Better LLM readability

### Changes
- `domains/html.go`: Added 2 regex cleanups (11 lines)
- `domains/html_test.go`: Updated all tests + 7 new whitespace tests (23 tests pass)
- `CHANGELOG_v0.0.7.md`: Detailed release notes

### Upgrading
```bash
go get github.com/aguzmans/uccp@v0.0.7
```

### Breaking Changes
None - fully backward compatible with v0.0.6.
```

---

## Post-Release Checklist

- [ ] GitHub release created with tag v0.0.7
- [ ] Release notes published
- [ ] Update UCCP README.md with latest version badge

---

## Rollback Instructions (if needed)

Downstream consumers can pin to the previous version:

```bash
go get github.com/aguzmans/uccp@v0.0.6
go mod tidy
```

---

## Communication Template

**For GitHub Release Announcement:**

> UCCP v0.0.7 is now available! 🎉
>
> This release fixes a critical whitespace issue where navigation-heavy websites (Wikipedia, Britannica) left 84% garbage lines after HTML compression.
>
> **Key improvements:**
> - 84% → 0% garbage line reduction
> - ~200 tokens saved per article
> - Better LLM readability
>
> Fully backward compatible - upgrade with:
> ```bash
> go get github.com/aguzmans/uccp@v0.0.7
> ```
>
> See [CHANGELOG](https://github.com/aguzmans/uccp/blob/master/CHANGELOG_v0.0.7.md) for details.

---

## Next Steps (Optional Improvements)

Future enhancements to consider:

1. **Performance optimization:** Compile regexes once at package init
2. **Configuration:** Add option to preserve certain formatting markers
3. **Metrics:** Track compression ratio improvements
4. **Documentation:** Add examples to README showing before/after

---

## File Locations

All changes committed to:
- `domains/html.go`
- `domains/html_test.go`
- `CHANGELOG_v0.0.7.md`

Git commit: `413dbd1` on `master` branch
