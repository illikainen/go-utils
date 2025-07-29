package structor_test

import (
	"testing"

	"github.com/illikainen/go-utils/src/structor"
	"github.com/illikainen/go-utils/src/test"
)

func TestParseValidatorTag(t *testing.T) {
	tags, err := structor.ParseValidatorTag("foo")
	test.AssertEq(t, err, nil)
	test.AssertEq(t, tags, []*structor.ValidatorTag{
		&structor.ValidatorTag{
			Name: "foo",
			Args: []string{},
		},
	})

	tags, err = structor.ParseValidatorTag("foo:bar")
	test.AssertEq(t, err, nil)
	test.AssertEq(t, tags, []*structor.ValidatorTag{
		&structor.ValidatorTag{
			Name: "foo",
			Args: []string{"bar"},
		},
	})

	tags, err = structor.ParseValidatorTag("foo:bar:baz")
	test.AssertEq(t, err, nil)
	test.AssertEq(t, tags, []*structor.ValidatorTag{
		&structor.ValidatorTag{
			Name: "foo",
			Args: []string{"bar", "baz"},
		},
	})

	tags, err = structor.ParseValidatorTag("foo:bar:baz,other:[\"111\", \"222\"]")
	test.AssertEq(t, err, nil)
	test.AssertEq(t, tags, []*structor.ValidatorTag{
		&structor.ValidatorTag{
			Name: "foo",
			Args: []string{"bar", "baz"},
		},
		&structor.ValidatorTag{
			Name: "other",
			Args: []string{"[\"111\", \"222\"]"},
		},
	})
}
