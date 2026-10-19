package kustolang

import "strings"

// Ported from src/Kusto.Language/Syntax/SyntaxNode_Semantics.cs. This file
// exposes semantic-info accessors over the Go syntax payload model, including
// relocation of diagnostics produced by detached alternate syntax trees.

const semanticInfoPayloadName = "SemanticInfo"

type DiagnosticsInclude uint8

const (
	DiagnosticsIncludeSyntactic DiagnosticsInclude = 1 << iota
	DiagnosticsIncludeSemantic
	DiagnosticsIncludeExpansion
)

const DiagnosticsIncludeDefault = DiagnosticsIncludeSyntactic | DiagnosticsIncludeSemantic

func (n *SyntaxNode) SetSemanticInfo(info SemanticInfo) {
	if n == nil {
		return
	}
	n.setPayload(semanticInfoPayloadName, info)
}

func (n *SyntaxNode) GetSemanticInfo() (SemanticInfo, bool) {
	if n == nil {
		return SemanticInfo{}, false
	}
	value, ok := n.Payload(semanticInfoPayloadName)
	if !ok {
		return SemanticInfo{}, false
	}
	switch info := value.(type) {
	case SemanticInfo:
		return info, true
	case *SemanticInfo:
		if info != nil {
			return *info, true
		}
	}
	return SemanticInfo{}, false
}

func (n *SyntaxNode) ReferencedSymbol() Symbol {
	if info, ok := n.GetSemanticInfo(); ok {
		return info.ReferencedSymbol()
	}
	return nil
}

func (n *SyntaxNode) ReferencedSignature() *Signature {
	if info, ok := n.GetSemanticInfo(); ok {
		return info.ReferencedSignature()
	}
	return nil
}

func (n *SyntaxNode) GetCalledFunctionBody() *SyntaxNode {
	if info, ok := n.GetSemanticInfo(); ok && info.CalledFunctionInfo != nil {
		if expansion := info.CalledFunctionInfo.Expansion(); expansion != nil {
			return expansion.Root
		}
	}
	return nil
}

func (n *SyntaxNode) GetCalledFunctionFacts() *FunctionBodyFacts {
	if info, ok := n.GetSemanticInfo(); ok && info.CalledFunctionInfo != nil {
		return info.CalledFunctionInfo.Facts
	}
	return nil
}

func (n *SyntaxNode) GetCalledFunctionDiagnostics() []Diagnostic {
	if info, ok := n.GetSemanticInfo(); ok && info.CalledFunctionInfo != nil {
		return info.CalledFunctionInfo.Diagnostics()
	}
	return []Diagnostic{}
}

func (n *SyntaxNode) CalledFunctionHasErrors() bool {
	if info, ok := n.GetSemanticInfo(); ok && info.CalledFunctionInfo != nil {
		return info.CalledFunctionInfo.HasErrors()
	}
	return false
}

func (n *SyntaxNode) Alternates() []*SyntaxNode {
	if info, ok := n.GetSemanticInfo(); ok {
		return append([]*SyntaxNode{}, info.Alternates...)
	}
	return nil
}

func (n *SyntaxNode) SemanticDiagnostics() []Diagnostic {
	if info, ok := n.GetSemanticInfo(); ok {
		return append([]Diagnostic{}, info.Diagnostics...)
	}
	return []Diagnostic{}
}

func (n *SyntaxNode) IsBound() bool {
	_, ok := n.GetSemanticInfo()
	return ok
}

func (n *SyntaxNode) RawResultType() *TypeSymbol {
	if info, ok := n.GetSemanticInfo(); ok {
		return info.ResultType
	}
	return nil
}

func (n *SyntaxNode) ResultType() *TypeSymbol {
	return ExpressionResultType(generatedExpression(n))
}

func (n *SyntaxNode) IsConstant() bool {
	if info, ok := n.GetSemanticInfo(); ok {
		return info.IsConstant
	}
	return false
}

func (n *SyntaxNode) ConstantValueInfo() *ValueInfo {
	return ExpressionConstantValueInfo(generatedExpression(n))
}

