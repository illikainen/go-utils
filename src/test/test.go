package test

import (
	"reflect"
	"testing"

	"github.com/pkg/errors"
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

func AssertErr(t *testing.T, actual error, expected error) {
	t.Helper()

	if !errors.Is(actual, expected) {
		t.Fatalf("%v != %v", actual, expected)
	}
}
