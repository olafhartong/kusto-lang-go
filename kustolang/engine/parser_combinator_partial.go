package kustolang

import (
	"reflect"
	"strconv"
	"strings"
)

// Ported from src/Kusto.Language/Parser/Combinators/PartialParser.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type PartialParserFailureHandler[TInput any] func(parser Parser[TInput], output *[]any)

func ScanPartial[TInput any](parser Parser[TInput], input Source[TInput], inputStart int) int {
	pathFinder := partialPathFinder[TInput]{}
	path := pathFinder.findPath(parser, input, inputStart)
	if path != nil {
		return path.inputLength
	}
	return -1
}

func ParsePartial[TInput any](parser Parser[TInput], input Source[TInput], inputStart int, output *[]any, outputStart int, onFailure ...PartialParserFailureHandler[TInput]) int {
	pathFinder := partialPathFinder[TInput]{}
	path := pathFinder.findPath(parser, input, inputStart)
	return parsePartialPath(path, input, inputStart, output, outputStart, firstPartialFailureHandler(onFailure))
}

func ScanPartialBest[TInput any](parsers []Parser[TInput], input Source[TInput], inputStart int) int {
	info := getPartialBestPath(parsers, input, inputStart)
	if info != nil && info.bestPath != nil {
		return info.bestPath.inputLength
	}
	return -1
}

func ParsePartialBest[TInput any](parsers []Parser[TInput], input Source[TInput], inputStart int, output *[]any, outputStart int, bestParsers *[]Parser[TInput], onFailure ...PartialParserFailureHandler[TInput]) int {
	best := getPartialBestPath(parsers, input, inputStart)
	if best == nil {
		return -1
	}
	if bestParsers != nil {
		*bestParsers = append(*bestParsers, best.bestParsers...)
	}
	return parsePartialPath(best.bestPath, input, inputStart, output, outputStart, firstPartialFailureHandler(onFailure))
}

func firstPartialFailureHandler[TInput any](handlers []PartialParserFailureHandler[TInput]) PartialParserFailureHandler[TInput] {
	if len(handlers) == 0 {
		return nil
	}
	return handlers[0]
}

type partialMemoizedInfo[TInput any] struct {
	bestPaths map[string]*partialBestPathInfo[TInput]
}

type partialBestPathInfo[TInput any] struct {
	bestPath    *partialPath[TInput]
	bestParsers []Parser[TInput]
}

func getPartialBestPath[TInput any](parsers []Parser[TInput], source Source[TInput], inputStart int) *partialBestPathInfo[TInput] {
	if source == nil {
		pathFinder := partialPathFinder[TInput]{}
		bestParsers := []Parser[TInput]{}
		return &partialBestPathInfo[TInput]{
			bestPath:    pathFinder.findBestPath(parsers, source, inputStart, &bestParsers),
			bestParsers: bestParsers,
		}
	}

	cached := SourceCacheGetOrCreate[*partialMemoizedInfo[TInput]](source.Cache())
	if cached.bestPaths == nil {
		cached.bestPaths = map[string]*partialBestPathInfo[TInput]{}
	}

	key := partialBestPathKey(parsers, inputStart)
	if info, ok := cached.bestPaths[key]; ok {
		return info
	}

	pathFinder := partialPathFinder[TInput]{}
	bestParsers := []Parser[TInput]{}
	path := pathFinder.findBestPath(parsers, source, inputStart, &bestParsers)
	info := &partialBestPathInfo[TInput]{
		bestPath:    path,
		bestParsers: bestParsers,
	}
	cached.bestPaths[key] = info
	return info
}

func partialBestPathKey[TInput any](parsers []Parser[TInput], inputStart int) string {
	var builder strings.Builder
	builder.WriteString(strconv.Itoa(inputStart))
	for _, parser := range parsers {
		builder.WriteByte('|')
		builder.WriteString(strconv.FormatUint(uint64(parserIdentity(parser)), 16))
	}
	return builder.String()
}

func parserIdentity(value any) uintptr {
	reflected := reflect.ValueOf(value)
	if !reflected.IsValid() {
		return 0
	}
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return reflected.Pointer()
	default:
		return 0
	}
}

