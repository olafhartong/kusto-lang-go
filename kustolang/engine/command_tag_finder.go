package kustolang

import (
	"slices"
	"strings"
)

// CommandParserExpectedTerms returns the terms used to describe a parser failure.
func CommandParserExpectedTerms[TInput any](parser Parser[TInput]) []string {
	finder := commandTagFinder[TInput]{}
	tags := make([]string, 0)
	finder.find(parser, &tags)
	if len(tags) == 0 {
		tags = append(tags, "<token>")
	}

	slices.Sort(tags)
	return slices.Compact(tags)
}

type commandTagFinder[TInput any] struct{}

func (f commandTagFinder[TInput]) find(parser Parser[TInput], tags *[]string) bool {
	if parser == nil {
		return false
	}

	return AcceptParserVisitorResult(parser, ParserVisitorResultFuncs[TInput, bool]{
		Apply: func(parser Parser[TInput], left Parser[TInput], _ Parser[TInput], _ ApplyKind) bool {
			return f.getTag(parser, tags) || f.find(left, tags)
		},
		Best: func(parser Parser[TInput], parsers []Parser[TInput]) bool {
			return f.getTag(parser, tags) || f.all(parsers, tags)
		},
		Convert: func(parser Parser[TInput], _ Parser[TInput]) bool {
			return f.getTag(parser, tags)
		},
		Fails: func(parser Parser[TInput], pattern Parser[TInput]) bool {
			return f.getTag(parser, tags) || f.find(pattern, tags)
		},
		First: func(parser Parser[TInput], parsers []Parser[TInput]) bool {
			return f.getTag(parser, tags) || f.all(parsers, tags)
		},
		Forward: func(parser Parser[TInput], deferred Parser[TInput]) bool {
			return f.getTag(parser, tags) || f.find(deferred, tags)
		},
		If: func(parser Parser[TInput], test Parser[TInput], body Parser[TInput]) bool {
			if f.getTag(parser, tags) {
				return true
			}
			if _, plain := parser.(*IfParser[TInput]); plain {
				return f.find(test, tags)
			}
			return f.find(body, tags)
		},
		Limit: func(parser Parser[TInput], _ Parser[TInput], limited Parser[TInput]) bool {
			return f.getTag(parser, tags) || f.find(limited, tags)
		},
		Map: func(parser Parser[TInput]) bool {
			return f.getTag(parser, tags)
		},
		Match: func(parser Parser[TInput]) bool {
			return f.getTag(parser, tags)
		},
		Not: func(parser Parser[TInput], _ Parser[TInput]) bool {
			return f.getTag(parser, tags)
		},
		OneOrMore: func(parser Parser[TInput], repeated Parser[TInput]) bool {
			return f.getTag(parser, tags) || f.find(repeated, tags)
		},
		Optional: func(parser Parser[TInput], optional Parser[TInput]) bool {
			return f.getTag(parser, tags) || f.find(optional, tags)
		},
		Produce: func(parser Parser[TInput], inner Parser[TInput]) bool {
			return f.getTag(parser, tags) || f.find(inner, tags)
		},
		Required: func(parser Parser[TInput], required Parser[TInput]) bool {
			return f.getTag(parser, tags) || f.find(required, tags)
		},
		Rule: func(parser Parser[TInput], parsers []Parser[TInput]) bool {
			return f.getTag(parser, tags) || f.first(parsers, tags)
		},
		Sequence: func(parser Parser[TInput], parsers []Parser[TInput]) bool {
			return f.getTag(parser, tags) || f.first(parsers, tags)
		},
		ZeroOrMore: func(parser Parser[TInput], repeated Parser[TInput], _ bool) bool {
			return f.getTag(parser, tags) || f.find(repeated, tags)
		},
		Unsupported: func(parser Parser[TInput]) bool {
			return f.getTag(parser, tags)
		},
	})
}

func (f commandTagFinder[TInput]) getTag(parser Parser[TInput], tags *[]string) bool {
	for _, annotation := range parser.Annotations() {
		switch item := annotation.(type) {
		case CompletionItem:
			addCommandTag(item.DisplayText, tags)
			return true
		case *CompletionItem:
			if item != nil {
				addCommandTag(item.DisplayText, tags)
				return true
			}
		}
	}

	if parser.Tag() != "" {
		addCommandTag(parser.Tag(), tags)
		return true
	}
	return false
}

func (f commandTagFinder[TInput]) all(parsers []Parser[TInput], tags *[]string) bool {
	any := false
	for _, parser := range parsers {
		if f.find(parser, tags) {
			any = true
		}
	}
	return any
}

func (f commandTagFinder[TInput]) first(parsers []Parser[TInput], tags *[]string) bool {
	if len(parsers) == 0 {
		return false
	}
	return f.find(parsers[0], tags)
}

func addCommandTag(tag string, tags *[]string) {
	for _, part := range strings.Split(tag, "|") {
		if part = strings.TrimSpace(part); part != "" {
			*tags = append(*tags, part)
		}
	}
}
