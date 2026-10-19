package environment

import (
	"encoding/json"
	"testing"
)

func TestCombinedEnvironmentRetainsSharedTableColumns(t *testing.T) {
	envs := KQLEnvironments{
		"m365": {
			TableDetails: map[string]TableDetails{"DeviceProcessEvents": {
				"Timestamp": "datetime", "Shared": "string",
			}},
			TableOptions: map[string]TableOptions{"DeviceProcessEvents": {IsOpen: true}},
		},
		"sentinel": {
			TableDetails: map[string]TableDetails{"DeviceProcessEvents": {
				"TimeGenerated": "datetime", "Type": "string", "Shared": "long", "Payload": "dynamic",
			}},
			TableOptions: map[string]TableOptions{"DeviceProcessEvents": {ColumnTypes: map[string]ColumnTypeDetails{
				"Payload": {Type: "dynamic", Properties: map[string]ColumnTypeDetails{"Name": {Type: "string"}}},
			}}},
		},
	}
	before, _ := json.Marshal(envs)
	AddM365WithSentinelIfPresent(envs)
	merged := envs["m365_with_sentinel"]
	columns := merged.TableDetails["DeviceProcessEvents"]
	for column, want := range map[string]string{
		"Timestamp": "datetime", "TimeGenerated": "datetime", "Type": "string", "Shared": "string", "Payload": "dynamic",
	} {
		if columns[column] != want {
			t.Fatalf("column %s = %q, want %q", column, columns[column], want)
		}
	}
	options := merged.TableOptions["DeviceProcessEvents"]
	if !options.IsOpen || options.ColumnTypes["Payload"].Properties["Name"].Type != "string" {
		t.Fatalf("merged options = %#v", options)
	}
	delete(envs, "m365_with_sentinel")
	after, _ := json.Marshal(envs)
	if string(before) != string(after) {
		t.Fatal("merge mutated source environments")
	}
	columns["Type"] = "long"
	options.ColumnTypes["Payload"].Properties["Name"] = ColumnTypeDetails{Type: "long"}
	if envs["sentinel"].TableDetails["DeviceProcessEvents"]["Type"] != "string" ||
		envs["sentinel"].TableOptions["DeviceProcessEvents"].ColumnTypes["Payload"].Properties["Name"].Type != "string" {
		t.Fatal("combined schema shares mutable column data with source environment")
	}
}
