package main

import (
	"fmt"
)

const (
	fixedRetry     retrierFactoryName = iota
	circuitBreaker                    = iota
	tokenBucket
)

type retrierFactoryName int

func (d retrierFactoryName) String() string {
	return [...]string{"fixed", "circuit-breaker", "token-bucket"}[d]
}

func getFactory(name retrierFactoryName) retrierFactory {
	switch name {
	case fixedRetry:
		return &fixedRetrierFactory{}
	case circuitBreaker:
		return &circuitBreakerRetrierFactory{}
	case tokenBucket:
		return &tokenBucketFactory{}
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
	}
	return t.r
}

type tokenBucketFactory struct {
	r *tokenBucketRetrier
}

func (t *tokenBucketFactory) get() retrier {
	if t.r == nil {
		t.r = &tokenBucketRetrier{
			maxBucketSize:  10,
			numberOfTokens: 10,
		}
	}
	return t.r
}