func parsePartialPath[TInput any](path *partialPath[TInput], input Source[TInput], inputStart int, output *[]any, outputStart int, onFailure PartialParserFailureHandler[TInput]) int {
	_ = outputStart
	if path == nil {
		return -1
	}

	inputPosition := inputStart
	for _, step := range partialPathSteps(path) {
		if step.isError {
			if onFailure != nil {
				onFailure(step.parser, output)
			}
			continue
		}
		inputPosition += step.parser.ParseList(input, inputPosition, output, outputCount(output))
	}
	return inputPosition - inputStart
}

func partialPathSteps[TInput any](path *partialPath[TInput]) []*partialPath[TInput] {
	steps := make([]*partialPath[TInput], path.length)
	for path != nil {
		steps[path.length-1] = path
		path = path.previous
	}
	return steps
}

type partialPathFinder[TInput any] struct{}

func (f partialPathFinder[TInput]) findPath(parser Parser[TInput], source Source[TInput], inputStart int) *partialPath[TInput] {
	return f.findBestPath([]Parser[TInput]{parser}, source, inputStart, nil)
}

func (f partialPathFinder[TInput]) findBestPath(parsers []Parser[TInput], source Source[TInput], inputStart int, bestParsers *[]Parser[TInput]) *partialPath[TInput] {
	input := partialScanInput[TInput]{source: source, inputStart: inputStart}
	results := make([]partialPathResult[TInput], 0, len(parsers))
	for _, parser := range parsers {
		output := f.accept(parser, input)
		for _, path := range output.pathsList() {
			results = append(results, partialPathResult[TInput]{parser: parser, path: path})
		}
	}
	if len(results) == 0 {
		return nil
	}

	maxInput := results[0].path.inputLengthValue()
	for _, result := range results[1:] {
		if result.path.inputLengthValue() > maxInput {
			maxInput = result.path.inputLengthValue()
		}
	}

	bestPaths := make([]*partialPath[TInput], 0, len(results))
	if bestParsers != nil {
		seen := map[uintptr]bool{}
		for _, result := range results {
			if result.path.inputLengthValue() == maxInput {
				id := parserIdentity(result.parser)
				if !seen[id] {
					seen[id] = true
					*bestParsers = append(*bestParsers, result.parser)
				}
				bestPaths = append(bestPaths, result.path)
			}
		}
	} else {
		for _, result := range results {
			if result.path.inputLengthValue() == maxInput {
				bestPaths = append(bestPaths, result.path)
			}
		}
	}

	return getBestPartialPath(bestPaths, nil)
}

