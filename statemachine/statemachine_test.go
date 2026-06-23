package statemachine

import (
	"errors"
	"reflect"
	"testing"
)

const (
	stateDraft    State = "draft"
	statePending  State = "pending"
	statePaid     State = "paid"
	stateCanceled State = "canceled"

	eventSubmit Event = "submit"
	eventPay    Event = "pay"
	eventCancel Event = "cancel"
)

func TestNewRejectsDuplicateTransitions(t *testing.T) {
	_, err := New(Config{
		Transitions: []Transition{
			{From: stateDraft, Event: eventSubmit, To: statePending},
			{From: stateDraft, Event: eventSubmit, To: stateCanceled},
		},
	})
	if !errors.Is(err, ErrDuplicateTransition) {
		t.Fatalf("New() error = %v, want ErrDuplicateTransition", err)
	}
}

func TestNewRejectsEmptyTransitionFields(t *testing.T) {
	tests := []struct {
		name       string
		transition Transition
	}{
		{
			name:       "empty from",
			transition: Transition{Event: eventSubmit, To: statePending},
		},
		{
			name:       "empty event",
			transition: Transition{From: stateDraft, To: statePending},
		},
		{
			name:       "empty to",
			transition: Transition{From: stateDraft, Event: eventSubmit},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(Config{
				Transitions: []Transition{tt.transition},
			})
			if !errors.Is(err, ErrInvalidConfig) {
				t.Fatalf("New() error = %v, want ErrInvalidConfig", err)
			}
		})
	}
}

func TestTriggerMovesToTargetState(t *testing.T) {
	machine := mustNew(t, Config{
		Transitions: []Transition{
			{From: stateDraft, Event: eventSubmit, To: statePending},
		},
	})

	next, err := machine.Trigger(stateDraft, eventSubmit)
	if err != nil {
		t.Fatalf("Trigger() error = %v, want nil", err)
	}
	if next != statePending {
		t.Fatalf("Trigger() state = %q, want %q", next, statePending)
	}
}

func TestTriggerReturnsCurrentStateForInvalidTransition(t *testing.T) {
	machine := mustNew(t, Config{
		Transitions: []Transition{
			{From: stateDraft, Event: eventSubmit, To: statePending},
		},
	})

	next, err := machine.Trigger(statePending, eventPay)
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("Trigger() error = %v, want ErrInvalidTransition", err)
	}
	if next != statePending {
		t.Fatalf("Trigger() state = %q, want current state %q", next, statePending)
	}
}

func TestCanReportsRegisteredTransition(t *testing.T) {
	machine := mustNew(t, Config{
		Transitions: []Transition{
			{From: statePending, Event: eventPay, To: statePaid},
		},
	})

	if !machine.Can(statePending, eventPay) {
		t.Fatal("Can() = false, want true")
	}
	if machine.Can(stateDraft, eventPay) {
		t.Fatal("Can() = true, want false")
	}
}

func TestAvailableEventsReturnsCopy(t *testing.T) {
	machine := mustNew(t, Config{
		Transitions: []Transition{
			{From: statePending, Event: eventPay, To: statePaid},
			{From: statePending, Event: eventCancel, To: stateCanceled},
		},
	})

	events := machine.AvailableEvents(statePending)
	events[0] = "mutated"

	expected := []Event{eventPay, eventCancel}
	if got := machine.AvailableEvents(statePending); !reflect.DeepEqual(got, expected) {
		t.Fatalf("AvailableEvents() = %#v, want %#v", got, expected)
	}
}

func TestAvailableTransitionsReturnsCopy(t *testing.T) {
	machine := mustNew(t, Config{
		Transitions: []Transition{
			{From: statePending, Event: eventPay, To: statePaid},
			{From: statePending, Event: eventCancel, To: stateCanceled},
		},
	})

	transitions := machine.AvailableTransitions(statePending)
	transitions[0].To = stateCanceled

	got := machine.AvailableTransitions(statePending)
	if got[0].To != statePaid {
		t.Fatalf("AvailableTransitions()[0].To = %q, want %q", got[0].To, statePaid)
	}
}

func TestNilMachine(t *testing.T) {
	var machine *Machine

	next, err := machine.Trigger(stateDraft, eventSubmit)
	if !errors.Is(err, ErrNilMachine) {
		t.Fatalf("Trigger() error = %v, want ErrNilMachine", err)
	}
	if next != stateDraft {
		t.Fatalf("Trigger() state = %q, want current state %q", next, stateDraft)
	}
	if machine.Can(stateDraft, eventSubmit) {
		t.Fatal("Can() = true, want false")
	}
	if events := machine.AvailableEvents(stateDraft); events != nil {
		t.Fatalf("AvailableEvents() = %#v, want nil", events)
	}
	if transitions := machine.AvailableTransitions(stateDraft); transitions != nil {
		t.Fatalf("AvailableTransitions() = %#v, want nil", transitions)
	}
}

func mustNew(t *testing.T, cfg Config) *Machine {
	t.Helper()

	machine, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	return machine
}
