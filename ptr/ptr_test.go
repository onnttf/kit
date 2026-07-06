package ptr

import "testing"

func TestTo(t *testing.T) {
	p := To("x")
	if p == nil || *p != "x" {
		t.Fatalf("To() = %v, want pointer to x", p)
	}
}

func TestValue(t *testing.T) {
	if got := Value[string](nil); got != "" {
		t.Fatalf("Value(nil) = %q, want zero value", got)
	}

	p := To(42)
	if got := Value(p); got != 42 {
		t.Fatalf("Value() = %d, want 42", got)
	}
}

func TestOr(t *testing.T) {
	if got := Or[int](nil, 7); got != 7 {
		t.Fatalf("Or(nil) = %d, want fallback", got)
	}

	p := To(3)
	if got := Or(p, 7); got != 3 {
		t.Fatalf("Or(non-nil) = %d, want pointed value", got)
	}
}
