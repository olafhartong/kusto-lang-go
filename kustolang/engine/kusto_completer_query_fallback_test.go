package kustolang

import (
	"slices"
	"testing"
)

func TestKustoCompleterPostPipeFallbackInventory(t *testing.T) {
	want := []SyntaxKind{
		SyntaxKindAsKeyword,
		SyntaxKindAssertSchemaKeyword,
		SyntaxKindConsumeKeyword,
		SyntaxKindCountKeyword,
		SyntaxKindDistinctKeyword,
		SyntaxKindExtendKeyword,
		SyntaxKindEvaluateKeyword,
		SyntaxKindExecuteAndCacheKeyword,
		SyntaxKindFacetKeyword,
		SyntaxKindForkKeyword,
		SyntaxKindFilterKeyword,
		SyntaxKindWhereKeyword,
		SyntaxKindGetSchemaKeyword,
		SyntaxKindInvokeKeyword,
		SyntaxKindJoinKeyword,
		SyntaxKindLookupKeyword,
		SyntaxKindMakeGraphKeyword,
		SyntaxKindMakeSeriesKeyword,
		SyntaxKindGraphMatchKeyword,
		SyntaxKindGraphMarkComponentsKeyword,
		SyntaxKindGraphShortestPathsKeyword,
		SyntaxKindGraphToTableKeyword,
		SyntaxKindMvDashApplyKeyword,
		SyntaxKindMvApplyKeyword,
		SyntaxKindMvDashExpandKeyword,
		SyntaxKindMvExpandKeyword,
		SyntaxKindParseKeyword,
		SyntaxKindParseKvKeyword,
		SyntaxKindParseWhereKeyword,
		SyntaxKindPartitionKeyword,
		SyntaxKindPartitionByKeyword,
		SyntaxKindProjectKeyword,
		SyntaxKindProjectAwayKeyword,
		SyntaxKindProjectByNamesKeyword,
		SyntaxKindProjectKeepKeyword,
		SyntaxKindProjectRenameKeyword,
		SyntaxKindProjectReorderKeyword,
		SyntaxKindReduceKeyword,
		SyntaxKindSampleKeyword,
		SyntaxKindSampleDistinctKeyword,
		SyntaxKindRenderKeyword,
		SyntaxKindScanKeyword,
		SyntaxKindSearchKeyword,
		SyntaxKindSerializeKeyword,
		SyntaxKindSortKeyword,
		SyntaxKindOrderKeyword,
		SyntaxKindSummarizeKeyword,
		SyntaxKindTakeKeyword,
		SyntaxKindLimitKeyword,
		SyntaxKindTopHittersKeyword,
		SyntaxKindTopNestedKeyword,
		SyntaxKindTopKeyword,
		SyntaxKindUnionKeyword,
	}

	if !slices.Equal(queryPipedOperatorCompletionKinds, want) {
		t.Fatalf("post-pipe fallback inventory = %#v, want %#v", queryPipedOperatorCompletionKinds, want)
	}
}
