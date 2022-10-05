package database

func Map[T any, R any](l []T, f func(T) R) []R {
	r := make([]R, len(l))
	for i, v := range l {
		r[i] = f(v)
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
