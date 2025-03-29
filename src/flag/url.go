package flag

import (
	"net/url"
)

type URL struct {
	Value *url.URL
}

func (u *URL) Set(value string) error {
	uri, err := url.Parse(value)
	if err != nil {
		return err
	}

	u.Value = uri
	return nil
}

func (u *URL) String() string {
	if u.Value != nil {
		return u.Value.String()
	}
	return ""
}

func (u *URL) Type() string {
	return "url"
}
