package main

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

type ScoreUpdate struct {
	Score  int64 `json:"score"`
	Clicks int64 `json:"clicks"`
	Hot    int64 `json:"hot"`
}
