package environment

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestDefenderTablesMatchBundledCatalog(t *testing.T) {
	data, err := os.ReadFile("testdata/defender_tables.json")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	if err := json.Unmarshal(data, &names); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{}
	for _, name := range names {
		want[strings.ToLower(name)] = true
	}
	if !reflect.DeepEqual(defenderTables, want) {
		t.Fatalf("Defender compatibility table names must match bundled m365 catalog: got %v, want %v", defenderTables, want)
	}
}
