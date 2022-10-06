package database

func Map[T any, R any](l []T, f func(T) R) []R {
	r := make([]R, len(l))
	for i, v := range l {
		r[i] = f(v)
	}
	return r
}

func Filter[T any](l []T, f func(T) bool) []T {
	var r []T
	for _, v := range l {
		if f(v) {
			r = append(r, v)
		}
	}
	return r
}

func MapErr[T any, R any](l []T, f func(T) (R, error)) ([]R, error) {
	r := make([]R, len(l))
	for i, v := range l {
		m, err := f(v)
		if err != nil {
			return nil, err
		}
		r[i] = m
	}
	return r, nil
}
