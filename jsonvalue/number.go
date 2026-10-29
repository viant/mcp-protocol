// Package jsonvalue preserves exact JSON numbers while retaining legacy Go
// float64 values when their JSON encoding has the same decimal value and
// integral values remain within the safe IEEE-754 integer range.
package jsonvalue

import (
	"encoding/json"
	"strconv"
	"strings"
)

// Normalize recursively normalizes JSON decoder values in place. Numbers whose
// float64 JSON encoding would change their decimal value remain json.Number.
// Integral values outside +/- (2^53-1) also remain json.Number, protecting Go
// callers that convert a generic float64 to an integer before serialization.
// It accepts only the generic maps, slices and scalars produced by UseNumber;
// declared numeric fields and custom Go values are left untouched.
func Normalize(value interface{}) interface{} {
	switch actual := value.(type) {
	case json.Number:
		number, err := actual.Float64()
		if err != nil {
			return actual
		}
		encoded, err := json.Marshal(number)
		if err != nil {
			return actual
		}
		left, leftExponent, leftOK := decimal(actual.String())
		right, rightExponent, rightOK := decimal(string(encoded))
		const maxSafeInteger = 1<<53 - 1
		if leftOK && leftExponent >= 0 && (number > maxSafeInteger || number < -maxSafeInteger) {
			return actual
		}
		if leftOK && rightOK && left == right && leftExponent == rightExponent {
			return number
		}
	case map[string]interface{}:
		for key, item := range actual {
			actual[key] = Normalize(item)
		}
	case []interface{}:
		for index, item := range actual {
			actual[index] = Normalize(item)
		}
	}
	return value
}

// decimal compares decimal values without expanding powers of ten. Exponents
// outside int64 cannot match a finite float64 encoding for an in-memory token.
func decimal(token string) (string, int64, bool) {
	negative := strings.HasPrefix(token, "-")
	token = strings.TrimPrefix(token, "-")
	mantissa, exponentText := token, "0"
	if index := strings.IndexAny(token, "eE"); index >= 0 {
		mantissa, exponentText = token[:index], token[index+1:]
	}
	fraction := 0
	if index := strings.IndexByte(mantissa, '.'); index >= 0 {
		fraction = len(mantissa) - index - 1
		mantissa = mantissa[:index] + mantissa[index+1:]
	}
	digits := strings.TrimLeft(mantissa, "0")
	if digits == "" {
		return "0", 0, true
	}
	trimmed := strings.TrimRight(digits, "0")
	adjustment := int64(len(digits)-len(trimmed)) - int64(fraction)
	exponent, err := strconv.ParseInt(exponentText, 10, 64)
	if err != nil {
		return "", 0, false
	}
	adjusted := exponent + adjustment
	if adjustment > 0 && adjusted < exponent || adjustment < 0 && adjusted > exponent {
		return "", 0, false
	}
	if negative {
		trimmed = "-" + trimmed
	}
	return trimmed, adjusted, true
}
