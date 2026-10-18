package launcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_OutputTail_WhenOutputExceedsLimit_KeepsLastWholeLines(t *testing.T) {
	tail := newOutputTail(30)

	_, _ = tail.Write([]byte("first line\nsecond line\n"))
	_, _ = tail.Write([]byte("client tools missing\n"))

	assert.Equal(t, "client tools missing", tail.GetContents())
}

func Test_OutputTail_WhenOutputFits_KeepsEverything(t *testing.T) {
	tail := newOutputTail(100)

	_, _ = tail.Write([]byte("first line\nsecond line\n"))

	assert.Equal(t, "first line\nsecond line", tail.GetContents())
}