func (f partialPathFinder[TInput]) accept(parser Parser[TInput], input partialScanInput[TInput]) partialScanOutput[TInput] {
	return AcceptParserVisitorResult(parser, ParserVisitorResultFuncs[TInput, partialScanOutput[TInput]]{
		Apply: func(parser Parser[TInput], left Parser[TInput], right Parser[TInput], kind ApplyKind) partialScanOutput[TInput] {
			return f.visitApply(parser, left, right, kind, input)
		},
		Best: func(parser Parser[TInput], parsers []Parser[TInput]) partialScanOutput[TInput] {
			return f.scanBest(parser, parsers, input)
		},
		Convert: func(parser Parser[TInput], pattern Parser[TInput]) partialScanOutput[TInput] {
			return f.scanParser(parser, pattern, input)
		},
		Fails: func(parser Parser[TInput], _ Parser[TInput]) partialScanOutput[TInput] {
			return f.scanParser(parser, parser, input)
		},
		First: func(parser Parser[TInput], parsers []Parser[TInput]) partialScanOutput[TInput] {
			return f.scanFirst(parser, parsers, input)
		},
		Forward: func(parser Parser[TInput], deferred Parser[TInput]) partialScanOutput[TInput] {
			return f.scanParser(parser, deferred, input)
		},
		If: func(parser Parser[TInput], test Parser[TInput], body Parser[TInput]) partialScanOutput[TInput] {
			return f.scanIf(parser, test, body, input)
		},
		Limit: func(parser Parser[TInput], limiter Parser[TInput], limited Parser[TInput]) partialScanOutput[TInput] {
			return f.visitLimit(parser, limiter, limited, input)
		},
		Map: func(parser Parser[TInput]) partialScanOutput[TInput] {
			return f.scanParser(parser, parser, input)
		},
		Match: func(parser Parser[TInput]) partialScanOutput[TInput] {
			return f.scanParser(parser, parser, input)
		},
		Not: func(parser Parser[TInput], _ Parser[TInput]) partialScanOutput[TInput] {
			return f.scanParser(parser, parser, input)
		},
		OneOrMore: func(parser Parser[TInput], repeated Parser[TInput]) partialScanOutput[TInput] {
			return f.scanRepeatOuter(parser, repeated, 1, partialMaxInt, input)
		},
		Optional: func(parser Parser[TInput], optional Parser[TInput]) partialScanOutput[TInput] {
			return f.scanOptional(parser, optional, input)
		},
		Produce: func(parser Parser[TInput], inner Parser[TInput]) partialScanOutput[TInput] {
			return adjustPartialPaths(f.accept(inner, input), parser, inner, 1)
		},
		Required: func(parser Parser[TInput], required Parser[TInput]) partialScanOutput[TInput] {
			return f.scanOptional(parser, required, input)
		},
		Rule: func(parser Parser[TInput], parsers []Parser[TInput]) partialScanOutput[TInput] {
			return f.scanSequence(parser, parsers, input)
		},
		Sequence: func(parser Parser[TInput], parsers []Parser[TInput]) partialScanOutput[TInput] {
			return f.scanSequence(parser, parsers, input)
		},
		ZeroOrMore: func(parser Parser[TInput], repeated Parser[TInput], zeroOrOne bool) partialScanOutput[TInput] {
			max := partialMaxInt
			if zeroOrOne {
				max = 1
			}
			return f.scanRepeatOuter(parser, repeated, 0, max, input)
		},
		Unsupported: func(parser Parser[TInput]) partialScanOutput[TInput] {
			return f.scanParser(parser, parser, input)
		},
	})
}

func (f partialPathFinder[TInput]) scanParser(parser Parser[TInput], scanned Parser[TInput], input partialScanInput[TInput]) partialScanOutput[TInput] {
	if scanned == nil {
		return newPartialScanOutput(input, input.path)
	}
	length := scanned.Scan(input.source, input.inputStart+input.path.inputLengthValue())
	if length > 0 {
		return newPartialScanOutput(input, input.path.add(parser, length))
	}
	return newPartialScanOutput(input, input.path)
}

func (f partialPathFinder[TInput]) scanIf(outer Parser[TInput], test Parser[TInput], inner Parser[TInput], input partialScanInput[TInput]) partialScanOutput[TInput] {
	length := test.Scan(input.source, input.inputStart+input.path.inputLengthValue())
	if length > 0 {
		return adjustPartialPaths(f.accept(inner, input), outer, inner, 1)
	}
	return newPartialScanOutput(input, input.path)
}

func (f partialPathFinder[TInput]) visitLimit(outer Parser[TInput], limiter Parser[TInput], limited Parser[TInput], input partialScanInput[TInput]) partialScanOutput[TInput] {
	length := limiter.Scan(input.source, input.inputStart+input.path.inputLengthValue())
	if length >= 0 {
		limitSource := NewLimitSource(input.source, input.inputStart+length)
		return adjustPartialPaths(f.accept(limited, input.withSource(limitSource)), outer, limited, 1)
	}
	return newPartialScanOutput(input, input.path)
}

func (f partialPathFinder[TInput]) scanSequence(outer Parser[TInput], stepParsers []Parser[TInput], input partialScanInput[TInput]) partialScanOutput[TInput] {
	if len(stepParsers) == 0 {
		return newPartialScanOutput(input, input.path.add(outer, 0))
	}
	paths := []*partialPath[TInput]{}
	f.scanSequenceSteps(stepParsers, 0, input, input.path, &paths)
	best := getBestPartialPath(paths, input.path)
	best = adjustPartialPathForSequence(outer, stepParsers, input.path, best)
	return newPartialScanOutput(input, best)
}

