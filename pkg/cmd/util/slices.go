package util

// Map applies the given function to all slice elements.
func Map[S ~[]E, E any](s S, fn func(e E) E) S {
	out := make([]E, len(s))

	for i, e := range s {
		out[i] = fn(e)
	}

	return out
}
