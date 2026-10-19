package kustolang

// Ported from src/Kusto.Language/Editor/Kusto/AnnotatedParserFinder.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

const annotatedParserFinderMaxDepth = 400

type annotatedParserFinder[TInput any] struct {
	source           Source[TInput]
	findOffset       int
	path             []ParserAndIndex[TInput]
	results          []Parser[TInput]
	pathResults      []*ParserPath[TInput]
	collectPaths     bool
	prevWasMissing   bool
	missingCount     int
	listMissingCount int
	depth            int
}

func FindAnnotatedParsers[TInput any](source Source[TInput], findOffset int, root Parser[TInput], start int) []Parser[TInput] {
	finder := newAnnotatedParserFinder[TInput](source, findOffset, false)
	finder.find(root, 0, start)
	return ListToReadOnly(finder.results)
}

func FindAnnotatedParserPaths[TInput any](source Source[TInput], findOffset int, root Parser[TInput], start int) []*ParserPath[TInput] {
	finder := newAnnotatedParserFinder[TInput](source, findOffset, true)
	finder.find(root, 0, start)
	return ListToReadOnly(finder.pathResults)
}

func newAnnotatedParserFinder[TInput any](source Source[TInput], findOffset int, collectPaths bool) *annotatedParserFinder[TInput] {
	return &annotatedParserFinder[TInput]{
		source:       source,
		findOffset:   findOffset,
		collectPaths: collectPaths,
	}
}

func (f *annotatedParserFinder[TInput]) find(parser Parser[TInput], indexInParent int, start int) int {
	if f.depth >= annotatedParserFinderMaxDepth {
		return -1
	}
	if parser == nil {
		return -1
	}

	f.depth++
	f.path = append(f.path, NewParserAndIndex(parser, indexInParent))

	var length int
	if parser.IsHidden() {
		length = parser.Scan(f.source, start)
	} else {
		if f.findOffset == start && !f.prevWasMissing && len(parser.Annotations()) > 0 {
			if len(f.results) == 0 || f.missingCount <= f.listMissingCount {
				if f.missingCount < f.listMissingCount {
					f.results = f.results[:0]
					f.pathResults = f.pathResults[:0]
				}
				f.results = append(f.results, parser)
				f.listMissingCount = f.missingCount
				if f.collectPaths {
					f.pathResults = append(f.pathResults, ParserPathFrom(f.path))
				}
			}
		}
		length = f.visit(parser, start)
	}

	f.path = f.path[:len(f.path)-1]
	f.depth--
	return length
}

func (f *annotatedParserFinder[TInput]) visit(parser Parser[TInput], start int) int {
	return AcceptParserVisitorResult(parser, ParserVisitorResultFuncs[TInput, int]{
		Apply: func(_ Parser[TInput], left Parser[TInput], right Parser[TInput], kind ApplyKind) int {
			return f.visitApply(left, right, kind, start)
		},
		Best: func(_ Parser[TInput], parsers []Parser[TInput]) int {
			return f.visitFirstOrBest(parsers, start, true)
		},
		Convert: func(_ Parser[TInput], pattern Parser[TInput]) int {
			return f.find(pattern, 0, start)
		},
		Fails: func(parser Parser[TInput], _ Parser[TInput]) int {
			length := parser.Scan(f.source, start)
			f.prevWasMissing = false
			return length
		},
		First: func(_ Parser[TInput], parsers []Parser[TInput]) int {
			return f.visitFirstOrBest(parsers, start, false)
		},
		Forward: func(_ Parser[TInput], deferred Parser[TInput]) int {
			return f.find(deferred, 0, start)
		},
		If: func(_ Parser[TInput], test Parser[TInput], body Parser[TInput]) int {
			return f.visitIf(test, body, start)
		},
		Limit: func(_ Parser[TInput], limiter Parser[TInput], limited Parser[TInput]) int {
			return f.visitLimit(limiter, limited, start)
		},
		Map: func(parser Parser[TInput]) int {
			length := parser.Scan(f.source, start)
			f.prevWasMissing = length >= 0
			return length
		},
		Match: func(parser Parser[TInput]) int {
			length := parser.Scan(f.source, start)
			if length >= 0 {
				f.prevWasMissing = false
			}
			return length
		},
		Not: func(_ Parser[TInput], pattern Parser[TInput]) int {
			return f.visitNot(pattern, start)
		},
		OneOrMore: func(_ Parser[TInput], repeated Parser[TInput]) int {
			return f.visitOneOrMore(repeated, start)
		},
		Optional: func(_ Parser[TInput], optional Parser[TInput]) int {
			return f.visitOptional(optional, start)
		},
		Produce: func(_ Parser[TInput], inner Parser[TInput]) int {
			return f.find(inner, 0, start)
		},
		Required: func(_ Parser[TInput], required Parser[TInput]) int {
			return f.visitRequired(required, start)
		},
		Rule: func(_ Parser[TInput], parsers []Parser[TInput]) int {
			return f.visitSequence(parsers, start)
		},
		Sequence: func(_ Parser[TInput], parsers []Parser[TInput]) int {
			return f.visitSequence(parsers, start)
		},
		ZeroOrMore: func(_ Parser[TInput], repeated Parser[TInput], zeroOrOne bool) int {
			return f.visitZeroOrMore(repeated, start, zeroOrOne)
		},
		Unsupported: func(parser Parser[TInput]) int {
			if parser == nil {
				return -1
			}
			return parser.Scan(f.source, start)
		},
	})
}