func (f partialPathFinder[TInput]) scanSequenceSteps(stepParsers []Parser[TInput], stepIndex int, initialInput partialScanInput[TInput], path *partialPath[TInput], resultPaths *[]*partialPath[TInput]) {
	stepParser := stepParsers[stepIndex]
	stepOutput := f.accept(stepParser, initialInput.withPath(path))
	if stepOutput.hasMultiplePaths() {
		for _, outputPath := range stepOutput.pathsList() {
			if stepIndex < len(stepParsers)-1 && isPartialSuccess(path, outputPath) {
				f.scanSequenceSteps(stepParsers, stepIndex+1, initialInput, outputPath, resultPaths)
			} else {
				*resultPaths = append(*resultPaths, addPartialSequenceError(stepParser, initialInput.path, stepOutput.input.path, outputPath))
			}
		}
		return
	}

	stepPath := stepOutput.pathValue()
	if stepIndex < len(stepParsers)-1 && isPartialSuccess(path, stepPath) {
		f.scanSequenceSteps(stepParsers, stepIndex+1, initialInput, stepPath, resultPaths)
	} else {
		*resultPaths = append(*resultPaths, addPartialSequenceError(stepParser, initialInput.path, stepOutput.input.path, stepPath))
	}
}

func addPartialSequenceError[TInput any](parser Parser[TInput], initialPath *partialPath[TInput], previousStepPath *partialPath[TInput], finalPath *partialPath[TInput]) *partialPath[TInput] {
	if finalPath != initialPath && finalPath == previousStepPath && !finalPath.isErrorValue() {
		return finalPath.addError(parser)
	}
	return finalPath
}

func (f partialPathFinder[TInput]) scanOptional(outer Parser[TInput], inner Parser[TInput], input partialScanInput[TInput]) partialScanOutput[TInput] {
	result := adjustPartialPaths(f.accept(inner, input), outer, inner, 1)
	if !result.hadSuccess() {
		result = result.addPath(input.path.add(outer, 0))
	}
	return result
}

func (f partialPathFinder[TInput]) scanRepeatOuter(outer Parser[TInput], inner Parser[TInput], min int, max int, input partialScanInput[TInput]) partialScanOutput[TInput] {
	result, maxSuccesses, maxSuccessInputLength := f.scanRepeat(inner, min, max, input, newPartialScanOutput(input, input.path))
	result = adjustPartialPaths(result, outer, inner, max)
	if maxSuccesses >= min && !result.hadSuccess() {
		result = result.addPath(input.path.add(outer, maxSuccessInputLength-input.path.inputLengthValue()))
	}
	return result
}

func (f partialPathFinder[TInput]) scanRepeat(inner Parser[TInput], min int, max int, input partialScanInput[TInput], previous partialScanOutput[TInput]) (partialScanOutput[TInput], int, int) {
	previousOutputs := []partialScanOutput[TInput]{previous}
	newOutputs := []partialScanOutput[TInput]{}
	resultPaths := []*partialPath[TInput]{}

	maxSuccessInputLength := input.path.inputLengthValue()
	maxSuccesses := 0
	for i := 0; i <= max && len(previousOutputs) > 0; i++ {
		newOutputs = newOutputs[:0]
		for _, prevOutput := range previousOutputs {
			if prevOutput.hasMultiplePaths() {
				for _, path := range prevOutput.pathsList() {
					if i == 0 || (i < max && isPartialSuccess(prevOutput.input.path, path)) {
						output := f.accept(inner, input.withPath(path))
						newOutputs = append(newOutputs, output)
						if output.hadSuccess() {
							maxSuccesses = i + 1
							if successInputLength := output.maxSuccessInputLength(); successInputLength > maxSuccessInputLength {
								maxSuccessInputLength = successInputLength
							}
						}
					} else {
						resultPath := path
						if i <= min {
							resultPath = addPartialSequenceError(inner, input.path, prevOutput.input.path, path)
						}
						resultPaths = append(resultPaths, resultPath)
					}
				}
			} else if i == 0 || (i < max && isPartialSuccess(prevOutput.input.path, prevOutput.pathValue())) {
				output := f.accept(inner, input.withPath(prevOutput.pathValue()))
				newOutputs = append(newOutputs, output)
				if output.hadSuccess() {
					maxSuccesses = i + 1
					if successInputLength := output.maxSuccessInputLength(); successInputLength > maxSuccessInputLength {
						maxSuccessInputLength = successInputLength
					}
				}
			} else {
				resultPath := prevOutput.pathValue()
				if i <= min {
					resultPath = addPartialSequenceError(inner, input.path, prevOutput.input.path, prevOutput.pathValue())
				}
				resultPaths = append(resultPaths, resultPath)
			}
		}
		previousOutputs, newOutputs = newOutputs, previousOutputs
	}

	return newPartialScanOutputWithPaths(input, resultPaths), maxSuccesses, maxSuccessInputLength
}

