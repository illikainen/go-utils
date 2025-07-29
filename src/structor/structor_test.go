package structor_test

import (
	"testing"

	"github.com/illikainen/go-utils/src/structor"
	"github.com/illikainen/go-utils/src/test"
)

func TestStructor(t *testing.T) {
	type last struct {
		Ints           []int   `default:"[9,8,7]" validate:"min:1,max:9"`
		String         *string `default:"foobar"`
		StringChoice   *string `default:"foo" validate:"in:[\"foo\", \"bar\"]"`
		IntChoice      *int    `default:"3" validate:"in:[3,4,5]"`
		SliceIntChoice []int   `default:"[5,7]" validate:"in:[5,6,7,8]"`
		Plain          string  `validator:"rx:^[a-z]+$"`
	}
	type inner struct {
		Name            string  `validate:"rx:^[a-z,]+$"`
		IP              *string `default:"127.0.0.1" validate:"ip"`
		Last            *last
		LastWithDefault *last `default:"{\"String\": \"ohai\"}"`
	}
	type good struct {
		String *string `default:"a string"`
		Inner  inner
	}

	actual := good{}
	expected := good{
		String: ptr("a string"),
		Inner: inner{
			IP: ptr("127.0.0.1"),
			LastWithDefault: &last{
				Ints:           []int{9, 8, 7},
				String:         ptr("ohai"),
				StringChoice:   ptr("foo"),
				IntChoice:      ptr(3),
				SliceIntChoice: []int{5, 7},
				Plain:          "",
			},
		},
	}
	err := structor.Apply(&actual, nil)
	test.AssertEq(t, err, nil)
	test.AssertEq(t, actual, expected)

}

func ptr[T int | string](v T) *T {
	return &v
}
