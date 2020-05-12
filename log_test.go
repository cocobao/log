package log

import (
	"testing"
	"time"
)

func TestDebug(t *testing.T) {
	Debug("Hello")
	time.Sleep(time.Second)
}