func (f partialPathFinder[TInput]) scanBest(outer Parser[TInput], alternates []Parser[TInput], input partialScanInput[TInput]) partialScanOutput[TInput] {
	candidates := []*partialPath[TInput]{}
	for _, alternate := range alternates {
		output := adjustPartialPaths(f.accept(alternate, input), outer, alternate, 1)
		candidates = output.getPaths(candidates)
	}
	return newPartialScanOutput(input, getBestPartialPath(candidates, input.path))
}

func (f partialPathFinder[TInput]) scanFirst(outer Parser[TInput], alternates []Parser[TInput], input partialScanInput[TInput]) partialScanOutput[TInput] {
	candidates := []*partialPath[TInput]{}
	for _, alternate := range alternates {
		output := adjustPartialPaths(f.accept(alternate, input), outer, alternate, 1)
		candidates = output.getPaths(candidates)
	}
	var first *partialPath[TInput]
	for _, candidate := range candidates {
		if isPartialSuccess(input.path, candidate) {
			first = candidate
			break
		}
	}
	if first == nil {
		first = getBestPartialPath(candidates, input.path)
	}
	return newPartialScanOutput(input, first)
}

func (f partialPathFinder[TInput]) visitApply(outer Parser[TInput], left Parser[TInput], right Parser[TInput], kind ApplyKind, input partialScanInput[TInput]) partialScanOutput[TInput] {
	leftOutput := f.accept(left, input)
	if !leftOutput.hadSuccess() {
		return leftOutput
	}

	min := 0
	if kind == ApplyKindOne {
		min = 1
	}
	max := partialMaxInt
	if kind == ApplyKindOne || kind == ApplyKindZeroOrOne {
		max = 1
	}

	result, maxSuccesses, maxSuccessInputLength := f.scanRepeat(right, min, max, input, leftOutput)
	result = adjustPartialApplyPaths(result, outer, left, right, max)
	if maxSuccesses >= min && !result.hadSuccess() {
		result = result.addPath(input.path.add(outer, maxSuccessInputLength-input.path.inputLengthValue()))
	}
	return result
}

func getBestPartialPath[TInput any](paths []*partialPath[TInput], input *partialPath[TInput]) *partialPath[TInput] {
	if len(paths) == 0 {
		return nil
	}
	maxInput := paths[0].inputLengthValue()
	for _, path := range paths[1:] {
		if path.inputLengthValue() > maxInput {
			maxInput = path.inputLengthValue()
		}
	}

	bestPaths := make([]*partialPath[TInput], 0, len(paths))
	for _, path := range paths {
		if path.inputLengthValue() == maxInput {
			bestPaths = append(bestPaths, path)
		}
	}

	if len(bestPaths) == 1 {
		return bestPaths[0]
	}
	for _, path := range bestPaths {
		if isPartialSuccess(input, path) {
			return path
		}
	}
	for _, path := range bestPaths {
		if !path.isErrorValue() {
			return path
		}
	}

	parsers := make([]Parser[TInput], 0, len(bestPaths))
	for _, path := range bestPaths {
		parsers = append(parsers, path.parser)
	}
	return bestPaths[0].previous.addError(NewBestParser(parsers))
}

