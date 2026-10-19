package kustolang

// Ported from src/Kusto.Language/Symbols/TableState.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type TableState uint8

const (
	TableStateNone       TableState = 0
	TableStateSerialized TableState = 1 << 0
	TableStateSorted     TableState = 1 << 1
	TableStateOpen       TableState = 1 << 2
)

func (s TableState) Has(flag TableState) bool {
	return (s & flag) != 0
}

func (s TableState) With(flag TableState) TableState {
	return s | flag
}

func (s TableState) Without(flag TableState) TableState {
	return s &^ flag
}

func (s TableState) New() TableState {
	return s.With(TableStateSerialized)
}

func (s TableState) Open() TableState {
	return s.With(TableStateOpen)
}

func (s TableState) Unsorted() TableState {
	return s.Without(TableStateSorted)
}
