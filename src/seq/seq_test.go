package seq_test

import (
	"fmt"
	"testing"

	"github.com/illikainen/go-utils/src/seq"
	"github.com/illikainen/go-utils/src/test"
)

func TestContains(t *testing.T) {
	if seq.Contains(nil, "foo") {
		t.Fatal("Contains()")
	}

	if seq.Contains([]string{}, "foo") {
		t.Fatal("Contains()")
	}

	if seq.Contains(nil, "foo") {
		t.Fatal("Contains()")
	}

	if !seq.Contains([]string{"foo"}, "foo") {
		t.Fatal("Contains()")
	}

	if !seq.Contains([]string{"foo", "bar", "baz"}, "foo") {
		t.Fatal("Contains()")
	}

	if !seq.Contains([]string{"bar", "foo", "baz"}, "foo") {
		t.Fatal("Contains()")
	}

	if !seq.Contains([]string{"bar", "baz", "foo"}, "foo") {
		t.Fatal("Contains()")
	}

	if seq.Contains([]string{"foo", "bar", "baz"}, "") {
		t.Fatal("Contains()")
	}
}

func TestContainsBy(t *testing.T) {
	if seq.ContainsBy(nil, func(elt string) bool {
		return false
	}) {
		t.Fatal("ContainsBy()")
	}

	if seq.ContainsBy([]string{"foo", "bar"}, func(elt string) bool {
		return false
	}) {
		t.Fatal("ContainsBy()")
	}

	if !seq.ContainsBy([]string{"foo", "bar"}, func(elt string) bool {
		return elt == "bar"
	}) {
		t.Fatal("ContainsBy()")
	}
}

func TestFindBy(t *testing.T) {
	if s, ok := seq.FindBy(nil, func(elt string) bool {
		return false
	}); s != "" || ok {
		t.Fatal("FindBy()")
	}

	if s, ok := seq.FindBy([]string{"foo", "bar"}, func(elt string) bool {
		return false
	}); s != "" || ok {
		t.Fatal("FindBy()")
	}

	if s, ok := seq.FindBy([]string{"foo"}, func(elt string) bool {
		return elt == "foo"
	}); s != "foo" || !ok {
		t.Fatal("FindBy()")
	}

	if s, ok := seq.FindBy([]string{"foo", "bar"}, func(elt string) bool {
		return elt == "bar"
	}); s != "bar" || !ok {
		t.Fatal("FindBy()")
	}
}

func TestFilter(t *testing.T) {
	test.AssertEq(
		t,
		seq.Filter([]string{"foo", "bar", "baz", "foo", "bar"}, "foo"),
		[]string{"bar", "baz", "bar"},
	)

	test.AssertEq(
		t,
		seq.Filter([]string{"foo", "bar", "baz"}, ""),
		[]string{"foo", "bar", "baz"},
	)
}

func TestFilterBy(t *testing.T) {
	test.AssertEq(
		t,
		seq.FilterBy(
			[]string{"foo", "bar", "baz", "foo", "bar"},
			func(elt string, _ int) bool {
				return elt != "foo"
			},
		),
		[]string{"bar", "baz", "bar"},
	)

	test.AssertEq(
		t,
		seq.FilterBy(
			[]string{"foo", "bar", "baz"},
			func(elt string, _ int) bool {
				return true
			},
		),
		[]string{"foo", "bar", "baz"},
	)

	test.AssertEq(
		t,
		seq.FilterBy(
			[]string{"foo", "bar", "baz"},
			func(elt string, _ int) bool {
				return false
			},
		),
		*new([]string),
	)
}

func TestUniq(t *testing.T) {
	test.AssertEq(
		t,
		seq.Uniq([]string{"foo", "bar", "bar", "baz"}),
		[]string{"foo", "bar", "baz"},
	)

	test.AssertEq(
		t,
		seq.Uniq([]string{"foo", "foo"}),
		[]string{"foo"},
	)

	test.AssertEq(
		t,
		seq.Uniq([]string{"foo"}),
		[]string{"foo"},
	)
}

