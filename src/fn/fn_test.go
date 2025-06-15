package fn_test

import (
	"fmt"
	"testing"

	"github.com/illikainen/go-utils/src/fn"
)

func TestTernary(t *testing.T) {
	if fn.Ternary(true, "a", "b") != "a" {
		t.Fatal("Ternary()")
	}

	if fn.Ternary(false, "foo", "bar") != "bar" {
		t.Fatal("Ternary()")
	}
}

func TestMin(t *testing.T) {
	if fn.Min("a", "b") != "a" {
		t.Fatal("Min()")
	}

	if fn.Min(-50, 20) != -50 {
		t.Fatal("Min()")
	}

	if fn.Min(10.0, 20.0) != 10.0 {
		t.Fatal("Min()")
	}
}

func TestMax(t *testing.T) {
	if fn.Max("a", "b") != "b" {
		t.Fatal("Max()")
	}

	if fn.Max(-50, 20) != 20 {
		t.Fatal("Max()")
	}

	if fn.Max(10.0, 20.0) != 20.0 {
		t.Fatal("Max()")
	}
}

func TestMustSuccess(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Fatal("Must() should not fail")
		}
	}()

	fn.Must(nil)
}

func TestMustError(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Must() should fail")
		}
	}()

	fn.Must(fmt.Errorf("err"))
}

func TestMust1Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Fatal("Must1() should not fail")
		}
	}()

	if fn.Must1("foo", nil) != "foo" {
		t.Fatal("Must1() return value")
	}
}

func TestMust1Error(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Must1() should fail")
		}
	}()

	fn.Must1("foo", fmt.Errorf("err"))
}
