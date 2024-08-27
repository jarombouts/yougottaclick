package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/websocket"
)

const (
	bitfieldSize            = 1024 * 1024
	updateInterval          = 333 * time.Millisecond
	fullStateInterval       = 30 * time.Second
	persistBitfieldInterval = 60 * time.Second
	minTimeBetweenChanges   = 50 * time.Millisecond
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	bitfield     = make([]byte, bitfieldSize/8)
	prevBitfield = make([]byte, bitfieldSize/8)
	clicks       int64
	hot          int64
	clients      = make(map[*websocket.Conn]time.Time)
	scores       = make(map[*websocket.Conn]int64)
	mutex        = &sync.Mutex{}
	broadcast    = make(chan []byte)
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	mutex.Lock()
	clients[conn] = time.Now().Add(-time.Second)
	mutex.Unlock()

	defer func() {
		mutex.Lock()
		delete(clients, conn)
		mutex.Unlock()
		conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		var flipMsg FlipMessage
		err = json.Unmarshal(msg, &flipMsg)
		if err != nil {
			log.Printf("Invalid message format: %v", err)
			continue
		}

		index := flipMsg.Flip
		if index >= 0 && index < bitfieldSize {
			mutex.Lock()
			lastUpdate, ok := clients[conn]
			if ok && time.Since(lastUpdate) >= minTimeBetweenChanges {
				// set new bit in bitfield
				byteIndex := index / 8
				bitOffset := index % 8
				bitfield[byteIndex] ^= 1 << bitOffset
				newState := int64(((bitfield[byteIndex]) & (1 << bitOffset)) >> bitOffset)
				// update score and last time since click
				scores[conn] = (scores[conn] + newState) * newState // +1 if new state is 'checked', go to 0 if new state is 'unchecked'
				clients[conn] = time.Now()
				// update click counter
				clicks += 1
				hot += newState*2 - 1
				mutex.Unlock()
			} else {
				// immediately unlock and apply sleep penalty
				mutex.Unlock()
				sleepDuration := time.Duration(500.+200.*rand.Float64()) * time.Millisecond
				time.Sleep(sleepDuration)
			}
		}
	}
}

func handleUpdates() {
	updateTicker := time.NewTicker(updateInterval)
	fullStateTicker := time.NewTicker(fullStateInterval)
	persistBitfieldTicker := time.NewTicker(persistBitfieldInterval)

	for {
		select {
		case <-updateTicker.C:
			update := BitfieldUpdate{}
			mutex.Lock()
			for i := 0; i < bitfieldSize; i++ {
				byteIndex := i / 8
				bitOffset := i % 8
				prevBit := prevBitfield[byteIndex]&(1<<bitOffset) != 0
				currentBit := bitfield[byteIndex]&(1<<bitOffset) != 0
				if prevBit != currentBit {
					if currentBit {
						update.One = append(update.One, i)
					} else {
						update.Zero = append(update.Zero, i)
					}
				}
			}
			copy(prevBitfield, bitfield)
			mutex.Unlock()

			if len(update.Zero) > 0 || len(update.One) > 0 {
				jsonUpdate, _ := json.Marshal(update)
				broadcast <- jsonUpdate
				log.Printf("Sent state diffs to %d clients (%d changed bits)", len(clients), len(update.Zero)+len(update.One))
			}

		case <-fullStateTicker.C:
			mutex.Lock()
			fullState := FullState{
				State: base64.StdEncoding.EncodeToString(bitfield),
			}
			hot = countOnes(bitfield)
			mutex.Unlock()

			jsonFullState, _ := json.Marshal(fullState)
			broadcast <- jsonFullState
			log.Printf("Sent full state to %d clients (%d hot bits, %d cumulative clicks)", len(clients), hot, clicks)

		case <-persistBitfieldTicker.C:
			err := saveBitfield()
			if err != nil {
				log.Printf("error saving bitfield: %v", err)
			}
		}
	}
}

func handleBroadcast() {
	for {
		msg := <-broadcast
		mutex.Lock()
		for client := range clients {
			sendIncomingMessage(client, msg)
			sendScoreMessage(client)
		}
		mutex.Unlock()
	}
}

func getState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mutex.Lock()
	fullState := FullState{
		State: base64.StdEncoding.EncodeToString(bitfield),
	}
	mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fullState)
}

func main() {
	log.Println("Starting...")
	err := loadBitfield()
	if err != nil {
		log.Fatalf("error loading bitfield: %v", err)
	}

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/state", getState)

	go handleUpdates()
	go handleBroadcast()

	log.Println("Server listening for connections on port 8008")
	log.Fatal(http.ListenAndServe(":8008", handlers.LoggingHandler(os.Stdout, http.DefaultServeMux)))
}
