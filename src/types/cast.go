package types

import (
	"math"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var ErrOutOfRange = errors.New("invalid signature")

func CheckedCast[T int, U int64 | uint32 | uint64](n T) (U, error) {
	var dst U

	switch any(dst).(type) {
	case int64:
		if n < math.MinInt64 || n > math.MaxInt64 {
			return 0, ErrOutOfRange
		}
	case uint32:
		if n < 0 || n > math.MaxUint32 {
			return 0, ErrOutOfRange
		}
	case uint64:
		// FIXME: math.MaxUint32 is the upper limit because:
		// cannot convert math.MaxUint64 (untyped int constant ...) to T
		if n < 0 || n > math.MaxUint32 || math.MaxUint32 > math.MaxUint64 {
			return 0, ErrOutOfRange
		}
	}

	return U(n), nil
}

func Cast[T int, U int64 | uint32 | uint64](n T) U {
	m, err := CheckedCast[T, U](n)
	if err != nil {
		log.Fatalf("%v", err)
	}
	return m
}
