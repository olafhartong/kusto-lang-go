package kustolang

import "strings"

// Ported from compact symbol types under src/Kusto.Language/Symbols:
// ClusterSymbol.cs, EntityGroupSymbol.cs, EntityGroupElementSymbol.cs,
// GraphModelSymbol.cs, GroupSymbol.cs and PatternSymbol.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type ClusterSymbol struct {
	TypeSymbol
	members []Symbol
	isOpen  bool
}

func NewClusterSymbol(name string, databases ...*DatabaseSymbol) *ClusterSymbol {
	return NewClusterSymbolWithDatabases(name, databases, false)
}

func NewClusterSymbolWithDatabases(name string, databases []*DatabaseSymbol, isOpen bool) *ClusterSymbol {
	members := make([]Symbol, 0, len(databases))
	for _, database := range databases {
		if database != nil {
			members = append(members, database)
		}
	}
	return NewClusterSymbolWithMembers(name, members, isOpen)
}

func NewClusterSymbolWithMembers(name string, members []Symbol, isOpen bool) *ClusterSymbol {
	return &ClusterSymbol{
		TypeSymbol: TypeSymbol{BaseSymbol: NewBaseSymbol(name)},
		members:    append([]Symbol{}, members...),
		isOpen:     isOpen,
	}
}

func (s *ClusterSymbol) SymbolKind() SymbolKind {
	return SymbolKindCluster
}

func (s *ClusterSymbol) SymbolTabularity() Tabularity {
	return TabularityTabular
}

func (s *ClusterSymbol) Members() []Symbol {
	if s == nil {
		return nil
	}
	return append([]Symbol{}, s.members...)
}

func (s *ClusterSymbol) IsOpen() bool {
	return s != nil && s.isOpen
}

func (s *ClusterSymbol) Databases() []*DatabaseSymbol {
	if s == nil {
		return nil
	}
	var databases []*DatabaseSymbol
	for _, member := range s.members {
		if database, ok := member.(*DatabaseSymbol); ok {
			databases = append(databases, database)
		}
	}
	return databases
}

func (s *ClusterSymbol) GetDatabase(databaseName string) *DatabaseSymbol {
	if s == nil || databaseName == "" {
		return nil
	}
	databases := s.Databases()
	for i := len(databases) - 1; i >= 0; i-- {
		database := databases[i]
		if database.SymbolName() == databaseName || database.AlternateName() == databaseName {
			return database
		}
	}
	return nil
}

func (s *ClusterSymbol) AddMembers(members ...Symbol) *ClusterSymbol {
	if s == nil {
		return nil
	}
	newMembers := append(s.Members(), members...)
	return NewClusterSymbolWithMembers(s.SymbolName(), newMembers, s.isOpen)
}

func (s *ClusterSymbol) WithDatabases(databases []*DatabaseSymbol) *ClusterSymbol {
	if s == nil {
		return nil
	}
	return NewClusterSymbolWithDatabases(s.SymbolName(), databases, s.isOpen)
}

func (s *ClusterSymbol) AddDatabase(database *DatabaseSymbol) *ClusterSymbol {
	if database == nil {
		return s
	}
	return s.AddMembers(database)
}

func (s *ClusterSymbol) UpdateDatabase(existingDatabase *DatabaseSymbol, newDatabase *DatabaseSymbol) *ClusterSymbol {
	if s == nil {
		return nil
	}
	newMembers := s.Members()
	for i, member := range newMembers {
		if member == existingDatabase {
			newMembers[i] = newDatabase
		}
	}
	return NewClusterSymbolWithMembers(s.SymbolName(), newMembers, s.isOpen)
}

func (s *ClusterSymbol) AddOrUpdateDatabase(newDatabase *DatabaseSymbol) *ClusterSymbol {
	if s == nil || newDatabase == nil {
		return s
	}
	if existingDatabase := s.GetDatabase(newDatabase.SymbolName()); existingDatabase != nil {
		return s.UpdateDatabase(existingDatabase, newDatabase)
	}
	return s.AddDatabase(newDatabase)
}

func (s *ClusterSymbol) RemoveDatabase(symbolToRemove *DatabaseSymbol) *ClusterSymbol {
	if s == nil {
		return nil
	}
	newMembers := make([]Symbol, 0, len(s.members))
	for _, member := range s.members {
		if member != symbolToRemove {
			newMembers = append(newMembers, member)
		}
	}
	return NewClusterSymbolWithMembers(s.SymbolName(), newMembers, s.isOpen)
}

