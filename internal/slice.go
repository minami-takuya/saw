package internal

func SliceCh[T any](in []T) <-chan T {
	out := make(chan T)

	go func() {
		defer close(out)

		for _, v := range in {
			out <- v
		}
	}()
	return out
}
