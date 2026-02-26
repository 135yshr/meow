package meowrt

import "fmt"

var methodRegistry = map[string]map[string]func(...Value) Value{}

// RegisterMethod registers a method for a named type.
func RegisterMethod(typeName, methodName string, fn func(...Value) Value) {
	if methodRegistry[typeName] == nil {
		methodRegistry[typeName] = map[string]func(...Value) Value{}
	}
	methodRegistry[typeName][methodName] = fn
}

// LookupMethod returns the method function for a type, if registered.
func LookupMethod(typeName, methodName string) (func(...Value) Value, bool) {
	if methods, ok := methodRegistry[typeName]; ok {
		if fn, ok := methods[methodName]; ok {
			return fn, true
		}
	}
	return nil, false
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