func (s *ClusterSymbol) RemoveDatabases(symbolsToRemove []*DatabaseSymbol) *ClusterSymbol {
	if s == nil {
		return nil
	}
	remove := map[*DatabaseSymbol]struct{}{}
	for _, database := range symbolsToRemove {
		if database != nil {
			remove[database] = struct{}{}
		}
	}
	newMembers := make([]Symbol, 0, len(s.members))
	for _, member := range s.members {
		if database, ok := member.(*DatabaseSymbol); ok {
			if _, found := remove[database]; found {
				continue
			}
		}
		newMembers = append(newMembers, member)
	}
	return NewClusterSymbolWithMembers(s.SymbolName(), newMembers, s.isOpen)
}

var UnknownClusterSymbol = NewClusterSymbolWithMembers("", nil, true)

type EntityGroupSymbol struct {
	TypeSymbol
	definition  string
	description string
	members     []Symbol
	signature   *Signature
}

func NewEntityGroupSymbol(name string, members ...Symbol) *EntityGroupSymbol {
	return NewEntityGroupSymbolWithMembers(name, members, "")
}

func NewEntityGroupSymbolWithMembers(name string, members []Symbol, description string) *EntityGroupSymbol {
	s := &EntityGroupSymbol{
		TypeSymbol:  TypeSymbol{BaseSymbol: NewBaseSymbol(name)},
		description: description,
		members:     append([]Symbol{}, members...),
	}
	s.signature = NewSymbolSignature(s)
	s.signature.Symbol = s
	return s
}

func NewEntityGroupSymbolFromDefinition(name string, definition string, description string) *EntityGroupSymbol {
	s := &EntityGroupSymbol{
		TypeSymbol:  TypeSymbol{BaseSymbol: NewBaseSymbol(name)},
		definition:  definition,
		description: description,
	}
	s.signature = NewComputedSignature(EntityGroupBodyFromDefinition(definition), TabularityTabular)
	s.signature.Symbol = s
	return s
}

func EntityGroupBodyFromDefinition(definition string) string {
	if definition == "" {
		return "entity_group []"
	}

	definition = strings.TrimSpace(definition)
	if strings.HasPrefix(definition, "entity_group") {
		return definition
	}

	expressionList := definition
	if strings.HasPrefix(definition, "[") && strings.HasSuffix(definition, "]") {
		expressionList = definition[1 : len(definition)-1]
	}
	expressionList = strings.TrimSpace(expressionList)
	if strings.HasPrefix(expressionList, `"`) || strings.HasPrefix(expressionList, `'`) {
		expressionList = GetStringLiteralValue(expressionList)
	}
	return "entity_group [" + expressionList + "]"
}

func (s *EntityGroupSymbol) SymbolKind() SymbolKind {
	return SymbolKindEntityGroup
}

func (s *EntityGroupSymbol) SymbolTabularity() Tabularity {
	return TabularityNone
}

func (s *EntityGroupSymbol) Members() []Symbol {
	if s == nil {
		return nil
	}
	return append([]Symbol{}, s.members...)
}

func (s *EntityGroupSymbol) Definition() string {
	if s == nil {
		return ""
	}
	return s.definition
}

func (s *EntityGroupSymbol) Description() string {
	if s == nil {
		return ""
	}
	return s.description
}

func (s *EntityGroupSymbol) Signature() *Signature {
	if s == nil {
		return nil
	}
	return s.signature
}

var EmptyEntityGroupSymbol = NewEntityGroupSymbol("")

type EntityGroupElementSymbol struct {
	TypeSymbol
	EntityGroup      *EntityGroupSymbol
	UnderlyingSymbol Symbol
}

func NewEntityGroupElementSymbol(name string, entityGroup *EntityGroupSymbol, underlyingSymbol Symbol) *EntityGroupElementSymbol {
	if entityGroup == nil {
		entityGroup = EmptyEntityGroupSymbol
	}
	if underlyingSymbol == nil {
		underlyingSymbol = firstEntityGroupMemberWithMembers(entityGroup)
	}
	if underlyingSymbol == nil {
		underlyingSymbol = firstEntityGroupMember(entityGroup)
	}
	if underlyingSymbol == nil {
		underlyingSymbol = Error
	}
	return &EntityGroupElementSymbol{
		TypeSymbol:       TypeSymbol{BaseSymbol: NewBaseSymbol(name)},
		EntityGroup:      entityGroup,
		UnderlyingSymbol: underlyingSymbol,
	}
}

func (s *EntityGroupElementSymbol) SymbolKind() SymbolKind {
	return SymbolKindEntityGroupElement
}

func (s *EntityGroupElementSymbol) SymbolTabularity() Tabularity {
	if s == nil || s.UnderlyingSymbol == nil {
		return TabularityUnknown
	}
	return s.UnderlyingSymbol.SymbolTabularity()
}

func (s *EntityGroupElementSymbol) Members() []Symbol {
	return EntityGroupElementSpecialMembers()
}

