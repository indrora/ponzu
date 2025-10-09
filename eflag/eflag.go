package eflag

import (
	"fmt"
	"strings"
)

type EnumFlag[T comparable] struct {
	value        *T
	allowedMap   map[string]T
	allowedNames []string
	placeholder  string
}

func NewEnumFlag[T comparable](target *T, defaultValue T, placeholder string, allowedMap map[string]T) *EnumFlag[T] {

	*target = defaultValue

	// Build a sorted list of names for help text
	names := make([]string, 0, len(allowedMap))
	for name := range allowedMap {
		names = append(names, name)
	}

	return &EnumFlag[T]{
		value:        target,
		allowedMap:   allowedMap,
		allowedNames: names,
		placeholder:  placeholder,
	}
}

func (e *EnumFlag[T]) String() string {
	// Find the name for the current value
	for name, val := range e.allowedMap {
		if val == *e.value {
			return name
		}
	}
	return ""
}

// Set sets the value from a string
func (e *EnumFlag[T]) Set(s string) error {
	val, ok := e.allowedMap[s]
	if !ok {
		return fmt.Errorf("invalid value %q, must be one of: %s", s, strings.Join(e.allowedNames, ", "))
	}
	*e.value = val
	return nil
}

func (e *EnumFlag[T]) Type() string {
	return e.placeholder
}
