package stringx

import (
	"os"
	"os/user"
	"strings"
)

func Interpolate(s string) (string, error) {
	host, err := os.Hostname()
	if err != nil {
		return "", err
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	repl := strings.NewReplacer(
		"{host}", host,
		"{user}", usr.Name,
		"{home}", usr.HomeDir,
	)

	return repl.Replace(s), nil
}

func Sanitize[T []byte | string](s T) T {
	var input []byte

	if value, ok := any(s).([]byte); ok {
		input = value
	} else {
		input = []byte(s)
	}

	for i, b := range input {
		if b != 0xa && (b < 0x20 || b > 0x7e) {
			input[i] = '_'
		}
	}

	return T(input)
}

func SplitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimRight(s, "\n")
	return strings.Split(s, "\n")
}
