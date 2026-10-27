# Development source oracle

This tool compiles the reviewed Microsoft Kusto.Language 13.0.0 source at
`328e05d10d6c73654e416675fc75af7ed0e05ffd`. It is used only to produce parity
evidence. The Go library does not load .NET or ship the source oracle.

Requirements: .NET SDK 10, Go, Python 3, and a clean checkout of the exact
upstream revision. No NuGet packages or Microsoft binaries are required.
The generator calls the same entry points as upstream's checked-in T4
templates, producing syntax nodes and the four command grammar/catalog pairs
in a temporary folder. The original upstream C# files are compiled unchanged.

From the repository root:

```sh
git clone https://github.com/microsoft/Kusto-Query-Language /tmp/kusto-source
git -C /tmp/kusto-source checkout 328e05d10d6c73654e416675fc75af7ed0e05ffd
python3 tools/upstream-oracle/refresh.py /tmp/kusto-source
```

The script independently replays both checked-in corpora against the source
oracle and `tools/upstream-parity-native`. It compares complete ordered
diagnostics, sorted columns, open-table state, and result type before writing
canonical outputs. It also records API-level Unicode, scanner, literal,
completion, and boundary cases under five cultures, requiring identical
results. Only after all comparisons pass does it refresh the source pin,
notice hash, result hashes, fixture, and supporting tool hashes in `UPSTREAM.json`.

The compatibility fixture uses .NET 10's ICU casing. Microsoft documents a
small set of BMP casing differences between ICU and older NLS hosts. The
port keeps the existing native raw-hex integer convenience; typed `long(0x10)`
and `int(0x10)` follow upstream's zero-value recovery.

Run normal Go checks after refreshing. They replay the corpus and API fixtures
without requiring .NET. `go run ./cmd/kusto-upstream-check -verify-provenance
-local-only` also detects edits to the oracle tools or compatibility fixture.
