# Third-Party Notices

kusto-lang-go contains a native Go source port derived from Microsoft Kusto Query
Language:

- Project: Microsoft Kusto Query Language
- Repository: https://github.com/microsoft/Kusto-Query-Language
- License: Apache License 2.0 (`Apache-2.0`)
- Current source revision: `6b5c2f739a14bd5de52115f9bf9695985f0852dd`

The preserved license text is distributed at `third_party/kusto/LICENSE`. Source
revision provenance is recorded in
`third_party/kusto/UPSTREAM.json`.

The offline Sentinel function catalog includes ASIM parser signature and schema
metadata derived from Microsoft's Azure Sentinel repository:

- Project: Azure Sentinel
- Repository: https://github.com/Azure/Azure-Sentinel
- License: MIT (`MIT`)
- Source revision: recorded in `kustolang/environment/sentinel_catalog.json`
- Preserved license: `third_party/sentinel/LICENSE`

The catalog contains function metadata and normalized schemas; parser bodies
are not redistributed.

