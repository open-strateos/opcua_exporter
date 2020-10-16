package main

import (
	"context"
	"log"
	"sync"
	"time"
)

func NewEventSummaryCounter(interval time.Duration) *EventSummaryCounter {
	return &EventSummaryCounter{
		Interval: interval,
		Counts:   make(map[string]int),
	}
}

type EventSummaryCounter struct {
	Interval time.Duration
	Counts   map[string]int
	Total    int
	mutex    sync.Mutex
}

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

func (esc *EventSummaryCounter) Reset() {
	esc.mutex.Lock()
	esc.Counts = make(map[string]int)
	esc.Total = 0
	esc.mutex.Unlock()
}

func (esc *EventSummaryCounter) Start(ctx context.Context) {
	go func() {
		log.Printf("Starting %v summary timer", esc.Interval.String())
		esc.Reset() // Initializes the map
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
