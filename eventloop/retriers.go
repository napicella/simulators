package main

import "fmt"

type retrier interface {
	recordSuccess()
	recordFailure()
	shouldRetry() bool
}

func newFixedRetrier() retrier {
	return &fixedRetrier{
		maxAttempts: 3,
	}
}

// a simple retrier that allows retrying a fixed amount of times
type fixedRetrier struct {
	currentAttempt int
	maxAttempts    int
}

func (t *fixedRetrier) recordSuccess() {
	// do nothing
}

func (t *fixedRetrier) recordFailure() {
	t.currentAttempt++
}

func (t *fixedRetrier) shouldRetry() bool {
	return t.currentAttempt <= t.maxAttempts
}

func newCircuitBreakerRetrier() retrier {
	return &circuitBreakerRetrier{
		r:        newFixedRetrier(),
		failures: 0,
		calls:    0,
		maxRate:  10,
	}
}

type circuitBreakerRetrier struct {
	r        retrier
	failures float64
	calls    float64
	maxRate  float64
}

func (t *circuitBreakerRetrier) recordSuccess() {
	t.calls++
	t.r.recordSuccess()
}

func (t *circuitBreakerRetrier) recordFailure() {
	t.failures++
	t.calls++
	t.r.recordFailure()
}

func (t *circuitBreakerRetrier) shouldRetry() bool {
	if t.failures/t.calls < t.maxRate {
		return t.r.shouldRetry()
	}
	return false
}

type retrierFactoryName int

func (d retrierFactoryName) String() string {
	return [...]string{"fixed", "circuit-breaker"}[d]
}

const (
	fixedRetry     retrierFactoryName = iota
	circuitBreaker                    = iota
)

func getFactory(name retrierFactoryName) retrierFactory {
	switch name {
	case fixedRetry:
		return &fixedRetrierFactory{}
	case circuitBreaker:
		return &circuitBreakerRetrierFactory{}
	default:
		panic(fmt.Sprintf("invalid retrier name %s", name))
	}
}

type retrierFactory interface {
	get() retrier
}

type fixedRetrierFactory struct{}

func (t *fixedRetrierFactory) get() retrier {
	return newFixedRetrier()
}

type circuitBreakerRetrierFactory struct {
	r *circuitBreakerRetrier
}

func (t *circuitBreakerRetrierFactory) get() retrier {
	if t.r == nil {
		t.r = &circuitBreakerRetrier{
			r:        newFixedRetrier(),
			failures: 0,
			calls:    0,
			maxRate:  0.5,
		}
	} else {
		// workaround to re-create the fixed retrier used by the circuitBreakerRetrier
		// every team get is called (which is every time the client creates a new call)
		// TODO: the retried should have a method called - start or call which should do
		//   initialization that need to happen before the call is made
		t.r.r = newFixedRetrier()
	}

	return t.r
}
