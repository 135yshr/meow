package meowrt

import (
	"fmt"
	"sync"
)

var (
	methodRegistry   = map[string]map[string]func(...Value) Value{}
	methodRegistryMu sync.RWMutex
)

// RegisterMethod registers a method for a named type.
func RegisterMethod(typeName, methodName string, fn func(...Value) Value) {
	methodRegistryMu.Lock()
	defer methodRegistryMu.Unlock()
	if methodRegistry[typeName] == nil {
		methodRegistry[typeName] = map[string]func(...Value) Value{}
	}
	methodRegistry[typeName][methodName] = fn
}

// LookupMethod returns the method function for a type, if registered.
func LookupMethod(typeName, methodName string) (func(...Value) Value, bool) {
	methodRegistryMu.RLock()
	defer methodRegistryMu.RUnlock()
	if methods, ok := methodRegistry[typeName]; ok {
		if fn, ok := methods[methodName]; ok {
			return fn, true
		}
	}
	return nil, false
}

// ClearMethods removes all registered methods from the registry.
func ClearMethods() {
	methodRegistryMu.Lock()
	defer methodRegistryMu.Unlock()
	methodRegistry = map[string]map[string]func(...Value) Value{}
}

// DispatchMethod calls a method on a value by looking up the method registry.
func DispatchMethod(obj Value, methodName string, args ...Value) Value {
	typeName := ""
	if k, ok := obj.(*Kitty); ok {
		typeName = k.TypeName
	}
	if typeName == "" {
		panic(fmt.Sprintf("Hiss! cannot call method %s on non-kitty value, nya~", methodName))
	}
	fn, ok := LookupMethod(typeName, methodName)
	if !ok {
		panic(fmt.Sprintf("Hiss! no method %s for type %s, nya~", methodName, typeName))
	}
	allArgs := make([]Value, 0, len(args)+1)
	allArgs = append(allArgs, obj)
	allArgs = append(allArgs, args...)
	return fn(allArgs...)
}