func adjustPartialPaths[TInput any](output partialScanOutput[TInput], outer Parser[TInput], inner Parser[TInput], maxRepeats int) partialScanOutput[TInput] {
	if output.hasMultiplePaths() {
		outputPaths := make([]*partialPath[TInput], 0, len(output.paths))
		for _, path := range output.pathsList() {
			outputPaths = append(outputPaths, adjustPartialPath(outer, inner, output.input.path, path, maxRepeats))
		}
		return output.withPaths(outputPaths)
	}
	return output.withPath(adjustPartialPath(outer, inner, output.input.path, output.pathValue(), maxRepeats))
}

func adjustPartialPath[TInput any](outer Parser[TInput], inner Parser[TInput], initial *partialPath[TInput], result *partialPath[TInput], maxRepeats int) *partialPath[TInput] {
	if !result.isErrorValue() {
		n := 0
		path := result
		for path != nil && path.parser == inner {
			path = path.previous
			n++
		}
		if path == initial && n > 0 && n <= maxRepeats {
			return initial.add(outer, result.inputLength-initial.inputLengthValue())
		}
	}
	return result
}

func adjustPartialPathForSequence[TInput any](outer Parser[TInput], inners []Parser[TInput], initial *partialPath[TInput], result *partialPath[TInput]) *partialPath[TInput] {
	if result == nil {
		return nil
	}
	if result.length == initial.pathLength()+len(inners) && !result.isError {
		path := result
		for i := len(inners) - 1; i >= 0; i-- {
			if path == nil || path.parser != inners[i] {
				return result
			}
			path = path.previous
		}
		return initial.add(outer, result.inputLength-initial.inputLengthValue())
	}
	return result
}

func adjustPartialApplyPaths[TInput any](output partialScanOutput[TInput], outer Parser[TInput], left Parser[TInput], right Parser[TInput], maxRight int) partialScanOutput[TInput] {
	if output.hasMultiplePaths() {
		outputPaths := make([]*partialPath[TInput], 0, len(output.paths))
		for _, path := range output.pathsList() {
			if isPartialSuccess(output.input.path, path) {
				outputPaths = append(outputPaths, adjustPartialApplyPath(outer, left, right, output.input.path, path, maxRight))
			}
		}
		return output.withPaths(outputPaths)
	}
	return output.withPath(adjustPartialApplyPath(outer, left, right, output.input.path, output.pathValue(), maxRight))
}

func adjustPartialApplyPath[TInput any](outer Parser[TInput], left Parser[TInput], right Parser[TInput], initial *partialPath[TInput], result *partialPath[TInput], maxRight int) *partialPath[TInput] {
	if !result.isErrorValue() {
		nRight := 0
		path := result
		for path != nil && path.parser == right {
			path = path.previous
			nRight++
		}
		if path != nil && path.parser == left && path.previous == initial && nRight <= maxRight {
			return initial.add(outer, result.inputLength-initial.inputLengthValue())
		}
	}
	return result
}

func isPartialSuccess[TInput any](input *partialPath[TInput], output *partialPath[TInput]) bool {
	return output != input && !output.isErrorValue()
}

func hasPartialProgress[TInput any](input *partialPath[TInput], output *partialPath[TInput]) bool {
	return output != input
}

type partialPathResult[TInput any] struct {
	parser Parser[TInput]
	path   *partialPath[TInput]
}

type partialPath[TInput any] struct {
	previous    *partialPath[TInput]
	parser      Parser[TInput]
	length      int
	inputLength int
	isError     bool
}

func (p *partialPath[TInput]) add(parser Parser[TInput], inputLength int) *partialPath[TInput] {
	return &partialPath[TInput]{
		previous:    p,
		parser:      parser,
		length:      p.pathLength() + 1,
		inputLength: p.inputLengthValue() + inputLength,
	}
}

func (p *partialPath[TInput]) addError(parser Parser[TInput]) *partialPath[TInput] {
	return &partialPath[TInput]{
		previous:    p,
		parser:      parser,
		length:      p.pathLength() + 1,
		inputLength: p.inputLengthValue(),
		isError:     true,
	}
}

