package influxclient

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type coll struct {
	data  []byte
	count int
}

func TestLinesCount(t *testing.T) {
	assert.EqualValues(t, 1, linesCount([]byte("a")))
	assert.EqualValues(t, 1, linesCount([]byte("a\n")))
	assert.EqualValues(t, 2, linesCount([]byte("a\nb")))
	assert.EqualValues(t, 2, linesCount([]byte("a\nb\n")))
}

func TestRetryBufferFutureRetries(t *testing.T) {
	var mu sync.Mutex
	input := make([]coll, 0, 10)
	output := make([]coll, 0, 10)

	rb := NewRetryBuffer(100, func(lines []byte, retryCountDown int, expires time.Time) {
		mu.Lock()
		defer mu.Unlock()
		output = append(output, coll{lines, retryCountDown})
	}, nil)
	for i := 0; i < 10; i++ {
		s := []byte(fmt.Sprintf("a%d", i))
		input = append(input, coll{s, i})
		rb.AddLines(s, i, 2, time.Now().Add(time.Second))
	}
	waitForCondition(t, 100, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(output) >= 10
	})
	assert.True(t, len(output) == 10)
	// Empty lines are ignored
	rb.AddLines([]byte{}, 1, 1, time.Now().Add(time.Second))
	assert.EqualValues(t, 0, rb.Close())
	// Ignored after close
	rb.AddLines([]byte("x"), 1, 1, time.Now().Add(time.Second))
	require.EqualValues(t, input, output)
}

func TestRetryBufferSkipOnHeavyLoad(t *testing.T) {
	var buff strings.Builder
	log.SetOutput(&buff)

	output := make([]coll, 0, 10)
	var rb *RetryBuffer
	i := 0
	rb = NewRetryBuffer(5, func(lines []byte, retryCountDown int, expires time.Time) {
		output = append(output, coll{lines, retryCountDown})
	}, func(lines []byte, retryCountDown int, expires time.Time) {
		i++
		for actual := rb.first; actual != nil; actual = actual.Next {
			if actual.Expires.Before(expires) {
				t.Errorf("%v entry was not remove but %v was", actual.Expires, expires)
			}
		}
	})
	for i := 0; i < 10; i++ {
		s := []byte(fmt.Sprintf("a%d", i))
		rb.AddLines(s, i, 100+i, time.Now().Add(time.Second))
	}
	rb.Flush()
	assert.Equal(t, 0, rb.Close())
	fmt.Println(buff.String())
	// 5 items should be logged
	assert.True(t, len(strings.Split(buff.String(), "\n")) > 0)
	// At most 5 items should be written
	assert.True(t, len(output) < 6)
}

func TestRetryBufferIgnoreOldWhenBigChunks(t *testing.T) {
	var buff strings.Builder
	log.SetOutput(&buff)
	output := make([]coll, 0, 10)
	rb := NewRetryBuffer(2, func(lines []byte, retryCountDown int, expires time.Time) {
		output = append(output, coll{lines, retryCountDown})
	}, nil)
	rb.AddLines([]byte("1\n2\n3\n"), 1, 100, time.Now().Add(time.Second))
	rb.AddLines([]byte("4\n5\n6\n"), 1, 100, time.Now().Add(time.Second))
	rb.Flush()
	fmt.Println(buff.String())
	// 5 items should be logged
	assert.True(t, len(strings.Split(buff.String(), "\n")) > 0)
	assert.EqualValues(t, []coll{{[]byte("4\n5\n6\n"), 1}}, output)
}

func TestRetryBufferRetriesAfterFlush(t *testing.T) {
	output := make([]coll, 0, 10)
	rb := NewRetryBuffer(5, func(lines []byte, retryCountDown int, expires time.Time) {
		output = append(output, coll{lines, retryCountDown})
	}, nil)
	rb.AddLines([]byte("a"), 1, 20, time.Now().Add(time.Second))
	rb.Flush()
	<-time.After(21 * time.Millisecond)
	assert.Equal(t, 0, rb.Close())
	assert.EqualValues(t, []coll{{[]byte("a"), 1}}, output)
}

func waitForCondition(t *testing.T, timeout int, a func() bool) {
	step := 5
	for {
		<-time.After(time.Duration(step) * time.Millisecond)
		timeout -= step
		if timeout < 0 {
			t.Fatal("wait timeout")
		}
		if a() {
			return
		}
	}

}
