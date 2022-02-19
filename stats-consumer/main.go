package consumer

import (
	"encoding/json"
	"strings"
	"sync"
	"time"
)

type Counts map[string]uint

type TStats struct {
	EventTypeCounts Counts
	WordCounts      Counts
}

type Event struct {
	EventType string     `json:"event_type"`
	Data      string     `json:"data"`
	TimeStamp *time.Time `json:"timestamp,omitempty"`
}

type IConsumer interface {
	Consume(jsonData json.RawMessage) error
	getStats() TStats
}

type Consumer struct {
	stats TStats
	mu    sync.Mutex
}

func NewConsumer() IConsumer {
	stats := TStats{
		EventTypeCounts: make(Counts),
		WordCounts:      make(Counts),
	}
	return &Consumer{
		stats: stats,
	}
}

func (c *Consumer) Consume(jsonData json.RawMessage) error {
	var event *Event
	err := json.Unmarshal(jsonData, &event)
	if err != nil {
		return err
	}

	evtTypeCount, ok := c.stats.EventTypeCounts[event.EventType]
	if ok {
		c.mu.Lock()
		evtTypeCount++
		c.mu.Unlock()
	} else {
		c.mu.Lock()
		c.stats.EventTypeCounts[event.EventType] = 0
		c.stats.EventTypeCounts[event.EventType]++
		c.mu.Unlock()
	}

	wordCounts := getWordCounts(event.Data)
	for word, count := range wordCounts {
		_, ok = c.stats.WordCounts[word]
		if !ok {
			c.stats.WordCounts[word] = 0
		}
		c.mu.Lock()
		c.stats.WordCounts[word] += count
		c.mu.Unlock()
	}

	return nil
}

func getWordCounts(data string) Counts {
	counts := make(Counts)
	words := strings.Split(data, " ")
	for _, word := range words {
		_, ok := counts[word]
		if !ok {
			counts[word] = 0
		}
		counts[word]++
	}
	return counts
}

func (c *Consumer) getStats() TStats {
	return c.stats
}

/*
JSON -> Event -> EventsChannel <- go ConsumeEvent -> TStats
N workers
*/
