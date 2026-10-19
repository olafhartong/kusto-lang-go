package kustolang

// Ported from:
//   src/Kusto.Language/Editor/ClientDirective.cs
//   src/Kusto.Language/Editor/ClientDirectiveArgument.cs
//   src/Kusto.Language/Editor/ClientDirectiveInfo.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

type ClientDirectiveArgument struct {
	argument internal.ClientDirectiveArgument
}

func NewClientDirectiveArgument(name, text string, value any) ClientDirectiveArgument {
	return wrapClientDirectiveArgument(internal.NewClientDirectiveArgument(name, text, value))
}

func wrapClientDirectiveArgument(argument internal.ClientDirectiveArgument) ClientDirectiveArgument {
	return ClientDirectiveArgument{argument: argument}
}

func (a ClientDirectiveArgument) Name() string { return a.argument.Name }

func (a ClientDirectiveArgument) Text() string { return a.argument.Text }

func (a ClientDirectiveArgument) Value() any { return a.argument.Value }

type ClientDirective struct {
	directive internal.ClientDirective
}

func NewClientDirective(text EditString, name string, afterNameText, argumentsText EditString, arguments []ClientDirectiveArgument, afterDirectiveText EditString) ClientDirective {
	converted := make([]internal.ClientDirectiveArgument, len(arguments))
	for i, argument := range arguments {
		converted[i] = argument.argument
	}
	return wrapClientDirective(internal.NewClientDirective(
		text.value,
		name,
		afterNameText.value,
		argumentsText.value,
		converted,
		afterDirectiveText.value,
	))
}

func wrapClientDirective(directive internal.ClientDirective) ClientDirective {
	return ClientDirective{directive: directive}
}

func (d ClientDirective) Text() EditString { return wrapEditString(d.directive.Text) }

func (d ClientDirective) Name() string { return d.directive.Name }

func (d ClientDirective) AfterNameText() EditString {
	return wrapEditString(d.directive.AfterNameText)
}

func (d ClientDirective) ArgumentsText() EditString {
	return wrapEditString(d.directive.ArgumentsText)
}

func (d ClientDirective) Arguments() []ClientDirectiveArgument {
	result := make([]ClientDirectiveArgument, len(d.directive.Arguments))
	for i, argument := range d.directive.Arguments {
		result[i] = wrapClientDirectiveArgument(argument)
	}
	return result
}

func (d ClientDirective) AfterDirectiveText() EditString {
	return wrapEditString(d.directive.AfterDirectiveText)
}

func TryParseClientDirective(text EditString) (ClientDirective, bool) {
	directive, ok := internal.TryParseClientDirective(text.value)
	if !ok {
		return ClientDirective{}, false
	}
	return wrapClientDirective(directive), true
}

type ClientDirectiveInfo struct {
	info internal.ClientDirectiveInfo
}

func NewClientDirectiveInfo(text EditString, directives []ClientDirective, remainingText EditString) ClientDirectiveInfo {
	converted := make([]internal.ClientDirective, len(directives))
	for i, directive := range directives {
		converted[i] = directive.directive
	}
	return wrapClientDirectiveInfo(internal.NewClientDirectiveInfo(text.value, converted, remainingText.value))
}

func wrapClientDirectiveInfo(info internal.ClientDirectiveInfo) ClientDirectiveInfo {
	return ClientDirectiveInfo{info: info}
}

func (i ClientDirectiveInfo) Text() EditString { return wrapEditString(i.info.Text) }

func (i ClientDirectiveInfo) Directives() []ClientDirective {
	result := make([]ClientDirective, len(i.info.Directives))
	for index, directive := range i.info.Directives {
		result[index] = wrapClientDirective(directive)
	}
	return result
}

func (i ClientDirectiveInfo) RemainingText() EditString {
	return wrapEditString(i.info.RemainingText)
}

func ParseClientDirectiveInfo(text EditString) ClientDirectiveInfo {
	return wrapClientDirectiveInfo(internal.ParseClientDirectiveInfo(text.value))
}
