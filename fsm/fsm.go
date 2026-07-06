package fsm

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrInvalidTransition   = errors.New("fsm: invalid transition")
	ErrDuplicateTransition = errors.New("fsm: duplicate transition")
)

type Transition[S comparable, E comparable] struct {
	From  S
	Event E
	To    S
}

type Machine[S comparable, E comparable] struct {
	mu          sync.RWMutex
	state       S
	transitions map[transitionKey[S, E]]Transition[S, E]
	list        []Transition[S, E]
}

type transitionKey[S comparable, E comparable] struct {
	from  S
	event E
}

func New[S comparable, E comparable](initial S, transitions []Transition[S, E]) (*Machine[S, E], error) {
	index := make(map[transitionKey[S, E]]Transition[S, E], len(transitions))
	for _, tr := range transitions {
		key := transitionKey[S, E]{from: tr.From, event: tr.Event}
		if _, ok := index[key]; ok {
			return nil, fmt.Errorf("%w: from=%v event=%v", ErrDuplicateTransition, tr.From, tr.Event)
		}
		index[key] = tr
	}
	list := append([]Transition[S, E](nil), transitions...)
	return &Machine[S, E]{state: initial, transitions: index, list: list}, nil
}

func (m *Machine[S, E]) State() S {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

func (m *Machine[S, E]) Can(event E) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.transitions[transitionKey[S, E]{from: m.state, event: event}]
	return ok
}

func (m *Machine[S, E]) Apply(event E) (S, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	tr, ok := m.transitions[transitionKey[S, E]{from: m.state, event: event}]
	if !ok {
		return m.state, fmt.Errorf("%w: from=%v event=%v", ErrInvalidTransition, m.state, event)
	}
	m.state = tr.To
	return m.state, nil
}

func (m *Machine[S, E]) Transitions() []Transition[S, E] {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]Transition[S, E](nil), m.list...)
}
