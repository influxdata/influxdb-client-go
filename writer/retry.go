package writer

import (
	"time"

	"github.com/influxdata/influxdb-client-go"
)

const defaultMaxAttempts = 5

// BackoffFunc is a function which when called with an
// attempt number returns a duration which should be
// waited for until a subsequent attempt is made
type BackoffFunc func(attempt int) time.Duration

// RetryWriter is a metrics writers which decorates other
// metrics writer implementations and automatically retries
// attempts to write metrics under certain error conditions
type RetryWriter struct {
	MetricsWriter

	sleep   func(time.Duration)
	now     func() time.Time
	backoff BackoffFunc

	maxAttempts int

	// Number of seconds to limit retry sleeping to. If zero there is no limit
	// imposed. If nonzero we will not sleep past this number of seconds since
	// the write attempt started. At the end of the limit there will be one
	// more write attempt.
	retrySleepLimit int
}

// NewRetryWriter returns a configured *RetryWriter which decorates
// the supplied MetricsWriter
func NewRetryWriter(w MetricsWriter, opts ...RetryOption) *RetryWriter {
	r := &RetryWriter{
		MetricsWriter: w,
		sleep:         time.Sleep,
		now:           time.Now,
		backoff:       func(int) time.Duration { return 0 },
		maxAttempts:   defaultMaxAttempts,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// sleepWithLimit sleeps for a given duration, but imposes a limit on the
// sleeping. If it is currently past the stop time then it does nothing and
// returns false (do not continue to try). Otherwise, sleeps either for
// duration, or until stop time, whichever comes sooner, then returns true
// (continue retrying).
func (r *RetryWriter) sleepWithLimit(duration time.Duration, stopAt time.Time) bool {
	if r.retrySleepLimit > 0 {
		allowedToSleep := stopAt.Sub(r.now())
		if allowedToSleep <= 0 {
			return false
		} else if duration > allowedToSleep {
			duration = allowedToSleep
		}
	}

	r.sleep(duration)
	return true
}

// Write delegates to underlying MetricsWriter and then
// automatically retries when errors occur
func (r *RetryWriter) Write(m ...influxdb.Metric) (n int, err error) {
	var stopAt time.Time
	if r.retrySleepLimit > 0 {
		stopAt = r.now().Add(time.Second * time.Duration(r.retrySleepLimit))
	}

	for i := 0; i < r.maxAttempts; i++ {
		n, err = r.MetricsWriter.Write(m...)
		if err == nil {
			return
		}

		ierr, ok := err.(*influxdb.Error)
		if !ok {
			return
		}

		switch ierr.Code {
		// retriable errors
		case influxdb.EUnavailable, influxdb.ETooManyRequests:
			if ierr.RetryAfter != nil {
				// given retry-after is configured attempt to sleep
				// for retry-after seconds
				// println( "-> retry after sleeping for", *ierr.RetryAfter, "seconds" )
				cont := r.sleepWithLimit(time.Duration(*ierr.RetryAfter)*time.Second, stopAt)
				if !cont {
					return
				}
			} else if backoffDuration := r.backoff(i + 1); backoffDuration > 0 {
				// given a backoff duration > 0
				// call sleep with backoff duration
				// println( "-> backoff sleeping for", duration )
				cont := r.sleepWithLimit(backoffDuration, stopAt)
				if !cont {
					return
				}
			}
		default:
			return
		}
	}

	return
}

// LinearBackoff returns a BackoffFunc which when called
// returns attempt * scale.
// e.g.
// LinearBackoff(time.Second)(5) returns 5 seconds
func LinearBackoff(scale time.Duration) BackoffFunc {
	return func(attempt int) time.Duration {
		return time.Duration(attempt) * scale
	}
}

// RetryOption is a functional option for the RetryWriters type
type RetryOption func(*RetryWriter)

// WithMaxAttempts sets the maximum number of attempts for a
// Write operation attempt
func WithMaxAttempts(maxAttempts int) RetryOption {
	return func(r *RetryWriter) {
		r.maxAttempts = maxAttempts
	}
}

// WithBackoff sets of the BackoffFunc on the RetryWriter
func WithBackoff(fn BackoffFunc) RetryOption {
	return func(r *RetryWriter) {
		r.backoff = fn
	}
}

// WithRetrySleepLimit sets the retry sleep limit. This optiona allows us to
// abort retry sleeps past some number of seconds.
func WithRetrySleepLimit(retrySleepLimit int) RetryOption {
	return func(r *RetryWriter) {
		r.retrySleepLimit = retrySleepLimit
	}
}
