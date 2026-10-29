package jsonvalue

import (
	"encoding/json"
	"math"
	"math/big"
	"strings"
	"testing"
)

func TestLegacyFloat64OnlyWhenJSONDecimalRoundTripIsExact(t *testing.T) {
	for _, token := range []string{"0", "-0", "7", "0.1", "1.25", "1e3", "1.00e+003", "9007199254740991", "-9007199254740991", "9.007199254740991e15", "1e-6", "-0e999999999999999999999999999999"} {
		t.Run(token, func(t *testing.T) {
			got := Normalize(json.Number(token))
			number, ok := got.(float64)
			if !ok {
				t.Fatalf("legacy extraction lost: %T %v", got, got)
			}
			if strings.HasPrefix(token, "-0") && number == 0 && !math.Signbit(number) {
				t.Fatal("negative zero lost")
			}
			encoded, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			if token == "-0e999999999999999999999999999999" {
				return
			}
			before, ok1 := new(big.Rat).SetString(token)
			after, ok2 := new(big.Rat).SetString(string(encoded))
			if !ok1 || !ok2 || before.Cmp(after) != 0 {
				t.Fatalf("decimal changed: %s -> %s", token, encoded)
			}
		})
	}
	for _, token := range []string{"9007199254740992", "-9007199254740992", "9007199254740992.0", "9.007199254740992e15", "-9.007199254740992e15", "1000000000000000100", "-1000000000000000100", "1.0000000000000001e18", "10000000000000001e2", "1e21", "9007199254740993", "9223372036854775807", "-9223372036854775808", "18446744073709551615", "0.12345678901234567890123456789", "1.234567890123456789e100", "1e400", "1e-400", "1e999999999999999999999999999999", "1e-999999999999999999999999999999"} {
		if got := Normalize(json.Number(token)); got != json.Number(token) {
			t.Fatalf("exact token changed: %s -> %T %v", token, got, got)
		}
	}
}

func TestNestedAndCustomValues(t *testing.T) {
	type declared struct{ Number json.Number }
	explicit := declared{Number: "7"}
	input := map[string]interface{}{"nested": []interface{}{json.Number("7"), json.Number("9007199254740993"), nil}, "typed": explicit, "text": "7", "bool": true}
	Normalize(input)
	values := input["nested"].([]interface{})
	if values[0].(float64) != 7 || values[1] != json.Number("9007199254740993") || values[2] != nil || input["typed"] != explicit || input["text"] != "7" || input["bool"] != true {
		t.Fatalf("changed shapes: %#v", input)
	}
}
