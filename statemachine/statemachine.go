package statemachine

import (
	"errors"
	"fmt"
	"slices"
)

var (
	ErrInvalidTransition = errors.New("statemachine: invalid transition")

	ErrDuplicateTransition = errors.New("statemachine: duplicate transition")

	ErrInvalidConfig = errors.New("statemachine: invalid config")

	ErrNilMachine = errors.New("statemachine: nil machine")
)

type State string

type Event string

type Transition struct {
	From  State
	Event Event
	To    State
}

type Config struct {
	Transitions []Transition
}

type Machine struct {
	transitions        map[transitionKey]Transition
	eventsByState      map[State][]Event
	transitionsByState map[State][]Transition
}

type transitionKey struct {
	from  State
	event Event
}

func New(cfg Config) (*Machine, error) {
	transitions := make(map[transitionKey]Transition, len(cfg.Transitions))
	eventsByState := make(map[State][]Event)
	transitionsByState := make(map[State][]Transition)

	for _, transition := range cfg.Transitions {
		if transition.From == "" {
			return nil, fmt.Errorf("%w: transition from is empty", ErrInvalidConfig)
		}
		if transition.Event == "" {
			return nil, fmt.Errorf("%w: transition event is empty", ErrInvalidConfig)
		}
		if transition.To == "" {
			return nil, fmt.Errorf("%w: transition to is empty", ErrInvalidConfig)
		}

		key := transitionKey{from: transition.From, event: transition.Event}
		if existing, ok := transitions[key]; ok {
			return nil, fmt.Errorf(
				"%w: from=%q event=%q existing_to=%q duplicate_to=%q",
				ErrDuplicateTransition,
				transition.From,
				transition.Event,
				existing.To,
				transition.To,
			)
		}

		transitions[key] = transition
		eventsByState[transition.From] = append(eventsByState[transition.From], transition.Event)
		transitionsByState[transition.From] = append(transitionsByState[transition.From], transition)
	}

	return &Machine{
		transitions:        transitions,
		eventsByState:      eventsByState,
		transitionsByState: transitionsByState,
	}, nil
}

func (m *Machine) Trigger(currentState State, event Event) (State, error) {
	if m == nil {
		return currentState, ErrNilMachine
	}

	transition, ok := m.transitions[transitionKey{from: currentState, event: event}]
	if !ok {
		return currentState, fmt.Errorf(
			"%w: from=%q event=%q",
			ErrInvalidTransition,
			currentState,
			event,
		)
	}

	return transition.To, nil
}

func (m *Machine) Can(currentState State, event Event) bool {
	if m == nil {
		return false
	}
	_, ok := m.transitions[transitionKey{from: currentState, event: event}]
	return ok
}

func (m *Machine) AvailableEvents(currentState State) []Event {
	if m == nil {
		return nil
	}
	return slices.Clone(m.eventsByState[currentState])
}

func (m *Machine) AvailableTransitions(currentState State) []Transition {
	if m == nil {
		return nil
	}
	return slices.Clone(m.transitionsByState[currentState])
}
