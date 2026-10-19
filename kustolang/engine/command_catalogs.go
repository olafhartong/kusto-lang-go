package kustolang

// Ported from the generated command symbol catalogs:
//   src/Kusto.Language/Parser/CodeGen/AriaBridgeCommands.cs
//   src/Kusto.Language/Parser/CodeGen/ClusterManagerCommands.cs
//   src/Kusto.Language/Parser/CodeGen/DataManagerCommands.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type commandSet struct {
	ShowVersion *CommandSymbol
	All         []*CommandSymbol
}

var (
	AriaBridgeCommands     = newShowVersionCommandSet()
	ClusterManagerCommands = newShowVersionCommandSet()
	DataManagerCommands    = newShowVersionCommandSet()
)

func newShowVersionCommandSet() *commandSet {
	c := &commandSet{}
	c.ShowVersion = NewCommandSymbol("ShowVersion", "(BuildVersion: string, BuildTime: datetime, ServiceType: string, ProductVersion: string)")
	c.All = []*CommandSymbol{
		c.ShowVersion,
	}
	return c
}