func (p *partialPath[TInput]) pathLength() int {
	if p == nil {
		return 0
	}
	return p.length
}

func (p *partialPath[TInput]) inputLengthValue() int {
	if p == nil {
		return 0
	}
	return p.inputLength
}

func (p *partialPath[TInput]) isErrorValue() bool {
	return p != nil && p.isError
}

type partialScanInput[TInput any] struct {
	source     Source[TInput]
	inputStart int
	path       *partialPath[TInput]
}

func (i partialScanInput[TInput]) withSource(source Source[TInput]) partialScanInput[TInput] {
	i.source = source
	return i
}

func (i partialScanInput[TInput]) withPath(path *partialPath[TInput]) partialScanInput[TInput] {
	i.path = path
	return i
}

type partialScanOutput[TInput any] struct {
	input partialScanInput[TInput]
	path  *partialPath[TInput]
	paths []*partialPath[TInput]
}

func newPartialScanOutput[TInput any](input partialScanInput[TInput], path *partialPath[TInput]) partialScanOutput[TInput] {
	return partialScanOutput[TInput]{input: input, path: path}
}

func newPartialScanOutputWithPaths[TInput any](input partialScanInput[TInput], paths []*partialPath[TInput]) partialScanOutput[TInput] {
	return partialScanOutput[TInput]{input: input, paths: append([]*partialPath[TInput]{}, paths...)}
}

func (o partialScanOutput[TInput]) pathValue() *partialPath[TInput] {
	if o.paths != nil && len(o.paths) == 1 {
		return o.paths[0]
	}
	return o.path
}

func (o partialScanOutput[TInput]) pathsList() []*partialPath[TInput] {
	if o.paths != nil {
		return o.paths
	}
	return []*partialPath[TInput]{o.path}
}

func (o partialScanOutput[TInput]) hasMultiplePaths() bool {
	return len(o.paths) > 1
}

func (o partialScanOutput[TInput]) madeProgress() bool {
	for _, path := range o.pathsList() {
		if hasPartialProgress(o.input.path, path) {
			return true
		}
	}
	return false
}

func (o partialScanOutput[TInput]) hadSuccess() bool {
	for _, path := range o.pathsList() {
		if isPartialSuccess(o.input.path, path) {
			return true
		}
	}
	return false
}

func (o partialScanOutput[TInput]) maxSuccessInputLength() int {
	maxInput := o.input.path.pathLength()
	for _, path := range o.pathsList() {
		if isPartialSuccess(o.input.path, path) && path.inputLengthValue() > maxInput {
			maxInput = path.inputLengthValue()
		}
	}
	return maxInput
}

func (o partialScanOutput[TInput]) withPath(path *partialPath[TInput]) partialScanOutput[TInput] {
	return newPartialScanOutput(o.input, path)
}

func (o partialScanOutput[TInput]) withPaths(paths []*partialPath[TInput]) partialScanOutput[TInput] {
	return newPartialScanOutputWithPaths(o.input, paths)
}

func (o partialScanOutput[TInput]) addPath(path *partialPath[TInput]) partialScanOutput[TInput] {
	return o.withPaths(distinctPartialPaths(append(o.pathsList(), path)))
}

func (o partialScanOutput[TInput]) addPaths(paths []*partialPath[TInput]) partialScanOutput[TInput] {
	return o.withPaths(distinctPartialPaths(append(o.pathsList(), paths...)))
}

func (o partialScanOutput[TInput]) getPaths(paths []*partialPath[TInput]) []*partialPath[TInput] {
	return append(paths, o.pathsList()...)
}

func distinctPartialPaths[TInput any](paths []*partialPath[TInput]) []*partialPath[TInput] {
	seen := map[*partialPath[TInput]]bool{}
	result := make([]*partialPath[TInput], 0, len(paths))
	for _, path := range paths {
		if !seen[path] {
			seen[path] = true
			result = append(result, path)
		}
	}
	return result
}

const partialMaxInt = int(^uint(0) >> 1)
