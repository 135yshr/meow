package interpreter

import (
	"fmt"

	"github.com/135yshr/meow/runtime/meowrt"
)

// Environment manages variable bindings with lexical scoping.
type Environment struct {
	vars   map[string]meowrt.Value
	parent *Environment
}

// NewEnvironment creates a new top-level environment.
func NewEnvironment() *Environment {
	return &Environment{vars: make(map[string]meowrt.Value)}
}

// Child creates a child scope.
func (e *Environment) Child() *Environment {
	return &Environment{vars: make(map[string]meowrt.Value), parent: e}
}

// Define binds a new variable in the current scope.
func (e *Environment) Define(name string, val meowrt.Value) {
	e.vars[name] = val
}

// Set updates an existing variable, walking up the scope chain.
// Panics if the variable is not found.
func (e *Environment) Set(name string, val meowrt.Value) {
	for env := e; env != nil; env = env.parent {
		if _, ok := env.vars[name]; ok {
			env.vars[name] = val
			return
		}
	}
	panic(fmt.Sprintf("Hiss! undefined variable %s, nya~", name))
}

// Get retrieves a variable, walking up the scope chain.
// Panics if the variable is not found.
func (e *Environment) Get(name string) meowrt.Value {
	for env := e; env != nil; env = env.parent {
		if v, ok := env.vars[name]; ok {
			return v
		}
	}
	panic(fmt.Sprintf("Hiss! undefined variable %s, nya~", name))
}

// Has returns true if name is defined in this scope chain.
func (e *Environment) Has(name string) bool {
	for env := e; env != nil; env = env.parent {
		if _, ok := env.vars[name]; ok {
			return true
		}
	}
	return false
}
