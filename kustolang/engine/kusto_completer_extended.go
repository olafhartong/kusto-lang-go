package kustolang

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Ported from src/Kusto.Language/Editor/Kusto/KustoCompleter.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

const (
	kustoExtendedCompletionMaxSize  = 10
	kustoExtendedCompletionMaxCount = 20
)

func (c *kustoCompleter) getExtendedSyntaxCompletions(position int, builder *kustoCompletionBuilder) {
	if c == nil || c.code == nil || builder == nil || len(c.code.Tokens) == 0 {
		return
	}
	root, startIndex, matchIndex, ok := c.extendedGrammarContext(position)
	if !ok {
		return
	}
	paths := FindAnnotatedParserPaths(NewArraySource(c.code.Tokens[:matchIndex+1]), matchIndex, root, startIndex)
	rootPaths := make([]*ParserPath[LexicalToken], 0, len(paths))
	seen := map[Parser[LexicalToken]]struct{}{}
	for _, path := range paths {
		if extendedPathIsInConditionalTest(path) {
			continue
		}
		root := extendedLocalRoot(path)
		if root == nil || root.Parent() == nil {
			continue
		}
		item, ok := extendedDeepCompletionItem(root.Parser())
		if !ok || item.Kind == CompletionKindPunctuation {
			continue
		}
		if _, exists := seen[root.Parser()]; exists {
			continue
		}
		seen[root.Parser()] = struct{}{}
		rootPaths = append(rootPaths, root)
	}

	for _, path := range rootPaths {
		local := newKustoCompletionBuilder()
		c.buildExtendedSyntaxCompletions(path, path.Parent(), nil, local)
		builder.addRange(local.items())
	}
}

func (c *kustoCompleter) extendedGrammarContext(position int) (Parser[LexicalToken], int, int, bool) {
	if c == nil || c.code == nil || len(c.code.Tokens) == 0 {
		return nil, 0, 0, false
	}
	if c.code.Kind == CodeKindCommand {
		context, ok := c.commandCompletionContext(position)
		if ok {
			grammar := CommandGrammarFrom(c.code.Globals)
			if grammar != nil && grammar.CommandCompletionParser() != nil {
				return grammar.CommandCompletionParser(), context.startIndex, context.findOffset, true
			}
		}
	}
	if root, startIndex, matchIndex, ok := c.extendedQueryGrammarContext(position); ok {
		return root, startIndex, matchIndex, true
	}
	if !kustoCodeIsQueryCompletionPosition(c.code, position) {
		return nil, 0, 0, false
	}
	grammar := QueryGrammarFrom(c.code.Globals)
	matchIndex := c.code.GetTokenIndex(position)
	if grammar == nil || grammar.QueryBlock == nil || matchIndex < 0 || matchIndex >= len(c.code.Tokens) {
		return nil, 0, 0, false
	}
	return grammar.QueryBlock, 0, matchIndex, true
}

