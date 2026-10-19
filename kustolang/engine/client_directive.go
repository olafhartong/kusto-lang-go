package kustolang

import (
	"strconv"
	"unicode/utf8"
)

// Ported from:
//   src/Kusto.Language/Editor/ClientDirective.cs
//   src/Kusto.Language/Editor/ClientDirectiveInfo.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type ClientDirective struct {
	Text               EditString
	Name               string
	AfterNameText      EditString
	ArgumentsText      EditString
	Arguments          []ClientDirectiveArgument
	AfterDirectiveText EditString
}

func NewClientDirective(text EditString, name string, afterNameText EditString, argumentsText EditString, arguments []ClientDirectiveArgument, afterDirectiveText EditString) ClientDirective {
	return ClientDirective{
		Text:               text,
		Name:               name,
		AfterNameText:      afterNameText,
		ArgumentsText:      argumentsText,
		Arguments:          ListToReadOnly(arguments),
		AfterDirectiveText: afterDirectiveText,
	}
}

func TryParseClientDirective(text EditString) (ClientDirective, bool) {
	pos := ScanTrivia(text.CurrentText, 0)
	if pos >= text.Length() || text.At(pos) != '#' {
		return ClientDirective{}, false
	}

	pos++
	nameLength := ScanIdentifier(text.CurrentText, pos)
	if nameLength < 0 {
		nameLength = 0
	}
	name := text.Substring(pos, nameLength).CurrentText
	pos += nameLength

	afterNameText := text.Substring(pos, text.Length()-pos)
	endOfLine := LineEnd(text.CurrentText, pos)

	pos += ScanWhitespace(text.CurrentText, pos)
	argLength := endOfLine - pos
	argumentText := EditStringEmpty
	if argLength > 0 {
		argumentText = text.Substring(pos, argLength)
	}

	afterDirectiveText := EditStringEmpty
	if afterDirectiveStart := NextLineStart(text.CurrentText, pos); afterDirectiveStart >= pos {
		afterDirectiveText = text.Substring(afterDirectiveStart)
	}

	arguments := parseClientDirectiveArguments(argumentText)
	return NewClientDirective(text, name, afterNameText, argumentText, arguments, afterDirectiveText), true
}

func parseClientDirectiveArguments(argumentText EditString) []ClientDirectiveArgument {
	var arguments []ClientDirectiveArgument
	pos := 0
	for pos < argumentText.Length() {
		var argumentName string
		hasArgumentName := false

		pos += ScanWhitespace(argumentText.CurrentText, pos)

		length := ScanIdentifier(argumentText.CurrentText, pos)
		if length > 0 {
			lookahead := length + ScanWhitespace(argumentText.CurrentText, pos+length)
			if pos+lookahead < argumentText.Length() && argumentText.At(pos+lookahead) == '=' {
				argumentName = argumentText.Substring(pos, length).CurrentText
				hasArgumentName = true
				pos += lookahead + 1
			}
		}

		pos += ScanWhitespace(argumentText.CurrentText, pos)
		argumentStart := pos

		if pos < argumentText.Length() && (argumentText.At(pos) == '"' || argumentText.At(pos) == '\'') {
			quote := argumentText.At(pos)
			pos++
			for pos < argumentText.Length() && argumentText.At(pos) != quote {
				pos++
			}

			if pos < argumentText.Length() && argumentText.At(pos) == quote {
				pos++
				argText := argumentText.Substring(argumentStart, pos-argumentStart)
				value := argumentText.Substring(argumentStart+1, (pos-argumentStart)-2).CurrentText
				arguments = append(arguments, NewClientDirectiveArgument(argumentNameOrEmpty(argumentName, hasArgumentName), argText.CurrentText, value))
				continue
			}

			argText := argumentText.Substring(argumentStart, pos-argumentStart)
			value := argumentText.Substring(argumentStart+1, (pos-argumentStart)-1).CurrentText
			arguments = append(arguments, NewClientDirectiveArgument(argumentNameOrEmpty(argumentName, hasArgumentName), argText.CurrentText, value))
			continue
		}

		for pos < argumentText.Length() {
			ch, size := utf8.DecodeRuneInString(argumentText.CurrentText[pos:])
			if IsWhitespace(ch) || ch == ',' || ch == ';' {
				break
			}
			pos += size
		}

		argLength := pos - argumentStart
		if argLength > 0 {
			argText := argumentText.Substring(argumentStart, argLength)
			valueText := argText.CurrentText
			if realLength := ScanRealLiteral(argumentText.CurrentText, argumentStart); realLength >= 0 && realLength == argLength {
				value, _ := strconv.ParseFloat(valueText, 64)
				arguments = append(arguments, NewClientDirectiveArgument(argumentNameOrEmpty(argumentName, hasArgumentName), valueText, value))
				continue
			}
			if longLength := ScanLongLiteral(argumentText.CurrentText, argumentStart); longLength > 0 && longLength == argLength {
				value, _ := strconv.ParseInt(valueText, 10, 64)
				arguments = append(arguments, NewClientDirectiveArgument(argumentNameOrEmpty(argumentName, hasArgumentName), valueText, value))
				continue
			}
			arguments = append(arguments, NewClientDirectiveArgument(argumentNameOrEmpty(argumentName, hasArgumentName), valueText, valueText))
			continue
		}

		if hasArgumentName {
			arguments = append(arguments, NewClientDirectiveArgument(argumentName, "", nil))
			continue
		}

		pos++
	}
	return ListToReadOnly(arguments)
}

func argumentNameOrEmpty(name string, ok bool) string {
	if ok {
		return name
	}
	return ""
}

type ClientDirectiveInfo struct {
	Text          EditString
	Directives    []ClientDirective
	RemainingText EditString
}

func NewClientDirectiveInfo(text EditString, directives []ClientDirective, remainingText EditString) ClientDirectiveInfo {
	return ClientDirectiveInfo{
		Text:          text,
		Directives:    ListToReadOnly(directives),
		RemainingText: remainingText,
	}
}

func ParseClientDirectiveInfo(text EditString) ClientDirectiveInfo {
	var directives []ClientDirective
	pos := 0
	for {
		whitespace := ScanWhitespace(text.CurrentText, pos)
		if pos+whitespace < text.Length() && text.At(pos+whitespace) == '#' {
			lineStart := LineStart(text.CurrentText, pos)
			lineLength := LineLength(text.CurrentText, lineStart)
			directiveText := text.Substring(lineStart, lineLength)
			if directive, ok := TryParseClientDirective(directiveText); ok {
				directives = append(directives, directive)
			}
			nextLineStart := NextLineStart(text.CurrentText, lineStart)
			if nextLineStart < 0 {
				pos = text.Length()
			} else {
				pos = nextLineStart
			}
			continue
		}
		break
	}

	remainingText := text.Substring(pos).Trim()
	return NewClientDirectiveInfo(text, directives, remainingText)
}
