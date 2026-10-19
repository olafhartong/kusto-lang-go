package kustolang

import "testing"

func TestCurrentMakeGraphPartitionEntityDoesNotSilentlyConsumeDottedPath(t *testing.T) {
	const text = `Edges | make-graph Source --> Target partitioned-by Tenant.Region (where Active | count)`
	query := ParseQuery(text)
	operator := firstBoundNode[*MakeGraphOperator](t, query)
	clause := operator.PartitionedByClause()
	if clause == nil || NameReferenceSimpleName(clause.Entity()) != "Tenant" || clause.Entity().Width() == 0 {
		t.Fatalf("make-graph partition entity = %#v", clause)
	}
	if !diagnosticsContain(query.ContainedDiagnostics(), "KS005") || !diagnosticsContain(query.ContainedDiagnostics(), "KS198") {
		t.Fatalf("dotted partition entity diagnostics = %#v", query.ContainedDiagnostics())
	}
	if skipped := query.GetChild(2); skipped == nil || skipped.Width() == 0 {
		t.Fatalf("dotted partition entity tail was silently discarded: %#v", skipped)
	}
}

func TestCurrentMakeGraphPartitionEntityMissingNameStillRecoversSubquery(t *testing.T) {
	const text = `Edges | make-graph Source --> Target partitioned-by (where Active | count)`
	query := ParseQuery(text)
	operator := firstBoundNode[*MakeGraphOperator](t, query)
	clause := operator.PartitionedByClause()
	if clause == nil || clause.Entity().Width() != 0 || clause.Subquery() == nil || clause.Subquery().Width() == 0 {
		t.Fatalf("missing partition entity recovery = %#v", clause)
	}
	if !diagnosticsContain(clause.Entity().ContainedDiagnostics(), "KS006") {
		t.Fatalf("missing partition entity diagnostics = %#v", clause.Entity().ContainedDiagnostics())
	}
}
