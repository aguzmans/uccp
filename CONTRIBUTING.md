# Contributing to UCCP

Thanks for your interest in improving UCCP.

## Ways to Contribute

- **Bug reports**: Open an issue with a minimal reproduction case.
- **Feature ideas**: Open a discussion before large PRs so we can align on scope.
- **Pull requests**: Fixes, new domains, benchmarks, docs — all welcome.

Priority areas are listed in [README.md](README.md#contributing) and [ROADMAP.md](ROADMAP.md).

## Development

```bash
git clone https://github.com/aguzmans/uccp
cd uccp
go build ./...
go test ./... -race
```

## Pull Request Checklist

- [ ] `go test ./...` passes locally
- [ ] `go vet ./...` is clean
- [ ] New behavior has tests
- [ ] Public API changes are noted in the PR description
- [ ] Docs/examples updated if behavior changed

CI runs `go build` and `go test -race` on push and PR — see [`.github/workflows/test.yml`](.github/workflows/test.yml).

## Adding a New Domain

Domains live in `domains/` and implement the `core.Compressor` interface. Add:

1. Compressor implementation in `domains/<name>.go`
2. Tests in `domains/<name>_test.go`
3. Benchmark scenario in `benchmarks/<name>_test.go`

See `domains/html.go` for a reference implementation.

## Code Style

- Standard Go formatting (`gofmt`)
- Keep exported symbols documented
- Prefer small, focused PRs

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
