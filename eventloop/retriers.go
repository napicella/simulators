package main

type retrier interface {
	initCall()
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

func (t *fixedRetrier) initCall() {
	// do nothing
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

type circuitBreakerRetrier struct {
	r        retrier
	failures float64
	calls    float64
	maxRate  float64
}

func (t *circuitBreakerRetrier) initCall() {
	t.r = newFixedRetrier()
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

type tokenBucketRetrier struct {
	maxBucketSize  int
	numberOfTokens int
}

func (t *tokenBucketRetrier) initCall() {
	// do nothing
}

func (t *tokenBucketRetrier) recordSuccess() {
	t.numberOfTokens = min(t.numberOfTokens+1, t.maxBucketSize)
}

func (t *tokenBucketRetrier) recordFailure() {
	t.numberOfTokens = max(t.numberOfTokens-1, 0)
}

func (t *tokenBucketRetrier) shouldRetry() bool {
	return t.numberOfTokens > 0
}

func max(x, y int) int {
	if x >= y {
		return x
	}
	return y
}

func min(x, y int) int {
	if x <= y {
		return x
	}
	return y
}
