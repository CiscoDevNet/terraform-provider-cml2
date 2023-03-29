package cmlclient

import (
	"fmt"
	"sync"
)

type clientState uint32

func (cs clientState) String() string {
	switch cs {
	case stateInitial:
		return "INITIAL"
	case stateCheckVersion:
		return "CHECKVERSION"
	case stateAuthRequired:
		return "AUTHREQUIRED"
	case stateAuthenticating:
		return "AUTHENTICATING"
	case stateAuthenticated:
		return "AUTHENTICATED"
	default:
		panic(fmt.Sprintf("unknown state %d", cs))
	}
}

const (
	stateInitial clientState = iota
	stateCheckVersion
	stateAuthRequired
	stateAuthenticating
	stateAuthenticated
)

type apiClientState struct {
	state clientState
	mu    *sync.RWMutex
}

func newState() *apiClientState {
	var mu sync.RWMutex
	return &apiClientState{
		mu:    &mu,
		state: stateInitial,
	}
}

func (acs *apiClientState) set(state clientState) {
	acs.mu.Lock()
	defer acs.mu.Unlock()
	acs.state = state
}

func (acs *apiClientState) get() clientState {
	acs.mu.RLock()
	defer acs.mu.RUnlock()
	return acs.state
}
