package model

import "time"

// Player represents a single connected player
type Player struct {
	ID        string
	Name      string
	Level     int
	Health    int
	Deaths    int
	Zombies   int
	Score     int
	Ping      int
	SteamID   string
	IP        string
}

// ServerStats holds the aggregate state of the server
type ServerStats struct {
	Host        string
	Time        string // In-game time
	Uptime      time.Duration
	HeapUsed    string
	HeapMax     string
	Fps         string
	PlayerCount int
}
