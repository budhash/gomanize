package core

import (
	"reflect"
	"testing"
)

// Every Options field is either a component switch (or Debug) or a style option.
// Setting any style option must make DefaultStyle false, so a newly added style
// field cannot slip past default-style-only components unnoticed.
func TestOptionsDefaultStyle(t *testing.T) {
	components := map[string]bool{"SchwaModel": true, "Lexicon": true, "Rerank": true, "Debug": true}
	if !(Options{}).DefaultStyle() {
		t.Fatal("zero options must be default style")
	}
	typ := reflect.TypeOf(Options{})
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Type.Kind() != reflect.Bool {
			t.Fatalf("Options.%s is not a bool; extend DefaultStyle and this test", field.Name)
		}
		var o Options
		reflect.ValueOf(&o).Elem().Field(i).SetBool(true)
		if got, want := o.DefaultStyle(), components[field.Name]; got != want {
			t.Errorf("Options{%s: true}.DefaultStyle() = %v, want %v", field.Name, got, want)
		}
	}
}
