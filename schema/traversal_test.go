package schema

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

type traversalChild struct{ Value string }
type traversalInput struct{ Child traversalChild }
type traversalRecursive struct {
	Value    string
	Children []*traversalRecursive
}

func TestLegacyTraversalIndependentCompilations(t *testing.T) {
	type useCase struct {
		description string
		input       any
	}
	for _, tc := range []useCase{
		{"ordinary nested object", traversalInput{}},
		{"recursive array", struct{ Root traversalRecursive }{}},
	} {
		t.Run(tc.description, func(t *testing.T) {
			var expected ToolInputSchema
			if err := expected.Load(tc.input); err != nil {
				t.Fatal(err)
			}
			entered, release := make(chan struct{}), make(chan struct{})
			var once sync.Once
			unblock := func() { once.Do(func() { close(release) }) }
			defer unblock()
			done := make(chan ToolInputSchema, 1)
			go func() {
				first := ToolInputSchema{Type: "object"}
				blocked := false
				// Enter reflection directly so the pause occurs during traversal,
				// not Load's preliminary field-name validation.
				first.Properties, first.Required = StructToProperties(reflect.TypeOf(tc.input), WithSkipFieldHook(func(f reflect.StructField) bool {
					if f.Name == "Value" && !blocked {
						blocked = true
						close(entered)
						<-release
					}
					return false
				}))
				done <- first
			}()
			select {
			case <-entered:
			case <-time.After(5 * time.Second):
				t.Fatal("first traversal never entered nested object")
			}
			var actual ToolInputSchema
			if err := actual.Load(tc.input); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(expected, actual) {
				t.Errorf("concurrent traversal lost schema: expected %#v, got %#v", expected, actual)
			}
			unblock()
			first := <-done
			if !reflect.DeepEqual(expected, first) {
				t.Error("first traversal changed after concurrent compilation")
			}
		})
	}
}
