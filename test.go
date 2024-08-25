package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var (
	addr            = flag.String("addr", "localhost:8080", "http service address")
	numConnections  = flag.Int("connections", 100, "number of WebSocket connections")
	messagesPerSec  = flag.Float64("rate", 1.0, "messages per second per connection")
	testDuration    = flag.Duration("duration", 60*time.Second, "test duration")
	totalMessages   = atomic.Int64{}
	successMessages = atomic.Int64{}
)

func runClient(ctx context.Context, id int, wg *sync.WaitGroup) {
	defer wg.Done()

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	c, done2 := connectWS(u, id)
	if done2 {
		return
	}
	defer c.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				//log.Printf("Client %d read error: %v", id, err)
				return
			}
			if len(message) < 100 {
				log.Printf("Client %d received: %s", id, message)
			} else {
				log.Printf("Client %d received big message of length %d", id, len(message))
			}
			successMessages.Add(1)
		}
	}()

	ticker := time.NewTicker(time.Duration(float64(time.Second) / *messagesPerSec))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Printf("Client %d write close error: %v", id, err)
			}
			select {
			case <-done:
				return
			case <-time.After(time.Second):
				return
			}
		case <-ticker.C:
			msg := FlipMessage{Flip: rand.Intn(100 * 1)}
			jsonMsg, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Client %d JSON marshal error: %v", id, err)
				continue
			}

			log.Printf("Client %d sent message: %v", id, msg)

			err = c.WriteMessage(websocket.TextMessage, jsonMsg)
			if err != nil {
				log.Printf("Client %d write error: %v", id, err)
				return
			}
			totalMessages.Add(1)
		}
	}
}

func connectWS(u url.URL, id int) (*websocket.Conn, bool) {
	for retries := 1; retries <= 5; retries++ {
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			// sleep for a while
			sleepDuration := time.Duration(float64(retries)+rand.Float64()) * time.Second
			time.Sleep(sleepDuration)
		} else {
			return c, false
		}
	}
	log.Printf("Client %d fails to connect after 5 retries", id)
	return nil, true
}

func test() {
	flag.Parse()
	log.SetFlags(0)

	ctx, cancel := context.WithTimeout(context.Background(), *testDuration)
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < *numConnections; i++ {
		wg.Add(1)
		go runClient(ctx, i, &wg)
	}

	<-ctx.Done()
	log.Println("Test duration completed. Waiting for goroutines to finish...")
	wg.Wait()

	totalMsgs := totalMessages.Load()
	successMsgs := successMessages.Load()
	log.Printf("Total messages sent: %d", totalMsgs)
	log.Printf("Messages sent per second: %.2f", float64(totalMsgs)/testDuration.Seconds())
	log.Printf("Successful messages received: %d", successMsgs)
	log.Printf("Messages received per second: %.2f", float64(successMsgs)/testDuration.Seconds())
}
