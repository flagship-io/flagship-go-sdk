package model

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
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

func (c Context) ToProtoMap() (map[string]*structpb.Value, error) {
	ret := map[string]*structpb.Value{}
	for key, value := range c {
		newProto, err := structpb.NewValue(value)
		if err != nil {
			return nil, errors.New(fmt.Sprint("error in context proto", err))
		}

		ret[key] = newProto

	}
	return ret, nil
}