func (n *SyntaxNode) ConstantValue() any {
	if info := n.ConstantValueInfo(); info != nil {
		return info.Value
	}
	return nil
}

func (n *SyntaxNode) IsLiteral() bool {
	return ExpressionIsLiteral(generatedExpression(n))
}

func (n *SyntaxNode) LiteralValueInfo() *ValueInfo {
	return ExpressionLiteralValueInfo(generatedExpression(n))
}

func (n *SyntaxNode) LiteralValue() any {
	if info := n.LiteralValueInfo(); info != nil {
		return info.Value
	}
	return nil
}

func generatedExpression(node *SyntaxNode) Expression {
	if node == nil {
		return nil
	}
	expression, _ := WrapGeneratedSyntaxNode(node).(Expression)
	return expression
}

// SyntaxList is a SyntaxNode in Kusto.Language. The Go syntax model keeps
// lists separate, so expose the semantic subset needed by list-owned
// alternates directly on SyntaxList.
func (l *SyntaxList) SetSemanticInfo(info SemanticInfo) {
	if l == nil {
		return
	}
	copy := info
	l.semanticInfo = &copy
}

func (l *SyntaxList) GetSemanticInfo() (SemanticInfo, bool) {
	if l == nil || l.semanticInfo == nil {
		return SemanticInfo{}, false
	}
	return *l.semanticInfo, true
}

func (l *SyntaxList) Alternates() []*SyntaxNode {
	if info, ok := l.GetSemanticInfo(); ok {
		return append([]*SyntaxNode{}, info.Alternates...)
	}
	return nil
}

func (l *SyntaxList) SemanticDiagnostics() []Diagnostic {
	if info, ok := l.GetSemanticInfo(); ok {
		return append([]Diagnostic{}, info.Diagnostics...)
	}
	return []Diagnostic{}
}

func (l *SyntaxList) IsBound() bool {
	_, ok := l.GetSemanticInfo()
	return ok
}

func ExpressionRawResultType(expression Expression) *TypeSymbol {
	if accessor, ok := expression.(interface{ RawResultType() *TypeSymbol }); ok {
		return accessor.RawResultType()
	}
	return nil
}

func ExpressionResultType(expression Expression) *TypeSymbol {
	rawType := ExpressionRawResultType(expression)
	if tuple := tupleResultSymbol(expression); tuple != nil && tuple.IsReducibleToScalar() {
		columns := tuple.Columns()
		if len(columns) == 1 && columns[0] != nil {
			return columns[0].Type
		}
	}
	if element := entityGroupElementResultSymbol(expression); element != nil {
		return returnTypeFromSymbol(element.UnderlyingSymbol)
	}
	return rawType
}

func ExpressionIsConstant(expression Expression) bool {
	if accessor, ok := expression.(interface{ IsConstant() bool }); ok {
		return accessor.IsConstant()
	}
	return false
}

func ExpressionIsLiteral(expression Expression) bool {
	return ExpressionLiteralValueInfo(expression) != nil || expressionIsValuelessLiteral(expression)
}

func ExpressionLiteralValueInfo(expression Expression) *ValueInfo {
	switch expr := expression.(type) {
	case *LiteralExpression:
		token := expr.Token()
		if token == nil {
			return nil
		}
		return NewValueInfo(token.RawText(), token.ValueText(), token.Value())
	case *CompoundStringLiteralExpression:
		if expr.Tokens() == nil {
			return NewValueInfo(expr.Text(IncludeTriviaMinimal), "", "")
		}
		var valueText strings.Builder
		for _, token := range expr.Tokens().Elements() {
			if token != nil {
				valueText.WriteString(token.ValueText())
			}
		}
		value := valueText.String()
		return NewValueInfo(expr.Text(IncludeTriviaMinimal), value, value)
	case *DynamicExpression:
		inner := ""
		var value any
		if expr.Expression() != nil {
			inner = expr.Expression().Text(IncludeTriviaMinimal)
			if expr.Expression().Kind() != SyntaxKindNullLiteralExpression {
				value = inner
			}
		}
		return NewValueInfo(expr.Text(IncludeTriviaMinimal), inner, value)
	default:
		return nil
	}
}

