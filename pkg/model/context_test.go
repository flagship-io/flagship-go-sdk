package model

import (
	"testing"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestValidate(t *testing.T) {

	context := Context{}
	context["test_string"] = structpb.NewStringValue("123")
	context["test_number"] = structpb.NewNumberValue(36.5)
	context["test_bool"] = structpb.NewBoolValue(true)
	context["test_int"] = structpb.NewNumberValue(4)
	// context["test_wrong"] = errors.New("wrong type")

	err := context.Validate()

	if err == nil {
		t.Error("Wrong context variable should raise an error")
	}
}
