package assoc

func Merge[K comparable, V any](maps ...map[K]V) map[K]V {
	result := map[K]V{}

	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}

func HasKey[K comparable, V any](m map[K]V, key K) bool {
	_, ok := m[key]
	return ok
}