func expressionIsValuelessLiteral(expression Expression) bool {
	_, ok := expression.(*TypeOfLiteralExpression)
	return ok
}

func ExpressionLiteralValue(expression Expression) any {
	if info := ExpressionLiteralValueInfo(expression); info != nil {
		return info.Value
	}
	return nil
}

func ExpressionConstantValueInfo(expression Expression) *ValueInfo {
	if ExpressionIsLiteral(expression) {
		return ExpressionLiteralValueInfo(expression)
	}
	if accessor, ok := expression.(interface{ ReferencedSymbol() Symbol }); ok {
		if variable, ok := accessor.ReferencedSymbol().(*VariableSymbol); ok && variable.IsConstant {
			return variable.ConstantValueInfo
		}
	}
	return nil
}

func ExpressionConstantValue(expression Expression) any {
	if value := ExpressionConstantValueInfo(expression); value != nil {
		return value.Value
	}
	return nil
}

func GetContainedDiagnostics(root SyntaxElement, include DiagnosticsInclude) []Diagnostic {
	return GetContainedDiagnosticsWithCancellation(root, include, CancellationToken{})
}

func GetContainedDiagnosticsWithCancellation(root SyntaxElement, include DiagnosticsInclude, cancellationToken CancellationToken) []Diagnostic {
	return getContainedDiagnostics(root, include, true, cancellationToken)
}

func GetContainedDiagnosticsPreserveDuplicates(root SyntaxElement, include DiagnosticsInclude) []Diagnostic {
	return GetContainedDiagnosticsPreserveDuplicatesWithCancellation(root, include, CancellationToken{})
}

func GetContainedDiagnosticsPreserveDuplicatesWithCancellation(root SyntaxElement, include DiagnosticsInclude, cancellationToken CancellationToken) []Diagnostic {
	return getContainedDiagnostics(root, include, false, cancellationToken)
}

// GetContainedDiagnosticsUTF16 is the public-coordinate counterpart to the
// native byte collector. Absolute syntax diagnostics are already public
// coordinates and are preserved; relative syntax and internal semantic
// diagnostics are located in UTF-16 code units.
func GetContainedDiagnosticsUTF16(root SyntaxElement, include DiagnosticsInclude, distinct bool, cancellationToken CancellationToken) []Diagnostic {
	cancellationToken.ThrowIfCancellationRequested()
	if root == nil || include == 0 {
		return []Diagnostic{}
	}
	diagnostics := []Diagnostic{}
	seen := map[diagnosticIdentity]struct{}{}
	add := func(items []Diagnostic) {
		for _, diagnostic := range items {
			if distinct {
				if identity, hasIdentity := diagnosticIdentityForEquality(diagnostic); hasIdentity {
					if _, exists := seen[identity]; exists {
						continue
					}
					seen[identity] = struct{}{}
				}
			}
			diagnostics = append(diagnostics, diagnostic)
		}
	}

	rootText := root.Text(IncludeTriviaAll)
	nextTokens := nextSyntaxTokensByTokenUTF16(root)
	var walk func(SyntaxElement, int)
	walk = func(element SyntaxElement, triviaStart int) {
		if element == nil {
			return
		}
		cancellationToken.ThrowIfCancellationRequested()
		if include&DiagnosticsIncludeSyntactic != 0 {
			add(diagnosticsWithElementLocationUTF16(element, triviaStart, element.Diagnostics(), nextTokens))
		}
		semantic, ok := element.(interface {
			SemanticDiagnostics() []Diagnostic
			Alternates() []*SyntaxNode
		})
		if ok && include&DiagnosticsIncludeSemantic != 0 {
			add(DiagnosticsToUTF16(rootText, semantic.SemanticDiagnostics()))
		}

		childStart := triviaStart
		for _, child := range element.Children() {
			walk(child, childStart)
			if child != nil {
				childStart += child.UTF16FullWidth()
			}
		}

		if include&DiagnosticsIncludeExpansion != 0 {
			if expr, ok := element.(Expression); ok {
				if body := expressionCalledFunctionBody(expr); body != nil {
					items := GetContainedDiagnosticsUTF16(body, include, distinct, cancellationToken)
					if firstError, ok := firstErrorDiagnostic(items); ok {
						start, length := expressionExpansionDiagnosticLocationUTF16(expr, triviaStart)
						add([]Diagnostic{DiagnosticErrorInExpansion(expressionReferencedSymbolName(expr), firstError.Message).WithLocation(start, length)})
					}
				}
			}
		}
		if ok && include&DiagnosticsIncludeSemantic != 0 {
			for _, alternate := range semantic.Alternates() {
				items := GetContainedDiagnosticsUTF16(alternate, include, distinct, cancellationToken)
				add(relocateDiagnosticsToTreeUTF16(alternate.Tree(), syntaxElementTree(element), items))
			}
		}
	}
	walk(root, 0)
	return diagnostics
}