var entityGroupElementSpecialMembers = []Symbol{
	NewVariableSymbol("$current_database", ScalarTypes.String.Type(), false, nil, nil),
	NewVariableSymbol("$current_cluster_endpoint", ScalarTypes.String.Type(), false, nil, nil),
}

func EntityGroupElementSpecialMembers() []Symbol {
	return append([]Symbol{}, entityGroupElementSpecialMembers...)
}

func firstEntityGroupMemberWithMembers(entityGroup *EntityGroupSymbol) Symbol {
	for _, member := range entityGroup.Members() {
		if isEntityGroupMemberType(member) && len(member.Members()) > 0 {
			return member
		}
	}
	return nil
}

func firstEntityGroupMember(entityGroup *EntityGroupSymbol) Symbol {
	for _, member := range entityGroup.Members() {
		if isEntityGroupMemberType(member) {
			return member
		}
	}
	return nil
}

func isEntityGroupMemberType(symbol Symbol) bool {
	switch symbol.(type) {
	case *TypeSymbol,
		*ScalarSymbol,
		*TableSymbol,
		*MaterializedViewSymbol,
		*ExternalTableSymbol,
		*StoredQueryResultSymbol,
		*DatabaseSymbol,
		*ClusterSymbol,
		*GraphSymbol,
		*GraphModelSymbol,
		*EntityGroupSymbol,
		*EntityGroupElementSymbol,
		*ErrorSymbol,
		*VoidSymbol:
		return true
	default:
		return false
	}
}

type GraphModelSymbol struct {
	TypeSymbol
	Edges               []*Signature
	Nodes               []*Signature
	Snapshots           []*GraphSnapshotSymbol
	ComputedGraphSymbol *GraphSymbol
}

func NewGraphModelSymbol(name string, edges []*Signature, nodes []*Signature, snapshots []*GraphSnapshotSymbol) *GraphModelSymbol {
	s := &GraphModelSymbol{
		TypeSymbol: TypeSymbol{BaseSymbol: NewBaseSymbol(name)},
		Edges:      append([]*Signature{}, edges...),
		Nodes:      append([]*Signature{}, nodes...),
		Snapshots:  append([]*GraphSnapshotSymbol{}, snapshots...),
	}
	for _, edge := range s.Edges {
		if edge != nil {
			edge.Symbol = s
		}
	}
	for _, node := range s.Nodes {
		if node != nil {
			node.Symbol = s
		}
	}
	for _, snapshot := range s.Snapshots {
		if snapshot != nil {
			snapshot.Model = s
		}
	}
	return s
}

func NewGraphModelSymbolFromNames(name string, edges []string, nodes []string, snapshots []string) *GraphModelSymbol {
	edgeSignatures := make([]*Signature, 0, len(edges))
	for _, edge := range edges {
		edgeSignatures = append(edgeSignatures, NewComputedSignature(edge, TabularityUnknown))
	}
	nodeSignatures := make([]*Signature, 0, len(nodes))
	for _, node := range nodes {
		nodeSignatures = append(nodeSignatures, NewComputedSignature(node, TabularityUnknown))
	}
	snapshotSymbols := make([]*GraphSnapshotSymbol, 0, len(snapshots))
	for _, snapshot := range snapshots {
		snapshotSymbols = append(snapshotSymbols, NewGraphSnapshotSymbol(snapshot))
	}
	return NewGraphModelSymbol(name, edgeSignatures, nodeSignatures, snapshotSymbols)
}

func (s *GraphModelSymbol) SymbolKind() SymbolKind {
	return SymbolKindGraphModel
}

func (s *GraphModelSymbol) SymbolTabularity() Tabularity {
	return TabularityNone
}

func (s *GraphModelSymbol) Members() []Symbol {
	if s == nil {
		return nil
	}
	out := make([]Symbol, 0, len(s.Snapshots))
	for _, snapshot := range s.Snapshots {
		if snapshot != nil {
			out = append(out, snapshot)
		}
	}
	return out
}

func (s *GraphModelSymbol) TryGetSnapshot(name string) (*GraphSnapshotSymbol, bool) {
	if s == nil {
		return nil, false
	}
	for i := len(s.Snapshots) - 1; i >= 0; i-- {
		snapshot := s.Snapshots[i]
		if snapshot != nil && snapshot.SymbolName() == name {
			return snapshot, true
		}
	}
	return nil, false
}

type GraphSnapshotSymbol struct {
	BaseSymbol
	Model *GraphModelSymbol
}

func NewGraphSnapshotSymbol(name string) *GraphSnapshotSymbol {
	return &GraphSnapshotSymbol{BaseSymbol: NewBaseSymbol(name)}
}

func (s *GraphSnapshotSymbol) SymbolKind() SymbolKind {
	return SymbolKindGraphSnapshot
}

