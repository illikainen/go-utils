package fn

import "golang.org/x/exp/constraints"

func Ternary[T any](cond bool, trueVal T, falseVal T) T {
	if cond {
		return trueVal
	}
	return falseVal
}

func Min[T constraints.Ordered](a T, b T) T {
	return Ternary(a < b, a, b)
}

func Max[T constraints.Ordered](a T, b T) T {
	return Ternary(a > b, a, b)
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Must1[T any](val T, err error) T {
	Must(err)
	return val
}