// OwnedDiagnostic retains the syntax element and source path that own a
// diagnostic. Consumers that combine syntax and independently produced
// semantic diagnostics need this information to preserve tree preorder.
type OwnedDiagnostic struct {
	Element    SyntaxElement
	Path       []int
	Diagnostic Diagnostic
}

func GetOwnedSyntacticDiagnostics(root SyntaxElement) []OwnedDiagnostic {
	if root == nil {
		return []OwnedDiagnostic{}
	}
	nextTokens := nextSyntaxTokensByToken(root)
	out := []OwnedDiagnostic{}
	var walk func(SyntaxElement, int, []int)
	walk = func(element SyntaxElement, triviaStart int, path []int) {
		if element == nil {
			return
		}
		for _, diagnostic := range diagnosticsWithElementLocation(element, triviaStart, element.Diagnostics(), nextTokens) {
			out = append(out, OwnedDiagnostic{Element: element, Path: append([]int{}, path...), Diagnostic: diagnostic})
		}
		childStart := triviaStart
		for i, child := range element.Children() {
			walk(child, childStart, append(path, i))
			if child != nil {
				childStart += child.FullWidth()
			}
		}
	}
	walk(root, 0, nil)
	return out
}

func getContainedDiagnostics(root SyntaxElement, include DiagnosticsInclude, distinct bool, cancellationToken CancellationToken) []Diagnostic {
	cancellationToken.ThrowIfCancellationRequested()
	if root == nil || include == 0 {
		return []Diagnostic{}
	}
	diagnostics := []Diagnostic{}
	seen := map[diagnosticIdentity]struct{}{}
	add := func(items []Diagnostic) {
		for _, diagnostic := range items {
			if distinct {
				if identity, hasIdentity := diagnosticIdentityForEquality(diagnostic); hasIdentity {
					if _, exists := seen[identity]; exists {
						continue
					}
					seen[identity] = struct{}{}
				}
			}
			diagnostics = append(diagnostics, diagnostic)
		}
	}

	nextTokens := nextSyntaxTokensByTokenWithCancellation(root, cancellationToken)
	var walk func(element SyntaxElement, triviaStart int)
	walk = func(element SyntaxElement, triviaStart int) {
		if element == nil {
			return
		}
		cancellationToken.ThrowIfCancellationRequested()
		if include&DiagnosticsIncludeSyntactic != 0 {
			add(diagnosticsWithElementLocation(element, triviaStart, element.Diagnostics(), nextTokens))
		}
		semantic, ok := element.(interface {
			SemanticDiagnostics() []Diagnostic
			Alternates() []*SyntaxNode
		})
		if ok && include&DiagnosticsIncludeSemantic != 0 {
			add(semantic.SemanticDiagnostics())
		}

		childStart := triviaStart
		for _, child := range element.Children() {
			walk(child, childStart)
			if child != nil {
				childStart += child.FullWidth()
			}
		}

		if include&DiagnosticsIncludeExpansion != 0 {
			if expr, ok := element.(Expression); ok {
				if body := expressionCalledFunctionBody(expr); body != nil {
					cancellationToken.ThrowIfCancellationRequested()
					if firstError, ok := firstErrorDiagnostic(getContainedDiagnostics(body, include, distinct, cancellationToken)); ok {
						start, length := expressionExpansionDiagnosticLocation(expr, triviaStart)
						add([]Diagnostic{DiagnosticErrorInExpansion(expressionReferencedSymbolName(expr), firstError.Message).WithByteLocation(start, length)})
					}
				}
			}
		}
		if ok && include&DiagnosticsIncludeSemantic != 0 {
			for _, alternate := range semantic.Alternates() {
				cancellationToken.ThrowIfCancellationRequested()
				items := getContainedDiagnostics(alternate, include, distinct, cancellationToken)
				add(relocateDiagnosticsToTree(alternate.Tree(), syntaxElementTree(element), items))
			}
		}
	}
	walk(root, 0)

	return diagnostics
}

