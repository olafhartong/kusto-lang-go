package kustolang

// Ported from src/Kusto.Language/Binder/ColumnMap.cs.
// ColumnMap is an internal binder helper that tracks columns by name and,
// when needed, by type while preserving upstream duplicate-column behavior.

type ColumnMap struct {
	nameMap           map[string]any
	unifyDynamicTypes bool
}

func NewColumnMap(unifyDynamicTypes ...bool) *ColumnMap {
	unify := true
	if len(unifyDynamicTypes) > 0 {
		unify = unifyDynamicTypes[0]
	}
	return &ColumnMap{
		nameMap:           map[string]any{},
		unifyDynamicTypes: unify,
	}
}

func NewColumnMapFromColumns(columns []*ColumnSymbol, unifyDynamicTypes ...bool) *ColumnMap {
	m := NewColumnMap(unifyDynamicTypes...)
	m.AddRange(columns)
	return m
}

func (m *ColumnMap) AddRange(columns []*ColumnSymbol) {
	for _, column := range columns {
		m.Add(column)
	}
}

func (m *ColumnMap) typeKey(typ *TypeSymbol) *TypeSymbol {
	if m != nil && m.unifyDynamicTypes && isColumnMapDynamicType(typ) {
		return ScalarTypes.Dynamic.Type()
	}
	return typ
}

func isColumnMapDynamicType(typ *TypeSymbol) bool {
	return typ != nil && typ.SymbolName() == ScalarTypes.Dynamic.SymbolName()
}

func (m *ColumnMap) Add(column *ColumnSymbol) {
	if m.nameMap == nil {
		m.nameMap = map[string]any{}
	}
	name := column.SymbolName()
	existing, ok := m.nameMap[name]
	if !ok {
		m.nameMap[name] = column
		return
	}

	switch value := existing.(type) {
	case *ColumnSymbol:
		existingKey := m.typeKey(value.Type)
		columnKey := m.typeKey(column.Type)
		if existingKey == columnKey {
			m.nameMap[name] = []*ColumnSymbol{value, column}
		} else {
			byType := newColumnMapByType()
			byType.set(existingKey, value)
			byType.set(columnKey, column)
			m.nameMap[name] = byType
		}
	case []*ColumnSymbol:
		listTypeKey := m.typeKey(value[0].Type)
		columnTypeKey := m.typeKey(column.Type)
		if listTypeKey == columnTypeKey {
			value = append(value, column)
			m.nameMap[name] = value
		} else {
			byType := newColumnMapByType()
			byType.set(listTypeKey, value)
			byType.set(columnTypeKey, column)
			// Upstream constructs this dictionary but does not publish it.
		}
	case *columnMapByType:
		columnKey := m.typeKey(column.Type)
		if columnOrList, ok := value.get(columnKey); ok {
			switch columns := columnOrList.(type) {
			case *ColumnSymbol:
				value.set(columnKey, []*ColumnSymbol{columns, column})
			case []*ColumnSymbol:
				columns = append(columns, column)
				value.set(columnKey, columns)
			}
		} else {
			value.set(columnKey, column)
		}
	}
}

func (m *ColumnMap) Remove(name string) {
	if m == nil || m.nameMap == nil {
		return
	}
	delete(m.nameMap, name)
}

func (m *ColumnMap) HasColumns(name string) bool {
	if m == nil || m.nameMap == nil {
		return false
	}
	_, ok := m.nameMap[name]
	return ok
}

func (m *ColumnMap) HasMultipleTypes(name string) bool {
	if m == nil || m.nameMap == nil {
		return false
	}
	byType, ok := m.nameMap[name].(*columnMapByType)
	return ok && byType.len() > 1
}

func (m *ColumnMap) GetTypes(name string) []*TypeSymbol {
	if m == nil || m.nameMap == nil {
		return nil
	}
	switch value := m.nameMap[name].(type) {
	case *columnMapByType:
		return value.keys()
	case []*ColumnSymbol:
		if len(value) == 0 {
			return nil
		}
		return []*TypeSymbol{value[0].Type}
	case *ColumnSymbol:
		return []*TypeSymbol{value.Type}
	default:
		return nil
	}
}