func (c *kustoCompleter) extendedQueryGrammarContext(position int) (Parser[LexicalToken], int, int, bool) {
	if c == nil || c.code == nil || c.code.Syntax() == nil {
		return nil, 0, 0, false
	}
	grammar := QueryGrammarFrom(c.code.Globals)
	if grammar == nil {
		return nil, 0, 0, false
	}
	refs, tokenIndex, ok := kustoServiceTokenLeftRef(c.code, position)
	if !ok {
		return nil, 0, 0, false
	}
	// SyntaxTree.GetTokenAt never selects zero-width recovery tokens. The native
	// reference index includes them, so move left to the same concrete token.
	for tokenIndex >= 0 && refs[tokenIndex].token != nil && refs[tokenIndex].token.Width() == 0 {
		tokenIndex--
	}
	if tokenIndex < 0 || refs[tokenIndex].token == nil {
		return nil, 0, 0, false
	}
	token := refs[tokenIndex].token
	nextToken := token.GetNextToken()
	matchIndex := c.code.GetTokenIndex(position)
	context := func(start int, parser Parser[LexicalToken]) (Parser[LexicalToken], int, int, bool) {
		if parser == nil || start < 0 || matchIndex < 0 || matchIndex >= len(c.code.Tokens) {
			return nil, 0, 0, false
		}
		return parser, c.code.GetTokenIndex(start), matchIndex, true
	}

	for current := token.Parent(); current != nil; current = current.Parent() {
		wrapped := WrapGeneratedSyntaxElement(current)
		if operator, ok := wrapped.(QueryOperator); ok {
			parent := WrapGeneratedSyntaxElement(current.Parent())
			if _, ok := parent.(*PipeExpression); ok && current.IndexInParent() == 2 {
				return context(operator.TextStart(), grammar.FollowingPipeElementExpression)
			}
			switch parent.(type) {
			case *ParenthesizedExpression, *ExpressionStatement, *FunctionBody:
				return context(operator.TextStart(), grammar.PipeExpression)
			}
			continue
		}

		switch node := wrapped.(type) {
		case *PipeExpression:
			bar := node.Bar()
			if bar != nil && position >= bar.TriviaStart() {
				if prior, ok := node.Expression().(*PipeExpression); ok && prior.Operator() != nil {
					return context(prior.Operator().TextStart(), grammar.PipeSubExpression)
				}
			}
		case *FunctionBody:
			return context(node.TextStart(), grammar.FunctionBody)
		case Statement:
			if extendedSyntaxIsQueryPart(node, 0) {
				return context(node.TextStart(), grammar.StatementList)
			}
		case *SeparatedElement:
			if _, isStatement := syntaxElementAs[Statement](node.Element()); isStatement && extendedSyntaxIsQueryPart(node, 0) {
				return context(node.TextStart(), grammar.StatementList)
			}
			if _, isExpression := syntaxElementAs[Expression](node.Element()); isExpression {
				parent := node.Parent()
				if parent != nil {
					parent = parent.Parent()
				}
				if _, ok := WrapGeneratedSyntaxElement(parent).(*DataTableExpression); ok {
					return context(node.TextStart(), grammar.LiteralList)
				}
			}
		case *FunctionCallExpression:
			arguments := node.ArgumentList()
			if arguments == nil || arguments.OpenParen() == nil || arguments.CloseParen() == nil {
				continue
			}
			close := arguments.CloseParen()
			insideComplete := close.Width() != 0 && position <= close.TextStart()
			insideRecovered := close.Width() == 0 && nextToken != nil && nextToken.Kind() == SyntaxKindEndOfTextToken
			if position > arguments.OpenParen().End() && (insideComplete || insideRecovered) {
				return context(node.TextStart(), grammar.UnnamedExpression)
			}
		case Command:
			commandGrammar := CommandGrammarFrom(c.code.Globals)
			if commandGrammar != nil && commandGrammar.CommandCompletionParser() != nil {
				return context(node.TextStart(), commandGrammar.CommandCompletionParser())
			}
		}
	}
	return nil, 0, 0, false
}

func extendedSyntaxIsQueryPart(element SyntaxElement, contextChildIndex int) bool {
	if element == nil {
		return false
	}
	switch element.Root().Kind() {
	case SyntaxKindQueryBlock:
		return true
	case SyntaxKindCommandBlock:
		if contextChildIndex > 0 {
			if token, ok := element.GetChild(contextChildIndex - 1).(*SyntaxToken); ok && (token.RawText() == "<|" || token.RawText() == "|") {
				return true
			}
		}
		return GetFirstAncestorOrSelf[*FunctionBody](element) != nil
	default:
		return false
	}
}

func extendedPathIsInConditionalTest(path *ParserPath[LexicalToken]) bool {
	_, index, ok := extendedFirstAncestor(path, nil, func(parser Parser[LexicalToken]) bool {
		return parser.IsConditional()
	})
	return ok && index == 0
}

func extendedLocalRoot(path *ParserPath[LexicalToken]) *ParserPath[LexicalToken] {
	current := path
	for current != nil {
		parent, index, ok := extendedFirstAncestor(current, nil, func(parser Parser[LexicalToken]) bool {
			return parser.IsSequence()
		})
		if !ok {
			return path
		}
		if parent.ChildCount() > 1 {
			return parent.GetChild(index)
		}
		current = parent
	}
	return path
}

