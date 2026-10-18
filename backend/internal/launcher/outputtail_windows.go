package launcher

import (
	"bytes"
	"sync"
)

// outputTail keeps the end of a child's console output, so the reason a child exited on its own
// can travel with the error up to the service's event log entry.
type outputTail struct {
	mutex     sync.Mutex
	limit     int
	contents  []byte
	isTrimmed bool
}

func newOutputTail(limit int) *outputTail {
	return &outputTail{limit: limit}
}

func (t *outputTail) Write(chunk []byte) (int, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.contents = append(t.contents, chunk...)

	if excess := len(t.contents) - t.limit; excess > 0 {
		t.contents = append([]byte(nil), t.contents[excess:]...)
		t.isTrimmed = true
	}

	return len(chunk), nil
}

// GetContents drops the partial first line left by trimming.
func (t *outputTail) GetContents() string {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	contents := t.contents
	if t.isTrimmed {
		if newline := bytes.IndexByte(contents, '\n'); newline >= 0 {
			contents = contents[newline+1:]
		}
	}

	return string(bytes.TrimSpace(contents))
}
