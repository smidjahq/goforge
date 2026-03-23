package ui

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestSpinner_Stop_WritesFinalMessage(t *testing.T) {
	var buf bytes.Buffer
	s := New(&buf)
	s.Start("working")
	time.Sleep(20 * time.Millisecond) // let at least one tick fire
	s.Stop("all done")

	out := buf.String()
	if !strings.Contains(out, "all done") {
		t.Errorf("Stop output %q does not contain final message", out)
	}
	if !strings.Contains(out, "✓") {
		t.Errorf("Stop output %q does not contain checkmark ✓", out)
	}
}

func TestSpinner_Fail_WritesCrossAndMessage(t *testing.T) {
	var buf bytes.Buffer
	s := New(&buf)
	s.Start("working")
	time.Sleep(20 * time.Millisecond)
	s.Fail("something went wrong")

	out := buf.String()
	if !strings.Contains(out, "something went wrong") {
		t.Errorf("Fail output %q does not contain error message", out)
	}
	if !strings.Contains(out, "✗") {
		t.Errorf("Fail output %q does not contain cross ✗", out)
	}
}

func TestSpinner_Stop_WithoutStart(t *testing.T) {
	// Stop on a never-started spinner should not panic.
	var buf bytes.Buffer
	s := New(&buf)
	s.Stop("done")

	if !strings.Contains(buf.String(), "done") {
		t.Errorf("expected final message in output, got %q", buf.String())
	}
}

func TestSpinner_Update_ChangesMessage(t *testing.T) {
	var buf bytes.Buffer
	s := New(&buf)
	s.Start("step one")
	time.Sleep(20 * time.Millisecond)
	s.Update("step two")
	time.Sleep(100 * time.Millisecond) // let the ticker pick up the new message
	s.Stop("done")

	out := buf.String()
	if !strings.Contains(out, "step two") {
		t.Errorf("output %q does not show updated message", out)
	}
}

func TestSpinner_DoubleStop_DoesNotPanic(t *testing.T) {
	var buf bytes.Buffer
	s := New(&buf)
	s.Start("working")
	s.Stop("done")
	// Second Stop must not panic (channel already closed).
	s.Stop("done again")
}
