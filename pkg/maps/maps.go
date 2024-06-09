package maps

// Merge merges the given maps. If keys are contained in multiple maps, values from later maps take precedence.
func Merge[M ~map[K]V, K comparable, V any](mm ...M) M {
	if len(mm) == 0 {
		return nil
	}

	out := make(M, len(mm[0]))
	for _, m := range mm {
		for k, v := range m {
			out[k] = v
		}
	}

	return out
}