func (c *kustoCompleter) buildExtendedSyntaxCompletions(path, ceiling *ParserPath[LexicalToken], parts []CompletionItem, builder *kustoCompletionBuilder) {
	for path != nil {
		if builder.count() >= kustoExtendedCompletionMaxCount {
			return
		}
		parser := path.Parser()
		if parser == nil || parser.IsHidden() || parser.IsForward() {
			break
		}
		switch {
		case parser.IsRequired():
			path = path.GetChild(0)
			continue
		case parser.IsConditional():
			path = path.GetChild(1)
			continue
		case c.extendedTrySpecialTexts(parser, &parts):
			path = extendedNextParserInSequence(path, ceiling)
			continue
		case extendedParserHasMatchCompletion(parser):
			item, _ := extendedCompletionItem(parser)
			if len(parts) == 0 && extendedCompletionIsPunctuation(item) {
				return
			}
			parts = append(parts, item)
			if end, ok := extendedEndBracket(item.BeforeText()); ok {
				parts = append(parts, NewCompletionItem(CompletionKindPunctuation, end,
					WithCompletionBeforeText(""), WithCompletionAfterText(end)))
				if sibling := extendedFindNextSibling(path, ceiling, end); sibling != nil {
					path = extendedNextParserInSequence(sibling, ceiling)
					continue
				}
				break
			}
			path = extendedNextParserInSequence(path, ceiling)
			continue
		case parser.IsMatch():
			break
		case parser.IsSequence() && path.ChildCount() > 0:
			if len(parts) >= kustoExtendedCompletionMaxSize {
				break
			}
			path = path.GetChild(0)
			continue
		case parser.IsOptional() && !parser.IsRepetition():
			withOptional := append([]CompletionItem(nil), parts...)
			c.buildExtendedSyntaxCompletions(path.GetChild(0), ceiling, withOptional, builder)
			path = extendedNextParserInSequence(path, ceiling)
			continue
		case parser.IsRepetition() && !parser.IsOptional():
			path = path.GetChild(0)
			continue
		case parser.IsRepetition() && parser.IsOptional():
			path = extendedNextParserInSequence(path, ceiling)
			continue
		case parser.IsAlternation():
			for index := 0; index < path.ChildCount(); index++ {
				alternativeParts := append([]CompletionItem(nil), parts...)
				alternativeBuilder := builder
				if len(parts) == 0 {
					alternativeBuilder = newKustoCompletionBuilder()
				}
				c.buildExtendedSyntaxCompletions(path.GetChild(index), ceiling, alternativeParts, alternativeBuilder)
				if alternativeBuilder != builder {
					builder.addRange(alternativeBuilder.items())
				}
			}
			return
		case path.ChildCount() == 1:
			path = path.GetChild(0)
			continue
		}
		break
	}

	if len(parts) <= 1 {
		return
	}
	item := createExtendedSyntaxCompletion(parts)
	first, ok := utf8FirstRune(item.DisplayText)
	if item.DisplayText != "" && ok && !unicode.IsSpace(first) && !unicode.IsPunct(first) {
		builder.add(item)
	}
}

func extendedParserHasMatchCompletion(parser Parser[LexicalToken]) bool {
	item, ok := extendedCompletionItem(parser)
	if !ok || item.DisplayText != item.BeforeText() || item.AfterText() != "" {
		return false
	}
	current := parser
	for count := 0; current != nil && count <= 20; count++ {
		if current.IsMatch() {
			return true
		}
		if current.IsHidden() || current.IsForward() || current.ChildParserCount() != 1 {
			return false
		}
		current = current.ChildParser(0)
	}
	return false
}

