# kusto-lang-go

An embeddable native Go implementation of Microsoft Kusto Query Language (KQL),
with parsing, semantic analysis, and editor services. No .NET runtime or
running service is required.

## Highlights

- **Understand KQL:** typed syntax trees, symbol binding, result schemas, and
  recoverable diagnostics for incomplete or malformed queries.
- **Build editor features:** completion, quick info, formatting, highlighting,
  outlines, code actions, and multi-block documents.
- **Bring your schema:** tables, functions, databases, clusters, and offline
  Microsoft Sentinel and ASIM function metadata.
- **Track upstream behavior:** a native source port backed by regression tests,
  differential parity evidence, and provenance checks.

## Quick start

```sh
go get github.com/olafhartong/kusto-lang-go
```

```go
package main

import (
	"fmt"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func main() {
	code := kustolang.ParseAndAnalyze(`print Answer = 6 * 7`)
	fmt.Println(code.GetDiagnostics())
}
```

## Documentation

- [Getting started and schema loading](docs/using-kustolang.md)
- [Language APIs, symbols, grammars, and coordinate conventions](docs/language-api.md)
- [Editor services and multi-block documents](docs/editor-services.md)
- [Sentinel functions, watchlists, and ASIM metadata](docs/sentinel-functions.md)
- [Development, validation, licensing, and upstream tracking](docs/development.md)
- [Upstream parity and remaining differences](docs/upstream-parity.md)
- [Microsoft upstream changelog](UPSTREAM_CHANGELOG.md)

Licensed under [GPL-3.0](LICENSE), with preserved Apache-2.0 and MIT
[third-party notices](THIRD_PARTY_NOTICES.md).