func syntaxElementTree(element SyntaxElement) *SyntaxTree {
	if node := syntaxNodeForBinderLocation(element); node != nil {
		return node.Tree()
	}
	if element != nil {
		return element.Tree()
	}
	return nil
}

func relocateDiagnosticsToTree(tree, target *SyntaxTree, diagnostics []Diagnostic) []Diagnostic {
	if tree == nil || target == nil || tree == target || len(diagnostics) == 0 {
		return diagnostics
	}
	offset := 0
	current := tree
	for current != nil && current != target && current.Original != nil {
		offset += current.OffsetInOriginal
		current = current.Original
	}
	if current != target {
		return diagnostics
	}
	out := make([]Diagnostic, len(diagnostics))
	for i, diagnostic := range diagnostics {
		if diagnostic.HasLocation() {
			diagnostic = diagnostic.WithByteLocation(diagnostic.Start+offset, diagnostic.Length)
		}
		out[i] = diagnostic
	}
	return out
}

func relocateDiagnosticsToTreeUTF16(tree, target *SyntaxTree, diagnostics []Diagnostic) []Diagnostic {
	if tree == nil || target == nil || tree == target || len(diagnostics) == 0 {
		return diagnostics
	}
	offset := 0
	current := tree
	for current != nil && current != target && current.Original != nil {
		originalText := ""
		if current.Original.Root != nil {
			originalText = current.Original.Root.Text(IncludeTriviaAll)
		}
		offset += utf16TextWidthOfBytePrefix(originalText, current.OffsetInOriginal)
		current = current.Original
	}
	if current != target {
		return diagnostics
	}
	out := make([]Diagnostic, len(diagnostics))
	for i, diagnostic := range diagnostics {
		if diagnostic.HasLocation() {
			diagnostic = diagnostic.WithLocation(diagnostic.Start+offset, diagnostic.Length)
		}
		out[i] = diagnostic
	}
	return out
}

func walkSyntaxWithTriviaStart(root SyntaxElement, triviaStart int, fn func(SyntaxElement, int)) {
	walkSyntaxWithTriviaStartWithCancellation(root, triviaStart, CancellationToken{}, fn)
}

func walkSyntaxWithTriviaStartWithCancellation(root SyntaxElement, triviaStart int, cancellationToken CancellationToken, fn func(SyntaxElement, int)) {
	if root == nil || fn == nil {
		return
	}
	cancellationToken.ThrowIfCancellationRequested()
	fn(root, triviaStart)
	childStart := triviaStart
	for _, child := range root.Children() {
		if child != nil {
			walkSyntaxWithTriviaStartWithCancellation(child, childStart, cancellationToken, fn)
			childStart += child.FullWidth()
		}
	}
}

type syntaxTokenPosition struct {
	token       *SyntaxToken
	triviaStart int
}

type syntaxTokenPositionUTF16 struct {
	token       *SyntaxToken
	triviaStart int
}

func (p syntaxTokenPositionUTF16) textStart() int {
	if p.token == nil {
		return p.triviaStart
	}
	return p.triviaStart + p.token.UTF16TriviaWidth()
}

func (p syntaxTokenPosition) textStart() int {
	if p.token == nil {
		return p.triviaStart
	}
	return p.triviaStart + p.token.TriviaWidth()
}

func nextSyntaxTokensByToken(root SyntaxElement) map[*SyntaxToken]syntaxTokenPosition {
	return nextSyntaxTokensByTokenWithCancellation(root, CancellationToken{})
}

