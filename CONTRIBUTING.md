# Contributing

Thanks for your interest. A few practical notes:

## Building and testing

- `make build` — compile the embeddable library and development utilities.
- `make check` — run tests and `go vet` (mirrors CI).
- `make lint` — run `golangci-lint` and `govulncheck` (CI enforces both).
- `make race` / `make cover` / `make fuzz-short` — mirror the corresponding CI steps.

Run the full language suite before submitting changes. KustoKITT integration
tests live in its separate repository.

## Ground rules

- **Generated files** (`*_generated.go`, catalogs) must not be edited by hand;
  regenerate them with the tools under `tools/` — CI verifies drift.
- The `kustolang` package is a port of
  [microsoft/Kusto-Query-Language](https://github.com/microsoft/Kusto-Query-Language);
  behavioral changes there must be justified by upstream parity, and are
  verified by the parity corpus in CI.
- Language runtime packages must remain standard-library-only. Development
  generators may use the existing YAML module; avoid adding runtime dependencies.
- Language behavior changes need regression coverage. KustoKITT is the concrete
consumer of the public engine and environment APIs.

## API and port boundaries

- Put new language implementation in `kustolang/engine`. Add to the
  `kustolang` facade only for a concrete external consumer, and record that
  consumer in the change. Do not add external tests that only prove a wrapper
  forwards to the internal package.
- `go generate ./kustolang` preserves the existing generated public facade by
  default. The `-allow-api-expansion` generator flag is reserved for a reviewed
  public API addition backed by a concrete external consumer. Any such change
  must include the public API snapshot update.
- Do not add `ThrowIfCancellationRequested` calls to hand-written Go. Prefer
  checking `context.Context` at bounded work boundaries.
- Do not add `recovered*` analysis flags without documenting their category and
  deletion condition.
- Do not add overload-style names such as
  `WithParserProfileAndOptionsContext`. Extend an options type or use the
  canonical entry point instead.
- Keep compatibility names only when they serve an existing caller. Mark them
  with a standard `Deprecated:` doc comment and remove them only in a major
  release.

## Absorbing upstream Kusto.Language drift (repin playbook)

The port is pinned to an exact upstream commit in
`third_party/kusto/UPSTREAM.json`. The `Kusto upstream check` workflow runs
`cmd/kusto-upstream-check` weekly and fails when upstream changes the configured
source paths (or when local provenance/snapshot hashes drift). Changes outside
those paths are informational when the comparison is complete. To absorb a new
upstream batch:

1. **Diff**: clone [microsoft/Kusto-Query-Language](https://github.com/microsoft/Kusto-Query-Language)
   and diff all `source_paths` from `UPSTREAM.json` between the pinned commit
   and the target commit, including `src/Kusto.Language.Generators` and
   `grammar` as well as `src/Kusto.Language`. Generator definitions can change
   command syntax and result schemas exposed by the language library.
2. **Classify** each change: port it, or record it as not applicable (e.g.
   .NET-internal memory-layout details) in the `notes` field of
   `UPSTREAM.json`.
3. **Port** the applicable changes in `kustolang/engine`, updating the
   pinned inventory tests alongside them.
4. **Regenerate**: run `go generate ./...` from the `kustolang/` directory
   (running the generators from their own directories breaks relative-root
   resolution). If the public API changed, regenerate the guard golden via
   `go test ./internal/guard -run TestPublicAPISnapshot -update`. If catalog
   membership changed, regenerate the inventory golden via
   `go test ./kustolang/engine -run TestCatalogInventoriesMatchPinnedUpstream -update-catalog-attestation`
   and review the diff against the upstream source diff.
5. **Re-attest**: update `commit`, the `current` profile, and the attestation
   commits in `UPSTREAM.json`. If the drift touches symbols exercised by the
   parity corpora, re-run the .NET oracle and refresh the attested hashes;
   otherwise verify (and note) that the corpora are unaffected. If
   `THIRD_PARTY_NOTICES.md` changed, recompute `notice_sha256`.
6. **Gate**: `make check`, `make lint`, `make race`, `make langsmoke`, and
   `go run ./cmd/kusto-upstream-check` must all pass; the last one should
   report "current at `<target>`".

## Security issues

See [SECURITY.md](SECURITY.md) — do not open public issues for vulnerabilities.