func (c *kustoCompleter) extendedTrySpecialTexts(parser Parser[LexicalToken], parts *[]CompletionItem) bool {
	if parser == nil || parts == nil || len(*parts) == 0 {
		return false
	}
	grammar := QueryGrammarFrom(c.code.Globals)
	if grammar == nil {
		return false
	}
	var rules *PredefinedRuleParsers
	if commandGrammar := CommandGrammarFrom(c.code.Globals); commandGrammar != nil {
		rules = commandGrammar.PredefinedRules
	}
	hasCaret := false
	for _, item := range *parts {
		if item.HasCaret() {
			hasCaret = true
			break
		}
	}

	appendItem := func(item CompletionItem) bool {
		*parts = append(*parts, item)
		return true
	}
	if rule := commandCompletionRuleFromAnnotations(parser.Annotations()); rule != "" {
		switch rule {
		case "TableNameReference", "ExternalTableNameReference", "MaterializedViewNameReference",
			"DatabaseFunctionNameReference", "DatabaseNameReference", "ClusterNameReference",
			"GraphModelNameReference", "GraphModelSnapshotNameReference", "DatabaseOrTableNameReference",
			"DatabaseOrTableOrColumnNameReference", "TableOrColumnNameReference",
			"DatabaseTableColumnNameReference", "ColumnNameReference", "TableColumnNameReference":
			return appendItem(NewCompletionItem(CompletionKindUnknown, "",
				WithCompletionBeforeText(" "), WithCompletionAfterText(" ")))
		case "QueryInput", "ScriptInput":
			return appendItem(NewCompletionItem(CompletionKindUnknown, "",
				WithCompletionBeforeText(" "), WithCompletionAfterText(" ")))
		case "NameDeclaration", "QualifiedNameDeclaration":
			return appendItem(NewCompletionItem(CompletionKindUnknown, "",
				WithCompletionApplyTexts(NewCompletionTextEdit("name"))))
		case "Value":
			return appendItem(NewCompletionItem(CompletionKindUnknown, "?",
				WithCompletionApplyTexts(NewCompletionTextEdit("0"))))
		case "Type":
			return appendItem(NewCompletionItem(CompletionKindUnknown, "",
				WithCompletionApplyTexts(NewCompletionTextEdit("real"))))
		case "FunctionDeclaration":
			return appendItem(NewCompletionItem(CompletionKindUnknown, "() { }"))
		case "FunctionBody":
			return appendItem(NewCompletionItem(CompletionKindUnknown, "{ }"))
		case "AnyGuidLiteralOrString", "GuidLiteral", "StringLiteral", "RawGuidLiteral":
			return appendItem(NewCompletionItem(CompletionKindUnknown, "\"\"",
				WithCompletionBeforeText("\""), WithCompletionAfterText("\"")))
		case "BracketedStringLiteral", "BracketedInputText":
			return appendItem(NewCompletionItem(CompletionKindUnknown, "[]",
				WithCompletionBeforeText("["), WithCompletionAfterText("]")))
		}
	}
	switch parser {
	case grammar.SimpleNameReference, grammar.WildcardedNameReference, grammar.BracedName, grammar.BracketedName:
		return appendItem(NewCompletionItem(CompletionKindUnknown, "",
			WithCompletionBeforeText(" "), WithCompletionAfterText(" ")))
	case grammar.SimpleNameDeclaration, grammar.SimpleNameDeclarationExpression, grammar.BracketedNameDeclaration:
		return appendItem(NewCompletionItem(CompletionKindUnknown, "",
			WithCompletionApplyTexts(NewCompletionTextEdit("name"))))
	case grammar.Literal:
		return appendItem(NewCompletionItem(CompletionKindUnknown, "?",
			WithCompletionApplyTexts(NewCompletionTextEdit("0"))))
	case grammar.Expression, grammar.NamedExpression, grammar.UnnamedExpression:
		if hasCaret {
			return appendItem(NewCompletionItem(CompletionKindUnknown, "",
				WithCompletionApplyTexts(NewCompletionTextEdit("expr"))))
		}
		return appendItem(NewCompletionItem(CompletionKindUnknown, "",
			WithCompletionBeforeText(" "), WithCompletionAfterText(" ")))
	case grammar.FunctionBody:
		return appendItem(NewCompletionItem(CompletionKindUnknown, "{ }"))
	case grammar.StringLiteral:
		return appendItem(NewCompletionItem(CompletionKindUnknown, "\"\"",
			WithCompletionBeforeText("\""), WithCompletionAfterText("\"")))
	}
	if rules == nil {
		return false
	}
	if extendedParserIsOneOf(parser,
		rules.TableNameReference,
		rules.ExternalTableNameReference,
		rules.MaterializedViewNameReference,
		rules.DatabaseFunctionNameReference,
		rules.DatabaseNameReference,
		rules.ClusterNameReference,
		rules.GraphModelNameReference,
		rules.GraphModelSnapshotNameReference,
		rules.DatabaseOrTableNameReference,
		rules.DatabaseOrTableOrColumnNameReference,
		rules.TableOrColumnNameReference,
		rules.DatabaseTableColumnNameReference,
		rules.ColumnNameReference,
		rules.TableColumnNameReference,
	) {
		return appendItem(NewCompletionItem(CompletionKindUnknown, "",
			WithCompletionBeforeText(" "), WithCompletionAfterText(" ")))
	}
	if parser == rules.ScriptInput || parser == rules.QueryInput {
		return appendItem(NewCompletionItem(CompletionKindUnknown, "",
			WithCompletionBeforeText(" "), WithCompletionAfterText(" ")))
	}
	if parser == rules.NameDeclaration || parser == rules.QualifiedNameDeclaration {
		return appendItem(NewCompletionItem(CompletionKindUnknown, "",
			WithCompletionApplyTexts(NewCompletionTextEdit("name"))))
	}
	if parser == rules.Value {
		return appendItem(NewCompletionItem(CompletionKindUnknown, "?",
			WithCompletionApplyTexts(NewCompletionTextEdit("0"))))
	}
	if parser == rules.Type {
		return appendItem(NewCompletionItem(CompletionKindUnknown, "",
			WithCompletionApplyTexts(NewCompletionTextEdit("real"))))
	}
	if parser == rules.FunctionDeclaration {
		return appendItem(NewCompletionItem(CompletionKindUnknown, "() { }"))
	}
	if parser == rules.FunctionBody {
		return appendItem(NewCompletionItem(CompletionKindUnknown, "{ }"))
	}
	if extendedParserIsOneOf(parser, rules.AnyGuidLiteralOrString, rules.GuidLiteral, rules.StringLiteral, rules.RawGuidLiteral) {
		return appendItem(NewCompletionItem(CompletionKindUnknown, "\"\"",
			WithCompletionBeforeText("\""), WithCompletionAfterText("\"")))
	}
	if parser == rules.BracketedStringLiteral || parser == rules.BracketedInputText {
		return appendItem(NewCompletionItem(CompletionKindUnknown, "[]",
			WithCompletionBeforeText("["), WithCompletionAfterText("]")))
	}
	return false
}

