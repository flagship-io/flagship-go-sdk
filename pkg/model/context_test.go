package model

import (
	"errors"
	"testing"
)

func TestValidate(t *testing.T) {

	context := Context{}
	context["test_string"] = "123"
	context["test_number"] = 36.5
	context["test_bool"] = true
	context["test_int"] = 4
	context["test_wrong"] = errors.New("wrong type")

	err := context.Validate()

	if err == nil {
		t.Error("Wrong context variable should raise an error")
	}
}
