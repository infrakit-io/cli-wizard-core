package wizard

import (
	"errors"
	"strings"
	"testing"
)

func TestNewUserError(t *testing.T) {
	err := NewUserError("invalid input", "provide a value")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); got != "invalid input" {
		t.Fatalf("unexpected error message: %q", got)
	}
	if got := ErrorHint(err); got != "provide a value" {
		t.Fatalf("unexpected hint: %q", got)
	}
}

func TestWithHint(t *testing.T) {
	base := errors.New("boom")
	err := WithHint(base, "retry with --dry-run")
	if err == nil {
		t.Fatal("expected wrapped error, got nil")
	}
	if got := err.Error(); got != "boom" {
		t.Fatalf("unexpected wrapped error message: %q", got)
	}
	if got := ErrorHint(err); got != "retry with --dry-run" {
		t.Fatalf("unexpected wrapped hint: %q", got)
	}
}

func TestErrorHint_NoHint(t *testing.T) {
	if got := ErrorHint(errors.New("plain")); got != "" {
		t.Fatalf("expected empty hint, got %q", got)
	}
}

func TestFormatCLIError_WithHint(t *testing.T) {
	out := FormatCLIError(NewUserError("doctor requires --spec", "use --spec examples/spec.sample.json"))
	if !strings.Contains(out, "error:") {
		t.Fatalf("expected error label, got %q", out)
	}
	if !strings.Contains(out, "hint:") {
		t.Fatalf("expected hint label, got %q", out)
	}
	if !strings.Contains(out, "doctor requires --spec") {
		t.Fatalf("expected error content, got %q", out)
	}
}

func TestFormatCLIError_WithoutHint(t *testing.T) {
	out := FormatCLIError(errors.New("plain failure"))
	if !strings.Contains(out, "plain failure") {
		t.Fatalf("expected error content, got %q", out)
	}
	if strings.Contains(out, "hint:") {
		t.Fatalf("did not expect hint label, got %q", out)
	}
}

func TestErrorHint_Nil(t *testing.T) {
	if got := ErrorHint(nil); got != "" {
		t.Fatalf("ErrorHint(nil) = %q, want empty", got)
	}
}

func TestWithHint_Nil(t *testing.T) {
	if err := WithHint(nil, "hint"); err != nil {
		t.Fatalf("WithHint(nil) should return nil, got %v", err)
	}
}

func TestFormatCLIError_Nil(t *testing.T) {
	if got := FormatCLIError(nil); got != "" {
		t.Fatalf("FormatCLIError(nil) = %q, want empty", got)
	}
}

func TestUserError_NilReceiver(t *testing.T) {
	var e *UserError
	if got := e.Error(); got != "" {
		t.Fatalf("(*UserError)(nil).Error() = %q, want empty", got)
	}
	if got := e.Hint(); got != "" {
		t.Fatalf("(*UserError)(nil).Hint() = %q, want empty", got)
	}
}

func TestHintedError_NilReceiver(t *testing.T) {
	var e *hintedError
	if got := e.Error(); got != "" {
		t.Fatalf("(*hintedError)(nil).Error() = %q, want empty", got)
	}
	if got := e.Unwrap(); got != nil {
		t.Fatalf("(*hintedError)(nil).Unwrap() = %v, want nil", got)
	}
	if got := e.Hint(); got != "" {
		t.Fatalf("(*hintedError)(nil).Hint() = %q, want empty", got)
	}
}

func TestIsInterrupted(t *testing.T) {
	if !IsInterrupted(ErrInterrupted) {
		t.Fatal("expected sentinel to be recognized")
	}
	wrapped := WithHint(ErrInterrupted, "stop")
	if !IsInterrupted(wrapped) {
		t.Fatal("expected wrapped interrupted sentinel to be recognized")
	}
	if IsInterrupted(errors.New("other")) {
		t.Fatal("unexpected interrupted match")
	}
}