func (m *ColumnMap) HasMutipleColumns(name string) bool {
	if m == nil || m.nameMap == nil {
		return false
	}
	switch value := m.nameMap[name].(type) {
	case *ColumnSymbol:
		return false
	case []*ColumnSymbol:
		return len(value) > 1
	case *columnMapByType:
		if value.len() > 1 {
			return true
		}
		if value.len() == 1 {
			switch columns := value.firstValue().(type) {
			case *ColumnSymbol:
				return true
			case []*ColumnSymbol:
				return len(columns) > 1
			}
		}
	}
	return false
}

func (m *ColumnMap) HasMultipleColumns(name string, typ *TypeSymbol) bool {
	if m == nil || m.nameMap == nil {
		return false
	}
	typeKey := m.typeKey(typ)
	switch value := m.nameMap[name].(type) {
	case *columnMapByType:
		columnOrList, ok := value.get(typeKey)
		columns, isList := columnOrList.([]*ColumnSymbol)
		return ok && isList && len(columns) > 0
	case []*ColumnSymbol:
		return len(value) > 0 && m.typeKey(value[0].Type) == typeKey && len(value) > 1
	default:
		return false
	}
}

func (m *ColumnMap) GetColumns(name string) []*ColumnSymbol {
	if m == nil || m.nameMap == nil {
		return nil
	}
	switch value := m.nameMap[name].(type) {
	case *columnMapByType:
		var columns []*ColumnSymbol
		for _, columnOrList := range value.valuesInOrder() {
			switch cols := columnOrList.(type) {
			case []*ColumnSymbol:
				columns = append(columns, cols...)
			case *ColumnSymbol:
				columns = append(columns, cols)
			}
		}
		return columns
	case []*ColumnSymbol:
		return append([]*ColumnSymbol{}, value...)
	case *ColumnSymbol:
		return []*ColumnSymbol{value}
	default:
		return nil
	}
}

func (m *ColumnMap) GetColumnsByType(name string, typ *TypeSymbol) []*ColumnSymbol {
	if m == nil || m.nameMap == nil {
		return nil
	}
	typeKey := m.typeKey(typ)
	switch value := m.nameMap[name].(type) {
	case *columnMapByType:
		if columnOrList, ok := value.get(typeKey); ok {
			switch cols := columnOrList.(type) {
			case []*ColumnSymbol:
				return append([]*ColumnSymbol{}, cols...)
			case *ColumnSymbol:
				return []*ColumnSymbol{cols}
			}
		}
	case []*ColumnSymbol:
		return append([]*ColumnSymbol{}, value...)
	case *ColumnSymbol:
		return []*ColumnSymbol{value}
	}
	return nil
}

type columnMapByType struct {
	order  []*TypeSymbol
	values map[*TypeSymbol]any
}

func newColumnMapByType() *columnMapByType {
	return &columnMapByType{values: map[*TypeSymbol]any{}}
}

func (m *columnMapByType) set(key *TypeSymbol, value any) {
	if _, ok := m.values[key]; !ok {
		m.order = append(m.order, key)
	}
	m.values[key] = value
}

func (m *columnMapByType) get(key *TypeSymbol) (any, bool) {
	value, ok := m.values[key]
	return value, ok
}

func (m *columnMapByType) len() int {
	if m == nil {
		return 0
	}
	return len(m.values)
}

func (m *columnMapByType) keys() []*TypeSymbol {
	if m == nil {
		return nil
	}
	return append([]*TypeSymbol{}, m.order...)
}

func (m *columnMapByType) valuesInOrder() []any {
	if m == nil {
		return nil
	}
	values := make([]any, 0, len(m.order))
	for _, key := range m.order {
		values = append(values, m.values[key])
	}
	return values
}

func (m *columnMapByType) firstValue() any {
	if m == nil || len(m.order) == 0 {
		return nil
	}
	return m.values[m.order[0]]
}
