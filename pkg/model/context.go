package model

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

// Context represents a visitor context object
type Context map[string]*structpb.Value

// Validate checks that the visitor context object is valid
func (c Context) Validate() []error {
	errorList := []error{}

	for key, val := range c {
		_, okBool := val.GetKind().(*structpb.Value_BoolValue)
		_, okString := val.GetKind().(*structpb.Value_StringValue)
		_, okNumber := val.GetKind().(*structpb.Value_NumberValue)

		if !okBool && !okString && !okNumber {
			errorList = append(errorList, fmt.Errorf("Value %v not handled for key %s. Type must be one of string, bool or number (int or float64)", val, key))
		}
	}
	return errorList
}