func nextSyntaxTokensByTokenWithCancellation(root SyntaxElement, cancellationToken CancellationToken) map[*SyntaxToken]syntaxTokenPosition {
	positions := []syntaxTokenPosition{}
	collectSyntaxTokenPositionsWithCancellation(root, 0, &positions, cancellationToken)
	if len(positions) < 2 {
		return nil
	}
	next := make(map[*SyntaxToken]syntaxTokenPosition, len(positions)-1)
	for i := 0; i+1 < len(positions); i++ {
		cancellationToken.ThrowIfCancellationRequested()
		for j := i + 1; j < len(positions); j++ {
			if positions[j].token != nil && positions[j].token.Width() > 0 {
				next[positions[i].token] = positions[j]
				break
			}
		}
	}
	return next
}

func nextSyntaxTokensByTokenUTF16(root SyntaxElement) map[*SyntaxToken]syntaxTokenPositionUTF16 {
	positions := []syntaxTokenPositionUTF16{}
	var collect func(SyntaxElement, int)
	collect = func(element SyntaxElement, triviaStart int) {
		if element == nil {
			return
		}
		if token, ok := element.(*SyntaxToken); ok {
			positions = append(positions, syntaxTokenPositionUTF16{token: token, triviaStart: triviaStart})
			return
		}
		childStart := triviaStart
		for _, child := range element.Children() {
			if child != nil {
				collect(child, childStart)
				childStart += child.UTF16FullWidth()
			}
		}
	}
	collect(root, 0)
	if len(positions) < 2 {
		return nil
	}
	next := make(map[*SyntaxToken]syntaxTokenPositionUTF16, len(positions)-1)
	for i := 0; i+1 < len(positions); i++ {
		for j := i + 1; j < len(positions); j++ {
			if positions[j].token != nil && positions[j].token.UTF16Width() > 0 {
				next[positions[i].token] = positions[j]
				break
			}
		}
	}
	return next
}

func collectSyntaxTokenPositions(element SyntaxElement, triviaStart int, positions *[]syntaxTokenPosition) {
	collectSyntaxTokenPositionsWithCancellation(element, triviaStart, positions, CancellationToken{})
}

func collectSyntaxTokenPositionsWithCancellation(element SyntaxElement, triviaStart int, positions *[]syntaxTokenPosition, cancellationToken CancellationToken) {
	if element == nil {
		return
	}
	cancellationToken.ThrowIfCancellationRequested()
	if token, ok := element.(*SyntaxToken); ok {
		*positions = append(*positions, syntaxTokenPosition{token: token, triviaStart: triviaStart})
		return
	}
	childStart := triviaStart
	for _, child := range element.Children() {
		if child != nil {
			collectSyntaxTokenPositionsWithCancellation(child, childStart, positions, cancellationToken)
			childStart += child.FullWidth()
		}
	}
}

func diagnosticsWithElementLocation(element SyntaxElement, triviaStart int, diagnostics []Diagnostic, nextTokens map[*SyntaxToken]syntaxTokenPosition) []Diagnostic {
	if len(diagnostics) == 0 {
		return []Diagnostic{}
	}
	out := make([]Diagnostic, len(diagnostics))
	for i, diagnostic := range diagnostics {
		out[i] = diagnosticWithElementLocation(element, triviaStart, diagnostic, nextTokens)
	}
	return out
}

func diagnosticWithElementLocation(element SyntaxElement, triviaStart int, diagnostic Diagnostic, nextTokens map[*SyntaxToken]syntaxTokenPosition) Diagnostic {
	switch diagnostic.LocationKind {
	case DiagnosticLocationRelative:
		if token, ok := element.(*SyntaxToken); ok && token.Width() == 0 {
			if next, ok := nextTokens[token]; ok && next.token != nil && next.textStart()-elementEnd(element, triviaStart) < 2 && !HasLineBreaks(next.token.Trivia()) {
				return diagnostic.WithByteLocation(next.textStart(), next.token.Width())
			}
		}
		return diagnostic.WithByteLocation(elementTextStart(element, triviaStart), element.Width())
	case DiagnosticLocationRelativeEnd:
		return diagnostic.WithByteLocation(elementEnd(element, triviaStart), 0)
	default:
		return diagnostic
	}
}

