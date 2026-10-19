package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestCustomNodeGetDefaultShape(t *testing.T) {
	shape := kustolang.CustomNodeGetDefaultShape(3)
	if len(shape) != 3 {
		t.Fatalf("shape length = %d, want 3", len(shape))
	}
	for i, descriptor := range shape {
		if descriptor != kustolang.CustomElementDescriptorDefault || !descriptor.IsDefault() {
			t.Fatalf("shape[%d] = %#v, want default descriptor", i, descriptor)
		}
	}
}

func TestCustomElementDescriptorFromConveniences(t *testing.T) {
	if got := kustolang.CustomElementDescriptorFrom(""); got != kustolang.CustomElementDescriptorDefault {
		t.Fatalf("From default = %#v, want %#v", got, kustolang.CustomElementDescriptorDefault)
	}

	named := kustolang.CustomElementDescriptorFrom("Target", kustolang.CustomElementDescriptorFromOptions{
		Hint:       kustolang.CompletionHintTable,
		IsOptional: true,
	})
	if named.Name != "Target" || named.Hint != kustolang.CompletionHintTable || !named.Optional {
		t.Fatalf("From named = %#v", named)
	}

	hinted := kustolang.CustomElementDescriptorFromHint(kustolang.CompletionHintColumn)
	if hinted.Name != "" || hinted.Hint != kustolang.CompletionHintColumn || hinted.Optional {
		t.Fatalf("From hint = %#v", hinted)
	}

	optional := kustolang.CustomElementDescriptorFromIsOptional(true)
	if optional.Name != "" || optional.Hint != kustolang.CompletionHintSyntax || !optional.Optional {
		t.Fatalf("From isOptional = %#v", optional)
	}
}

func TestCustomElementDescriptorWithIsOptionalPreservesPinnedBug(t *testing.T) {
	descriptor := kustolang.CustomElementDescriptorFrom("Target")

	// Pinned CustomNode.cs constructs the changed value with this.IsOptional
	// instead of the argument. Keep WithIsOptional bug-compatible; callers that
	// want to change the value retain the existing behavior-correct WithOptional.
	if got := descriptor.WithIsOptional(true); got.Optional {
		t.Fatalf("WithIsOptional(true) = %#v, want pinned no-op behavior", got)
	}
	if got := descriptor.WithOptional(true); !got.Optional {
		t.Fatalf("WithOptional(true) = %#v, want optional descriptor", got)
	}
}
