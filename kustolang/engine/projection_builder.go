package kustolang

// Ported from src/Kusto.Language/Binder/ProjectionBuilder.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type ProjectionBuilder struct {
	projection         []*ColumnSymbol
	doNotAdd           map[*ColumnSymbol]bool
	columnIndexMap     map[string]int
	declaredNames      map[string]bool
	uniqueNames        *UniqueNameTable
	preserveErrorTypes bool
}

type ProjectionAddRangeOptions struct {
	Declare     bool
	DoNotRepeat bool
}

type ProjectionAddOptions struct {
	BaseName    string
	HasBaseName bool
	Replace     bool
	DoNotRepeat bool
}

func NewProjectionBuilder() *ProjectionBuilder {
	builder := &ProjectionBuilder{}
	builder.ensure()
	return builder
}

func (b *ProjectionBuilder) Clear() {
	if b == nil {
		return
	}
	b.projection = nil
	b.doNotAdd = map[*ColumnSymbol]bool{}
	b.columnIndexMap = map[string]int{}
	b.declaredNames = map[string]bool{}
	b.uniqueNames = NewUniqueNameTable()
}

func (b *ProjectionBuilder) GetProjection() []*ColumnSymbol {
	if b == nil {
		return nil
	}
	return ListToReadOnly(b.projection)
}

func (b *ProjectionBuilder) AddRange(columns []*ColumnSymbol, options ...ProjectionAddRangeOptions) {
	if b == nil {
		return
	}
	b.ensure()
	config := ProjectionAddRangeOptions{}
	if len(options) > 0 {
		config = options[0]
	}
	for _, column := range columns {
		added := b.Add(column, ProjectionAddOptions{DoNotRepeat: config.DoNotRepeat})
		if config.Declare && added != nil {
			b.declaredNames[added.SymbolName()] = true
		}
	}
}

func (b *ProjectionBuilder) DoNotAdd(column *ColumnSymbol) {
	if b == nil || column == nil {
		return
	}
	b.ensure()
	b.doNotAdd[column] = true
}

func (b *ProjectionBuilder) DoNotAddAny(columns []*ColumnSymbol) {
	for _, column := range columns {
		b.DoNotAdd(column)
	}
}

func (b *ProjectionBuilder) CanAdd(column *ColumnSymbol) bool {
	if b == nil || column == nil {
		return false
	}
	b.ensure()
	return !b.doNotAdd[column]
}

func (b *ProjectionBuilder) Add(column *ColumnSymbol, options ...ProjectionAddOptions) *ColumnSymbol {
	if b == nil || column == nil {
		return column
	}
	b.ensure()
	config := ProjectionAddOptions{}
	if len(options) > 0 {
		config = options[0]
	}

	if column.SymbolName() == "" {
		return column
	}

	if b.doNotAdd[column] {
		return column
	}

	if config.Replace {
		if index, ok := b.columnIndexMap[column.SymbolName()]; ok {
			b.projection[index] = column
			if config.DoNotRepeat {
				b.doNotAdd[column] = true
			}
			return column
		}
	}

	var uniqueName string
	if config.HasBaseName {
		uniqueName = b.uniqueNames.GetOrAddName(column.SymbolName(), config.BaseName)
	} else {
		uniqueName = b.uniqueNames.GetOrAddName(column.SymbolName())
	}
	if uniqueName != column.SymbolName() {
		if b.preserveErrorTypes && isErrorTypeSymbol(column.Type) {
			column = &ColumnSymbol{
				BaseSymbol:      NewBaseSymbol(uniqueName),
				Type:            column.Type,
				Description:     column.Description,
				OriginalColumns: trulyOriginalColumns([]*ColumnSymbol{column}),
			}
		} else {
			column = NewColumnSymbolFull(uniqueName, column.Type, column.Description, []*ColumnSymbol{column}, nil, nil)
		}
	}

	b.projection = append(b.projection, column)
	b.columnIndexMap[column.SymbolName()] = len(b.projection) - 1

	if config.DoNotRepeat {
		b.doNotAdd[column] = true
	}

	return column
}

func (b *ProjectionBuilder) Declare(column *ColumnSymbol, diagnostics *[]Diagnostic, location SyntaxElement, replace ...bool) {
	if b == nil || column == nil {
		return
	}
	b.ensure()
	allowReplace := len(replace) > 0 && replace[0]

	if b.declaredNames[column.SymbolName()] {
		appendProjectionDiagnostic(diagnostics, DiagnosticDuplicateColumnDeclaration(column.SymbolName()), location)
		return
	}

	if allowReplace {
		if index, ok := b.columnIndexMap[column.SymbolName()]; ok {
			b.projection[index] = column
			b.declaredNames[column.SymbolName()] = true
			return
		}
	}

	added := b.Add(column)
	if added != nil {
		b.declaredNames[added.SymbolName()] = true
	}
}

func (b *ProjectionBuilder) Rename(oldName string, newName string, diagnostics *[]Diagnostic, location SyntaxElement) *ColumnSymbol {
	if b == nil {
		return nil
	}
	b.ensure()
	index, ok := b.columnIndexMap[oldName]
	if !ok {
		return nil
	}

	if b.declaredNames[newName] {
		appendProjectionDiagnostic(diagnostics, DiagnosticDuplicateColumnDeclaration(newName), location)
		return nil
	}

	oldColumn := b.projection[index]
	newColumn := NewColumnSymbolFull(newName, oldColumn.Type, oldColumn.Description, []*ColumnSymbol{oldColumn}, nil, nil)

	b.projection[index] = newColumn
	delete(b.columnIndexMap, oldName)
	b.columnIndexMap[newName] = index
	b.declaredNames[newName] = true
	b.uniqueNames.AddName(newName)

	return newColumn
}

func (b *ProjectionBuilder) ensure() {
	if b.doNotAdd == nil {
		b.doNotAdd = map[*ColumnSymbol]bool{}
	}
	if b.columnIndexMap == nil {
		b.columnIndexMap = map[string]int{}
	}
	if b.declaredNames == nil {
		b.declaredNames = map[string]bool{}
	}
	if b.uniqueNames == nil {
		b.uniqueNames = NewUniqueNameTable()
	}
}

func appendProjectionDiagnostic(diagnostics *[]Diagnostic, diagnostic Diagnostic, location SyntaxElement) {
	if diagnostics == nil {
		return
	}
	if location != nil {
		diagnostic = diagnostic.WithByteLocation(elementTextStart(location, 0), location.Width())
	}
	*diagnostics = append(*diagnostics, diagnostic)
}
