package influxclient

import (
	"math/rand"
)

// RetryStrategy is a strategy for calculating retry delays.
type RetryStrategy interface {
	// NextDelay returns delay for a next retry
	//  - error - reason for retrying
	//  - failedAttempts - a count of already failed attempts, 1 being the first
	// Returns milliseconds to wait before retrying
	NextDelay(err error, failedAttempts int) int
	// Success implementation should reset its state, this is mandatory to call upon success
	Success()
}

// NewRetryStrategyF factory function creates a new RetryStrategy
type NewRetryStrategyF func(params RetryParams) RetryStrategy

func NewDefaultRetryStrategy(params RetryParams) RetryStrategy {
	return &retryStrategy{
		params:       params,
		currentDelay: 0,
	}
}

var retryStrategyCreator NewRetryStrategyF = NewDefaultRetryStrategy

func GetRetryStrategyFactory() NewRetryStrategyF {
	return retryStrategyCreator
}

func SetRetryStrategyFactory(f NewRetryStrategyF) {
	retryStrategyCreator = f
}

type retryStrategy struct {
	params       RetryParams
	currentDelay int
}

// NextDelay calculates retry delay.
//
//	It calculates as random value within the interval
//
// [retry_interval * exponential_base^(attempts) and retry_interval * exponential_base^(attempts+1)]
// and adds jitter
func (r *retryStrategy) NextDelay(err error, failedAttempts int) int {
	if se, ok := err.(*ServerError); ok && se.RetryAfter > 0 {
		return se.RetryAfter * 1000
	}
	minDelay := r.params.RetryInterval * powi(r.params.ExponentialBase, failedAttempts)
	maxDelay := r.params.RetryInterval * powi(r.params.ExponentialBase, failedAttempts+1)
	diff := maxDelay - minDelay
	if diff <= 0 { //check overflows
		return r.params.MaxRetryInterval
	}
	retryDelay := rand.Intn(diff) + minDelay
	if retryDelay > r.params.MaxRetryInterval {
		retryDelay = r.params.MaxRetryInterval
	}
	if r.params.RetryJitter > 0 {
		retryDelay = retryDelay + rand.Intn(r.params.RetryJitter)
	}
	r.currentDelay = retryDelay
	return retryDelay
}

// powi computes x**y
func powi(x, y int) int {
	p := 1
	if y == 0 {
		return 1
	}
	for i := 1; i <= y; i++ {
		p = p * x
	}
	return p
}

func (r *retryStrategy) Success() {
	r.currentDelay = 0
}
