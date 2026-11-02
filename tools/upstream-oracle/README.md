# Development source oracle

This tool compiles the reviewed Microsoft Kusto.Language 13.0.0 source at
`82267c7628e70898cd8bec54c20fc4e61c0a523a`. It is used only to produce parity
evidence. The Go library does not load .NET or ship the source oracle.

Requirements: .NET SDK 10, Go, Python 3, and a clean checkout of the exact
upstream revision. No NuGet packages or Microsoft binaries are required.
The generator calls the same entry points as upstream's checked-in T4
templates, producing syntax nodes and the four command grammar/catalog pairs
in a temporary folder. The original upstream C# files are compiled unchanged.

From the repository root:

Before refreshing a pin, update `UPSTREAM_CHANGELOG.md` with a dated entry
linking to the exact upstream commit. Keep the Area / Impacted functions or
symbols / Confirmed behavior or update table. Give concrete behavior examples
supported by source-oracle comparisons or regressions; label unmeasured
changes as **Source update**. Follow the repository's `AGENTS.md` instructions.
The local provenance check requires the entry for the pinned revision, so CI
rejects a pin update that omits it.

```sh
git clone https://github.com/microsoft/Kusto-Query-Language /tmp/kusto-source
git -C /tmp/kusto-source checkout 82267c7628e70898cd8bec54c20fc4e61c0a523a
python3 tools/upstream-oracle/refresh.py /tmp/kusto-source
```

The script independently replays both checked-in corpora against the source
oracle and `tools/upstream-parity-native`. It compares complete ordered
diagnostics, sorted columns, open-table state, and result type before writing
canonical outputs. It also records API-level Unicode, scanner, literal,
completion, and boundary cases under five cultures, requiring identical
results. The cache/grammar fixture additionally records source and global cache
factory calls, retries, and all eight shared grammar variants. Revision-specific
fixture filenames preserve historical evidence instead of overwriting it. Only after all comparisons pass does it refresh the source pin,
notice hash, result hashes, fixture, and supporting tool hashes in `UPSTREAM.json`.

The compatibility fixture uses .NET 10's ICU casing. Microsoft documents a
small set of BMP casing differences between ICU and older NLS hosts. The
port keeps the existing native raw-hex integer convenience; typed `long(0x10)`
and `int(0x10)` follow upstream's zero-value recovery.

Run normal Go checks after refreshing. They replay the corpus and API fixtures
without requiring .NET. `go run ./cmd/kusto-upstream-check -verify-provenance
-local-only` also detects edits to the oracle tools or compatibility fixture.
