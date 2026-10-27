# Repository instructions

Only use subagents when the user explicitly requests them.

## Upstream compatibility updates

Every Microsoft Kusto.Language compatibility update must refresh
`UPSTREAM_CHANGELOG.md` in the same change as the implementation and upstream
pin. This applies to source synchronizations even when Microsoft does not
change its version number.

- Add a dated entry linking to the exact upstream commit. Use the upstream
  commit date; distinguish source revisions from package release dates.
- Keep the existing table format: **Area**, **Impacted functions or symbols**,
  and **Confirmed behavior or update**. Name the affected upstream functions,
  methods, or symbols, and the corresponding Go entry points when helpful.
- Describe confirmed behavior with concrete before/after examples. Support
  confirmation with source-oracle comparisons or regression tests and link
  relevant evidence. Describe unmeasured changes as **Source update**; do not
  present inferred behavior as confirmed.
- Account for each affected area, including tooling-only updates and changes
  that require no Go modification. Preserve historical entries and distinguish
  intentional Go differences from upstream behavior.
- Refresh parity attestations and provenance, run the applicable checks, and
  verify that the changelog entry matches the pinned revision before committing.

The existing local provenance check also requires a populated changelog table
in the section linking to the current upstream commit. It runs in both CI
workflows; its structural validation does not replace reviewing the evidence.
