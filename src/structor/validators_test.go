package structor_test

import (
	"encoding/base64"
	"path/filepath"
	"testing"

	"github.com/illikainen/go-utils/src/structor"
	"github.com/illikainen/go-utils/src/test"
)

func TestBase64(t *testing.T) {
	type s struct {
		Base64 string `validate:"base64"`
	}

	test.AssertErr(t, structor.Apply(&s{
		Base64: base64.StdEncoding.EncodeToString([]byte("foobar")),
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Base64: "invalid",
	}, nil), structor.ErrValidation)
}

func TestEither(t *testing.T) {
	type s struct {
		First  string
		Second string `validate:"either:First"`
		Third  string
	}

	test.AssertErr(t, structor.Apply(&s{
		First: "foobar",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Second: "foobar",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Third: "foobar",
	}, nil), structor.ErrValidation)
}

func TestMultipleEither(t *testing.T) {
	type s struct {
		First  string
		Second string `validate:"either:First:Third"`
		Third  string
	}

	test.AssertErr(t, structor.Apply(&s{
		First: "foobar",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Second: "foobar",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Third: "foobar",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{}, nil), structor.ErrValidation)
}

func TestEmail(t *testing.T) {
	type s struct {
		Email string `validate:"email"`
	}

	test.AssertErr(t, structor.Apply(&s{
		Email: "foo@example.invalid",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Email: "invalid",
	}, nil), structor.ErrValidation)
}

func TestExists(t *testing.T) {
	tmp := t.TempDir()

	type s struct {
		Path string `validate:"exists"`
		File string `validate:"exists:file"`
		Dir  string `validate:"exists:dir"`
	}

	test.AssertErr(t, structor.Apply(&s{
		Path: tmp,
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Dir: tmp,
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		File: tmp,
	}, nil), structor.ErrValidation)

	test.AssertErr(t, structor.Apply(&s{
		Path: filepath.Join(tmp, "enoent"),
	}, nil), structor.ErrValidation)
}

func TestGroup(t *testing.T) {
	type s struct {
		Group string `validate:"group"`
	}

	test.AssertErr(t, structor.Apply(&s{
		Group: "valid",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Group: "-invalid",
	}, nil), structor.ErrValidation)
}

func TestHex(t *testing.T) {
	type s struct {
		Hex string `validate:"hex"`
	}

	test.AssertErr(t, structor.Apply(&s{
		Hex: "0123456789abcdef",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Hex: "0123456789abcdefg",
	}, nil), structor.ErrValidation)
}

func TestHost(t *testing.T) {
	type s struct {
		Host string `validate:"host"`
	}

	test.AssertErr(t, structor.Apply(&s{
		Host: "foobar",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Host: "_foobar",
	}, nil), structor.ErrValidation)

	test.AssertErr(t, structor.Apply(&s{
		Host: "example.invalid",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Host: "foobar.invalid_",
	}, nil), structor.ErrValidation)
}

func TestIn(t *testing.T) {
	type s struct {
		In string `validate:"in:[\"foo\", \"bar\"]"`
	}

	test.AssertErr(t, structor.Apply(&s{
		In: "foo",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		In: "bar",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		In: "foobar",
	}, nil), structor.ErrValidation)
}

func TestIP(t *testing.T) {
	type s struct {
		IP       string `validate:"ip"`
		IPv4     string `validate:"ip:v4"`
		IPv6     string `validate:"ip:v6"`
		CIDR     string `validate:"ip:cidr"`
		CIDRv4   string `validate:"ip:v4:cidr"`
		CIDRv6   string `validate:"ip:v6:cidr"`
		CIDRorIP string `validate:"ip:autocidr"`
	}

	test.AssertErr(t, structor.Apply(&s{
		IP:     "127.0.0.1",
		IPv4:   "127.0.0.1",
		IPv6:   "::1",
		CIDR:   "127.0.0.1/8",
		CIDRv4: "127.0.0.1/32",
		CIDRv6: "::1/128",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		IPv6: "127.0.0.1",
	}, nil), structor.ErrValidation)

	test.AssertErr(t, structor.Apply(&s{
		CIDRv6: "127.0.0.1/8",
	}, nil), structor.ErrValidation)

	test.AssertErr(t, structor.Apply(&s{
		IPv4: "::1",
	}, nil), structor.ErrValidation)

	test.AssertErr(t, structor.Apply(&s{
		CIDRv4: "::1/128",
	}, nil), structor.ErrValidation)

	test.AssertErr(t, structor.Apply(&s{
		IP: "foobar",
	}, nil), structor.ErrValidation)

	test.AssertErr(t, structor.Apply(&s{
		CIDR: "foobar",
	}, nil), structor.ErrValidation)

	test.AssertErr(t, structor.Apply(&s{
		CIDRorIP: "127.0.0.1",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		CIDRorIP: "127.0.0.1/8",
	}, nil), nil)
}

func TestMinMax(t *testing.T) {
	type s struct {
		Value int `validate:"min:2,max:5"`
	}

	test.AssertErr(t, structor.Apply(&s{
		Value: 1,
	}, nil), structor.ErrValidation)

	test.AssertErr(t, structor.Apply(&s{
		Value: 2,
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Value: 3,
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Value: 5,
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Value: 6,
	}, nil), structor.ErrValidation)
}

func TestPort(t *testing.T) {
	type s struct {
		Value int `validate:"port"`
	}

	test.AssertErr(t, structor.Apply(&s{
		Value: -1,
	}, nil), structor.ErrValidation)

	test.AssertErr(t, structor.Apply(&s{
		Value: 0,
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Value: 22,
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Value: 65535,
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Value: 65536,
	}, nil), structor.ErrValidation)
}

func TestPortrange(t *testing.T) {
	type s struct {
		Value string `validate:"port"`
	}

	test.AssertErr(t, structor.Apply(&s{
		Value: "0",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Value: "65535",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Value: "65536",
	}, nil), structor.ErrValidation)

	test.AssertErr(t, structor.Apply(&s{
		Value: "1-65535",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Value: "1-65536",
	}, nil), structor.ErrValidation)

	test.AssertErr(t, structor.Apply(&s{
		Value: "1024-22",
	}, nil), structor.ErrValidation)
}

func TestPrintable(t *testing.T) {
	type s struct {
		ASCII   string `validate:"printable"`
		Unicode string `validate:"printable:unicode"`
	}

	test.AssertErr(t, structor.Apply(&s{
		ASCII: "foo",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		ASCII: "",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		ASCII: "foo\x1b",
	}, nil), structor.ErrValidation)

	test.AssertErr(t, structor.Apply(&s{
		Unicode: "foo",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Unicode: "",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Unicode: "foo\x1b",
	}, nil), structor.ErrValidation)
}

func TestRequired(t *testing.T) {
	type s struct {
		Value string `validate:"required"`
	}

	test.AssertErr(t, structor.Apply(&s{}, nil), structor.ErrValidation)

	test.AssertErr(t, structor.Apply(&s{
		Value: "foo",
	}, nil), nil)
}

func TestRegexp(t *testing.T) {
	type s struct {
		Value string `validate:"rx:^[a-z0-9]+$"`
	}

	test.AssertErr(t, structor.Apply(&s{
		Value: "foo",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		Value: "foO",
	}, nil), structor.ErrValidation)
}

func TestURI(t *testing.T) {
	type s struct {
		URI         string `validate:"uri"`
		HTTPS       string `validate:"uri:scheme=https"`
		HTTPorHTTPS string `validate:"uri:scheme=http|https"`
	}

	test.AssertErr(t, structor.Apply(&s{
		URI: "https://example.invalid",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		HTTPS: "https://example.invalid",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		HTTPS: "http://example.invalid",
	}, nil), structor.ErrValidation)

	test.AssertErr(t, structor.Apply(&s{
		HTTPorHTTPS: "http://example.invalid",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		HTTPorHTTPS: "https://example.invalid",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		HTTPorHTTPS: "ssh://example.invalid",
	}, nil), structor.ErrValidation)

	test.AssertErr(t, structor.Apply(&s{
		URI: "example.invalid",
	}, nil), structor.ErrValidation)
}

func TestUser(t *testing.T) {
	type s struct {
		User string `validate:"user"`
	}

	test.AssertErr(t, structor.Apply(&s{
		User: "valid",
	}, nil), nil)

	test.AssertErr(t, structor.Apply(&s{
		User: "-invalid",
	}, nil), structor.ErrValidation)
}
