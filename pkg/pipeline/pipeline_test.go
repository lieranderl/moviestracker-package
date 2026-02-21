package pipeline

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProducerEmitsAllValues(t *testing.T) {
	ctx := context.Background()
	input := []int{1, 2, 3}

	out, err := Producer(ctx, input)
	require.NoError(t, err)

	var got []int
	for v := range out {
		got = append(got, v)
	}

	require.Equal(t, input, got)
}

func TestStepProcessesValues(t *testing.T) {
	ctx := context.Background()

	input, err := Producer(ctx, []int{1, 2, 3})
	require.NoError(t, err)

	out, errs := Step(ctx, input, func(v int) (int, error) {
		return v * 2, nil
	}, 2)

	results := make([]int, 0, 3)
	for out != nil || errs != nil {
		select {
		case v, ok := <-out:
			if !ok {
				out = nil
				continue
			}
			results = append(results, v)
		case e, ok := <-errs:
			if !ok {
				errs = nil
				continue
			}
			require.NoError(t, e)
		}
	}

	require.ElementsMatch(t, []int{2, 4, 6}, results)
}

func TestStepEmitsErrors(t *testing.T) {
	ctx := context.Background()

	input, err := Producer(ctx, []int{1, 2, 3})
	require.NoError(t, err)

	out, errs := Step(ctx, input, func(v int) (int, error) {
		if v == 2 {
			return 0, errors.New("bad value")
		}
		return v, nil
	}, 3)

	var (
		gotValues []int
		gotErr    error
	)

	for out != nil || errs != nil {
		select {
		case v, ok := <-out:
			if !ok {
				out = nil
				continue
			}
			gotValues = append(gotValues, v)
		case e, ok := <-errs:
			if !ok {
				errs = nil
				continue
			}
			if gotErr == nil {
				gotErr = e
			}
		}
	}

	require.Error(t, gotErr)
	require.ElementsMatch(t, []int{1, 3}, gotValues)
}

func TestMergeCombinesChannels(t *testing.T) {
	ctx := context.Background()

	c1 := make(chan int, 2)
	c2 := make(chan int, 1)
	c1 <- 1
	c1 <- 2
	c2 <- 3
	close(c1)
	close(c2)

	out := Merge(ctx, c1, c2)

	var got []int
	for v := range out {
		got = append(got, v)
	}

	require.ElementsMatch(t, []int{1, 2, 3}, got)
}
