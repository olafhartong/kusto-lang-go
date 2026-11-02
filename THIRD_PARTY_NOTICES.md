# Third-Party Notices

kusto-lang-go contains a native Go source port derived from Microsoft Kusto Query
Language:

- Project: Microsoft Kusto Query Language
- Repository: https://github.com/microsoft/Kusto-Query-Language
- License: Apache License 2.0 (`Apache-2.0`)
- Current source revision: `82267c7628e70898cd8bec54c20fc4e61c0a523a`

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

