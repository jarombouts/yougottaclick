package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
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

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
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
			}

		case <-fullStateTicker.C:
			mutex.Lock()
			fullState := FullState{
				State: base64.StdEncoding.EncodeToString(bitfield),
			}
			mutex.Unlock()
			jsonFullState, _ := json.Marshal(fullState)
			broadcast <- jsonFullState
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
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
		mutex.Unlock()
	}
}

func main() {
	fmt.Println("Starting...")

	http.HandleFunc("/ws", handleConnections)
	go handleUpdates()
	go handleBroadcast()

	fmt.Println("Server listening for connections.")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