func diagnosticsWithElementLocationUTF16(element SyntaxElement, triviaStart int, diagnostics []Diagnostic, nextTokens map[*SyntaxToken]syntaxTokenPositionUTF16) []Diagnostic {
	if len(diagnostics) == 0 {
		return []Diagnostic{}
	}
	out := make([]Diagnostic, len(diagnostics))
	for i, diagnostic := range diagnostics {
		switch diagnostic.LocationKind {
		case DiagnosticLocationRelative:
			if token, ok := element.(*SyntaxToken); ok && token.UTF16Width() == 0 {
				if next, ok := nextTokens[token]; ok && next.token != nil && next.textStart()-elementEndUTF16(element, triviaStart) < 2 && !HasLineBreaks(next.token.Trivia()) {
					diagnostic = diagnostic.WithLocation(next.textStart(), next.token.UTF16Width())
					break
				}
			}
			diagnostic = diagnostic.WithLocation(elementTextStartUTF16(element, triviaStart), element.UTF16Width())
		case DiagnosticLocationRelativeEnd:
			diagnostic = diagnostic.WithLocation(elementEndUTF16(element, triviaStart), 0)
		}
		out[i] = diagnostic
	}
	return out
}

func elementTextStart(element SyntaxElement, triviaStart int) int {
	if element == nil {
		return triviaStart
	}
	return triviaStart + element.TriviaWidth()
}

func elementEnd(element SyntaxElement, triviaStart int) int {
	if element == nil {
		return triviaStart
	}
	return triviaStart + element.FullWidth()
}

func elementTextStartUTF16(element SyntaxElement, triviaStart int) int {
	if element == nil {
		return triviaStart
	}
	return triviaStart + element.UTF16TriviaWidth()
}

func elementEndUTF16(element SyntaxElement, triviaStart int) int {
	if element == nil {
		return triviaStart
	}
	return triviaStart + element.UTF16FullWidth()
}

func expressionExpansionDiagnosticLocation(expression Expression, triviaStart int) (int, int) {
	if call, ok := expression.(*FunctionCallExpression); ok {
		if name := call.Name(); name != nil {
			return elementTextStart(name, triviaStart), name.Width()
		}
	}
	return elementTextStart(expression, triviaStart), expression.Width()
}

func expressionExpansionDiagnosticLocationUTF16(expression Expression, triviaStart int) (int, int) {
	if call, ok := expression.(*FunctionCallExpression); ok {
		if name := call.Name(); name != nil {
			return elementTextStartUTF16(name, triviaStart), name.UTF16Width()
		}
	}
	return elementTextStartUTF16(expression, triviaStart), expression.UTF16Width()
}

func expressionCalledFunctionBody(expression Expression) *SyntaxNode {
	if accessor, ok := expression.(interface{ GetCalledFunctionBody() *SyntaxNode }); ok {
		return accessor.GetCalledFunctionBody()
	}
	return nil
}

func expressionReferencedSymbolName(expression Expression) string {
	if accessor, ok := expression.(interface{ ReferencedSymbol() Symbol }); ok {
		if symbol := accessor.ReferencedSymbol(); symbol != nil && symbol.SymbolName() != "" {
			return symbol.SymbolName()
		}
	}
	return "<unknown>"
}

func firstErrorDiagnostic(diagnostics []Diagnostic) (Diagnostic, bool) {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == DiagnosticSeverityError {
			return diagnostic, true
		}
	}
	return Diagnostic{}, false
}

func tupleResultSymbol(expression Expression) *TupleSymbol {
	if accessor, ok := expression.(interface{ ReferencedSymbol() Symbol }); ok {
		switch symbol := GetResultType(accessor.ReferencedSymbol()).(type) {
		case *TupleSymbol:
			return symbol
		}
	}
	return nil
}

func entityGroupElementResultSymbol(expression Expression) *EntityGroupElementSymbol {
	if accessor, ok := expression.(interface{ ReferencedSymbol() Symbol }); ok {
		switch symbol := GetResultType(accessor.ReferencedSymbol()).(type) {
		case *EntityGroupElementSymbol:
			return symbol
		}
	}
	return nil
}
