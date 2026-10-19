package kustolang

import "strconv"

// Ported from src/Kusto.Language/Utils/UniqueNameTable.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

const maxInt32 = int(1<<31 - 1)

// UniqueNameTable maintains a set of unique names and allocates suffixed names
// for repeated candidates.
type UniqueNameTable struct {
	nameMap map[string]int
}

func NewUniqueNameTable() *UniqueNameTable {
	return &UniqueNameTable{nameMap: map[string]int{}}
}

// AddNames adds names already known to be unique to the table.
func (t *UniqueNameTable) AddNames(names []string) {
	for _, name := range names {
		t.AddName(name)
	}
}

func (t *UniqueNameTable) AddName(name string) {
	t.ensureMap()
	if name == "" {
		return
	}
	if _, ok := t.nameMap[name]; !ok {
		t.nameMap[name] = 0
	}
}

// GetOrAddName adds name to the table if it is unique, otherwise creates a new
// unique name. Omitting baseName matches reference implementation's null optional value; passing ""
// explicitly uses an empty suffix prefix, just as baseName: "" does upstream.
func (t *UniqueNameTable) GetOrAddName(name string, baseName ...string) string {
	t.ensureMap()
	if lastUsedSuffix, ok := t.nameMap[name]; ok {
		prefix := name
		if len(baseName) > 0 {
			prefix = baseName[0]
		}
		for suffix := lastUsedSuffix + 1; suffix < maxInt32; suffix++ {
			newName := prefix + strconv.Itoa(suffix)
			if _, exists := t.nameMap[newName]; !exists {
				t.nameMap[name] = suffix
				t.nameMap[newName] = 0
				return newName
			}
		}
	} else {
		t.nameMap[name] = 0
	}
	return name
}

func (t *UniqueNameTable) Clear() {
	t.ensureMap()
	clear(t.nameMap)
}

func (t *UniqueNameTable) ensureMap() {
	if t.nameMap == nil {
		t.nameMap = map[string]int{}
	}
}
