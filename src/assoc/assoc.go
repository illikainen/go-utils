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
