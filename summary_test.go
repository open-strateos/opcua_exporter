package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventSummary(t *testing.T) {
	duration := 4 * time.Millisecond
	esc := NewEventSummaryCounter(duration)
	esc.Inc("foo")
	esc.Inc("foo")
	esc.Inc("foo")
	esc.Inc("bar")

	assert.Equal(t, 3, esc.Counts["foo"])
	assert.Equal(t, 1, esc.Counts["bar"])
	assert.Equal(t, 4, esc.Total)
	assert.Equal(t, duration, esc.Interval)

	esc.Reset()
	assert.Empty(t, esc.Counts)
	assert.Zero(t, esc.Total)
}
