package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func fixture(t *testing.T, parser string) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"ASIM/dev/ASimTester/ASimTester.csv": "ColumnName,ColumnType,Schema\nTimeGenerated,datetime,Common\nDnsQuery,string,Dns\nLatitude,double,Dns\n",
		"Parsers/ASimDns/Parsers/imDns.yaml": parser,
	}
	for name, contents := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

const parserFixture = `ParserName: imDns
EquivalentBuiltInParser: _Im_Dns
Normalization:
  Schema: Dns
ParserParams:
  - Name: required
    Type: string
  - Name: pack
    Type: bool
    Default: false
  - Name: starttime
    Type: datetime
    Default: null
`

func TestGenerateASIMCatalog(t *testing.T) {
	root := fixture(t, parserFixture)
	c, err := generate(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Functions) != 2 || !reflect.DeepEqual(c.Functions["imDns"], c.Functions["_Im_Dns"]) {
		t.Fatalf("parser aliases = %#v", c.Functions)
	}
	args := c.Functions["_Im_Dns"].Arguments
	if len(args) != 3 || args[0].Optional || !args[1].Optional || !args[2].Optional {
		t.Fatalf("required/false/null defaults = %#v", args)
	}
	if c.Schemas["Dns"]["TimeGenerated"] != "datetime" || c.Schemas["Dns"]["Latitude"] != "real" {
		t.Fatalf("normalized schema = %#v", c.Schemas)
	}
	first, _ := json.Marshal(c)
	secondCatalog, err := generate(root)
	if err != nil {
		t.Fatal(err)
	}
	second, _ := json.Marshal(secondCatalog)
	if string(first) != string(second) {
		t.Fatal("generation is not deterministic")
	}
}

func TestGenerateASIMCatalogRejectsUnknownSchemaAndDuplicates(t *testing.T) {
	root := fixture(t, strings.Replace(parserFixture, "Schema: Dns", "Schema: Unknown", 1))
	if _, err := generate(root); err == nil {
		t.Fatal("unknown schema should fail generation")
	}
	root = fixture(t, parserFixture)
	if err := os.WriteFile(filepath.Join(root, "Parsers/ASimDns/Parsers/duplicate.yaml"), []byte(parserFixture), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := generate(root); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate parser error = %v", err)
	}
}
