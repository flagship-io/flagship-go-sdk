package bucketing

import (
	"testing"
)

func testTargetingNumber(operator TargetingOperator, targetingValue float64, value float64, t *testing.T, shouldMatch bool, shouldRaiseError bool) {
	match, err := targetingMatchOperatorNumber(operator, targetingValue, value)

	if ((err != nil && !shouldRaiseError) || (shouldRaiseError && err == nil)) || (match != shouldMatch) {
		t.Errorf("Targeting number %v not working - tv : %f, v: %f, match : %v, err: %v", operator, targetingValue, value, match, err)
	}
}

func testTargetingBoolean(operator TargetingOperator, targetingValue bool, value bool, t *testing.T, shouldMatch bool, shouldRaiseError bool) {
	match, err := targetingMatchOperatorBool(operator, targetingValue, value)

	if ((err != nil && !shouldRaiseError) || (shouldRaiseError && err == nil)) || (match != shouldMatch) {
		t.Errorf("Targeting number %v not working - tv : %v, v: %v, match : %v, err: %v", operator, targetingValue, value, match, err)
	}
}

func testTargetingString(operator TargetingOperator, targetingValue string, value string, t *testing.T, shouldMatch bool, shouldRaiseError bool) {
	match, err := targetingMatchOperatorString(operator, targetingValue, value)

	if ((err != nil && !shouldRaiseError) || (shouldRaiseError && err == nil)) || (match != shouldMatch) {
		t.Errorf("Targeting string %v not working - tv : %v, v: %v, match : %v, err: %v", operator, targetingValue, value, match, err)
	}
}

func testTargetingCast(operator TargetingOperator, targeting interface{}, value interface{}, t *testing.T, shouldMatch bool, shouldRaiseError bool) {
	match, err := targetingMatchOperator(operator, targeting, value)

	if ((err != nil && !shouldRaiseError) || (shouldRaiseError && err == nil)) || (match != shouldMatch) {
		t.Errorf("Targeting cast value %v not working - tv : %v, v: %v, match : %v, err: %v", operator, targeting, value, match, err)
	}
}

func testTargetingListString(operator TargetingOperator, targetingValues []string, value string, t *testing.T, shouldMatch bool, shouldRaiseError bool) {
	stringValues := []string{}
	for _, str := range targetingValues {
		stringValues = append(stringValues, str)
	}

	match, err := targetingMatchOperator(operator, stringValues, value)

	if ((err != nil && !shouldRaiseError) || (shouldRaiseError && err == nil)) || (match != shouldMatch) {
		t.Errorf("Targeting list string %v not working - tv : %v, v: %v, match : %v, err: %v", operator, targetingValues, value, match, err)
	}
}

// TestNumberTargeting checks all possible number targeting
func TestNumberTargeting(t *testing.T) {
	testTargetingNumber(LOWER_THAN, 11, 10, t, true, false)
	testTargetingNumber(LOWER_THAN, 10, 10, t, false, false)
	testTargetingNumber(LOWER_THAN, 9, 10, t, false, false)

	testTargetingNumber(LOWER_THAN_OR_EQUALS, 11, 10, t, true, false)
	testTargetingNumber(LOWER_THAN_OR_EQUALS, 10, 10, t, true, false)
	testTargetingNumber(LOWER_THAN_OR_EQUALS, 9, 10, t, false, false)

	testTargetingNumber(GREATER_THAN, 11, 10, t, false, false)
	testTargetingNumber(GREATER_THAN, 10, 10, t, false, false)
	testTargetingNumber(GREATER_THAN, 9, 10, t, true, false)

	testTargetingNumber(GREATER_THAN_OR_EQUALS, 11, 10, t, false, false)
	testTargetingNumber(GREATER_THAN_OR_EQUALS, 10, 10, t, true, false)
	testTargetingNumber(GREATER_THAN_OR_EQUALS, 9, 10, t, true, false)

	testTargetingNumber(NOT_EQUALS, 11, 10, t, true, false)
	testTargetingNumber(NOT_EQUALS, 10, 10, t, false, false)
	testTargetingNumber(NOT_EQUALS, 9, 10, t, true, false)

	testTargetingNumber(EQUALS, 11, 10, t, false, false)
	testTargetingNumber(EQUALS, 10, 10, t, true, false)
	testTargetingNumber(EQUALS, 9, 10, t, false, false)

	testTargetingNumber(CONTAINS, 11, 10, t, false, true)
	testTargetingNumber(ENDS_WITH, 10, 10, t, false, true)
	testTargetingNumber(STARTS_WITH, 9, 10, t, false, true)
}

// TestBooleanTargeting checks all possible boolean targeting
func TestBooleanTargeting(t *testing.T) {
	testTargetingBoolean(NOT_EQUALS, true, false, t, true, false)
	testTargetingBoolean(NOT_EQUALS, true, true, t, false, false)
	testTargetingBoolean(NOT_EQUALS, false, true, t, true, false)

	testTargetingBoolean(EQUALS, true, false, t, false, false)
	testTargetingBoolean(EQUALS, true, true, t, true, false)
	testTargetingBoolean(EQUALS, false, true, t, false, false)

	testTargetingBoolean(CONTAINS, true, false, t, false, true)
	testTargetingBoolean(ENDS_WITH, true, false, t, false, true)
	testTargetingBoolean(STARTS_WITH, true, false, t, false, true)
	testTargetingBoolean(GREATER_THAN, true, false, t, false, true)
	testTargetingBoolean(GREATER_THAN_OR_EQUALS, true, false, t, false, true)
	testTargetingBoolean(LOWER_THAN, true, false, t, false, true)
	testTargetingBoolean(LOWER_THAN_OR_EQUALS, true, false, t, false, true)
}

