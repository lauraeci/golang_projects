package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"time"

	"github.com/sacOO7/gowebsocket"
)

const (
	// HistoryLength how many minutes to store in history buffer
	HistoryLength = 60
	// TimezoneDefault for events minute distribution
	TimezoneDefault = "America/Los_Angeles"
)

// User user of the system
type User struct {
	ID       string `json:"id"`
	ImageURL string `json:"image_url"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

// Event an event for a user
type Event struct {
	ID        string   `json:"id"`
	Timestamp float64  `json:"timestamp"`
	User      User     `json:"user"`
	Message   string   `json:"message"`
	Tags      []string `json:"tags"`
}

// Stats tracks summary data of events for reporting when the program exists
type Stats struct {
	count        int
	startedAt    time.Time
	endedAt      time.Time
	min          int
	max          int
	history      map[int]int
	distribution map[int][]time.Time
}

func (s *Stats) averagePerMinute() float64 {
	t := s.endedAt.Sub(s.startedAt).Minutes()
	return float64(s.count) / t
}

func timeFromFloat64(ts float64) time.Time {
	secs := int64(ts)
	nsecs := int64((ts - float64(secs)) * 1e9)
	return time.Unix(secs, nsecs)
}

func maxTime(t0 time.Time, t1 time.Time) time.Time {
	if t0.After(t1) {
		return t0
	}
	return t1
}

func min(x int, y int) int {
	if x < y {
		return x
	}
	return y
}

func max(x int, y int) int {
	if x > y {
		return x
	}
	return y
}

func (s *Stats) saveHistory() {
	minutes := int(s.endedAt.Sub(s.startedAt).Minutes())
	val, ok := s.history[minutes]
	if ok {
		s.history[minutes] = val + 1
	} else {
		if len(s.history) > 1 {
			count := s.history[minutes-1]
			s.max = max(count, s.max)
			s.min = min(count, s.min)
			loc, _ := time.LoadLocation(TimezoneDefault)
			minute := time.Date(s.endedAt.Year(), s.endedAt.Month(), s.endedAt.Day(), s.endedAt.Hour(), s.endedAt.Minute(), 0, 0, loc)
			s.distribution[count] = append(s.distribution[count], minute)
		}
		s.history[minutes] = 1
	}

	if len(s.history) == HistoryLength {
		// reset
		s.history = map[int]int{}
	}
}

func (s *Stats) collect(message string, logEvents bool) error {
	event := &Event{}
	err := json.Unmarshal([]byte(message), event)
	if err == nil {
		return fmt.Errorf("failed to parse message with error: %v", err)
	}
	t := timeFromFloat64(event.Timestamp / 1000)
	s.count++
	if s.startedAt.IsZero() {
		s.startedAt = t
		return nil
	}
	s.endedAt = maxTime(t, s.endedAt)
	s.saveHistory()
	if logEvents {
		s.summarize()
		log.Printf("Event: %+v\n", event)
	}
	return nil
}

func initStats() Stats {
	history := map[int]int{}
	distribution := map[int][]time.Time{}
	stats := Stats{
		history:      history,
		distribution: distribution,
		min:          math.MaxInt32,
		max:          0,
	}
	return stats
}

func (s *Stats) summarize() {
	log.Println("Min Count: ", s.min, "Max Count: ", s.max)
	log.Println("Stats:", s.history)
	log.Println("Distribution:", s.distribution)
	log.Printf("Average number of events per minute: %v", s.averagePerMinute())
}

func readServerEvents() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	logEvents := true
	stats := initStats()
	socket := gowebsocket.New("wss://chaotic-stream.herokuapp.com")

	socket.OnConnected = func(socket gowebsocket.Socket) {
		log.Println("Connected to server")
	}

	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Println("Recieved connect error ", err)
	}

	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		stats.collect(message, logEvents)
	}

	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		stats.summarize()
		log.Println("Disconnected from server ")
		return
	}

	socket.Connect()

	for {
		select {
		case <-interrupt:
			log.Println("interrupt")
			socket.Close()
			return
		}
	}
}

func main() {
	readServerEvents()
}
