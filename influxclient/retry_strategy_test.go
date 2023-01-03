package influxclient

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPowi(t *testing.T) {
	assert.EqualValues(t, 1, powi(10, 0))
	assert.EqualValues(t, 10, powi(10, 1))
	assert.EqualValues(t, 4, powi(2, 2))
	assert.EqualValues(t, 1, powi(1, 2))
	assert.EqualValues(t, 125, powi(5, 3))
}

func TestRetryStrategy_NextDelay(t *testing.T) {
	rs := GetRetryStrategyFactory()(DefaultRetryParams)
	assertBetween(t, rs.NextDelay(nil, 0), 5_000, 10_000)
	assertBetween(t, rs.NextDelay(nil, 1), 10_000, 20_000)
	assertBetween(t, rs.NextDelay(nil, 2), 20_000, 40_000)
	assertBetween(t, rs.NextDelay(nil, 3), 40_000, 80_000)
	assertBetween(t, rs.NextDelay(nil, 4), 80_000, 125_000)

	for i := 5; i < 200; i++ { //test also limiting higher values
		assert.EqualValues(t, 125_000, rs.NextDelay(nil, i))
	}

	assert.EqualValues(t, 30_000, rs.NextDelay(&ServerError{RetryAfter: 30}, 1))
	assertBetween(t, rs.NextDelay(NewServerError("error"), 1), 10_000, 20_000)
	params := DefaultRetryParams
	params.RetryInterval = 100
	rs = GetRetryStrategyFactory()(params)
	assertBetween(t, rs.NextDelay(nil, 0), 100, 200)
}

type custRS struct {
}

func newCustRS(params RetryParams) RetryStrategy {
	return &custRS{}
}

func (c *custRS) NextDelay(err error, failedAttempts int) int {
	return (failedAttempts + 1) * 10
}

func (c *custRS) Success() {
}

func TestCustomRetryStrategy(t *testing.T) {
	oldRs := GetRetryStrategyFactory()
	defer SetRetryStrategyFactory(oldRs)
	SetRetryStrategyFactory(newCustRS)
	rs := GetRetryStrategyFactory()(DefaultRetryParams)
	assert.EqualValues(t, 10, rs.NextDelay(nil, 0))
	assert.EqualValues(t, 110, rs.NextDelay(nil, 10))

}

func assertBetween(t *testing.T, val, min, max int) {
	t.Helper()
	assert.True(t, val >= min && val <= max, fmt.Sprintf("%d is outside <%d;%d>", val, min, max))
}