// TestStringTargeting checks all possible string targeting
func TestStringTargeting(t *testing.T) {
	testTargetingString(LOWER_THAN, "abc", "abd", t, false, false)
	testTargetingString(LOWER_THAN, "abc", "abc", t, false, false)
	testTargetingString(LOWER_THAN, "abd", "abc", t, true, false)

	testTargetingString(LOWER_THAN_OR_EQUALS, "abc", "abd", t, false, false)
	testTargetingString(LOWER_THAN_OR_EQUALS, "abc", "abc", t, true, false)
	testTargetingString(LOWER_THAN_OR_EQUALS, "abd", "abc", t, true, false)

	testTargetingString(GREATER_THAN, "abc", "abd", t, true, false)
	testTargetingString(GREATER_THAN, "abc", "abc", t, false, false)
	testTargetingString(GREATER_THAN, "abd", "abc", t, false, false)

	testTargetingString(GREATER_THAN_OR_EQUALS, "abc", "abd", t, true, false)
	testTargetingString(GREATER_THAN_OR_EQUALS, "abc", "abd", t, true, false)
	testTargetingString(GREATER_THAN_OR_EQUALS, "abd", "abc", t, false, false)

	testTargetingString(NOT_EQUALS, "abc", "abd", t, true, false)
	testTargetingString(NOT_EQUALS, "abc", "abc", t, false, false)
	testTargetingString(NOT_EQUALS, "", "", t, false, false)
	testTargetingString(NOT_EQUALS, "", " ", t, true, false)

	testTargetingString(EQUALS, "abc", "abd", t, false, false)
	testTargetingString(EQUALS, "abc", "abc", t, true, false)
	testTargetingString(EQUALS, "ABC", "abc", t, true, false)
	testTargetingString(EQUALS, "", "", t, true, false)
	testTargetingString(EQUALS, "", " ", t, false, false)

	testTargetingString(CONTAINS, "b", "abc", t, true, false)
	testTargetingString(CONTAINS, "B", "abc", t, true, false)
	testTargetingString(CONTAINS, "d", "abc", t, false, false)

	testTargetingString(NOT_CONTAINS, "d", "abc", t, true, false)
	testTargetingString(NOT_CONTAINS, "D", "abc", t, true, false)
	testTargetingString(NOT_CONTAINS, "b", "abc", t, false, false)

	testTargetingString(ENDS_WITH, "c", "abc", t, true, false)
	testTargetingString(ENDS_WITH, "C", "abc", t, true, false)
	testTargetingString(ENDS_WITH, "d", "abc", t, false, false)
	testTargetingString(ENDS_WITH, "a", "abc", t, false, false)
	testTargetingString(ENDS_WITH, "", "abc", t, true, false)

	testTargetingString(STARTS_WITH, "a", "abc", t, true, false)
	testTargetingString(STARTS_WITH, "A", "abc", t, true, false)
	testTargetingString(STARTS_WITH, "d", "abc", t, false, false)
	testTargetingString(STARTS_WITH, "c", "abc", t, false, false)
	testTargetingString(STARTS_WITH, "", "abc", t, true, false)

	testTargetingString(NULL, "", "abc", t, false, true)
}

// TestListStringTargeting checks all possible string list targeting
func TestListStringTargeting(t *testing.T) {
	testTargetingListString(EQUALS, []string{"abc"}, "abd", t, false, false)
	testTargetingListString(EQUALS, []string{"abc"}, "abc", t, true, false)
	testTargetingListString(NOT_EQUALS, []string{"abc"}, "abd", t, true, false)
	testTargetingListString(NOT_EQUALS, []string{"abc"}, "abc", t, false, false)

	testTargetingListString(EQUALS, []string{"abc", "bcd"}, "abd", t, false, false)
	testTargetingListString(EQUALS, []string{"abc", "bcd"}, "abc", t, true, false)
	testTargetingListString(NOT_EQUALS, []string{"abc", "bcd"}, "abd", t, true, false)
	testTargetingListString(NOT_EQUALS, []string{"abc", "bcd"}, "abc", t, false, false)
}

// TestTargetingCast checks all possible targeting type
func TestTargetingCast(t *testing.T) {
	testTargetingCast(EQUALS, "abc", "abc", t, true, false)
	testTargetingCast(EQUALS, 1, "abc", t, false, true)

	testTargetingCast(EQUALS, 200, 200, t, true, false)
	testTargetingCast(EQUALS, 200, 400, t, false, false)

	testTargetingCast(EQUALS, 200.0, 200.0, t, true, false)
	testTargetingCast(EQUALS, 200.0, 400.0, t, false, false)

	testTargetingCast(EQUALS, true, true, t, true, false)
	testTargetingCast(EQUALS, true, false, t, false, false)
}

func TestTargetingMatch(t *testing.T) {
	vg := VariationGroup{
		Targeting: TargetingWrapper{
			TargetingGroups: []*TargetingGroup{
				{
					Targetings: []*Targeting{
						{
							Operator: EQUALS,
							Key:      "test",
							Value:    1,
						},
					},
				},
			},
		},
	}

	context := map[string]interface{}{
		"test": true,
	}
	_, err := TargetingMatch(&vg, testVID, context)

	if err == nil {
		t.Error("Expected error as targeting and context value type do not match")
	}
}
