package test

import (
	"reflect"
	"testing"
)

func AssertEq(t *testing.T, x any, y any) {
	t.Helper()

	if !reflect.DeepEqual(x, y) {
		t.Fatalf("%v != %v", x, y)
	}
}

func AssertNe(t *testing.T, x any, y any) {
	t.Helper()

	if reflect.DeepEqual(x, y) {
		t.Fatalf("%v != %v", x, y)
	}
}
