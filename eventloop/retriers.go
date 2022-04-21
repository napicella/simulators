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

// circuitBreakerRetrier allows retries if the rate of failures is less than the maxRate.
// If the failure rate is less than maxRate, a fixedRetrier is used to determine whether
// or not the call should be retried (i.e. if it reached the maximum number of attempts)
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

// tokenBucketRetrier allows retrying a request as long as the number of tokens in the
// bucket is not zero. With this strategy a request is retried potentially a number of
// times equal to the number of tokens in the bucket
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

// tokenBucketFixedRetrier combines a tokenBucketRetrier with a fixedRetrier. That is, a
// request is allowed to be retried a fix number of times as long as the tokens in the
// token bucket is not zero
type tokenBucketFixedRetrier struct {
	tokenBucketRetrier
	fixedRetrier retrier
}

func (t *tokenBucketFixedRetrier) initCall() {
	t.tokenBucketRetrier.initCall()
	t.fixedRetrier = newFixedRetrier()
}

func (t *tokenBucketFixedRetrier) recordSuccess() {
	t.tokenBucketRetrier.recordSuccess()
}

func (t *tokenBucketFixedRetrier) recordFailure() {
	t.tokenBucketRetrier.recordFailure()
}

func (t *tokenBucketFixedRetrier) shouldRetry() bool {
	return t.tokenBucketRetrier.shouldRetry() && t.fixedRetrier.shouldRetry()
}
