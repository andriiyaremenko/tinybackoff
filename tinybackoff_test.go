package tinybackoff

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	attempts = 4
	delay    = time.Millisecond * 100
)

func TestBackOff(t *testing.T) {
	t.Run("ConstantBackOff returns same delay", testConstantBackOff)
	t.Run("LinearBackOff returns delay with linear growth", testLinearBackOff)
	t.Run("PowerBackOff returns delay with growth equals power of `base`", testPowerBackOff)
	t.Run("ExponentialBackOff returns delay with exponential growth until it reaches `maxDelay`",
		testExponentialBackOff)
	t.Run("Combine can combine bock-offs in one back-off", testCombine)
}

func testConstantBackOff(t *testing.T) {
	assert := assert.New(t)
	delay := time.Millisecond * 100
	backOff := WithMaxAttempts(NewConstantBackOff(delay), uint64(attempts))

	for i := 0; i < attempts; i++ {
		assert.Equal(true, backOff.Continue())
		assert.Equal(delay, backOff.NextDelay())
	}

	assert.Equal(false, backOff.Continue())
}

func testLinearBackOff(t *testing.T) {
	assert := assert.New(t)
	multiplier := time.Millisecond * 50
	delay := time.Millisecond * 100
	backOff := NewLinearBackOff(delay, multiplier, uint64(attempts))

	for i := 0; i < attempts; i++ {
		attempt := i + 1
		expected := delay + multiplier*time.Duration(attempt)

		assert.Equal(expected, backOff.NextDelay())
	}
}

func testPowerBackOff(t *testing.T) {
	assert := assert.New(t)
	base := 2
	backOff := NewPowerBackOff(delay, uint64(base), uint64(attempts))

	for i := 0; i < attempts; i++ {
		attempt := i + 1
		expected := delay * time.Duration(math.Pow(float64(base), float64(attempt)))

		assert.Equal(expected, backOff.NextDelay())
	}
}

func testExponentialBackOff(t *testing.T) {
	assert := assert.New(t)
	maxDelay := time.Hour * 24
	attempts := 7
	backOff := NewExponentialBackOff(maxDelay, uint64(attempts))

	for i := 0; i < attempts; i++ {
		attempt := i + 1
		expected := time.Duration(float64(maxDelay) / math.Exp(float64(attempts-attempt)))

		assert.Equal(expected, backOff.NextDelay())
	}
}

func testCombine(t *testing.T) {
	assert := assert.New(t)
	now := time.Now()
	delay := now.Sub(now.Add(-time.Minute * 50))

	backOff := Combine(delay,
		WithMaxAttempts(NewConstantBackOff(time.Second*10), 2),
		WithMaxAttempts(NewLinearBackOff(time.Minute, time.Second*30, 5), 5),
	)

	for i := 0; backOff.Continue(); i++ {
		nextDelay := backOff.NextDelay()

		if i == 0 {
			assert.Equal(delay, nextDelay)
			continue
		}

		if i > 0 && i < 3 {
			assert.Equal(time.Second*10, nextDelay)
			continue
		}

		if i >= 3 && i < 8 {
			expected := time.Minute + time.Second*time.Duration(30*(i-2))
			assert.Equal(expected, nextDelay)
			continue
		}

		assert.Fail("Continue() was true for more times than expected: %d", i-1)
	}
}