func extendedParserIsOneOf(parser Parser[LexicalToken], candidates ...Parser[LexicalToken]) bool {
	for _, candidate := range candidates {
		if parser == candidate {
			return true
		}
	}
	return false
}

func extendedCompletionItem(parser Parser[LexicalToken]) (CompletionItem, bool) {
	if parser == nil {
		return CompletionItem{}, false
	}
	items := syntaxCompletionItemsFromAnnotations(parser.Annotations())
	if len(items) == 0 {
		return CompletionItem{}, false
	}
	return items[0], true
}

func extendedDeepCompletionItem(parser Parser[LexicalToken]) (CompletionItem, bool) {
	for parser != nil {
		if item, ok := extendedCompletionItem(parser); ok {
			return item, true
		}
		if parser.ChildParserCount() == 0 {
			break
		}
		parser = parser.ChildParser(0)
	}
	return CompletionItem{}, false
}

func extendedFindNextSibling(path, ceiling *ParserPath[LexicalToken], text string) *ParserPath[LexicalToken] {
	sequence, index, ok := extendedFirstAncestor(path, ceiling, func(parser Parser[LexicalToken]) bool {
		return parser.IsSequence()
	})
	if !ok {
		return nil
	}
	for i := index; i < sequence.ChildCount(); i++ {
		child := sequence.GetChild(i)
		if item, found := extendedDeepCompletionItem(child.Parser()); found && item.BeforeText() == text && item.AfterText() == "" {
			return child
		}
	}
	return nil
}

