package model

import (
	"fmt"
)

// Context represents a visitor context object
type Context map[string]interface{}

// Validate checks that the visitor context object is valid
func (c Context) Validate() []error {
	errorList := []error{}

	for key, val := range c {
		_, okBool := val.(bool)
		_, okString := val.(string)
		_, okFloat64 := val.(float64)
		intVal, okInt := val.(int)

		if !okBool && !okString && !okFloat64 && !okInt {
			errorList = append(errorList, fmt.Errorf("Value %v not handled for key %s. Type must be one of string, bool or number (int or float64)", val, key))
		}

		if okInt {
			c[key] = float64(intVal)
		}
	}
	return errorList
}
