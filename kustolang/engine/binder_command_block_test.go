package kustolang

import "testing"

func commandResultsReference(t *testing.T, root SyntaxElement) *NameReference {
	t.Helper()
	return namedReference(t, root, "$command_results")
}

func namedReference(t *testing.T, root SyntaxElement, name string) *NameReference {
	t.Helper()
	var reference *NameReference
	WalkSyntax(root, func(element SyntaxElement) {
		if reference != nil {
			return
		}
		candidate, ok := WrapGeneratedSyntaxElement(element).(*NameReference)
		if ok && NameReferenceSimpleName(candidate) == name {
			reference = candidate
		}
	})
	if reference == nil {
		t.Fatalf("name reference %q not found", name)
	}
	return reference
}
