package seq

import "github.com/illikainen/go-utils/src/fn"

func Contains[T comparable](elts []T, elt T) bool {
	for _, cur := range elts {
		if cur == elt {
			return true
		}
	}
	return false
}

func ContainsBy[T any](elts []T, fun func(elt T) bool) bool {
	for _, cur := range elts {
		if fun(cur) {
			return true
		}
	}
	return false
}

func FindBy[T any](elts []T, fun func(elt T) bool) (T, bool) {
	for _, cur := range elts {
		if fun(cur) {
			return cur, true
		}
	}

	var null T
	return null, false
}

func Filter[T comparable](elts []T, exclude ...T) []T {
	var result []T

	for _, cur := range elts {
		if !Contains(exclude, cur) {
			result = append(result, cur)
		}
	}

	return result
}

func FilterBy[T any](elts []T, fun func(elt T, idx int) bool) []T {
	var result []T

	for i, cur := range elts {
		if fun(cur, i) {
			result = append(result, cur)
		}
	}

	return result
}

func Uniq[T comparable](elts []T) []T {
	var result []T

	for _, cur := range elts {
		if !Contains(result, cur) {
			result = append(result, cur)
		}
	}

	return result
}

func Intersect[T comparable](first []T, second []T) []T {
	var result []T

	for _, cur := range Uniq(first) {
		if Contains(second, cur) {
			result = append(result, cur)
		}
	}

	return result
}

func Coalesce[T comparable](elts ...T) (T, bool) {
	var null T

	for _, cur := range elts {
		if cur != null {
			return cur, true
		}
	}

	return null, false
}

func MaxBy[T any](elts []T, fun func(elt T, max T) bool) T {
	var max T

	for i, cur := range elts {
		if i == 0 {
			max = cur
		}
		if fun(cur, max) {
			max = cur
		}
	}

	return max
}

func MinBy[T comparable](elts []T, fun func(elt T, min T) bool) T {
	return MaxBy(elts, fun)
}

func ExpandBy[T any, R any](elts []T, fun func(elt T, idx int) []R) []R {
	var result []R

	for i, cur := range elts {
		r := fun(cur, i)
		if r != nil {
			result = append(result, r...)
		}
	}

	return result
}

func Chunk[T any](elts []T, n int) [][]T {
	var result [][]T
	i := 0

	for i < len(elts) {
		add := fn.Min(n, len(elts)-i)
		result = append(result, elts[i:i+add])
		i += add
	}

	return result
}
