# Security policy

Report vulnerabilities privately through
[GitHub private vulnerability reporting](https://github.com/olafhartong/kusto-lang-go/security/advisories/new).
Include the library revision, a reproducing query/schema, and the embedding API.
Only the latest release receives security fixes.

The library processes caller-supplied queries and schemas in the embedding
process. Use cancellation and configurable analysis limits when processing
untrusted input. Parser panics, cancellation failures, memory exhaustion, and
incorrect editor edit boundaries are in scope. KustoKITT HTTP deployment and
schema-fetch behavior belong to that project's security policy.