func (s *GraphSnapshotSymbol) SymbolTabularity() Tabularity {
	return TabularityNone
}

type GroupSymbol struct {
	TypeSymbol
	members []Symbol
}

func NewGroupSymbol(symbols ...Symbol) *GroupSymbol {
	return &GroupSymbol{
		TypeSymbol: TypeSymbol{BaseSymbol: NewBaseSymbol("group")},
		members:    append([]Symbol{}, symbols...),
	}
}

func (s *GroupSymbol) SymbolKind() SymbolKind {
	return SymbolKindGroup
}

func (s *GroupSymbol) Members() []Symbol {
	if s == nil {
		return nil
	}
	return append([]Symbol{}, s.members...)
}

func (s *GroupSymbol) SymbolTabularity() Tabularity {
	if s == nil || len(s.members) == 0 || s.members[0] == nil {
		return TabularityUnknown
	}
	return s.members[0].SymbolTabularity()
}

type PatternSymbol struct {
	TypeSymbol
	parameters    []*Parameter
	pathParameter *Parameter
	signatures    []*PatternSignature
}

func NewPatternSymbol(name string, parameters []*Parameter, pathParameter *Parameter, signatures []*PatternSignature) *PatternSymbol {
	s := &PatternSymbol{
		TypeSymbol:    TypeSymbol{BaseSymbol: NewBaseSymbol(name)},
		parameters:    append([]*Parameter{}, parameters...),
		pathParameter: pathParameter,
		signatures:    append([]*PatternSignature{}, signatures...),
	}
	for _, signature := range s.signatures {
		if signature != nil {
			signature.Symbol = s
		}
	}
	return s
}

func NewNamedPatternSymbol(name string) *PatternSymbol {
	return NewPatternSymbol(name, nil, nil, nil)
}

func (s *PatternSymbol) SymbolKind() SymbolKind {
	return SymbolKindPattern
}

func (s *PatternSymbol) SymbolTabularity() Tabularity {
	return TabularityTabular
}

func (s *PatternSymbol) Parameters() []*Parameter {
	if s == nil {
		return nil
	}
	return append([]*Parameter{}, s.parameters...)
}

func (s *PatternSymbol) PathParameter() *Parameter {
	if s == nil {
		return nil
	}
	return s.pathParameter
}

func (s *PatternSymbol) Signatures() []*PatternSignature {
	if s == nil {
		return nil
	}
	return append([]*PatternSignature{}, s.signatures...)
}

type PatternSignature struct {
	Symbol         *PatternSymbol
	ArgumentValues []string
	PathValue      string
	hasPathValue   bool
	body           *string
	declaration    *FunctionBody
	signature      *Signature
}

func NewPatternSignature(argumentValues []string, pathValue string, body string, hasPathValue ...bool) *PatternSignature {
	pathPresent := pathValue != ""
	if len(hasPathValue) > 0 {
		pathPresent = hasPathValue[0]
	}
	return &PatternSignature{
		ArgumentValues: append([]string{}, argumentValues...),
		PathValue:      pathValue,
		hasPathValue:   pathPresent,
		body:           &body,
	}
}

func NewPatternSignatureWithFunctionBody(argumentValues []string, pathValue string, body *FunctionBody, hasPathValue ...bool) *PatternSignature {
	pathPresent := pathValue != ""
	if len(hasPathValue) > 0 {
		pathPresent = hasPathValue[0]
	}
	return &PatternSignature{
		ArgumentValues: append([]string{}, argumentValues...),
		PathValue:      pathValue,
		hasPathValue:   pathPresent,
		declaration:    body,
	}
}

func NewPatternSignatureFromFunctionBody(argumentValues []string, pathValue string, body *FunctionBody, hasPathValue ...bool) *PatternSignature {
	return NewPatternSignatureWithFunctionBody(argumentValues, pathValue, body, hasPathValue...)
}

func (s *PatternSignature) HasPathValue() bool {
	return s != nil && s.hasPathValue
}

func (s *PatternSignature) BodyText() string {
	if s == nil || s.body == nil {
		return ""
	}
	return *s.body
}

func (s *PatternSignature) Signature() *Signature {
	if s == nil || s.Symbol == nil {
		return nil
	}
	if s.signature == nil {
		parameters := s.Symbol.Parameters()
		if pathParameter := s.Symbol.PathParameter(); pathParameter != nil {
			parameters = append(parameters, pathParameter)
		}
		if s.declaration != nil {
			s.signature = NewDeclarationSignature(s.declaration, parameters...)
		} else if s.body != nil {
			s.signature = NewComputedSignature(*s.body, TabularityUnspecified, parameters...)
		} else {
			return nil
		}
		s.signature.Symbol = s.Symbol
	}
	return s.signature
}
