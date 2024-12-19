package internal

func ChunkCh[T any](in <-chan T, size int) <-chan []T {
	out := make(chan []T)

	go func() {
		defer close(out)

		var chunk []T
		for v := range in {
			chunk = append(chunk, v)
			if len(chunk) == size {
				out <- chunk
				chunk = nil
			}
		}

		if len(chunk) > 0 {
			out <- chunk
		}
	}()

	return out
}
