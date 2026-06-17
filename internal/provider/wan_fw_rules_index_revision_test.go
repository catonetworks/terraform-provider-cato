package provider

import (
	"errors"
	"testing"
)

func TestWanReorderPolicyConflictMessage_transportError(t *testing.T) {
	t.Parallel()

	err := errors.New("Cannot reorder policy while other active revisions exist")
	if got := wanReorderPolicyConflictMessage(nil, err); got != err.Error() {
		t.Fatalf("expected transport error text, got %q", got)
	}
}

func TestWanReorderPolicyConflictMessage_payloadError(t *testing.T) {
	t.Parallel()

	msg := wanReorderPolicyConflictMessage(
		wanReorderResponseWithError("Cannot reorder policy while other active revisions exist"),
		nil,
	)
	if msg == "" {
		t.Fatal("expected payload error message")
	}
	if !isActiveRevisionConflict(msg) {
		t.Fatalf("expected active revision conflict, got %q", msg)
	}
}

func TestWanReorderPolicyConflictMessage_success(t *testing.T) {
	t.Parallel()

	if got := wanReorderPolicyConflictMessage(wanReorderResponseSuccess(), nil); got != "" {
		t.Fatalf("expected empty message on success, got %q", got)
	}
}
