package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
)

func sendScoreMessage(client *websocket.Conn) {
	scoreMsg := ScoreUpdate{
		Score:  scores[client],
		Hot:    hot,
		Clicks: clicks,
	}
	jsonScoreMsg, _ := json.Marshal(scoreMsg)
	err := client.WriteMessage(websocket.TextMessage, jsonScoreMsg)
	if err != nil {
		client.Close()
		delete(clients, client)
	}
}

func sendIncomingMessage(client *websocket.Conn, msg []byte) {
	err := client.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		//log.Printf("error: %v", err)
		client.Close()
		delete(clients, client)
	}
}
