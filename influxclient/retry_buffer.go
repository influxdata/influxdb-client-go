package influxclient

import (
	"log"
	"strings"
	"sync"
	"time"
)

type RetryItem struct {
	Lines      []byte
	RetryCount int
	RetryTime  time.Time
	Expires    time.Time
	Next       *RetryItem
}

func findRemovableItem(first *RetryItem) (found, parent *RetryItem) {
	parent = nil
	found = first
	currentParent := first
	for currentParent.Next != nil {
		if currentParent.Next.Expires.Before(found.Expires) {
			parent = currentParent
			found = currentParent.Next
		}
		currentParent = currentParent.Next
	}
	return
}

type RetryLines func(lines []byte, retryCountDown int, expires time.Time)

// OnRemoveCallback is callback used when inform about removed batch from retry buffer
// Params:
//
//	lines - lines that were skipped
//	retryCountDown - number of remaining attempts
//	expires - time when batch expires
type OnRemoveCallback func(lines []byte, retryCountDown int, expires time.Time)

type RetryBuffer struct {
	first      *RetryItem
	size       int
	maxSize    int
	retryLines RetryLines
	onRemove   OnRemoveCallback
	timer      *time.Timer
	closed     bool
	mu         sync.Mutex
}

func NewRetryBuffer(maxSize int, retryLines RetryLines, onRemove OnRemoveCallback) *RetryBuffer {
	return &RetryBuffer{
		maxSize:    maxSize,
		retryLines: retryLines,
		onRemove:   onRemove,
	}
}

func linesCount(lines []byte) int {
	c := 0
	s := 0
	for {
		i := strings.Index(string(lines[s:]), "\n")
		if i == -1 {
			break
		}
		c++
		s += i + 1
	}
	if len(lines[s:]) > 0 {
		c++
	}
	return c
}

func (r *RetryBuffer) AddLines(lines []byte, retryCount, delay int, expires time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return
	}
	if len(lines) == 0 {
		return
	}
	c := linesCount(lines)
	retryTime := time.Now().Add(time.Millisecond * time.Duration(delay))
	if expires.Before(retryTime) {
		retryTime = expires
	}
	// ensure at most maxLines are in the Buffer
	if r.first != nil && r.size+c > r.maxSize {
		origSize := r.size
		newSize := int(0.7 * float32(origSize)) // reduce to 70 %
		for r.first != nil && r.size+c > newSize {
			// remove "oldest" item
			found, parent := findRemovableItem(r.first)
			r.size -= linesCount(found.Lines)
			if parent != nil {
				parent.Next = found.Next
			} else {
				r.first = found.Next
				if r.first != nil {
					r.scheduleRetry(time.Until(r.first.RetryTime))
				}
			}
			found.Next = nil
			if r.onRemove != nil {
				r.onRemove(found.Lines, found.RetryCount, found.Expires)
			}
		}
		log.Printf("![E] RetryBuffer: %d oldest lines removed to keep buffer size under the limit of %d lines", origSize-r.size, r.maxSize)
	}
	toAdd := &RetryItem{
		Lines:      lines,
		RetryCount: retryCount,
		RetryTime:  retryTime,
		Expires:    expires,
	}
	// insert sorted according to retryTime
	current := r.first
	var parent *RetryItem = nil
	for {
		if current == nil || current.RetryTime.After(retryTime) {
			toAdd.Next = current
			if parent != nil {
				parent.Next = toAdd
			} else {
				r.first = toAdd
				r.scheduleRetry(time.Until(retryTime))
			}
			break
		}
		parent = current
		current = current.Next
	}
	r.size += c
}

func (r *RetryBuffer) removeLines() *RetryItem {
	if r.first != nil {
		toRetry := r.first
		r.first = r.first.Next
		toRetry.Next = nil
		r.size -= linesCount(toRetry.Lines)
		return toRetry
	}
	return nil
}

func (r *RetryBuffer) scheduleRetry(delay time.Duration) {
	if r.timer != nil {
		r.timer.Stop()
	}
	if delay < 0 {
		delay = 0
	}
	r.timer = time.AfterFunc(delay, func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		toRetry := r.removeLines()
		if toRetry != nil {
			r.retryLines(
				toRetry.Lines,
				toRetry.RetryCount,
				toRetry.Expires)

			if r.first != nil {
				r.scheduleRetry(time.Until(r.first.RetryTime))
			}
		} else {
			r.timer = nil
		}
	})
}

func (r *RetryBuffer) Flush() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for {
		toRetry := r.removeLines()
		if toRetry == nil {
			break
		}
		r.retryLines(toRetry.Lines, toRetry.RetryCount, toRetry.Expires)
	}
}

func (r *RetryBuffer) Close() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.timer != nil {
		r.timer.Stop()
		r.timer = nil
	}
	r.closed = true
	return r.size
}