func (f *annotatedParserFinder[TInput]) visitApply(left Parser[TInput], right Parser[TInput], kind ApplyKind, start int) int {
	leftLength := f.find(left, 0, start)
	if leftLength < 0 {
		return leftLength
	}

	for {
		rightLength := f.find(right, 1, start+leftLength)
		if rightLength < 0 {
			if kind == ApplyKindOne {
				return -leftLength + rightLength
			}
			return leftLength
		}

		leftLength += rightLength
		if kind != ApplyKindZeroOrMore {
			return leftLength
		}
	}
}

func (f *annotatedParserFinder[TInput]) visitFirstOrBest(parsers []Parser[TInput], start int, isBest bool) int {
	oldPrevWasMissing := f.prevWasMissing
	oldMissingCount := f.missingCount

	bestLength := -1
	bestMissingCount := oldMissingCount
	bestPrevWasMissing := oldPrevWasMissing

	for i, parser := range parsers {
		f.prevWasMissing = oldPrevWasMissing
		f.missingCount = oldMissingCount

		length := f.find(parser, i, start)
		if i == 0 {
			bestLength = length
			bestMissingCount = f.missingCount
			bestPrevWasMissing = f.prevWasMissing
		} else if length < 0 && bestLength < 0 && length < bestLength {
			bestLength = length
			bestMissingCount = f.missingCount
			bestPrevWasMissing = f.prevWasMissing
		} else if length > 0 && (bestLength < 0 || (isBest && length > bestLength)) {
			bestLength = length
			bestMissingCount = f.missingCount
			bestPrevWasMissing = f.prevWasMissing
		}
	}

	f.missingCount = bestMissingCount
	f.prevWasMissing = bestPrevWasMissing
	return bestLength
}

func (f *annotatedParserFinder[TInput]) visitIf(test Parser[TInput], body Parser[TInput], start int) int {
	oldPrevWasMissing := f.prevWasMissing
	oldMissingCount := f.missingCount

	length := test.Scan(f.source, start)
	if length >= 0 {
		f.find(test, 0, start)
		length = f.find(body, 1, start)
	} else {
		length = f.find(test, 0, start)
	}

	if length < 0 {
		f.prevWasMissing = oldPrevWasMissing
		f.missingCount = oldMissingCount
	}

	return length
}

func (f *annotatedParserFinder[TInput]) visitLimit(limiter Parser[TInput], limited Parser[TInput], start int) int {
	length := limiter.Scan(f.source, start)
	if length > 0 {
		oldSource := f.source
		f.source = NewLimitSource(f.source, start+length)
		result := f.find(limited, 0, start)
		f.source = oldSource
		return result
	}
	return -1
}

func (f *annotatedParserFinder[TInput]) visitNot(pattern Parser[TInput], start int) int {
	f.prevWasMissing = false
	if f.source.IsEnd(start) {
		return 0
	}
	length := pattern.Scan(f.source, start)
	if length >= 0 {
		return -1
	}
	return 1
}

func (f *annotatedParserFinder[TInput]) visitOneOrMore(repeated Parser[TInput], start int) int {
	length := f.find(repeated, 0, start)
	if length < 0 {
		return length
	}

	totalLength := 0
	for length > 0 {
		totalLength += length
		length = f.find(repeated, 0, start+totalLength)
	}
	return totalLength
}

func (f *annotatedParserFinder[TInput]) visitOptional(optional Parser[TInput], start int) int {
	oldPrevWasMissing := f.prevWasMissing
	length := f.find(optional, 0, start)
	if length < 0 {
		f.prevWasMissing = oldPrevWasMissing
		return 0
	}
	return length
}

func (f *annotatedParserFinder[TInput]) visitRequired(required Parser[TInput], start int) int {
	length := f.find(required, 0, start)
	if length < 0 {
		f.prevWasMissing = true
		f.missingCount++
		return 0
	}
	return length
}

func (f *annotatedParserFinder[TInput]) visitSequence(parsers []Parser[TInput], start int) int {
	totalLength := 0
	for i, parser := range parsers {
		length := f.find(parser, i, start+totalLength)
		if length < 0 {
			return length - totalLength
		}
		totalLength += length
	}
	return totalLength
}

func (f *annotatedParserFinder[TInput]) visitZeroOrMore(repeated Parser[TInput], start int, zeroOrOne bool) int {
	totalLength := 0
	if zeroOrOne {
		length := f.find(repeated, 0, start)
		if length > 0 {
			totalLength += length
			length = f.find(repeated, 0, start+length)
			if length > 0 {
				totalLength += length
			}
		}
		return totalLength
	}

	length := f.find(repeated, 0, start)
	for length > 0 {
		totalLength += length
		length = f.find(repeated, 0, start+totalLength)
	}
	return totalLength
}
