package pipeline

import (
	"context"
	"sync"

	"golang.org/x/sync/semaphore"
)

func Step[In any, Out any](
	ctx context.Context,
	inputChannel <-chan In,
	fn func(In) (Out, error),
	limit int64,
) (chan Out, chan error) {
	if limit < 1 {
		limit = 1
	}

	outputChannel := make(chan Out)
	errorChannel := make(chan error)

	// Use all CPU cores to maximize efficiency. We'll set the limit to 2 so you
	// can see the values being processed in batches of 2 at a time, in parallel
	// limit := int64(runtime.NumCPU())
	sem1 := semaphore.NewWeighted(limit)

	go func() {
		defer close(outputChannel)
		defer close(errorChannel)

		for {
			var s In
			var ok bool

			select {
			case <-ctx.Done():
				return
			case s, ok = <-inputChannel:
				if !ok {
					if err := sem1.Acquire(ctx, limit); err != nil {
						select {
						case errorChannel <- err:
						case <-ctx.Done():
						}
					}
					return
				}
			}

			if err := sem1.Acquire(ctx, 1); err != nil {
				select {
				case errorChannel <- err:
				case <-ctx.Done():
				}
				return
			}

			go func(s In) {
				defer sem1.Release(1)

				result, err := fn(s)
				if err != nil {
					select {
					case errorChannel <- err:
					case <-ctx.Done():
					}
				} else {
					select {
					case outputChannel <- result:
					case <-ctx.Done():
					}
				}
			}(s)
		}
	}()

	return outputChannel, errorChannel
}

func Merge[T any](ctx context.Context, cs ...<-chan T) <-chan T {
	var wg sync.WaitGroup
	out := make(chan T)

	output := func(c <-chan T) {
		defer wg.Done()
		for n := range c {
			select {
			case out <- n:
			case <-ctx.Done():
				return
			}
		}
	}

	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
