package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"math/bits"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	bitfieldSize      = 1024 * 1024
	updateInterval    = time.Second
	fullStateInterval = 30 * time.Second
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	bitfield     = make([]byte, bitfieldSize/8)
	prevBitfield = make([]byte, bitfieldSize/8)
	clients      = make(map[*websocket.Conn]time.Time)
	mutex        = &sync.Mutex{}
	broadcast    = make(chan []byte)
)

type FlipMessage struct {
	Flip int `json:"flip"`
}

type BitfieldUpdate struct {
	Zero []int `json:"0,omitempty"`
	One  []int `json:"1,omitempty"`
}

type FullState struct {
	State string `json:"state"`
}

func countOnes(bitfield []byte) int {
	count := 0
	for _, b := range bitfield {
		count += bits.OnesCount8(b)
	}
	return count
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		//log.Println(err)
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
			if ok && time.Since(lastUpdate) >= time.Second {
				byteIndex := index / 8
				bitOffset := index % 8
				bitfield[byteIndex] ^= 1 << bitOffset
				clients[conn] = time.Now()
			}
			mutex.Unlock()
		}
	}
}

func handleUpdates() {
	updateTicker := time.NewTicker(updateInterval)
	fullStateTicker := time.NewTicker(fullStateInterval)

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
				log.Printf("Sent state diffs: %d changed bits", len(update.Zero)+len(update.One))
			}

		case <-fullStateTicker.C:
			mutex.Lock()
			fullState := FullState{
				State: base64.StdEncoding.EncodeToString(bitfield),
			}
			mutex.Unlock()
			jsonFullState, _ := json.Marshal(fullState)
			broadcast <- jsonFullState
			log.Printf("Sent full state: %d hot bits out of %d", countOnes(bitfield), bitfieldSize)
		}
	}
}

func handleBroadcast() {
	for {
		msg := <-broadcast
		mutex.Lock()
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				//log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
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

// withCORS adds CORS headers to responses
func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			return
		}
		next.ServeHTTP(w, r)
	}
}

func main() {
	log.Println("Starting...")

	http.Handle("/", http.FileServer(http.Dir("./templates")))
	http.HandleFunc("/ws", withCORS(handleConnections))
	http.HandleFunc("/state", withCORS(getState))

	go handleUpdates()
	go handleBroadcast()

	log.Println("Server listening for connections.")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