func extendedNextParserInSequence(path, ceiling *ParserPath[LexicalToken]) *ParserPath[LexicalToken] {
	current := path
	for current != nil {
		sequence, index, ok := extendedFirstAncestor(current, ceiling, func(parser Parser[LexicalToken]) bool {
			return parser.IsSequence()
		})
		if !ok {
			return nil
		}
		if index+1 < sequence.ChildCount() {
			return sequence.GetChild(index + 1)
		}
		current = sequence
	}
	return nil
}

func extendedFirstAncestor(path, ceiling *ParserPath[LexicalToken], match func(Parser[LexicalToken]) bool) (*ParserPath[LexicalToken], int, bool) {
	child := path
	for count := 0; child != nil && count <= 100; count++ {
		if child == ceiling || child.Parent() == nil {
			break
		}
		ancestor := child.Parent()
		if match(ancestor.Parser()) {
			return ancestor, child.IndexInParent(), true
		}
		child = ancestor
	}
	return nil, -1, false
}

func extendedEndBracket(text string) (string, bool) {
	switch text {
	case "(":
		return ")", true
	case "[":
		return "]", true
	case "{":
		return "}", true
	default:
		return "", false
	}
}

func extendedCompletionIsPunctuation(item CompletionItem) bool {
	first, ok := utf8FirstRune(item.DisplayText)
	return ok && unicode.IsPunct(first)
}

func createExtendedSyntaxCompletion(items []CompletionItem) CompletionItem {
	texts := make([]CompletionText, 0, len(items))
	displays := make([]string, 0, len(items))
	for _, item := range items {
		texts = append(texts, item.ApplyTexts...)
		displays = append(displays, item.DisplayText)
	}
	displayText := ""
	for _, display := range displays {
		displayText = extendedCombineWithSpacing(displayText, display)
	}
	matchText := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, displayText)
	return NewCompletionItem(items[0].Kind, displayText,
		WithCompletionMatchText(matchText),
		WithCompletionApplyTexts(extendedCombineTexts(texts)...),
		WithCompletionRank(CompletionRankOther),
		WithCompletionPriority(CompletionPriorityLow))
}

func extendedCombineTexts(texts []CompletionText) []CompletionText {
	result := make([]CompletionText, 0, len(texts))
	for _, current := range texts {
		if len(result) == 0 {
			result = append(result, current)
			continue
		}
		previous := result[len(result)-1]
		switch {
		case !previous.Caret && !current.Caret:
			result[len(result)-1] = NewCompletionText(extendedCombineWithSpacing(previous.Text, current.Text), false, false)
		case extendedNeedsSpacing(previous.Text, current.Text) && !current.Caret:
			current.Text = " " + current.Text
			result = append(result, current)
		case extendedNeedsSpacing(previous.Text, current.Text) && !previous.Caret:
			result[len(result)-1].Text += " "
			result = append(result, current)
		default:
			result = append(result, current)
		}
	}
	return result
}

func extendedCombineWithSpacing(previous, next string) string {
	if extendedNeedsSpacing(previous, next) {
		return previous + " " + next
	}
	return previous + next
}

func extendedNeedsSpacing(previous, next string) bool {
	if previous == "" || next == "" {
		return false
	}
	left, _ := utf8.DecodeLastRuneInString(previous)
	right, _ := utf8.DecodeRuneInString(next)
	if strings.ContainsRune("=<>!~", left) || strings.ContainsRune("=<>!~", right) {
		return true
	}
	if unicode.IsPunct(left) && unicode.IsPunct(right) {
		return false
	}
	if unicode.IsSpace(left) || unicode.IsSpace(right) {
		return false
	}
	return true
}