func TestIntersect(t *testing.T) {
	test.AssertEq(
		t,
		seq.Intersect([]string{"foo", "bar", "bar", "baz"}, []string{"foo", "bar"}),
		[]string{"foo", "bar"},
	)

	test.AssertEq(
		t,
		seq.Intersect([]string{"foo", "bar", "bar", "baz"}, []string{"bar", "b0rk"}),
		[]string{"bar"},
	)

	if len(seq.Intersect([]string{"foo", "bar", "bar", "baz"}, []string{"bar123"})) != 0 {
		t.Fatal("Intersect()")
	}
}

func TestCoalesce(t *testing.T) {
	s, ok := seq.Coalesce("", "", "foo")
	test.AssertEq(t, ok, true)
	test.AssertEq(t, s, "foo")

	s, ok = seq.Coalesce("", "", "")
	test.AssertEq(t, ok, false)
	test.AssertEq(t, s, "")

	n, ok := seq.Coalesce(0, 1)
	test.AssertEq(t, ok, true)
	test.AssertEq(t, n, 1)

	n, ok = seq.Coalesce(0)
	test.AssertEq(t, ok, false)
	test.AssertEq(t, n, 0)
}

func TestMaxBy(t *testing.T) {
	test.AssertEq(
		t,
		seq.MaxBy([]string{"string1", "s2", "string3"}, func(cur string, max string) bool {
			return len(cur) > len(max)
		}),
		"string1",
	)

	test.AssertEq(
		t,
		seq.MaxBy([]string{"", "foo"}, func(cur string, max string) bool {
			return len(cur) > len(max)
		}),
		"foo",
	)

	test.AssertEq(
		t,
		seq.MaxBy([]string{}, func(cur string, max string) bool {
			return len(cur) > len(max)
		}),
		"",
	)
}

func TestMinBy(t *testing.T) {
	test.AssertEq(
		t,
		seq.MinBy([]string{"string2", "s1", "s3"}, func(cur string, min string) bool {
			return len(cur) < len(min)
		}),
		"s1",
	)

	test.AssertEq(
		t,
		seq.MinBy([]string{"foo", ""}, func(cur string, min string) bool {
			return len(cur) < len(min)
		}),
		"",
	)

	test.AssertEq(
		t,
		seq.MinBy([]string{}, func(cur string, min string) bool {
			return len(cur) < len(min)
		}),
		"",
	)
}

func TestExpandBy(t *testing.T) {
	test.AssertEq(
		t,
		seq.ExpandBy([]string{"foo", "bar", "baz"}, func(elt string, _ int) []byte {
			return []byte(elt)
		}),
		[]byte("foobarbaz"),
	)

	test.AssertEq(
		t,
		seq.ExpandBy([]int{0, 2, 4, 6}, func(elt int, _ int) []string {
			return []string{fmt.Sprintf("%d", elt), fmt.Sprintf("%d", elt+1)}
		}),
		[]string{"0", "1", "2", "3", "4", "5", "6", "7"},
	)
}

func TestChunk(t *testing.T) {
	test.AssertEq(
		t,
		seq.Chunk([]byte("foobarbaz"), 3),
		[][]byte{[]byte("foo"), []byte("bar"), []byte("baz")},
	)

	test.AssertEq(
		t,
		seq.Chunk([]byte("foobarbaz12"), 3),
		[][]byte{[]byte("foo"), []byte("bar"), []byte("baz"), []byte("12")},
	)

	test.AssertEq(
		t,
		seq.Chunk([]string{"f", "o", "o", "b", "a", "r", "b", "a", "z"}, 3),
		[][]string{[]string{"f", "o", "o"}, []string{"b", "a", "r"}, []string{"b", "a", "z"}},
	)

	test.AssertEq(
		t,
		seq.Chunk([]string{"f", "o", "o", "b", "a", "r", "b", "a", "z", "1", "2"}, 3),
		[][]string{[]string{"f", "o", "o"}, []string{"b", "a", "r"}, []string{"b", "a", "z"}, []string{"1", "2"}},
	)

}
