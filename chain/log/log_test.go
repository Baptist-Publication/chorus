package log

import "testing"

func TestFoo(t *testing.T) {
	l := Initialize("production", "/tmp/cc.log", "/tmp/ce.log")

	l.Debug("debug")
	l.Info("info")
	l.Warn("warn")
	l.Error("error")

}
