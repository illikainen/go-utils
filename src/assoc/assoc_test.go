package assoc_test

import (
	"testing"

	"github.com/illikainen/go-utils/src/assoc"
	"github.com/illikainen/go-utils/src/test"
)

func TestMerge(t *testing.T) {
	test.AssertEq(
		t,
		assoc.Merge(
			map[string]string{
				"foo":        "bar",
				"overridden": "old",
			},
			map[string]string{
				"hello":      "world",
				"another":    "value",
				"overridden": "new",
			},
		),
		map[string]string{
			"foo":        "bar",
			"hello":      "world",
			"another":    "value",
			"overridden": "new",
		},
	)
}
