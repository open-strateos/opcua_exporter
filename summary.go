package main

import (
	"context"
	"log"
	"sync"
	"time"
)

// NewEventSummaryCounter creates a new event counter that will
// log a summary at every interval.
func NewEventSummaryCounter(interval time.Duration) *EventSummaryCounter {
	return &EventSummaryCounter{
		Interval: interval,
		Counts:   make(map[string]int),
	}
}

// EventSummaryCounter keeps a set of counters, one for each event name string it receives
// Periodically, it logs all the counts and then clears its timers.
// This is to reduce log volume while allowing some visibility into the shape of the OPC-UA event traffic
// being received by the exporter
type EventSummaryCounter struct {
	Interval time.Duration
	Counts   map[string]int
	Total    int
	mutex    sync.Mutex
}

// Inc adds one to the counter for the given channel
func (esc *EventSummaryCounter) Inc(channel string) {
	esc.mutex.Lock()
	if _, ok := esc.Counts[channel]; ok {
		esc.Counts[channel]++
	} else {
		esc.Counts[channel] = 1
	}
	esc.Total++
	esc.mutex.Unlock()
}

// Reset all the counters to zero
func (esc *EventSummaryCounter) Reset() {
	esc.mutex.Lock()
	esc.Counts = make(map[string]int)
	esc.Total = 0
	esc.mutex.Unlock()
}

// Start the goroutine that periodically logs the counter summary,
// then resets the counters.
func (esc *EventSummaryCounter) Start(ctx context.Context) {
	go func() {
		log.Printf("Starting %v summary timer", esc.Interval.String())
		esc.Reset()
		ticker := time.NewTicker(esc.Interval)
		for {
			select {
			case <-ticker.C:
				esc.logSummary()
				esc.Reset()
			case <-ctx.Done():
				log.Println("Exiting summary printing loop")
				break
			}
		}
	}()
}

func (esc *EventSummaryCounter) logSummary() {
	log.Printf("Received %d events on %d channels in the last %v", esc.Total, len(esc.Counts), esc.Interval.String())
	for channel, count := range esc.Counts {
		rate := float64(count) / esc.Interval.Seconds()
		log.Printf("CHANNEL: %v\tEVENTS: %v (%.3f per second)", channel, count, rate)
	}
}
