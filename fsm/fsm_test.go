package fsm

import (
	"errors"
	"sync"
	"testing"
)

func TestMachine(t *testing.T) {
	m, err := New("draft", []Transition[string, string]{
		{From: "draft", Event: "submit", To: "review"},
		{From: "review", Event: "approve", To: "done"},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if !m.Can("submit") {
		t.Fatalf("Can(submit) = false")
	}
	if got, err := m.Apply("submit"); err != nil || got != "review" {
		t.Fatalf("Apply() = %q, %v", got, err)
	}
	if _, err := m.Apply("submit"); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("Apply(invalid) error = %v", err)
	}
}

func TestDuplicateTransition(t *testing.T) {
	_, err := New("a", []Transition[string, string]{
		{From: "a", Event: "go", To: "b"},
		{From: "a", Event: "go", To: "c"},
	})
	if !errors.Is(err, ErrDuplicateTransition) {
		t.Fatalf("New() error = %v", err)
	}
}

func TestTransitionsReturnsCopyAndConcurrentAccess(t *testing.T) {
	m, err := New(0, []Transition[int, int]{{From: 0, Event: 1, To: 1}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	transitions := m.Transitions()
	transitions[0].To = 99
	if got := m.Transitions()[0].To; got != 1 {
		t.Fatalf("Transitions() leaked internal slice: %v", got)
	}

	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			_ = m.State()
			_ = m.Can(1)
		})
	}
	wg.Wait()
}

func TestMachineTransitionsAreIndependentFromInput(t *testing.T) {
	input := []Transition[string, string]{{From: "a", Event: "go", To: "b"}}
	m, err := New("a", input)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	input[0].To = "mutated"

	got, err := m.Apply("go")
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if got != "b" {
		t.Fatalf("Apply() = %q, want b", got)
	}
}

func TestMachineConcurrentApplySingleWinner(t *testing.T) {
	m, err := New(0, []Transition[int, string]{{From: 0, Event: "go", To: 1}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 20)
	for range 20 {
		wg.Go(func() {
			_, err := m.Apply("go")
			errs <- err
		})
	}
	wg.Wait()
	close(errs)

	var successes, failures int
	for err := range errs {
		if err == nil {
			successes++
			continue
		}
		if !errors.Is(err, ErrInvalidTransition) {
			t.Fatalf("Apply() error = %v", err)
		}
		failures++
	}
	if successes != 1 || failures != 19 || m.State() != 1 {
		t.Fatalf("successes/failures/state = %d/%d/%d", successes, failures, m.State())
	}
}

func TestEmptyMachine(t *testing.T) {
	m, err := New[string, string]("idle", nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if got := m.State(); got != "idle" {
		t.Fatalf("State() = %q", got)
	}
	if m.Can("go") {
		t.Fatalf("Can(go) = true")
	}
	if _, err := m.Apply("go"); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("Apply(go) error = %v", err)
	}
	if got := m.Transitions(); len(got) != 0 {
		t.Fatalf("Transitions() = %#v", got)
	}
}
