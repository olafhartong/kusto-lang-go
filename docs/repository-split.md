# Repository split

kusto-lang-go was extracted from the local KQLAnalyzer working tree on 2026-10-19,
starting from commit `cbe3bff9bffa53c7af61bf552ba3ebac5a88c933`. Both new repositories start with fresh
Git histories. The source folder was left intact. Local binaries, reports,
caches, prior Git history, and ignored deliverables were not copied.

## Ownership

| Repository | Owns |
| --- | --- |
| kusto-lang-go | `kustolang`, `kustolang/engine`, `kustolang/environment`, language tests, generators, metadata, parity artifacts, upstream provenance and checks |
| KustoKITT | Analyzer API, CLI, HTTP service, playground, rule scanner, operational schemas, NRT catalogs, binary fixtures, Docker and executable releases |

Existing language imports change from `github.com/olafhartong/kustokitt/kustolang`
to `github.com/olafhartong/kusto-lang-go/kustolang`. The `/kustolang` suffix is
retained. Internal language consumers now use the public `/kustolang/engine`
and `/kustolang/environment` APIs. Engine parsed-code positions remain byte
based; facade diagnostics and editor services retain their UTF-16 contracts.
No standalone language server was added. Existing GPL project licenses and
upstream Apache/MIT attribution are retained; no relicensing is implied.

## Local development

Keep both folders beside each other. From KustoKITT:

```sh
go work init . ../kusto-lang-go
go work edit -replace=github.com/olafhartong/kusto-lang-go@v0.1.0=../kusto-lang-go
make check
```

The split already includes this ignored local `go.work`. Do not run `go work
init` again if it exists. The language repository builds independently without
KustoKITT. Its separate `tools/langsmoke` module uses a local replacement of
its own root module, with `GOWORK=off`.

## Publish the initial repositories

1. Create `olafhartong/kusto-lang-go` and publish its initial commit on `main`.
2. Tag and push the language library as `v0.1.0` after its own checks pass.
3. Publish `olafhartong/KustoKITT` on `main`. Its `go.mod` requires that tag.
4. After publication, run `GOWORK=off go mod tidy` in KustoKITT and commit the
   generated language-module checksums. Independent consumers can then use
   the published dependency without a sibling checkout or workspace.

The initial version is a planned tag, not a claim that it already exists.
KustoKITT CI and releases check out the exact version parsed from `go.mod`
into a sibling folder and create a temporary workspace. These workflows need
the library repository and tag to exist; local checks already work before
publication. CI intentionally uses the pinned version rather than language
`main`.

For a library upgrade, release a new language tag, then update KustoKITT with
`GOWORK=off go get github.com/olafhartong/kusto-lang-go@<version>` and commit
both module files. Test against the pinned release with `GOWORK=off go test
./...`; the local workspace otherwise selects the sibling checkout.

## Containers

KustoKITT's Compose build supplies the sibling language checkout as a named
build context. For a direct local build, run from KustoKITT:

```sh
docker build --build-context kusto_lang_go=../kusto-lang-go -t kustokitt .
```

The Dockerfile creates a build-time workspace and does not ship a Go runtime.
The supplied language checkout determines the compiled engine; use the tag
required by `go.mod` for release-equivalent builds. Docker Compose requires
support for `additional_contexts`.

## Checksums and attribution

The native and oracle parity JSON artifacts, source snapshot, preserved
third-party licenses, and corpus generators are unchanged by the extraction.
The provenance manifest records the same upstream commit. Its notice checksum
was refreshed because notices now describe the standalone library. API
snapshots were regenerated for the package paths and repository ownership.
