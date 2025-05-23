package internal

import (
	"bufio"
	"io"
)

func ReadByLine(in io.Reader) <-chan string {
	ch := make(chan string, 100)

	go func() {
		defer close(ch)

		scanner := bufio.NewScanner(in)
		for scanner.Scan() {
			ch <- scanner.Text()
		}
	}()
	return ch
}
