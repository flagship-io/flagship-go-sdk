package bucketing

import (
	"errors"
	"reflect"
	"strings"
)

func isANDListOperator(operator TargetingOperator) bool {
	return operator == NOT_EQUALS || operator == NOT_CONTAINS
}

func isORListOperator(operator TargetingOperator) bool {
	return operator == EQUALS || operator == CONTAINS
}

// TargetingMatch returns true if a visitor ID and context match the variationGroup targeting
func TargetingMatch(variationGroup *VariationGroup, visitorID string, context map[string]interface{}) (bool, error) {
	globalMatch := false
	for _, targetingGroup := range variationGroup.Targeting.TargetingGroups {
		matchGroup := len(targetingGroup.Targetings) > 0
		for _, targeting := range targetingGroup.Targetings {
			v, ok := context[targeting.Key]
			switch targeting.Key {
			case "fs_all_users":
				return true, nil
			case "fs_users":
				v = visitorID
				ok = true
			}

			if ok {
				matchTargeting, err := targetingMatchOperator(targeting.Operator, targeting.Value, v)
				if err != nil {
					return false, err
				}

				matchGroup = matchGroup && matchTargeting
			} else {
				matchGroup = false
			}
		}
		globalMatch = globalMatch || matchGroup
	}

	return globalMatch, nil
}

func targetingMatchOperator(operator TargetingOperator, targetingValue interface{}, contextValue interface{}) (bool, error) {
	match := false
	var err error

	isList := strings.Contains(reflect.TypeOf(targetingValue).String(), "[]")

	if isList {
		return targetingMatchOperatorList(operator, targetingValue, contextValue)
	}

	// Except for targeting value of type list, check that context and targeting types are equals
	if reflect.TypeOf(targetingValue) != reflect.TypeOf(contextValue) {
		return false, errors.New("Targeting and Context value kinds mismatch")
	}

	switch targetingValue.(type) {
	case string:
		targetingValueCasted := targetingValue.(string)
		contextValueCasted := contextValue.(string)
		match, err = targetingMatchOperatorString(operator, targetingValueCasted, contextValueCasted)
		break
	case bool:
		targetingValueCasted := targetingValue.(bool)
		contextValueCasted := contextValue.(bool)
		match, err = targetingMatchOperatorBool(operator, targetingValueCasted, contextValueCasted)
		break
	case int:
		targetingValueCasted := targetingValue.(int)
		contextValueCasted := contextValue.(int)
		match, err = targetingMatchOperatorNumber(operator, float64(targetingValueCasted), float64(contextValueCasted))
	case float64:
		targetingValueCasted := targetingValue.(float64)
		contextValueCasted := contextValue.(float64)
		match, err = targetingMatchOperatorNumber(operator, targetingValueCasted, contextValueCasted)
	}

	return match, err
}

func targetingMatchOperatorList(operator TargetingOperator, targetingValue interface{}, contextValue interface{}) (bool, error) {
	targetingList, convOk := takeSliceArg(targetingValue)
	if !convOk {
		return false, errors.New("Could not convert list targeting")
	}
	match := isANDListOperator(operator)
	for _, v := range targetingList {
		subValueMatch, err := targetingMatchOperator(operator, v, contextValue)
		if err != nil {
			return false, err
		}

		if isANDListOperator(operator) {
			match = match && subValueMatch
		}
		if isORListOperator(operator) {
			match = match || subValueMatch
		}
	}

	return match, nil
}

func targetingMatchOperatorString(operator TargetingOperator, targetingValue string, contextValue string) (bool, error) {
	switch operator {
	case LOWER_THAN:
		return strings.ToLower(contextValue) < strings.ToLower(targetingValue), nil
	case GREATER_THAN:
		return strings.ToLower(contextValue) > strings.ToLower(targetingValue), nil
	case LOWER_THAN_OR_EQUALS:
		return strings.ToLower(contextValue) <= strings.ToLower(targetingValue), nil
	case GREATER_THAN_OR_EQUALS:
		return strings.ToLower(contextValue) >= strings.ToLower(targetingValue), nil
	case EQUALS:
		return strings.ToLower(contextValue) == strings.ToLower(targetingValue), nil
	case NOT_EQUALS:
		return strings.ToLower(contextValue) != strings.ToLower(targetingValue), nil
	case STARTS_WITH:
		return strings.HasPrefix(strings.ToLower(contextValue), strings.ToLower(targetingValue)), nil
	case ENDS_WITH:
		return strings.HasSuffix(strings.ToLower(contextValue), strings.ToLower(targetingValue)), nil
	case CONTAINS:
		return strings.Contains(strings.ToLower(contextValue), strings.ToLower(targetingValue)), nil
	case NOT_CONTAINS:
		return !strings.Contains(strings.ToLower(contextValue), strings.ToLower(targetingValue)), nil
	// case "regex":
	// 	match, err := regexp.MatchString(targetingValue, contextValue)
	// 	return match, err
	default:
		return false, errors.New("Operator not handled")
	}
}

func targetingMatchOperatorNumber(operator TargetingOperator, targetingValue float64, contextValue float64) (bool, error) {
	switch operator {
	case LOWER_THAN:
		return contextValue < targetingValue, nil
	case GREATER_THAN:
		return contextValue > targetingValue, nil
	case LOWER_THAN_OR_EQUALS:
		return contextValue <= targetingValue, nil
	case GREATER_THAN_OR_EQUALS:
		return contextValue >= targetingValue, nil
	case EQUALS:
		return contextValue == targetingValue, nil
	case NOT_EQUALS:
		return contextValue != targetingValue, nil
	default:
		return false, errors.New("Operator not handled")
	}
}

func targetingMatchOperatorBool(operator TargetingOperator, targetingValue bool, contextValue bool) (bool, error) {
	switch operator {
	case EQUALS:
		return contextValue == targetingValue, nil
	case NOT_EQUALS:
		return contextValue != targetingValue, nil
	default:
		return false, errors.New("Operator not handled")
	}
}

func takeSliceArg(arg interface{}) (out []interface{}, ok bool) {
	slice, success := takeArg(arg, reflect.Slice)
	if !success {
		ok = false
		return
	}
	c := slice.Len()
	out = make([]interface{}, c)
	for i := 0; i < c; i++ {
		out[i] = slice.Index(i).Interface()
	}
	return out, true
}

func takeArg(arg interface{}, kind reflect.Kind) (val reflect.Value, ok bool) {
	val = reflect.ValueOf(arg)
	if val.Kind() == kind {
		ok = true
	}
	return
}
