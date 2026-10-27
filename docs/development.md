# Development, validation, and provenance

## Requirements and installation

Use Go 1.27.2 or newer (the module retains Go 1.25 language compatibility).

Install the library as a dependency with:

```sh
go get github.com/olafhartong/kusto-lang-go
```

The [repository split guide](repository-split.md) records the original
`v0.1.0` publication plan, local workspaces, and KustoKITT migration commands.
The language repository builds independently of KustoKITT.

## Build and validate

```sh
make check
make race
make generate
make provenance
```

`make check` runs the tests, vet, provenance validation, and the independent
consumer smoke module. `make langsmoke` runs that smoke check on its own;
`make lint` runs the linter and vulnerability scanner. `make race` checks the
public language, engine, and environment packages. `make generate` refreshes
the generated public facade and catalogs, which CI verifies are deterministic.
The ASIM metadata refresh utility uses a Go YAML dependency at development
time; the runtime language packages use only the Go standard library.

## Repository layout

- `kustolang/` — public facade and editor API
- `kustolang/engine/` — native source port and its tests
- `kustolang/environment/` — schema model, environment support, and metadata
- `tools/` — generators and independent consumer smoke module
- `cmd/kusto-upstream-check/` — development-time upstream/provenance checker
- `parity-artifacts/` — commit-bound differential evidence
- `third_party/` — upstream licenses and source snapshots

## Licensing and provenance

Copyright 2026 Olaf Hartong. The existing GPL-3.0 project license is retained
in [LICENSE](../LICENSE); this extraction does not relicense the project.
Ported Microsoft code retains its Apache-2.0 notices and preserved license.
Sentinel metadata retains its MIT attribution. See
[Third-party notices](../THIRD_PARTY_NOTICES.md). FalconForce NRT catalogs remain
in KustoKITT and are not distributed by this repository.

## Upstream tracking

The pinned Microsoft revision is checked weekly in CI. The scheduled check also
classifies changes to grammar/parser syntax, functions/operators/plug-ins,
commands, diagnostics, and editor sources, and compares vendored generator
source snapshots with the exact files at the pinned GitHub commit. Run the same
standard-library-only revision check locally with:

```bash
go run ./cmd/kusto-upstream-check
```

The command resolves upstream `HEAD` and compares it with the pinned revision
using GitHub's commit-comparison API. Identical revisions pass. Upstream
advancement is informational when a complete comparison contains no changes
within the configured source paths. Changes in tracked source paths, behind or
diverged histories, missing pins, incomplete or malformed comparisons, and API
errors fail the check. See [upstream parity](upstream-parity.md) for the detailed
scope and failure rules.

Verify the recorded revisions, license hash, notice, manifest paths, and
commit-bound differential artifacts without network access with:

```bash
go run ./cmd/kusto-upstream-check -verify-provenance -local-only
```

Changing the pinned revision requires a fresh differential attestation for
that exact commit. The checker verifies each checked-in corpus, generator,
C#-oracle result, and native result hash and rejects stale or missing evidence.

Every compatibility update must also refresh the
[upstream changelog](../UPSTREAM_CHANGELOG.md) with affected areas, functions or
symbols, and confirmed behavior or a clearly labeled source-update description.
Local and CI provenance checks require a populated entry for the pinned
revision. The [repository instructions](../AGENTS.md) describe the evidence and
review requirements; the [source-oracle guide](../tools/upstream-oracle/README.md)
explains how to rebuild the upstream reference and refresh parity attestations.
